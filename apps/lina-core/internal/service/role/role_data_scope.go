// This file applies shared data-scope rules to role authorization user lists.

package role

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/service/datascope"
)

// applyRoleUserDataScope filters users displayed in role authorization pages.
func (s *serviceImpl) applyRoleUserDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, bool, error) {
	return s.currentScopeSvc().ApplyUserScope(ctx, model, qualifiedSysUserIDColumn())
}

// ensureRoleUsersVisible rejects authorization changes targeting invisible users.
func (s *serviceImpl) ensureRoleUsersVisible(ctx context.Context, userIDs []int) error {
	return s.currentScopeSvc().EnsureUsersVisible(ctx, userIDs)
}

// currentScopeSvc returns the shared data-scope service for role user operations.
func (s *serviceImpl) currentScopeSvc() datascope.Service {
	return datascope.New(datascope.Dependencies{
		BizCtxSvc: s.bizCtxSvc,
		RoleSvc:   s,
		OrgCapSvc: s.orgCapSvc,
	})
}

// qualifiedSysUserIDColumn returns the fully qualified user ID column.
func qualifiedSysUserIDColumn() string {
	return dao.SysUser.Table() + "." + dao.SysUser.Columns().Id
}
