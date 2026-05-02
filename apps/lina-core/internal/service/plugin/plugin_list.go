// This file exposes root-facade list and manifest synchronization methods.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
)

// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
func (s *serviceImpl) SyncSourcePlugins(ctx context.Context) error {
	_, err := s.SyncAndList(ctx)
	return err
}

// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
// returns the combined list of source and dynamic plugin items.
func (s *serviceImpl) SyncAndList(ctx context.Context) (*ListOutput, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	rootCtx := ctx
	syncCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	syncCtx, err = s.integrationSvc.WithStartupDataSnapshot(syncCtx)
	if err != nil {
		return nil, err
	}

	covered := make(map[string]struct{}, len(manifests))
	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		covered[manifest.ID] = struct{}{}
		registry, syncErr := s.catalogSvc.SyncManifest(syncCtx, manifest)
		if syncErr != nil {
			return nil, syncErr
		}
		items = append(items, s.runtimeSvc.BuildPluginItem(syncCtx, manifest, registry))
	}

	runtimeItems, err := s.runtimeSvc.BuildRuntimeItems(rootCtx, covered)
	if err != nil {
		return nil, err
	}
	items = append(items, runtimeItems...)
	runtime.SortPluginItems(items)
	if err = s.integrationSvc.RefreshEnabledSnapshot(rootCtx); err != nil {
		return nil, err
	}
	return &ListOutput{List: items, Total: len(items)}, nil
}

// List returns the read-only plugin list with optional in-memory filtering applied.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.ReadOnlyList(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]*PluginItem, 0, len(out.List))
	for _, item := range out.List {
		if in.ID != "" && !strings.Contains(item.Id, in.ID) {
			continue
		}
		if in.Name != "" && !strings.Contains(item.Name, in.Name) {
			continue
		}
		if in.Type != "" && !matchPluginType(item.Type, in.Type) {
			continue
		}
		if in.Status != nil && item.Enabled != *in.Status {
			continue
		}
		if in.Installed != nil && item.Installed != *in.Installed {
			continue
		}
		filtered = append(filtered, item)
	}
	return &ListOutput{List: filtered, Total: len(filtered)}, nil
}

// ReadOnlyList scans plugin manifests and projects current registry state
// without synchronizing governance tables.
func (s *serviceImpl) ReadOnlyList(ctx context.Context) (*ListOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	readCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	registries, err := s.catalogSvc.ListAllRegistries(readCtx)
	if err != nil {
		return nil, err
	}

	registryByPluginID := buildRegistryByPluginID(registries)
	covered := make(map[string]struct{}, len(manifests))
	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		covered[manifest.ID] = struct{}{}
		if item := s.runtimeSvc.BuildPluginItem(readCtx, manifest, registryByPluginID[manifest.ID]); item != nil {
			items = append(items, item)
		}
	}

	runtimeItems, err := s.runtimeSvc.BuildRuntimeItemsReadOnly(readCtx, covered)
	if err != nil {
		return nil, err
	}
	items = append(items, runtimeItems...)
	runtime.SortPluginItems(items)
	return &ListOutput{List: items, Total: len(items)}, nil
}

// buildRegistryByPluginID indexes registry rows by plugin ID for read-only list projection.
func buildRegistryByPluginID(registries []*entity.SysPlugin) map[string]*entity.SysPlugin {
	result := make(map[string]*entity.SysPlugin, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		result[registry.PluginId] = registry
	}
	return result
}

// ListEnabledPluginIDs returns the IDs of plugins that are currently
// installed and enabled.
func (s *serviceImpl) ListEnabledPluginIDs(ctx context.Context) ([]string, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}

	pluginIDs := make([]string, 0, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
			continue
		}
		pluginIDs = append(pluginIDs, strings.TrimSpace(registry.PluginId))
	}
	return pluginIDs, nil
}

// matchPluginType compares normalized plugin types so list filtering accepts
// user input that differs only by case or alias formatting.
func matchPluginType(actual string, expected string) bool {
	actualType := catalog.NormalizeType(actual)
	expectedType := catalog.NormalizeType(expected)
	if expectedType == "" {
		return true
	}
	return actualType == expectedType
}
