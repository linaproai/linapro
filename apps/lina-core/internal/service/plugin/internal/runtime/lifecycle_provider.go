// This file exposes public lifecycle provider methods so the lifecycle sub-package
// can trigger reconciliation and artifact checks without importing the runtime package.

package runtime

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// ReconcileDynamicPluginRequest implements lifecycle.ReconcileProvider.
// It submits a desired-state transition to the reconciler loop.
func (s *serviceImpl) ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string) error {
	return s.reconcileDynamicPluginRequest(ctx, pluginID, catalog.HostState(desiredState))
}

// EnsureRuntimeArtifactAvailable implements lifecycle.ReconcileProvider.
// It verifies the WASM artifact is present for the given lifecycle action label.
func (s *serviceImpl) EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error {
	return s.ensureArtifactAvailable(manifest, actionLabel)
}

// ShouldRefreshInstalledDynamicRelease implements lifecycle.ReconcileProvider.
// It type-asserts registry to *entity.SysPlugin then delegates to the private helper.
func (s *serviceImpl) ShouldRefreshInstalledDynamicRelease(
	ctx context.Context,
	registry interface{},
	manifest *catalog.Manifest,
) bool {
	reg, ok := registry.(*entity.SysPlugin)
	if !ok {
		return false
	}
	return s.shouldRefreshInstalledRelease(ctx, reg, manifest)
}

// BuildPluginItem returns a PluginItem projection for one manifest + registry pair.
// Used by the plugin facade SyncAndList coordination method.
func (s *serviceImpl) BuildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *entity.SysPlugin) *PluginItem {
	return s.buildPluginItem(ctx, manifest, registry)
}

// BuildRuntimeItems returns PluginItems for dynamic plugins present in the registry
// but absent from the given manifest map. Used by the plugin facade SyncAndList.
func (s *serviceImpl) BuildRuntimeItems(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error) {
	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*PluginItem, 0)
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		if _, ok := covered[registry.PluginId]; ok {
			continue
		}
		registry, err = s.reconcileRegistryArtifactState(ctx, registry)
		if err != nil {
			return nil, err
		}
		if item := s.buildPluginItem(ctx, nil, registry); item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// CheckIsInstalled reports whether a plugin is installed after reconciling artifact state.
// Used by the plugin facade UpdateStatus guard.
func (s *serviceImpl) CheckIsInstalled(ctx context.Context, pluginID string) (bool, error) {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return false, err
	}
	if registry == nil {
		return false, nil
	}
	registry, err = s.reconcileRegistryArtifactState(ctx, registry)
	if err != nil {
		return false, err
	}
	if registry == nil {
		return false, nil
	}
	return registry.Installed == catalog.InstalledYes, nil
}
