// This file exposes root-facade list and manifest synchronization methods.

package plugin

import (
	"context"
	"sort"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
)

// WithStartupDataSnapshot returns a child context carrying catalog and
// integration startup snapshots for one host startup orchestration.
func (s *serviceImpl) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	startupCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
	startupCtx, err = s.integrationSvc.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		return ctx, err
	}
	return startupCtx, nil
}

// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
func (s *serviceImpl) SyncSourcePlugins(ctx context.Context) error {
	_, err := s.SyncAndList(ctx)
	return err
}

// SyncSourcePluginsStrict synchronizes source plugins discovered by the
// running host. Tooling is responsible for official submodule preflight before
// plugin-full operations reach the runtime API.
func (s *serviceImpl) SyncSourcePluginsStrict(ctx context.Context) (*ListOutput, error) {
	return s.SyncAndList(ctx)
}

// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
// returns the combined list of source and dynamic plugin items.
func (s *serviceImpl) SyncAndList(ctx context.Context) (*ListOutput, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	startupstats.Add(ctx, startupstats.CounterPluginScans, 1)
	startupstats.Add(ctx, startupstats.CounterPluginScanItems, len(manifests))

	syncCtx, err := s.WithStartupDataSnapshot(ctx)
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
		items = append(items, s.buildServicePluginItem(syncCtx, s.runtimeSvc.BuildPluginItem(syncCtx, manifest, registry)))
	}

	runtimeItems, err := s.runtimeSvc.BuildRuntimeItems(syncCtx, covered)
	if err != nil {
		return nil, err
	}
	items = append(items, s.buildServicePluginItems(syncCtx, runtimeItems)...)
	sortServicePluginItems(items)
	if err = s.integrationSvc.RefreshEnabledSnapshot(syncCtx); err != nil {
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

// Get returns one read-only plugin detail projection by exact plugin ID.
func (s *serviceImpl) Get(ctx context.Context, pluginID string) (*PluginItem, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	out, err := s.ReadOnlyList(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range out.List {
		if item != nil && item.Id == normalizedPluginID {
			return item, nil
		}
	}
	return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
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
	startupstats.Add(ctx, startupstats.CounterPluginScans, 1)
	startupstats.Add(ctx, startupstats.CounterPluginScanItems, len(manifests))

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
		if item := s.buildServicePluginItem(readCtx, s.runtimeSvc.BuildPluginItem(readCtx, manifest, registryByPluginID[manifest.ID])); item != nil {
			items = append(items, item)
		}
	}

	runtimeItems, err := s.runtimeSvc.BuildRuntimeItemsReadOnly(readCtx, covered)
	if err != nil {
		return nil, err
	}
	items = append(items, s.buildServicePluginItems(readCtx, runtimeItems)...)
	sortServicePluginItems(items)
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

// buildServicePluginItems wraps runtime projections with facade-level metadata.
func (s *serviceImpl) buildServicePluginItems(ctx context.Context, items []*runtime.PluginItem) []*PluginItem {
	out := make([]*PluginItem, 0, len(items))
	for _, item := range items {
		if wrapped := s.buildServicePluginItem(ctx, item); wrapped != nil {
			out = append(out, wrapped)
		}
	}
	return out
}

// buildServicePluginItem wraps one runtime projection and attaches dependency status.
func (s *serviceImpl) buildServicePluginItem(ctx context.Context, item *runtime.PluginItem) *PluginItem {
	if item == nil {
		return nil
	}
	out := &PluginItem{PluginItem: *item}
	if dependencyCheck, err := s.CheckPluginDependencies(ctx, item.Id); err == nil {
		out.DependencyCheck = dependencyCheck
	}
	return out
}

// sortServicePluginItems sorts facade plugin projections by plugin ID.
func sortServicePluginItems(items []*PluginItem) {
	sort.Slice(items, func(i int, j int) bool {
		if items[i] == nil {
			return false
		}
		if items[j] == nil {
			return true
		}
		return items[i].Id < items[j].Id
	})
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
