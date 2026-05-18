// This file synchronizes manifest-declared plugin menus and dynamic route
// permission entries into sys_menu.

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
	"github.com/gogf/gf/v2/util/gconv"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/startupstats"
)

// Plugin menu defaults and synthetic permission-menu settings used during menu sync.
const (
	pluginMenuDefaultVisible = 1
	pluginMenuDefaultStatus  = 1
	pluginMenuDefaultIsFrame = 0
	pluginMenuDefaultIsCache = 0
)

// SyncPluginMenusAndPermissions reconciles all manifest menus and dynamic route permission
// entries into sys_menu.
// It implements runtime.MenuManager and catalog.MenuSyncer.
func (s *serviceImpl) SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}
	changed, err := s.pluginMenusAndPermissionsNeedSync(ctx, manifest)
	if err != nil {
		return err
	}
	if !changed {
		startupstats.Add(ctx, startupstats.CounterPluginMenuSyncNoop, 1)
		return nil
	}
	startupstats.Add(ctx, startupstats.CounterPluginMenuSyncChanged, 1)
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
	changed, err := s.pluginDeclaredMenusNeedSync(ctx, manifest)
	if err != nil {
		return err
	}
	if !changed {
		startupstats.Add(ctx, startupstats.CounterPluginMenuSyncNoop, 1)
		return nil
	}
	startupstats.Add(ctx, startupstats.CounterPluginMenuSyncChanged, 1)
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

// pluginMenusAndPermissionsNeedSync checks whether the full menu projection
// differs before opening a transaction.
func (s *serviceImpl) pluginMenusAndPermissionsNeedSync(ctx context.Context, manifest *catalog.Manifest) (bool, error) {
	changed, existingByKey, err := s.pluginDeclaredMenusNeedSyncWithExisting(ctx, manifest)
	if err != nil || changed {
		return changed, err
	}
	return s.dynamicRoutePermissionMenusNeedSync(manifest, existingByKey)
}

// pluginDeclaredMenusNeedSync checks whether manifest-declared menus differ.
func (s *serviceImpl) pluginDeclaredMenusNeedSync(ctx context.Context, manifest *catalog.Manifest) (bool, error) {
	changed, _, err := s.pluginDeclaredMenusNeedSyncWithExisting(ctx, manifest)
	return changed, err
}

// pluginDeclaredMenusNeedSyncWithExisting returns whether declared plugin menus
// differ and the existing plugin-owned menu lookup for permission checks.
func (s *serviceImpl) pluginDeclaredMenusNeedSyncWithExisting(
	ctx context.Context,
	manifest *catalog.Manifest,
) (bool, map[string]*entity.SysMenu, error) {
	declaredKeys := s.listDeclaredPluginMenuKeys(manifest)
	existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
	if err != nil {
		return false, nil, err
	}

	existingByKey := make(map[string]*entity.SysMenu, len(existingMenus))
	for _, item := range existingMenus {
		if item == nil {
			continue
		}
		existingByKey[item.MenuKey] = item
		if _, ok := declaredKeys[item.MenuKey]; !ok && !isDynamicRoutePermissionMenuKey(item.MenuKey) {
			return true, existingByKey, nil
		}
	}

	externalParents, err := s.listPluginMenuExternalParents(ctx, manifest)
	if err != nil {
		return false, nil, err
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
				return false, nil, err
			}
			if !resolved {
				nextPending = append(nextPending, spec)
				continue
			}

			data, err := buildPluginMenuData(spec, parentID)
			if err != nil {
				return false, nil, err
			}
			existing := existingByKey[spec.Key]
			if existing == nil || existing.DeletedAt != nil || !pluginMenuMatches(existing, data) {
				return true, existingByKey, nil
			}
			resolvedIDs[spec.Key] = existing.Id
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
			return false, nil, gerror.Newf("plugin menu parent_key cannot be resolved: %s", strings.Join(unresolved, ", "))
		}

		pendingMenus = nextPending
	}
	return false, existingByKey, nil
}

