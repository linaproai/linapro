// This file manages the sys_plugin registry table — creating, reading, updating,
// and synchronizing plugin registry rows for both source and dynamic plugins.

package store

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/dialect"
)

// GetRegistry returns the sys_plugin row for the given plugin ID, or nil if not found.
func (s *serviceImpl) GetRegistry(ctx context.Context, pluginID string) (*PluginRecord, error) {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return nil, nil
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.registry(normalizedID), nil
	}

	return s.getRegistryFromDB(ctx, normalizedID)
}

// ListAllRegistries returns all sys_plugin rows ordered by plugin_id.
func (s *serviceImpl) ListAllRegistries(ctx context.Context) ([]*PluginRecord, error) {
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.listRegistries(), nil
	}
	return s.listAllRegistriesFromDB(ctx)
}

// SyncManifest creates or updates the registry row for a discovered manifest and
// then synchronizes the release metadata snapshot and node state record.
func (s *serviceImpl) SyncManifest(ctx context.Context, manifest *catalog.Manifest) (*PluginRecord, error) {
	installedState := plugintypes.InstalledNo

	existing, err := s.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		stableState := plugintypes.DeriveHostState(installedState, plugintypes.StatusDisabled)
		data := do.SysPlugin{
			PluginId:     manifest.ID,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Type:         manifest.Type,
			Installed:    installedState,
			Status:       plugintypes.StatusDisabled,
			DesiredState: stableState,
			CurrentState: stableState,
			Generation:   int64(1),
			ManifestPath: manifest.ManifestPath,
			Checksum:     s.BuildRegistryChecksum(manifest),
			Remark:       manifest.Description,
			ScopeNature:  plugintypes.NormalizeScopeNature(manifest.ScopeNature).String(),
			InstallMode:  plugintypes.NormalizeInstallMode(manifest.DefaultInstallMode).String(),
		}

		insertID, insertErr := dao.SysPlugin.Ctx(ctx).Data(data).InsertAndGetId()
		err = insertErr
		if err != nil {
			if !dialect.IsUniqueConstraintViolation(err) {
				return nil, err
			}
			registry, refreshErr := s.refreshStartupRegistry(ctx, manifest.ID)
			if refreshErr != nil {
				return nil, err
			}
			if registry == nil {
				return nil, err
			}
			existing = registry
			goto existingRegistry
		}
		startupstats.Add(ctx, startupstats.CounterPluginSyncChanged, 1)
		registry := insertStartupRegistry(ctx, int(insertID), data)
		if registry == nil {
			registry, err = s.refreshStartupRegistry(ctx, manifest.ID)
			if err != nil {
				return nil, err
			}
		}
		if err = s.syncReleaseMetadata(ctx, manifest, registry); err != nil {
			return nil, err
		}
		registry, err = s.syncRegistryReleaseReference(ctx, registry, manifest)
		if err != nil {
			return nil, err
		}
		return registry, nil
	}

