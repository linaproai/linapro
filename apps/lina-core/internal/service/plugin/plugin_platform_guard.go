// This file centralizes platform-governance guards used before plugin
// lifecycle, package, sync, and provisioning mutations.

package plugin

import (
	"context"

	"lina-core/internal/service/plugin/internal/governance"
)

// platformGovernanceTenantCapability is the tenant-capability slice required by
// plugin governance guards.
type platformGovernanceTenantCapability = governance.TenantCapability

// SetTenantPlatformGovernanceCapability wires platform plugin-governance checks.
func (s *serviceImpl) SetTenantPlatformGovernanceCapability(service platformGovernanceTenantCapability) {
	if s == nil {
		return
	}
	s.tenantGovernance = service
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
