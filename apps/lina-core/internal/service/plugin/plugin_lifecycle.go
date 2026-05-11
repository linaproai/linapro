// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"
	"strconv"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// Install executes the install lifecycle and optionally persists one host-confirmed
// host service authorization snapshot when the target is a dynamic plugin. The
// options.InstallMockData flag is threaded through context so deeply nested
// runtime/reconciler code can detect mock opt-in without mass signature changes.
//
// On a rolled-back mock-data load the plugin is fully installed (registry, menus,
// release state) — only the mock data was reverted. Install returns a stable
// bizerr (CodePluginInstallMockDataFailed) carrying pluginId, failedFile,
// rolledBackFiles, and cause so the caller can render a precise message.
func (s *serviceImpl) Install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (err error) {
	ctx = withInstallMockData(ctx, options.InstallMockData)
	defer func() {
		err = wrapMockDataLoadError(err)
	}()

	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if err = applyInstallModeSelection(manifest, options.InstallMode); err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		if err = s.installSourcePlugin(ctx, manifest); err != nil {
			if !isMockDataLoadError(err) {
				return err
			}
			if snapshotErr := s.syncEnabledSnapshotFromRegistry(ctx, pluginID); snapshotErr != nil {
				return snapshotErr
			}
			if markErr := s.markRuntimeCacheChanged(ctx, "source_plugin_installed"); markErr != nil {
				return markErr
			}
			return err
		}
		if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
			return err
		}
		if err = s.markRuntimeCacheChanged(ctx, "source_plugin_installed"); err != nil {
			return err
		}
		return notifyPluginInstalled(ctx, pluginID)
	}
	if err = s.persistDynamicPluginAuthorization(ctx, manifest, options.Authorization); err != nil {
		return err
	}
	if err = s.lifecycleSvc.Install(ctx, pluginID); err != nil {
		return err
	}
	// Dynamic lifecycle reloads the manifest from the runtime artifact. Re-sync
	// the operator-selected governance fields so installMode cannot be reset to
	// the artifact default after the request has already been validated.
	if _, err = s.catalogSvc.SyncManifest(ctx, manifest); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	return notifyPluginInstalled(ctx, pluginID)
}

// applyInstallModeSelection validates the explicit install-mode request and
// applies it to the short-lived desired manifest before registry synchronization.
func applyInstallModeSelection(manifest *catalog.Manifest, installMode string) error {
	if manifest == nil {
		return nil
	}
	scopeNature := catalog.NormalizeScopeNature(manifest.ScopeNature)
	if strings.TrimSpace(installMode) == "" {
		installMode = manifest.DefaultInstallMode
	}
	if !catalog.IsSupportedInstallMode(installMode) {
		return bizerr.NewCode(CodePluginInstallModeInvalid)
	}
	mode := catalog.NormalizeInstallMode(installMode)
	if scopeNature == catalog.ScopeNaturePlatformOnly && mode != catalog.InstallModeGlobal {
		return bizerr.NewCode(
			CodePluginInstallModeInvalidForScopeNature,
			bizerr.P("pluginId", manifest.ID),
			bizerr.P("scopeNature", scopeNature.String()),
			bizerr.P("installMode", mode.String()),
		)
	}
	manifest.DefaultInstallMode = mode.String()
	return nil
}

// Uninstall executes the uninstall lifecycle for an installed plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string) error {
	return s.UninstallWithOptions(ctx, pluginID, UninstallOptions{PurgeStorageData: true})
}

// UninstallWithOptions executes the uninstall lifecycle for an installed plugin using one explicit policy snapshot.
func (s *serviceImpl) UninstallWithOptions(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if err = s.ensureLifecycleGuardAllowed(ctx, pluginID, pluginhost.GuardHookCanUninstall, options.Force); err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		if err = s.uninstallSourcePlugin(ctx, manifest, options); err != nil {
			return err
		}
		if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
			return err
		}
		if err = s.markRuntimeCacheChanged(ctx, "source_plugin_uninstalled"); err != nil {
			return err
		}
		return notifyPluginUninstalled(ctx, pluginID)
	}
	if err = s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	return notifyPluginUninstalled(ctx, pluginID)
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled,
// and optionally persists one host-confirmed host service authorization snapshot
// before enabling a dynamic plugin.
func (s *serviceImpl) UpdateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	authorization *HostServiceAuthorizationInput,
) error {
	return s.updateStatus(ctx, pluginID, status, authorization)
}

// updateStatus centralizes enable/disable validation so source and dynamic
// plugins both honor installed-state checks before status transitions.
func (s *serviceImpl) updateStatus(
	ctx context.Context,
	pluginID string,
	status int,
	authorization *HostServiceAuthorizationInput,
) error {
	if status != catalog.StatusDisabled && status != catalog.StatusEnabled {
		return bizerr.NewCode(CodePluginStatusInvalid)
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if status == catalog.StatusEnabled && catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		if err = s.runtimeSvc.EnsureRuntimeArtifactAvailable(manifest, "enable"); err != nil {
			return err
		}
	}
	if status == catalog.StatusDisabled {
		if err = s.ensureLifecycleGuardAllowed(ctx, pluginID, pluginhost.GuardHookCanDisable, false); err != nil {
			return err
		}
	}
	if err = s.SyncSourcePlugins(ctx); err != nil {
		return err
	}
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	if err != nil {
		return err
	}
	if !installed {
		return bizerr.NewCode(CodePluginNotInstalled)
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		if status == catalog.StatusEnabled {
			if err = s.persistDynamicPluginAuthorization(ctx, manifest, authorization); err != nil {
				return err
			}
		}
		if err = s.reconcileDynamicPluginStatus(ctx, pluginID, status); err != nil {
			return err
		}
		if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
			return err
		}
		if status == catalog.StatusEnabled {
			return notifyPluginEnabled(ctx, pluginID)
		}
		return notifyPluginDisabled(ctx, pluginID)
	}
	if err = s.catalogSvc.SetPluginStatus(ctx, pluginID, status); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	if err = s.markRuntimeCacheChanged(ctx, "source_plugin_status_changed"); err != nil {
		return err
	}
	if status == catalog.StatusEnabled {
		return notifyPluginEnabled(ctx, pluginID)
	}
	return notifyPluginDisabled(ctx, pluginID)
}