// dynamicRoutePermissionMenusNeedSync checks whether synthetic route permission
// menus differ from the current plugin-owned menu projection.
func (s *serviceImpl) dynamicRoutePermissionMenusNeedSync(
	manifest *catalog.Manifest,
	existingByKey map[string]*entity.SysMenu,
) (bool, error) {
	permissionMenus := s.buildDynamicRoutePermissionMenuSpecs(manifest)
	desiredKeys := make(map[string]struct{}, len(permissionMenus))
	for _, spec := range permissionMenus {
		if spec == nil {
			continue
		}
		desiredKeys[spec.Key] = struct{}{}
		parentID, err := s.resolveDynamicRoutePermissionParentID(spec, existingByKey)
		if err != nil {
			return false, err
		}
		data, err := buildPluginMenuData(spec, parentID)
		if err != nil {
			return false, err
		}
		existing := existingByKey[spec.Key]
		if existing == nil || existing.DeletedAt != nil || !pluginMenuMatches(existing, data) {
			return true, nil
		}
	}
	for _, menu := range existingByKey {
		if menu == nil || !isDynamicRoutePermissionMenuKey(menu.MenuKey) {
			continue
		}
		if _, ok := desiredKeys[strings.TrimSpace(menu.MenuKey)]; !ok {
			return true, nil
		}
	}
	return false, nil
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
			return gerror.Newf("plugin menu parent_key cannot be resolved: %s", strings.Join(unresolved, ", "))
		}

		pendingMenus = nextPending
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
		parentID, err := s.resolveDynamicRoutePermissionParentID(spec, existingByKey)
		if err != nil {
			return err
		}
		menuID, err := s.upsertPluginMenu(ctx, spec, parentID, existingByKey[spec.Key])
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
	return s.deletePluginMenusByKeys(ctx, staleKeys)
}

// buildDynamicRoutePermissionMenuSpecs derives synthetic hidden button menus from manifest routes.
func (s *serviceImpl) buildDynamicRoutePermissionMenuSpecs(manifest *catalog.Manifest) []*catalog.MenuSpec {
	if manifest == nil || len(manifest.Routes) == 0 {
		return []*catalog.MenuSpec{}
	}

	items := make([]*catalog.MenuSpec, 0)
	seen := make(map[string]struct{})
	parentKey := dynamicRoutePermissionParentKey(manifest)
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
			Key:       buildDynamicRoutePermissionMenuKey(manifest.ID, permission),
			ParentKey: parentKey,
			Name:      catalog.DynamicRoutePermissionMenuNamePrefix + permission,
			Perms:     permission,
			Type:      catalog.MenuTypeButton.String(),
			Visible:   &visible,
			Status:    &status,
			Remark:    buildDynamicRoutePermissionMenuRemark(manifest.ID),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Perms < items[j].Perms
	})
	return items
}

// dynamicRoutePermissionParentKey returns the plugin menu that should own
// synthetic route-permission buttons. The first non-button root menu is the
// plugin entry point in current manifests; if none exists, route permissions
// keep the legacy root placement.
func dynamicRoutePermissionParentKey(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	for _, spec := range manifest.Menus {
		if spec == nil || strings.TrimSpace(spec.Key) == "" {
			continue
		}
		if catalog.NormalizeMenuType(spec.Type) == catalog.MenuTypeButton {
			continue
		}
		if strings.TrimSpace(spec.ParentKey) == "" {
			return strings.TrimSpace(spec.Key)
		}
	}
	for _, spec := range manifest.Menus {
		if spec == nil || strings.TrimSpace(spec.Key) == "" {
			continue
		}
		if catalog.NormalizeMenuType(spec.Type) != catalog.MenuTypeButton {
			return strings.TrimSpace(spec.Key)
		}
	}
	return ""
}

// resolveDynamicRoutePermissionParentID maps the synthetic permission menu's
// parent_key to the persisted plugin menu row created earlier in the same sync.
func (s *serviceImpl) resolveDynamicRoutePermissionParentID(
	spec *catalog.MenuSpec,
	existingByKey map[string]*entity.SysMenu,
) (int, error) {
	if spec == nil || strings.TrimSpace(spec.ParentKey) == "" {
		return 0, nil
	}
	parent, ok := existingByKey[strings.TrimSpace(spec.ParentKey)]
	if !ok || parent == nil || parent.DeletedAt != nil {
		return 0, gerror.Newf("dynamic route permission parent menu does not exist: %s -> %s", spec.Key, spec.ParentKey)
	}
	return parent.Id, nil
}

