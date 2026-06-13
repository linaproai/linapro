// This file exposes public lifecycle provider methods so the lifecycle sub-package
// can trigger reconciliation and artifact checks without importing the runtime package.

package runtime

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
)

// ReconcileDynamicPluginRequest implements lifecycle.RuntimeOrchestrator. It
// submits a desired-state transition to the reconciler loop with explicit
// operation options.
func (s *serviceImpl) ReconcileDynamicPluginRequest(
	ctx context.Context,
	pluginID string,
	desiredState string,
	options DynamicReconcileOptions,
) error {
	return s.reconcileDynamicPluginRequest(ctx, pluginID, plugintypes.HostState(desiredState), options)
}

// EnsureRuntimeArtifactAvailable implements lifecycle.RuntimeOrchestrator.
// It verifies the WASM artifact is present for the given lifecycle action label.
func (s *serviceImpl) EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error {
	return s.ensureArtifactAvailable(manifest, actionLabel)
}

// ShouldRefreshInstalledDynamicRelease implements lifecycle.RuntimeOrchestrator.
// It type-asserts registry to *store.PluginRecord then delegates to the private helper.
func (s *serviceImpl) ShouldRefreshInstalledDynamicRelease(
	ctx context.Context,
	registry interface{},
	manifest *catalog.Manifest,
) bool {
	reg, ok := registry.(*store.PluginRecord)
	if !ok {
		return false
	}
	return s.shouldRefreshInstalledRelease(ctx, reg, manifest)
}

// BuildPluginItem returns a PluginItem projection for one manifest + registry pair.
// Used by the plugin facade SyncAndList coordination method.
func (s *serviceImpl) BuildPluginItem(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem {
	return s.buildPluginItem(ctx, manifest, registry)
}

// BuildPluginSummaryItem returns the lightweight management-list projection for
// one manifest + registry pair.
func (s *serviceImpl) BuildPluginSummaryItem(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem {
	return s.buildPluginSummaryItem(ctx, manifest, registry)
}

// BuildPluginItemReadOnly returns one detail projection without mutating
// governance state when a dynamic artifact is missing from storage.
func (s *serviceImpl) BuildPluginItemReadOnly(ctx context.Context, manifest *catalog.Manifest, registry *store.PluginRecord) *PluginItem {
	if manifest == nil {
		registry = s.projectRegistryArtifactState(ctx, registry)
	}
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

// BuildRuntimeSummaryItemsReadOnly returns lightweight dynamic PluginItems
// without mutating governance state when artifacts are missing from runtime storage.
func (s *serviceImpl) BuildRuntimeSummaryItemsReadOnly(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error) {
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
		registry = s.projectRegistryArtifactState(ctx, registry)
		if item := s.buildPluginSummaryItem(ctx, nil, registry); item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// BuildRuntimeItemsReadOnly returns dynamic PluginItems without mutating
// governance state when an artifact is missing from runtime storage.
func (s *serviceImpl) BuildRuntimeItemsReadOnly(ctx context.Context, covered map[string]struct{}) ([]*PluginItem, error) {
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
		registry = s.projectRegistryArtifactState(ctx, registry)
		if item := s.buildPluginItem(ctx, nil, registry); item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// CheckIsInstalled reports whether a plugin is installed after reconciling artifact state.
// Used by the plugin facade UpdateStatus guard.
func (s *serviceImpl) CheckIsInstalled(ctx context.Context, pluginID string) (bool, error) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
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
	return registry.Installed == plugintypes.InstalledYes, nil
}
