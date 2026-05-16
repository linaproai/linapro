// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// Install executes the install lifecycle and returns the dependency plan/result
// generated before target plugin side effects. It optionally persists one
// host-confirmed host service authorization snapshot when the target is a dynamic
// plugin. The options.InstallMockData flag is threaded through context so deeply
// nested runtime/reconciler code can detect mock opt-in without mass signature
// changes.
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
	ctx = withInstallMockData(ctx, options.InstallMockData)
	defer func() {
		err = wrapMockDataLoadError(err)
	}()

	result, ctx, err = s.prepareInstallDependencies(ctx, pluginID, options)
	if err != nil {
		return result, err
	}
	options.dependencyResult = result

	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return result, err
	}
	if err = applyInstallModeSelection(manifest, options.InstallMode); err != nil {
		return result, err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		if err = s.installSourcePlugin(ctx, manifest); err != nil {
			if !isMockDataLoadError(err) {
				return result, err
			}
			if snapshotErr := s.syncEnabledSnapshotFromRegistry(ctx, pluginID); snapshotErr != nil {
				return result, snapshotErr
			}
			if markErr := s.markRuntimeCacheChanged(ctx, "source_plugin_installed"); markErr != nil {
				return result, markErr
			}
			return result, err
		}
		if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
			return result, err
		}
		if err = s.markRuntimeCacheChanged(ctx, "source_plugin_installed"); err != nil {
			return result, err
		}
		if err = notifyPluginInstalled(ctx, pluginID); err != nil {
			return result, err
		}
		s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterInstall)
		return result, nil
	}
	if err = s.ensureDynamicPluginInstallLifecyclePreconditionAllowed(ctx, manifest, options.Authorization); err != nil {
		return result, err
	}
	if err = s.persistDynamicPluginAuthorization(ctx, manifest, options.Authorization); err != nil {
		return result, err
	}
	if err = s.lifecycleSvc.Install(ctx, pluginID); err != nil {
		return result, err
	}
	// Dynamic lifecycle reloads the manifest from the runtime artifact. Re-sync
	// the operator-selected governance fields so installMode cannot be reset to
	// the artifact default after the request has already been validated.
	if _, err = s.catalogSvc.SyncManifest(ctx, manifest); err != nil {
		return result, err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return result, err
	}
	if err = notifyPluginInstalled(ctx, pluginID); err != nil {
		return result, err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:  manifest.ID,
		Operation: pluginhost.LifecycleHookAfterInstall,
	})
	return result, nil
}

// ensureDynamicPluginInstallLifecyclePreconditionAllowed runs BeforeInstall
// with the same host-service authorization snapshot that install will persist.
func (s *serviceImpl) ensureDynamicPluginInstallLifecyclePreconditionAllowed(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) error {
	authorizedManifest, err := cloneManifestWithAuthorizedHostServices(manifest, authorization)
	if err != nil {
		return err
	}
	return s.ensureDynamicPluginLifecyclePreconditionAllowed(
		ctx,
		authorizedManifest,
		pluginhost.LifecycleHookBeforeInstall,
		false,
	)
}

// cloneManifestWithAuthorizedHostServices applies one operation-local
// host-service authorization decision to a shallow manifest clone.
func cloneManifestWithAuthorizedHostServices(
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) (*catalog.Manifest, error) {
	if manifest == nil {
		return nil, nil
	}
	hostServices, err := buildLifecycleAuthorizedHostServices(manifest.HostServices, authorization)
	if err != nil {
		return nil, err
	}
	clone := *manifest
	clone.HostServices = hostServices
	clone.HostCapabilities = pluginbridge.CapabilityMapFromHostServices(hostServices)
	return &clone, nil
}

// buildLifecycleAuthorizedHostServices narrows lifecycle bridge execution to
// operation-confirmed host services. When no confirmation is provided, only
// capability-only services are exposed.
func buildLifecycleAuthorizedHostServices(
	hostServices []*pluginbridge.HostServiceSpec,
	authorization *HostServiceAuthorizationInput,
) ([]*pluginbridge.HostServiceSpec, error) {
	if authorization != nil {
		return catalog.BuildAuthorizedHostServiceSpecs(hostServices, authorization)
	}
	requested, err := pluginbridge.NormalizeHostServiceSpecs(hostServices)
	if err != nil {
		return nil, err
	}
	authorized := make([]*pluginbridge.HostServiceSpec, 0, len(requested))
	for _, spec := range requested {
		if spec == nil || len(spec.Paths) > 0 || len(spec.Resources) > 0 || len(spec.Tables) > 0 {
			continue
		}
		authorized = append(authorized, spec)
	}
	return pluginbridge.NormalizeHostServiceSpecs(authorized)
}

