// This file synchronizes manifest-declared plugin menus and dynamic route
// permission entries into sys_menu and ensures admin role access is granted.

package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

const (
	pluginMenuDefaultVisible = 1
	pluginMenuDefaultStatus  = 1
	pluginMenuDefaultIsFrame = 0
	pluginMenuDefaultIsCache = 0
	pluginDefaultAdminRoleID = 1
)

// SyncPluginMenusAndPermissions reconciles all manifest menus and dynamic route permission
// entries into sys_menu, then ensures the admin role has access to them.
// It implements runtime.MenuManager and catalog.MenuSyncer.
func (s *serviceImpl) SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	return dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.syncPluginMenusInTx(ctx, manifest); err != nil {
			return err
		}
		return s.syncDynamicRoutePermissionMenus(ctx, manifest)
	})
}

// SyncPluginMenus reconciles only the manifest-declared menus, skipping route-permission entries.
// Used during reconciler rollback to restore the previous menu state without touching permissions.
// It implements runtime.MenuManager.
func (s *serviceImpl) SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	return dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		return s.syncPluginMenusInTx(ctx, manifest)
	})
}

// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
// It implements runtime.MenuManager.
func (s *serviceImpl) DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
	if err != nil {
		return err
	}
	menuKeys := make([]string, 0, len(existingMenus))
	for _, item := range existingMenus {
		if item == nil || strings.TrimSpace(item.MenuKey) == "" {
			continue
		}
		menuKeys = append(menuKeys, strings.TrimSpace(item.MenuKey))
	}
	return s.deletePluginMenusByKeys(ctx, menuKeys)
}

// syncPluginMenusInTx reconciles one plugin's declared menus inside the caller's transaction.
func (s *serviceImpl) syncPluginMenusInTx(ctx context.Context, manifest *catalog.Manifest) error {
	declaredKeys := s.listDeclaredPluginMenuKeys(manifest)
	existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
	if err != nil {
		return err
	}

	existingByKey := make(map[string]*entity.SysMenu, len(existingMenus))
	staleKeys := make([]string, 0)
	for _, item := range existingMenus {
		if item == nil {
			continue
		}
		existingByKey[item.MenuKey] = item
		if _, ok := declaredKeys[item.MenuKey]; !ok {
			// Only remove declared menu keys, not permission menu synthetic keys.
			if !isDynamicRoutePermissionMenuKey(item.MenuKey) {
				staleKeys = append(staleKeys, item.MenuKey)
			}
		}
	}

	externalParents, err := s.listPluginMenuExternalParents(ctx, manifest)
	if err != nil {
		return err
	}

	resolvedIDs := make(map[string]int, len(manifest.Menus))
	pendingMenus := append([]*catalog.MenuSpec(nil), manifest.Menus...)
	for len(pendingMenus) > 0 {
		nextPending := make([]*catalog.MenuSpec, 0, len(pendingMenus))
		progressed := false

		for _, spec := range pendingMenus {
			if spec == nil {
				continue
			}

			parentID, resolved, err := s.resolvePluginMenuParentID(spec, declaredKeys, resolvedIDs, externalParents)
			if err != nil {
				return err
			}
			if !resolved {
				nextPending = append(nextPending, spec)
				continue
			}

			menuID, err := s.upsertPluginMenu(ctx, spec, parentID, existingByKey[spec.Key])
			if err != nil {
				return err
			}
			resolvedIDs[spec.Key] = menuID
			progressed = true
		}

		if !progressed {
			unresolved := make([]string, 0, len(nextPending))
			for _, spec := range nextPending {
				if spec == nil {
					continue
				}
				unresolved = append(unresolved, spec.Key)
			}
			sort.Strings(unresolved)
			return gerror.Newf("插件菜单 parent_key 无法解析: %s", strings.Join(unresolved, ", "))
		}

		pendingMenus = nextPending
	}

	if err := s.ensurePluginMenuAdminBindings(ctx, resolvedIDs); err != nil {
		return err
	}
	return s.deletePluginMenusByKeys(ctx, staleKeys)
}

