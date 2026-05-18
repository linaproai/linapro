// This file defines the deployment-aware runtime-parameter revision
// controllers selected during config-service construction.

package config

import (
	"context"
	"time"

	"lina-core/internal/service/cachecoord"
)

// Runtime-configuration cache coordination reasons.
const (
	// runtimeParamCacheDomain coordinates protected runtime configuration snapshots.
	runtimeParamCacheDomain cachecoord.Domain = "runtime-config"
	// runtimeParamCacheChangeReason records protected runtime parameter mutations.
	runtimeParamCacheChangeReason cachecoord.ChangeReason = "runtime_params_changed"
	// runtimeParamCacheMaxStale is the runtime-config freshness budget.
	runtimeParamCacheMaxStale = 10 * time.Second
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
// the current process so single-node deployments avoid any cachecoord traffic.
type localRuntimeParamRevisionController struct{}

// clusterRuntimeParamRevisionController coordinates revision changes through
// cachecoord while still caching the last synchronized value in process memory.
type clusterRuntimeParamRevisionController struct {
	cacheCoordSvc cachecoord.Service
}

// newCacheCoordRuntimeParamRevisionController selects the deployment-specific
// revision strategy backed by cachecoord in cluster mode.
func newCacheCoordRuntimeParamRevisionController(clusterEnabled bool) runtimeParamRevisionController {
	if clusterEnabled {
		cacheCoordSvc := cachecoord.Default(cachecoord.NewStaticTopology(true))
		configureRuntimeParamCacheDomain(cacheCoordSvc)
		return &clusterRuntimeParamRevisionController{
			cacheCoordSvc: cacheCoordSvc,
		}
	}
	return &localRuntimeParamRevisionController{}
}

// configureRuntimeParamCacheDomain declares the runtime-config consistency
// contract without making cachecoord own a global domain registry.
func configureRuntimeParamCacheDomain(cacheCoordSvc cachecoord.Service) {
	if cacheCoordSvc == nil {
		return
	}
	if err := cacheCoordSvc.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           runtimeParamCacheDomain,
		AuthoritySource:  "sys_config protected runtime parameters",
		ConsistencyModel: cachecoord.ConsistencySharedRevision,
		MaxStale:         runtimeParamCacheMaxStale,
		SyncMechanism:    "persistent sys_cache_revision plus request or watcher refresh",
		FailureStrategy:  cachecoord.FailureStrategyReturnVisibleError,
	}); err != nil {
		panic(err)
	}
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

// CurrentRevision verifies freshness through cachecoord on the request path so
// protected runtime readers do not indefinitely trust a process-local revision.
func (c *clusterRuntimeParamRevisionController) CurrentRevision(ctx context.Context) (int64, error) {
	return c.ensureFresh(ctx)
}

// SyncRevision always refreshes from cachecoord because watcher-driven sync must
// observe cross-node writes even when this process already has a local copy.
func (c *clusterRuntimeParamRevisionController) SyncRevision(ctx context.Context) (int64, error) {
	return c.ensureFresh(ctx)
}

// MarkChanged publishes one cross-node revision bump and then mirrors the new
// value locally so the mutating node does not wait for the next watcher cycle.
func (c *clusterRuntimeParamRevisionController) MarkChanged(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.MarkChanged(
		ctx,
		runtimeParamCacheDomain,
		cachecoord.ScopeGlobal,
		runtimeParamCacheChangeReason,
	)
	if err != nil {
		return 0, err
	}

	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// ensureFresh confirms that the local runtime-parameter snapshot has consumed
// the latest coordinated revision or returns a visible freshness error.
func (c *clusterRuntimeParamRevisionController) ensureFresh(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.EnsureFresh(
		ctx,
		runtimeParamCacheDomain,
		cachecoord.ScopeGlobal,
		func(_ context.Context, revision int64) error {
			storeLocalRuntimeParamRevision(revision)
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}