// applyInstallModeSelection validates the explicit install-mode request and
// applies it to the short-lived desired manifest before registry synchronization.
func applyInstallModeSelection(manifest *catalog.Manifest, installMode string) error {
	if manifest == nil {
		return nil
	}
	scopeNature := catalog.NormalizeScopeNature(manifest.ScopeNature)
	if strings.TrimSpace(installMode) != "" && !catalog.IsSupportedInstallMode(installMode) {
		return bizerr.NewCode(CodePluginInstallModeInvalid)
	}
	if !manifest.SupportsTenantGovernance() {
		manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
		if strings.TrimSpace(installMode) != "" && catalog.NormalizeInstallMode(installMode) != catalog.InstallModeGlobal {
			return bizerr.NewCode(
				CodePluginInstallModeInvalidForScopeNature,
				bizerr.P("scopeNature", scopeNature.String()),
				bizerr.P("installMode", catalog.NormalizeInstallMode(installMode).String()),
			)
		}
		return nil
	}
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

// Uninstall executes the uninstall lifecycle for an installed plugin using one explicit policy snapshot.
func (s *serviceImpl) Uninstall(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return s.uninstallWithoutDesiredManifest(ctx, pluginID, options, err)
	}
	if err = s.ensureNoReverseDependencies(ctx, pluginID); err != nil {
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
		if err = notifyPluginUninstalled(ctx, pluginID); err != nil {
			return err
		}
		s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterUninstall)
		return nil
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if err = s.ensureDynamicPluginActiveLifecyclePreconditionAllowed(ctx, registry, pluginhost.LifecycleHookBeforeUninstall, options.Force); err != nil {
		return err
	}
	activeManifest := s.loadActiveDynamicLifecycleManifestBestEffort(ctx, pluginID)
	if err = s.uninstallDynamicPlugin(ctx, pluginID, options); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	if err = notifyPluginUninstalled(ctx, pluginID); err != nil {
		return err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:  pluginID,
		Operation: pluginhost.LifecycleHookAfterUninstall,
	})
	return nil
}

// uninstallWithoutDesiredManifest keeps dynamic-plugin uninstall recoverable
// when the mutable staging artifact is missing but the registry still carries
// enough active-release state to complete or force one uninstall.
func (s *serviceImpl) uninstallWithoutDesiredManifest(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
	discoveryErr error,
) error {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return discoveryErr
	}
	if err = s.ensureNoReverseDependencies(ctx, pluginID); err != nil {
		return err
	}
	if err = s.ensureDynamicPluginActiveLifecyclePreconditionAllowed(ctx, registry, pluginhost.LifecycleHookBeforeUninstall, options.Force); err != nil {
		return err
	}
	activeManifest := s.loadActiveDynamicLifecycleManifestBestEffort(ctx, pluginID)
	if err = s.uninstallDynamicPlugin(ctx, pluginID, options); err != nil {
		return err
	}
	if err = s.syncEnabledSnapshotFromRegistry(ctx, pluginID); err != nil {
		return err
	}
	if err = notifyPluginUninstalled(ctx, pluginID); err != nil {
		return err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:  pluginID,
		Operation: pluginhost.LifecycleHookAfterUninstall,
	})
	return nil
}

// uninstallDynamicPlugin chooses between the full active-release uninstall and
// the restricted orphan cleanup path used only when active dynamic artifacts are
// no longer readable.
func (s *serviceImpl) uninstallDynamicPlugin(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData)
	}
	if registry.Installed != catalog.InstalledYes {
		if options.Force {
			return s.forceUninstallMissingDynamicArtifact(ctx, registry)
		}
		return s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData)
	}
	if s.dynamicFullUninstallRecoverable(ctx, registry) {
		if err = s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData); err == nil {
			return nil
		}
		refreshed, refreshErr := s.catalogSvc.GetRegistry(ctx, pluginID)
		if refreshErr != nil {
			return refreshErr
		}
		if !options.Force || s.dynamicFullUninstallRecoverable(ctx, refreshed) {
			return err
		}
		registry = refreshed
	}
	if !options.Force {
		return bizerr.NewCode(
			CodePluginDynamicArtifactMissingForUninstall,
			bizerr.P("pluginId", pluginID),
		)
	}
	return s.forceUninstallMissingDynamicArtifact(ctx, registry)
}

