// This file implements the leader-aware dynamic-plugin reconciler. Management
// APIs persist the desired host state, while the primary node archives the
// staged artifact, performs migrations and menu switches, advances generation,
// and updates per-node convergence rows.

package runtime

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

const runtimeReconcilerInterval = 2 * time.Second

var (
	reconcilerOnce sync.Once
	reconcileMu    sync.Mutex
)

// StartRuntimeReconciler starts the background loop that keeps dynamic-plugin
// desired state, active release, and current-node projection converged.
func (s *serviceImpl) StartRuntimeReconciler(ctx context.Context) {
	if !s.isClusterModeEnabled() {
		return
	}
	reconcilerOnce.Do(func() {
		go s.runReconciler(context.WithoutCancel(ctx))
	})
}

func (s *serviceImpl) runReconciler(ctx context.Context) {
	ticker := time.NewTicker(runtimeReconcilerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ReconcileRuntimePlugins(ctx); err != nil {
				logger.Warningf(ctx, "dynamic plugin reconciler tick failed: %v", err)
			}
		}
	}
}

// ReconcileRuntimePlugins runs one convergence pass. It is safe to call from
// both the background loop and synchronous management flows.
func (s *serviceImpl) ReconcileRuntimePlugins(ctx context.Context) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return err
	}

	isPrimary := s.isPrimaryNode()

	var firstErr error
	for _, registry := range registries {
		if registry == nil {
			continue
		}

		registry, err = s.reconcileRegistryArtifactState(ctx, registry)
		if err != nil {
			logger.Warningf(ctx, "reconcile runtime registry artifact state failed plugin=%s err=%v", registry.PluginId, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if registry == nil {
			continue
		}

		if isPrimary {
			if err = s.reconcilePluginIfNeeded(ctx, registry); err != nil {
				logger.Warningf(ctx, "reconcile dynamic plugin failed plugin=%s err=%v", registry.PluginId, err)
				if firstErr == nil {
					firstErr = err
				}
			}
			refreshedRegistry, getErr := s.catalogSvc.GetRegistry(ctx, registry.PluginId)
			if getErr != nil {
				logger.Warningf(ctx, "reload dynamic plugin registry failed plugin=%s err=%v", registry.PluginId, getErr)
				if firstErr == nil {
					firstErr = getErr
				}
			}
			registry = refreshedRegistry
		}
		if registry == nil {
			continue
		}
		if err = s.reconcileCurrentNodeProjection(ctx, registry); err != nil {
			logger.Warningf(ctx, "reconcile current node projection failed plugin=%s err=%v", registry.PluginId, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *serviceImpl) reconcileDynamicPluginRequest(
	ctx context.Context,
	pluginID string,
	desiredState catalog.HostState,
) error {
	if err := s.updateDesiredState(ctx, pluginID, desiredState); err != nil {
		return err
	}
	if !s.isPrimaryNode() {
		return nil
	}
	return s.ReconcileRuntimePlugins(ctx)
}

// reconcilePluginIfNeeded selects the smallest convergence action for the current
// registry row: install, upgrade, same-version refresh, state toggle, or uninstall.
func (s *serviceImpl) reconcilePluginIfNeeded(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = catalog.BuildStableHostState(registry)
	}
	stableState := catalog.BuildStableHostState(registry)
	if desiredState == catalog.HostStateUninstalled.String() {
		if registry.Installed != catalog.InstalledYes {
			return nil
		}
		return s.applyUninstall(ctx, registry)
	}

	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return gerror.New("动态插件目标清单不存在")
	}

	if registry.Installed != catalog.InstalledYes {
		return s.applyInstall(ctx, registry, desiredManifest, desiredState)
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		// Version drift means upgrade semantics, including upgrade SQL and release switch.
		return s.applyUpgrade(ctx, registry, desiredManifest, desiredState)
	}
	if s.shouldRefreshInstalledRelease(ctx, registry, desiredManifest) {
		// Same semantic version can still require refresh when the staged artifact,
		// archive bytes, or synthesized checksum changed after a rebuild.
		return s.applyRefresh(ctx, registry, desiredManifest, desiredState)
	}
	if desiredState != stableState {
		return s.applyStateToggle(ctx, registry, desiredManifest, desiredState)
	}
	return nil
}

func (s *serviceImpl) reconcileCurrentNodeProjection(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	if registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled && registry.ReleaseId > 0 {
		manifest, err := s.loadActiveManifest(ctx, registry)
		if err != nil {
			return s.syncNodeProjection(ctx, nodeProjectionInput{
				PluginID:     registry.PluginId,
				ReleaseID:    registry.ReleaseId,
				DesiredState: registry.DesiredState,
				CurrentState: catalog.NodeStateFailed.String(),
				Generation:   registry.Generation,
				Message:      err.Error(),
			})
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.syncNodeProjection(ctx, nodeProjectionInput{
					PluginID:     registry.PluginId,
					ReleaseID:    registry.ReleaseId,
					DesiredState: registry.DesiredState,
					CurrentState: catalog.NodeStateFailed.String(),
					Generation:   registry.Generation,
					Message:      err.Error(),
				})
			}
		}
	}

	return s.syncNodeProjection(ctx, nodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    registry.ReleaseId,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Current node converged to host plugin generation.",
	})
}

