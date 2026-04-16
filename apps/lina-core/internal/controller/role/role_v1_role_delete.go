package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleDelete deletes the specified role.
func (c *ControllerV1) RoleDelete(ctx context.Context, req *v1.RoleDeleteReq) (res *v1.RoleDeleteRes, err error) {
	// Delete role
	err = c.roleSvc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.RoleDeleteRes{}, nil
}