// Enable enables the specified plugin.
func (s *serviceImpl) Enable(ctx context.Context, pluginID string) error {
	return s.updateStatus(ctx, pluginID, catalog.StatusEnabled, nil)
}

// Disable disables the specified plugin.
func (s *serviceImpl) Disable(ctx context.Context, pluginID string) error {
	return s.updateStatus(ctx, pluginID, catalog.StatusDisabled, nil)
}

// persistDynamicPluginAuthorization refreshes the release snapshot for dynamic
// plugins so install/enable flows can reuse one governance preparation path.
func (s *serviceImpl) persistDynamicPluginAuthorization(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) error {
	if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil
	}
	if _, err := s.catalogSvc.SyncManifest(ctx, manifest); err != nil {
		return err
	}
	if _, err := s.catalogSvc.PersistReleaseHostServiceAuthorization(ctx, manifest, authorization); err != nil {
		return err
	}
	return nil
}

// reconcileDynamicPluginStatus converts facade enable/disable requests into the
// runtime reconciler host state transitions used by dynamic plugins.
func (s *serviceImpl) reconcileDynamicPluginStatus(ctx context.Context, pluginID string, status int) error {
	targetState := catalog.HostStateInstalled.String()
	if status == catalog.StatusEnabled {
		targetState = catalog.HostStateEnabled.String()
	}
	return s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, targetState)
}

// IsInstalled returns whether a plugin is installed.
func (s *serviceImpl) IsInstalled(ctx context.Context, pluginID string) bool {
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	return err == nil && installed
}

// IsEnabled returns whether a plugin is enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "is_enabled")
	return s.integrationSvc.IsEnabled(ctx, pluginID)
}

// EnsureTenantDeleteAllowed runs plugin lifecycle guards before tenant deletion
// continues in the tenant capability provider.
func (s *serviceImpl) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	return s.ensureTenantLifecycleGuardAllowed(ctx, tenantID, pluginhost.GuardHookCanTenantDelete)
}

// ensureTenantLifecycleGuardAllowed runs tenant-scoped lifecycle guards and
// converts vetoes to the same stable lifecycle guard error used by plugin
// disable and uninstall operations.
func (s *serviceImpl) ensureTenantLifecycleGuardAllowed(ctx context.Context, tenantID int, hook pluginhost.GuardHook) error {
	result := pluginhost.RunLifecycleGuards(ctx, pluginhost.GuardRequest{
		Hook:         hook,
		TenantID:     tenantID,
		Participants: pluginhost.ListLifecycleGuardParticipants(),
	})
	if result.OK {
		return nil
	}

	return bizerr.NewCode(
		CodePluginLifecycleGuardVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", "tenant:"+strconv.Itoa(tenantID)),
		bizerr.P("reasons", summarizeGuardVetoReasons(result.Decisions)),
	)
}

// ensureLifecycleGuardAllowed runs source-plugin lifecycle guards before a
// protected plugin action and converts vetoes to stable caller-visible errors.
func (s *serviceImpl) ensureLifecycleGuardAllowed(
	ctx context.Context,
	pluginID string,
	hook pluginhost.GuardHook,
	force bool,
) error {
	result := pluginhost.RunLifecycleGuards(ctx, pluginhost.GuardRequest{
		Hook:         hook,
		Participants: pluginhost.ListLifecycleGuardParticipantsForPlugin(pluginID),
	})
	if result.OK {
		return nil
	}

	reasons := summarizeGuardVetoReasons(result.Decisions)
	if force && hook == pluginhost.GuardHookCanUninstall {
		if !s.configSvc.GetPlugin(ctx).AllowForceUninstall {
			return bizerr.NewCode(CodePluginForceUninstallDisabled)
		}
		logger.Warningf(
			ctx,
			"plugin lifecycle guard force bypass operation=%s plugin=%s reasons=%s",
			hook,
			pluginID,
			reasons,
		)
		return nil
	}

	return bizerr.NewCode(
		CodePluginLifecycleGuardVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", pluginID),
		bizerr.P("reasons", reasons),
	)
}

// summarizeGuardVetoReasons builds one deterministic reason string for bizerr
// params and audit logs.
func summarizeGuardVetoReasons(decisions []pluginhost.GuardDecision) string {
	items := make([]string, 0, len(decisions))
	for _, decision := range decisions {
		if decision.OK {
			continue
		}
		reason := strings.TrimSpace(decision.Reason)
		if reason == "" && decision.Err != nil {
			reason = decision.Err.Error()
		}
		if reason == "" {
			reason = "plugin." + strings.TrimSpace(decision.PluginID) + ".guard.vetoed"
		}
		items = append(items, strings.TrimSpace(decision.PluginID)+":"+reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}
