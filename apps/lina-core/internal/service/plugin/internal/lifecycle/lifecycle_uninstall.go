// This file owns plugin uninstall orchestration for source and dynamic plugins,
// including reverse dependency guards, lifecycle preconditions, force handling,
// cache publication, and observer notifications.

package lifecycle

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// UninstallOptions defines one plugin uninstall policy snapshot.
type UninstallOptions struct {
	// PurgeStorageData reports whether uninstall should also clear plugin-owned
	// table data and stored files.
	PurgeStorageData bool
	// Force reports whether an authorized caller requested precondition veto bypass.
	Force bool
	// AllowForceUninstall is the root-facade host configuration snapshot that
	// permits destructive force uninstall behavior for this request.
	AllowForceUninstall bool
}

// Uninstall executes the uninstall lifecycle for an installed source or dynamic plugin.
func (s *serviceImpl) Uninstall(ctx context.Context, pluginID string, options UninstallOptions) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return s.uninstallWithoutDesiredManifest(ctx, pluginID, options, err)
	}
	if err = s.ensureNoReverseDependencies(ctx, pluginID); err != nil {
		return err
	}
	if plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeSource {
		return s.uninstallSource(ctx, manifest, options)
	}
	return s.uninstallDynamic(ctx, pluginID, options)
}

// uninstallWithoutDesiredManifest keeps dynamic-plugin uninstall recoverable
// when the mutable staging artifact is missing but registry state still carries
// enough active-release data to complete or force one uninstall.
func (s *serviceImpl) uninstallWithoutDesiredManifest(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
	discoveryErr error,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || plugintypes.NormalizeType(registry.Type) != pluginv1.PluginTypeDynamic {
		return discoveryErr
	}
	if err = s.ensureNoReverseDependencies(ctx, pluginID); err != nil {
		return err
	}
	return s.uninstallDynamicWithRegistry(ctx, pluginID, registry, options)
}

// uninstallSource executes source-plugin uninstall and shared lifecycle epilogue.
func (s *serviceImpl) uninstallSource(ctx context.Context, manifest *catalog.Manifest, options UninstallOptions) error {
	if err := s.uninstallSourcePlugin(ctx, manifest, options); err != nil {
		return wrapUninstallExecutionError(err, manifest.ID)
	}
	return s.completeSourceUninstall(ctx, manifest, options)
}

// completeSourceUninstall refreshes runtime visibility and emits source
// uninstall notifications after source-plugin side effects commit.
func (s *serviceImpl) completeSourceUninstall(ctx context.Context, manifest *catalog.Manifest, options UninstallOptions) error {
	pluginID := ""
	if manifest != nil {
		pluginID = manifest.ID
	}
	if err := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "source_plugin_uninstalled"); err != nil {
		return err
	}
	if err := s.notifyPluginUninstalled(ctx, pluginID); err != nil {
		return err
	}
	s.executeSourcePluginAfterLifecycle(ctx, manifest, pluginhost.LifecycleHookAfterUninstall, sourceLifecyclePolicy{
		purgeStorageData: options.PurgeStorageData,
	})
	return nil
}

// uninstallDynamic executes dynamic-plugin uninstall and shared lifecycle epilogue.
func (s *serviceImpl) uninstallDynamic(ctx context.Context, pluginID string, options UninstallOptions) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if err = s.uninstallDynamicWithRegistry(ctx, pluginID, registry, options); err != nil {
		return err
	}
	return nil
}

// uninstallDynamicWithRegistry executes dynamic preconditions, runtime
// reconciliation, cache publication, observer notification, and best-effort
// AfterUninstall callback using one registry snapshot.
func (s *serviceImpl) uninstallDynamicWithRegistry(
	ctx context.Context,
	pluginID string,
	registry *store.PluginRecord,
	options UninstallOptions,
) error {
	if strings.TrimSpace(pluginID) == "" && registry != nil {
		pluginID = registry.PluginId
	}
	if err := s.ensureDynamicPluginActiveLifecyclePreconditionAllowed(
		ctx,
		registry,
		pluginhost.LifecycleHookBeforeUninstall,
		options,
	); err != nil {
		return err
	}
	activeManifest := s.loadActiveDynamicLifecycleManifestBestEffort(ctx, pluginID)
	if err := s.uninstallDynamicPlugin(ctx, pluginID, options); err != nil {
		return wrapUninstallExecutionError(err, pluginID)
	}
	if err := s.completeDynamicUninstall(ctx, pluginID); err != nil {
		return err
	}
	s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:         pluginID,
		Operation:        pluginhost.LifecycleHookAfterUninstall,
		PurgeStorageData: options.PurgeStorageData,
	})
	return nil
}

// completeDynamicUninstall refreshes runtime visibility and emits dynamic
// uninstall observer notifications after runtime side effects commit.
func (s *serviceImpl) completeDynamicUninstall(ctx context.Context, pluginID string) error {
	if err := s.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, "dynamic_plugin_uninstalled"); err != nil {
		return err
	}
	return s.notifyPluginUninstalled(ctx, pluginID)
}

