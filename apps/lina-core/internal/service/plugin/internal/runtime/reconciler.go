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

// runtimeReconcilerInterval is the clustered background reconcile cadence.
const runtimeReconcilerInterval = 2 * time.Second

// Background reconciler singletons ensure only one reconcile loop and one
// convergence pass run at a time inside the current process.
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

// runReconciler executes the periodic background convergence loop used by
// clustered deployments.
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
		if err = s.reconcileRuntimeRegistry(ctx, registry, isPrimary); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// reconcileDynamicPluginRequest records the requested target state and lets the
// primary node converge the addressed plugin immediately.
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
	return s.reconcileRuntimePlugin(ctx, pluginID)
}

// reconcileRuntimePlugin converges one target plugin synchronously for
// management requests. Unlike the background full scan, it must not fail a
// user-triggered install/refresh because some unrelated staged dynamic plugin is
// temporarily broken in the shared registry during other tests or uploads.
func (s *serviceImpl) reconcileRuntimePlugin(ctx context.Context, pluginID string) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return gerror.New("plugin does not exist")
	}
	return s.reconcileRuntimeRegistry(ctx, registry, true)
}

// reconcileRuntimeRegistry converges one runtime registry row, optionally
// performing primary-only lifecycle work before updating current-node state.
func (s *serviceImpl) reconcileRuntimeRegistry(
	ctx context.Context,
	registry *entity.SysPlugin,
	isPrimary bool,
) error {
	if registry == nil {
		return nil
	}

	pluginID := registry.PluginId

	// Refresh the registry against current artifact presence before any lifecycle
	// action so missing or newly restored packages are reflected consistently.
	refreshedRegistry, err := s.reconcileRegistryArtifactState(ctx, registry)
	if err != nil {
		logger.Warningf(ctx, "reconcile runtime registry artifact state failed plugin=%s err=%v", pluginID, err)
		return err
	}
	if refreshedRegistry == nil {
		return nil
	}
	registry = refreshedRegistry

	if isPrimary {
		// Only the primary node mutates shared lifecycle state such as release
		// activation, migrations, and desired/current host states.
		if err = s.reconcilePluginIfNeeded(ctx, registry); err != nil {
			logger.Warningf(ctx, "reconcile dynamic plugin failed plugin=%s err=%v", pluginID, err)
			return err
		}
		// Reload after lifecycle work so node projection sees the latest release
		// binding, generation, and stable host state.
		registry, err = s.catalogSvc.GetRegistry(ctx, registry.PluginId)
		if err != nil {
			logger.Warningf(ctx, "reload dynamic plugin registry failed plugin=%s err=%v", pluginID, err)
			return err
		}
	}
	if registry == nil {
		return nil
	}
	if err = s.reconcileCurrentNodeProjection(ctx, registry); err != nil {
		logger.Warningf(ctx, "reconcile current node projection failed plugin=%s err=%v", pluginID, err)
		return err
	}
	return nil
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
		return gerror.New("dynamic plugin desired manifest does not exist")
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

// reconcileCurrentNodeProjection verifies the current node can serve the active
// dynamic plugin state and then persists the node-local convergence snapshot.
func (s *serviceImpl) reconcileCurrentNodeProjection(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil
	}

	// Enabled dynamic plugins must prove their active manifest and optional
	// frontend bundle still load on this node before we mark the node converged.
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
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
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
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_installed")
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
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
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
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_upgraded")
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
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_disabled")
	} else {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_enabled")
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
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
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
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_refreshed")
	}
	return s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release refreshed on primary node.")
}

// applyUninstall removes live governance, runs uninstall cleanup according to
// the stored uninstall snapshot, and returns the registry to the uninstalled
// stable state.
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

// rollbackInstallOrUpgrade attempts to restore the last stable plugin state
// after install, upgrade, or refresh work fails midway through reconciliation.
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

// rollbackReleaseFailure marks the release and node projection as failed when a
// reconciliation error cannot be hidden behind a fully restored stable state.
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
	return s.UninstallWithOptions(ctx, pluginID, true)
}

// UninstallWithOptions executes uninstall lifecycle for an installed dynamic
// plugin using one explicit cleanup policy snapshot.
func (s *serviceImpl) UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return gerror.New("source plugins are compiled into the host and cannot be uninstalled")
	}

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		return nil
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
