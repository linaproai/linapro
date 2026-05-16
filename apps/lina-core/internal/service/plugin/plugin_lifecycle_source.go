// This file implements explicit install and uninstall orchestration for source
// plugins so discovery no longer implies automatic installation.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// installSourcePlugin performs the explicit lifecycle for one discovered source plugin.
func (s *serviceImpl) installSourcePlugin(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginSourceManifestRequired)
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if registry.Installed == catalog.InstalledYes {
		return nil
	}

	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return bizerr.NewCode(
			CodePluginReleaseNotFound,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("version", manifest.Version),
		)
	}

	if err = s.executeSourcePluginBeforeLifecycle(ctx, manifest, pluginhost.LifecycleHookBeforeInstall, false); err != nil {
		return err
	}
	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		return err
	}

	if err = s.integrationSvc.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}
	if err = s.applySourcePluginStableState(ctx, registry, catalog.InstalledYes, catalog.StatusDisabled); err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}

	registry, err = s.catalogSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}
	if registry == nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return bizerr.NewCode(CodePluginSourceRegistryAfterInstallNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if err = s.catalogSvc.UpdateReleaseState(
		ctx,
		release.Id,
		catalog.BuildReleaseStatus(registry.Installed, registry.Status),
		s.catalogSvc.BuildPackagePath(manifest),
	); err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Source plugin installed from management API."); err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}
	if err = s.integrationSvc.DispatchPluginHookEvent(
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
	// Mock-data load is the final, optional install decoration. It runs only when
	// the operator explicitly opted in via InstallOptions.InstallMockData. Mock
	// failure does NOT undo the install: the registry, menu sync, release state,
	// and hook event all remain effective. The returned error tells the caller
	// install succeeded but mock data was rolled back and they can choose to
	// accept the no-mock state or uninstall + reinstall after fixing the SQL.
	return s.loadSourcePluginMockData(ctx, manifest)
}

// loadSourcePluginMockData runs the optional mock-data load phase for one source
// plugin install. Returns nil when the operator did not opt in or when the plugin
// has no mock-data SQL files. On failure returns a *lifecycle.MockDataLoadError
// so the plugin facade can wrap it once into a user-facing bizerr regardless of
// install path (source vs dynamic).
func (s *serviceImpl) loadSourcePluginMockData(ctx context.Context, manifest *catalog.Manifest) error {
	if !shouldInstallMockData(ctx) {
		return nil
	}
	if !s.catalogSvc.HasMockSQLData(manifest) {
		return nil
	}
	return executeMockDataLoadTransaction(ctx, s.lifecycleSvc, manifest)
}

// uninstallSourcePlugin performs the explicit lifecycle for one installed source plugin.
func (s *serviceImpl) uninstallSourcePlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	options UninstallOptions,
) error {
	if manifest == nil {
		return bizerr.NewCode(CodePluginSourceManifestRequired)
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, manifest)
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
		release, err = s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return err
		}
	}

	if err = s.executeSourcePluginBeforeLifecycle(ctx, manifest, pluginhost.LifecycleHookBeforeUninstall, options.Force); err != nil {
		return err
	}
	if options.PurgeStorageData {
		if err = s.executeSourcePluginUninstallHandler(ctx, manifest, options); err != nil {
			return err
		}
		if err = s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
			return err
		}
	}
	if err = s.integrationSvc.DeletePluginMenusByManifest(ctx, manifest); err != nil {
		return err
	}
	if err = s.deleteSourcePluginResourceRefs(ctx, manifest, release); err != nil {
		return err
	}
	if err = s.applySourcePluginStableState(ctx, registry, catalog.InstalledNo, catalog.StatusDisabled); err != nil {
		return err
	}

	registry, err = s.catalogSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryAfterUninstallNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if release != nil {
		if err = s.catalogSvc.UpdateReleaseState(
			ctx,
			release.Id,
			catalog.BuildReleaseStatus(registry.Installed, registry.Status),
			s.catalogSvc.BuildPackagePath(manifest),
		); err != nil {
			return err
		}
	}
	if err = s.catalogSvc.SyncMetadata(ctx, manifest, registry, "Source plugin uninstalled from management API."); err != nil {
		return err
	}
	return s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUninstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
}