// dynamicFullUninstallRecoverable reports whether the installed dynamic plugin
// can run full uninstall from its archived active release or repair that archive
// from the current same-version staging artifact.
func (s *serviceImpl) dynamicFullUninstallRecoverable(ctx context.Context, registry *entity.SysPlugin) bool {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return false
	}
	manifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err == nil && manifest != nil {
		return true
	}
	desiredManifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil || desiredManifest == nil {
		return false
	}
	if catalog.NormalizeType(desiredManifest.Type) != catalog.TypeDynamic {
		return false
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		return false
	}
	return desiredManifest.RuntimeArtifact != nil
}

// forceUninstallMissingDynamicArtifact validates host policy before clearing
// host-owned orphan governance for a dynamic plugin with unreadable artifacts.
func (s *serviceImpl) forceUninstallMissingDynamicArtifact(ctx context.Context, registry *entity.SysPlugin) error {
	if err := s.ensureForceUninstallEnabled(ctx); err != nil {
		return err
	}
	return s.runtimeSvc.ForceUninstallMissingArtifact(ctx, registry)
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
		} else {
			registry, registryErr := s.catalogSvc.GetRegistry(ctx, pluginID)
			if registryErr != nil {
				return registryErr
			}
			if err = s.ensureDynamicPluginActiveLifecyclePreconditionAllowed(
				ctx,
				registry,
				pluginhost.LifecycleHookBeforeDisable,
				false,
			); err != nil {
				return err
			}
		}
		var activeManifest *catalog.Manifest
		if status == catalog.StatusDisabled {
			activeManifest = s.loadActiveDynamicLifecycleManifestBestEffort(ctx, pluginID)
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
		if err = notifyPluginDisabled(ctx, pluginID); err != nil {
			return err
		}
		s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
			PluginID:  pluginID,
			Operation: pluginhost.LifecycleHookAfterDisable,
		})
		return nil
	}
	if status == catalog.StatusDisabled {
		if err = s.executeSourcePluginBeforeLifecycle(ctx, manifest, pluginhost.LifecycleHookBeforeDisable, false); err != nil {
			return err
		}
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
	if err = notifyPluginDisabled(ctx, pluginID); err != nil {
		return err
	}
	s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterDisable)
	return nil
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

// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before tenant
// deletion continues in the tenant capability provider.
func (s *serviceImpl) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	if err := s.ensureTenantLifecyclePreconditionAllowed(ctx, tenantID, pluginhost.LifecycleHookBeforeTenantDelete); err != nil {
		return err
	}
	return s.ensureDynamicTenantLifecyclePreconditionAllowed(ctx, tenantID, pluginhost.LifecycleHookBeforeTenantDelete)
}

// ensureTenantLifecyclePreconditionAllowed runs tenant-scoped lifecycle
// preconditions and converts vetoes to the same stable lifecycle error used by
// plugin disable and uninstall operations.
func (s *serviceImpl) ensureTenantLifecyclePreconditionAllowed(
	ctx context.Context,
	tenantID int,
	hook pluginhost.LifecycleHook,
) error {
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         hook,
		TenantInput:  pluginhost.NewSourcePluginTenantLifecycleInput(hook.String(), tenantID),
		Participants: pluginhost.ListSourcePluginLifecycleParticipants(),
	})
	if result.OK {
		return nil
	}

	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", "tenant"),
		bizerr.P("reasons", summarizeLifecycleVetoReasons(result.Decisions)),
	)
}

