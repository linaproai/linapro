// This file implements the tenant-scoped plugin-disable endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
)

// TenantPluginDisable disables a tenant-scoped plugin for the current tenant.
func (c *ControllerV1) TenantPluginDisable(ctx context.Context, req *v1.TenantPluginDisableReq) (res *v1.TenantPluginDisableRes, err error) {
	if err = c.tenantPluginSvc.SetEnabled(ctx, req.PluginId, false); err != nil {
		return nil, err
	}
	return &v1.TenantPluginDisableRes{}, nil
}
