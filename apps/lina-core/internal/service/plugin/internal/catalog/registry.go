// This file manages the sys_plugin registry table — creating, reading, updating,
// and synchronizing plugin registry rows for both source and dynamic plugins.

package catalog

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// GetRegistry returns the sys_plugin row for the given plugin ID, or nil if not found.
func (s *serviceImpl) GetRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
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
func (s *serviceImpl) ListAllRegistries(ctx context.Context) ([]*entity.SysPlugin, error) {
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.listRegistries(), nil
	}
	return s.listAllRegistriesFromDB(ctx)
}

// SyncManifest creates or updates the registry row for a discovered manifest and
// then synchronizes the release metadata snapshot and node state record.
func (s *serviceImpl) SyncManifest(ctx context.Context, manifest *Manifest) (*entity.SysPlugin, error) {
	installedState := InstalledNo

	existing, err := s.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		stableState := DeriveHostState(installedState, StatusDisabled)
		data := do.SysPlugin{
			PluginId:     manifest.ID,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Type:         manifest.Type,
			Installed:    installedState,
			Status:       StatusDisabled,
			DesiredState: stableState,
			CurrentState: stableState,
			Generation:   int64(1),
			ManifestPath: manifest.ManifestPath,
			Checksum:     s.BuildRegistryChecksum(manifest),
			Remark:       manifest.Description,
		}

		_, err = dao.SysPlugin.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return nil, err
		}
		registry, err := s.refreshStartupRegistry(ctx, manifest.ID)
		if err != nil {
			return nil, err
		}
		if err = s.syncMetadata(ctx, manifest, registry, PluginNodeStateMessageManifestSynchronized); err != nil {
			return nil, err
		}
		return s.syncRegistryReleaseReference(ctx, registry, manifest)
	}

	data := do.SysPlugin{
		Name:   manifest.Name,
		Type:   manifest.Type,
		Remark: manifest.Description,
	}
	if NormalizeType(manifest.Type) == TypeSource {
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.BuildRegistryChecksum(manifest)
		data.Installed = existing.Installed
		if existing.Installed == InstalledYes {
			if strings.TrimSpace(existing.Version) == "" {
				data.Version = manifest.Version
			}
			data.Status = existing.Status
			data.DesiredState = DeriveHostState(existing.Installed, existing.Status)
			data.CurrentState = DeriveHostState(existing.Installed, existing.Status)
			if existing.InstalledAt == nil {
				data.InstalledAt = gtime.Now()
			}
		} else {
			data.Version = manifest.Version
			data.Status = StatusDisabled
			data.DesiredState = DeriveHostState(InstalledNo, StatusDisabled)
			data.CurrentState = DeriveHostState(InstalledNo, StatusDisabled)
		}
		if existing.Generation <= 0 {
			data.Generation = int64(1)
		}
	} else if !ShouldTrackStagedDynamicRelease(existing, manifest) {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.BuildRegistryChecksum(manifest)
		if existing.DesiredState == "" {
			data.DesiredState = DeriveHostState(existing.Installed, existing.Status)
		}
		if existing.CurrentState == "" {
			data.CurrentState = DeriveHostState(existing.Installed, existing.Status)
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

		registry, err = s.refreshStartupRegistry(ctx, manifest.ID)
		if err != nil {
			return nil, err
		}
	}
	if NormalizeType(manifest.Type) == TypeSource &&
		registry != nil &&
		registry.Installed == InstalledYes &&
		strings.TrimSpace(registry.Version) == strings.TrimSpace(manifest.Version) &&
		s.menuSyncer != nil {
		if err = s.menuSyncer.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
			return nil, err
		}
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
	if err = s.syncMetadata(ctx, manifest, registry, PluginNodeStateMessageManifestSynchronized); err != nil {
		return nil, err
	}
	return s.syncRegistryReleaseReference(ctx, registry, manifest)
}

