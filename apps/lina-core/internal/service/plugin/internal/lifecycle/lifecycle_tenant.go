// This file owns tenant-scoped plugin lifecycle preconditions, notifications,
// and startup tenant auto-provisioning policy reconciliation.

package lifecycle

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/governance"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// ReconcileAutoEnabledTenantPlugins applies plugin.autoEnable entries to
// tenant-scoped plugin governance after tenant-capability providers register.
func (s *serviceImpl) ReconcileAutoEnabledTenantPlugins(ctx context.Context, entries []AutoEnableEntry) error {
	if s == nil {
		return nil
	}
	if len(entries) == 0 {
		return nil
	}

	requiresProvisioning := false
	for _, entry := range entries {
		eligible, err := s.reconcileAutoEnabledTenantPluginPolicy(ctx, entry)
		if err != nil {
			return err
		}
		requiresProvisioning = requiresProvisioning || eligible
	}
	if !requiresProvisioning {
		return nil
	}
	if s.tenantProvisioning != nil {
		if err := s.tenantProvisioning.ProvisionAutoEnabledTenantPlugins(ctx); err != nil {
			return bizerr.WrapCode(err, CodePluginAutoEnableTenantProvisioningFailed, bizerr.P("pluginId", "all"))
		}
	}
	if err := s.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		return bizerr.WrapCode(err, CodePluginEnabledSnapshotRefreshFailed)
	}
	return nil
}

// reconcileAutoEnabledTenantPluginPolicy reports whether one auto-enabled
// plugin is eligible for tenant provisioning and enables its new-tenant policy
// when needed.
func (s *serviceImpl) reconcileAutoEnabledTenantPluginPolicy(
	ctx context.Context,
	entry AutoEnableEntry,
) (bool, error) {
	pluginID := strings.TrimSpace(entry.ID)
	if pluginID == "" {
		return false, nil
	}
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return false, bizerr.WrapCode(err, CodePluginRegistryReadFailed, bizerr.P("pluginId", pluginID))
	}
	if !s.isAutoEnabledTenantPluginCandidate(ctx, registry) {
		return false, nil
	}
	if !registry.AutoEnableForNewTenants {
		if err = s.storeSvc.SetAutoEnableForNewTenants(ctx, pluginID, true); err != nil {
			return false, bizerr.WrapCode(err, CodePluginAutoEnableTenantProvisioningFailed, bizerr.P("pluginId", pluginID))
		}
	}
	return true, nil
}

// isAutoEnabledTenantPluginCandidate checks whether a registry row should be
// provisioned for existing and future tenants.
func (s *serviceImpl) isAutoEnabledTenantPluginCandidate(ctx context.Context, registry *store.PluginRecord) bool {
	if registry == nil ||
		registry.Installed != plugintypes.InstalledYes ||
		registry.Status != plugintypes.StatusEnabled ||
		plugintypes.NormalizeInstallMode(registry.InstallMode) != plugintypes.InstallModeTenantScoped {
		return false
	}
	return governance.RegistrySupportsTenantGovernance(s.catalogSvc, registry)
}

// EnsureTenantPluginDisableAllowed runs source and dynamic lifecycle
// preconditions before one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || tenantID <= 0 {
		return nil
	}
	if err := s.ensureSourceTenantPluginLifecyclePreconditionAllowed(
		ctx,
		normalizedPluginID,
		tenantID,
		pluginhost.LifecycleHookBeforeTenantDisable,
	); err != nil {
		return err
	}
	return s.ensureDynamicTenantPluginLifecyclePreconditionAllowed(
		ctx,
		normalizedPluginID,
		tenantID,
		pluginhost.LifecycleHookBeforeTenantDisable,
	)
}

// NotifyTenantPluginDisabled runs best-effort source and dynamic lifecycle
// callbacks after one tenant loses access to a tenant-scoped plugin.
func (s *serviceImpl) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || tenantID <= 0 {
		return
	}
	s.executeSourceTenantPluginLifecycleNotification(
		ctx,
		normalizedPluginID,
		tenantID,
		pluginhost.LifecycleHookAfterTenantDisable,
	)
	s.executeDynamicTenantPluginLifecycleNotification(
		ctx,
		normalizedPluginID,
		tenantID,
		pluginhost.LifecycleHookAfterTenantDisable,
	)
}

// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before tenant
// deletion continues in the tenant capability provider.
func (s *serviceImpl) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	if err := s.ensureTenantLifecyclePreconditionAllowed(ctx, tenantID, pluginhost.LifecycleHookBeforeTenantDelete); err != nil {
		return err
	}
	return s.ensureDynamicTenantLifecyclePreconditionAllowed(ctx, tenantID, pluginhost.LifecycleHookBeforeTenantDelete)
}

// NotifyTenantDeleted runs best-effort source and dynamic lifecycle callbacks
// after a tenant has been deleted.
func (s *serviceImpl) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	if tenantID <= 0 {
		return
	}
	s.executeTenantLifecycleNotification(ctx, tenantID, pluginhost.LifecycleHookAfterTenantDelete)
	s.executeDynamicTenantLifecycleNotification(ctx, tenantID, pluginhost.LifecycleHookAfterTenantDelete)
}

// ensureSourceTenantPluginLifecyclePreconditionAllowed runs source-plugin
// lifecycle preconditions for one plugin and tenant pair.
func (s *serviceImpl) ensureSourceTenantPluginLifecyclePreconditionAllowed(
	ctx context.Context,
	pluginID string,
	tenantID int,
	hook pluginhost.LifecycleHook,
) error {
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         hook,
		TenantInput:  pluginhost.NewSourcePluginTenantLifecycleInput(hook.String(), tenantID),
		Participants: pluginhost.ListSourcePluginLifecycleParticipantsForPlugin(pluginID),
	})
	if result.OK {
		return nil
	}
	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", pluginID),
		bizerr.P("reasons", s.summarizeLocalizedLifecycleVetoReasons(ctx, result.Decisions)),
	)
}

// ensureTenantLifecyclePreconditionAllowed runs tenant-scoped lifecycle
// preconditions and converts vetoes to the shared lifecycle error.
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
		bizerr.P("reasons", s.summarizeLocalizedLifecycleVetoReasons(ctx, result.Decisions)),
	)
}

// ensureDynamicTenantLifecyclePreconditionAllowed runs dynamic-plugin
// tenant-scoped lifecycle preconditions before tenant deletion continues.
func (s *serviceImpl) ensureDynamicTenantLifecyclePreconditionAllowed(
	ctx context.Context,
	tenantID int,
	hook pluginhost.LifecycleHook,
) error {
	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return err
	}
	decisions := make([]runtime.DynamicLifecycleDecision, 0)
	for _, registry := range registries {
		if registry == nil ||
			plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
			registry.Installed != plugintypes.InstalledYes ||
			registry.Status != plugintypes.StatusEnabled {
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
				UninstallOptions{},
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
			return s.dynamicLifecycleError(ctx, hook, activeManifest.ID, decisions, UninstallOptions{})
		}
	}
	if tenantDynamicLifecycleDecisionsAllowed(decisions) {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, "tenant", decisions, UninstallOptions{})
}

// ensureDynamicTenantPluginLifecyclePreconditionAllowed runs dynamic-plugin
// tenant-scoped lifecycle preconditions for one plugin and tenant pair.
func (s *serviceImpl) ensureDynamicTenantPluginLifecyclePreconditionAllowed(
	ctx context.Context,
	pluginID string,
	tenantID int,
	hook pluginhost.LifecycleHook,
) error {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil ||
		plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
		registry.Installed != plugintypes.InstalledYes ||
		registry.Status != plugintypes.StatusEnabled {
		return nil
	}
	activeManifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		return s.dynamicLifecycleError(
			ctx,
			hook,
			registry.PluginId,
			[]runtime.DynamicLifecycleDecision{
				dynamicLifecycleFailureDecision(registry.PluginId, hook, err),
			},
			UninstallOptions{},
		)
	}
	if activeManifest == nil {
		return nil
	}
	decision, err := s.runtimeSvc.RunDynamicLifecyclePrecondition(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:  activeManifest.ID,
		Operation: hook,
		TenantID:  tenantID,
	})
	if decision == nil {
		return nil
	}
	decisions := []runtime.DynamicLifecycleDecision{*decision}
	if err != nil {
		return s.dynamicLifecycleError(ctx, hook, activeManifest.ID, decisions, UninstallOptions{})
	}
	if decision.OK {
		return nil
	}
	return s.dynamicLifecycleError(ctx, hook, activeManifest.ID, decisions, UninstallOptions{})
}