// buildDynamicRoutePermissionMenuKey returns the synthetic menu key for a route permission.
func buildDynamicRoutePermissionMenuKey(pluginID string, permission string) string {
	encodedPermission := base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(permission)))
	return catalog.MenuKeyPrefix + strings.TrimSpace(pluginID) + catalog.DynamicRoutePermissionMenuKeySeparator + encodedPermission
}

// isDynamicRoutePermissionMenuKey reports whether a menu key belongs to one
// synthetic dynamic-route permission menu.
func isDynamicRoutePermissionMenuKey(menuKey string) bool {
	return strings.Contains(strings.TrimSpace(menuKey), catalog.DynamicRoutePermissionMenuKeySeparator)
}

// buildDynamicRoutePermissionMenuRemark returns the stable remark used for
// synthetic permission menus generated from dynamic routes.
func buildDynamicRoutePermissionMenuRemark(pluginID string) string {
	return catalog.MenuRemarkPrefix + strings.TrimSpace(pluginID) + catalog.DynamicRoutePermissionMenuRemarkSuffix
}

// listDeclaredPluginMenuKeys returns the normalized set of menu keys declared
// directly by the manifest.
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

// listPluginMenuExternalParents resolves manifest parent_key values that point
// at menus owned outside the current plugin.
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
		if !menusvc.IsStableCatalogKey(spec.ParentKey) {
			return nil, gerror.Newf("plugin menu parent_key can only mount to a stable host catalog: %s -> %s", spec.Key, spec.ParentKey)
		}
		if allowed, official := menusvc.IsExpectedStableParentKey(manifest.ID, spec.ParentKey); official && !allowed {
			expectedParentKeys, _ := menusvc.ExpectedStableParentKeys(manifest.ID)
			return nil, gerror.Newf("official plugin top-level menu parent_key is invalid: %s -> %s, expected %s", spec.Key, spec.ParentKey, strings.Join(expectedParentKeys, " or "))
		}
		if _, ok := seen[spec.ParentKey]; ok {
			continue
		}
		seen[spec.ParentKey] = struct{}{}
		parentKeys = append(parentKeys, spec.ParentKey)
	}
	return s.listMenusByKeys(ctx, parentKeys, false)
}

// resolvePluginMenuParentID resolves the parent menu ID for one manifest menu
// while supporting both intra-plugin and external parent references.
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
		return 0, false, gerror.Newf("plugin menu parent_key does not exist: %s -> %s", spec.Key, spec.ParentKey)
	}
	return parent.Id, true, nil
}

// upsertPluginMenu inserts or updates one plugin-owned sys_menu row from a
// normalized manifest menu specification.
func (s *serviceImpl) upsertPluginMenu(
	ctx context.Context,
	spec *catalog.MenuSpec,
	parentID int,
	existing *entity.SysMenu,
) (int, error) {
	if spec == nil {
		return 0, gerror.New("plugin menu declaration cannot be nil")
	}

	data, err := buildPluginMenuData(spec, parentID)
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
		if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
			snapshot.deleteMenus([]string{existing.MenuKey})
		}
		existing = nil
	}

	if existing == nil {
		menuID, err := dao.SysMenu.Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}
		if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
			snapshot.storeMenu(buildPluginMenuEntity(int(menuID), data))
		}
		return int(menuID), nil
	}

	if pluginMenuMatches(existing, data) {
		return existing.Id, nil
	}

	if _, err = dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{Id: existing.Id}).
		Data(data).
		Update(); err != nil {
		return 0, err
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.storeMenu(buildPluginMenuEntity(existing.Id, data))
	}
	return existing.Id, nil
}

