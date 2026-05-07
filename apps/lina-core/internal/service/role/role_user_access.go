// This file assembles one user's effective role, menu, and permission snapshot
// for permission checks and login bootstrap responses.

package role

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/user/accountpolicy"
	"lina-core/pkg/bizerr"
)

// UserAccessContext describes the role, menu, and permission data required by the current user session.
type UserAccessContext struct {
	RoleIds              []int           // RoleIds contains all role IDs bound to the user.
	RoleNames            []string        // RoleNames contains enabled role names bound to the user.
	MenuIds              []int           // MenuIds contains all menu IDs reachable through the user's roles.
	Permissions          []string        // Permissions contains effective menu and button permissions after plugin filtering.
	DataScope            datascope.Scope // DataScope is the widest enabled role data-scope for governed resources.
	DataScopeUnsupported bool            // DataScopeUnsupported reports whether an enabled role carries an unsupported data-scope value.
	UnsupportedDataScope int             // UnsupportedDataScope stores the first unsupported role data-scope value.
	IsSuperAdmin         bool            // IsSuperAdmin reports whether the user is the built-in admin account.
}

// GetUserAccessContext loads the user's roles, menus, and permissions with token-aware caching when available.
func (s *serviceImpl) GetUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error) {
	tokenID := s.resolveAccessTokenID(ctx)
	if tokenID != "" {
		return s.getTokenAccessContext(ctx, tokenID, userId)
	}
	if _, err := s.getAccessRevision(ctx); err != nil {
		return nil, err
	}
	return s.loadUserAccessContext(ctx, userId)
}

// loadUserAccessContext loads the user's roles, menus, and permissions directly from storage.
func (s *serviceImpl) loadUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error) {
	isSuperAdmin, err := s.isDefaultAdminUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	roleIds, err := s.GetUserRoleIds(ctx, userId)
	if err != nil {
		return nil, err
	}

	roles, err := s.getUserRolesByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}

	menuIds := []int{}
	permissions := []string{}
	if isSuperAdmin {
		menuIds, permissions, err = s.loadAllEnabledMenuAccess(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		menuIds, err = s.getUserMenuIdsByRoleIds(ctx, roleIds)
		if err != nil {
			return nil, err
		}

		permissions, err = s.getUserPermissionsByMenuIds(ctx, menuIds)
		if err != nil {
			return nil, err
		}
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		roleNames = append(roleNames, s.DisplayName(ctx, role))
	}

	if roleNames == nil {
		roleNames = []string{}
	}
	if menuIds == nil {
		menuIds = []int{}
	}
	if permissions == nil {
		permissions = []string{}
	}
	dataScope, unsupported, unsupportedValue := resolveEffectiveDataScope(roles, isSuperAdmin)

	return &UserAccessContext{
		RoleIds:              roleIds,
		RoleNames:            roleNames,
		MenuIds:              menuIds,
		Permissions:          permissions,
		DataScope:            dataScope,
		DataScopeUnsupported: unsupported,
		UnsupportedDataScope: unsupportedValue,
		IsSuperAdmin:         isSuperAdmin,
	}, nil
}

// GetUserDataScopeSnapshot returns the user's effective role data-scope using
// the same token-bound access snapshot as menu and button permissions.
func (s *serviceImpl) GetUserDataScopeSnapshot(ctx context.Context, userId int) (*datascope.AccessSnapshot, error) {
	if userId <= 0 {
		return &datascope.AccessSnapshot{UserID: userId, Scope: datascope.ScopeNone}, nil
	}
	accessContext, err := s.GetUserAccessContext(ctx, userId)
	if err != nil {
		return nil, err
	}
	if accessContext == nil {
		return &datascope.AccessSnapshot{UserID: userId, Scope: datascope.ScopeNone}, nil
	}
	if accessContext.DataScopeUnsupported {
		return nil, bizerr.NewCode(
			datascope.CodeDataScopeUnsupported,
			bizerr.P("scope", accessContext.UnsupportedDataScope),
		)
	}
	return &datascope.AccessSnapshot{
		UserID:       userId,
		Scope:        accessContext.DataScope,
		IsSuperAdmin: accessContext.IsSuperAdmin,
	}, nil
}

// resolveEffectiveDataScope collapses enabled role data-scope values into the
// widest range that can be cached with the token access context.
func resolveEffectiveDataScope(roles []*entity.SysRole, isSuperAdmin bool) (datascope.Scope, bool, int) {
	if isSuperAdmin {
		return datascope.ScopeAll, false, 0
	}

	scope := datascope.ScopeNone
	for _, role := range roles {
		if role == nil {
			continue
		}
		switch datascope.Scope(role.DataScope) {
		case datascope.ScopeAll:
			return datascope.ScopeAll, false, 0
		case datascope.ScopeDept:
			if scope == datascope.ScopeNone || scope == datascope.ScopeSelf {
				scope = datascope.ScopeDept
			}
		case datascope.ScopeSelf:
			if scope == datascope.ScopeNone {
				scope = datascope.ScopeSelf
			}
		default:
			return datascope.ScopeNone, true, role.DataScope
		}
	}
	return scope, false, 0
}

