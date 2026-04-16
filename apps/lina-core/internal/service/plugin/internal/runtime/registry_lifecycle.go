// This file centralizes sys_plugin state mutations used by lifecycle actions
// and the background reconciler so generation, release, and stable state fields
// stay consistent across install, enable, disable, upgrade, and rollback flows.

package runtime

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// updateDesiredState records the management intent that the primary reconciler
// should eventually converge to.
func (s *serviceImpl) updateDesiredState(
	ctx context.Context,
	pluginID string,
	desiredState catalog.HostState,
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
	registry *entity.SysPlugin,
	desiredState catalog.HostState,
) error {
	if registry == nil {
		return nil
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			DesiredState: desiredState.String(),
			CurrentState: catalog.HostStateReconciling.String(),
		}).
		Update()
	return err
}

// finalizeState applies one stable lifecycle state together with the release
// pointer and next generation number after a successful switch.
func (s *serviceImpl) finalizeState(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
	installed int,
	enabled int,
) (*entity.SysPlugin, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := catalog.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
		Generation:   catalog.NextGeneration(registry),
	}
	if manifest != nil {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.catalogSvc.BuildRegistryChecksum(manifest)
	}
	if release != nil {
		data.ReleaseId = release.Id
	}
	if installed == catalog.InstalledYes {
		if registry.Installed != catalog.InstalledYes {
			data.InstalledAt = gtime.Now()
		}
		if enabled == catalog.StatusEnabled {
			data.EnabledAt = gtime.Now()
		} else {
			data.DisabledAt = gtime.Now()
		}
	} else {
		data.Status = catalog.StatusDisabled
		data.ReleaseId = 0
		data.DisabledAt = gtime.Now()
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}
	return s.catalogSvc.GetRegistry(ctx, registry.PluginId)
}

// restoreStableState clears a transient reconcile marker and rewrites
// desired/current state back to the stable registry flags.
func (s *serviceImpl) restoreStableState(
	ctx context.Context,
	registry *entity.SysPlugin,
) (*entity.SysPlugin, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := catalog.BuildStableHostState(registry)
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
	return s.catalogSvc.GetRegistry(ctx, registry.PluginId)
}
