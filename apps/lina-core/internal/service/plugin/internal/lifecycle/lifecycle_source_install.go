// This file implements explicit source-plugin install, uninstall, and rollback
// orchestration now owned by lifecycle.

package lifecycle

import (
	"context"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// SourceInstallOptions carries source-plugin install metadata.
type SourceInstallOptions struct {
	// StartupAutoEnable reports whether plugin.autoEnable initiated this install.
	StartupAutoEnable bool
	// InstallMockData requests loading optional mock-data SQL during install.
	InstallMockData bool
}

// sourceLifecyclePolicy carries host-side action options into source-plugin
// generic lifecycle callbacks.
type sourceLifecyclePolicy struct {
	force               bool
	allowForceUninstall bool
	startupAutoEnable   bool
	purgeStorageData    bool
}

// timePtr returns a pointer to value for generated DO time fields that preserve
// database NULL semantics with *time.Time.
func timePtr(value time.Time) *time.Time {
	return &value
}

// installSourcePlugin performs the explicit lifecycle for one discovered source plugin.
func (s *serviceImpl) installSourcePlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	options SourceInstallOptions,
) error {
	registry, release, shouldInstall, err := s.prepareSourcePluginInstall(ctx, manifest)
	if err != nil || !shouldInstall {
		return err
	}
	if err = s.executeSourcePluginBeforeLifecycle(
		ctx,
		manifest,
		pluginhost.LifecycleHookBeforeInstall,
		sourceLifecyclePolicy{startupAutoEnable: options.StartupAutoEnable},
	); err != nil {
		return err
	}
	if err = s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionInstall); err != nil {
		return err
	}
	if err = s.applySourcePluginInstallGovernance(ctx, manifest, registry, release); err != nil {
		s.rollbackSourcePluginInstall(ctx, manifest, release)
		return err
	}
	if err = s.dispatchSourcePluginInstallHook(ctx, manifest); err != nil {
		return err
	}
	return s.loadSourcePluginMockData(ctx, manifest, options.InstallMockData)
}

// prepareSourcePluginInstall resolves registry and release state before source
// install side effects begin.
func (s *serviceImpl) prepareSourcePluginInstall(
	ctx context.Context,
	manifest *catalog.Manifest,
) (*store.PluginRecord, *store.ReleaseRecord, bool, error) {
	if manifest == nil {
		return nil, nil, false, bizerr.NewCode(CodePluginSourceManifestRequired)
	}

	registry, err := s.storeSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return nil, nil, false, err
	}
	if registry == nil {
		return nil, nil, false, bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if registry.Installed == plugintypes.InstalledYes {
		return registry, nil, false, nil
	}

	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, nil, false, err
	}
	if release == nil {
		return nil, nil, false, bizerr.NewCode(
			CodePluginReleaseNotFound,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("version", manifest.Version),
		)
	}
	return registry, release, true, nil
}

// applySourcePluginInstallGovernance applies rollback-protected source install
// side effects after install SQL succeeds.
func (s *serviceImpl) applySourcePluginInstallGovernance(
	ctx context.Context,
	manifest *catalog.Manifest,
	registry *store.PluginRecord,
	release *store.ReleaseRecord,
) error {
	if err := s.integrationSvc.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return err
	}
	if err := s.applySourcePluginStableState(ctx, registry, plugintypes.InstalledYes, plugintypes.StatusDisabled); err != nil {
		return err
	}

	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryAfterInstallNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if err = s.storeSvc.UpdateReleaseState(
		ctx,
		release.Id,
		plugintypes.BuildReleaseStatus(registry.Installed, registry.Status),
		s.storeSvc.BuildPackagePath(manifest),
	); err != nil {
		return err
	}
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Source plugin installed from management API."); err != nil {
		return err
	}
	if err = s.integrationSvc.SyncPluginResourceReferences(ctx, manifest); err != nil {
		return err
	}
	return nil
}

// dispatchSourcePluginInstallHook emits the source install host hook after
// source governance state has been persisted.
func (s *serviceImpl) dispatchSourcePluginInstallHook(ctx context.Context, manifest *catalog.Manifest) error {
	return s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
}

