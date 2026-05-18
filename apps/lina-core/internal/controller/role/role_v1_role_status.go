package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleStatus updates the status of the specified role.
func (c *ControllerV1) RoleStatus(ctx context.Context, req *v1.RoleStatusReq) (res *v1.RoleStatusRes, err error) {
	// Update role status
	err = c.roleSvc.UpdateStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}

	return &v1.RoleStatusRes{}, nil
}
