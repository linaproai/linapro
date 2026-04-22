package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	v1 "lina-core/api/menu/v1"
	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
	"lina-core/pkg/menutype"
)

// Runtime menu route conversion constants shared by hosted-asset menu items.
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
			menuTree = buildFilteredTree(allMenus.List, menuIds)
		}
	}

	// Convert to Vben route format
	routes := convertToRouteItems(menuTree)

	return &v1.GetAllRes{List: routes}, nil
}

// buildFilteredTree builds a tree from one flat menu list and automatically
// keeps ancestor directories required by the selected menu IDs.
func buildFilteredTree(allMenus []*entity.SysMenu, selectedIDs []int) []*menusvc.MenuItem {
	if len(allMenus) == 0 || len(selectedIDs) == 0 {
		return []*menusvc.MenuItem{}
	}

	entityMap := make(map[int]*entity.SysMenu, len(allMenus))
	for _, item := range allMenus {
		if item == nil {
			continue
		}
		entityMap[item.Id] = item
	}

	selectedMap := make(map[int]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		currentID := id
		for currentID > 0 {
			if _, ok := selectedMap[currentID]; ok {
				break
			}
			selectedMap[currentID] = struct{}{}
			parent, ok := entityMap[currentID]
			if !ok {
				break
			}
			currentID = parent.ParentId
		}
	}

	items := make([]*menusvc.MenuItem, 0, len(selectedMap))
	for _, item := range allMenus {
		if item == nil {
			continue
		}
		if _, ok := selectedMap[item.Id]; !ok {
			continue
		}
		items = append(items, cloneMenuItem(item))
	}

	nodeMap := make(map[int]*menusvc.MenuItem, len(items))
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

// cloneMenuItem detaches one menu item from the service tree so controller-side
// filtering can rebuild children without mutating shared slices.
func cloneMenuItem(item *entity.SysMenu) *menusvc.MenuItem {
	if item == nil {
		return nil
	}
	createdAt := ""
	if item.CreatedAt != nil {
		createdAt = item.CreatedAt.String()
	}
	updatedAt := ""
	if item.UpdatedAt != nil {
		updatedAt = item.UpdatedAt.String()
	}
	return &menusvc.MenuItem{
		Id:         item.Id,
		ParentId:   item.ParentId,
		Name:       item.Name,
		Path:       item.Path,
		Component:  item.Component,
		Perms:      item.Perms,
		Icon:       item.Icon,
		Type:       item.Type,
		Sort:       item.Sort,
		Visible:    item.Visible,
		Status:     item.Status,
		IsFrame:    item.IsFrame,
		IsCache:    item.IsCache,
		QueryParam: item.QueryParam,
		Remark:     item.Remark,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Children:   []*menusvc.MenuItem{},
	}
}

// convertToRouteItems converts menu items to Vben route format
func convertToRouteItems(items []*menusvc.MenuItem) []*v1.MenuRouteItem {
	result := make([]*v1.MenuRouteItem, 0, len(items))
	for _, item := range items {
		if item.Type == menutype.Button.String() {
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
		if menuLinkTarget := normalizeMenuLinkTarget(item.Path); item.Type == menutype.Menu.String() && menuLinkTarget != "" {
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
		} else if item.Type == menutype.Menu.String() {
			// Set component for menu type (M) - actual pages.
			route.Component = generateComponentPath(item.Component)
		}

		// Convert children recursively, excluding button-type nodes.
		if len(item.Children) > 0 {
			route.Children = convertToRouteItems(item.Children)
		}

		// Hide empty directory menus so stable host catalogs do not leave empty
		// shells in navigation when all child menus are unavailable.
		if item.Type == menutype.Directory.String() && len(route.Children) == 0 {
			continue
		}

		// Set redirect for directory type (D) with children.
		if item.Type == menutype.Directory.String() && len(route.Children) > 0 {
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
	// Child routes normally use relative paths so Vue Router appends them to the
	// parent path. When the menu explicitly stores an absolute path, keep it so
	// grouped directory menus can preserve existing stable URLs.
	if item.ParentId != 0 {
		if item.Path[0] == '/' {
			return item.Path
		}
		return item.Path
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

// parseMenuQueryParams decodes the persisted JSON query payload into trimmed
// string pairs used by route metadata.
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

// mergeMenuQueryParams overlays non-empty query parameters onto an existing map
// and returns nil when the merged result is empty.
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

// isRuntimeEmbeddedMountMenu reports whether the menu entry points at the
// hosted runtime page component using embedded-mount semantics.
func isRuntimeEmbeddedMountMenu(item *menusvc.MenuItem, menuQuery map[string]string) bool {
	if normalizeMenuComponentPath(item.Component) != menuRuntimePageComponentPath {
		return false
	}
	return strings.TrimSpace(menuQuery[menuQueryKeyPluginAccessMode]) == menuPluginAccessModeEmbeddedMount
}

// normalizeMenuComponentPath normalizes one stored component path into the
// canonical relative path used by runtime menu comparisons.
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

// buildMenuLinkRouteSlug derives one stable router slug for menus that point
// to hosted assets or external links.
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

// collapseHyphen removes repeated hyphen runs while preserving single
// separators in generated route slugs.
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