// isDefaultAdminUser reports whether the requested user ID belongs to the
// built-in administrator account.
func (s *serviceImpl) isDefaultAdminUser(ctx context.Context, userId int) (bool, error) {
	if userId <= 0 {
		return false, nil
	}

	var (
		cols = dao.SysUser.Columns()
		user *entity.SysUser
	)

	err := dao.SysUser.Ctx(ctx).
		Fields(cols.Id, cols.Username).
		Where(cols.Id, userId).
		Scan(&user)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return accountpolicy.IsBuiltInAdminUsername(user.Username), nil
}

// getUserRolesByRoleIds loads only enabled roles because disabled roles must no
// longer contribute names or grants to the effective access snapshot.
func (s *serviceImpl) getUserRolesByRoleIds(ctx context.Context, roleIds []int) ([]*entity.SysRole, error) {
	if len(roleIds) == 0 {
		return []*entity.SysRole{}, nil
	}

	var (
		cols  = dao.SysRole.Columns()
		roles []*entity.SysRole
	)

	err := dao.SysRole.Ctx(ctx).
		WhereIn(cols.Id, roleIds).
		Where(cols.Status, 1).
		Scan(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// getUserMenuIdsByRoleIds flattens the role-menu relation into one deduplicated
// menu ID list that later drives permission resolution.
func (s *serviceImpl) getUserMenuIdsByRoleIds(ctx context.Context, roleIds []int) ([]int, error) {
	if len(roleIds) == 0 {
		return []int{}, nil
	}

	var (
		rmCols    = dao.SysRoleMenu.Columns()
		roleMenus []*entity.SysRoleMenu
	)

	err := dao.SysRoleMenu.Ctx(ctx).
		WhereIn(rmCols.RoleId, roleIds).
		Scan(&roleMenus)
	if err != nil {
		return nil, err
	}

	menuIds := make([]int, 0, len(roleMenus))
	menuIdSet := make(map[int]bool, len(roleMenus))
	for _, roleMenu := range roleMenus {
		if roleMenu == nil || menuIdSet[roleMenu.MenuId] {
			continue
		}
		menuIds = append(menuIds, roleMenu.MenuId)
		menuIdSet[roleMenu.MenuId] = true
	}
	return menuIds, nil
}

// getUserPermissionsByMenuIds resolves enabled menu permissions and lets the
// plugin layer filter out permissions hidden by plugin enablement or callbacks.
func (s *serviceImpl) getUserPermissionsByMenuIds(ctx context.Context, menuIds []int) ([]string, error) {
	if len(menuIds) == 0 {
		return []string{}, nil
	}

	var (
		menuCols = dao.SysMenu.Columns()
		menus    []*entity.SysMenu
	)

	err := dao.SysMenu.Ctx(ctx).
		WhereIn(menuCols.Id, menuIds).
		Where(menuCols.Status, 1).
		Scan(&menus)
	if err != nil {
		return nil, err
	}

	return s.collectPermissionsFromMenus(ctx, menus)
}

// loadAllEnabledMenuAccess resolves the enabled menu and permission snapshot
// for the built-in admin account without depending on role bindings.
func (s *serviceImpl) loadAllEnabledMenuAccess(ctx context.Context) ([]int, []string, error) {
	var (
		cols  = dao.SysMenu.Columns()
		menus []*entity.SysMenu
	)

	err := dao.SysMenu.Ctx(ctx).
		Where(cols.Status, 1).
		OrderAsc(cols.Id).
		Scan(&menus)
	if err != nil {
		return nil, nil, err
	}

	menuIds := make([]int, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		menuIds = append(menuIds, menu.Id)
	}

	permissions, err := s.collectPermissionsFromMenus(ctx, menus)
	if err != nil {
		return nil, nil, err
	}
	return menuIds, permissions, nil
}

// collectPermissionsFromMenus normalizes distinct permission strings from the
// supplied menu set after plugin-state filtering is applied.
func (s *serviceImpl) collectPermissionsFromMenus(ctx context.Context, menus []*entity.SysMenu) ([]string, error) {
	// Plugin runtime state can hide permission menus even when the backing menu
	// rows exist, so filtering must happen before permission strings are emitted.
	menus = s.permissionFilter.FilterPermissionMenus(ctx, menus)

	perms := make([]string, 0, len(menus))
	seen := make(map[string]struct{}, len(menus))
	for _, menu := range menus {
		if menu == nil || menu.Perms == "" {
			continue
		}
		if _, ok := seen[menu.Perms]; ok {
			continue
		}
		seen[menu.Perms] = struct{}{}
		perms = append(perms, menu.Perms)
	}
	return perms, nil
}
