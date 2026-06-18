// This file adapts host role and menu storage to plugin-visible
// authorization capability contracts.
package capabilityhost

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service exposes the authorization domain service and management commands.
type authzCapabilityService interface {
	authz.Service
	authz.AdminService
}

// adapter implements authorization-domain checks without exposing role tables.
type authzCapabilityAdapter struct{}

var (
	_ authz.Service      = (*authzCapabilityAdapter)(nil)
	_ authz.AdminService = (*authzCapabilityAdapter)(nil)
)

// New creates the host-owned authorization capability adapter.
func newAuthzCapabilityAdapter() authzCapabilityService {
	return &authzCapabilityAdapter{}
}

// BatchGetPermissions returns stable permission projections for non-empty keys.
func (a *authzCapabilityAdapter) BatchGetPermissions(_ context.Context, _ capmodel.CapabilityContext, keys []authz.PermissionKey) (*capmodel.BatchResult[*authz.PermissionProjection, authz.PermissionKey], error) {
	result := &capmodel.BatchResult[*authz.PermissionProjection, authz.PermissionKey]{
		Items:      make(map[authz.PermissionKey]*authz.PermissionProjection, len(keys)),
		MissingIDs: []authz.PermissionKey{},
	}
	for _, key := range keys {
		if strings.TrimSpace(string(key)) == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		result.Items[key] = &authz.PermissionProjection{
			Key:      key,
			LabelKey: "permission." + string(key),
		}
	}
	return result, nil
}

// BatchHasPermissions reports whether the current actor snapshot grants each permission key.
func (a *authzCapabilityAdapter) BatchHasPermissions(_ context.Context, capCtx capmodel.CapabilityContext, keys []authz.PermissionKey) (map[authz.PermissionKey]bool, error) {
	if len(keys) > authz.MaxBatchHasPermissions {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", authz.MaxBatchHasPermissions))
	}
	result := make(map[authz.PermissionKey]bool, len(keys))
	granted := make(map[string]struct{}, len(capCtx.Authorization.Permissions))
	for _, permission := range capCtx.Authorization.Permissions {
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
func (a *authzCapabilityAdapter) HasPermission(ctx context.Context, capCtx capmodel.CapabilityContext, key authz.PermissionKey) (bool, error) {
	result, err := a.BatchHasPermissions(ctx, capCtx, []authz.PermissionKey{key})
	if err != nil {
		return false, err
	}
	return result[key], nil
}

// IsPlatformAdmin reports whether one user has a platform all-data role.
func (a *authzCapabilityAdapter) IsPlatformAdmin(ctx context.Context, _ capmodel.CapabilityContext, userID authz.UserID) (bool, error) {
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
		Where("ur."+userRoleCols.TenantId, PlatformTenantID).
		Where("r."+roleCols.DataScope, AuthorizationPlatformAllDataScope).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReplaceRolePermissions replaces one role's permission bindings and advances
// the authorization revision after the transaction commits successfully.
func (a *authzCapabilityAdapter) ReplaceRolePermissions(ctx context.Context, capCtx capmodel.CapabilityContext, roleID authz.RoleID, keys []authz.PermissionKey) error {
	tenantID, err := TenantID(capCtx.TenantID)
	if err != nil || tenantID < PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedRoleID, err := strconv.Atoi(strings.TrimSpace(string(roleID)))
	if err != nil || parsedRoleID <= 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	menuIDs, err := a.permissionMenuIDs(ctx, keys)
	if err != nil {
		return err
	}
	return dao.SysRoleMenu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		roleCols := dao.SysRole.Columns()
		count, countErr := tx.Model(dao.SysRole.Table()).Safe().Ctx(ctx).
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
		if _, err = tx.Model(dao.SysRoleMenu.Table()).Safe().Ctx(ctx).
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
			if _, err = tx.Model(dao.SysRoleMenu.Table()).Safe().Ctx(ctx).Data(data).Insert(); err != nil {
				return err
			}
		}
		return BumpSharedRevision(
			ctx,
			tx,
			AuthorizationCacheDomain,
			AuthorizationCacheScopeGlobal,
			AuthorizationChangeReason,
		)
	})
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
