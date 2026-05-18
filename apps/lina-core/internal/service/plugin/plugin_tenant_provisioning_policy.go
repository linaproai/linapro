// This file manages platform-owned plugin provisioning policy for newly created tenants.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
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
	if !s.registrySupportsTenantGovernance(ctx, registry) ||
		catalog.NormalizeInstallMode(registry.InstallMode) != catalog.InstallModeTenantScoped {
		return bizerr.NewCode(CodePluginTenantProvisioningPolicyInvalid, bizerr.P("pluginId", normalizedPluginID))
	}
	return s.catalogSvc.SetAutoEnableForNewTenants(ctx, normalizedPluginID, autoEnableForNewTenants)
}

// registrySupportsTenantGovernance resolves the current manifest declaration
// for one registry and falls back to the persisted scope if the manifest is
// unavailable to keep registry-only tests and startup projections deterministic.
func (s *serviceImpl) registrySupportsTenantGovernance(ctx context.Context, registry *entity.SysPlugin) bool {
	if registry == nil {
		return false
	}
	if strings.TrimSpace(registry.ManifestPath) != "" {
		manifest := &catalog.Manifest{}
		if loadErr := s.catalogSvc.LoadManifestFromYAML(registry.ManifestPath, manifest); loadErr == nil {
			if manifest.SupportsMultiTenant == nil {
				return catalog.NormalizeScopeNature(manifest.ScopeNature) == catalog.ScopeNatureTenantAware
			}
			return manifest.SupportsTenantGovernance()
		}
	}
	manifest, err := s.catalogSvc.GetActiveManifest(ctx, registry.PluginId)
	if err == nil && manifest != nil {
		return manifest.SupportsTenantGovernance()
	}
	return catalog.NormalizeScopeNature(registry.ScopeNature) == catalog.ScopeNatureTenantAware
}
