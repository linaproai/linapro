// This file implements the platform tenant-create endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
)

// TenantCreate creates a tenant.
func (c *ControllerV1) TenantCreate(ctx context.Context, req *v1.TenantCreateReq) (res *v1.TenantCreateRes, err error) {
	id, err := c.tenantSvc.Create(ctx, tenantsvc.CreateInput{
		Code:   req.Code,
		Name:   req.Name,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantCreateRes{Id: id}, nil
}
