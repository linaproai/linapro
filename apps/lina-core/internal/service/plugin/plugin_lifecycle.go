// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// Install executes the install lifecycle and returns the dependency plan/result
// generated before target plugin side effects. It optionally persists one
// host-confirmed host service authorization snapshot when the target is a
// dynamic plugin.
//
// On a rolled-back mock-data load the plugin is fully installed (registry, menus,
// release state) — only the mock data was reverted. Install returns a stable
// bizerr (CodePluginInstallMockDataFailed) carrying pluginId, failedFile,
// rolledBackFiles, and cause so the caller can render a precise message.
func (s *serviceImpl) Install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (result *DependencyCheckResult, err error) {
	if err = s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	return s.install(ctx, pluginID, options)
}

// install executes plugin install side effects for platform-guarded callers
// and trusted startup reconciliation.
func (s *serviceImpl) install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (result *DependencyCheckResult, err error) {
	defer func() {
		err = wrapMockDataLoadError(err)
	}()
	return s.lifecycleSvc.Install(ctx, pluginID, lifecycle.InstallOptions{
		Authorization:    options.Authorization,
		InstallMode:      options.InstallMode,
		InstallMockData:  options.InstallMockData,
		FrameworkVersion: s.frameworkVersion(ctx),
	})
}

// Uninstall executes the uninstall lifecycle for an installed plugin using one explicit policy snapshot.
func (s *serviceImpl) Uninstall(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	return s.lifecycleSvc.Uninstall(ctx, pluginID, lifecycle.UninstallOptions{
		PurgeStorageData:    options.PurgeStorageData,
		Force:               options.Force,
		AllowForceUninstall: s.configSvc.GetPlugin(ctx).AllowForceUninstall,
	})
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
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
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
	return s.lifecycleSvc.UpdateStatus(ctx, pluginID, status, lifecycle.UpdateStatusOptions{
		Authorization:    authorization,
		FrameworkVersion: s.frameworkVersion(ctx),
	})
}

// Enable enables the specified plugin.
func (s *serviceImpl) Enable(ctx context.Context, pluginID string) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	return s.updateStatus(ctx, pluginID, plugintypes.StatusEnabled, nil)
}

// Disable disables the specified plugin.
func (s *serviceImpl) Disable(ctx context.Context, pluginID string) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	return s.updateStatus(ctx, pluginID, plugintypes.StatusDisabled, nil)
}

// persistDynamicPluginAuthorization refreshes the release snapshot for dynamic
// plugins so install/enable flows can reuse one governance preparation path.
func (s *serviceImpl) persistDynamicPluginAuthorization(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) error {
	if manifest == nil || plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic {
		return nil
	}
	if _, err := s.storeSvc.SyncManifest(ctx, manifest); err != nil {
		return err
	}
	if _, err := s.storeSvc.PersistReleaseHostServiceAuthorization(ctx, manifest, authorization); err != nil {
		return err
	}
	return nil
}

// IsInstalled returns whether a plugin is installed.
func (s *serviceImpl) IsInstalled(ctx context.Context, pluginID string) bool {
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	return err == nil && installed
}

// IsEnabled returns whether a plugin is enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "is_enabled")
	return s.integrationSvc.CanExposeBusinessEntries(ctx, pluginID)
}

// IsProviderEnabled returns whether pluginID is platform-enabled for framework
// capability provider use.
func (s *serviceImpl) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "provider_enabled")
	return s.integrationSvc.IsProviderEnabled(ctx, pluginID)
}

// IsEnabledAuthoritative returns whether pluginID is installed, enabled, and
// allowed to expose business entries after forcing a persisted registry read
// instead of reusing a process-local platform snapshot.
func (s *serviceImpl) IsEnabledAuthoritative(ctx context.Context, pluginID string) bool {
	ctx = integration.WithAuthoritativeEnablement(ctx)
	s.ensureRuntimeCacheFreshBestEffort(ctx, "is_enabled_authoritative")
	return s.integrationSvc.CanExposeBusinessEntries(ctx, pluginID)
}

// EnsureTenantPluginDisableAllowed runs source and dynamic lifecycle
// preconditions before one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	return s.lifecycleSvc.EnsureTenantPluginDisableAllowed(ctx, pluginID, tenantID)
}

// NotifyTenantPluginDisabled runs best-effort source and dynamic lifecycle
// callbacks after one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	s.lifecycleSvc.NotifyTenantPluginDisabled(ctx, pluginID, tenantID)
}

// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before tenant
// deletion continues in the tenant capability provider.
func (s *serviceImpl) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	return s.lifecycleSvc.EnsureTenantDeleteAllowed(ctx, tenantID)
}

// NotifyTenantDeleted runs best-effort source and dynamic lifecycle callbacks
// after a tenant has been deleted.
func (s *serviceImpl) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	s.lifecycleSvc.NotifyTenantDeleted(ctx, tenantID)
}

