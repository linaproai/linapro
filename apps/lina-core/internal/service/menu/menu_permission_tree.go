// This file builds the role authorization tree projection used by role forms.

package menu

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/pkg/menutype"
)

const (
	// permissionTreeDynamicDirectoryID is a synthetic display-only node ID.
	permissionTreeDynamicDirectoryID = -1000001
	// permissionTreeDynamicButtonMenuID is a synthetic display-only node ID.
	permissionTreeDynamicButtonMenuID = -1000002
	// permissionTreeDynamicDirectorySort keeps dynamic permissions after host directories.
	permissionTreeDynamicDirectorySort = 10000
	// permissionTreeDynamicButtonMenuSort keeps orphan buttons after real dynamic menus.
	permissionTreeDynamicButtonMenuSort = 10000
)

// permissionTreeNode carries sort and synthetic-node metadata while building
// the public MenuTreeNode projection.
type permissionTreeNode struct {
	Id        int
	ParentId  int
	MenuKey   string
	Label     string
	Type      string
	Icon      string
	Sort      int
	Synthetic bool
	Children  []*permissionTreeNode
}

// buildPermissionTreeNodes creates a role authorization projection that keeps
// directories above menus and menus above buttons for a cleaner permission tree.
func (s *serviceImpl) buildPermissionTreeNodes(ctx context.Context, list []*entity.SysMenu) []*MenuTreeNode {
	nodeMap := make(map[int]*permissionTreeNode, len(list))
	for _, menu := range list {
		if menu == nil {
			continue
		}
		nodeMap[menu.Id] = newPermissionTreeNodeFromMenu(menu)
	}

	roots := make([]*permissionTreeNode, 0)
	for _, menu := range list {
		if menu == nil {
			continue
		}
		node := nodeMap[menu.Id]
		if node == nil {
			continue
		}
		parent, ok := nodeMap[menu.ParentId]
		if ok && parent != nil {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}

	roots = s.normalizePermissionTreeRoots(ctx, roots)
	return convertPermissionTreeNodes(roots)
}

// newPermissionTreeNodeFromMenu projects one persisted menu into the internal
// authorization tree node representation.
func newPermissionTreeNodeFromMenu(menu *entity.SysMenu) *permissionTreeNode {
	return &permissionTreeNode{
		Id:       menu.Id,
		ParentId: menu.ParentId,
		MenuKey:  menu.MenuKey,
		Label:    menu.Name,
		Type:     menu.Type,
		Icon:     menu.Icon,
		Sort:     menu.Sort,
		Children: make([]*permissionTreeNode, 0),
	}
}

// normalizePermissionTreeRoots moves root-level menus and buttons into a
// display-only dynamic permission directory.
func (s *serviceImpl) normalizePermissionTreeRoots(
	ctx context.Context,
	roots []*permissionTreeNode,
) []*permissionTreeNode {
	normalizedRoots := make([]*permissionTreeNode, 0, len(roots)+1)
	dynamicDirectory := s.newDynamicPermissionDirectory(ctx)
	var dynamicButtonMenu *permissionTreeNode

	for _, root := range roots {
		if root == nil {
			continue
		}
		rootType := normalizePermissionTreeMenuType(root.Type)
		switch rootType {
		case menutype.Directory:
			if IsStableCatalogKey(root.MenuKey) {
				s.normalizePermissionDirectoryNode(ctx, root)
				normalizedRoots = append(normalizedRoots, root)
				continue
			}
			root.ParentId = dynamicDirectory.Id
			dynamicDirectory.Children = append(dynamicDirectory.Children, root)
		case menutype.Button:
			if dynamicButtonMenu == nil {
				dynamicButtonMenu = s.newDynamicButtonPermissionMenu(ctx)
			}
			root.ParentId = dynamicButtonMenu.Id
			dynamicButtonMenu.Children = append(dynamicButtonMenu.Children, root)
		default:
			root.ParentId = dynamicDirectory.Id
			dynamicDirectory.Children = append(dynamicDirectory.Children, root)
		}
	}

	if dynamicButtonMenu != nil && len(dynamicButtonMenu.Children) > 0 {
		dynamicButtonMenu.ParentId = dynamicDirectory.Id
		dynamicDirectory.Children = append(dynamicDirectory.Children, dynamicButtonMenu)
	}
	if len(dynamicDirectory.Children) > 0 {
		s.normalizePermissionDirectoryNode(ctx, dynamicDirectory)
		normalizedRoots = append(normalizedRoots, dynamicDirectory)
	}
	return normalizedRoots
}

// normalizePermissionDirectoryNode keeps directories as grouping nodes, menu
// nodes as permission rows, and moves any direct buttons under a synthetic menu.
func (s *serviceImpl) normalizePermissionDirectoryNode(ctx context.Context, directory *permissionTreeNode) {
	if directory == nil {
		return
	}

	normalizedChildren := make([]*permissionTreeNode, 0, len(directory.Children)+1)
	buttonChildren := make([]*permissionTreeNode, 0)
	for _, child := range directory.Children {
		if child == nil {
			continue
		}
		switch normalizePermissionTreeMenuType(child.Type) {
		case menutype.Directory:
			s.normalizePermissionDirectoryNode(ctx, child)
			normalizedChildren = append(normalizedChildren, child)
		case menutype.Button:
			buttonChildren = append(buttonChildren, child)
		default:
			s.normalizePermissionMenuNode(ctx, child)
			normalizedChildren = append(normalizedChildren, child)
		}
	}

	if len(buttonChildren) > 0 {
		buttonMenu := s.newDynamicButtonPermissionMenu(ctx)
		buttonMenu.Id = syntheticButtonMenuIDForDirectory(directory.Id)
		buttonMenu.ParentId = directory.Id
		for _, child := range buttonChildren {
			child.ParentId = buttonMenu.Id
			buttonMenu.Children = append(buttonMenu.Children, child)
		}
		normalizedChildren = append(normalizedChildren, buttonMenu)
	}
	directory.Children = normalizedChildren
}

// normalizePermissionMenuNode keeps only button permissions under menu nodes.
// Other malformed descendants are left out of this menu row to preserve the
// directory -> menu -> button display contract.
func (s *serviceImpl) normalizePermissionMenuNode(ctx context.Context, menu *permissionTreeNode) {
	if menu == nil {
		return
	}

	buttonChildren := make([]*permissionTreeNode, 0, len(menu.Children))
	for _, child := range menu.Children {
		if child == nil {
			continue
		}
		if normalizePermissionTreeMenuType(child.Type) == menutype.Button {
			child.ParentId = menu.Id
			buttonChildren = append(buttonChildren, child)
		}
	}
	menu.Children = buttonChildren
}

// newDynamicPermissionDirectory creates the display-only directory used for
// dynamically discovered menu and permission nodes.
func (s *serviceImpl) newDynamicPermissionDirectory(ctx context.Context) *permissionTreeNode {
	label := "Dynamic Permissions"
	if s != nil && s.i18nSvc != nil {
		label = s.i18nSvc.Translate(ctx, "menu.dynamic-permissions.title", label)
	}
	return &permissionTreeNode{
		Id:        permissionTreeDynamicDirectoryID,
		ParentId:  0,
		Label:     label,
		Type:      menutype.Directory.String(),
		Icon:      "lucide:puzzle",
		Sort:      permissionTreeDynamicDirectorySort,
		Synthetic: true,
		Children:  make([]*permissionTreeNode, 0),
	}
}

// newDynamicButtonPermissionMenu creates the display-only menu used to hold
// root-level button permission nodes.
func (s *serviceImpl) newDynamicButtonPermissionMenu(ctx context.Context) *permissionTreeNode {
	label := "Runtime Route Permissions"
	if s != nil && s.i18nSvc != nil {
		label = s.i18nSvc.Translate(ctx, "menu.dynamic-permissions.route-permissions.title", label)
	}
	return &permissionTreeNode{
		Id:        permissionTreeDynamicButtonMenuID,
		ParentId:  permissionTreeDynamicDirectoryID,
		Label:     label,
		Type:      menutype.Menu.String(),
		Sort:      permissionTreeDynamicButtonMenuSort,
		Synthetic: true,
		Children:  make([]*permissionTreeNode, 0),
	}
}

// syntheticButtonMenuIDForDirectory derives a stable negative display ID for a
// directory-owned synthetic button container.
func syntheticButtonMenuIDForDirectory(directoryID int) int {
	if directoryID < 0 {
		return directoryID*10 - 1
	}
	return -2000000 - directoryID
}

// convertPermissionTreeNodes strips internal sort metadata from tree nodes.
func convertPermissionTreeNodes(nodes []*permissionTreeNode) []*MenuTreeNode {
	result := make([]*MenuTreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		item := &MenuTreeNode{
			Id:       node.Id,
			ParentId: node.ParentId,
			Label:    node.Label,
			Type:     node.Type,
			Icon:     node.Icon,
			Children: convertPermissionTreeNodes(node.Children),
		}
		result = append(result, item)
	}
	return result
}

// normalizePermissionTreeMenuType normalizes raw menu types for authorization
// tree structural decisions.
func normalizePermissionTreeMenuType(value string) menutype.Code {
	normalized := menutype.Normalize(value)
	if normalized == "" && strings.TrimSpace(value) != "" {
		return menutype.Menu
	}
	return normalized
}
