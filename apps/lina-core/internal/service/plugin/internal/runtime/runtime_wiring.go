// This file contains runtime service dependency wiring and nil-safe provider
// adapters used by lifecycle and reconciliation flows.

package runtime

import (
	"context"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/internal/service/session"
	"lina-core/pkg/pluginhost"
)

// SetTopology wires the cluster topology provider.
func (s *serviceImpl) SetTopology(t TopologyProvider) {
	s.topology = t
	s.configureReconcilerRevisionController()
}

// SetMenuManager wires the menu synchronization provider.
func (s *serviceImpl) SetMenuManager(m MenuManager) {
	s.menuMgr = m
}

// SetHookDispatcher wires the lifecycle hook dispatcher.
func (s *serviceImpl) SetHookDispatcher(d HookDispatcher) {
	s.hookDispatcher = d
}

// SetJwtConfigProvider wires the JWT configuration provider for route token validation.
func (s *serviceImpl) SetJwtConfigProvider(p JwtConfigProvider) {
	s.jwtConfig = p
}

// SetUploadSizeProvider wires the upload-size provider for dynamic package uploads.
func (s *serviceImpl) SetUploadSizeProvider(p UploadSizeProvider) {
	s.uploadSize = p
}

// SetUserContextSetter wires the user-context injection provider.
func (s *serviceImpl) SetUserContextSetter(p UserContextSetter) {
	s.userCtx = p
}

// SetSessionStore wires the online-session store used for dynamic route requests.
func (s *serviceImpl) SetSessionStore(store session.Store) {
	if store != nil {
		s.sessionStore = store
	}
}

// SetPermissionMenuFilter wires the plugin-level permission menu filter.
func (s *serviceImpl) SetPermissionMenuFilter(f PermissionMenuFilter) {
	s.menuFilter = f
}

// SetRuntimeCacheChangeNotifier wires cluster cache revision publication.
func (s *serviceImpl) SetRuntimeCacheChangeNotifier(n CacheChangeNotifier) {
	s.cacheChangeNotifier = n
}

// SetDependencyValidator wires release dependency validation.
func (s *serviceImpl) SetDependencyValidator(v DependencyValidator) {
	s.dependencyValidator = v
}

// isClusterModeEnabled is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isClusterModeEnabled() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsClusterModeEnabled()
}

// isPrimaryNode is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isPrimaryNode() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsPrimaryNode()
}

// currentNodeID is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) currentNodeID() string {
	if s.topology == nil {
		return ""
	}
	return s.topology.CurrentNodeID()
}

// dispatchHookEvent is a nil-safe wrapper for hook event dispatch.
func (s *serviceImpl) dispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if s.hookDispatcher == nil {
		return nil
	}
	return s.hookDispatcher.DispatchPluginHookEvent(ctx, event, values)
}

// syncPluginMenusAndPermissions is a nil-safe wrapper for menu synchronization.
func (s *serviceImpl) syncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenusAndPermissions(ctx, manifest)
}

// syncPluginMenus is a nil-safe wrapper for partial menu synchronization (rollback path).
func (s *serviceImpl) syncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenus(ctx, manifest)
}

// deletePluginMenusByManifest is a nil-safe wrapper for menu deletion.
func (s *serviceImpl) deletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.DeletePluginMenusByManifest(ctx, manifest)
}

// ensureFrontendBundle delegates to frontendSvc to guarantee an in-memory bundle exists.
func (s *serviceImpl) ensureFrontendBundle(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.EnsureBundle(ctx, manifest)
}

// validateFrontendMenuBindings delegates frontend menu binding validation.
func (s *serviceImpl) validateFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.ValidateRuntimeFrontendMenuBindings(ctx, manifest)
}

// invalidateRuntimeCaches removes cached runtime frontend assets and runtime i18n
// bundles after one plugin lifecycle change. Only the dynamic-plugin sector for
// the affected plugin is invalidated; host and source-plugin sectors stay hot
// for unrelated locales and plugins.
func (s *serviceImpl) invalidateRuntimeCaches(ctx context.Context, pluginID string, reason string) {
	wasm.InvalidateAllCache(ctx)
	if s.frontendSvc != nil {
		s.frontendSvc.InvalidateBundle(ctx, pluginID, reason)
	}
	if s.i18nSvc != nil {
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors:         []i18nsvc.Sector{i18nsvc.SectorDynamicPlugin},
			DynamicPluginID: pluginID,
		})
	}
}

// notifyRuntimeCacheChanged publishes a successful dynamic runtime mutation to
// other cluster nodes through the root plugin facade.
func (s *serviceImpl) notifyRuntimeCacheChanged(ctx context.Context, reason string) error {
	if s.cacheChangeNotifier == nil {
		return nil
	}
	return s.cacheChangeNotifier.MarkRuntimeCacheChanged(ctx, reason)
}