// applyInstall performs the first activation of a discovered dynamic plugin,
// including artifact archive, SQL install, permission/menu projection, optional
// frontend bundle preparation, and registry finalization.
func (s *serviceImpl) applyInstall(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}
	if err = s.markReconciling(ctx, registry, catalog.HostState(desiredState)); err != nil {
		return err
	}

	archivedPath, err := s.archiveReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if desiredState == catalog.HostStateEnabled.String() {
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
			}
		}
	}

	enabled := catalog.StatusDisabled
	if desiredState == catalog.HostStateEnabled.String() {
		enabled = catalog.StatusEnabled
	}
	registry, err = s.finalizeState(ctx, registry, manifest, release, catalog.InstalledYes, enabled)
	if err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.BuildReleaseStatus(catalog.InstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release installed on primary node."); err != nil {
		return err
	}
	if err = s.dispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	); err != nil {
		return err
	}
	if enabled == catalog.StatusEnabled {
		return s.dispatchHookEvent(
			ctx,
			pluginhost.ExtensionPointPluginEnabled,
			pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
				PluginID: manifest.ID,
				Name:     manifest.Name,
				Version:  manifest.Version,
				Status:   &enabled,
			}),
		)
	}
	return nil
}

// applyUpgrade moves an installed plugin to a new semantic version. Unlike
// refresh, this path runs upgrade SQL and may replace the active release.
func (s *serviceImpl) applyUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	activeManifest, err := s.loadActiveManifest(ctx, registry)
	if err != nil {
		return err
	}
	// Invalidate the Wasm module cache for the previous active artifact before
	// replacing it so subsequent requests compile from the new artifact.
	if activeManifest != nil && activeManifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(activeManifest.RuntimeArtifact.Path)
	}
	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}

	if err = s.markReconciling(ctx, registry, catalog.HostState(desiredState)); err != nil {
		return err
	}
	archivedPath, err := s.archiveReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUpgrade); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if desiredState == catalog.HostStateEnabled.String() {
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
			}
		}
	}

	enabled := catalog.StatusDisabled
	if desiredState == catalog.HostStateEnabled.String() {
		enabled = catalog.StatusEnabled
	}
	previousReleaseID := registry.ReleaseId
	registry, err = s.finalizeState(ctx, registry, manifest, release, catalog.InstalledYes, enabled)
	if err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if previousReleaseID > 0 && previousReleaseID != release.Id {
		if err = s.catalogSvc.UpdateReleaseState(ctx, previousReleaseID, catalog.ReleaseStatusInstalled, ""); err != nil {
			return err
		}
	}
	if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.BuildReleaseStatus(catalog.InstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	return s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release upgraded on primary node.")
}

// applyStateToggle flips enable/disable status for the current active release
// without changing the installed version or artifact archive.
func (s *serviceImpl) applyStateToggle(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if err = s.markReconciling(ctx, registry, catalog.HostState(desiredState)); err != nil {
		return err
	}

	enabled := catalog.StatusDisabled
	eventName := pluginhost.ExtensionPointPluginDisabled
	if desiredState == catalog.HostStateEnabled.String() {
		enabled = catalog.StatusEnabled
		eventName = pluginhost.ExtensionPointPluginEnabled
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, 0, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackReleaseFailure(ctx, registry, 0, err)
			}
		}
	}

	registry, err = s.finalizeState(ctx, registry, manifest, release, catalog.InstalledYes, enabled)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, 0, err)
	}
	if release != nil {
		if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.BuildReleaseStatus(catalog.InstalledYes, enabled), ""); err != nil {
			return err
		}
	}
	if enabled == catalog.StatusDisabled {
		s.invalidateFrontendBundle(ctx, manifest.ID, "plugin_disabled")
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin status converged on primary node."); err != nil {
		return err
	}
	return s.dispatchHookEvent(
		ctx,
		eventName,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
			Status:   &enabled,
		}),
	)
}

