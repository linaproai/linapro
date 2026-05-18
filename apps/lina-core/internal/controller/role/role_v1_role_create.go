package role

import (
	"context"

	"lina-core/api/role/v1"
	"lina-core/internal/service/role"
)

// RoleCreate creates a new role.
func (c *ControllerV1) RoleCreate(ctx context.Context, req *v1.RoleCreateReq) (res *v1.RoleCreateRes, err error) {
	// Create role
	id, err := c.roleSvc.Create(ctx, role.CreateInput{
		Name:      req.Name,
		Key:       req.Key,
		Sort:      req.Sort,
		DataScope: req.DataScope,
		Status:    req.Status,
		Remark:    req.Remark,
		MenuIds:   req.MenuIds,
	})
	if err != nil {
		return nil, err
	}

	return &v1.RoleCreateRes{Id: id}, nil
}
