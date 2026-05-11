// This file validates persisted plugin and tenant-governance state before the
// host starts serving requests.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// ValidateStartupConsistency verifies persisted startup state that must be
// coherent before HTTP routes and plugin callbacks become reachable.
func (s *serviceImpl) ValidateStartupConsistency(ctx context.Context) error {
	if s == nil || s.catalogSvc == nil {
		return nil
	}
	var details []string
	pluginDetails, err := s.validatePluginStartupConsistency(ctx)
	if err != nil {
		return err
	}
	details = append(details, pluginDetails...)
	providerDetails, err := s.validateTenantProviderStartupConsistency(ctx)
	if err != nil {
		return err
	}
	details = append(details, providerDetails...)
	membershipDetails, err := tenantcapsvc.New(s).ValidateUserMembershipStartupConsistency(ctx)
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
	return details, nil
}

// validateTenantProviderStartupConsistency verifies the tenant capability
// provider is registered when the multi-tenant plugin is enabled.
func (s *serviceImpl) validateTenantProviderStartupConsistency(ctx context.Context) ([]string, error) {
	enabled := s.IsEnabled(ctx, pkgtenantcap.ProviderPluginID)
	if !enabled || pkgtenantcap.HasProvider() {
		return nil, nil
	}
	return []string{"multi-tenant plugin is enabled but tenantcap provider is not registered"}, nil
}
