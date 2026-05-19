// This file implements the localized menu tree list endpoint for the menu
// management page.

package menu

import (
	"context"

	"lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
	"lina-core/pkg/menutype"
)

// List queries menus with filters and returns tree nodes.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	// Query flat list
	out, err := c.menuSvc.List(ctx, menusvc.ListInput{
		Name:      req.Name,
		Status:    enabledPtrToInt(req.Status),
		Visible:   visibilityPtrToInt(req.Visible),
		Localized: true,
	})
	if err != nil {
		return nil, err
	}

	// Build tree structure
	tree := c.menuSvc.BuildTree(out.List)

	// Convert to API response
	items := make([]*v1.MenuItem, 0, len(tree))
	for _, node := range tree {
		items = append(items, convertMenuItem(node))
	}

	return &v1.ListRes{List: items}, nil
}

// convertMenuItem converts service MenuItem to API MenuItem recursively.
func convertMenuItem(node *menusvc.MenuItem) *v1.MenuItem {
	item := &v1.MenuItem{
		Id:         node.Id,
		ParentId:   node.ParentId,
		Name:       node.Name,
		Path:       node.Path,
		Component:  node.Component,
		Perms:      node.Perms,
		Icon:       node.Icon,
		Type:       menutype.Code(node.Type),
		Sort:       node.Sort,
		Visible:    statusflagVisibility(node.Visible),
		Status:     statusflagEnabled(node.Status),
		IsFrame:    statusflagYesNo(node.IsFrame),
		IsCache:    statusflagYesNo(node.IsCache),
		QueryParam: node.QueryParam,
		Remark:     node.Remark,
		CreatedAt:  node.CreatedAt,
		UpdatedAt:  node.UpdatedAt,
		Children:   make([]*v1.MenuItem, 0),
	}

	for _, child := range node.Children {
		item.Children = append(item.Children, convertMenuItem(child))
	}

	return item
}
