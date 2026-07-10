// This file owns explicit source and dynamic runtime-upgrade execution
// orchestration, including lock coordination, state re-read, failure recovery,
// dynamic authorization persistence, and cache publication.

package upgrade

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"strings"
	"sync"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
func (s *serviceImpl) ExecuteRuntimeUpgrade(
	ctx context.Context,
	pluginID string,
	options RuntimeUpgradeOptions,
) (*RuntimeUpgradeResult, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	if !options.Confirmed {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeConfirmationRequired,
			bizerr.P("pluginId", normalizedPluginID),
		)
	}
	unlock, err := s.lockRuntimeUpgrade(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	defer unlock()

	if err := s.cacheFreshener.EnsureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}

	targetManifest, registry, projection, err := s.loadRuntimeUpgradeExecutionState(ctx, normalizedPluginID)
	if err != nil {
		return nil, err
	}
	if !canExecute(projection.State) {
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeUnavailable,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("runtimeState", projection.State.String()),
		)
	}
	if plugintypes.NormalizeType(targetManifest.Type) == pluginv1.PluginTypeDynamic {
		if err = s.ensureDynamicPluginUpgradeLifecyclePreconditionAllowed(
			ctx,
			registry,
			targetManifest,
			options.Authorization,
		); err != nil {
			return nil, err
		}
	}
	if err = s.markRuntimeUpgradeRunning(ctx, registry); err != nil {
		return nil, err
	}

	result := &RuntimeUpgradeResult{
		PluginID:          normalizedPluginID,
		FromVersion:       projection.EffectiveVersion,
		ToVersion:         projection.DiscoveredVersion,
		EffectiveVersion:  projection.EffectiveVersion,
		DiscoveredVersion: projection.DiscoveredVersion,
	}

	if err = s.executeRuntimeUpgradeByType(ctx, registry, targetManifest, options); err != nil {
		if restoreErr := s.restoreRuntimeUpgradeStableState(ctx, normalizedPluginID); restoreErr != nil {
			logger.Warningf(
				ctx,
				"restore runtime upgrade stable state after failure failed plugin=%s err=%v",
				normalizedPluginID,
				restoreErr,
			)
		}
		if syncErr := s.cachePublisher.SyncEnabledSnapshotAndPublishRuntimeChange(
			ctx,
			normalizedPluginID,
			"runtime_upgrade_failed",
		); syncErr != nil {
			logger.Warningf(
				ctx,
				"sync plugin snapshot after runtime upgrade failure failed plugin=%s err=%v",
				normalizedPluginID,
				syncErr,
			)
		}
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeExecutionFailed,
			bizerr.P("pluginId", normalizedPluginID),
			bizerr.P("fromVersion", result.FromVersion),
			bizerr.P("toVersion", result.ToVersion),
		)
	}
	if plugintypes.NormalizeType(targetManifest.Type) == pluginv1.PluginTypeDynamic {
		s.executeDynamicPluginUpgradeLifecycleNotification(
			ctx,
			registry,
			targetManifest,
			options.Authorization,
		)
	}

	if err = s.cachePublisher.SyncEnabledSnapshotAndPublishRuntimeChange(
		ctx,
		normalizedPluginID,
		"runtime_upgrade_succeeded",
	); err != nil {
		return nil, err
	}
	refreshedManifest, refreshedRegistry, refreshedProjection, err := s.loadRuntimeUpgradeExecutionState(
		ctx,
		normalizedPluginID,
	)
	if err != nil {
		return nil, err
	}
	result.Executed = true
	result.RuntimeState = RuntimeUpgradeState(refreshedProjection.State)
	result.EffectiveVersion = refreshedProjection.EffectiveVersion
	result.DiscoveredVersion = refreshedProjection.DiscoveredVersion
	if refreshedManifest == nil && refreshedRegistry != nil {
		result.EffectiveVersion = refreshedRegistry.Version
		result.DiscoveredVersion = refreshedRegistry.Version
	}
	return result, nil
}

// loadRuntimeUpgradeExecutionState re-reads the target manifest, registry row,
// and version-drift projection from authoritative storage.
func (s *serviceImpl) loadRuntimeUpgradeExecutionState(
	ctx context.Context,
	pluginID string,
) (*catalog.Manifest, *store.PluginRecord, plugintypes.RuntimeUpgradeProjection, error) {
	targetManifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return nil, nil, plugintypes.RuntimeUpgradeProjection{}, err
	}
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, nil, plugintypes.RuntimeUpgradeProjection{}, err
	}
	if registry == nil {
		return nil, nil, plugintypes.RuntimeUpgradeProjection{}, bizerr.NewCode(
			CodePluginNotFound,
			bizerr.P("pluginId", pluginID),
		)
	}
	projection, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, targetManifest)
	if err != nil {
		return nil, nil, plugintypes.RuntimeUpgradeProjection{}, err
	}
	return targetManifest, registry, projection, nil
}

