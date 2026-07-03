// This file adapts host role and menu storage to plugin-visible
// authorization capability contracts.
package role

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
)

// adapter implements authorization-domain checks without exposing role tables.
type authzCapabilityAdapter struct {
	access        Service
	bizCtx        bizctxcap.Service
	cacheCoordSvc cachecoord.Service
}

var _ authz.Service = (*authzCapabilityAdapter)(nil)

// NewCapabilityAdapter creates the host-owned authorization capability adapter.
func NewCapabilityAdapter(
	access Service,
	bizCtx bizctxcap.Service,
	cacheCoordSvc cachecoord.Service,
) authz.Service {
	return &authzCapabilityAdapter{access: access, bizCtx: bizCtx, cacheCoordSvc: cacheCoordSvc}
}

// BatchGetPermissions returns stable permission projections for non-empty keys.
func (a *authzCapabilityAdapter) BatchGetPermissions(_ context.Context, keys []authz.PermissionKey) (*capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey], error) {
	result := &capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey]{
		Items:      make(map[authz.PermissionKey]*authz.PermissionInfo, len(keys)),
		MissingIDs: []authz.PermissionKey{},
	}
	for _, key := range keys {
		if strings.TrimSpace(string(key)) == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &authz.PermissionInfo{
			Key:      key,
			LabelKey: "permission." + string(key),
		}
	}
	return result, nil
}

// BatchHasPermissions reports whether the current request grants each permission key.
func (a *authzCapabilityAdapter) BatchHasPermissions(ctx context.Context, keys []authz.PermissionKey) (map[authz.PermissionKey]bool, error) {
	if len(keys) > authz.MaxBatchHasPermissions {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", authz.MaxBatchHasPermissions))
	}
	result := make(map[authz.PermissionKey]bool, len(keys))
	current := a.current(ctx)
	if current.IsSuperAdmin {
		for _, key := range keys {
			result[key] = strings.TrimSpace(string(key)) != ""
		}
		return result, nil
	}
	permissions, err := a.currentPermissions(ctx, current)
	if err != nil {
		return nil, err
	}
	granted := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		normalizedPermission := strings.TrimSpace(permission)
		if normalizedPermission == "" {
			continue
		}
		granted[normalizedPermission] = struct{}{}
	}
	for _, key := range keys {
		normalizedKey := strings.TrimSpace(string(key))
		if normalizedKey == "" {
			result[key] = false
			continue
		}
		_, ok := granted[normalizedKey]
		result[key] = ok
	}
	return result, nil
}

// HasPermission reports whether the actor has one permission in the current scope.
func (a *authzCapabilityAdapter) HasPermission(ctx context.Context, key authz.PermissionKey) (bool, error) {
	result, err := a.BatchHasPermissions(ctx, []authz.PermissionKey{key})
	if err != nil {
		return false, err
	}
	return result[key], nil
}

