// This file verifies admin-account permission bypass semantics backed by the database.

package role

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/user/accountpolicy"
)

// TestAdminUserRetainsBypassWithoutRoleBinding verifies the built-in admin
// account still receives full access even if its admin-role association is missing.
func TestAdminUserRetainsBypassWithoutRoleBinding(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)

	adminUserID, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)

	_, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: adminUserID, RoleId: adminRoleID}).
		Delete()
	if err != nil {
		t.Fatalf("delete admin user-role binding: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUserRole.Ctx(ctx).
			Data(do.SysUserRole{UserId: adminUserID, RoleId: adminRoleID}).
			Insert(); cleanupErr != nil {
			t.Fatalf("restore admin user-role binding: %v", cleanupErr)
		}
	})

	if !svc.IsSuperAdmin(ctx, adminUserID) {
		t.Fatal("expected built-in admin user to bypass super-admin checks without role binding")
	}

	accessContext, err := svc.GetUserAccessContext(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin access context: %v", err)
	}
	if accessContext == nil || !accessContext.IsSuperAdmin {
		t.Fatal("expected admin access context to keep built-in admin bypass flag")
	}
	if accessContext.DataScope != datascope.ScopeAll {
		t.Fatalf("expected admin access context to carry all-data scope, got %d", accessContext.DataScope)
	}

	menuIDs, err := svc.GetUserMenuIds(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin menu ids: %v", err)
	}
	if len(menuIDs) == 0 {
		t.Fatal("expected built-in admin user to receive enabled menu ids")
	}

	permissions, err := svc.GetUserPermissions(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin permissions: %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("expected built-in admin user to receive enabled permissions")
	}
}

// TestAdminRoleDoesNotUpgradeOtherUsers verifies assigning the admin role to a
// non-admin username does not trigger the built-in admin bypass flag.
func TestAdminRoleDoesNotUpgradeOtherUsers(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)

	_, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)
	username := fmt.Sprintf("role-admin-%d", time.Now().UnixNano())

	userID, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: "Role Admin Test",
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert temp user: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUser.Ctx(ctx).
			Unscoped().
			Where(do.SysUser{Id: int(userID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup temp user: %v", cleanupErr)
		}
	})

	if _, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: int(userID),
		RoleId: adminRoleID,
	}).Insert(); err != nil {
		t.Fatalf("bind temp user admin role: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUserRole.Ctx(ctx).
			Where(do.SysUserRole{UserId: int(userID), RoleId: adminRoleID}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup temp user-role binding: %v", cleanupErr)
		}
	})

	if svc.IsSuperAdmin(ctx, int(userID)) {
		t.Fatal("expected non-admin username to stay outside built-in admin bypass")
	}

	accessContext, err := svc.GetUserAccessContext(ctx, int(userID))
	if err != nil {
		t.Fatalf("load temp user access context: %v", err)
	}
	if accessContext == nil {
		t.Fatal("expected temp user access context to exist")
	}
	if accessContext.IsSuperAdmin {
		t.Fatal("expected temp user access context to keep IsSuperAdmin=false")
	}
	if accessContext.DataScope != datascope.ScopeAll {
		t.Fatalf("expected admin-role user to carry all-data scope without super-admin bypass, got %d", accessContext.DataScope)
	}
}

// mustQueryAdminUserAndRoleID resolves the built-in admin user ID and role ID for tests.
func mustQueryAdminUserAndRoleID(t *testing.T, ctx context.Context) (int, int) {
	t.Helper()

	var (
		adminUser *entity.SysUser
		adminRole *entity.SysRole
	)

	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: accountpolicy.DefaultAdminUsername}).
		Scan(&adminUser)
	if err != nil {
		t.Fatalf("query built-in admin user: %v", err)
	}
	if adminUser == nil {
		t.Fatal("expected built-in admin user to exist")
	}

	err = dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Key: "admin"}).
		Scan(&adminRole)
	if err != nil {
		t.Fatalf("query built-in admin role: %v", err)
	}
	if adminRole == nil {
		t.Fatal("expected built-in admin role to exist")
	}

	return adminUser.Id, adminRole.Id
}
