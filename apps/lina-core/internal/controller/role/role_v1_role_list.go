package role

import (
	"context"

	"lina-core/api/role/v1"
	"lina-core/internal/service/role"
)

// RoleList queries roles with pagination and filters.
func (c *ControllerV1) RoleList(ctx context.Context, req *v1.RoleListReq) (res *v1.RoleListRes, err error) {
	// Prepare status filter
	var status *int
	if req.Status > 0 {
		status = &req.Status
	}

	// Query role list
	out, err := c.roleSvc.List(ctx, role.ListInput{
		Name:   req.Name,
		Key:    req.Key,
		Status: status,
		Page:   req.Page,
		Size:   req.Size,
	})
	if err != nil {
		return nil, err
	}

	// Convert to API response format
	list := make([]*v1.RoleListItem, 0, len(out.List))
	for _, r := range out.List {
		list = append(list, &v1.RoleListItem{
			Id:        r.Id,
			Name:      r.Name,
			Key:       r.Key,
			Sort:      r.Sort,
			DataScope: r.DataScope,
			Status:    r.Status,
			Remark:    r.Remark,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	return &v1.RoleListRes{
		List:  list,
		Total: out.Total,
	}, nil
}
