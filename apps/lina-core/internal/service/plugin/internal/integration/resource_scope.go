// This file applies host role data-scope rules to plugin-owned generic resource
// queries so dynamic plugins can reuse Lina governance semantics.

package integration

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

const (
	pluginResourceDataScopeNone = 0
	pluginResourceDataScopeAll  = 1
	pluginResourceDataScopeDept = 2
	pluginResourceDataScopeSelf = 3
)

// applyPluginResourceDataScope injects host role data-scope constraints into one plugin resource query.
func (s *serviceImpl) applyPluginResourceDataScope(
	ctx context.Context,
	model *gdb.Model,
	resource *catalog.ResourceSpec,
) (*gdb.Model, error) {
	if model == nil || resource == nil || resource.DataScope == nil {
		return model, nil
	}

	currentUserID := s.getCurrentPluginResourceUserID(ctx)
	if currentUserID <= 0 {
		return model.Where("1 = 0"), nil
	}

	scope, err := s.getCurrentPluginResourceDataScope(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	switch scope {
	case pluginResourceDataScopeAll:
		return model, nil
	case pluginResourceDataScopeDept:
		if resource.DataScope.DeptColumn == "" {
			return model.Where("1 = 0"), nil
		}
		deptIDs, deptErr := s.getCurrentPluginResourceDeptIDs(ctx, currentUserID)
		if deptErr != nil {
			return nil, deptErr
		}
		if len(deptIDs) == 0 {
			return model.Where("1 = 0"), nil
		}
		return model.WhereIn(resource.DataScope.DeptColumn, deptIDs), nil
	case pluginResourceDataScopeSelf:
		if resource.DataScope.UserColumn == "" {
			return model.Where("1 = 0"), nil
		}
		return model.Where(resource.DataScope.UserColumn, currentUserID), nil
	default:
		return model.Where("1 = 0"), nil
	}
}

// getCurrentPluginResourceUserID returns the current request user ID from the business context.
func (s *serviceImpl) getCurrentPluginResourceUserID(ctx context.Context) int {
	if s.bizCtxSvc == nil {
		return 0
	}
	return s.bizCtxSvc.GetUserId(ctx)
}

// getCurrentPluginResourceDataScope resolves the effective data scope for the given user
// by scanning all assigned roles and applying the most permissive scope that applies.
func (s *serviceImpl) getCurrentPluginResourceDataScope(ctx context.Context, userID int) (int, error) {
	roleIDs, err := s.getPluginResourceRoleIDs(ctx, userID)
	if err != nil {
		return pluginResourceDataScopeNone, err
	}
	if len(roleIDs) == 0 {
		return pluginResourceDataScopeNone, nil
	}

	var roles []*entity.SysRole
	err = dao.SysRole.Ctx(ctx).
		WhereIn(dao.SysRole.Columns().Id, roleIDs).
		Scan(&roles)
	if err != nil {
		return pluginResourceDataScopeNone, err
	}

	scope := pluginResourceDataScopeNone
	for _, roleItem := range roles {
		if roleItem == nil {
			continue
		}
		switch roleItem.DataScope {
		case pluginResourceDataScopeAll:
			return pluginResourceDataScopeAll, nil
		case pluginResourceDataScopeDept:
			if scope == pluginResourceDataScopeNone || scope == pluginResourceDataScopeSelf {
				scope = pluginResourceDataScopeDept
			}
		case pluginResourceDataScopeSelf:
			if scope == pluginResourceDataScopeNone {
				scope = pluginResourceDataScopeSelf
			}
		default:
			return pluginResourceDataScopeNone, gerror.Newf("unsupported role data scope: %d", roleItem.DataScope)
		}
	}
	return scope, nil
}

// getPluginResourceRoleIDs returns the deduplicated role IDs assigned to the given user.
func (s *serviceImpl) getPluginResourceRoleIDs(ctx context.Context, userID int) ([]int, error) {
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

// getCurrentPluginResourceDeptIDs returns the deduplicated department IDs for the given user.
func (s *serviceImpl) getCurrentPluginResourceDeptIDs(ctx context.Context, userID int) ([]int, error) {
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