// buildPluginMenuData converts one normalized manifest menu specification into
// the persisted menu projection.
func buildPluginMenuData(spec *catalog.MenuSpec, parentID int) (do.SysMenu, error) {
	if spec == nil {
		return do.SysMenu{}, gerror.New("plugin menu declaration cannot be nil")
	}

	queryParam, err := buildMenuQueryParam(spec)
	if err != nil {
		return do.SysMenu{}, err
	}
	visible, err := normalizeMenuFlag(spec.Visible, pluginMenuDefaultVisible)
	if err != nil {
		return do.SysMenu{}, err
	}
	status, err := normalizeMenuFlag(spec.Status, pluginMenuDefaultStatus)
	if err != nil {
		return do.SysMenu{}, err
	}
	isFrame, err := normalizeMenuFlag(spec.IsFrame, pluginMenuDefaultIsFrame)
	if err != nil {
		return do.SysMenu{}, err
	}
	isCache, err := normalizeMenuFlag(spec.IsCache, pluginMenuDefaultIsCache)
	if err != nil {
		return do.SysMenu{}, err
	}

	return do.SysMenu{
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
	}, nil
}

// buildPluginMenuEntity creates the startup snapshot projection for one menu row
// after an insert or update.
func buildPluginMenuEntity(menuID int, data do.SysMenu) *entity.SysMenu {
	return &entity.SysMenu{
		Id:         menuID,
		ParentId:   dataInt(data.ParentId),
		MenuKey:    dataString(data.MenuKey),
		Name:       dataString(data.Name),
		Path:       dataString(data.Path),
		Component:  dataString(data.Component),
		Perms:      dataString(data.Perms),
		Icon:       dataString(data.Icon),
		Type:       dataString(data.Type),
		Sort:       dataInt(data.Sort),
		Visible:    dataInt(data.Visible),
		Status:     dataInt(data.Status),
		IsFrame:    dataInt(data.IsFrame),
		IsCache:    dataInt(data.IsCache),
		QueryParam: dataString(data.QueryParam),
		Remark:     dataString(data.Remark),
	}
}

// pluginMenuMatches reports whether a persisted plugin menu already matches
// the desired manifest projection.
func pluginMenuMatches(existing *entity.SysMenu, data do.SysMenu) bool {
	if existing == nil {
		return false
	}
	return existing.ParentId == dataInt(data.ParentId) &&
		existing.MenuKey == dataString(data.MenuKey) &&
		existing.Name == dataString(data.Name) &&
		existing.Path == dataString(data.Path) &&
		existing.Component == dataString(data.Component) &&
		existing.Perms == dataString(data.Perms) &&
		existing.Icon == dataString(data.Icon) &&
		existing.Type == dataString(data.Type) &&
		existing.Sort == dataInt(data.Sort) &&
		existing.Visible == dataInt(data.Visible) &&
		existing.Status == dataInt(data.Status) &&
		existing.IsFrame == dataInt(data.IsFrame) &&
		existing.IsCache == dataInt(data.IsCache) &&
		existing.QueryParam == dataString(data.QueryParam) &&
		existing.Remark == dataString(data.Remark)
}

// dataString normalizes a DO field into its persisted string value.
func dataString(value any) string {
	return gconv.String(value)
}

// dataInt normalizes a DO field into its persisted integer value.
func dataInt(value any) int {
	return gconv.Int(value)
}

// listPluginMenusByPlugin loads all menus owned by the given plugin, including
// soft-deleted rows when needed by cleanup flows.
func (s *serviceImpl) listPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.pluginMenus(pluginID), nil
	}

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

// listMenusByKeys resolves menus by key and returns them as a lookup map.
func (s *serviceImpl) listMenusByKeys(ctx context.Context, menuKeys []string, unscoped bool) (map[string]*entity.SysMenu, error) {
	result := make(map[string]*entity.SysMenu, len(menuKeys))
	if len(menuKeys) == 0 {
		return result, nil
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.menusByKeys(menuKeys, unscoped), nil
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

// deletePluginMenusByKeys deletes plugin-owned menus and their role bindings in
// a deterministic order derived from the provided keys.
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
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.deleteMenus(menuKeys)
	}
	return nil
}

// normalizeMenuFlag validates and returns a plugin menu integer flag (0 or 1).
func normalizeMenuFlag(value *int, defaultValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	if *value != 0 && *value != 1 {
		return 0, gerror.New("only 0 or 1 is supported")
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
