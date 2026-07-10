// This file owns plugin install orchestration for source and dynamic plugins,
// including dependency checks, install-mode selection, lifecycle callbacks, and
// cache/observer publication.

package lifecycle

import (
	"context"
	"errors"
	pluginv1 "lina-core/api/plugin/v1"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// InstallOptions captures per-request install decisions needed by lifecycle.
type InstallOptions struct {
	// Authorization optionally carries a host-service authorization snapshot for
	// dynamic plugins that require explicit confirmation before install.
	Authorization *store.HostServiceAuthorizationInput
	// InstallMode optionally carries the platform operator's explicit tenant
	// governance selection. Empty means use the plugin manifest default.
	InstallMode string
	// InstallMockData enables the optional mock-data load phase.
	InstallMockData bool
	// StartupAutoEnable marks install requests initiated by plugin.autoEnable.
	StartupAutoEnable bool
	// FrameworkVersion is the current LinaPro framework version used for
	// dependency compatibility checks.
	FrameworkVersion string
}

// LifecycleObserver receives synchronous plugin lifecycle callbacks from the
// host plugin service.
type LifecycleObserver interface {
	// OnPluginInstalled handles one successful plugin install transition.
	OnPluginInstalled(ctx context.Context, pluginID string) error
	// OnPluginEnabled handles one successful plugin enable transition.
	OnPluginEnabled(ctx context.Context, pluginID string) error
	// OnPluginDisabled handles one successful plugin disable transition.
	OnPluginDisabled(ctx context.Context, pluginID string) error
	// OnPluginUninstalled handles one successful plugin uninstall transition.
	OnPluginUninstalled(ctx context.Context, pluginID string) error
}

// Install executes the install lifecycle for a discovered source or dynamic
// plugin and returns the dependency check result produced before side effects.
func (s *serviceImpl) Install(
	ctx context.Context,
	pluginID string,
	options InstallOptions,
) (result *plugindep.CheckProjection, err error) {
	s.catalogSvc.InvalidateManifestCache(pluginID)
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return result, err
	}
	if err = applyInstallModeSelection(manifest, options.InstallMode); err != nil {
		return result, err
	}

	result, err = s.prepareInstallDependencies(ctx, manifest, options.FrameworkVersion)
	if err != nil {
		return result, err
	}

	if plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeSource {
		if err = s.installSourcePlugin(ctx, manifest, SourceInstallOptions{
			StartupAutoEnable: options.StartupAutoEnable,
			InstallMockData:   options.InstallMockData,
		}); err != nil {
			if !isMockDataLoadError(err) {
				return result, err
			}
			if syncErr := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "source_plugin_installed"); syncErr != nil {
				return result, syncErr
			}
			return result, err
		}
		if err = s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "source_plugin_installed"); err != nil {
			return result, err
		}
		if err = s.notifyPluginInstalled(ctx, pluginID); err != nil {
			return result, err
		}
		s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterInstall, sourceLifecyclePolicy{})
		return result, nil
	}

	if err = s.ensureDynamicPluginInstallLifecyclePreconditionAllowed(ctx, manifest, options.Authorization); err != nil {
		return result, err
	}
	if err = s.persistDynamicPluginAuthorization(ctx, manifest, options.Authorization); err != nil {
		return result, err
	}
	if err = s.installDynamic(ctx, pluginID, manifest, runtime.DynamicReconcileOptions{
		InstallMockData: options.InstallMockData,
	}); err != nil {
		return result, err
	}
	if err = s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "dynamic_plugin_installed"); err != nil {
		return result, err
	}
	if err = s.notifyPluginInstalled(ctx, pluginID); err != nil {
		return result, err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, manifest, runtime.DynamicLifecycleInput{
		PluginID:  manifest.ID,
		Operation: pluginhost.LifecycleHookAfterInstall,
	})
	return result, nil
}

