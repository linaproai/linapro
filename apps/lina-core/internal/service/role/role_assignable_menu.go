// This file centralizes role-menu assignability checks. It keeps tenant
// context, menu-key, permission-string, and plugin scope metadata decisions in
// one path so role form trees and sys_role_menu writes cannot diverge.

package role

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

const (
	// rolePluginMenuKeyPrefix prefixes plugin-owned menu keys and remarks.
	rolePluginMenuKeyPrefix = "plugin:"
	// rolePluginScopePlatformOnly is the persisted plugin scope that marks all
	// non-exempt plugin menus as platform-control-plane permissions.
	rolePluginScopePlatformOnly = "platform_only"
)

var (
	// rolePlatformOnlyMenuKeys marks host menu keys that belong to platform
	// governance regardless of their permission string.
	rolePlatformOnlyMenuKeys = map[string]struct{}{
		"platform":              {},
		"extension:plugin:list": {},
	}
	// roleTenantAllowedPermissionPrefixes are tenant self-service namespaces
	// that remain assignable in tenant contexts.
	roleTenantAllowedPermissionPrefixes = []string{
		"system:tenant:plugin:",
	}
	// roleTenantBlockedPermissionPrefixes are platform tenant-control namespaces
	// excluded from tenant role assignment unless specifically allow-listed.
	roleTenantBlockedPermissionPrefixes = []string{
		"system:tenant:",
	}
	// rolePlatformOnlyPermissions are host platform-governance permissions
	// whose menu rows must not be assigned by tenant contexts.
	rolePlatformOnlyPermissions = map[string]struct{}{
		"plugin:list":        {},
		"plugin:query":       {},
		"plugin:install":     {},
		"plugin:uninstall":   {},
		"plugin:enable":      {},
		"plugin:disable":     {},
		"plugin:edit":        {},
		"system:menu:add":    {},
		"system:menu:edit":   {},
		"system:menu:remove": {},
	}
)

// FilterAssignableMenus returns only menus assignable by the current request
// context after plugin enablement/permission filters and tenant boundaries.
func (s *serviceImpl) FilterAssignableMenus(ctx context.Context, menus []*entity.SysMenu) ([]*entity.SysMenu, error) {
	return s.filterAssignableMenus(ctx, menus)
}

// FilterAssignableMenuIDs filters checked role-menu IDs to the current
// assignable set so historical platform-only grants are not projected as
// tenant-editable permissions.
func (s *serviceImpl) FilterAssignableMenuIDs(ctx context.Context, menuIDs []int) ([]int, error) {
	normalized := normalizeRoleMenuIDs(menuIDs)
	if len(normalized) == 0 {
		return []int{}, nil
	}
	assignableIDs, err := s.assignableMenuIDSet(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]int, 0, len(normalized))
	for _, menuID := range normalized {
		if _, ok := assignableIDs[menuID]; ok {
			filtered = append(filtered, menuID)
		}
	}
	return filtered, nil
}

// EnsureAssignableMenuIDs rejects role writes that include menu IDs outside
// the current context's assignable set.
func (s *serviceImpl) EnsureAssignableMenuIDs(ctx context.Context, menuIDs []int) error {
	normalized := normalizeRoleMenuIDs(menuIDs)
	if len(normalized) == 0 {
		return nil
	}
	assignableIDs, err := s.assignableMenuIDSet(ctx)
	if err != nil {
		return err
	}
	for _, menuID := range normalized {
		if _, ok := assignableIDs[menuID]; !ok {
			return bizerr.NewCode(CodeRoleMenuAssignmentForbidden, bizerr.P("menuId", menuID))
		}
	}
	return nil
}

// assignableMenuIDSet loads the full permission topology before classifying
// submitted IDs so ancestor-based platform-only decisions cannot be bypassed by
// submitting only child IDs.
func (s *serviceImpl) assignableMenuIDSet(ctx context.Context) (map[int]struct{}, error) {
	menus, err := loadRoleAssignmentMenus(ctx)
	if err != nil {
		return nil, err
	}
	assignableMenus, err := s.filterAssignableMenus(ctx, menus)
	if err != nil {
		return nil, err
	}
	result := make(map[int]struct{}, len(assignableMenus))
	for _, menu := range assignableMenus {
		if menu == nil || menu.Id <= 0 {
			continue
		}
		result[menu.Id] = struct{}{}
	}
	return result, nil
}

