package menu

import (
	"context"

	"lina-core/api/menu/v1"
)

// Get returns the detail of the specified menu.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	// Get menu detail
	m, err := c.menuSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Get parent name
	parentName := c.menuSvc.GetParentName(ctx, m.ParentId)

	return &v1.GetRes{
		MenuItem: &v1.MenuItem{
			Id:         m.Id,
			ParentId:   m.ParentId,
			Name:       m.Name,
			Path:       m.Path,
			Component:  m.Component,
			Perms:      m.Perms,
			Icon:       m.Icon,
			Type:       m.Type,
			Sort:       m.Sort,
			Visible:    m.Visible,
			Status:     m.Status,
			IsFrame:    m.IsFrame,
			IsCache:    m.IsCache,
			QueryParam: m.QueryParam,
			Remark:     m.Remark,
		},
		ParentName: parentName,
	}, nil
}
