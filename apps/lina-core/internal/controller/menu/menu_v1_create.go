// This file implements menu creation and converts typed public flag fields into
// the menu service's integer model.

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
		Type:       string(req.Type),
		Sort:       req.Sort,
		Visible:    req.Visible.Int(),
		Status:     req.Status.Int(),
		IsFrame:    req.IsFrame.Int(),
		IsCache:    req.IsCache.Int(),
		QueryParam: req.QueryParam,
		Remark:     req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &v1.CreateRes{Id: id}, nil
}
