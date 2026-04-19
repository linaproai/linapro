// This file defines deployment-aware permission-topology revision controllers
// selected during role-service construction.

package role

import (
	"context"

	"lina-core/internal/service/kvcache"
)

// accessRevisionController hides the single-node and clustered revision
// synchronization strategies behind one common contract.
type accessRevisionController interface {
	// CurrentRevision returns the effective permission-topology revision currently visible to the process.
	CurrentRevision(ctx context.Context) (int64, error)
	// SyncRevision refreshes the process-local revision from the active source and clears stale token snapshots when needed.
	SyncRevision(ctx context.Context, onRevisionChange func()) (int64, error)
	// MarkChanged records one topology mutation and returns the new effective revision.
	MarkChanged(ctx context.Context) (int64, error)
}

// localAccessRevisionController keeps permission-topology invalidation fully
// in process memory for single-node deployments.
type localAccessRevisionController struct{}

// clusterAccessRevisionController synchronizes topology revision through shared
// KV while preserving a local hot-path copy for request-time permission checks.
type clusterAccessRevisionController struct {
	kvCacheSvc kvcache.Service
}

// newAccessRevisionController selects the deployment-specific controller once
// during service construction so access-cache call sites remain branch-free.
func newAccessRevisionController(
	clusterEnabled bool,
	kvCacheSvc kvcache.Service,
) accessRevisionController {
	if clusterEnabled {
		return &clusterAccessRevisionController{kvCacheSvc: kvCacheSvc}
	}
	return &localAccessRevisionController{}
}

// CurrentRevision reuses the last in-process value or lazily initializes one
// so single-node permission checks never need a distributed coordination step.
func (c *localAccessRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if revision, ok := getLocalAccessRevisionForce(); ok {
		return revision, nil
	}
	revision := bumpLocalAccessRevision()
	return revision, nil
}

// SyncRevision is the same as CurrentRevision in single-node mode because no
// other node can advance the permission topology independently.
func (c *localAccessRevisionController) SyncRevision(ctx context.Context, _ func()) (int64, error) {
	return c.CurrentRevision(ctx)
}

// MarkChanged advances the local revision immediately after one topology write
// so token snapshots created afterward bind to the new version.
func (c *localAccessRevisionController) MarkChanged(_ context.Context) (int64, error) {
	return bumpLocalAccessRevision(), nil
}

// CurrentRevision first tries the short-lived local copy and only re-reads
// shared KV when this process needs to resynchronize.
func (c *clusterAccessRevisionController) CurrentRevision(ctx context.Context) (int64, error) {
	if revision, ok := getLocalAccessRevision(); ok {
		return revision, nil
	}

	revision, err := c.getSharedRevision(ctx)
	if err != nil {
		// Keep permission checks soft-degraded during transient shared-KV
		// failures by reusing the last synchronized revision when one exists.
		if fallbackRevision, ok := getLocalAccessRevisionForce(); ok {
			return fallbackRevision, nil
		}
		return 0, err
	}

	storeLocalAccessRevision(revision)
	return revision, nil
}

// SyncRevision is used by the background watcher, so it must always consult
// shared KV and optionally trigger token-cache eviction when another node wrote
// a newer topology revision.
func (c *clusterAccessRevisionController) SyncRevision(
	ctx context.Context,
	onRevisionChange func(),
) (int64, error) {
	revision, err := c.getSharedRevision(ctx)
	if err != nil {
		return 0, err
	}

	if localRevision, ok := getLocalAccessRevisionForce(); ok && localRevision != revision && onRevisionChange != nil {
		onRevisionChange()
	}
	storeLocalAccessRevision(revision)
	return revision, nil
}

// MarkChanged publishes one shared revision bump and then updates the local
// copy so the writing node observes its own topology mutation immediately.
func (c *clusterAccessRevisionController) MarkChanged(ctx context.Context) (int64, error) {
	item, err := c.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		accessRevisionCacheKey,
		1,
		0,
	)
	if err != nil {
		return 0, err
	}

	storeLocalAccessRevision(item.IntValue)
	return item.IntValue, nil
}

// getSharedRevision reads the current shared counter without incrementing it so
// background refreshes stay side-effect free.
func (c *clusterAccessRevisionController) getSharedRevision(ctx context.Context) (int64, error) {
	revision, _, err := c.kvCacheSvc.GetInt(
		ctx,
		kvcache.OwnerTypeModule,
		accessRevisionCacheKey,
	)
	return revision, err
}