// ensureDynamicPluginActiveLifecyclePreconditionAllowed runs a dynamic plugin
// lifecycle precondition from the archived active release when that release is
// still readable.
func (s *serviceImpl) ensureDynamicPluginActiveLifecyclePreconditionAllowed(
	ctx context.Context,
	registry *store.PluginRecord,
	hook pluginhost.LifecycleHook,
	options UninstallOptions,
) error {
	if registry == nil ||
		plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
		registry.Installed != plugintypes.InstalledYes {
		return nil
	}
	manifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		if hook == pluginhost.LifecycleHookBeforeUninstall {
			manifest = s.loadSameVersionDesiredDynamicManifestBestEffort(registry)
			if manifest == nil {
				return nil
			}
			return s.ensureDynamicPluginLifecyclePreconditionAllowed(ctx, manifest, hook, options)
		}
		return s.dynamicLifecycleError(
			ctx,
			hook,
			registry.PluginId,
			[]runtime.DynamicLifecycleDecision{
				dynamicLifecycleFailureDecision(registry.PluginId, hook, err),
			},
			options.Force,
		)
	}
	return s.ensureDynamicPluginLifecyclePreconditionAllowed(ctx, manifest, hook, options)
}

// loadSameVersionDesiredDynamicManifestBestEffort returns the staged manifest
// only when it is the same version as the active dynamic release. This mirrors
// runtime uninstall repair, which may rebuild a missing active archive from the
// same-version staging artifact.
func (s *serviceImpl) loadSameVersionDesiredDynamicManifestBestEffort(registry *store.PluginRecord) *catalog.Manifest {
	if registry == nil {
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil ||
		manifest == nil ||
		plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic ||
		strings.TrimSpace(manifest.Version) != strings.TrimSpace(registry.Version) {
		return nil
	}
	return manifest
}

// loadActiveDynamicLifecycleManifestBestEffort returns the active dynamic
// manifest for best-effort post-lifecycle notifications.
func (s *serviceImpl) loadActiveDynamicLifecycleManifestBestEffort(ctx context.Context, pluginID string) *catalog.Manifest {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return nil
	}
	manifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		return nil
	}
	return manifest
}

// dynamicLifecycleFailureDecision creates a synthetic fail-closed decision when
// the host cannot even load the active dynamic plugin handler contract.
func dynamicLifecycleFailureDecision(
	pluginID string,
	hook pluginhost.LifecycleHook,
	err error,
) runtime.DynamicLifecycleDecision {
	return runtime.DynamicLifecycleDecision{
		PluginID:  pluginID,
		Operation: hook,
		OK:        false,
		Reason:    "plugin." + strings.TrimSpace(pluginID) + ".lifecycle." + hook.String() + ".failed",
		Err:       err,
	}
}

// ensureDynamicPluginLifecyclePreconditionAllowed runs one dynamic-plugin
// lifecycle precondition and converts vetoes to the shared lifecycle bizerr.
func (s *serviceImpl) ensureDynamicPluginLifecyclePreconditionAllowed(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	options UninstallOptions,
) error {
	if manifest == nil {
		return nil
	}
	decision, err := s.runtimeSvc.RunDynamicLifecyclePrecondition(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:         manifest.ID,
		Operation:        hook,
		PurgeStorageData: options.PurgeStorageData,
	})
	if decision == nil {
		return nil
	}
	decisions := []runtime.DynamicLifecycleDecision{*decision}
	if err != nil {
		return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, options.Force)
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, options.Force)
}

// executeDynamicPluginLifecycleNotification runs one dynamic After* lifecycle
// callback as a best-effort notification after the host transition succeeded.
func (s *serviceImpl) executeDynamicPluginLifecycleNotification(
	ctx context.Context,
	manifest *catalog.Manifest,
	input runtime.DynamicLifecycleInput,
) {
	if manifest == nil {
		return
	}
	if strings.TrimSpace(input.PluginID) == "" {
		input.PluginID = manifest.ID
	}
	if input.Operation == "" {
		return
	}
	decision, err := s.runtimeSvc.RunDynamicLifecycleCallback(ctx, manifest, input)
	if err == nil && (decision == nil || decision.OK) {
		return
	}
	decisions := make([]runtime.DynamicLifecycleDecision, 0, 1)
	if decision != nil {
		decisions = append(decisions, *decision)
	}
	if err != nil && len(decisions) == 0 {
		decisions = append(decisions, dynamicLifecycleFailureDecision(input.PluginID, input.Operation, err))
	}
	reasons := summarizeDynamicLifecycleVetoReasons(decisions)
	logger.Warningf(
		ctx,
		"dynamic plugin after lifecycle callback failed operation=%s plugin=%s reasons=%s err=%v",
		input.Operation,
		input.PluginID,
		reasons,
		err,
	)
}

