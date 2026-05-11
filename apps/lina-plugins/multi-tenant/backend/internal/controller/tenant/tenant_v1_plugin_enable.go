// This file implements the tenant-scoped plugin-enable endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
)

// TenantPluginEnable enables a tenant-scoped plugin for the current tenant.
func (c *ControllerV1) TenantPluginEnable(ctx context.Context, req *v1.TenantPluginEnableReq) (res *v1.TenantPluginEnableRes, err error) {
	if err = c.tenantPluginSvc.SetEnabled(ctx, req.PluginId, true); err != nil {
		return nil, err
	}
	return &v1.TenantPluginEnableRes{}, nil
}