// executeSourcePluginBeforeLifecycle invokes lifecycle facade precondition
// callbacks registered by one source plugin before host side effects run.
func (s *serviceImpl) executeSourcePluginBeforeLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	force bool,
) error {
	if manifest == nil || manifest.SourcePlugin == nil {
		return nil
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInput(
			manifest.ID,
			hook.String(),
		),
		Participants: []pluginhost.LifecycleParticipant{
			{
				PluginID: manifest.ID,
				Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
			},
		},
	})
	if result.OK {
		return nil
	}
	reasons := summarizeLifecycleVetoReasons(result.Decisions)
	if force && hook == pluginhost.LifecycleHookBeforeUninstall {
		if err := s.ensureForceUninstallEnabled(ctx); err != nil {
			return err
		}
		logger.Warningf(
			ctx,
			"source plugin lifecycle callback force bypass operation=%s plugin=%s reasons=%s",
			hook,
			manifest.ID,
			reasons,
		)
		return nil
	}
	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", manifest.ID),
		bizerr.P("reasons", reasons),
	)
}

// executeSourcePluginAfterLifecycle invokes non-blocking lifecycle callbacks
// registered by one source plugin after host side effects have succeeded.
func (s *serviceImpl) executeSourcePluginAfterLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
) {
	if manifest == nil || manifest.SourcePlugin == nil {
		return
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInput(
			manifest.ID,
			hook.String(),
		),
		Participants: []pluginhost.LifecycleParticipant{
			{
				PluginID: manifest.ID,
				Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
			},
		},
	})
	if result.OK {
		return
	}
	logger.Warningf(
		ctx,
		"source plugin after lifecycle callback failed operation=%s plugin=%s reasons=%s",
		hook,
		manifest.ID,
		summarizeLifecycleVetoReasons(result.Decisions),
	)
}

// executeSourcePluginUninstallHandler invokes one optional source-plugin cleanup callback
// before uninstall SQL removes plugin-owned tables.
func (s *serviceImpl) executeSourcePluginUninstallHandler(
	ctx context.Context,
	manifest *catalog.Manifest,
	options UninstallOptions,
) error {
	if manifest == nil || manifest.SourcePlugin == nil || !options.PurgeStorageData {
		return nil
	}
	handler := manifest.SourcePlugin.GetUninstallHandler()
	if handler == nil {
		return nil
	}
	return handler(
		ctx,
		pluginhost.NewSourcePluginUninstallInput(manifest.ID, options.PurgeStorageData),
	)
}

// applySourcePluginStableState updates one source plugin registry row to a stable installed/disabled state.
func (s *serviceImpl) applySourcePluginStableState(
	ctx context.Context,
	registry *entity.SysPlugin,
	installed int,
	enabled int,
) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryRequired)
	}

	stableState := catalog.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if registry.Generation <= 0 {
		data.Generation = int64(1)
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
		data.DisabledAt = gtime.Now()
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	_, err = s.catalogSvc.RefreshStartupRegistry(ctx, registry.PluginId)
	return err
}

// deleteSourcePluginResourceRefs removes governance resource refs for the given source-plugin release.
func (s *serviceImpl) deleteSourcePluginResourceRefs(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) error {
	if manifest == nil || release == nil {
		return nil
	}
	_, err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{
			PluginId:  manifest.ID,
			ReleaseId: release.Id,
		}).
		Delete()
	return err
}

// rollbackSourcePluginInstall best-effort restores source-plugin governance after a failed install.
func (s *serviceImpl) rollbackSourcePluginInstall(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) {
	if manifest == nil {
		return
	}

	if err := s.lifecycleSvc.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
		logger.Warningf(ctx, "rollback source plugin uninstall SQL failed plugin=%s err=%v", manifest.ID, err)
	}
	if err := s.integrationSvc.DeletePluginMenusByManifest(ctx, manifest); err != nil {
		logger.Warningf(ctx, "rollback source plugin menus failed plugin=%s err=%v", manifest.ID, err)
	}
	if err := s.deleteSourcePluginResourceRefs(ctx, manifest, release); err != nil {
		logger.Warningf(ctx, "rollback source plugin resource refs failed plugin=%s err=%v", manifest.ID, err)
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		logger.Warningf(ctx, "rollback source plugin registry lookup failed plugin=%s err=%v", manifest.ID, err)
	} else if registry != nil {
		if err = s.applySourcePluginStableState(ctx, registry, catalog.InstalledNo, catalog.StatusDisabled); err != nil {
			logger.Warningf(ctx, "rollback source plugin stable state failed plugin=%s err=%v", manifest.ID, err)
		}
	}
	if release != nil {
		if err = s.catalogSvc.UpdateReleaseState(
			ctx,
			release.Id,
			catalog.ReleaseStatusUninstalled,
			s.catalogSvc.BuildPackagePath(manifest),
		); err != nil {
			logger.Warningf(ctx, "rollback source plugin release state failed plugin=%s release=%d err=%v", manifest.ID, release.Id, err)
		}
	}
}
