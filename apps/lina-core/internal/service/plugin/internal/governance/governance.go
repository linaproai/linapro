// Package governance owns platform-scope plugin governance checks and startup
// consistency helpers that do not require the root plugin service graph.
package governance

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TenantCapability is the tenant-governance slice required by plugin platform guards.
type TenantCapability interface {
	// Available reports whether multi-tenancy governance is active.
	Available(ctx context.Context) bool
	// PlatformBypass reports whether the request is a platform all-data context.
	PlatformBypass(ctx context.Context) bool
}

// ManifestResolver is the catalog slice required to resolve tenant governance support.
type ManifestResolver interface {
	// LoadManifestFromYAML parses a plugin.yaml file at the given path into a Manifest.
	LoadManifestFromYAML(filePath string, manifest *catalog.Manifest) error
	// GetActiveManifest returns the manifest currently in use by the host for serving.
	GetActiveManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error)
}

// EnsurePlatformContext verifies the current request can mutate platform plugin governance state.
func EnsurePlatformContext(ctx context.Context, tenantSvc TenantCapability) error {
	if tenantSvc == nil || !tenantSvc.Available(ctx) || tenantSvc.PlatformBypass(ctx) {
		return nil
	}
	return bizerr.NewCode(tenantcap.CodePlatformPermissionRequired)
}

// ValidatePluginRegistryRows verifies sys_plugin governance enum combinations
// for all synchronized plugin rows.
func ValidatePluginRegistryRows(registries []*entity.SysPlugin) []string {
	details := make([]string, 0)
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		scope := strings.TrimSpace(strings.ToLower(registry.ScopeNature))
		mode := strings.TrimSpace(strings.ToLower(registry.InstallMode))
		if !catalog.IsSupportedScopeNature(scope) {
			details = append(details, "plugin "+registry.PluginId+" has invalid scope_nature "+registry.ScopeNature)
		}
		if !catalog.IsSupportedInstallMode(mode) {
			details = append(details, "plugin "+registry.PluginId+" has invalid install_mode "+registry.InstallMode)
		}
		if catalog.NormalizeScopeNature(scope) == catalog.ScopeNaturePlatformOnly &&
			catalog.NormalizeInstallMode(mode) != catalog.InstallModeGlobal {
			details = append(details, "platform_only plugin "+registry.PluginId+" must use global install_mode")
		}
	}
	return details
}

// RegistrySupportsTenantGovernance resolves the current manifest declaration
// for one registry and falls back to the persisted scope if the manifest is
// unavailable to keep registry-only tests and startup projections deterministic.
func RegistrySupportsTenantGovernance(
	ctx context.Context,
	resolver ManifestResolver,
	registry *entity.SysPlugin,
) bool {
	if registry == nil {
		return false
	}
	if strings.TrimSpace(registry.ManifestPath) != "" {
		manifest := &catalog.Manifest{}
		if loadErr := resolver.LoadManifestFromYAML(registry.ManifestPath, manifest); loadErr == nil {
			if manifest.SupportsMultiTenant == nil {
				return catalog.NormalizeScopeNature(manifest.ScopeNature) == catalog.ScopeNatureTenantAware
			}
			return manifest.SupportsTenantGovernance()
		}
	}
	manifest, err := resolver.GetActiveManifest(ctx, registry.PluginId)
	if err == nil && manifest != nil {
		return manifest.SupportsTenantGovernance()
	}
	return catalog.NormalizeScopeNature(registry.ScopeNature) == catalog.ScopeNatureTenantAware
}
