// This file manages platform-owned plugin provisioning policy for newly created tenants.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
)

// UpdateTenantProvisioningPolicy updates the platform-owned new-tenant plugin provisioning policy.
func (s *serviceImpl) UpdateTenantProvisioningPolicy(
	ctx context.Context,
	pluginID string,
	autoEnableForNewTenants bool,
) error {
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
	if catalog.NormalizeScopeNature(registry.ScopeNature) != catalog.ScopeNatureTenantAware ||
		catalog.NormalizeInstallMode(registry.InstallMode) != catalog.InstallModeTenantScoped {
		return bizerr.NewCode(CodePluginTenantProvisioningPolicyInvalid, bizerr.P("pluginId", normalizedPluginID))
	}
	return s.catalogSvc.SetAutoEnableForNewTenants(ctx, normalizedPluginID, autoEnableForNewTenants)
}
