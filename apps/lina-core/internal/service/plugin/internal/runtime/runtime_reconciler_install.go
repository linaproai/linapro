// This file contains dynamic-plugin install reconciliation steps and optional
// mock-data loading.

package runtime

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginhost"
)

// applyInstall performs the first activation of a discovered dynamic plugin,
// including artifact archive, SQL install, permission/menu projection, optional
// frontend bundle preparation, and registry finalization.
func (s *serviceImpl) applyInstall(
	ctx context.Context,
	registry *store.PluginRecord,
	manifest *catalog.Manifest,
	desiredState string,
	options DynamicReconcileOptions,
) error {
	if err := s.validateCandidateDependencies(ctx, manifest); err != nil {
		return err
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
		return s.rollbackReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionInstall); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if desiredState == plugintypes.HostStateEnabled.String() {
		if err = s.validateFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
		}
		if frontend.HasFrontendAssets(manifest) {
			if err = s.ensureFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
			}
		}
	}

	enabled := plugintypes.StatusDisabled
	if desiredState == plugintypes.HostStateEnabled.String() {
		enabled = plugintypes.StatusEnabled
	}
	registry, err = s.finalizeState(ctx, registry, manifest, release, plugintypes.InstalledYes, enabled)
	if err != nil {
		return s.rollbackInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.storeSvc.UpdateReleaseState(ctx, release.Id, plugintypes.BuildReleaseStatus(plugintypes.InstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release installed on primary node."); err != nil {
		return err
	}
	if err = s.syncPluginResourceReferences(ctx, manifest); err != nil {
		return err
	}
	if enabled == plugintypes.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest, runtimeChangeReasonPluginInstalled)
	}
	if err = s.notifyRuntimeCacheChanged(ctx, manifest, runtimeChangeReasonPluginInstalled); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, runtimeChangeReasonPluginInstalled); err != nil {
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
	if enabled == plugintypes.StatusEnabled {
		if err = s.dispatchHookEvent(
			ctx,
			pluginhost.ExtensionPointPluginEnabled,
			pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
				PluginID: manifest.ID,
				Name:     manifest.Name,
				Version:  manifest.Version,
				Status:   &enabled,
			}),
		); err != nil {
			return err
		}
	}
	// Mock-data load is the final, optional install decoration. It runs only when
	// the operator opted in via the install request. Mock failure does NOT undo
	// the install; the typed *migration.MockDataLoadError is propagated so the
	// plugin facade can wrap it once into a stable user-facing bizerr.
	return s.loadDynamicPluginMockData(ctx, manifest, options.InstallMockData)
}

// loadDynamicPluginMockData runs the optional mock-data load phase for one
// dynamic plugin install. Returns nil when the operator did not opt in or when
// the artifact carries no mock SQL. Returns *migration.MockDataLoadError on
// rollback so the facade can convert it to a user-facing bizerr.
func (s *serviceImpl) loadDynamicPluginMockData(
	ctx context.Context,
	manifest *catalog.Manifest,
	installMockData bool,
) error {
	if !installMockData {
		return nil
	}
	if !s.catalogSvc.HasMockSQLData(manifest) {
		return nil
	}
	return s.migrationSvc.ExecuteManifestMockSQLFiles(ctx, manifest)
}
