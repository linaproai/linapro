package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleGet returns the detail of the specified role.
func (c *ControllerV1) RoleGet(ctx context.Context, req *v1.RoleGetReq) (res *v1.RoleGetRes, err error) {
	// Get role detail with menu IDs
	out, err := c.roleSvc.GetDetail(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Convert to API response
	createdAt := ""
	if out.Role.CreatedAt != nil {
		createdAt = out.Role.CreatedAt.String()
	}
	updatedAt := ""
	if out.Role.UpdatedAt != nil {
		updatedAt = out.Role.UpdatedAt.String()
	}

	return &v1.RoleGetRes{
		Id:        out.Role.Id,
		Name:      out.Role.Name,
		Key:       out.Role.Key,
		Sort:      out.Role.Sort,
		DataScope: out.Role.DataScope,
		Status:    out.Role.Status,
		Remark:    out.Role.Remark,
		MenuIds:   out.MenuIds,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