// applyRefresh reapplies host projections for the same semantic version when
// the artifact checksum or archived bytes changed. It intentionally skips
// upgrade SQL because the version contract did not advance.
func (s *serviceImpl) applyRefresh(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}
	if err = s.markReconciling(ctx, registry, catalog.HostState(desiredState)); err != nil {
		return err
	}

	// Invalidate any previously cached compiled module so the refreshed artifact
	// is recompiled on next bridge invocation.
	if manifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(manifest.RuntimeArtifact.Path)
	}
	archivedPath, err := s.archiveReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}

	enabled := catalog.StatusDisabled
	if desiredState == catalog.HostStateEnabled.String() {
		enabled = catalog.StatusEnabled
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
			}
		}
	}

	registry, err = s.finalizeState(ctx, registry, manifest, release, catalog.InstalledYes, enabled)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.catalogSvc.UpdateReleaseState(ctx, release.Id, catalog.BuildReleaseStatus(catalog.InstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	return s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release refreshed on primary node.")
}

func (s *serviceImpl) applyUninstall(ctx context.Context, registry *entity.SysPlugin) error {
	manifest, err := s.loadActiveManifest(ctx, registry)
	if err != nil {
		return err
	}
	if manifest != nil && manifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(manifest.RuntimeArtifact.Path)
	}
	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
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
	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
		return s.rollbackReleaseFailure(ctx, registry, 0, err)
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
	s.invalidateFrontendBundle(ctx, manifest.ID, "plugin_uninstalled")
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

func (s *serviceImpl) rollbackInstallOrUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
	restoreManifest *catalog.Manifest,
	failedManifest *catalog.Manifest,
	failedReleaseID int,
	reconcileErr error,
) error {
	if failedManifest != nil {
		if rollbackErr := s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, failedManifest, catalog.MigrationDirectionRollback); rollbackErr != nil {
			logger.Warningf(ctx, "rollback dynamic plugin SQL failed plugin=%s err=%v", failedManifest.ID, rollbackErr)
		}
		if menuErr := s.deletePluginMenusByManifest(ctx, failedManifest); menuErr != nil {
			logger.Warningf(ctx, "delete dynamic plugin menus during rollback failed plugin=%s err=%v", failedManifest.ID, menuErr)
		}
	}
	if restoreManifest != nil {
		if restoreErr := s.syncPluginMenusAndPermissions(ctx, restoreManifest); restoreErr != nil {
			logger.Warningf(ctx, "restore previous plugin menus and permissions failed plugin=%s err=%v", restoreManifest.ID, restoreErr)
		}
	}
	return s.rollbackReleaseFailure(ctx, registry, failedReleaseID, reconcileErr)
}

func (s *serviceImpl) rollbackReleaseFailure(
	ctx context.Context,
	registry *entity.SysPlugin,
	releaseID int,
	reconcileErr error,
) error {
	restoredRegistry := registry
	if releaseID > 0 {
		if updateErr := s.catalogSvc.UpdateReleaseState(ctx, releaseID, catalog.ReleaseStatusFailed, ""); updateErr != nil {
			logger.Warningf(ctx, "mark dynamic plugin release failed failed plugin=%s releaseID=%d err=%v", registry.PluginId, releaseID, updateErr)
		}
	}
	if restored, err := s.restoreStableState(ctx, registry); err == nil && restored != nil {
		restoredRegistry = restored
	}
	if restoredRegistry != nil {
		if syncErr := s.syncNodeProjection(ctx, nodeProjectionInput{
			PluginID:     restoredRegistry.PluginId,
			ReleaseID:    restoredRegistry.ReleaseId,
			DesiredState: restoredRegistry.DesiredState,
			CurrentState: catalog.NodeStateFailed.String(),
			Generation:   restoredRegistry.Generation,
			Message:      reconcileErr.Error(),
		}); syncErr != nil {
			logger.Warningf(ctx, "sync failed-node projection failed plugin=%s err=%v", restoredRegistry.PluginId, syncErr)
		}
	}
	return reconcileErr
}

// shouldRefreshInstalledRelease decides whether an already installed dynamic release
// should be re-converged even though the semantic version did not change. It compares
// desired checksum, registry checksum, and archived release content.
func (s *serviceImpl) shouldRefreshInstalledRelease(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return false
	}
	if registry.Installed != catalog.InstalledYes {
		return false
	}
	if strings.TrimSpace(registry.Checksum) == "" {
		return true
	}
	desiredChecksum := strings.TrimSpace(s.catalogSvc.BuildRegistryChecksum(manifest))
	if desiredChecksum == "" {
		return true
	}
	if desiredChecksum != strings.TrimSpace(registry.Checksum) {
		return true
	}

	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		return true
	}
	packagePath, err := s.resolveReleasePackagePath(ctx, release)
	if err != nil {
		return true
	}
	archivedManifest, err := s.catalogSvc.LoadManifestFromArtifactPath(packagePath)
	if err != nil || archivedManifest == nil {
		return true
	}
	return strings.TrimSpace(s.catalogSvc.BuildRegistryChecksum(archivedManifest)) != desiredChecksum
}

// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return gerror.New("源码插件随宿主编译集成，不支持卸载")
	}

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		return nil
	}
	return s.reconcileDynamicPluginRequest(ctx, pluginID, catalog.HostStateUninstalled)
}