// pluginRegistryDataMatches reports whether a registry row already contains all
// desired non-nil projection fields prepared for a startup manifest sync.
func pluginRegistryDataMatches(existing *entity.SysPlugin, data do.SysPlugin) bool {
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
func pluginRegistryTimeFieldMatches(existing *gtime.Time, desired *gtime.Time) bool {
	if desired == nil {
		return true
	}
	if existing == nil {
		return false
	}
	return existing.String() == desired.String()
}

// SetPluginStatus updates the enabled flag on a plugin registry row and fires the
// matching lifecycle hook event, then syncs release state and node state records.
func (s *serviceImpl) SetPluginStatus(ctx context.Context, pluginID string, enabled int) error {
	registry, err := s.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	installed := InstalledYes
	if registry != nil {
		installed = registry.Installed
	}
	stableState := DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if enabled == StatusEnabled {
		data.EnabledAt = gtime.Now()
	} else {
		data.DisabledAt = gtime.Now()
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}

	if s.hookDispatcher != nil {
		eventName := pluginhost.ExtensionPointPluginDisabled
		if enabled == StatusEnabled {
			eventName = pluginhost.ExtensionPointPluginEnabled
		}
		if err = s.hookDispatcher.DispatchPluginHookEvent(
			ctx,
			eventName,
			pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
				PluginID: pluginID,
				Status:   &enabled,
			}),
		); err != nil {
			return err
		}
	}

	registry, err = s.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}
	if s.releaseStateSyncer != nil {
		if err = s.releaseStateSyncer.SyncPluginReleaseRuntimeState(ctx, registry); err != nil {
			return err
		}
	}
	if s.nodeStateSyncer != nil {
		return s.nodeStateSyncer.SyncPluginNodeState(
			ctx,
			registry.PluginId,
			registry.Version,
			registry.Installed,
			registry.Status,
			PluginNodeStateMessageStatusUpdated,
		)
	}
	return nil
}

// SetPluginInstalled updates the installed flag and derived lifecycle states for one plugin registry row.
func (s *serviceImpl) SetPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	registry, err := s.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	enabled := StatusDisabled
	if registry != nil {
		enabled = registry.Status
	}
	stableState := DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if installed == InstalledYes {
		data.InstalledAt = gtime.Now()
	}
	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	return err
}

// BuildPluginStatusKey returns the display key for a plugin's status record.
func (s *serviceImpl) BuildPluginStatusKey(pluginID string) string {
	return PluginStatusKeyPrefix + pluginID
}

// syncRegistryReleaseReference links the registry row to the matching release row
// when the registry and manifest versions agree.
func (s *serviceImpl) syncRegistryReleaseReference(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *Manifest,
) (*entity.SysPlugin, error) {
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
	return s.refreshStartupRegistry(ctx, registry.PluginId)
}

// SyncRegistryReleaseReference is the exported form of syncRegistryReleaseReference for
// use by runtime-level callers that cannot call the private method directly.
func (s *serviceImpl) SyncRegistryReleaseReference(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *Manifest,
) (*entity.SysPlugin, error) {
	return s.syncRegistryReleaseReference(ctx, registry, manifest)
}

// SyncMetadata orchestrates release metadata, resource reference, and node state
// synchronization after a manifest or lifecycle change. It is the exported form
// used by the runtime package after reconciler state transitions.
func (s *serviceImpl) SyncMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin, message string) error {
	return s.syncMetadata(ctx, manifest, registry, message)
}

// syncMetadata orchestrates release metadata, resource reference, and node state
// synchronization after a manifest or lifecycle change.
func (s *serviceImpl) syncMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin, message string) error {
	if manifest == nil || registry == nil {
		return nil
	}
	if err := s.syncReleaseMetadata(ctx, manifest, registry); err != nil {
		return err
	}
	if registry.Installed == InstalledYes && s.resourceRefSyncer != nil {
		if err := s.resourceRefSyncer.SyncPluginResourceReferences(ctx, manifest); err != nil {
			return err
		}
	}
	if s.nodeStateSyncer != nil {
		return s.nodeStateSyncer.SyncPluginNodeState(ctx, registry.PluginId, registry.Version, registry.Installed, registry.Status, message)
	}
	return nil
}

// shouldDetachDynamicManifestGovernance reports whether metadata sync should
// avoid reattaching release-bound governance for an uninstalled dynamic plugin.
func shouldDetachDynamicManifestGovernance(plugin *entity.SysPlugin) bool {
	if plugin == nil {
		return false
	}
	if NormalizeType(plugin.Type) != TypeDynamic {
		return false
	}
	if plugin.Installed == InstalledYes {
		return false
	}
	return plugin.InstalledAt != nil
}