// syncDynamicRoutePermissionMenus materializes route permission entries as hidden button menus.
func (s *serviceImpl) syncDynamicRoutePermissionMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	permissionMenus := s.buildDynamicRoutePermissionMenuSpecs(manifest)
	existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
	if err != nil {
		return err
	}

	resolvedIDs := make(map[string]int, len(permissionMenus))
	desiredKeys := make(map[string]struct{}, len(permissionMenus))
	staleKeys := make([]string, 0)
	existingByKey := make(map[string]*entity.SysMenu, len(existingMenus))
	for _, menu := range existingMenus {
		if menu == nil {
			continue
		}
		existingByKey[menu.MenuKey] = menu
	}
	for _, spec := range permissionMenus {
		desiredKeys[spec.Key] = struct{}{}
		menuID, err := s.upsertPluginMenu(ctx, spec, 0, existingByKey[spec.Key])
		if err != nil {
			return err
		}
		resolvedIDs[spec.Key] = menuID
	}
	for _, menu := range existingMenus {
		if menu == nil || !isDynamicRoutePermissionMenuKey(menu.MenuKey) {
			continue
		}
		if _, ok := desiredKeys[strings.TrimSpace(menu.MenuKey)]; ok {
			continue
		}
		staleKeys = append(staleKeys, strings.TrimSpace(menu.MenuKey))
	}
	if err = s.ensurePluginMenuAdminBindings(ctx, resolvedIDs); err != nil {
		return err
	}
	return s.deletePluginMenusByKeys(ctx, staleKeys)
}

// buildDynamicRoutePermissionMenuSpecs derives synthetic hidden button menus from manifest routes.
func (s *serviceImpl) buildDynamicRoutePermissionMenuSpecs(manifest *catalog.Manifest) []*catalog.MenuSpec {
	if manifest == nil || len(manifest.Routes) == 0 {
		return []*catalog.MenuSpec{}
	}

	items := make([]*catalog.MenuSpec, 0)
	seen := make(map[string]struct{})
	for _, route := range manifest.Routes {
		if route == nil || strings.TrimSpace(route.Permission) == "" {
			continue
		}
		permission := strings.TrimSpace(route.Permission)
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		visible := 0
		status := pluginMenuDefaultStatus
		items = append(items, &catalog.MenuSpec{
			Key:     buildDynamicRoutePermissionMenuKey(manifest.ID, permission),
			Name:    catalog.DynamicRoutePermissionMenuNamePrefix + permission,
			Perms:   permission,
			Type:    catalog.MenuTypeButton.String(),
			Visible: &visible,
			Status:  &status,
			Remark:  buildDynamicRoutePermissionMenuRemark(manifest.ID),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Perms < items[j].Perms
	})
	return items
}

// buildDynamicRoutePermissionMenuKey returns the synthetic menu key for a route permission.
func buildDynamicRoutePermissionMenuKey(pluginID string, permission string) string {
	encodedPermission := base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(permission)))
	return catalog.MenuKeyPrefix + strings.TrimSpace(pluginID) + catalog.DynamicRoutePermissionMenuKeySeparator + encodedPermission
}

func isDynamicRoutePermissionMenuKey(menuKey string) bool {
	return strings.Contains(strings.TrimSpace(menuKey), catalog.DynamicRoutePermissionMenuKeySeparator)
}

func buildDynamicRoutePermissionMenuRemark(pluginID string) string {
	return catalog.MenuRemarkPrefix + strings.TrimSpace(pluginID) + catalog.DynamicRoutePermissionMenuRemarkSuffix
}

func (s *serviceImpl) listDeclaredPluginMenuKeys(manifest *catalog.Manifest) map[string]struct{} {
	declaredKeys := make(map[string]struct{}, len(manifest.Menus))
	if manifest == nil {
		return declaredKeys
	}
	for _, spec := range manifest.Menus {
		if spec == nil || strings.TrimSpace(spec.Key) == "" {
			continue
		}
		declaredKeys[strings.TrimSpace(spec.Key)] = struct{}{}
	}
	return declaredKeys
}

