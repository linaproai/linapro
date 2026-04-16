package role

import (
	"context"

	"lina-core/api/role/v1"
	"lina-core/internal/service/role"
)

// RoleUsers queries the users assigned to the specified role.
func (c *ControllerV1) RoleUsers(ctx context.Context, req *v1.RoleUsersReq) (res *v1.RoleUsersRes, err error) {
	// Prepare status filter
	var status *int
	if req.Status > 0 {
		status = &req.Status
	}

	// Query users assigned to role
	out, err := c.roleSvc.GetUsers(ctx, role.GetUsersInput{
		RoleId:   req.Id,
		Username: req.Username,
		Phone:    req.Phone,
		Status:   status,
		Page:     req.Page,
		Size:     req.Size,
	})
	if err != nil {
		return nil, err
	}

	// Convert to API response
	items := make([]*v1.RoleUserItem, 0, len(out.List))
	for _, u := range out.List {
		items = append(items, &v1.RoleUserItem{
			Id:        u.Id,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Phone:     u.Phone,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		})
	}

	return &v1.RoleUsersRes{
		List:  items,
		Total: out.Total,
	}, nil
}
