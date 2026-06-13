// This file contains dynamic-plugin uninstall and orphan cleanup reconciliation
// steps.

package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginhost"
)

// applyUninstall removes live governance, runs uninstall cleanup according to
// the stored uninstall snapshot, and returns the registry to the uninstalled
// stable state.
func (s *serviceImpl) applyUninstall(ctx context.Context, registry *store.PluginRecord) error {
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
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	purgeStorageData := true
	if release != nil {
		snapshot, parseErr := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
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
			Status:       plugintypes.StatusDisabled,
			DesiredState: plugintypes.HostStateUninstalled.String(),
			CurrentState: plugintypes.HostStateReconciling.String(),
		}).
		Update()
	if err != nil {
		return err
	}
	if purgeStorageData {
		if err = s.executeDynamicUninstallLifecycleCallback(ctx, manifest, purgeStorageData); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
		if err = s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionUninstall); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
		storageSvc := s.storageCleanupServiceForPlugin(manifest.ID)
		if err = wasm.PurgeAuthorizedStoragePaths(ctx, storageSvc, manifest.HostServices); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
	}
	if err = s.deletePluginMenusByManifest(ctx, manifest); err != nil {
		return s.rollbackReleaseFailure(ctx, registry, 0, err)
	}
	registry, err = s.finalizeState(ctx, registry, manifest, nil, plugintypes.InstalledNo, plugintypes.StatusDisabled)
	if err != nil {
		return err
	}
	if release != nil {
		if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.ReleaseStatusUninstalled, ""); err != nil {
			return err
		}
	}
	s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginUninstalled)
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
	if err = s.notifyRuntimeCacheChanged(ctx, manifest, runtimeChangeReasonPluginUninstalled); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginUninstalled); err != nil {
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

// storageCleanupServiceForPlugin resolves the plugin-scoped storage service
// used by dynamic uninstall cleanup.
func (s *serviceImpl) storageCleanupServiceForPlugin(pluginID string) storagecap.Service {
	if s == nil || s.storageCleanupServices == nil {
		return nil
	}
	services := capability.ServicesForPlugin(s.storageCleanupServices.StorageCleanupServices(), pluginID)
	if services == nil {
		return nil
	}
	return services.Storage()
}

// repairActiveReleaseBeforeUninstall re-archives the same-version staging
// artifact when the active release artifact is missing or corrupt, allowing a
// full uninstall to proceed without falling back to orphan cleanup.
func (s *serviceImpl) repairActiveReleaseBeforeUninstall(ctx context.Context, registry *store.PluginRecord) error {
	if registry == nil {
		return gerror.New("dynamic plugin registry cannot be nil")
	}
	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || plugintypes.NormalizeType(desiredManifest.Type) != plugintypes.TypeDynamic {
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
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
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
	return s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.BuildReleaseStatus(registry.Installed, registry.Status), archivedPath)
}

// ForceUninstallMissingArtifact clears host-owned governance for a dynamic
// plugin whose staged and archived runtime artifacts are both unavailable. The
// guest uninstall SQL and authorized plugin storage cleanup are intentionally
// skipped because the host cannot load a trusted manifest snapshot.
func (s *serviceImpl) ForceUninstallMissingArtifact(ctx context.Context, registry *store.PluginRecord) error {
	if registry == nil {
		return nil
	}
	pluginID := strings.TrimSpace(registry.PluginId)
	if pluginID == "" || plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic {
		return nil
	}

	latest, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if latest != nil {
		registry = latest
	}
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
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

	nextGeneration := store.NextGeneration(registry)
	stableState := plugintypes.HostStateUninstalled.String()
	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			Installed:    plugintypes.InstalledNo,
			Status:       plugintypes.StatusDisabled,
			DesiredState: stableState,
			CurrentState: stableState,
			Generation:   nextGeneration,
			ReleaseId:    0,
			DisabledAt:   timePtr(time.Now()),
		}).
		Update()
	if err != nil {
		return err
	}
	if release != nil {
		if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.ReleaseStatusUninstalled, ""); err != nil {
			return err
		}
	}

	registry, err = s.storeSvc.RefreshStartupRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}

	s.invalidateRuntimeCaches(ctx, &catalog.Manifest{ID: pluginID}, runtimeChangeReasonPluginOrphanUninstalled)
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
	if err = s.notifyRuntimeCacheChanged(ctx, &catalog.Manifest{ID: pluginID}, runtimeChangeReasonPluginOrphanUninstalled); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginOrphanUninstalled); err != nil {
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
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		return nil
	}
	if plugintypes.NormalizeType(registry.Type) == plugintypes.TypeSource {
		return gerror.New("source plugins are compiled into the host and cannot be uninstalled")
	}
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("dynamic plugin is missing active release: %s", pluginID)
	}
	if _, err = s.storeSvc.PersistReleaseUninstallPurgePolicy(ctx, release, purgeStorageData); err != nil {
		return err
	}
	return s.reconcileDynamicPluginRequest(ctx, pluginID, plugintypes.HostStateUninstalled, DynamicReconcileOptions{})
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
		PluginID:         manifest.ID,
		Operation:        pluginhost.LifecycleHookUninstall,
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