// ensureDynamicTenantLifecyclePreconditionAllowed runs dynamic-plugin
// tenant-scoped lifecycle preconditions before tenant deletion continues.
func (s *serviceImpl) ensureDynamicTenantLifecyclePreconditionAllowed(
	ctx context.Context,
	tenantID int,
	hook pluginhost.LifecycleHook,
) error {
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return err
	}
	decisions := make([]runtime.DynamicLifecycleDecision, 0)
	for _, registry := range registries {
		if registry == nil ||
			catalog.NormalizeType(registry.Type) != catalog.TypeDynamic ||
			registry.Installed != catalog.InstalledYes ||
			registry.Status != catalog.StatusEnabled {
			continue
		}
		activeManifest, activeErr := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
		if activeErr != nil {
			return s.dynamicLifecycleError(
				ctx,
				hook,
				registry.PluginId,
				[]runtime.DynamicLifecycleDecision{
					dynamicLifecycleFailureDecision(registry.PluginId, hook, activeErr),
				},
				false,
			)
		}
		if activeManifest == nil {
			continue
		}
		decision, runErr := s.runtimeSvc.RunDynamicLifecyclePrecondition(ctx, activeManifest, runtime.DynamicLifecycleInput{
			PluginID:  activeManifest.ID,
			Operation: hook,
			TenantID:  tenantID,
		})
		if decision != nil {
			decisions = append(decisions, *decision)
		}
		if runErr != nil {
			return s.dynamicLifecycleError(ctx, hook, activeManifest.ID, decisions, false)
		}
	}
	if dynamicLifecycleDecisionsAllowed(decisions) {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, "tenant", decisions, false)
}

// ensureDynamicPluginActiveLifecyclePreconditionAllowed runs a dynamic plugin
// lifecycle precondition from the archived active release when that release is
// still readable.
func (s *serviceImpl) ensureDynamicPluginActiveLifecyclePreconditionAllowed(
	ctx context.Context,
	registry *entity.SysPlugin,
	hook pluginhost.LifecycleHook,
	force bool,
) error {
	if registry == nil ||
		catalog.NormalizeType(registry.Type) != catalog.TypeDynamic ||
		registry.Installed != catalog.InstalledYes {
		return nil
	}
	manifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		if hook == pluginhost.LifecycleHookBeforeUninstall {
			manifest = s.loadSameVersionDesiredDynamicManifestBestEffort(registry)
			if manifest == nil {
				return nil
			}
			return s.ensureDynamicPluginLifecyclePreconditionAllowed(ctx, manifest, hook, force)
		}
		return s.dynamicLifecycleError(
			ctx,
			hook,
			registry.PluginId,
			[]runtime.DynamicLifecycleDecision{
				dynamicLifecycleFailureDecision(registry.PluginId, hook, err),
			},
			force,
		)
	}
	return s.ensureDynamicPluginLifecyclePreconditionAllowed(ctx, manifest, hook, force)
}