// loadSourcePluginMockData runs the optional mock-data load phase for one source
// plugin install.
func (s *serviceImpl) loadSourcePluginMockData(
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

// executeSourcePluginBeforeLifecycle invokes lifecycle facade precondition
// callbacks registered by one source plugin before host side effects run.
func (s *serviceImpl) executeSourcePluginBeforeLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	policy sourceLifecyclePolicy,
) error {
	if manifest == nil || manifest.SourcePlugin == nil {
		return nil
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInputWithPolicy(
			manifest.ID,
			hook.String(),
			pluginhost.SourcePluginLifecyclePolicy{
				StartupAutoEnable: policy.startupAutoEnable,
				PurgeStorageData:  policy.purgeStorageData,
			},
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
	reasons := s.summarizeLocalizedLifecycleVetoReasons(ctx, result.Decisions)
	if policy.force && hook == pluginhost.LifecycleHookBeforeUninstall {
		if err := ensureForceUninstallEnabled(UninstallOptions{
			Force:               true,
			AllowForceUninstall: policy.allowForceUninstall,
		}); err != nil {
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

// uninstallSourcePlugin performs the explicit lifecycle for one installed source plugin.
func (s *serviceImpl) uninstallSourcePlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	options UninstallOptions,
) error {
	registry, release, shouldUninstall, err := s.prepareSourcePluginUninstall(ctx, manifest)
	if err != nil || !shouldUninstall {
		return err
	}
	if err = s.executeSourcePluginBeforeLifecycle(ctx, manifest, pluginhost.LifecycleHookBeforeUninstall, sourceLifecyclePolicy{
		force:               options.Force,
		allowForceUninstall: options.AllowForceUninstall,
		purgeStorageData:    options.PurgeStorageData,
	}); err != nil {
		return err
	}
	if err = s.purgeSourcePluginStorage(ctx, manifest, options); err != nil {
		return err
	}
	if err = s.applySourcePluginUninstallGovernance(ctx, manifest, registry, release); err != nil {
		return err
	}
	return s.dispatchSourcePluginUninstallHook(ctx, manifest)
}

// prepareSourcePluginUninstall resolves registry and release state before
// source uninstall side effects begin.
func (s *serviceImpl) prepareSourcePluginUninstall(
	ctx context.Context,
	manifest *catalog.Manifest,
) (*store.PluginRecord, *store.ReleaseRecord, bool, error) {
	if manifest == nil {
		return nil, nil, false, bizerr.NewCode(CodePluginSourceManifestRequired)
	}

	registry, err := s.storeSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return nil, nil, false, err
	}
	if registry == nil || registry.Installed != plugintypes.InstalledYes {
		return registry, nil, false, nil
	}

	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, nil, false, err
	}
	if release == nil {
		release, err = s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return nil, nil, false, err
		}
	}
	return registry, release, true, nil
}

// purgeSourcePluginStorage runs optional source uninstall cleanup before
// governance state is marked uninstalled.
func (s *serviceImpl) purgeSourcePluginStorage(ctx context.Context, manifest *catalog.Manifest, options UninstallOptions) error {
	if !options.PurgeStorageData {
		return nil
	}
	if err := s.executeSourcePluginUninstallHandler(ctx, manifest, options); err != nil {
		return err
	}
	return s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionUninstall)
}

// applySourcePluginUninstallGovernance removes source governance projections
// and writes the stable uninstalled registry state.
func (s *serviceImpl) applySourcePluginUninstallGovernance(
	ctx context.Context,
	manifest *catalog.Manifest,
	registry *store.PluginRecord,
	release *store.ReleaseRecord,
) error {
	if err := s.integrationSvc.DeletePluginMenusByManifest(ctx, manifest); err != nil {
		return err
	}
	if err := s.deleteSourcePluginResourceRefs(ctx, manifest, release); err != nil {
		return err
	}
	if err := s.applySourcePluginStableState(ctx, registry, plugintypes.InstalledNo, plugintypes.StatusDisabled); err != nil {
		return err
	}

	var err error
	registry, err = s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryAfterUninstallNotFound, bizerr.P("pluginId", manifest.ID))
	}
	if release != nil {
		if err = s.storeSvc.UpdateReleaseState(
			ctx,
			release.Id,
			plugintypes.BuildReleaseStatus(registry.Installed, registry.Status),
			s.storeSvc.BuildPackagePath(manifest),
		); err != nil {
			return err
		}
	}
	if err = s.storeSvc.SyncMetadata(ctx, manifest, registry, "Source plugin uninstalled from management API."); err != nil {
		return err
	}
	return nil
}

// dispatchSourcePluginUninstallHook emits the source uninstall host hook after
// source governance state has been persisted.
func (s *serviceImpl) dispatchSourcePluginUninstallHook(ctx context.Context, manifest *catalog.Manifest) error {
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
	var services pluginhost.Services
	if s.sourceServices != nil {
		services = s.sourceServices.SourceServicesForPlugin(manifest.ID)
	}
	return handler(
		ctx,
		pluginhost.NewSourcePluginUninstallInputWithServices(
			manifest.ID,
			options.PurgeStorageData,
			services,
		),
	)
}

