package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	v1 "lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

const (
	menuRuntimePageComponentPath      = "system/plugin/dynamic-page"
	menuQueryKeyEmbeddedSource        = "embeddedSrc"
	menuQueryKeyPluginAccessMode      = "pluginAccessMode"
	menuPluginAccessModeEmbeddedMount = "embedded-mount"
)

// GetAll returns all menus for the current user in Vben route format
func (c *ControllerV1) GetAll(ctx context.Context, req *v1.GetAllReq) (res *v1.GetAllRes, err error) {
	// Get user ID from business context (set by auth middleware)
	bizCtx := c.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return &v1.GetAllRes{List: []*v1.MenuRouteItem{}}, nil
	}
	userId := bizCtx.UserId

	// Check if super admin
	isSuperAdmin := c.roleSvc.IsSuperAdmin(ctx, userId)

	var menuTree []*menusvc.MenuItem

	statusNormal := 1
	if isSuperAdmin {
		// Super admin gets all enabled menus
		allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
			Status: &statusNormal,
		})
		if err != nil {
			return nil, err
		}
		menuTree = c.menuSvc.BuildTree(allMenus.List)
	} else {
		// Regular user gets menus based on roles
		menuIds, err := c.roleSvc.GetUserMenuIds(ctx, userId)
		if err != nil {
			return nil, err
		}
		if len(menuIds) > 0 {
			allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
				Status: &statusNormal,
			})
			if err != nil {
				return nil, err
			}
			// Filter menus by user's menu IDs
			menuMap := make(map[int]bool)
			for _, id := range menuIds {
				menuMap[id] = true
			}
			filteredMenus := make([]*menusvc.MenuItem, 0)
			for _, m := range allMenus.List {
				if menuMap[m.Id] {
					filteredMenus = append(filteredMenus, &menusvc.MenuItem{
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
						Children:   make([]*menusvc.MenuItem, 0),
					})
				}
			}
			menuTree = buildFilteredTree(filteredMenus)
		}
	}

	// Convert to Vben route format
	routes := convertToRouteItems(menuTree)

	return &v1.GetAllRes{List: routes}, nil
}

// buildFilteredTree builds a tree from filtered menu items
func buildFilteredTree(items []*menusvc.MenuItem) []*menusvc.MenuItem {
	nodeMap := make(map[int]*menusvc.MenuItem)
	for _, m := range items {
		nodeMap[m.Id] = m
	}

	var roots []*menusvc.MenuItem
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

// convertToRouteItems converts menu items to Vben route format
func convertToRouteItems(items []*menusvc.MenuItem) []*v1.MenuRouteItem {
	result := make([]*v1.MenuRouteItem, 0, len(items))
	for _, item := range items {
		if item.Type == "B" {
			continue
		}

		routeName := generateRouteName(item)
		routePath := generateRoutePath(item)
		route := &v1.MenuRouteItem{
			Id:       item.Id,
			ParentId: item.ParentId,
			Name:     routeName,
			Path:     routePath,
			Meta: &v1.MenuRouteMeta{
				Title:            item.Name,
				Icon:             item.Icon,
				HideInMenu:       item.Visible == 0,
				KeepAlive:        item.IsCache == 1,
				Order:            item.Sort,
				Authority:        item.Perms,
				IgnoreAccess:     false,
				HideInBreadcrumb: false,
				HideInTab:        false,
				ActiveIcon:       "",
			},
		}
		menuQuery := parseMenuQueryParams(item.QueryParam)
		if len(menuQuery) > 0 {
			route.Meta.Query = menuQuery
		}

		// Runtime hosted assets and generic external links must be converted into
		// router-level iframe/new-window semantics before normal view resolution.
		if menuLinkTarget := normalizeMenuLinkTarget(item.Path); item.Type == "M" && menuLinkTarget != "" {
			route.Name = buildMenuLinkRouteName(item)
			route.Path = buildMenuLinkRoutePath(item)
			if isRuntimeEmbeddedMountMenu(item, menuQuery) {
				// Embedded mount keeps the host runtime shell component while the
				// actual asset URL is forwarded through route query parameters.
				route.Component = generateComponentPath(item.Component)
				route.Meta.Query = mergeMenuQueryParams(menuQuery, map[string]string{
					menuQueryKeyEmbeddedSource:   menuLinkTarget,
					menuQueryKeyPluginAccessMode: menuPluginAccessModeEmbeddedMount,
				})
			} else if item.IsFrame == 1 {
				route.Component = "BasicLayout"
				route.Meta.Link = menuLinkTarget
				route.Meta.OpenInNewWindow = true
			} else {
				route.Component = "IFrameView"
				route.Meta.IframeSrc = menuLinkTarget
			}
		} else if item.Type == "M" {
			// Set component for menu type (M) - actual pages.
			route.Component = generateComponentPath(item.Component)
		}

		// Convert children recursively, excluding button-type nodes.
		if len(item.Children) > 0 {
			route.Children = convertToRouteItems(item.Children)
		}

		// Set redirect for directory type (D) with children.
		if item.Type == "D" && len(route.Children) > 0 {
			route.Redirect = route.Children[0].Path
		}

		result = append(result, route)
	}
	return result
}

// generateRouteName generates route name from menu
func generateRouteName(item *menusvc.MenuItem) string {
	if normalizeMenuLinkTarget(item.Path) != "" {
		return buildMenuLinkRouteName(item)
	}
	if item.Path != "" {
		// Convert path to PascalCase name
		return toPascalCase(item.Path)
	}
	return toPascalCase(item.Name)
}

// generateRoutePath generates route path
func generateRoutePath(item *menusvc.MenuItem) string {
	if normalizeMenuLinkTarget(item.Path) != "" {
		return buildMenuLinkRoutePath(item)
	}
	if item.Path == "" {
		return ""
	}
	// For child routes (parentId != 0), return relative path without leading /
	// Vue Router will append this to parent path
	if item.ParentId != 0 {
		path := item.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		return path
	}
	// For root routes, ensure path starts with /
	if item.Path[0] != '/' {
		return "/" + item.Path
	}
	return item.Path
}

// generateComponentPath generates component path for Vben
func generateComponentPath(component string) string {
	if component == "" {
		return ""
	}
	// Vben expects component path like #/views/xxx/index.vue
	if component[0] == '#' {
		return component
	}
	return "#/views/" + component
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]byte, 0, len(s))
	upperNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '_' || c == '/' || c == ' ' {
			upperNext = true
			continue
		}
		if upperNext {
			if c >= 'a' && c <= 'z' {
				c = c - 32
			}
			upperNext = false
		}
		result = append(result, c)
	}
	return string(result)
}