// loadSameVersionDesiredDynamicManifestBestEffort returns the staged manifest
// only when it is the same version as the active dynamic release. This mirrors
// runtime uninstall repair, which may rebuild a missing active archive from the
// same-version staging artifact.
func (s *serviceImpl) loadSameVersionDesiredDynamicManifestBestEffort(registry *entity.SysPlugin) *catalog.Manifest {
	if registry == nil {
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(registry.PluginId)
	if err != nil ||
		manifest == nil ||
		catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic ||
		strings.TrimSpace(manifest.Version) != strings.TrimSpace(registry.Version) {
		return nil
	}
	return manifest
}

// loadActiveDynamicLifecycleManifestBestEffort returns the active dynamic
// manifest for best-effort post-lifecycle notifications.
func (s *serviceImpl) loadActiveDynamicLifecycleManifestBestEffort(ctx context.Context, pluginID string) *catalog.Manifest {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
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

// ensureDynamicPluginUpgradeLifecyclePreconditionAllowed runs BeforeUpgrade for
// dynamic plugins before upgrade state markers or release switch side effects.
func (s *serviceImpl) ensureDynamicPluginUpgradeLifecyclePreconditionAllowed(
	ctx context.Context,
	registry *entity.SysPlugin,
	targetManifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) error {
	if registry == nil || targetManifest == nil {
		return nil
	}
	manifest, err := s.applyTargetReleaseAuthorizedHostServices(ctx, targetManifest, authorization)
	if err != nil {
		return err
	}
	decision, err := s.runtimeSvc.RunDynamicLifecyclePrecondition(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:    targetManifest.ID,
		Operation:   pluginhost.LifecycleHookBeforeUpgrade,
		FromVersion: strings.TrimSpace(registry.Version),
		ToVersion:   strings.TrimSpace(targetManifest.Version),
	})
	if decision == nil {
		return nil
	}
	decisions := []runtime.DynamicLifecycleDecision{*decision}
	if err != nil {
		return s.dynamicLifecycleError(ctx, pluginhost.LifecycleHookBeforeUpgrade, targetManifest.ID, decisions, false)
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, pluginhost.LifecycleHookBeforeUpgrade, targetManifest.ID, decisions, false)
}

// executeDynamicPluginUpgradeLifecycleNotification runs AfterUpgrade for a
// dynamic plugin after the target release has become effective.
func (s *serviceImpl) executeDynamicPluginUpgradeLifecycleNotification(
	ctx context.Context,
	registry *entity.SysPlugin,
	targetManifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) {
	if registry == nil || targetManifest == nil {
		return
	}
	manifest, err := s.applyTargetReleaseAuthorizedHostServices(ctx, targetManifest, authorization)
	if err != nil {
		logger.Warningf(
			ctx,
			"dynamic plugin after lifecycle authorization snapshot failed operation=%s plugin=%s err=%v",
			pluginhost.LifecycleHookAfterUpgrade,
			targetManifest.ID,
			err,
		)
		return
	}
	s.executeDynamicPluginLifecycleNotification(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:    targetManifest.ID,
		Operation:   pluginhost.LifecycleHookAfterUpgrade,
		FromVersion: strings.TrimSpace(registry.Version),
		ToVersion:   strings.TrimSpace(targetManifest.Version),
	})
}

// applyTargetReleaseAuthorizedHostServices overlays the target release's
// already-confirmed host-service snapshot when it exists, keeping BeforeUpgrade
// execution aligned with the bridge authorization that will become effective.
func (s *serviceImpl) applyTargetReleaseAuthorizedHostServices(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *HostServiceAuthorizationInput,
) (*catalog.Manifest, error) {
	if manifest == nil {
		return nil, nil
	}
	if authorization != nil {
		return cloneManifestWithAuthorizedHostServices(manifest, authorization)
	}
	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return cloneManifestWithAuthorizedHostServices(manifest, nil)
	}
	snapshot, err := s.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || !snapshot.HostServiceAuthRequired || !snapshot.HostServiceAuthConfirmed {
		return cloneManifestWithAuthorizedHostServices(manifest, nil)
	}
	hostServices, err := pluginbridge.NormalizeHostServiceSpecs(snapshot.AuthorizedHostServices)
	if err != nil {
		return nil, err
	}
	clone := *manifest
	clone.HostServices = hostServices
	clone.HostCapabilities = pluginbridge.CapabilityMapFromHostServices(hostServices)
	return &clone, nil
}

// ensureDynamicPluginLifecyclePreconditionAllowed runs one dynamic-plugin
// lifecycle precondition and converts vetoes to the shared lifecycle bizerr.
func (s *serviceImpl) ensureDynamicPluginLifecyclePreconditionAllowed(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	force bool,
) error {
	if manifest == nil {
		return nil
	}
	decision, err := s.runtimeSvc.RunDynamicLifecyclePrecondition(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:  manifest.ID,
		Operation: hook,
	})
	if decision == nil {
		return nil
	}
	decisions := []runtime.DynamicLifecycleDecision{*decision}
	if err != nil {
		return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, force)
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, force)
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

	reasons := summarizeLifecycleVetoReasons(result.Decisions)
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
	reasons := summarizeDynamicLifecycleVetoReasons(decisions)
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

// summarizeLifecycleVetoReasons builds one deterministic reason string for
// bizerr params and audit logs.
func summarizeLifecycleVetoReasons(decisions []pluginhost.LifecycleDecision) string {
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
		items = append(items, strings.TrimSpace(decision.PluginID)+":"+reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}

// summarizeDynamicLifecycleVetoReasons builds one deterministic reason string
// for dynamic lifecycle precondition results.
func summarizeDynamicLifecycleVetoReasons(decisions []runtime.DynamicLifecycleDecision) string {
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
		items = append(items, strings.TrimSpace(decision.PluginID)+":"+reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
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
