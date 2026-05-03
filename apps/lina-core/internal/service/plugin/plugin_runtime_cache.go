// This file coordinates plugin runtime cache freshness across cluster nodes.

package plugin

import (
	"context"

	"lina-core/internal/service/cachecoord"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/pluginruntimecache"
	"lina-core/pkg/logger"
)

// pluginRuntimeCacheObservedRevision records the revision consumed by the root
// plugin facade cache domain inside this process.
var pluginRuntimeCacheObservedRevision = pluginruntimecache.NewObservedRevision()

// runtimeBundleInvalidator defines the narrow i18n cache invalidation method
// needed by plugin runtime cache refreshes.
type runtimeBundleInvalidator interface {
	// InvalidateRuntimeBundleCache clears cached runtime bundles for one explicit scope.
	InvalidateRuntimeBundleCache(scope i18nsvc.InvalidateScope)
}

// newRuntimeCacheRevisionController creates the cluster-aware revision
// controller used by the root plugin service.
func newRuntimeCacheRevisionController(
	topology Topology,
	integrationSvc pluginRuntimeIntegrationRefresher,
	frontendSvc pluginRuntimeFrontendInvalidator,
	i18nSvc runtimeBundleInvalidator,
) *pluginruntimecache.Controller {
	clusterEnabled := false
	if topology != nil {
		clusterEnabled = topology.IsEnabled()
	}
	return pluginruntimecache.NewControllerWithCoordinator(
		clusterEnabled,
		cachecoord.Default(topology),
		pluginRuntimeCacheObservedRevision,
		func(ctx context.Context) error {
			if integrationSvc != nil {
				if err := integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
					return err
				}
			}
			if frontendSvc != nil {
				frontendSvc.InvalidateAllBundles(ctx, "cluster_runtime_revision_changed")
			}
			wasm.InvalidateAllCache(ctx)
			if i18nSvc != nil {
				i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
					Sectors: []i18nsvc.Sector{
						i18nsvc.SectorSourcePlugin,
						i18nsvc.SectorDynamicPlugin,
					},
				})
			}
			return nil
		},
	)
}

// pluginRuntimeIntegrationRefresher narrows the integration cache refresh dependency.
type pluginRuntimeIntegrationRefresher interface {
	// RefreshEnabledSnapshot rebuilds the in-memory plugin enablement snapshot.
	RefreshEnabledSnapshot(ctx context.Context) error
}

// pluginRuntimeFrontendInvalidator narrows the frontend bundle invalidation dependency.
type pluginRuntimeFrontendInvalidator interface {
	// InvalidateAllBundles removes every cached runtime frontend bundle.
	InvalidateAllBundles(ctx context.Context, reason string)
}

// ensureRuntimeCacheFresh synchronizes plugin runtime caches with the shared
// cluster revision before read paths consume process-local snapshots.
func (s *serviceImpl) ensureRuntimeCacheFresh(ctx context.Context) error {
	if s == nil || s.runtimeCacheRevisionCtrl == nil {
		return nil
	}
	return s.runtimeCacheRevisionCtrl.EnsureFresh(ctx)
}

// ensureRuntimeCacheFreshBestEffort logs revision refresh failures for methods
// that cannot return an error to their caller.
func (s *serviceImpl) ensureRuntimeCacheFreshBestEffort(ctx context.Context, operation string) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		logger.Warningf(ctx, "refresh plugin runtime cache failed operation=%s err=%v", operation, err)
	}
}

// MarkRuntimeCacheChanged publishes one successful runtime cache mutation to
// other cluster nodes. It implements the dynamic runtime cache-change notifier.
func (s *serviceImpl) MarkRuntimeCacheChanged(ctx context.Context, reason string) error {
	return s.markRuntimeCacheChanged(ctx, reason)
}

// markRuntimeCacheChanged bumps the shared plugin runtime cache revision in
// cluster mode and is a no-op in single-node deployments.
func (s *serviceImpl) markRuntimeCacheChanged(ctx context.Context, reason string) error {
	if s == nil || s.runtimeCacheRevisionCtrl == nil {
		return nil
	}
	revision, err := s.runtimeCacheRevisionCtrl.MarkChanged(ctx)
	if err != nil {
		return err
	}
	if revision > 0 {
		logger.Debugf(ctx, "plugin runtime cache revision bumped reason=%s revision=%d", reason, revision)
	}
	return nil
}