// IsPlatformAdmin reports whether one user has a platform all-data role.
func (a *authzCapabilityAdapter) IsPlatformAdmin(ctx context.Context, userID authz.UserID) (bool, error) {
	parsedID, err := strconv.Atoi(strings.TrimSpace(string(userID)))
	if err != nil || parsedID <= 0 {
		return false, nil
	}
	userRoleCols := dao.SysUserRole.Columns()
	roleCols := dao.SysRole.Columns()
	count, err := dao.SysUserRole.Ctx(ctx).
		As("ur").
		InnerJoin(dao.SysRole.Table()+" r", "r."+roleCols.Id+" = ur."+userRoleCols.RoleId).
		Where("ur."+userRoleCols.UserId, parsedID).
		Where("ur."+userRoleCols.TenantId, datascope.PlatformTenantID).
		Where("r."+roleCols.DataScope, roleDataScopeAll).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReplaceRolePermissions replaces one role's permission bindings and advances
// the authorization revision after the transaction commits successfully.
func (a *authzCapabilityAdapter) ReplaceRolePermissions(ctx context.Context, roleID authz.RoleID, keys []authz.PermissionKey) error {
	current := a.current(ctx)
	if current.UserID <= 0 && !current.PlatformBypass {
		return bizerr.NewCode(capmodel.CodeCapabilityCurrentUserRequired)
	}
	tenantID := current.TenantID
	if tenantID < datascope.PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedRoleID, err := strconv.Atoi(strings.TrimSpace(string(roleID)))
	if err != nil || parsedRoleID <= 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if a == nil || a.cacheCoordSvc == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "cachecoord"))
	}
	menuIDs, err := a.permissionMenuIDs(ctx, keys)
	if err != nil {
		return err
	}
	if err = dao.SysRoleMenu.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		roleCols := dao.SysRole.Columns()
		count, countErr := dao.SysRole.Ctx(ctx).
			Where(roleCols.Id, parsedRoleID).
			Where(roleCols.TenantId, tenantID).
			Count()
		if countErr != nil {
			return countErr
		}
		if count == 0 {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		roleMenuCols := dao.SysRoleMenu.Columns()
		if _, err = dao.SysRoleMenu.Ctx(ctx).
			Where(roleMenuCols.RoleId, parsedRoleID).
			Where(roleMenuCols.TenantId, tenantID).
			Delete(); err != nil {
			return err
		}
		if len(menuIDs) > 0 {
			data := make([]do.SysRoleMenu, 0, len(menuIDs))
			for _, menuID := range menuIDs {
				data = append(data, do.SysRoleMenu{
					TenantId: tenantID,
					RoleId:   parsedRoleID,
					MenuId:   menuID,
				})
			}
			if _, err = dao.SysRoleMenu.Ctx(ctx).Data(data).Insert(); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return a.markAuthorizationChanged(ctx)
}

// markAuthorizationChanged publishes the permission-access revision after the
// role-menu transaction has committed successfully.
func (a *authzCapabilityAdapter) markAuthorizationChanged(ctx context.Context) error {
	revision, err := a.cacheCoordSvc.MarkTenantChanged(
		ctx,
		accessTopologyCacheDomain,
		cachecoord.ScopeGlobal,
		accessRevisionInvalidationScope(ctx),
		accessTopologyCacheChangeReason,
	)
	if err != nil {
		return err
	}
	storeLocalAccessRevision(revision)
	return nil
}

func (a *authzCapabilityAdapter) current(ctx context.Context) bizctxcap.CurrentContext {
	if a != nil && a.bizCtx != nil {
		return a.bizCtx.Current(ctx)
	}
	return bizctxcap.CurrentFromContext(ctx)
}

func (a *authzCapabilityAdapter) currentPermissions(ctx context.Context, current bizctxcap.CurrentContext) ([]string, error) {
	if len(current.Permissions) > 0 {
		return append([]string(nil), current.Permissions...), nil
	}
	if a == nil || a.access == nil || current.UserID <= 0 {
		return nil, nil
	}
	if strings.TrimSpace(current.TokenID) != "" {
		projection, err := a.access.BuildDynamicRouteAccessProjection(ctx, current.TokenID, current.UserID, current.TenantID)
		if err != nil {
			return nil, err
		}
		if projection == nil {
			return nil, nil
		}
		return append([]string(nil), projection.Permissions...), nil
	}
	access, err := a.access.GetUserAccessContext(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, nil
	}
	return append([]string(nil), access.Permissions...), nil
}

// permissionMenuIDs resolves permission keys to menu IDs in one bounded query.
func (a *authzCapabilityAdapter) permissionMenuIDs(ctx context.Context, keys []authz.PermissionKey) ([]int, error) {
	requestedKeys := make([]string, 0, len(keys))
	requested := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		normalizedKey := strings.TrimSpace(string(key))
		if normalizedKey == "" {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		if _, exists := requested[normalizedKey]; exists {
			continue
		}
		requested[normalizedKey] = struct{}{}
		requestedKeys = append(requestedKeys, normalizedKey)
	}
	if len(requestedKeys) == 0 {
		return []int{}, nil
	}

	rows := make([]*struct {
		Id    int
		Perms string
	}, 0, len(requestedKeys))
	cols := dao.SysMenu.Columns()
	if err := dao.SysMenu.Ctx(ctx).
		Fields(cols.Id, cols.Perms).
		WhereIn(cols.Perms, requestedKeys).
		Scan(&rows); err != nil {
		return nil, err
	}
	menuIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		normalizedPerms := strings.TrimSpace(row.Perms)
		if _, ok := requested[normalizedPerms]; !ok {
			continue
		}
		delete(requested, normalizedPerms)
		menuIDs = append(menuIDs, row.Id)
	}
	if len(requested) > 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return menuIDs, nil
}
