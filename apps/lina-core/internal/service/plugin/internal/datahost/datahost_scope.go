// This file applies host role data-scope rules to structured data host
// service requests.

package datahost

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

const (
	resourceDataScopeNone = 0
	resourceDataScopeAll  = 1
	resourceDataScopeDept = 2
	resourceDataScopeSelf = 3
)

func applyResourceDataScope(
	ctx context.Context,
	model *gdb.Model,
	resource *catalog.ResourceSpec,
	identity *pluginbridge.IdentitySnapshotV1,
) (*gdb.Model, error) {
	if model == nil || resource == nil || resource.DataScope == nil {
		return model, nil
	}
	if identity != nil && identity.IsSuperAdmin {
		return model, nil
	}
	if identity == nil || identity.UserID <= 0 {
		return nil, gerror.Newf("data table %s 需要用户上下文以应用数据范围", resource.Table)
	}

	scope, err := getCurrentResourceDataScope(ctx, int(identity.UserID))
	if err != nil {
		return nil, err
	}
	switch scope {
	case resourceDataScopeAll:
		return model, nil
	case resourceDataScopeDept:
		if resource.DataScope.DeptColumn == "" {
			return model.Where("1 = 0"), nil
		}
		deptIDs, deptErr := getCurrentResourceDeptIDs(ctx, int(identity.UserID))
		if deptErr != nil {
			return nil, deptErr
		}
		if len(deptIDs) == 0 {
			return model.Where("1 = 0"), nil
		}
		return model.WhereIn(resource.DataScope.DeptColumn, deptIDs), nil
	case resourceDataScopeSelf:
		if resource.DataScope.UserColumn == "" {
			return model.Where("1 = 0"), nil
		}
		return model.Where(resource.DataScope.UserColumn, identity.UserID), nil
	default:
		return model.Where("1 = 0"), nil
	}
}

func getCurrentResourceDataScope(ctx context.Context, userID int) (int, error) {
	roleIDs, err := getCurrentResourceRoleIDs(ctx, userID)
	if err != nil {
		return resourceDataScopeNone, err
	}
	if len(roleIDs) == 0 {
		return resourceDataScopeNone, nil
	}

	var roles []*entity.SysRole
	err = dao.SysRole.Ctx(ctx).
		WhereIn(dao.SysRole.Columns().Id, roleIDs).
		Scan(&roles)
	if err != nil {
		return resourceDataScopeNone, err
	}

	scope := resourceDataScopeNone
	for _, roleItem := range roles {
		if roleItem == nil {
			continue
		}
		switch roleItem.DataScope {
		case resourceDataScopeAll:
			return resourceDataScopeAll, nil
		case resourceDataScopeDept:
			if scope == resourceDataScopeNone || scope == resourceDataScopeSelf {
				scope = resourceDataScopeDept
			}
		case resourceDataScopeSelf:
			if scope == resourceDataScopeNone {
				scope = resourceDataScopeSelf
			}
		default:
			return resourceDataScopeNone, gerror.Newf("unsupported role data scope: %d", roleItem.DataScope)
		}
	}
	return scope, nil
}

func getCurrentResourceRoleIDs(ctx context.Context, userID int) ([]int, error) {
	var userRoles []*entity.SysUserRole
	err := dao.SysUserRole.Ctx(ctx).
		Where(dao.SysUserRole.Columns().UserId, userID).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int, 0, len(userRoles))
	seen := make(map[int]struct{}, len(userRoles))
	for _, item := range userRoles {
		if item == nil {
			continue
		}
		if _, ok := seen[item.RoleId]; ok {
			continue
		}
		seen[item.RoleId] = struct{}{}
		roleIDs = append(roleIDs, item.RoleId)
	}
	return roleIDs, nil
}

func getCurrentResourceDeptIDs(ctx context.Context, userID int) ([]int, error) {
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Where(dao.SysUserDept.Columns().UserId, userID).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	seen := make(map[int]struct{}, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		if _, ok := seen[item.DeptId]; ok {
			continue
		}
		seen[item.DeptId] = struct{}{}
		deptIDs = append(deptIDs, item.DeptId)
	}
	return deptIDs, nil
}
