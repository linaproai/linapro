package user

import (
	"context"
	"strings"

	v1 "lina-core/api/user/v1"
	"lina-core/internal/service/menu"
)

// GetInfo returns current logged-in user information
func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	user, err := c.userSvc.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	realName := user.Nickname
	if realName == "" {
		realName = user.Username
	}

	accessContext, err := c.roleSvc.GetUserAccessContext(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	roleNames := accessContext.RoleNames
	permissions := accessContext.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	if len(roleNames) == 0 {
		roleNames = []string{}
	}

	isSuperAdmin := accessContext.IsSuperAdmin

	// Get user menus
	var menuTree []*menu.MenuItem

	if isSuperAdmin {
		// Super admin gets all enabled menus
		allMenus, err := c.menuSvc.List(ctx, menu.ListInput{
			Status: intPtr(1),
		})
		if err != nil {
			return nil, err
		}
		menuTree = c.menuSvc.BuildTree(allMenus.List)
		// Add wildcard permission for super admin
		permissions = append(permissions, "*:*:*")
	} else {
		// Regular user gets menus based on roles
		menuIds := accessContext.MenuIds
		if len(menuIds) > 0 {
			allMenus, err := c.menuSvc.List(ctx, menu.ListInput{
				Status: intPtr(1),
			})
			if err != nil {
				return nil, err
			}
			// Filter menus by user's menu IDs
			menuMap := make(map[int]bool)
			for _, id := range menuIds {
				menuMap[id] = true
			}
			filteredMenus := make([]*menu.MenuItem, 0)
			for _, m := range allMenus.List {
				if menuMap[m.Id] {
					filteredMenus = append(filteredMenus, &menu.MenuItem{
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
						CreatedAt:  m.CreatedAt.String(),
						UpdatedAt:  m.UpdatedAt.String(),
						Children:   make([]*menu.MenuItem, 0),
					})
				}
			}
			menuTree = buildFilteredTree(filteredMenus)
		}
	}

	return &v1.GetInfoRes{
		UserId:      user.Id,
		Username:    user.Username,
		RealName:    realName,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Roles:       roleNames,
		HomePath:    resolveHomePath(menuTree),
		Menus:       convertToMenuTree(menuTree),
		Permissions: permissions,
	}, nil
}

func intPtr(i int) *int {
	return &i
}

// buildFilteredTree builds a tree from filtered menu items
func buildFilteredTree(items []*menu.MenuItem) []*menu.MenuItem {
	// Build map for quick lookup
	nodeMap := make(map[int]*menu.MenuItem)
	for _, m := range items {
		nodeMap[m.Id] = m
	}

	// Build tree
	var roots []*menu.MenuItem
	for _, m := range items {
		if m.ParentId == 0 {
			roots = append(roots, m)
		} else {
			if parent, ok := nodeMap[m.ParentId]; ok {
				parent.Children = append(parent.Children, m)
			}
		}
	}
	return roots
}

// resolveHomePath returns the first accessible internal route from the menu tree.
func resolveHomePath(items []*menu.MenuItem) string {
	if homePath := findFirstAccessiblePath(items, ""); homePath != "" {
		return homePath
	}
	return "/profile"
}

// findFirstAccessiblePath traverses the menu tree in order and returns the first accessible path.
func findFirstAccessiblePath(items []*menu.MenuItem, parentPath string) string {
	for _, item := range items {
		currentPath := joinMenuPath(parentPath, item.Path)
		if item.Type == "M" && item.IsFrame == 0 && currentPath != "" && !isExternalPath(currentPath) {
			return currentPath
		}
		if len(item.Children) == 0 {
			continue
		}
		if homePath := findFirstAccessiblePath(item.Children, currentPath); homePath != "" {
			return homePath
		}
	}
	return ""
}

// joinMenuPath combines the current menu path with its parent route path.
func joinMenuPath(parentPath string, currentPath string) string {
	currentPath = strings.TrimSpace(currentPath)
	if currentPath == "" {
		return parentPath
	}
	if strings.HasPrefix(currentPath, "/") {
		return currentPath
	}
	if parentPath == "" {
		return "/" + strings.TrimLeft(currentPath, "/")
	}
	return strings.TrimRight(parentPath, "/") + "/" + strings.TrimLeft(currentPath, "/")
}

// isExternalPath reports whether the path points to an external address.
func isExternalPath(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func convertToMenuTree(items []*menu.MenuItem) []*v1.MenuTree {
	result := make([]*v1.MenuTree, 0, len(items))
	for _, item := range items {
		if item.Type == "B" {
			continue
		}

		node := &v1.MenuTree{
			Id:        item.Id,
			ParentId:  item.ParentId,
			Name:      item.Name,
			Path:      item.Path,
			Component: item.Component,
			Perms:     item.Perms,
			Icon:      item.Icon,
			Type:      item.Type,
			Sort:      item.Sort,
			Visible:   item.Visible,
			Status:    item.Status,
			IsFrame:   item.IsFrame,
			IsCache:   item.IsCache,
			Children:  convertToMenuTree(item.Children),
		}
		result = append(result, node)
	}
	return result
}
