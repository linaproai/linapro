// This file owns plugin enable and disable orchestration for source and dynamic
// plugins, preserving dependency checks, lifecycle vetoes, runtime convergence,
// cache publication, and observer notifications.

package lifecycle

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// UpdateStatusOptions carries per-request state-change policy values.
type UpdateStatusOptions struct {
	// Authorization optionally carries a host-service authorization snapshot for
	// dynamic plugins that require explicit confirmation before enable.
	Authorization *store.HostServiceAuthorizationInput
	// FrameworkVersion is the current LinaPro framework version used for
	// dependency compatibility checks before enabling a plugin.
	FrameworkVersion string
}

// UpdateStatus executes source or dynamic enable/disable lifecycle orchestration.
func (s *serviceImpl) UpdateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	options UpdateStatusOptions,
) error {
	if status != statusflag.Disabled.Int() && status != statusflag.EnabledValue.Int() {
		return bizerr.NewCode(CodePluginStatusInvalid)
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if status == statusflag.EnabledValue.Int() && plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeDynamic {
		if err = s.runtimeSvc.EnsureRuntimeArtifactAvailable(manifest, "enable"); err != nil {
			return err
		}
	}
	if err = s.syncDiscoveredManifestsForStatus(ctx); err != nil {
		return err
	}
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	if err != nil {
		return err
	}
	if !installed {
		return bizerr.NewCode(CodePluginNotInstalled)
	}
	if status == statusflag.EnabledValue.Int() {
		if err = s.ensureEnableDependencies(ctx, pluginID, options.FrameworkVersion); err != nil {
			return err
		}
	}
	if status == statusflag.Disabled.Int() {
		if err = s.ensureNoReverseDependencies(ctx, pluginID); err != nil {
			return err
		}
	}
	if plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeDynamic {
		return s.updateDynamicStatus(ctx, manifest, pluginID, status, options)
	}
	return s.updateSourceStatus(ctx, manifest, pluginID, status)
}

// syncDiscoveredManifestsForStatus keeps source/dynamic registry rows current
// before status transitions without pulling list projection assembly into lifecycle.
func (s *serviceImpl) syncDiscoveredManifestsForStatus(ctx context.Context) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}
	syncCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return err
	}
	syncCtx = management.WithManifestSnapshot(syncCtx, manifests)
	for _, manifest := range manifests {
		if _, err = s.storeSvc.SyncManifest(syncCtx, manifest); err != nil {
			return err
		}
	}
	return s.integrationSvc.RefreshEnabledSnapshot(syncCtx)
}

// ensureEnableDependencies blocks enable when hard plugin dependencies are not satisfied.
func (s *serviceImpl) ensureEnableDependencies(ctx context.Context, pluginID string, frameworkVersion string) error {
	check, err := s.resolveInstallDependencies(ctx, pluginID, frameworkVersion)
	if err != nil {
		return err
	}
	if !plugindep.HasBlockers(check.Blockers) {
		return nil
	}
	return buildDependencyBlockedError(pluginID, check.Blockers)
}

// updateDynamicStatus dispatches dynamic enable or disable into named branches.
func (s *serviceImpl) updateDynamicStatus(
	ctx context.Context,
	manifest *catalog.Manifest,
	pluginID string,
	status int,
	options UpdateStatusOptions,
) error {
	if status == statusflag.EnabledValue.Int() {
		return s.enableDynamicPlugin(ctx, manifest, pluginID, options)
	}
	return s.disableDynamicPlugin(ctx, pluginID)
}

// enableDynamicPlugin persists authorization, converges runtime state, and
// publishes successful dynamic enable side effects.
func (s *serviceImpl) enableDynamicPlugin(
	ctx context.Context,
	manifest *catalog.Manifest,
	pluginID string,
	options UpdateStatusOptions,
) error {
	if err := s.persistDynamicPluginAuthorization(ctx, manifest, options.Authorization); err != nil {
		return err
	}
	if err := s.reconcileDynamicPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		return err
	}
	if err := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "dynamic_plugin_status_changed"); err != nil {
		return err
	}
	return s.notifyPluginEnabled(ctx, pluginID)
}

// disableDynamicPlugin runs BeforeDisable, converges runtime state, and emits
// observer plus best-effort AfterDisable notifications.
func (s *serviceImpl) disableDynamicPlugin(ctx context.Context, pluginID string) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if err = s.ensureDynamicPluginActiveLifecyclePreconditionAllowed(
		ctx,
		registry,
		pluginhost.LifecycleHookBeforeDisable,
		UninstallOptions{},
	); err != nil {
		return err
	}
	activeManifest := s.loadActiveDynamicLifecycleManifestBestEffort(ctx, pluginID)
	if err = s.reconcileDynamicPluginStatus(ctx, pluginID, statusflag.Disabled.Int()); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "dynamic_plugin_status_changed"); err != nil {
		return err
	}
	if err = s.notifyPluginDisabled(ctx, pluginID); err != nil {
		return err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:  pluginID,
		Operation: pluginhost.LifecycleHookAfterDisable,
	})
	return nil
}

// updateSourceStatus dispatches source enable or disable into named branches.
func (s *serviceImpl) updateSourceStatus(
	ctx context.Context,
	manifest *catalog.Manifest,
	pluginID string,
	status int,
) error {
	if status == statusflag.EnabledValue.Int() {
		return s.enableSourcePlugin(ctx, pluginID)
	}
	return s.disableSourcePlugin(ctx, manifest, pluginID)
}

// enableSourcePlugin updates source governance state and publishes successful enable side effects.
func (s *serviceImpl) enableSourcePlugin(ctx context.Context, pluginID string) error {
	if err := s.storeSvc.SetPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		return err
	}
	if err := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "source_plugin_status_changed"); err != nil {
		return err
	}
	return s.notifyPluginEnabled(ctx, pluginID)
}

// disableSourcePlugin runs BeforeDisable, updates source governance state, and
// emits observer plus best-effort AfterDisable notifications.
func (s *serviceImpl) disableSourcePlugin(ctx context.Context, manifest *catalog.Manifest, pluginID string) error {
	if err := s.executeSourcePluginBeforeLifecycle(
		ctx,
		manifest,
		pluginhost.LifecycleHookBeforeDisable,
		sourceLifecyclePolicy{},
	); err != nil {
		return err
	}
	if err := s.storeSvc.SetPluginStatus(ctx, pluginID, statusflag.Disabled.Int()); err != nil {
		return err
	}
	if err := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "source_plugin_status_changed"); err != nil {
		return err
	}
	if err := s.notifyPluginDisabled(ctx, pluginID); err != nil {
		return err
	}
	s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterDisable, sourceLifecyclePolicy{})
	return nil
}

// reconcileDynamicPluginStatus converts enable/disable requests into runtime
// reconciler host state transitions used by dynamic plugins.
func (s *serviceImpl) reconcileDynamicPluginStatus(ctx context.Context, pluginID string, status int) error {
	targetState := plugintypes.HostStateInstalled.String()
	if status == statusflag.EnabledValue.Int() {
		targetState = plugintypes.HostStateEnabled.String()
	}
	return s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, targetState, runtime.DynamicReconcileOptions{})
}
