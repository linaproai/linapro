package user

import (
	"context"

	"lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
)

// BatchUpdate updates selected users in one transaction.
func (c *ControllerV1) BatchUpdate(ctx context.Context, req *v1.BatchUpdateReq) (res *v1.BatchUpdateRes, err error) {
	if err = c.userSvc.BatchUpdate(ctx, usersvc.BatchUpdateInput{
		Ids:          req.Ids,
		UpdateStatus: req.UpdateStatus,
		Status:       req.Status,
		UpdateRoles:  req.UpdateRoles,
		RoleIds:      req.RoleIds,
		UpdateTenant: req.UpdateTenant,
		TenantIds:    req.TenantIds,
	}); err != nil {
		return nil, err
	}
	return &v1.BatchUpdateRes{}, nil
}