existingRegistry:
	data := do.SysPlugin{
		Name:        manifest.Name,
		Type:        manifest.Type,
		Remark:      manifest.Description,
		ScopeNature: plugintypes.NormalizeScopeNature(manifest.ScopeNature).String(),
		InstallMode: plugintypes.NormalizeInstallMode(manifest.DefaultInstallMode).String(),
	}
	if plugintypes.NormalizeType(manifest.Type) == plugintypes.TypeSource {
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.BuildRegistryChecksum(manifest)
		data.Installed = existing.Installed
		if existing.Installed == plugintypes.InstalledYes {
			if strings.TrimSpace(existing.Version) == "" {
				data.Version = manifest.Version
			}
			data.Status = existing.Status
			data.DesiredState = plugintypes.DeriveHostState(existing.Installed, existing.Status)
			data.CurrentState = plugintypes.DeriveHostState(existing.Installed, existing.Status)
			if existing.InstalledAt == nil {
				data.InstalledAt = timePtr(time.Now())
			}
		} else {
			data.Version = manifest.Version
			data.Status = plugintypes.StatusDisabled
			data.DesiredState = plugintypes.DeriveHostState(plugintypes.InstalledNo, plugintypes.StatusDisabled)
			data.CurrentState = plugintypes.DeriveHostState(plugintypes.InstalledNo, plugintypes.StatusDisabled)
		}
		if existing.Generation <= 0 {
			data.Generation = int64(1)
		}
	} else if !ShouldTrackStagedDynamicRelease(existing, manifest) {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.BuildRegistryChecksum(manifest)
		if existing.DesiredState == "" {
			data.DesiredState = plugintypes.DeriveHostState(existing.Installed, existing.Status)
		}
		if existing.CurrentState == "" {
			data.CurrentState = plugintypes.DeriveHostState(existing.Installed, existing.Status)
		}
		if existing.Generation <= 0 {
			data.Generation = int64(1)
		}
	} else {
		data.ManifestPath = existing.ManifestPath
		data.Checksum = existing.Checksum
	}

	registry := existing
	if !pluginRegistryDataMatches(existing, data) {
		_, err = dao.SysPlugin.Ctx(ctx).
			Where(do.SysPlugin{PluginId: manifest.ID}).
			Data(data).
			Update()
		if err != nil {
			return nil, err
		}
		startupstats.Add(ctx, startupstats.CounterPluginSyncChanged, 1)

		if updated := updateStartupRegistry(ctx, manifest.ID, data); updated != nil {
			registry = updated
		} else {
			registry, err = s.refreshStartupRegistry(ctx, manifest.ID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		startupstats.Add(ctx, startupstats.CounterPluginSyncNoop, 1)
	}
	// After a dynamic plugin has been uninstalled once, later workspace scans
	// should keep its staged release snapshot up to date but must not restore
	// active release bindings or governance projections until it is installed again.
	if shouldDetachDynamicManifestGovernance(registry) {
		if err = s.syncReleaseMetadata(ctx, manifest, registry); err != nil {
			return nil, err
		}
		return registry, nil
	}
	if err = s.syncReleaseMetadata(ctx, manifest, registry); err != nil {
		return nil, err
	}
	registry, err = s.syncRegistryReleaseReference(ctx, registry, manifest)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// pluginRegistryDataMatches reports whether a registry row already contains all
// desired non-nil projection fields prepared for a startup manifest sync.
func pluginRegistryDataMatches(existing *PluginRecord, data do.SysPlugin) bool {
	if existing == nil {
		return false
	}
	return pluginRegistryFieldMatches(existing.Name, data.Name) &&
		pluginRegistryFieldMatches(existing.Version, data.Version) &&
		pluginRegistryFieldMatches(existing.Type, data.Type) &&
		pluginRegistryFieldMatches(existing.Installed, data.Installed) &&
		pluginRegistryFieldMatches(existing.Status, data.Status) &&
		pluginRegistryFieldMatches(existing.DesiredState, data.DesiredState) &&
		pluginRegistryFieldMatches(existing.CurrentState, data.CurrentState) &&
		pluginRegistryFieldMatches(existing.Generation, data.Generation) &&
		pluginRegistryFieldMatches(existing.ManifestPath, data.ManifestPath) &&
		pluginRegistryFieldMatches(existing.Checksum, data.Checksum) &&
		pluginRegistryFieldMatches(existing.Remark, data.Remark) &&
		pluginRegistryFieldMatches(existing.ScopeNature, data.ScopeNature) &&
		pluginRegistryFieldMatches(existing.InstallMode, data.InstallMode) &&
		pluginRegistryTimeFieldMatches(existing.InstalledAt, data.InstalledAt)
}

// pluginRegistryFieldMatches treats nil DO fields as omitted updates and compares
// non-nil fields using GoFrame's conversion semantics.
func pluginRegistryFieldMatches(existing any, desired any) bool {
	if desired == nil {
		return true
	}
	return gconv.String(existing) == gconv.String(desired)
}

// pluginRegistryTimeFieldMatches treats nil time DO fields as omitted updates.
func pluginRegistryTimeFieldMatches(existing *time.Time, desired *time.Time) bool {
	if desired == nil {
		return true
	}
	if existing == nil {
		return false
	}
	return existing.Equal(*desired)
}

// SetPluginStatus updates the enabled flag on a plugin registry row and fires the
// matching lifecycle hook event, then syncs release state and node state records.
func (s *serviceImpl) SetPluginStatus(ctx context.Context, pluginID string, enabled int) error {
	registry, err := s.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	installed := plugintypes.InstalledYes
	if registry != nil {
		installed = registry.Installed
	}
	stableState := plugintypes.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if enabled == plugintypes.StatusEnabled {
		data.EnabledAt = timePtr(time.Now())
	} else {
		data.DisabledAt = timePtr(time.Now())
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	if updated := updateStartupRegistry(ctx, pluginID, data); updated != nil {
		registry = updated
	}

	if registry == nil || startupDataSnapshotFromContext(ctx) == nil {
		registry, err = s.GetRegistry(ctx, pluginID)
		if err != nil {
			return err
		}
	}
	if registry == nil {
		return nil
	}
	return nil
}

// SetPluginInstalled updates the installed flag and derived lifecycle states for one plugin registry row.
func (s *serviceImpl) SetPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	registry, err := s.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	enabled := plugintypes.StatusDisabled
	if registry != nil {
		enabled = registry.Status
	}
	stableState := plugintypes.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if installed == plugintypes.InstalledYes {
		data.InstalledAt = timePtr(time.Now())
	}
	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err == nil {
		updateStartupRegistry(ctx, pluginID, data)
	}
	return err
}

// SetRegistryRuntimeState updates runtime state fields without changing the
// stable desired lifecycle state such as installed or enabled.
func (s *serviceImpl) SetRegistryRuntimeState(ctx context.Context, pluginID string, patch RuntimeStatePatch) error {
	data := do.SysPlugin{}
	if strings.TrimSpace(patch.DesiredState) != "" {
		data.DesiredState = strings.TrimSpace(patch.DesiredState)
	}
	if strings.TrimSpace(patch.CurrentState) != "" {
		data.CurrentState = strings.TrimSpace(patch.CurrentState)
	}
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	_, err = s.RefreshStartupRegistry(ctx, pluginID)
	return err
}

// SetAutoEnableForNewTenants updates the platform-owned tenant provisioning policy.
func (s *serviceImpl) SetAutoEnableForNewTenants(ctx context.Context, pluginID string, enabled bool) error {
	data := do.SysPlugin{
		AutoEnableForNewTenants: enabled,
	}
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err == nil {
		updateStartupRegistry(ctx, pluginID, data)
	}
	return err
}

// BuildPluginStatusKey returns the display key for a plugin's status record.
func (s *serviceImpl) BuildPluginStatusKey(pluginID string) string {
	return plugintypes.PluginStatusKeyPrefix + pluginID
}

// syncRegistryReleaseReference links the registry row to the matching release row
// when the registry and manifest versions agree.
func (s *serviceImpl) syncRegistryReleaseReference(
	ctx context.Context,
	registry *PluginRecord,
	manifest *catalog.Manifest,
) (*PluginRecord, error) {
	if registry == nil || manifest == nil {
		return registry, nil
	}
	if strings.TrimSpace(registry.Version) != strings.TrimSpace(manifest.Version) {
		return registry, nil
	}

	release, err := s.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil || registry.ReleaseId == release.Id {
		return registry, nil
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{ReleaseId: release.Id}).
		Update()
	if err != nil {
		return nil, err
	}
	if updated := updateStartupRegistry(ctx, registry.PluginId, do.SysPlugin{ReleaseId: release.Id}); updated != nil {
		return updated, nil
	}
	return s.refreshStartupRegistry(ctx, registry.PluginId)
}

// SyncRegistryReleaseReference is the exported form of syncRegistryReleaseReference for
// use by runtime-level callers that cannot call the private method directly.
func (s *serviceImpl) SyncRegistryReleaseReference(
	ctx context.Context,
	registry *PluginRecord,
	manifest *catalog.Manifest,
) (*PluginRecord, error) {
	return s.syncRegistryReleaseReference(ctx, registry, manifest)
}

// SyncMetadata orchestrates release metadata, resource reference, and node state
// synchronization after a manifest or lifecycle change. It is the exported form
// used by the runtime package after reconciler state transitions.
func (s *serviceImpl) SyncMetadata(ctx context.Context, manifest *catalog.Manifest, registry *PluginRecord, message string) error {
	return s.syncMetadata(ctx, manifest, registry, message)
}

// syncMetadata orchestrates release metadata, resource reference, and node state
// synchronization after a manifest or lifecycle change.
func (s *serviceImpl) syncMetadata(ctx context.Context, manifest *catalog.Manifest, registry *PluginRecord, message string) error {
	if manifest == nil || registry == nil {
		return nil
	}
	_ = message
	return s.syncReleaseMetadata(ctx, manifest, registry)
}

// shouldDetachDynamicManifestGovernance reports whether metadata sync should
// avoid reattaching release-bound governance for an uninstalled dynamic plugin.
func shouldDetachDynamicManifestGovernance(plugin *PluginRecord) bool {
	if plugin == nil {
		return false
	}
	if plugintypes.NormalizeType(plugin.Type) != plugintypes.TypeDynamic {
		return false
	}
	if plugin.Installed == plugintypes.InstalledYes {
		return false
	}
	return plugin.InstalledAt != nil
}
