// This file implements the tenant-scoped plugin-list endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
)

// TenantPluginList returns tenant-controllable plugins.
func (c *ControllerV1) TenantPluginList(ctx context.Context, _ *v1.TenantPluginListReq) (res *v1.TenantPluginListRes, err error) {
	out, err := c.tenantPluginSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.TenantPluginEntity, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPITenantPlugin(item))
	}
	return &v1.TenantPluginListRes{List: list, Total: out.Total}, nil
}
