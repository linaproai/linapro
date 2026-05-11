// This file implements the platform tenant-update endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
)

// TenantUpdate updates tenant profile fields.
func (c *ControllerV1) TenantUpdate(ctx context.Context, req *v1.TenantUpdateReq) (res *v1.TenantUpdateRes, err error) {
	err = c.tenantSvc.Update(ctx, tenantsvc.UpdateInput{
		Id:     req.Id,
		Name:   req.Name,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantUpdateRes{}, nil
}
