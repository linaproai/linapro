// This file centralizes platform-context checks for plugin governance entry
// points. Startup reconciliation uses dedicated internal helpers, while HTTP
// management actions must fail closed for tenant and impersonation contexts.

package plugin

import (
	"context"

	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// platformGovernanceTenantCapability is the tenant-capability slice required by
// plugin governance guards.
type platformGovernanceTenantCapability interface {
	// Enabled reports whether multi-tenancy governance is active.
	Enabled(ctx context.Context) bool
	// PlatformBypass reports whether the request is a platform all-data context.
	PlatformBypass(ctx context.Context) bool
}

// ensurePlatformGovernance verifies the current request can mutate platform
// plugin governance state.
func (s *serviceImpl) ensurePlatformGovernance(ctx context.Context) error {
	return ensurePlatformGovernanceContext(ctx, s)
}

// ensurePlatformGovernanceContext applies platform-governance checks without
// coupling tests to the full tenantcap service interface.
func ensurePlatformGovernanceContext(ctx context.Context, holder interface {
	platformGovernanceTenantCapability() platformGovernanceTenantCapability
}) error {
	if holder == nil {
		return nil
	}
	tenantSvc := holder.platformGovernanceTenantCapability()
	if tenantSvc == nil || !tenantSvc.Enabled(ctx) || tenantSvc.PlatformBypass(ctx) {
		return nil
	}
	return bizerr.NewCode(pkgtenantcap.CodePlatformPermissionRequired)
}

// platformGovernanceTenantCapability returns the tenant capability used by the
// plugin governance guard.
func (s *serviceImpl) platformGovernanceTenantCapability() platformGovernanceTenantCapability {
	if s == nil {
		return nil
	}
	return s.tenantSvc
}
