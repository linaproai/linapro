// This file keeps platform-context checks for global menu-governance writes in
// one place. sys_menu remains the host-wide permission topology, so tenant and
// impersonation contexts must fail before any mutation or topology revision.

package menu

import (
	"context"

	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// platformMenuTenantCapability is the tenant-capability slice required by
// global menu governance guards.
type platformMenuTenantCapability interface {
	// Enabled reports whether multi-tenancy governance is active.
	Enabled(ctx context.Context) bool
	// PlatformBypass reports whether the request is a platform all-data context.
	PlatformBypass(ctx context.Context) bool
}

// ensurePlatformMenuGovernance verifies the current request can mutate the
// global menu topology.
func (s *serviceImpl) ensurePlatformMenuGovernance(ctx context.Context) error {
	return ensurePlatformMenuGovernanceContext(ctx, s)
}

// ensurePlatformMenuGovernanceContext applies platform-menu checks without
// coupling tests to the full tenantcap service interface.
func ensurePlatformMenuGovernanceContext(ctx context.Context, holder interface {
	platformMenuTenantCapability() platformMenuTenantCapability
}) error {
	if holder == nil {
		return nil
	}
	tenantSvc := holder.platformMenuTenantCapability()
	if tenantSvc == nil || !tenantSvc.Enabled(ctx) || tenantSvc.PlatformBypass(ctx) {
		return nil
	}
	return bizerr.NewCode(pkgtenantcap.CodePlatformPermissionRequired)
}

// platformMenuTenantCapability returns the tenant capability used by the menu
// governance guard.
func (s *serviceImpl) platformMenuTenantCapability() platformMenuTenantCapability {
	if s == nil {
		return nil
	}
	return s.tenantSvc
}
