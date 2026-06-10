// This file coordinates startup snapshots, startup consistency validation, and
// platform governance guards for plugin mutations.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/governance"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/pkg/bizerr"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
)

// platformGovernanceTenantCapability is the tenant-capability slice required by
// plugin governance guards.
type platformGovernanceTenantCapability = governance.TenantCapability

// pluginTenantStartupCapability is the tenant slice needed by plugin startup
// consistency checks. It excludes request resolution, data-scope, membership
// writes, and provisioning.
type pluginTenantStartupCapability interface {
	// Available reports whether an active tenant provider can serve framework calls.
	Available(ctx context.Context) bool
	// ValidateUserMembershipStartupConsistency returns startup consistency failures detected by the provider.
	ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error)
}

// SetTenantStartupCapability wires tenant provider availability and startup
// consistency checks.
func (s *serviceImpl) SetTenantStartupCapability(service pluginTenantStartupCapability) {
	if s == nil {
		return
	}
	s.tenantStartup = service
}

// SetTenantProvisioningCapability wires tenant plugin auto-provisioning.
func (s *serviceImpl) SetTenantProvisioningCapability(service tenantcapsvc.PluginProvisioningService) {
	if s == nil {
		return
	}
	s.tenantProvisioning = service
}

// SetOrganizationCapability wires the runtime-owned organization capability
// used by plugin-owned resource data-scope filtering.
func (s *serviceImpl) SetOrganizationCapability(service orgcapsvc.Service) {
	if s == nil || s.integrationSvc == nil {
		return
	}
	s.integrationSvc.SetOrganizationCapability(service)
}

// SetTenantPlatformGovernanceCapability wires platform plugin-governance checks.
func (s *serviceImpl) SetTenantPlatformGovernanceCapability(service platformGovernanceTenantCapability) {
	if s == nil {
		return
	}
	s.tenantGovernance = service
}

// WithStartupDataSnapshot returns a child context carrying catalog and
// integration startup snapshots for one host startup orchestration.
func (s *serviceImpl) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	startupCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
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
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	return governance.ValidatePluginRegistryRows(registries), nil
}

// validateProviderStartupConsistency verifies the tenant capability provider
// is active when the linapro-tenant-core plugin is enabled.
func (s *serviceImpl) validateProviderStartupConsistency(ctx context.Context) ([]string, error) {
	enabled := s.IsEnabled(ctx, tenantcap.ProviderPluginID)
	if !enabled || (s.tenantStartup != nil && s.tenantStartup.Available(ctx)) {
		return nil, nil
	}
	return []string{"linapro-tenant-core plugin is enabled but capability tenant provider is not active"}, nil
}

// validateTenantMembershipStartupConsistency delegates tenant membership
// checks to the startup-owned tenant capability instance.
func (s *serviceImpl) validateTenantMembershipStartupConsistency(ctx context.Context) ([]string, error) {
	if s == nil || s.tenantStartup == nil {
		return nil, bizerr.NewCode(
			CodePluginStartupConsistencyFailed,
			bizerr.P("details", "plugin startup consistency requires injected tenant capability service"),
		)
	}
	return s.tenantStartup.ValidateUserMembershipStartupConsistency(ctx)
}

// ensurePlatformGovernance verifies the current request can mutate platform
// plugin governance state.
func (s *serviceImpl) ensurePlatformGovernance(ctx context.Context) error {
	return governance.EnsurePlatformContext(ctx, s.platformGovernanceTenantCapability())
}

// platformGovernanceTenantCapability returns the tenant capability used by the
// plugin governance guard.
func (s *serviceImpl) platformGovernanceTenantCapability() platformGovernanceTenantCapability {
	if s == nil {
		return nil
	}
	return s.tenantGovernance
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
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", pluginID))
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, normalizedPluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return bizerr.NewCode(CodePluginSourceRegistryNotFound, bizerr.P("pluginId", normalizedPluginID))
	}
	if !s.registrySupportsTenantGovernance(ctx, registry) ||
		catalog.NormalizeInstallMode(registry.InstallMode) != catalog.InstallModeTenantScoped {
		return bizerr.NewCode(CodePluginTenantProvisioningPolicyInvalid, bizerr.P("pluginId", normalizedPluginID))
	}
	if err = s.catalogSvc.SetAutoEnableForNewTenants(ctx, normalizedPluginID, autoEnableForNewTenants); err != nil {
		return err
	}
	_, err = s.markRuntimeCacheChanged(ctx, "plugin_tenant_provisioning_policy_updated")
	return err
}

// registrySupportsTenantGovernance resolves the current manifest declaration
// for one registry and falls back to the persisted scope if the manifest is
// unavailable to keep registry-only tests and startup projections deterministic.
func (s *serviceImpl) registrySupportsTenantGovernance(ctx context.Context, registry *entity.SysPlugin) bool {
	return governance.RegistrySupportsTenantGovernance(ctx, s.catalogSvc, registry)
}
