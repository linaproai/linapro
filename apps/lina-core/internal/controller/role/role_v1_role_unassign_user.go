package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleUnassignUser removes one user from the specified role.
func (c *ControllerV1) RoleUnassignUser(ctx context.Context, req *v1.RoleUnassignUserReq) (res *v1.RoleUnassignUserRes, err error) {
	// Unassign user from role
	err = c.roleSvc.UnassignUser(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}

	return &v1.RoleUnassignUserRes{}, nil
}
