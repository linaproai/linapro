// This file defines the deployment-aware runtime-parameter revision
// controllers selected during config-service construction.

package config

import (
	"context"

	"lina-core/internal/service/kvcache"
)

// runtimeParamRevisionController hides the single-node and clustered revision
// synchronization strategies behind one common contract.
type runtimeParamRevisionController interface {
	// CurrentRevision returns the effective revision currently visible to the process.
	CurrentRevision(ctx context.Context) (int64, error)
	// SyncRevision refreshes the process-local revision from the active source.
	SyncRevision(ctx context.Context) (int64, error)
	// MarkChanged records one runtime-parameter mutation and returns the new revision.
	MarkChanged(ctx context.Context) (int64, error)
}

// localRuntimeParamRevisionController keeps revision ownership entirely inside
// the current process so single-node deployments avoid any shared-KV traffic.
type localRuntimeParamRevisionController struct{}

// clusterRuntimeParamRevisionController coordinates revision changes through
// shared KV while still caching the last synchronized value in process memory.
type clusterRuntimeParamRevisionController struct {
	kvCacheSvc kvcache.Service
}

// newRuntimeParamRevisionController selects the deployment-specific revision
// strategy once during service construction so business methods can stay branch-free.
func newRuntimeParamRevisionController(
	clusterEnabled bool,
	kvCacheSvc kvcache.Service,
) runtimeParamRevisionController {
	if clusterEnabled {
		return &clusterRuntimeParamRevisionController{kvCacheSvc: kvCacheSvc}
	}
	return &localRuntimeParamRevisionController{}
}

// CurrentRevision lazily initializes the local revision so a single-node
// process can invalidate snapshots without depending on any external store.
func (c *localRuntimeParamRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if revision, ok := getLocalRuntimeParamRevision(); ok {
		return revision, nil
	}
	revision := bumpLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// SyncRevision is equivalent to CurrentRevision in single-node mode because
// there is no remote source that can advance independently of this process.
func (c *localRuntimeParamRevisionController) SyncRevision(ctx context.Context) (int64, error) {
	return c.CurrentRevision(ctx)
}

// MarkChanged advances the in-process revision immediately after one protected
// config write so subsequent reads rebuild against the new local version.
func (c *localRuntimeParamRevisionController) MarkChanged(_ context.Context) (int64, error) {
	revision := bumpLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// CurrentRevision prefers the synchronized local copy on the request path and
// only falls back to shared KV when this process has not seen a revision yet.
func (c *clusterRuntimeParamRevisionController) CurrentRevision(ctx context.Context) (int64, error) {
	if revision, ok := getLocalRuntimeParamRevision(); ok {
		return revision, nil
	}

	revision, err := c.getSharedRevision(ctx)
	if err != nil {
		return 0, err
	}
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// SyncRevision always refreshes from shared KV because watcher-driven sync must
// observe cross-node writes even when this process already has a local copy.
func (c *clusterRuntimeParamRevisionController) SyncRevision(ctx context.Context) (int64, error) {
	revision, err := c.getSharedRevision(ctx)
	if err != nil {
		return 0, err
	}
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// MarkChanged publishes one cross-node revision bump and then mirrors the new
// value locally so the mutating node does not wait for the next watcher cycle.
func (c *clusterRuntimeParamRevisionController) MarkChanged(ctx context.Context) (int64, error) {
	item, err := c.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		runtimeParamRevisionCacheKey,
		1,
		0,
	)
	if err != nil {
		return 0, err
	}

	storeLocalRuntimeParamRevision(item.IntValue)
	return item.IntValue, nil
}

// getSharedRevision uses the pure read path so clustered refreshes do not
// mutate the shared counter when they only need the current effective version.
func (c *clusterRuntimeParamRevisionController) getSharedRevision(ctx context.Context) (int64, error) {
	revision, _, err := c.kvCacheSvc.GetInt(
		ctx,
		kvcache.OwnerTypeModule,
		runtimeParamRevisionCacheKey,
	)
	return revision, err
}
