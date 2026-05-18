// This file implements cachecoord-backed plugin runtime cache revision control.

package pluginruntimecache

import (
	"context"

	"lina-core/internal/service/cachecoord"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// WithTenantScope returns the controller after applying tenant invalidation scope.
func (c *Controller) WithTenantScope(tenantID pkgtenantcap.TenantID, cascadeToTenants bool) *Controller {
	if c == nil {
		return nil
	}
	c.tenantScope = cachecoord.InvalidationScope{
		TenantID:         cachecoord.TenantID(tenantID),
		CascadeToTenants: cascadeToTenants,
	}
	return c
}

// EnsureFresh refreshes this process-local cache domain when cluster mode is
// enabled and cachecoord reports a newer plugin runtime revision.
func (c *Controller) EnsureFresh(ctx context.Context) error {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return nil
	}
	revision, err := c.CurrentRevision(ctx)
	if err != nil {
		return err
	}
	return c.observed.Ensure(ctx, revision, c.refresher)
}

// CurrentRevision returns the current shared revision for this controller.
func (c *Controller) CurrentRevision(ctx context.Context) (int64, error) {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return 0, nil
	}
	return c.cacheCoordSvc.CurrentRevision(
		ctx,
		runtimeCacheDomain,
		c.scope,
	)
}

// IsObserved reports whether this process-local domain has consumed revision.
func (c *Controller) IsObserved(revision int64) bool {
	if c == nil || !c.clusterEnabled {
		return true
	}
	return c.observed.Matches(revision)
}

// StoreObserved records that this process-local domain has consumed revision.
func (c *Controller) StoreObserved(revision int64) {
	if c == nil || !c.clusterEnabled {
		return
	}
	c.observed.Store(revision)
}

// MarkChanged publishes one plugin runtime cache mutation to other cluster
// nodes. Single-node deployments skip cachecoord and return revision 0.
func (c *Controller) MarkChanged(ctx context.Context) (int64, error) {
	return c.markChanged(ctx, true)
}

// PublishChanged publishes one shared revision without recording it as consumed
// by the local process. It is used by callers that need the background consumer
// on the same node to retry work if the foreground mutation fails afterward.
func (c *Controller) PublishChanged(ctx context.Context) (int64, error) {
	return c.markChanged(ctx, false)
}

// markChanged increments the shared revision and optionally records the
// returned value as already consumed by this local cache domain.
func (c *Controller) markChanged(ctx context.Context, storeObserved bool) (int64, error) {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return 0, nil
	}
	revision, err := c.cacheCoordSvc.MarkTenantChanged(
		ctx,
		runtimeCacheDomain,
		c.scope,
		c.tenantScope,
		c.changeReason,
	)
	if err != nil {
		return 0, err
	}
	if storeObserved {
		c.observed.Store(revision)
	}
	return revision, nil
}

// configureRuntimeCacheDomain declares plugin-runtime consistency policy in
// the package that owns the plugin runtime cache semantics.
func configureRuntimeCacheDomain(clusterEnabled bool, cacheCoordSvc cachecoord.Service) {
	if !clusterEnabled || cacheCoordSvc == nil {
		return
	}
	if err := cacheCoordSvc.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           runtimeCacheDomain,
		AuthoritySource:  "plugin registry, active releases, plugin node state, and artifacts",
		ConsistencyModel: cachecoord.ConsistencySharedRevision,
		MaxStale:         runtimeCacheMaxStale,
		SyncMechanism:    "persistent sys_cache_revision plus runtime cache invalidation",
		FailureStrategy:  cachecoord.FailureStrategyConservativeHide,
	}); err != nil {
		panic(err)
	}
}
