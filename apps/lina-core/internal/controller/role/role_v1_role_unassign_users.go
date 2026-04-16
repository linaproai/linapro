package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleUnassignUsers removes multiple users from the specified role.
func (c *ControllerV1) RoleUnassignUsers(ctx context.Context, req *v1.RoleUnassignUsersReq) (res *v1.RoleUnassignUsersRes, err error) {
	// Batch unassign users from role
	err = c.roleSvc.UnassignUsers(ctx, req.Id, req.UserIds)
	if err != nil {
		return nil, err
	}

	return &v1.RoleUnassignUsersRes{}, nil
}
