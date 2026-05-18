package menu

import (
	"context"

	"lina-core/api/menu/v1"
)

// RoleMenuTree returns the menu tree and checked menu IDs for a role.
func (c *ControllerV1) RoleMenuTree(ctx context.Context, req *v1.RoleMenuTreeReq) (res *v1.RoleMenuTreeRes, err error) {
	// Get role menu tree
	out, err := c.menuSvc.GetRoleMenuTree(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	// Convert to API response
	items := make([]*v1.MenuTreeNode, 0, len(out.Menus))
	for _, node := range out.Menus {
		items = append(items, convertMenuTreeNode(node))
	}

	return &v1.RoleMenuTreeRes{
		Menus:       items,
		CheckedKeys: out.CheckedKeys,
	}, nil
}
