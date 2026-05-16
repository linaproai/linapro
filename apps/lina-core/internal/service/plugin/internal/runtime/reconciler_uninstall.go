// This file contains dynamic-plugin uninstall and orphan cleanup reconciliation
// steps.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/pluginhost"
)

// applyUninstall removes live governance, runs uninstall cleanup according to
// the stored uninstall snapshot, and returns the registry to the uninstalled
// stable state.
func (s *serviceImpl) applyUninstall(ctx context.Context, registry *entity.SysPlugin) error {
	manifest, err := s.loadActiveManifest(ctx, registry)
	if err != nil {
		if repairErr := s.repairActiveReleaseBeforeUninstall(ctx, registry); repairErr != nil {
			return gerror.Wrapf(repairErr, "load active dynamic manifest failed before uninstall: %v", err)
		}
		manifest, err = s.loadActiveManifest(ctx, registry)
		if err != nil {
			return err
		}
	}
	if manifest != nil && manifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(ctx, manifest.RuntimeArtifact.Path)
	}
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	purgeStorageData := true
	if release != nil {
		snapshot, parseErr := s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
		if parseErr != nil {
			return parseErr
		}
		if snapshot != nil && snapshot.UninstallPurgeStorageData != nil {
			purgeStorageData = *snapshot.UninstallPurgeStorageData
		}
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			Status:       catalog.StatusDisabled,
			DesiredState: catalog.HostStateUninstalled.String(),
			CurrentState: catalog.HostStateReconciling.String(),
		}).
		Update()
	if err != nil {
		return err
	}
	if purgeStorageData {
		if err = s.executeDynamicUninstallLifecycleCallback(ctx, manifest, purgeStorageData); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
		if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
		if err = wasm.PurgeAuthorizedStoragePaths(ctx, manifest.ID, manifest.HostServices); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
	}
	if err = s.deletePluginMenusByManifest(ctx, manifest); err != nil {
		return s.rollbackReleaseFailure(ctx, registry, 0, err)
	}
	registry, err = s.finalizeState(ctx, registry, manifest, nil, catalog.InstalledNo, catalog.StatusDisabled)
	if err != nil {
		return err
	}
	if release != nil {
		if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.ReleaseStatusUninstalled, ""); err != nil {
			return err
		}
	}
	s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_uninstalled")
	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: manifest.ID}).
		Delete(); err != nil {
		return err
	}
	if err = s.syncNodeProjection(ctx, nodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    0,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Dynamic plugin uninstalled on primary node.",
	}); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_uninstalled"); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, "plugin_uninstalled"); err != nil {
		return err
	}
	return s.dispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUninstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
}

// repairActiveReleaseBeforeUninstall re-archives the same-version staging
// artifact when the active release artifact is missing or corrupt, allowing a
// full uninstall to proceed without falling back to orphan cleanup.
func (s *serviceImpl) repairActiveReleaseBeforeUninstall(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil {
		return gerror.New("dynamic plugin registry cannot be nil")
	}
	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return gerror.Newf("dynamic plugin desired manifest does not exist: %s", registry.PluginId)
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		return gerror.Newf(
			"dynamic plugin active release cannot be repaired from a different staged version: %s active=%s staged=%s",
			registry.PluginId,
			registry.Version,
			desiredManifest.Version,
		)
	}
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("dynamic plugin is missing active release: %s", registry.PluginId)
	}
	archivedPath, err := s.archiveReleaseArtifact(ctx, desiredManifest)
	if err != nil {
		return err
	}
	return s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.BuildReleaseStatus(registry.Installed, registry.Status), archivedPath)
}

// ForceUninstallMissingArtifact clears host-owned governance for a dynamic
// plugin whose staged and archived runtime artifacts are both unavailable. The
// guest uninstall SQL and authorized plugin storage cleanup are intentionally
// skipped because the host cannot load a trusted manifest snapshot.
func (s *serviceImpl) ForceUninstallMissingArtifact(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil {
		return nil
	}
	pluginID := strings.TrimSpace(registry.PluginId)
	if pluginID == "" || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	latest, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if latest != nil {
		registry = latest
	}
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}

	if err = s.deletePluginMenusByManifest(ctx, &catalog.Manifest{ID: pluginID}); err != nil {
		return err
	}
	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Delete(); err != nil {
		return err
	}

	nextGeneration := catalog.NextGeneration(registry)
	stableState := catalog.HostStateUninstalled.String()
	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			Installed:    catalog.InstalledNo,
			Status:       catalog.StatusDisabled,
			DesiredState: stableState,
			CurrentState: stableState,
			Generation:   nextGeneration,
			ReleaseId:    0,
			DisabledAt:   gtime.Now(),
		}).
		Update()
	if err != nil {
		return err
	}
	if release != nil {
		if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.ReleaseStatusUninstalled, ""); err != nil {
			return err
		}
	}

	registry, err = s.catalogSvc.RefreshStartupRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}

	s.invalidateRuntimeCaches(ctx, pluginID, "plugin_orphan_uninstalled")
	if err = s.syncNodeProjection(ctx, nodeProjectionInput{
		PluginID:     pluginID,
		ReleaseID:    0,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Dynamic plugin force-uninstalled with missing artifact; skipped embedded uninstall SQL, Wasm lifecycle hooks, and authorized plugin storage cleanup.",
	}); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_orphan_uninstalled"); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, "plugin_orphan_uninstalled"); err != nil {
		return err
	}
	return s.dispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUninstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Name:     registry.Name,
			Version:  registry.Version,
		}),
	)
}

// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string) error {
	return s.UninstallWithOptions(ctx, pluginID, true)
}

// UninstallWithOptions executes uninstall lifecycle for an installed dynamic
// plugin using one explicit cleanup policy snapshot.
func (s *serviceImpl) UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		return nil
	}
	if catalog.NormalizeType(registry.Type) == catalog.TypeSource {
		return gerror.New("source plugins are compiled into the host and cannot be uninstalled")
	}
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("dynamic plugin is missing active release: %s", pluginID)
	}
	if _, err = s.catalogSvc.PersistReleaseUninstallPurgePolicy(ctx, release, purgeStorageData); err != nil {
		return err
	}
	return s.reconcileDynamicPluginRequest(ctx, pluginID, catalog.HostStateUninstalled)
}

// executeDynamicUninstallLifecycleCallback invokes the plugin-owned dynamic
// uninstall execution phase when cleanup has been requested and declared.
func (s *serviceImpl) executeDynamicUninstallLifecycleCallback(
	ctx context.Context,
	manifest *catalog.Manifest,
	purgeStorageData bool,
) error {
	if manifest == nil || !purgeStorageData {
		return nil
	}
	decision, err := s.RunDynamicLifecycleCallback(ctx, manifest, DynamicLifecycleInput{
		PluginID:          manifest.ID,
		Operation:         pluginhost.LifecycleHookUninstall,
		PurgeStorageData: purgeStorageData,
	})
	if err != nil {
		return err
	}
	if decision != nil && !decision.OK {
		return gerror.New(strings.TrimSpace(decision.Reason))
	}
	return nil
}
