// This file implements the platform tenant-delete endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
)

// TenantDelete deletes a tenant after lifecycle guard checks pass.
func (c *ControllerV1) TenantDelete(ctx context.Context, req *v1.TenantDeleteReq) (res *v1.TenantDeleteRes, err error) {
	if err = c.tenantSvc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.TenantDeleteRes{}, nil
}
