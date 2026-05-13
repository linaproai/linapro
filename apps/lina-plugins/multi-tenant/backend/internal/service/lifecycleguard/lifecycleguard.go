// Package lifecycleguard implements multi-tenant plugin lifecycle guard checks.
package lifecycleguard

import (
	"context"

	"lina-plugin-multi-tenant/backend/internal/service/tenant"
)

const (
	// ReasonUninstallTenantsExist is returned when existing tenants block uninstall.
	ReasonUninstallTenantsExist = "plugin.multi-tenant.uninstall_blocked.tenants_exist"
	// ReasonDisableTenantsExist is returned when existing tenants block disable.
	ReasonDisableTenantsExist = "plugin.multi-tenant.disable_blocked.tenants_exist"
)

// Guard implements plugin-owned lifecycle guard checks.
type Guard struct {
	tenantSvc tenant.Service
}

// New creates and returns a lifecycle guard.
func New(tenantSvc tenant.Service) *Guard {
	return &Guard{tenantSvc: tenantSvc}
}

// CanUninstall rejects uninstall while tenants exist.
func (g *Guard) CanUninstall(ctx context.Context) (bool, string, error) {
	count, err := g.tenantSvc.CountExisting(ctx)
	if err != nil {
		return false, ReasonUninstallTenantsExist, err
	}
	if count > 0 {
		return false, ReasonUninstallTenantsExist, nil
	}
	return true, "", nil
}

// CanDisable rejects global disable while tenants exist.
func (g *Guard) CanDisable(ctx context.Context) (bool, string, error) {
	count, err := g.tenantSvc.CountExisting(ctx)
	if err != nil {
		return false, ReasonDisableTenantsExist, err
	}
	if count > 0 {
		return false, ReasonDisableTenantsExist, nil
	}
	return true, "", nil
}

// CanTenantDelete reserves the cross-plugin tenant delete guard surface.
func (g *Guard) CanTenantDelete(ctx context.Context, tenantID int) (bool, string, error) {
	return true, "", nil
}
