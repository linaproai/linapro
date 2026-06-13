// This file contains dynamic-plugin upgrade, refresh, and enablement state
// convergence steps.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// applyUpgrade moves an installed plugin to a new semantic version. Unlike
// refresh, this path runs upgrade SQL and may replace the active release.
func (s *serviceImpl) applyUpgrade(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	if err := s.validateCandidateDependencies(ctx, manifest); err != nil {
		return err
	}
	activeManifest, err := s.loadActiveManifest(ctx, registry)
	if err != nil {
		return err
	}
	// Invalidate the Wasm module cache for the previous active artifact before
	// replacing it so subsequent requests compile from the new artifact.
	if activeManifest != nil && activeManifest.RuntimeArtifact != nil {
		if s.wasmRuntime != nil {
			s.wasmRuntime.InvalidateCache(ctx, activeManifest.RuntimeArtifact.Path)
		}
	}
	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
	}

	if err = s.markReconciling(ctx, registry, plugintypes.HostState(desiredState)); err != nil {
		return err
	}
	archivedPath, err := s.archiveReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackReleaseFailure(
			ctx,
			registry,
			release.Id,
			tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch, err),
		)
	}
	if err = s.executeDynamicUpgradeLifecycleCallback(ctx, registry, activeManifest, manifest, release); err != nil {
		return s.rollbackInstallOrUpgrade(
			ctx,
			registry,
			activeManifest,
			manifest,
			release.Id,
			tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseUpgradeCallback, err),
		)
	}
	if err = s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionUpgrade); err != nil {
		return s.rollbackInstallOrUpgrade(
			ctx,
			registry,
			activeManifest,
			manifest,
			release.Id,
			tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseSQL, err),
		)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackInstallOrUpgrade(
			ctx,
			registry,
			activeManifest,
			manifest,
			release.Id,
			tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err),
		)
	}
	if desiredState == plugintypes.HostStateEnabled.String() {
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackInstallOrUpgrade(
				ctx,
				registry,
				activeManifest,
				manifest,
				release.Id,
				tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err),
			)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackInstallOrUpgrade(
					ctx,
					registry,
					activeManifest,
					manifest,
					release.Id,
					tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err),
				)
			}
		}
	}

	enabled := plugintypes.StatusDisabled
	if desiredState == plugintypes.HostStateEnabled.String() {
		enabled = plugintypes.StatusEnabled
	}
	previousReleaseID := registry.ReleaseId
	registry, err = s.finalizeState(ctx, registry, manifest, release, plugintypes.InstalledYes, enabled)
	if err != nil {
		return s.rollbackInstallOrUpgrade(
			ctx,
			registry,
			activeManifest,
			manifest,
			release.Id,
			tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch, err),
		)
	}
	if previousReleaseID > 0 && previousReleaseID != release.Id {
		if err = s.storeSvc.UpdateReleaseState(ctx, previousReleaseID, plugintypes.ReleaseStatusInstalled, ""); err != nil {
			return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch, err)
		}
	}
	if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.BuildReleaseStatus(plugintypes.InstalledYes, enabled), archivedPath); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch, err)
	}
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if enabled == plugintypes.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginUpgraded)
	}
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release upgraded on primary node."); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err)
	}
	if err = s.syncPluginResourceReferences(ctx, manifest); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err)
	}
	if err = s.notifyRuntimeCacheChanged(ctx, manifest, runtimeChangeReasonPluginUpgraded); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseCacheInvalidation, err)
	}
	if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginUpgraded); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseCacheInvalidation, err)
	}
	if err = s.dispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUpgraded,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
			Status:   &enabled,
		}),
	); err != nil {
		return tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseGovernance, err)
	}
	return nil
}

// executeDynamicUpgradeLifecycleCallback invokes the plugin-owned dynamic
// upgrade execution phase when the target artifact declares it.
func (s *serviceImpl) executeDynamicUpgradeLifecycleCallback(
	ctx context.Context,
	registry *store.PluginRecord,
	activeManifest *catalog.Manifest,
	targetManifest *catalog.Manifest,
	targetRelease *store.ReleaseRecord,
) error {
	if targetManifest == nil {
		return nil
	}
	input, err := s.buildDynamicUpgradeLifecycleInput(ctx, registry, activeManifest, targetManifest, targetRelease)
	if err != nil {
		return err
	}
	decision, err := s.RunDynamicLifecycleCallback(ctx, targetManifest, input)
	if err != nil {
		return err
	}
	if decision != nil && !decision.OK {
		return gerror.New(strings.TrimSpace(decision.Reason))
	}
	return nil
}

// buildDynamicUpgradeLifecycleInput creates the source-equivalent dynamic
// lifecycle payload for the target artifact's Upgrade handler.
func (s *serviceImpl) buildDynamicUpgradeLifecycleInput(
	ctx context.Context,
	registry *store.PluginRecord,
	activeManifest *catalog.Manifest,
	targetManifest *catalog.Manifest,
	targetRelease *store.ReleaseRecord,
) (DynamicLifecycleInput, error) {
	input := DynamicLifecycleInput{
		PluginID:  targetManifest.ID,
		Operation: pluginhost.LifecycleHookUpgrade,
	}
	if registry != nil {
		input.FromVersion = strings.TrimSpace(registry.Version)
	}
	input.ToVersion = strings.TrimSpace(targetManifest.Version)

	fromSnapshot, err := s.dynamicLifecycleSnapshotFromRegistry(ctx, registry, activeManifest)
	if err != nil {
		return input, err
	}
	toSnapshot, err := s.dynamicLifecycleSnapshotFromReleaseOrManifest(targetRelease, targetManifest)
	if err != nil {
		return input, err
	}
	input.FromManifest = fromSnapshot
	input.ToManifest = toSnapshot
	return input, nil
}