func (s *serviceImpl) listPluginMenuExternalParents(ctx context.Context, manifest *catalog.Manifest) (map[string]*entity.SysMenu, error) {
	declaredKeys := s.listDeclaredPluginMenuKeys(manifest)
	parentKeys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, spec := range manifest.Menus {
		if spec == nil || spec.ParentKey == "" {
			continue
		}
		if _, ok := declaredKeys[spec.ParentKey]; ok {
			continue
		}
		if _, ok := seen[spec.ParentKey]; ok {
			continue
		}
		seen[spec.ParentKey] = struct{}{}
		parentKeys = append(parentKeys, spec.ParentKey)
	}
	return s.listMenusByKeys(ctx, parentKeys, false)
}

func (s *serviceImpl) resolvePluginMenuParentID(
	spec *catalog.MenuSpec,
	declaredKeys map[string]struct{},
	resolvedIDs map[string]int,
	externalParents map[string]*entity.SysMenu,
) (int, bool, error) {
	if spec == nil || strings.TrimSpace(spec.ParentKey) == "" {
		return 0, true, nil
	}

	parentKey := strings.TrimSpace(spec.ParentKey)
	if _, ok := declaredKeys[parentKey]; ok {
		parentID, resolved := resolvedIDs[parentKey]
		return parentID, resolved, nil
	}

	parent, ok := externalParents[parentKey]
	if !ok || parent == nil {
		return 0, false, gerror.Newf("插件菜单 parent_key 不存在: %s -> %s", spec.Key, spec.ParentKey)
	}
	return parent.Id, true, nil
}

func (s *serviceImpl) upsertPluginMenu(
	ctx context.Context,
	spec *catalog.MenuSpec,
	parentID int,
	existing *entity.SysMenu,
) (int, error) {
	if spec == nil {
		return 0, gerror.New("插件菜单声明不能为空")
	}

	queryParam, err := buildMenuQueryParam(spec)
	if err != nil {
		return 0, err
	}
	visible, err := normalizeMenuFlag(spec.Visible, pluginMenuDefaultVisible)
	if err != nil {
		return 0, err
	}
	status, err := normalizeMenuFlag(spec.Status, pluginMenuDefaultStatus)
	if err != nil {
		return 0, err
	}
	isFrame, err := normalizeMenuFlag(spec.IsFrame, pluginMenuDefaultIsFrame)
	if err != nil {
		return 0, err
	}
	isCache, err := normalizeMenuFlag(spec.IsCache, pluginMenuDefaultIsCache)
	if err != nil {
		return 0, err
	}

	if existing != nil && existing.DeletedAt != nil {
		if _, err = dao.SysMenu.Ctx(ctx).
			Unscoped().
			Where(do.SysMenu{Id: existing.Id}).
			Delete(); err != nil {
			return 0, err
		}
		existing = nil
	}

	data := do.SysMenu{
		ParentId:   parentID,
		MenuKey:    spec.Key,
		Name:       spec.Name,
		Path:       spec.Path,
		Component:  spec.Component,
		Perms:      spec.Perms,
		Icon:       spec.Icon,
		Type:       catalog.NormalizeMenuType(spec.Type).String(),
		Sort:       spec.Sort,
		Visible:    visible,
		Status:     status,
		IsFrame:    isFrame,
		IsCache:    isCache,
		QueryParam: queryParam,
		Remark:     spec.Remark,
	}

	if existing == nil {
		menuID, err := dao.SysMenu.Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}
		return int(menuID), nil
	}

	if _, err = dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{Id: existing.Id}).
		Data(data).
		Update(); err != nil {
		return 0, err
	}
	return existing.Id, nil
}

