package role

import (
	"context"

	"lina-core/api/role/v1"
	"lina-core/internal/service/role"
)

// RoleUpdate updates the specified role.
func (c *ControllerV1) RoleUpdate(ctx context.Context, req *v1.RoleUpdateReq) (res *v1.RoleUpdateRes, err error) {
	// Update role
	err = c.roleSvc.Update(ctx, role.UpdateInput{
		Id:        req.Id,
		Name:      req.Name,
		Key:       req.Key,
		Sort:      &req.Sort,
		DataScope: &req.DataScope,
		Status:    &req.Status,
		Remark:    &req.Remark,
		MenuIds:   req.MenuIds,
	})
	if err != nil {
		return nil, err
	}

	return &v1.RoleUpdateRes{}, nil
}
