package menu

import (
	"context"

	"lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

// Update updates the specified menu.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	// Convert required string fields to pointers for service
	var path, component, perms, icon, menuType, queryParam, remark *string
	if req.Path != "" {
		path = &req.Path
	}
	if req.Component != "" {
		component = &req.Component
	}
	if req.Perms != "" {
		perms = &req.Perms
	}
	if req.Icon != "" {
		icon = &req.Icon
	}
	if req.Type != "" {
		menuType = &req.Type
	}
	if req.QueryParam != "" {
		queryParam = &req.QueryParam
	}
	if req.Remark != "" {
		remark = &req.Remark
	}

	err = c.menuSvc.Update(ctx, menusvc.UpdateInput{
		Id:         req.Id,
		ParentId:   req.ParentId,
		Name:       req.Name,
		Path:       path,
		Component:  component,
		Perms:      perms,
		Icon:       icon,
		Type:       menuType,
		Sort:       req.Sort,
		Visible:    req.Visible,
		Status:     req.Status,
		IsFrame:    req.IsFrame,
		IsCache:    req.IsCache,
		QueryParam: queryParam,
		Remark:     remark,
	})
	if err != nil {
		return nil, err
	}

	return &v1.UpdateRes{}, nil
}
