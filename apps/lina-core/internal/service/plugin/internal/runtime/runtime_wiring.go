// This file contains nil-safe runtime provider adapters used by lifecycle and
// reconciliation flows.

package runtime

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/plugin/pluginhost"
)

// isClusterModeEnabled is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isClusterModeEnabled() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsEnabled()
}

// isPrimaryNode is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) isPrimaryNode() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsPrimary()
}

// currentNodeID is a nil-safe wrapper around the topology provider.
func (s *serviceImpl) currentNodeID() string {
	if s.topology == nil {
		return ""
	}
	return s.topology.NodeID()
}

// dispatchHookEvent is a nil-safe wrapper for hook event dispatch.
func (s *serviceImpl) dispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if s.integrationSvc == nil {
		return nil
	}
	return s.integrationSvc.DispatchPluginHookEvent(ctx, event, values)
}

// syncPluginMenusAndPermissions is a nil-safe wrapper for menu synchronization.
func (s *serviceImpl) syncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if s.integrationSvc == nil {
		return nil
	}
	return s.integrationSvc.SyncPluginMenusAndPermissions(ctx, manifest)
}

// syncPluginResourceReferences is a nil-safe wrapper for governance resource references.
func (s *serviceImpl) syncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error {
	if s.integrationSvc == nil {
		return nil
	}
	return s.integrationSvc.SyncPluginResourceReferences(ctx, manifest)
}

// deletePluginMenusByManifest is a nil-safe wrapper for menu deletion.
func (s *serviceImpl) deletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if s.integrationSvc == nil {
		return nil
	}
	return s.integrationSvc.DeletePluginMenusByManifest(ctx, manifest)
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
func (s *serviceImpl) invalidateRuntimeCaches(ctx context.Context, manifest *catalog.Manifest, reason runtimeChangeReason) {
	var pluginID string
	if manifest != nil {
		pluginID = manifest.ID
		if manifest.RuntimeArtifact != nil && s.wasmRuntime != nil {
			s.wasmRuntime.InvalidateCache(ctx, manifest.RuntimeArtifact.Path)
		}
	}
	if s.frontendSvc != nil {
		s.frontendSvc.InvalidateBundle(ctx, pluginID, string(reason))
	}
	if s.i18nSvc != nil {
		s.i18nSvc.InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
			Sectors:         []i18nsvc.Sector{i18nsvc.SectorDynamicPlugin},
			DynamicPluginID: pluginID,
		})
	}
}

// notifyRuntimeCacheChanged publishes a successful dynamic runtime mutation for
// one plugin to other cluster nodes through the root plugin facade.
func (s *serviceImpl) notifyRuntimeCacheChanged(
	ctx context.Context,
	manifest *catalog.Manifest,
	reason runtimeChangeReason,
) error {
	if s.cacheChangeNotifier == nil {
		return nil
	}
	pluginID := ""
	if manifest != nil {
		pluginID = manifest.ID
	}
	return s.cacheChangeNotifier.PublishPluginChange(ctx, pluginID, pluginv1.PluginTypeDynamic.String(), string(reason))
}