// uninstallDynamicPlugin chooses between the full active-release uninstall and
// the restricted orphan cleanup path used only when active dynamic artifacts are
// no longer readable.
func (s *serviceImpl) uninstallDynamicPlugin(
	ctx context.Context,
	pluginID string,
	options UninstallOptions,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || plugintypes.NormalizeType(registry.Type) != pluginv1.PluginTypeDynamic {
		return s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData)
	}
	if registry.Installed != statusflag.Installed.Int() {
		if options.Force {
			return s.forceUninstallMissingDynamicArtifact(ctx, registry, options)
		}
		return s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData)
	}
	if s.dynamicFullUninstallRecoverable(ctx, registry) {
		if err = s.runtimeSvc.UninstallWithOptions(ctx, pluginID, options.PurgeStorageData); err == nil {
			return nil
		}
		refreshed, refreshErr := s.storeSvc.GetRegistry(ctx, pluginID)
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
	return s.forceUninstallMissingDynamicArtifact(ctx, registry, options)
}

// dynamicFullUninstallRecoverable reports whether the installed dynamic plugin
// can run full uninstall from its archived active release or repair that archive
// from the current same-version staging artifact.
func (s *serviceImpl) dynamicFullUninstallRecoverable(ctx context.Context, registry *store.PluginRecord) bool {
	if registry == nil || plugintypes.NormalizeType(registry.Type) != pluginv1.PluginTypeDynamic {
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
	if plugintypes.NormalizeType(desiredManifest.Type) != pluginv1.PluginTypeDynamic {
		return false
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		return false
	}
	return desiredManifest.RuntimeArtifact != nil
}

// forceUninstallMissingDynamicArtifact validates host policy before clearing
// host-owned orphan governance for a dynamic plugin with unreadable artifacts.
func (s *serviceImpl) forceUninstallMissingDynamicArtifact(
	ctx context.Context,
	registry *store.PluginRecord,
	options UninstallOptions,
) error {
	if err := ensureForceUninstallEnabled(options); err != nil {
		return err
	}
	return s.runtimeSvc.ForceUninstallMissingArtifact(ctx, registry)
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
		plugintypes.NormalizeType(registry.Type) != pluginv1.PluginTypeDynamic ||
		registry.Installed != statusflag.Installed.Int() {
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
			options,
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
		plugintypes.NormalizeType(manifest.Type) != pluginv1.PluginTypeDynamic ||
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

// ensureNoReverseDependencies blocks uninstall when installed downstream plugins depend on target.
func (s *serviceImpl) ensureNoReverseDependencies(ctx context.Context, pluginID string) error {
	result, err := s.resolveReverseDependencies(ctx, pluginID, "")
	if err != nil {
		return err
	}
	if !plugindep.HasBlockers(result.Blockers) {
		return nil
	}
	return buildReverseDependencyBlockedError(pluginID, result)
}

// resolveReverseDependencies evaluates installed downstream dependencies for one target.
func (s *serviceImpl) resolveReverseDependencies(
	ctx context.Context,
	pluginID string,
	candidateVersion string,
) (*plugindep.ReverseCheckResult, error) {
	snapshots, err := s.buildDependencySnapshots(ctx, nil)
	if err != nil {
		return nil, err
	}
	resolver := s.dependencyResolver
	if resolver == nil {
		resolver = plugindep.New()
	}
	return resolver.CheckReverse(plugindep.ReverseCheckInput{
		TargetID:         strings.TrimSpace(pluginID),
		CandidateVersion: strings.TrimSpace(candidateVersion),
		Plugins:          snapshots,
	}), nil
}

// buildReverseDependencyBlockedError converts reverse dependency blockers into one structured error.
func buildReverseDependencyBlockedError(
	pluginID string,
	result *plugindep.ReverseCheckResult,
) error {
	dependents := plugindep.ToReverseDependentProjections(result.Dependents)
	dependencyID, requiredVersion, currentVersion := plugindep.FirstBlockerFields(result.Blockers)
	return bizerr.NewCode(
		CodePluginReverseDependencyBlocked,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
		bizerr.P("dependencyId", dependencyID),
		bizerr.P("requiredVersion", requiredVersion),
		bizerr.P("currentVersion", currentVersion),
		bizerr.P("dependents", strings.Join(plugindep.ReverseDependentIDs(dependents), ",")),
		bizerr.P("ownerHostServices", plugindep.FormatReverseDependentOwnerHostServices(dependents)),
		bizerr.P("blockers", plugindep.FormatBlockers(result.Blockers)),
	)
}

// wrapUninstallExecutionError preserves stable business errors and wraps
// low-level uninstall side-effect failures before they reach API callers.
func wrapUninstallExecutionError(err error, pluginID string) error {
	if err == nil {
		return nil
	}
	if _, ok := bizerr.As(err); ok {
		return err
	}
	return bizerr.WrapCode(
		err,
		CodePluginUninstallExecutionFailed,
		bizerr.P("pluginId", strings.TrimSpace(pluginID)),
	)
}

// ensureForceUninstallEnabled verifies that host configuration explicitly
// permits destructive force-uninstall flows.
func ensureForceUninstallEnabled(options UninstallOptions) error {
	if !options.AllowForceUninstall {
		return bizerr.NewCode(CodePluginForceUninstallDisabled)
	}
	return nil
}

// dynamicLifecycleError converts dynamic lifecycle vetoes to the shared
// caller-visible bizerr used by source-plugin lifecycle preconditions.
func (s *serviceImpl) dynamicLifecycleError(
	ctx context.Context,
	hook pluginhost.LifecycleHook,
	pluginID string,
	decisions []runtime.DynamicLifecycleDecision,
	options UninstallOptions,
) error {
	reasons := s.summarizeLocalizedDynamicLifecycleVetoReasons(ctx, decisions)
	if options.Force && hook == pluginhost.LifecycleHookBeforeUninstall {
		if err := ensureForceUninstallEnabled(options); err != nil {
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