func (s *serviceImpl) ensurePluginMenuAdminBindings(ctx context.Context, resolvedIDs map[string]int) error {
	menuIDs := make([]int, 0, len(resolvedIDs))
	for _, menuID := range resolvedIDs {
		if menuID <= 0 {
			continue
		}
		menuIDs = append(menuIDs, menuID)
	}
	sort.Ints(menuIDs)

	for _, menuID := range menuIDs {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			Data(do.SysRoleMenu{
				RoleId: pluginDefaultAdminRoleID,
				MenuId: menuID,
			}).
			Save(); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceImpl) listPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	pattern := fmt.Sprintf("%s%s:%%", catalog.MenuKeyPrefix, strings.TrimSpace(pluginID))
	cols := dao.SysMenu.Columns()
	items := make([]*entity.SysMenu, 0)
	err := dao.SysMenu.Ctx(ctx).
		Unscoped().
		WhereLike(cols.MenuKey, pattern).
		OrderAsc(cols.Id).
		Scan(&items)
	return items, err
}

func (s *serviceImpl) listMenusByKeys(ctx context.Context, menuKeys []string, unscoped bool) (map[string]*entity.SysMenu, error) {
	result := make(map[string]*entity.SysMenu, len(menuKeys))
	if len(menuKeys) == 0 {
		return result, nil
	}

	m := dao.SysMenu.Ctx(ctx)
	if unscoped {
		m = m.Unscoped()
	}

	cols := dao.SysMenu.Columns()
	items := make([]*entity.SysMenu, 0)
	if err := m.WhereIn(cols.MenuKey, menuKeys).OrderAsc(cols.Id).Scan(&items); err != nil {
		return nil, err
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		result[item.MenuKey] = item
	}
	return result, nil
}

func (s *serviceImpl) deletePluginMenusByKeys(ctx context.Context, menuKeys []string) error {
	if len(menuKeys) == 0 {
		return nil
	}

	menuMap, err := s.listMenusByKeys(ctx, menuKeys, true)
	if err != nil {
		return err
	}

	menuIDs := make([]int, 0, len(menuMap))
	for _, item := range menuMap {
		if item == nil {
			continue
		}
		menuIDs = append(menuIDs, item.Id)
	}
	sort.Ints(menuIDs)

	if len(menuIDs) > 0 {
		menuIDValues := make([]interface{}, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			menuIDValues = append(menuIDValues, menuID)
		}
		if _, err = dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDValues).
			Delete(); err != nil {
			return err
		}
	}

	if _, err = dao.SysMenu.Ctx(ctx).
		Unscoped().
		WhereIn(dao.SysMenu.Columns().MenuKey, menuKeys).
		Delete(); err != nil {
		return err
	}
	return nil
}

// normalizeMenuFlag validates and returns a plugin menu integer flag (0 or 1).
func normalizeMenuFlag(value *int, defaultValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	if *value != 0 && *value != 1 {
		return 0, gerror.New("仅支持 0 或 1")
	}
	return *value, nil
}

// BuildDynamicRoutePermissionMenuKey is the exported form of buildDynamicRoutePermissionMenuKey for cross-package access.
func BuildDynamicRoutePermissionMenuKey(pluginID string, permission string) string {
	return buildDynamicRoutePermissionMenuKey(pluginID, permission)
}

// ListPluginMenusByPlugin is the exported form of listPluginMenusByPlugin for cross-package access.
func (s *serviceImpl) ListPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	return s.listPluginMenusByPlugin(ctx, pluginID)
}

// buildMenuQueryParam serializes the query map or query_param field of a menu spec.
func buildMenuQueryParam(spec *catalog.MenuSpec) (string, error) {
	if spec == nil {
		return "", nil
	}
	if strings.TrimSpace(spec.QueryParam) != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(spec.QueryParam), &payload); err != nil {
			return "", err
		}
		if len(payload) == 0 {
			return "", nil
		}
		content, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	if len(spec.Query) == 0 {
		return "", nil
	}
	content, err := json.Marshal(spec.Query)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