// markRuntimeUpgradeRunning records an observable in-progress state before the
// upgrade starts. Projection later reports pending/failed/normal from the
// authoritative release and version state after execution completes.
func (s *serviceImpl) markRuntimeUpgradeRunning(ctx context.Context, registry *store.PluginRecord) error {
	if registry == nil {
		return bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", ""))
	}
	return s.storeSvc.SetRegistryRuntimeState(ctx, registry.PluginId, store.RuntimeStatePatch{
		CurrentState: plugintypes.RuntimeUpgradeStateUpgradeRunning.String(),
	})
}

// restoreRuntimeUpgradeStableState clears transient runtime-upgrade markers
// after a failed explicit upgrade so projection can expose upgrade_failed or
// pending_upgrade from release/version state instead of staying in running.
func (s *serviceImpl) restoreRuntimeUpgradeStableState(ctx context.Context, pluginID string) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}
	stableState := store.BuildStableHostState(registry)
	return s.storeSvc.SetRegistryRuntimeState(ctx, registry.PluginId, store.RuntimeStatePatch{
		DesiredState: stableState,
		CurrentState: stableState,
	})
}

// executeRuntimeUpgradeByType dispatches the confirmed upgrade to the source or
// dynamic strategy after common validation has completed.
func (s *serviceImpl) executeRuntimeUpgradeByType(
	ctx context.Context,
	registry *store.PluginRecord,
	targetManifest *catalog.Manifest,
	options RuntimeUpgradeOptions,
) error {
	if targetManifest == nil {
		return bizerr.NewCode(CodePluginNotFound, bizerr.P("pluginId", ""))
	}

	switch plugintypes.NormalizeType(targetManifest.Type) {
	case pluginv1.PluginTypeSource:
		_, err := s.ExecuteSourcePluginUpgrade(ctx, targetManifest.ID)
		return err
	case pluginv1.PluginTypeDynamic:
		if err := s.persistDynamicPluginAuthorization(ctx, targetManifest, options.Authorization); err != nil {
			return err
		}
		if registry != nil && registry.Installed == statusflag.Installed.Int() {
			if err := s.runtimeSvc.UpgradeDynamicPluginRequest(ctx, targetManifest.ID); err != nil {
				s.markDynamicPluginUpgradeFailed(ctx, targetManifest, err)
				return err
			}
			return nil
		}
		return bizerr.NewCode(CodePluginNotInstalled)
	default:
		return bizerr.NewCode(
			CodePluginRuntimeUpgradeTypeUnsupported,
			bizerr.P("pluginId", targetManifest.ID),
			bizerr.P("pluginType", targetManifest.Type),
		)
	}
}

// markDynamicPluginUpgradeFailed writes the runtime-reported dynamic failure
// phase into the same upgrade migration ledger source plugins use. It only
// records diagnostics after runtime has marked the target release failed, so a
// cache or hook error after a successful release switch cannot demote the
// effective release into an upgrade_failed projection.
func (s *serviceImpl) markDynamicPluginUpgradeFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	upgradeErr error,
) {
	if manifest == nil || upgradeErr == nil {
		return
	}
	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		logger.Warningf(
			ctx,
			"load dynamic plugin failed release for upgrade diagnostics failed plugin=%s err=%v",
			manifest.ID,
			err,
		)
		return
	}
	if release == nil || strings.TrimSpace(release.Status) != plugintypes.ReleaseStatusFailed.String() {
		return
	}
	phase := runtime.DynamicUpgradeFailurePhase(upgradeErr)
	if phase == "" {
		phase = plugintypes.RuntimeUpgradeFailurePhaseRelease
	}
	if err = recordRuntimeUpgradeFailureMigration(ctx, manifest.ID, release.Id, phase.String(), upgradeErr); err != nil {
		logger.Warningf(
			ctx,
			"record dynamic plugin upgrade failure failed plugin=%s phase=%s err=%v",
			manifest.ID,
			phase.String(),
			err,
		)
	}
}

