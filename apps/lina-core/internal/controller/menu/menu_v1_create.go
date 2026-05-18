package menu

import (
	"context"

	"lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

// Create creates a new menu.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := c.menuSvc.Create(ctx, menusvc.CreateInput{
		ParentId:   req.ParentId,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Perms:      req.Perms,
		Icon:       req.Icon,
		Type:       req.Type,
		Sort:       req.Sort,
		Visible:    req.Visible,
		Status:     req.Status,
		IsFrame:    req.IsFrame,
		IsCache:    req.IsCache,
		QueryParam: req.QueryParam,
		Remark:     req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &v1.CreateRes{Id: id}, nil
}
