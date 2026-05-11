// This file implements the platform tenant-list endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
)

// TenantList queries tenants by page.
func (c *ControllerV1) TenantList(ctx context.Context, req *v1.TenantListReq) (res *v1.TenantListRes, err error) {
	out, err := c.tenantSvc.List(ctx, tenantsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Code:     req.Code,
		Name:     req.Name,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.TenantEntity, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPITenant(item))
	}
	return &v1.TenantListRes{List: list, Total: out.Total}, nil
}
