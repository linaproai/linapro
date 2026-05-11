// This file implements the platform tenant-status endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// TenantStatus updates tenant lifecycle status.
func (c *ControllerV1) TenantStatus(ctx context.Context, req *v1.TenantStatusReq) (res *v1.TenantStatusRes, err error) {
	err = c.tenantSvc.ChangeStatus(ctx, req.Id, shared.TenantStatus(req.Status))
	if err != nil {
		return nil, err
	}
	return &v1.TenantStatusRes{}, nil
}
