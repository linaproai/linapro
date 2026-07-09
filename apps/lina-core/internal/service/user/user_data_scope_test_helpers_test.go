// This file keeps user data-scope aliases and direct scope helpers scoped to tests.

package user

import (
	"context"

	"lina-core/internal/service/datascope"
)

// userDataScope represents the role data range used by host user management tests.
type userDataScope = datascope.Scope

// User data-scope levels follow sys_role.data_scope values in tests.
const (
	userDataScopeNone   userDataScope = datascope.ScopeNone
	userDataScopeAll    userDataScope = datascope.ScopeAll
	userDataScopeTenant userDataScope = datascope.ScopeTenant
	userDataScopeDept   userDataScope = datascope.ScopeDept
	userDataScopeSelf   userDataScope = datascope.ScopeSelf
)

// currentUserDataScope computes the widest enabled role data-scope for the current test user.
func (s *serviceImpl) currentUserDataScope(ctx context.Context) (userDataScope, int, error) {
	currentScope, err := s.currentScopeSvc().Current(ctx)
	if err != nil {
		return userDataScopeNone, 0, err
	}
	return userDataScope(currentScope.Scope), currentScope.UserID, nil
}
