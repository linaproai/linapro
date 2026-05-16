// This file contains dynamic-plugin upgrade, refresh, and enablement state
// convergence steps.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// applyUpgrade moves an installed plugin to a new semantic version. Unlike
// refresh, this path runs upgrade SQL and may replace the active release.
func (s *serviceImpl) applyUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
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
		wasm.InvalidateCache(ctx, activeManifest.RuntimeArtifact.Path)
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
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_upgraded")
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release upgraded on primary node."); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_upgraded"); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, "plugin_upgraded"); err != nil {
		return err
	}
	return s.dispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUpgraded,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
			Status:   &enabled,
		}),
	)
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
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_status_changed"); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, "plugin_status_changed"); err != nil {
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
	if err := s.validateCandidateDependencies(ctx, manifest); err != nil {
		return err
	}
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

	activeManifest, activeManifestErr := s.loadActiveManifest(ctx, registry)
	if activeManifestErr != nil {
		logger.Warningf(ctx, "load active dynamic manifest before refresh failed plugin=%s err=%v", manifest.ID, activeManifestErr)
	}
	if activeManifest != nil && activeManifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(ctx, activeManifest.RuntimeArtifact.Path)
	}
	if manifest.RuntimeArtifact != nil {
		wasm.InvalidateCache(ctx, manifest.RuntimeArtifact.Path)
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
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_refreshed")
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release refreshed on primary node."); err != nil {
		return err
	}
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_refreshed"); err != nil {
		return err
	}
	return s.notifyReconcilerChanged(ctx, "plugin_refreshed")
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
