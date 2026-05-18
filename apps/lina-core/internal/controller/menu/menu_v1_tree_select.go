package menu

import (
	"context"

	"lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

// TreeSelect returns the selectable menu tree.
func (c *ControllerV1) TreeSelect(ctx context.Context, req *v1.TreeSelectReq) (res *v1.TreeSelectRes, err error) {
	// Get menu tree
	nodes, err := c.menuSvc.GetTreeSelect(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to API response
	items := make([]*v1.MenuTreeNode, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, convertMenuTreeNode(node))
	}

	return &v1.TreeSelectRes{List: items}, nil
}

// convertMenuTreeNode converts service MenuTreeNode to API MenuTreeNode recursively.
func convertMenuTreeNode(node *menusvc.MenuTreeNode) *v1.MenuTreeNode {
	item := &v1.MenuTreeNode{
		Id:       node.Id,
		ParentId: node.ParentId,
		Label:    node.Label,
		Type:     node.Type,
		Icon:     node.Icon,
		Children: make([]*v1.MenuTreeNode, 0),
	}

	for _, child := range node.Children {
		item.Children = append(item.Children, convertMenuTreeNode(child))
	}

	return item
}
