// This file contains dynamic-plugin install reconciliation steps and optional
// mock-data loading.

package runtime

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// applyInstall performs the first activation of a discovered dynamic plugin,
// including artifact archive, SQL install, permission/menu projection, optional
// frontend bundle preparation, and registry finalization.
func (s *serviceImpl) applyInstall(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	desiredState string,
) error {
	if err := s.validateCandidateDependencies(ctx, manifest); err != nil {
		return err
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
	s.cleanupStaleReleaseArtifacts(ctx, manifest.ID)
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Dynamic plugin release installed on primary node."); err != nil {
		return err
	}
	if enabled == catalog.StatusEnabled {
		s.invalidateRuntimeCaches(ctx, manifest.ID, "plugin_installed")
	}
	if err = s.notifyRuntimeCacheChanged(ctx, "plugin_installed"); err != nil {
		return err
	}
	if err = s.notifyReconcilerChanged(ctx, "plugin_installed"); err != nil {
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
	// the install; the typed *lifecycle.MockDataLoadError is propagated so the
	// plugin facade can wrap it once into a stable user-facing bizerr.
	return s.loadDynamicPluginMockData(ctx, manifest)
}

// loadDynamicPluginMockData runs the optional mock-data load phase for one
// dynamic plugin install. Returns nil when the operator did not opt in or when
// the artifact carries no mock SQL. Returns *lifecycle.MockDataLoadError on
// rollback so the facade can convert it to a user-facing bizerr.
func (s *serviceImpl) loadDynamicPluginMockData(ctx context.Context, manifest *catalog.Manifest) error {
	if !catalog.ShouldInstallMockData(ctx) {
		return nil
	}
	if !s.catalogSvc.HasMockSQLData(manifest) {
		return nil
	}

	var (
		executedFiles []string
		failedFile    string
		causeErr      error
	)
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result := s.lifecycleSvc.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		executedFiles = append(executedFiles[:0], result.ExecutedFiles...)
		failedFile = result.FailedFile
		if result.Err != nil {
			causeErr = result.Err
			return result.Err
		}
		return nil
	})
	if txErr == nil {
		return nil
	}
	if causeErr == nil {
		causeErr = txErr
	}
	logger.Warningf(
		ctx,
		"dynamic plugin mock data load rolled back plugin=%s failedFile=%s cause=%v",
		manifest.ID,
		failedFile,
		causeErr,
	)
	rolledBack := make([]string, 0, len(executedFiles)+1)
	rolledBack = append(rolledBack, executedFiles...)
	if failedFile != "" {
		rolledBack = append(rolledBack, failedFile)
	}
	return &lifecycle.MockDataLoadError{
		PluginID:        manifest.ID,
		FailedFile:      failedFile,
		RolledBackFiles: rolledBack,
		Cause:           causeErr,
	}
}
