// This file resolves plugin enablement, tenant-specific plugin state, and the
// process-local business-entry snapshot used by source integration guards.

package integration

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginhost"
)

// isEnabled reports whether the plugin with the given ID is currently enabled.
func (r *filterRuntime) isEnabled(pluginID string) bool {
	if r == nil {
		return false
	}
	return r.enabledByID[strings.TrimSpace(pluginID)]
}

// IsEnabled reports whether the plugin with the given ID can expose business entries.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	manifest, _ := s.catalogSvc.GetDesiredManifest(pluginID)
	enabled, err := s.registryBusinessEntryEnabledForTenant(ctx, registry, manifest)
	return err == nil && enabled
}

// SetTenantPluginEnabledState persists one tenant-scoped plugin enablement row.
func (s *serviceImpl) SetTenantPluginEnabledState(ctx context.Context, pluginID string, tenantID int, enabled bool) error {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}
	identity := do.SysPluginState{
		PluginId: normalizedPluginID,
		TenantId: tenantID,
		StateKey: pluginTenantEnablementStateKey,
	}
	return dao.SysPluginState.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		_, err := dao.SysPluginState.Ctx(ctx).Data(do.SysPluginState{
			PluginId:   identity.PluginId,
			TenantId:   identity.TenantId,
			StateKey:   identity.StateKey,
			StateValue: pluginTenantEnablementStateValue(enabled),
			Enabled:    enabled,
		}).InsertIgnore()
		if err != nil {
			return err
		}
		_, err = dao.SysPluginState.Ctx(ctx).
			Where(identity).
			Data(do.SysPluginState{
				StateValue: pluginTenantEnablementStateValue(enabled),
				Enabled:    enabled,
			}).
			Update()
		return err
	})
}

// registryEnabledForTenant resolves the effective plugin enablement state for
// the current tenant context.
func (s *serviceImpl) registryEnabledForTenant(ctx context.Context, registry *entity.SysPlugin) (bool, error) {
	if registry == nil ||
		registry.Installed != catalog.InstalledYes ||
		registry.Status != catalog.StatusEnabled {
		return false, nil
	}
	tenantID := datascope.CurrentTenantID(ctx)
	// Platform-only describes tenant governance visibility, not runtime
	// availability. Global platform-only plugins such as multi-tenant can still
	// publish tenant-context APIs and permissions while tenant administrators
	// remain unable to control them through the tenant plugin list.
	if catalog.NormalizeInstallMode(registry.InstallMode) != catalog.InstallModeTenantScoped || tenantID == datascope.PlatformTenantID {
		return true, nil
	}
	return s.tenantPluginEnabled(ctx, registry.PluginId, tenantID)
}

// registryBusinessEntryEnabledForTenant resolves plugin enablement and runtime
// upgrade state before allowing plugin-owned routes, menus, cron jobs, or hooks.
func (s *serviceImpl) registryBusinessEntryEnabledForTenant(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
) (bool, error) {
	enabled, err := s.registryEnabledForTenant(ctx, registry)
	if err != nil || !enabled {
		return enabled, err
	}
	state, err := s.catalogSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return false, err
	}
	return catalog.RuntimeStateAllowsBusinessEntry(state.State), nil
}

// tenantPluginEnabled reads one tenant-scoped plugin enablement row.
func (s *serviceImpl) tenantPluginEnabled(ctx context.Context, pluginID string, tenantID int) (bool, error) {
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: strings.TrimSpace(pluginID),
			TenantId: tenantID,
			StateKey: pluginTenantEnablementStateKey,
		}).
		Value(dao.SysPluginState.Columns().Enabled)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return false, nil
	}
	return value.Bool(), nil
}

// pluginTenantEnablementStateValue converts one enablement flag to the stable
// state_value payload used for diagnostics and manual inspection.
func pluginTenantEnablementStateValue(enabled bool) string {
	if enabled {
		return pluginTenantEnabledValue
	}
	return pluginTenantDisabledValue
}

// RefreshEnabledSnapshot rebuilds the in-memory business-entry snapshot used by runtime guards.
func (s *serviceImpl) RefreshEnabledSnapshot(ctx context.Context) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}
	enabledByID, err := s.buildEnabledPluginMapFromCatalog(ctx, manifests, false)
	if err != nil {
		return err
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot = enabledByID
	s.sharedState.enabledSnapshotLoaded = true
	return nil
}

// SetPluginEnabledState updates one plugin entry in the in-memory business-entry snapshot.
func (s *serviceImpl) SetPluginEnabledState(pluginID string, enabled bool) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot[normalizedPluginID] = enabled
	s.sharedState.enabledSnapshotLoaded = true
}

// DeletePluginEnabledState removes one plugin entry from the in-memory business-entry snapshot.
func (s *serviceImpl) DeletePluginEnabledState(pluginID string) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	delete(s.sharedState.enabledSnapshot, normalizedPluginID)
	s.sharedState.enabledSnapshotLoaded = true
}

// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
func (s *serviceImpl) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	s.sharedState.sourceRouteBindingsMu.RLock()
	defer s.sharedState.sourceRouteBindingsMu.RUnlock()

	items := make([]pluginhost.SourceRouteBinding, 0)
	for _, bindings := range s.sharedState.sourceRouteBindings {
		items = append(items, pluginhost.CloneSourceRouteBindings(bindings)...)
	}
	return items
}

// buildFilterRuntime builds a filter runtime by scanning all manifests and
// loading whether each discovered plugin can expose business entries.
func (s *serviceImpl) buildFilterRuntime(ctx context.Context) (*filterRuntime, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	return s.buildFilterRuntimeFromManifests(ctx, manifests)
}

