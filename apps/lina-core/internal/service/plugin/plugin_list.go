// This file exposes root-facade list and manifest synchronization methods.

package plugin

import (
	"context"
	"strings"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/management"
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
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	out, err := s.syncAndList(ctx)
	if err != nil {
		return err
	}
	if _, err = s.markRuntimeCacheChanged(ctx, "source_plugins_synced"); err != nil {
		return err
	}
	s.managementListCache.Store(s.managementListCacheKey(ctx), out)
	return nil
}

// SyncSourcePluginsStrict synchronizes source plugins discovered by the
// running host. Tooling is responsible for official submodule preflight before
// plugin-full operations reach the runtime API.
func (s *serviceImpl) SyncSourcePluginsStrict(ctx context.Context) (*ListOutput, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	out, err := s.syncAndList(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.markRuntimeCacheChanged(ctx, "source_plugins_synced"); err != nil {
		return nil, err
	}
	s.managementListCache.Store(s.managementListCacheKey(ctx), out)
	return out, nil
}

// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
// returns the combined list of source and dynamic plugin items.
func (s *serviceImpl) SyncAndList(ctx context.Context) (*ListOutput, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	out, err := s.syncAndList(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.markRuntimeCacheChanged(ctx, "plugins_synced_and_listed"); err != nil {
		return nil, err
	}
	s.managementListCache.Store(s.managementListCacheKey(ctx), out)
	return out, nil
}

// syncAndList scans plugin manifests and mutates plugin governance tables for
// trusted startup or already-guarded platform management paths.
func (s *serviceImpl) syncAndList(ctx context.Context) (*ListOutput, error) {
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
	syncCtx = management.WithManifestSnapshot(syncCtx, manifests)
	syncCtx = s.WithDependencySnapshotCache(syncCtx)

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
	management.SortPluginItems(items)
	if err = s.integrationSvc.RefreshEnabledSnapshot(syncCtx); err != nil {
		return nil, err
	}
	return &ListOutput{List: items, Total: len(items)}, nil
}

// List returns the read-only plugin list with optional in-memory filtering applied.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.managementList(ctx)
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
		if in.Type != "" && !management.MatchesPluginType(item.Type, in.Type) {
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
	out, err := s.managementList(ctx)
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
	return s.buildManagementList(ctx)
}

// PrewarmManagementList builds the complete plugin management list read model
// so the first administrator request can reuse hot discovery and dependency
// projections. Failures are returned to foreground callers and logged by
// asynchronous startup callers.
func (s *serviceImpl) PrewarmManagementList(ctx context.Context) error {
	if _, err := s.managementList(ctx); err != nil {
		return err
	}
	return nil
}

// managementList returns the complete unfiltered plugin management read model.
func (s *serviceImpl) managementList(ctx context.Context) (*ListOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	cacheKey := s.managementListCacheKey(ctx)
	if cached, ok := s.managementListCache.Get(cacheKey); ok {
		return cached, nil
	}
	out, err := s.buildManagementList(ctx)
	if err != nil {
		return nil, err
	}
	s.managementListCache.Store(s.managementListCacheKey(ctx), out)
	return management.CloneListOutput(out), nil
}

// InvalidateManagementListCache clears this process-local read model. Cluster
// peers observe the same plugin-runtime revision and invalidate through the
// root runtime-cache refresh callback.
func (s *serviceImpl) InvalidateManagementListCache(_ context.Context, _ string) {
	if s == nil || s.managementListCache == nil {
		return
	}
	s.managementListCache.Invalidate()
}

// managementListCacheKey returns the current cache partition because plugin
// display metadata is localized during projection and can change when the
// runtime translation bundle version changes.
func (s *serviceImpl) managementListCacheKey(ctx context.Context) management.ListCacheKey {
	if s == nil || s.i18nSvc == nil {
		return management.ListCacheKey{Locale: i18nsvc.DefaultLocale}
	}
	locale := normalizeManagementListCacheLocale(s.i18nSvc.GetLocale(ctx))
	return management.ListCacheKey{
		Locale:               locale,
		RuntimeBundleVersion: s.i18nSvc.BundleVersion(locale),
	}
}

// normalizeManagementListCacheLocale keeps cache keys stable for detached
// startup contexts and tests that do not carry business locale metadata.
func normalizeManagementListCacheLocale(locale string) string {
	if locale == "" {
		return i18nsvc.DefaultLocale
	}
	return locale
}

// buildManagementList scans plugin manifests and projects current registry
// state without synchronizing governance tables.
func (s *serviceImpl) buildManagementList(ctx context.Context) (*ListOutput, error) {
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
	readCtx = management.WithManifestSnapshot(readCtx, manifests)
	readCtx = s.WithDependencySnapshotCache(readCtx)
	registries, err := s.catalogSvc.ListAllRegistries(readCtx)
	if err != nil {
		return nil, err
	}

	registryByPluginID := management.RegistryByPluginID(registries)
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
	management.SortPluginItems(items)
	return &ListOutput{List: items, Total: len(items)}, nil
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