// filterAssignableMenus applies plugin permission visibility first and then
// evaluates platform-only boundaries for the current tenant context.
func (s *serviceImpl) filterAssignableMenus(ctx context.Context, menus []*entity.SysMenu) ([]*entity.SysMenu, error) {
	if len(menus) == 0 {
		return []*entity.SysMenu{}, nil
	}
	filteredMenus := s.permissionFilter.FilterPermissionMenus(ctx, menus)
	classifier, err := s.newRoleMenuAssignmentClassifier(ctx, filteredMenus)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.SysMenu, 0, len(filteredMenus))
	for _, menu := range filteredMenus {
		if classifier.assignable(menu) {
			result = append(result, menu)
		}
	}
	return result, nil
}

// loadRoleAssignmentMenus reads the full global menu topology for role
// assignment classification.
func loadRoleAssignmentMenus(ctx context.Context) ([]*entity.SysMenu, error) {
	cols := dao.SysMenu.Columns()
	var menus []*entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).
		OrderAsc(cols.ParentId).
		OrderAsc(cols.Sort).
		OrderAsc(cols.Id).
		Scan(&menus)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// newRoleMenuAssignmentClassifier prepares ancestor and plugin-scope lookup
// state for one role assignment filtering pass.
func (s *serviceImpl) newRoleMenuAssignmentClassifier(
	ctx context.Context,
	menus []*entity.SysMenu,
) (*roleMenuAssignmentClassifier, error) {
	pluginScopes, err := loadRoleMenuPluginScopes(ctx, menus)
	if err != nil {
		return nil, err
	}
	menusByID := make(map[int]*entity.SysMenu, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		menusByID[menu.Id] = menu
	}
	return &roleMenuAssignmentClassifier{
		platformContext: s.roleMenuAssignmentPlatformContext(ctx),
		menusByID:       menusByID,
		pluginScopes:    pluginScopes,
	}, nil
}

// roleMenuAssignmentPlatformContext reports whether current context may assign
// platform-control-plane permissions. Disabled tenancy is treated as platform
// context because no tenant-scoped boundary exists.
func (s *serviceImpl) roleMenuAssignmentPlatformContext(ctx context.Context) bool {
	if s == nil || s.tenantSvc == nil {
		return true
	}
	if !s.tenantSvc.Available(ctx) {
		return true
	}
	return s.tenantSvc.PlatformBypass(ctx)
}

// loadRoleMenuPluginScopes reads plugin scope metadata for plugin-owned menu
// rows participating in the current classification pass.
func loadRoleMenuPluginScopes(ctx context.Context, menus []*entity.SysMenu) (map[string]string, error) {
	pluginIDs := collectRoleMenuPluginIDs(menus)
	if len(pluginIDs) == 0 {
		return map[string]string{}, nil
	}
	cols := dao.SysPlugin.Columns()
	var registries []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Fields(cols.PluginId, cols.ScopeNature).
		WhereIn(cols.PluginId, pluginIDs).
		Scan(&registries)
	if err != nil {
		return nil, err
	}
	scopes := make(map[string]string, len(registries))
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if pluginID == "" {
			continue
		}
		scopes[pluginID] = strings.TrimSpace(strings.ToLower(registry.ScopeNature))
	}
	return scopes, nil
}

// collectRoleMenuPluginIDs extracts distinct plugin IDs from menu keys.
func collectRoleMenuPluginIDs(menus []*entity.SysMenu) []string {
	seen := make(map[string]struct{}, len(menus))
	ids := make([]string, 0)
	for _, menu := range menus {
		pluginID := roleMenuPluginID(menu)
		if pluginID == "" {
			continue
		}
		if _, ok := seen[pluginID]; ok {
			continue
		}
		seen[pluginID] = struct{}{}
		ids = append(ids, pluginID)
	}
	return ids
}

// roleMenuAssignmentClassifier evaluates whether one menu is assignable for
// the current platform or tenant context.
type roleMenuAssignmentClassifier struct {
	platformContext bool
	menusByID       map[int]*entity.SysMenu
	pluginScopes    map[string]string
}

// assignable reports whether one menu row may be granted by the current
// request context.
func (c *roleMenuAssignmentClassifier) assignable(menu *entity.SysMenu) bool {
	if menu == nil {
		return false
	}
	if c.platformContext {
		return true
	}
	if roleMenuIsTenantAllowedSelfService(menu) {
		return true
	}
	return !c.platformOnly(menu)
}