// normalizeMenuLinkTarget returns the real target URL for iframe/new-window menus.
func normalizeMenuLinkTarget(path string) string {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return ""
	}

	lowerPath := strings.ToLower(trimmedPath)
	if strings.HasPrefix(lowerPath, "http://") || strings.HasPrefix(lowerPath, "https://") {
		return trimmedPath
	}

	normalizedHostedPath := "/" + strings.TrimLeft(trimmedPath, "/")
	if strings.HasPrefix(normalizedHostedPath, "/plugin-assets/") {
		return normalizedHostedPath
	}
	return ""
}

func parseMenuQueryParams(queryParam string) map[string]string {
	trimmedQuery := strings.TrimSpace(queryParam)
	if trimmedQuery == "" {
		return nil
	}

	rawQuery := make(map[string]interface{})
	if err := json.Unmarshal([]byte(trimmedQuery), &rawQuery); err != nil {
		return nil
	}

	query := make(map[string]string, len(rawQuery))
	for key, value := range rawQuery {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" || value == nil {
			continue
		}
		query[trimmedKey] = strings.TrimSpace(fmt.Sprint(value))
	}
	if len(query) == 0 {
		return nil
	}
	return query
}

func mergeMenuQueryParams(base map[string]string, overrides map[string]string) map[string]string {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}

	merged := make(map[string]string, len(base)+len(overrides))
	for key, value := range base {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		merged[key] = value
	}
	for key, value := range overrides {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func isRuntimeEmbeddedMountMenu(item *menusvc.MenuItem, menuQuery map[string]string) bool {
	if normalizeMenuComponentPath(item.Component) != menuRuntimePageComponentPath {
		return false
	}
	return strings.TrimSpace(menuQuery[menuQueryKeyPluginAccessMode]) == menuPluginAccessModeEmbeddedMount
}

func normalizeMenuComponentPath(component string) string {
	normalizedComponent := strings.TrimSpace(component)
	normalizedComponent = strings.TrimPrefix(normalizedComponent, "#")
	normalizedComponent = strings.TrimPrefix(normalizedComponent, "/")
	normalizedComponent = strings.TrimPrefix(normalizedComponent, "views/")
	normalizedComponent = strings.TrimPrefix(normalizedComponent, "views\\")
	normalizedComponent = strings.TrimSuffix(normalizedComponent, ".vue")
	return strings.ReplaceAll(normalizedComponent, "\\", "/")
}

// buildMenuLinkRoutePath creates one internal router path for a menu that actually targets a hosted asset URL.
func buildMenuLinkRoutePath(item *menusvc.MenuItem) string {
	slug := buildMenuLinkRouteSlug(item)
	if item.ParentId == 0 {
		return "/" + slug
	}
	return slug
}

// buildMenuLinkRouteName creates a stable route name for hosted-link menus.
func buildMenuLinkRouteName(item *menusvc.MenuItem) string {
	return toPascalCase(buildMenuLinkRouteSlug(item))
}

func buildMenuLinkRouteSlug(item *menusvc.MenuItem) string {
	var builder strings.Builder
	builder.WriteString("link-")
	builder.WriteString(strconv.Itoa(item.Id))
	builder.WriteString("-")

	for _, currentRune := range strings.ToLower(strings.TrimSpace(item.Path)) {
		if unicode.IsLetter(currentRune) || unicode.IsDigit(currentRune) {
			builder.WriteRune(currentRune)
			continue
		}
		builder.WriteRune('-')
	}

	slug := strings.Trim(builder.String(), "-")
	slug = collapseHyphen(slug)
	if slug == "" {
		return "link-" + strconv.Itoa(item.Id)
	}
	return slug
}

func collapseHyphen(value string) string {
	var (
		builder      strings.Builder
		previousDash bool
	)

	for _, currentRune := range value {
		if currentRune == '-' {
			if previousDash {
				continue
			}
			previousDash = true
			builder.WriteRune(currentRune)
			continue
		}
		previousDash = false
		builder.WriteRune(currentRune)
	}
	return builder.String()
}
