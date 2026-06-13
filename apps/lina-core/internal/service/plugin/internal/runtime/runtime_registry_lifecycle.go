// This file centralizes sys_plugin state mutations used by lifecycle actions
// and the background reconciler so generation, release, and stable state fields
// stay consistent across install, enable, disable, upgrade, and rollback flows.

package runtime

import (
	"context"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
)

// updateDesiredState records the management intent that the primary reconciler
// should eventually converge to.
func (s *serviceImpl) updateDesiredState(
	ctx context.Context,
	pluginID string,
	desiredState plugintypes.HostState,
) error {
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{DesiredState: desiredState.String()}).
		Update()
	return err
}

// markReconciling marks the host row as entering a transient reconciliation
// window while keeping the requested desired state persisted.
func (s *serviceImpl) markReconciling(
	ctx context.Context,
	registry *store.PluginRecord,
	desiredState plugintypes.HostState,
) error {
	if registry == nil {
		return nil
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			DesiredState: desiredState.String(),
			CurrentState: plugintypes.HostStateReconciling.String(),
		}).
		Update()
	return err
}

// finalizeState applies one stable lifecycle state together with the release
// pointer and next generation number after a successful switch.
func (s *serviceImpl) finalizeState(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	release *store.ReleaseRecord,
	installed int,
	enabled int,
) (*store.PluginRecord, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := plugintypes.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
		Generation:   store.NextGeneration(registry),
	}
	if manifest != nil {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.catalogSvc.BuildRegistryChecksum(manifest)
	}
	if release != nil {
		data.ReleaseId = release.Id
	}
	if installed == plugintypes.InstalledYes {
		if registry.Installed != plugintypes.InstalledYes {
			data.InstalledAt = timePtr(time.Now())
		}
		if enabled == plugintypes.StatusEnabled {
			data.EnabledAt = timePtr(time.Now())
		} else {
			data.DisabledAt = timePtr(time.Now())
		}
	} else {
		data.Status = plugintypes.StatusDisabled
		data.ReleaseId = 0
		data.DisabledAt = timePtr(time.Now())
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}
	return s.storeSvc.RefreshStartupRegistry(ctx, registry.PluginId)
}

// restoreStableState clears a transient reconcile marker and rewrites
// desired/current state back to the stable registry flags.
func (s *serviceImpl) restoreStableState(
	ctx context.Context,
	registry *store.PluginRecord,
) (*store.PluginRecord, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := store.BuildStableHostState(registry)
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			Installed:    registry.Installed,
			Status:       registry.Status,
			DesiredState: stableState,
			CurrentState: stableState,
			ReleaseId:    registry.ReleaseId,
		}).
		Update()
	if err != nil {
		return nil, err
	}
	return s.storeSvc.RefreshStartupRegistry(ctx, registry.PluginId)
}
