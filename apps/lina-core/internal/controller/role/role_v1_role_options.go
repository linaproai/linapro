package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleOptions returns enabled role options.
func (c *ControllerV1) RoleOptions(ctx context.Context, req *v1.RoleOptionsReq) (res *v1.RoleOptionsRes, err error) {
	// Get role options
	list, err := c.roleSvc.GetOptions(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to API response
	items := make([]*v1.RoleOptionItem, 0, len(list))
	for _, r := range list {
		items = append(items, &v1.RoleOptionItem{
			Id:   r.Id,
			Name: r.Name,
			Key:  r.Key,
		})
	}

	return &v1.RoleOptionsRes{List: items}, nil
}