// ensureLifecyclePreconditionAllowed runs source-plugin lifecycle
// preconditions before a protected plugin action and converts vetoes to stable
// caller-visible errors.
func (s *serviceImpl) ensureLifecyclePreconditionAllowed(
	ctx context.Context,
	pluginID string,
	hook pluginhost.LifecycleHook,
	force bool,
) error {
	pluginInput := pluginhost.NewSourcePluginLifecycleInput(pluginID, hook.String())
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         hook,
		PluginInput:  pluginInput,
		Participants: pluginhost.ListSourcePluginLifecycleParticipantsForPlugin(pluginID),
	})
	if result.OK {
		return nil
	}

	reasons := s.summarizeLocalizedLifecycleVetoReasons(ctx, result.Decisions)
	if force && hook == pluginhost.LifecycleHookBeforeUninstall {
		if err := s.ensureForceUninstallEnabled(ctx); err != nil {
			return err
		}
		logger.Warningf(
			ctx,
			"plugin lifecycle precondition force bypass operation=%s plugin=%s reasons=%s",
			hook,
			pluginID,
			reasons,
		)
		return nil
	}

	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", pluginID),
		bizerr.P("reasons", reasons),
	)
}

// dynamicLifecycleError converts dynamic lifecycle vetoes to the same shared
// caller-visible bizerr used by source-plugin lifecycle preconditions.
func (s *serviceImpl) dynamicLifecycleError(
	ctx context.Context,
	hook pluginhost.LifecycleHook,
	pluginID string,
	decisions []runtime.DynamicLifecycleDecision,
	force bool,
) error {
	reasons := s.summarizeLocalizedDynamicLifecycleVetoReasons(ctx, decisions)
	if force && hook == pluginhost.LifecycleHookBeforeUninstall {
		if err := s.ensureForceUninstallEnabled(ctx); err != nil {
			return err
		}
		logger.Warningf(
			ctx,
			"dynamic plugin lifecycle precondition force bypass operation=%s plugin=%s reasons=%s",
			hook,
			pluginID,
			reasons,
		)
		return nil
	}
	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", pluginID),
		bizerr.P("reasons", reasons),
	)
}

// ensureForceUninstallEnabled verifies that host configuration explicitly
// permits destructive force-uninstall flows.
func (s *serviceImpl) ensureForceUninstallEnabled(ctx context.Context) error {
	if !s.configSvc.GetPlugin(ctx).AllowForceUninstall {
		return bizerr.NewCode(CodePluginForceUninstallDisabled)
	}
	return nil
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
	includePluginPrefix := translate == nil || countLifecycleVetoes(decisions) > 1
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
			reason = "plugin." + strings.TrimSpace(decision.PluginID) + ".lifecycle.vetoed"
		}
		if translate != nil {
			if translated := strings.TrimSpace(translate(reason)); translated != "" {
				reason = translated
			}
		}
		pluginID := strings.TrimSpace(decision.PluginID)
		if includePluginPrefix && pluginID != "" {
			items = append(items, pluginID+":"+reason)
			continue
		}
		items = append(items, reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}

// summarizeDynamicLifecycleVetoReasons builds one deterministic raw reason
// string for dynamic lifecycle precondition results.
func summarizeDynamicLifecycleVetoReasons(decisions []runtime.DynamicLifecycleDecision) string {
	return summarizeDynamicLifecycleVetoReasonsWithTranslator(decisions, nil)
}

// summarizeLocalizedDynamicLifecycleVetoReasons builds one deterministic
// localized reason string for caller-visible dynamic lifecycle errors.
func (s *serviceImpl) summarizeLocalizedDynamicLifecycleVetoReasons(
	ctx context.Context,
	decisions []runtime.DynamicLifecycleDecision,
) string {
	return summarizeDynamicLifecycleVetoReasonsWithTranslator(decisions, func(key string) string {
		if s == nil || s.i18nSvc == nil {
			return ""
		}
		return s.i18nSvc.Translate(ctx, key, "")
	})
}

// summarizeDynamicLifecycleVetoReasonsWithTranslator applies an optional
// translator to dynamic lifecycle reason keys.
func summarizeDynamicLifecycleVetoReasonsWithTranslator(
	decisions []runtime.DynamicLifecycleDecision,
	translate func(key string) string,
) string {
	includePluginPrefix := translate == nil || countDynamicLifecycleVetoes(decisions) > 1
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
			reason = "plugin." + strings.TrimSpace(decision.PluginID) + ".lifecycle.vetoed"
		}
		if translate != nil {
			if translated := strings.TrimSpace(translate(reason)); translated != "" {
				reason = translated
			}
		}
		pluginID := strings.TrimSpace(decision.PluginID)
		if includePluginPrefix && pluginID != "" {
			items = append(items, pluginID+":"+reason)
			continue
		}
		items = append(items, reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}

// countLifecycleVetoes returns how many source lifecycle decisions blocked the action.
func countLifecycleVetoes(decisions []pluginhost.LifecycleDecision) int {
	count := 0
	for _, decision := range decisions {
		if !decision.OK {
			count++
		}
	}
	return count
}

// countDynamicLifecycleVetoes returns how many dynamic lifecycle decisions blocked the action.
func countDynamicLifecycleVetoes(decisions []runtime.DynamicLifecycleDecision) int {
	count := 0
	for _, decision := range decisions {
		if !decision.OK {
			count++
		}
	}
	return count
}

// dynamicLifecycleDecisionsAllowed reports whether all dynamic decisions allowed the action.
func dynamicLifecycleDecisionsAllowed(decisions []runtime.DynamicLifecycleDecision) bool {
	for _, decision := range decisions {
		if !decision.OK {
			return false
		}
	}
	return true
}
