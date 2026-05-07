// This file applies role data-scope rules to host user-management queries and
// target-record checks.

package user

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// userDataScope represents the role data range used by host user management tests.
type userDataScope = datascope.Scope

// User data-scope levels follow sys_role.data_scope values.
const (
	userDataScopeNone userDataScope = datascope.ScopeNone
	userDataScopeAll  userDataScope = datascope.ScopeAll
	userDataScopeDept userDataScope = datascope.ScopeDept
	userDataScopeSelf userDataScope = datascope.ScopeSelf
)

// applyUserDataScope injects the current user's data-scope filter into a
// sys_user model. The empty flag lets callers return an empty result when a
// scope resolves to no visible rows.
func (s *serviceImpl) applyUserDataScope(ctx context.Context, m *gdb.Model) (*gdb.Model, bool, error) {
	scopedModel, empty, err := s.currentScopeSvc().ApplyUserScope(ctx, m, qualifiedSysUserIDColumn())
	return scopedModel, empty, mapDataScopeError(err)
}

// ensureUserVisible rejects detail and mutation operations for rows outside
// the current request user's effective data-scope.
func (s *serviceImpl) ensureUserVisible(ctx context.Context, userID int) error {
	return s.ensureUsersVisible(ctx, []int{userID})
}

// ensureUsersVisible rejects a multi-target operation unless every target user
// is visible under the current request user's effective data-scope.
func (s *serviceImpl) ensureUsersVisible(ctx context.Context, userIDs []int) error {
	return mapDataScopeError(s.currentScopeSvc().EnsureUsersVisible(ctx, userIDs))
}

// qualifiedSysUserIDColumn returns the fully qualified sys_user ID column used
// by correlated orgcap constraints.
func qualifiedSysUserIDColumn() string {
	return dao.SysUser.Table() + "." + dao.SysUser.Columns().Id
}

// currentUserDataScope computes the widest enabled role data-scope for the
// current request user. The built-in administrator always receives all data.
func (s *serviceImpl) currentUserDataScope(ctx context.Context) (userDataScope, int, error) {
	currentScope, err := s.currentScopeSvc().Current(ctx)
	if err != nil {
		return userDataScopeNone, 0, mapDataScopeError(err)
	}
	return currentScope.Scope, currentScope.UserID, nil
}

// currentScopeSvc returns the shared data-scope service, lazily constructing it
// for tests that instantiate serviceImpl directly.
func (s *serviceImpl) currentScopeSvc() datascope.Service {
	return datascope.New(datascope.Dependencies{
		BizCtxSvc: s.bizCtxSvc,
		RoleSvc:   s.roleSvc,
		OrgCapSvc: s.orgCapSvc,
	})
}

// mapDataScopeError preserves user-management legacy business error codes at
// the module boundary while reusing shared data-scope internals.
func mapDataScopeError(err error) error {
	switch {
	case err == nil:
		return nil
	case bizerr.Is(err, datascope.CodeDataScopeDenied):
		return bizerr.NewCode(CodeUserDataScopeDenied)
	case bizerr.Is(err, datascope.CodeDataScopeNotAuthenticated):
		return bizerr.NewCode(CodeUserNotAuthenticated)
	case bizerr.Is(err, datascope.CodeDataScopeUnsupported):
		messageErr, ok := bizerr.As(err)
		if !ok {
			return bizerr.NewCode(CodeUserDataScopeUnsupported)
		}
		return bizerr.NewCode(CodeUserDataScopeUnsupported, bizerr.P("scope", messageErr.Params()["scope"]))
	default:
		return err
	}
}