// applyInstallModeSelection validates the explicit install-mode request and
// applies it to the short-lived desired manifest before registry synchronization.
func applyInstallModeSelection(manifest *catalog.Manifest, installMode string) error {
	if manifest == nil {
		return nil
	}
	scopeNature := plugintypes.NormalizeScopeNature(manifest.ScopeNature)
	if strings.TrimSpace(installMode) != "" && !plugintypes.IsSupportedInstallMode(installMode) {
		return bizerr.NewCode(CodePluginInstallModeInvalid)
	}
	if !manifest.SupportsTenantGovernance() {
		manifest.DefaultInstallMode = pluginv1.InstallModeGlobal.String()
		if strings.TrimSpace(installMode) != "" && plugintypes.NormalizeInstallMode(installMode) != pluginv1.InstallModeGlobal {
			return bizerr.NewCode(
				CodePluginInstallModeInvalidForScopeNature,
				bizerr.P("scopeNature", scopeNature.String()),
				bizerr.P("installMode", plugintypes.NormalizeInstallMode(installMode).String()),
			)
		}
		return nil
	}
	if strings.TrimSpace(installMode) == "" {
		installMode = manifest.DefaultInstallMode
	}
	if !plugintypes.IsSupportedInstallMode(installMode) {
		return bizerr.NewCode(CodePluginInstallModeInvalid)
	}
	mode := plugintypes.NormalizeInstallMode(installMode)
	if scopeNature == pluginv1.ScopeNaturePlatformOnly && mode != pluginv1.InstallModeGlobal {
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

// prepareInstallDependencies verifies a target before lifecycle side effects.
func (s *serviceImpl) prepareInstallDependencies(
	ctx context.Context,
	manifest *catalog.Manifest,
	frameworkVersion string,
) (*plugindep.CheckProjection, error) {
	if manifest == nil {
		return nil, nil
	}
	normalizedID := strings.TrimSpace(manifest.ID)
	if normalizedID == "" {
		return nil, nil
	}
	check, err := s.resolveInstallDependenciesForManifest(ctx, manifest, frameworkVersion)
	if err != nil {
		return nil, err
	}
	result := plugindep.ToCheckProjection(check)
	if plugindep.HasBlockers(check.Blockers) {
		return result, buildDependencyBlockedError(normalizedID, check.Blockers)
	}
	return result, nil
}

// resolveInstallDependencies evaluates dependency status for one discovered target.
func (s *serviceImpl) resolveInstallDependencies(
	ctx context.Context,
	pluginID string,
	frameworkVersion string,
) (*plugindep.InstallCheckResult, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if manifest := management.ManifestByIDFromContext(ctx, normalizedPluginID); manifest != nil {
		return s.resolveInstallDependenciesForManifest(ctx, manifest, frameworkVersion)
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(normalizedPluginID)
	if err != nil {
		return nil, err
	}
	return s.resolveInstallDependenciesForManifest(ctx, manifest, frameworkVersion)
}

// resolveInstallDependenciesForManifest evaluates dependency status using a
// candidate manifest override.
func (s *serviceImpl) resolveInstallDependenciesForManifest(
	ctx context.Context,
	manifest *catalog.Manifest,
	frameworkVersion string,
) (*plugindep.InstallCheckResult, error) {
	snapshots, err := s.buildDependencySnapshots(ctx, manifest)
	if err != nil {
		return nil, err
	}
	resolver := s.dependencyResolver
	if resolver == nil {
		resolver = plugindep.New()
	}
	return resolver.CheckInstall(plugindep.InstallCheckInput{
		TargetID:         strings.TrimSpace(manifest.ID),
		FrameworkVersion: strings.TrimSpace(frameworkVersion),
		Plugins:          snapshots,
	}), nil
}

// buildDependencySnapshots combines discovered manifests with installed
// registry release snapshots.
func (s *serviceImpl) buildDependencySnapshots(
	ctx context.Context,
	candidate *catalog.Manifest,
) ([]*plugindep.PluginSnapshot, error) {
	manifests := management.ManifestSnapshotFromContext(ctx)
	if manifests == nil {
		var err error
		manifests, err = s.catalogSvc.ScanManifests()
		if err != nil {
			return nil, err
		}
	}
	snapshotByID := make(map[string]*plugindep.PluginSnapshot, len(manifests)+1)
	for _, manifest := range manifests {
		if manifest == nil || strings.TrimSpace(manifest.ID) == "" {
			continue
		}
		snapshotByID[manifest.ID] = &plugindep.PluginSnapshot{
			ID:           strings.TrimSpace(manifest.ID),
			Name:         strings.TrimSpace(manifest.Name),
			Version:      strings.TrimSpace(manifest.Version),
			Manifest:     manifest,
			Dependencies: plugintypes.CloneDependencySpec(manifest.Dependencies),
		}
	}
	if candidate != nil && strings.TrimSpace(candidate.ID) != "" {
		snapshotByID[candidate.ID] = &plugindep.PluginSnapshot{
			ID:           strings.TrimSpace(candidate.ID),
			Name:         strings.TrimSpace(candidate.Name),
			Version:      strings.TrimSpace(candidate.Version),
			Manifest:     candidate,
			Dependencies: plugintypes.CloneDependencySpec(candidate.Dependencies),
		}
	}

	readCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	registries, err := s.storeSvc.ListAllRegistries(readCtx)
	if err != nil {
		return nil, err
	}
	candidateID := ""
	if candidate != nil {
		candidateID = strings.TrimSpace(candidate.ID)
	}
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		registryPluginID := strings.TrimSpace(registry.PluginId)
		if registryPluginID == "" {
			continue
		}
		snapshot := snapshotByID[registryPluginID]
		if snapshot == nil {
			if registry.ReleaseId <= 0 {
				continue
			}
			snapshot = &plugindep.PluginSnapshot{ID: registryPluginID}
			snapshotByID[registryPluginID] = snapshot
		}
		if registryPluginID == candidateID {
			snapshot.Installed = registry.Installed == statusflag.Installed.Int()
			continue
		}
		plugindep.ApplyRegistrySnapshot(readCtx, s.storeSvc, snapshot, registry)
	}

	out := make([]*plugindep.PluginSnapshot, 0, len(snapshotByID))
	for _, snapshot := range snapshotByID {
		out = append(out, snapshot)
	}
	return out, nil
}

// buildDependencyBlockedError converts resolver blockers into one structured business error.
func buildDependencyBlockedError(pluginID string, blockers []*plugindep.Blocker) error {
	dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(blockers)
	return bizerr.NewCode(
		CodePluginDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("chain", plugindep.FirstBlockerChain(blockers)),
		bizerr.P("blockers", plugindep.FormatBlockers(blockers)),
	)
}

// ensureDynamicPluginInstallLifecyclePreconditionAllowed runs BeforeInstall
// with the same host-service authorization snapshot that install will persist.
func (s *serviceImpl) ensureDynamicPluginInstallLifecyclePreconditionAllowed(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
) error {
	authorizedManifest, err := cloneManifestWithAuthorizedHostServices(manifest, authorization)
	if err != nil {
		return err
	}
	return s.ensureDynamicPluginLifecyclePreconditionAllowed(
		ctx,
		authorizedManifest,
		pluginhost.LifecycleHookBeforeInstall,
		UninstallOptions{},
	)
}

// cloneManifestWithAuthorizedHostServices applies one operation-local
// host-service authorization decision to a shallow manifest clone.
func cloneManifestWithAuthorizedHostServices(
	manifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
) (*catalog.Manifest, error) {
	if manifest == nil {
		return nil, nil
	}
	hostServices, err := buildLifecycleAuthorizedHostServices(manifest.ID, manifest.HostServices, authorization)
	if err != nil {
		return nil, err
	}
	clone := *manifest
	clone.HostServices = hostServices
	clone.HostCapabilities = protocol.CapabilityMapFromHostServices(hostServices)
	return &clone, nil
}

// buildLifecycleAuthorizedHostServices narrows lifecycle bridge execution to
// operation-confirmed host services. When no confirmation is provided, only
// capability-only services are exposed.
func buildLifecycleAuthorizedHostServices(
	pluginID string,
	hostServices []*protocol.HostServiceSpec,
	authorization *store.HostServiceAuthorizationInput,
) ([]*protocol.HostServiceSpec, error) {
	if authorization != nil {
		return store.BuildAuthorizedHostServiceSpecsForPlugin(pluginID, hostServices, authorization)
	}
	requested, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, hostServices)
	if err != nil {
		return nil, err
	}
	authorized := make([]*protocol.HostServiceSpec, 0, len(requested))
	for _, spec := range requested {
		if spec == nil || len(spec.Paths) > 0 || len(spec.Resources) > 0 || len(spec.Tables) > 0 || len(spec.Keys) > 0 {
			continue
		}
		authorized = append(authorized, spec)
	}
	return protocol.NormalizeHostServiceSpecsForPlugin(pluginID, authorized)
}

// persistDynamicPluginAuthorization refreshes the release snapshot for dynamic
// plugins so install/enable flows can reuse one governance preparation path.
func (s *serviceImpl) persistDynamicPluginAuthorization(
	ctx context.Context,
	manifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
) error {
	if manifest == nil || plugintypes.NormalizeType(manifest.Type) != pluginv1.PluginTypeDynamic {
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
		PluginID:  manifest.ID,
		Operation: hook,
	})
	if decision == nil {
		return nil
	}
	decisions := []runtime.DynamicLifecycleDecision{*decision}
	if err != nil {
		return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, options)
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, manifest.ID, decisions, options)
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

// dynamicLifecycleFailureDecision creates a synthetic fail-closed decision when
// the host cannot even load a dynamic plugin handler contract.
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

// summarizeDynamicLifecycleVetoReasons builds one deterministic raw reason
// string for dynamic lifecycle precondition results.
func summarizeDynamicLifecycleVetoReasons(decisions []runtime.DynamicLifecycleDecision) string {
	return summarizeDynamicLifecycleVetoReasonsWithTranslator(decisions, nil)
}

// summarizeDynamicLifecycleVetoReasonsWithTranslator applies an optional
// translator to dynamic lifecycle reason keys.
func summarizeDynamicLifecycleVetoReasonsWithTranslator(
	decisions []runtime.DynamicLifecycleDecision,
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

// isMockDataLoadError reports whether err represents an install that succeeded
// except for the optional mock-data load phase.
func isMockDataLoadError(err error) bool {
	var mockErr *migration.MockDataLoadError
	return errors.As(err, &mockErr)
}
