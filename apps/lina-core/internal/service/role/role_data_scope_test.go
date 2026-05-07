// This file verifies role authorization user operations respect data scope.

package role

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// TestRoleUsersApplyDataScope verifies role authorization pages and assignment
// mutations only expose or accept users inside the current data scope.
func TestRoleUsersApplyDataScope(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertRoleScopeUser(t, ctx, "role-scope-current")
	hiddenUserID := insertRoleScopeUser(t, ctx, "role-scope-hidden")
	scopeRoleID := insertRoleScopeRole(t, ctx, "role-scope-self", 3)
	managedRoleID := insertRoleScopeRole(t, ctx, "role-scope-managed", 3)
	t.Cleanup(func() {
		cleanupRoleScopeUsers(t, ctx, []int{currentUserID, hiddenUserID})
		cleanupRoleScopeRoles(t, ctx, []int{scopeRoleID, managedRoleID})
	})
	insertRoleScopeUserRole(t, ctx, currentUserID, scopeRoleID)
	insertRoleScopeUserRole(t, ctx, currentUserID, managedRoleID)
	insertRoleScopeUserRole(t, ctx, hiddenUserID, managedRoleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	out, err := svc.GetUsers(ctx, GetUsersInput{RoleId: managedRoleID, Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("get role users: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].Id != currentUserID {
		t.Fatalf("expected only current user in role users, got total=%d list=%#v", out.Total, out.List)
	}

	if err = svc.AssignUsers(ctx, managedRoleID, []int{hiddenUserID}); !bizerr.Is(err, datascope.CodeDataScopeDenied) {
		t.Fatalf("expected hidden assignment to be denied, got %v", err)
	}
	if err = svc.UnassignUser(ctx, managedRoleID, hiddenUserID); !bizerr.Is(err, datascope.CodeDataScopeDenied) {
		t.Fatalf("expected hidden unassign to be denied, got %v", err)
	}
}

// TestRoleRejectsDepartmentScopeWhenOrgCapabilityDisabled verifies role writes
// cannot persist department data scope when the organization plugin is not
// enabled.
func TestRoleRejectsDepartmentScopeWhenOrgCapabilityDisabled(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)

	if err := svc.ensureRoleDataScopeAllowed(ctx, roleDataScopeDept); !bizerr.Is(err, CodeRoleDataScopeDeptUnavailable) {
		t.Fatalf("expected department data-scope unavailable error, got %v", err)
	}
	if err := svc.ensureRoleDataScopeAllowed(ctx, roleDataScopeSelf); err != nil {
		t.Fatalf("expected self data-scope to remain available, got %v", err)
	}
}

// TestRoleAllowsDepartmentScopeWhenOrgCapabilityEnabled verifies the backend
// accepts department data scope only when an organization capability state is
// explicitly available.
func TestRoleAllowsDepartmentScopeWhenOrgCapabilityEnabled(t *testing.T) {
	ctx := context.Background()
	svc := New(roleScopeEnabledOrgState{}).(*serviceImpl)

	if err := svc.ensureRoleDataScopeAllowed(ctx, roleDataScopeDept); err != nil {
		t.Fatalf("expected enabled organization capability to allow department scope, got %v", err)
	}
}

// roleScopeEnabledOrgState exposes both permission filtering and organization
// capability state to the role service.
type roleScopeEnabledOrgState struct{}

// FilterPermissionMenus leaves test menus unchanged.
func (roleScopeEnabledOrgState) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// Enabled reports organization capability as available.
func (roleScopeEnabledOrgState) Enabled(context.Context) bool { return true }

// roleScopeStaticBizCtx returns a fixed business context.
type roleScopeStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by role data-scope tests.
func (s roleScopeStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s roleScopeStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by role data-scope tests.
func (s roleScopeStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by role data-scope tests.
func (s roleScopeStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetUserAccess is unused by role data-scope tests.
func (s roleScopeStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// insertRoleScopeUser inserts one temporary user.
func insertRoleScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()
	return insertRoleTestUser(t, ctx, prefix)
}

// insertRoleScopeRole inserts one temporary role.
func insertRoleScopeRole(t *testing.T, ctx context.Context, prefix string, scope int) int {
	t.Helper()
	roleID := insertTestRole(t, ctx, prefix)
	if _, err := dao.SysRole.Ctx(ctx).
		Where(dao.SysRole.Columns().Id, roleID).
		Data(do.SysRole{DataScope: scope}).
		Update(); err != nil {
		t.Fatalf("update role-scope data scope: %v", err)
	}
	return roleID
}

// insertRoleScopeUserRole binds one user to one role.
func insertRoleScopeUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{UserId: userID, RoleId: roleID}).Insert(); err != nil {
		t.Fatalf("insert role-scope user-role: %v", err)
	}
}

// cleanupRoleScopeUsers removes temporary users.
func cleanupRoleScopeUsers(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, ids).Delete(); err != nil {
		t.Fatalf("cleanup role-scope user roles: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup role-scope users: %v", err)
	}
}

// cleanupRoleScopeRoles removes temporary roles.
func cleanupRoleScopeRoles(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, ids).Delete(); err != nil {
		t.Fatalf("cleanup role-scope user roles by role: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup role-scope roles: %v", err)
	}
}