// dynamicLifecycleSnapshotFromRegistry returns the effective release snapshot
// before an upgrade, falling back to the active manifest when needed.
func (s *serviceImpl) dynamicLifecycleSnapshotFromRegistry(
	ctx context.Context,
	registry *store.PluginRecord,
	activeManifest *catalog.Manifest,
) (*store.ManifestSnapshot, error) {
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release != nil {
		snapshot, parseErr := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
		if parseErr != nil {
			return nil, parseErr
		}
		if snapshot != nil {
			return snapshot, nil
		}
	}
	if activeManifest == nil {
		return nil, nil
	}
	return s.dynamicLifecycleSnapshotFromManifest(activeManifest)
}

// dynamicLifecycleSnapshotFromReleaseOrManifest returns the target release
// snapshot, falling back to the provided target manifest when the persisted
// snapshot is unavailable.
func (s *serviceImpl) dynamicLifecycleSnapshotFromReleaseOrManifest(
	release *store.ReleaseRecord,
	manifest *catalog.Manifest,
) (*store.ManifestSnapshot, error) {
	if release != nil {
		snapshot, err := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
		if err != nil {
			return nil, err
		}
		if snapshot != nil {
			return snapshot, nil
		}
	}
	return s.dynamicLifecycleSnapshotFromManifest(manifest)
}

// dynamicLifecycleSnapshotFromManifest builds a review snapshot from one
// in-memory manifest.
func (s *serviceImpl) dynamicLifecycleSnapshotFromManifest(manifest *catalog.Manifest) (*store.ManifestSnapshot, error) {
	if manifest == nil {
		return nil, nil
	}
	content, err := s.storeSvc.BuildManifestSnapshot(manifest)
	if err != nil {
		return nil, err
	}
	return s.storeSvc.ParseManifestSnapshot(content)
}

// applyStateToggle flips enable/disable status for the current active release
// without changing the installed version or artifact archive.
func (s *serviceImpl) applyStateToggle(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if err = s.markReconciling(ctx, registry, plugintypes.HostState(desiredState)); err != nil {
		return err
	}

	enabled := plugintypes.StatusDisabled
	eventName := pluginhost.ExtensionPointPluginDisabled
	if desiredState == plugintypes.HostStateEnabled.String() {
		enabled = plugintypes.StatusEnabled
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

	registry, err = s.finalizeState(ctx, registry, manifest, release, plugintypes.InstalledYes, enabled)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, 0, err)
	}
	if release != nil {
		if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.BuildReleaseStatus(plugintypes.InstalledYes, enabled), ""); err != nil {
			return err
		}
	}
	if enabled == plugintypes.StatusDisabled {
		s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginDisabled)
	} else {
		s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginEnabled)
	}
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin status converged on primary node."); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, manifest, runtimeChangeReasonPluginStatusChanged); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginStatusChanged); err != nil {
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
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	if err := s.validateCandidateDependencies(ctx, manifest); err != nil {
		return err
	}
	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
	}
	if err = s.markReconciling(ctx, registry, plugintypes.HostState(desiredState)); err != nil {
		return err
	}

	activeManifest, activeManifestErr := s.loadActiveManifest(ctx, registry)
	if activeManifestErr != nil {
		logger.Warningf(ctx, "load active dynamic manifest before refresh failed plugin=%s err=%v", manifest.ID, activeManifestErr)
	}
	if activeManifest != nil && activeManifest.RuntimeArtifact != nil {
		if s.wasmRuntime != nil {
			s.wasmRuntime.InvalidateCache(ctx, activeManifest.RuntimeArtifact.Path)
		}
	}
	if manifest.RuntimeArtifact != nil {
		if s.wasmRuntime != nil {
			s.wasmRuntime.InvalidateCache(ctx, manifest.RuntimeArtifact.Path)
		}
	}
	archivedPath, err := s.archiveReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}

	enabled := plugintypes.StatusDisabled
	if desiredState == plugintypes.HostStateEnabled.String() {
		enabled = plugintypes.StatusEnabled
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
			}
		}
	}

	registry, err = s.finalizeState(ctx, registry, manifest, release, plugintypes.InstalledYes, enabled)
	if err != nil {
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.BuildReleaseStatus(plugintypes.InstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if enabled == plugintypes.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginRefreshed)
	}
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release refreshed on primary node."); err != nil {
		return err
	}
	if err = s.syncPluginResourceReferences(ctx, manifest); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, manifest, runtimeChangeReasonPluginRefreshed); err != nil {
		return err
	}
	return s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginRefreshed)
}

// validateCandidateDependencies delegates release dependency checks to the root
// plugin facade. Runtime keeps this as an explicit seam to avoid importing the
// facade package from the internal runtime package.
func (s *serviceImpl) validateCandidateDependencies(ctx context.Context, manifest *catalog.Manifest) error {
	if s.dependencyValidator == nil {
		return nil
	}
	return s.dependencyValidator.ValidateDynamicPluginCandidate(ctx, manifest)
}

// shouldRefreshInstalledRelease decides whether an already installed dynamic release
// should be re-converged even though the semantic version did not change. It compares
// desired checksum, registry checksum, and archived release content.
func (s *serviceImpl) shouldRefreshInstalledRelease(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic {
		return false
	}
	if registry.Installed != plugintypes.InstalledYes {
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

	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
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
