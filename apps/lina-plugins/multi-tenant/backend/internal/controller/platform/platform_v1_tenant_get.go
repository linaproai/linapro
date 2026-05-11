// This file implements the platform tenant-detail endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
)

// TenantGet retrieves tenant details.
func (c *ControllerV1) TenantGet(ctx context.Context, req *v1.TenantGetReq) (res *v1.TenantGetRes, err error) {
	item, err := c.tenantSvc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TenantGetRes{TenantEntity: toAPITenant(item)}, nil
}