// executeSourcePluginAfterLifecycle invokes non-blocking lifecycle callbacks
// registered by one source plugin after host side effects have succeeded.
func (s *serviceImpl) executeSourcePluginAfterLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	policy sourceLifecyclePolicy,
) {
	if manifest == nil || manifest.SourcePlugin == nil {
		return
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInputWithPolicy(
			manifest.ID,
			hook.String(),
			pluginhost.SourcePluginLifecyclePolicy{
				StartupAutoEnable: policy.startupAutoEnable,
				PurgeStorageData:  policy.purgeStorageData,
			},
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

// applySourcePluginStableState updates one source plugin registry row to a stable installed/disabled state.
func (s *serviceImpl) applySourcePluginStableState(
	ctx context.Context,
	registry *store.PluginRecord,
	installed int,
	enabled int,
) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryRequired)
	}

	stableState := plugintypes.DeriveHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if registry.Generation <= 0 {
		data.Generation = int64(1)
	}
	if installed == plugintypes.InstalledYes {
		if registry.Installed != plugintypes.InstalledYes {
			data.InstalledAt = timePtr(time.Now())
		}
		if enabled == plugintypes.StatusEnabled {
			data.EnabledAt = timePtr(time.Now())
		} else {
			data.DisabledAt = timePtr(time.Now())
		}
	} else {
		data.Status = plugintypes.StatusDisabled
		data.DisabledAt = timePtr(time.Now())
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	_, err = s.storeSvc.RefreshStartupRegistry(ctx, registry.PluginId)
	return err
}

// deleteSourcePluginResourceRefs removes governance resource refs for the given source-plugin release.
func (s *serviceImpl) deleteSourcePluginResourceRefs(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *store.ReleaseRecord,
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
	release *store.ReleaseRecord,
) {
	if manifest == nil {
		return
	}

	if err := s.migrationSvc.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionUninstall); err != nil {
		logger.Warningf(ctx, "rollback source plugin uninstall SQL failed plugin=%s err=%v", manifest.ID, err)
	}
	if err := s.integrationSvc.DeletePluginMenusByManifest(ctx, manifest); err != nil {
		logger.Warningf(ctx, "rollback source plugin menus failed plugin=%s err=%v", manifest.ID, err)
	}
	if err := s.deleteSourcePluginResourceRefs(ctx, manifest, release); err != nil {
		logger.Warningf(ctx, "rollback source plugin resource refs failed plugin=%s err=%v", manifest.ID, err)
	}
	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		logger.Warningf(ctx, "rollback source plugin registry lookup failed plugin=%s err=%v", manifest.ID, err)
	} else if registry != nil {
		if err = s.applySourcePluginStableState(ctx, registry, plugintypes.InstalledNo, plugintypes.StatusDisabled); err != nil {
			logger.Warningf(ctx, "rollback source plugin stable state failed plugin=%s err=%v", manifest.ID, err)
		}
	}
	if release != nil {
		if err = s.storeSvc.UpdateReleaseState(
			ctx,
			release.Id,
			plugintypes.ReleaseStatusUninstalled,
			s.storeSvc.BuildPackagePath(manifest),
		); err != nil {
			logger.Warningf(ctx, "rollback source plugin release state failed plugin=%s release=%d err=%v", manifest.ID, release.Id, err)
		}
	}
}

// summarizeLifecycleVetoReasons builds one deterministic raw reason string for
// audit logs and development diagnostics.
func summarizeLifecycleVetoReasons(decisions []pluginhost.LifecycleDecision) string {
	return summarizeLifecycleVetoReasonsWithTranslator(decisions, nil)
}

// summarizeLocalizedLifecycleVetoReasons builds one deterministic localized
// reason string for caller-visible lifecycle precondition errors.
func (s *serviceImpl) summarizeLocalizedLifecycleVetoReasons(
	ctx context.Context,
	decisions []pluginhost.LifecycleDecision,
) string {
	return summarizeLifecycleVetoReasonsWithTranslator(decisions, func(key string) string {
		if s == nil || s.i18nSvc == nil {
			return ""
		}
		return s.i18nSvc.Translate(ctx, key, "")
	})
}

// summarizeLifecycleVetoReasonsWithTranslator applies an optional translator to
// reason keys while preserving the existing plugin-prefixed reason format used
// by lifecycle callers.
func summarizeLifecycleVetoReasonsWithTranslator(
	decisions []pluginhost.LifecycleDecision,
	translate func(key string) string,
) string {
	shared := make([]lifecycleVetoDecision, 0, len(decisions))
	for _, decision := range decisions {
		shared = append(shared, lifecycleVetoDecision{
			PluginID: decision.PluginID,
			OK:       decision.OK,
			Reason:   decision.Reason,
			Err:      decision.Err,
		})
	}
	return summarizeLifecycleVetoDecisionReasons(shared, translate)
}