// buildFilterRuntimeFromManifests builds a filter runtime for the given manifest list.
func (s *serviceImpl) buildFilterRuntimeFromManifests(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (*filterRuntime, error) {
	enabledByID, err := s.buildEnabledPluginMap(ctx, manifests)
	if err != nil {
		return nil, err
	}
	return &filterRuntime{
		manifests:   manifests,
		enabledByID: enabledByID,
	}, nil
}

// buildEnabledPluginMap queries whether each plugin can expose business entries.
func (s *serviceImpl) buildEnabledPluginMap(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (map[string]bool, error) {
	return s.buildEnabledPluginMapFromCatalog(ctx, manifests, true)
}

// buildEnabledPluginMapFromCatalog queries or reuses business-entry state for
// the supplied manifests. Refresh callers can disable snapshot reuse to rebuild
// the process-wide view after lifecycle changes.
func (s *serviceImpl) buildEnabledPluginMapFromCatalog(
	ctx context.Context,
	manifests []*catalog.Manifest,
	allowLoadedSnapshot bool,
) (map[string]bool, error) {
	var (
		enabledByID = make(map[string]bool, len(manifests))
		pluginIDs   = make([]string, 0, len(manifests))
	)
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		if _, ok := enabledByID[pluginID]; ok {
			continue
		}
		enabledByID[pluginID] = false
		pluginIDs = append(pluginIDs, pluginID)
	}
	if len(pluginIDs) == 0 {
		return enabledByID, nil
	}
	if allowLoadedSnapshot &&
		datascope.CurrentTenantID(ctx) == datascope.PlatformTenantID &&
		s.applyLoadedEnabledSnapshot(enabledByID) {
		return enabledByID, nil
	}

	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	manifestByID := manifestByPluginID(manifests)

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if _, ok := enabledByID[pluginID]; !ok {
			continue
		}
		enabled, err := s.registryBusinessEntryEnabledForTenant(ctx, registry, manifestByID[pluginID])
		if err != nil {
			return nil, err
		}
		enabledByID[pluginID] = enabled
	}
	s.storeLoadedEnabledSnapshot(enabledByID)
	return enabledByID, nil
}

// storeLoadedEnabledSnapshot refreshes the process-local business-entry snapshot
// from one registry read so later filters in the same process can reuse it.
func (s *serviceImpl) storeLoadedEnabledSnapshot(enabledByID map[string]bool) {
	if s == nil || s.sharedState == nil {
		return
	}
	snapshot := make(map[string]bool, len(enabledByID))
	for pluginID, enabled := range enabledByID {
		normalizedPluginID := strings.TrimSpace(pluginID)
		if normalizedPluginID != "" {
			snapshot[normalizedPluginID] = enabled
		}
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot = snapshot
	s.sharedState.enabledSnapshotLoaded = true
}

// applyLoadedEnabledSnapshot copies the process-local business-entry snapshot into
// the requested plugin map when a lifecycle path has already warmed it.
func (s *serviceImpl) applyLoadedEnabledSnapshot(enabledByID map[string]bool) bool {
	if s == nil || s.sharedState == nil || len(enabledByID) == 0 {
		return false
	}
	s.sharedState.enabledSnapshotMu.RLock()
	defer s.sharedState.enabledSnapshotMu.RUnlock()
	if !s.sharedState.enabledSnapshotLoaded {
		return false
	}
	for pluginID := range enabledByID {
		enabledByID[pluginID] = s.sharedState.enabledSnapshot[pluginID]
	}
	return true
}

// manifestByPluginID indexes discovered manifests by plugin ID.
func manifestByPluginID(manifests []*catalog.Manifest) map[string]*catalog.Manifest {
	result := make(map[string]*catalog.Manifest, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		result[strings.TrimSpace(manifest.ID)] = manifest
	}
	return result
}

// buildBackgroundEnabledChecker returns a PluginEnabledChecker for use in source plugin
// route and cron registrars that need to guard runtime access.
func (s *serviceImpl) buildBackgroundEnabledChecker() pluginhost.PluginEnabledChecker {
	return func(ctx context.Context, pluginID string) bool {
		normalizedPluginID := strings.TrimSpace(pluginID)
		if normalizedPluginID == "" {
			return false
		}
		if ctx == nil {
			ctx = context.Background()
		}
		if datascope.CurrentTenantID(ctx) != datascope.PlatformTenantID {
			return s.IsEnabled(ctx, normalizedPluginID)
		}

		s.sharedState.enabledSnapshotMu.RLock()
		enabled, ok := s.sharedState.enabledSnapshot[normalizedPluginID]
		loaded := s.sharedState.enabledSnapshotLoaded
		s.sharedState.enabledSnapshotMu.RUnlock()
		if ok || loaded {
			return enabled
		}
		return s.IsEnabled(ctx, normalizedPluginID)
	}
}

// buildPrimaryNodeChecker returns a PrimaryNodeChecker for use in source plugin cron registrars.
func (s *serviceImpl) buildPrimaryNodeChecker() pluginhost.PrimaryNodeChecker {
	return func() bool {
		if s.topology == nil {
			return false
		}
		return s.topology.IsPrimaryNode()
	}
}

// setSourceRouteBindings stores the latest host-captured route bindings for one
// source plugin after registration completes.
func (s *serviceImpl) setSourceRouteBindings(pluginID string, bindings []pluginhost.SourceRouteBinding) {
	s.sharedState.sourceRouteBindingsMu.Lock()
	defer s.sharedState.sourceRouteBindingsMu.Unlock()
	s.sharedState.sourceRouteBindings[strings.TrimSpace(pluginID)] = pluginhost.CloneSourceRouteBindings(bindings)
}