// platformOnly reports whether one menu is part of the platform control plane
// for tenant role-assignment purposes.
func (c *roleMenuAssignmentClassifier) platformOnly(menu *entity.SysMenu) bool {
	if menu == nil {
		return false
	}
	menuKey := strings.TrimSpace(menu.MenuKey)
	if _, ok := rolePlatformOnlyMenuKeys[menuKey]; ok {
		return true
	}
	if c.hasPlatformOnlyAncestor(menu) {
		return true
	}
	if roleMenuPermissionPlatformOnly(menu.Perms) {
		return true
	}
	pluginID := roleMenuPluginID(menu)
	if pluginID != "" && c.pluginScopes[pluginID] == rolePluginScopePlatformOnly {
		return true
	}
	return false
}

// hasPlatformOnlyAncestor walks the in-memory menu parent chain to catch
// platform-control-plane descendants even when the descendant's own permission
// string is blank or newly introduced.
func (c *roleMenuAssignmentClassifier) hasPlatformOnlyAncestor(menu *entity.SysMenu) bool {
	visited := make(map[int]struct{})
	parentID := menu.ParentId
	for parentID > 0 {
		if _, ok := visited[parentID]; ok {
			return false
		}
		visited[parentID] = struct{}{}
		parent := c.menusByID[parentID]
		if parent == nil {
			return false
		}
		if roleMenuIsTenantAllowedSelfService(parent) {
			return false
		}
		if _, ok := rolePlatformOnlyMenuKeys[strings.TrimSpace(parent.MenuKey)]; ok {
			return true
		}
		if roleMenuPermissionPlatformOnly(parent.Perms) {
			return true
		}
		parentID = parent.ParentId
	}
	return false
}

// roleMenuPermissionPlatformOnly classifies stable permission-string namespaces
// that belong to platform control-plane governance.
func roleMenuPermissionPlatformOnly(permission string) bool {
	trimmed := strings.TrimSpace(permission)
	if trimmed == "" {
		return false
	}
	if rolePermissionHasAnyPrefix(trimmed, roleTenantAllowedPermissionPrefixes) {
		return false
	}
	if _, ok := rolePlatformOnlyPermissions[trimmed]; ok {
		return true
	}
	return rolePermissionHasAnyPrefix(trimmed, roleTenantBlockedPermissionPrefixes)
}

// roleMenuIsTenantAllowedSelfService preserves tenant plugin self-service
// permissions even though the owning linapro-tenant-core plugin also exposes platform
// tenant-management permissions.
func roleMenuIsTenantAllowedSelfService(menu *entity.SysMenu) bool {
	if menu == nil {
		return false
	}
	return rolePermissionHasAnyPrefix(strings.TrimSpace(menu.Perms), roleTenantAllowedPermissionPrefixes)
}

// rolePermissionHasAnyPrefix checks one normalized permission string against a
// stable list of namespace prefixes.
func rolePermissionHasAnyPrefix(permission string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(permission, prefix) {
			return true
		}
	}
	return false
}

// roleMenuPluginID extracts the owner plugin ID from the menu key.
func roleMenuPluginID(menu *entity.SysMenu) string {
	if menu == nil {
		return ""
	}
	return parseRoleMenuPluginID(menu.MenuKey)
}

// parseRoleMenuPluginID extracts the plugin ID from "plugin:<id>:..." or
// "plugin:<id> ..." markers.
func parseRoleMenuPluginID(value string) string {
	tagged := strings.TrimSpace(value)
	if !strings.HasPrefix(tagged, rolePluginMenuKeyPrefix) {
		return ""
	}
	suffix := tagged[len(rolePluginMenuKeyPrefix):]
	end := len(suffix)
	for _, separator := range []string{":", " "} {
		if idx := strings.Index(suffix, separator); idx >= 0 && idx < end {
			end = idx
		}
	}
	return strings.TrimSpace(suffix[:end])
}

// normalizeRoleMenuIDs returns distinct positive menu IDs while preserving
// caller order, matching sys_role_menu persistence normalization.
func normalizeRoleMenuIDs(menuIDs []int) []int {
	seen := make(map[int]struct{}, len(menuIDs))
	normalized := make([]int, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		normalized = append(normalized, menuID)
	}
	return normalized
}