// tenantDynamicLifecycleDecisionsAllowed reports whether all dynamic tenant
// lifecycle participants allowed the operation.
func tenantDynamicLifecycleDecisionsAllowed(decisions []runtime.DynamicLifecycleDecision) bool {
	for _, decision := range decisions {
		if !decision.OK {
			return false
		}
	}
	return true
}

// executeTenantLifecycleNotification runs source-plugin tenant lifecycle
// notifications after tenant-wide lifecycle side effects have succeeded.
func (s *serviceImpl) executeTenantLifecycleNotification(
	ctx context.Context,
	tenantID int,
	hook pluginhost.LifecycleHook,
) {
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         hook,
		TenantInput:  pluginhost.NewSourcePluginTenantLifecycleInput(hook.String(), tenantID),
		Participants: pluginhost.ListSourcePluginLifecycleParticipants(),
	})
	if result.OK {
		return
	}
	logger.Warningf(
		ctx,
		"source plugin tenant after lifecycle callback failed operation=%s tenantID=%d reasons=%s",
		hook,
		tenantID,
		summarizeLifecycleVetoReasons(result.Decisions),
	)
}

// executeSourceTenantPluginLifecycleNotification runs one source-plugin tenant
// lifecycle notification after tenant-plugin state changed.
func (s *serviceImpl) executeSourceTenantPluginLifecycleNotification(
	ctx context.Context,
	pluginID string,
	tenantID int,
	hook pluginhost.LifecycleHook,
) {
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         hook,
		TenantInput:  pluginhost.NewSourcePluginTenantLifecycleInput(hook.String(), tenantID),
		Participants: pluginhost.ListSourcePluginLifecycleParticipantsForPlugin(pluginID),
	})
	if result.OK {
		return
	}
	logger.Warningf(
		ctx,
		"source plugin tenant after lifecycle callback failed operation=%s plugin=%s tenantID=%d reasons=%s",
		hook,
		pluginID,
		tenantID,
		summarizeLifecycleVetoReasons(result.Decisions),
	)
}

// executeDynamicTenantLifecycleNotification runs best-effort dynamic-plugin
// tenant lifecycle callbacks after tenant-wide side effects have succeeded.
func (s *serviceImpl) executeDynamicTenantLifecycleNotification(
	ctx context.Context,
	tenantID int,
	hook pluginhost.LifecycleHook,
) {
	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		logger.Warningf(ctx, "list dynamic tenant lifecycle registries failed operation=%s tenantID=%d err=%v", hook, tenantID, err)
		return
	}
	for _, registry := range registries {
		if registry == nil ||
			plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
			registry.Installed != plugintypes.InstalledYes ||
			registry.Status != plugintypes.StatusEnabled {
			continue
		}
		activeManifest, activeErr := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
		if activeErr != nil {
			logger.Warningf(
				ctx,
				"load dynamic tenant lifecycle manifest failed operation=%s plugin=%s tenantID=%d err=%v",
				hook,
				registry.PluginId,
				tenantID,
				activeErr,
			)
			continue
		}
		s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
			PluginID:  registry.PluginId,
			Operation: hook,
			TenantID:  tenantID,
		})
	}
}

// executeDynamicTenantPluginLifecycleNotification runs one dynamic-plugin
// tenant lifecycle notification after tenant-plugin state changed.
func (s *serviceImpl) executeDynamicTenantPluginLifecycleNotification(
	ctx context.Context,
	pluginID string,
	tenantID int,
	hook pluginhost.LifecycleHook,
) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		logger.Warningf(ctx, "load dynamic tenant plugin lifecycle registry failed operation=%s plugin=%s tenantID=%d err=%v", hook, pluginID, tenantID, err)
		return
	}
	if registry == nil ||
		plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
		registry.Installed != plugintypes.InstalledYes ||
		registry.Status != plugintypes.StatusEnabled {
		return
	}
	activeManifest, err := s.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		logger.Warningf(
			ctx,
			"load dynamic tenant plugin lifecycle manifest failed operation=%s plugin=%s tenantID=%d err=%v",
			hook,
			pluginID,
			tenantID,
			err,
		)
		return
	}
	s.executeDynamicPluginLifecycleNotification(ctx, activeManifest, runtime.DynamicLifecycleInput{
		PluginID:  pluginID,
		Operation: hook,
		TenantID:  tenantID,
	})
}
