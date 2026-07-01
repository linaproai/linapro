// This file coordinates startup snapshots, startup consistency validation, and
// platform governance guards for plugin mutations.

package plugin

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"strings"

	"lina-core/internal/service/plugin/internal/governance"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// WithStartupDataSnapshot returns a child context carrying catalog and
// integration startup snapshots for one host startup orchestration.
func (s *serviceImpl) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	startupCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
	startupCtx, err = s.integrationSvc.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		return ctx, err
	}
	return startupCtx, nil
}

// ValidateStartupConsistency verifies persisted startup state that must be
// coherent before HTTP routes and plugin callbacks become reachable.
func (s *serviceImpl) ValidateStartupConsistency(ctx context.Context) error {
	if s == nil || s.catalogSvc == nil {
		return nil
	}
	ctx = integration.WithAuthoritativeEnablement(ctx)
	var details []string
	pluginDetails, err := s.validatePluginStartupConsistency(ctx)
	if err != nil {
		return err
	}
	details = append(details, pluginDetails...)
	providerDetails, err := s.validateProviderStartupConsistency(ctx)
	if err != nil {
		return err
	}
	details = append(details, providerDetails...)
	membershipDetails, err := s.validateTenantMembershipStartupConsistency(ctx)
	if err != nil {
		return err
	}
	details = append(details, membershipDetails...)
	if len(details) == 0 {
		return nil
	}
	return bizerr.NewCode(
		CodePluginStartupConsistencyFailed,
		bizerr.P("details", strings.Join(details, "; ")),
	)
}

// validatePluginStartupConsistency verifies sys_plugin governance enum
// combinations for all synchronized plugin rows.
func (s *serviceImpl) validatePluginStartupConsistency(ctx context.Context) ([]string, error) {
	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	return governance.ValidatePluginRegistryRows(registries), nil
}

// validateProviderStartupConsistency verifies the tenant capability provider
// is active when the linapro-tenant-core plugin is enabled.
func (s *serviceImpl) validateProviderStartupConsistency(ctx context.Context) ([]string, error) {
	enabled := s.IsEnabled(ctx, tenantcap.ProviderPluginID)
	if !enabled || (s.tenantSvc != nil && s.tenantSvc.Available(ctx)) {
		return nil, nil
	}
	return []string{"linapro-tenant-core plugin is enabled but capability tenant provider is not active"}, nil
}

// validateTenantMembershipStartupConsistency delegates tenant membership
// checks to the startup-owned tenant capability instance.
func (s *serviceImpl) validateTenantMembershipStartupConsistency(ctx context.Context) ([]string, error) {
	if s == nil || s.tenantSvc == nil {
		return nil, bizerr.NewCode(
			CodePluginStartupConsistencyFailed,
			bizerr.P("details", "plugin startup consistency requires injected tenant capability service"),
		)
	}
	return s.tenantSvc.ValidateUserMembershipStartupConsistency(ctx)
}

// ensurePlatformGovernance verifies the current request can mutate platform
// plugin governance state.
func (s *serviceImpl) ensurePlatformGovernance(ctx context.Context) error {
	var tenantSvc tenantspi.Service
	if s != nil {
		tenantSvc = s.tenantSvc
	}
	return governance.EnsurePlatformContext(ctx, tenantSvc)
}

// ensureBuiltinManagementActionAllowed rejects ordinary management mutations
// for project built-in plugins. It consults both the registry and the desired
// manifest so uninstalled builtin plugins are guarded before a registry row
// exists, while registry-only dynamic cleanup paths can continue when the
// mutable manifest is unavailable.
func (s *serviceImpl) ensureBuiltinManagementActionAllowed(ctx context.Context, pluginID string) error {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}
	if s == nil {
		return nil
	}

	if s.storeSvc != nil {
		registry, err := s.storeSvc.GetRegistry(ctx, normalizedPluginID)
		if err != nil {
			return err
		}
		if registry != nil && registry.Distribution == pluginv1.PluginDistributionBuiltin.String() {
			return bizerr.NewCode(CodePluginBuiltinManagementActionDenied, bizerr.P("pluginId", normalizedPluginID))
		}
	}

	if s.catalogSvc == nil {
		return nil
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(normalizedPluginID)
	if err != nil || manifest == nil {
		return nil
	}
	if plugintypes.NormalizeDistribution(manifest.Distribution) == pluginv1.PluginDistributionBuiltin {
		return bizerr.NewCode(CodePluginBuiltinManagementActionDenied, bizerr.P("pluginId", normalizedPluginID))
	}
	return nil
}

// UpdateTenantProvisioningPolicy updates the platform-owned new-tenant plugin provisioning policy.
func (s *serviceImpl) UpdateTenantProvisioningPolicy(
	ctx context.Context,
	pluginID string,
	autoEnableForNewTenants bool,
) error {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return err
	}
	if err := s.ensureBuiltinManagementActionAllowed(ctx, pluginID); err != nil {
		return err
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", pluginID))
	}
	registry, err := s.storeSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	if !s.registrySupportsTenantGovernance(ctx, registry) ||
		plugintypes.NormalizeInstallMode(registry.InstallMode) != pluginv1.InstallModeTenantScoped {
		return bizerr.NewCode(CodePluginTenantProvisioningPolicyInvalid, bizerr.P("pluginId", normalizedPluginID))
	}
	if err = s.storeSvc.SetAutoEnableForNewTenants(ctx, normalizedPluginID, autoEnableForNewTenants); err != nil {
		return err
	}
	_, err = s.publishPluginChange(ctx, pluginChangePublishInput{
		pluginID:   normalizedPluginID,
		pluginType: registry.Type,
		reason:     "plugin_tenant_provisioning_policy_updated",
	})
	return err
}

// registrySupportsTenantGovernance resolves the current manifest declaration
// for one registry and falls back to the persisted scope if the manifest is
// unavailable to keep registry-only tests and startup projections deterministic.
func (s *serviceImpl) registrySupportsTenantGovernance(ctx context.Context, registry *store.PluginRecord) bool {
	return governance.RegistrySupportsTenantGovernance(s.catalogSvc, registry)
}