// lockRuntimeUpgrade serializes explicit runtime upgrades for one plugin within
// the current process and, when cluster mode is enabled, across all nodes via
// the configured coordination lock store.
func (s *serviceImpl) lockRuntimeUpgrade(ctx context.Context, pluginID string) (func(), error) {
	localUnlock := s.lockRuntimeUpgradeLocal(pluginID)
	if s == nil || s.topology == nil || !s.topology.IsEnabled() {
		return localUnlock, nil
	}
	if s.runtimeUpgradeLockStore == nil {
		localUnlock()
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeLockUnavailable,
			bizerr.P("pluginId", pluginID),
		)
	}

	lockName := distributedLockName(pluginID)
	owner := distributedLockOwner(s.topology)
	handle, ok, err := s.runtimeUpgradeLockStore.Acquire(
		ctx,
		lockName,
		owner,
		distributedLockReason,
		distributedLockLease,
	)
	if err != nil {
		localUnlock()
		return nil, bizerr.WrapCode(
			err,
			CodePluginRuntimeUpgradeLockUnavailable,
			bizerr.P("pluginId", pluginID),
		)
	}
	if !ok || handle == nil {
		localUnlock()
		return nil, bizerr.NewCode(
			CodePluginRuntimeUpgradeAlreadyRunning,
			bizerr.P("pluginId", pluginID),
		)
	}

	return func() {
		if releaseErr := s.runtimeUpgradeLockStore.Release(ctx, handle); releaseErr != nil {
			logger.Warningf(
				ctx,
				"release runtime upgrade distributed lock failed plugin=%s lock=%s err=%v",
				pluginID,
				lockName,
				releaseErr,
			)
		}
		localUnlock()
	}, nil
}

// lockRuntimeUpgradeLocal serializes explicit runtime upgrades for one plugin
// within the current process.
func (s *serviceImpl) lockRuntimeUpgradeLocal(pluginID string) func() {
	if s == nil {
		return func() {}
	}
	s.runtimeUpgradeLocksMu.Lock()
	lock := s.runtimeUpgradeLocks[pluginID]
	if lock == nil {
		lock = &sync.Mutex{}
		s.runtimeUpgradeLocks[pluginID] = lock
	}
	s.runtimeUpgradeLocksMu.Unlock()
	lock.Lock()
	return lock.Unlock
}

// persistDynamicPluginAuthorization refreshes the release snapshot for dynamic
// plugins so the runtime switch consumes the operator-confirmed hostServices.
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

// ensureDynamicPluginUpgradeLifecyclePreconditionAllowed runs BeforeUpgrade for
// dynamic plugins before upgrade state markers or release switch side effects.
func (s *serviceImpl) ensureDynamicPluginUpgradeLifecyclePreconditionAllowed(
	ctx context.Context,
	registry *store.PluginRecord,
	targetManifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
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
		return s.dynamicLifecycleError(ctx, pluginhost.LifecycleHookBeforeUpgrade, targetManifest.ID, decisions)
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, pluginhost.LifecycleHookBeforeUpgrade, targetManifest.ID, decisions)
}

// executeDynamicPluginUpgradeLifecycleNotification runs AfterUpgrade for a
// dynamic plugin after the target release has become effective.
func (s *serviceImpl) executeDynamicPluginUpgradeLifecycleNotification(
	ctx context.Context,
	registry *store.PluginRecord,
	targetManifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
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
	authorization *store.HostServiceAuthorizationInput,
) (*catalog.Manifest, error) {
	if manifest == nil {
		return nil, nil
	}
	if authorization != nil {
		return cloneManifestWithAuthorizedHostServices(manifest, authorization)
	}
	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return cloneManifestWithAuthorizedHostServices(manifest, nil)
	}
	snapshot, err := s.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || !snapshot.HostServiceAuthRequired || !snapshot.HostServiceAuthConfirmed {
		return cloneManifestWithAuthorizedHostServices(manifest, nil)
	}
	hostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(manifest.ID, snapshot.AuthorizedHostServices)
	if err != nil {
		return nil, err
	}
	clone := *manifest
	clone.HostServices = hostServices
	clone.HostCapabilities = protocol.CapabilityMapFromHostServices(hostServices)
	return &clone, nil
}

// cloneManifestWithAuthorizedHostServices returns a manifest copy carrying the
// authorized hostServices that should be visible to dynamic lifecycle handlers.
func cloneManifestWithAuthorizedHostServices(
	manifest *catalog.Manifest,
	authorization *store.HostServiceAuthorizationInput,
) (*catalog.Manifest, error) {
	if manifest == nil {
		return nil, nil
	}
	hostServices, err := buildAuthorizedHostServicesForPlugin(
		manifest.ID,
		manifest.HostServices,
		authorization,
	)
	if err != nil {
		return nil, err
	}
	clone := *manifest
	clone.HostServices = hostServices
	clone.HostCapabilities = protocol.CapabilityMapFromHostServices(hostServices)
	return &clone, nil
}

// buildAuthorizedHostServicesForPlugin derives lifecycle-visible hostServices
// from the current request authorization or from safe unscoped declarations.
func buildAuthorizedHostServicesForPlugin(
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

// dynamicLifecycleError converts dynamic lifecycle vetoes to the shared
// caller-visible bizerr used by plugin lifecycle preconditions.
func (s *serviceImpl) dynamicLifecycleError(
	ctx context.Context,
	hook pluginhost.LifecycleHook,
	pluginID string,
	decisions []runtime.DynamicLifecycleDecision,
) error {
	reasons := s.summarizeLocalizedDynamicLifecycleVetoReasons(ctx, decisions)
	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", pluginID),
		bizerr.P("reasons", reasons),
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
