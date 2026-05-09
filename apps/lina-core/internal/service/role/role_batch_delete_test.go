// This file verifies role batch deletion behavior and association cleanup.

package role

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
)

// TestBatchDeleteRemovesRolesAndAssociations verifies batch role deletion
// clears role-menu and user-role associations in one service call.
func TestBatchDeleteRemovesRolesAndAssociations(t *testing.T) {
	ctx := context.Background()
	svc := New(nil)
	roleID := insertTestRole(t, ctx, "batch-delete-role")
	userID := insertRoleTestUser(t, ctx, "batch-delete-user")
	menuID := insertRoleTestMenu(t, ctx, "batch-delete-menu")
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{menuID})
	})

	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId: roleID,
		MenuId: menuID,
	}).Insert(); err != nil {
		t.Fatalf("insert role-menu relation: %v", err)
	}
	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: userID,
		RoleId: roleID,
	}).Insert(); err != nil {
		t.Fatalf("insert user-role relation: %v", err)
	}

	if err := svc.BatchDelete(ctx, []int{roleID}); err != nil {
		t.Fatalf("batch delete role: %v", err)
	}

	if count := mustCountRole(t, ctx, roleID); count != 0 {
		t.Fatalf("expected role to be soft-deleted, visible count=%d", count)
	}
	if count := mustCountRoleMenu(t, ctx, roleID); count != 0 {
		t.Fatalf("expected role-menu relations to be deleted, got %d", count)
	}
	if count := mustCountUserRole(t, ctx, roleID); count != 0 {
		t.Fatalf("expected user-role relations to be deleted, got %d", count)
	}
}

// TestBatchDeleteRejectsBuiltinAdminRoleAtomically verifies a mixed batch with
// the built-in admin role is rejected before any custom role is deleted.
func TestBatchDeleteRejectsBuiltinAdminRoleAtomically(t *testing.T) {
	ctx := context.Background()
	svc := New(nil)
	roleID := insertTestRole(t, ctx, "batch-delete-admin-guard")
	_, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, nil, nil)
	})

	err := svc.BatchDelete(ctx, []int{roleID, adminRoleID})
	if err == nil {
		t.Fatal("expected builtin admin role batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeRoleBuiltinDeleteDenied) {
		t.Fatalf("expected CodeRoleBuiltinDeleteDenied, got %v", err)
	}
	if count := mustCountRole(t, ctx, roleID); count != 1 {
		t.Fatalf("expected custom role to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRejectsEmptyList verifies empty role batches return a stable
// bizerr code before touching the database.
func TestBatchDeleteRejectsEmptyList(t *testing.T) {
	err := New(nil).BatchDelete(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeRoleDeleteIdsRequired) {
		t.Fatalf("expected CodeRoleDeleteIdsRequired, got %v", err)
	}
}

// insertTestRole inserts one temporary role row.
func insertTestRole(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: roleDataScopeAll,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test role: %v", err)
	}
	return int(id)
}

// insertRoleTestUser inserts one temporary user row for role tests.
func insertRoleTestUser(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	username := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return int(id)
}

// insertRoleTestMenu inserts one temporary menu row for role tests.
func insertRoleTestMenu(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	key := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		MenuKey: key,
		Name:    key,
		Type:    "M",
		Sort:    1,
		Visible: 1,
		Status:  1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test menu: %v", err)
	}
	return int(id)
}

// cleanupRoleTestRows removes temporary role, user, and menu rows unscoped.
func cleanupRoleTestRows(t *testing.T, ctx context.Context, roleIDs []int, userIDs []int, menuIDs []int) {
	t.Helper()

	if len(roleIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().RoleId, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role-menu rows: %v", err)
		}
		if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup user-role rows: %v", err)
		}
		if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role rows: %v", err)
		}
	}
	if len(userIDs) > 0 {
		if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, userIDs).Delete(); err != nil {
			t.Fatalf("cleanup user-role user rows: %v", err)
		}
		if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, userIDs).Delete(); err != nil {
			t.Fatalf("cleanup user rows: %v", err)
		}
	}
	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup role-menu menu rows: %v", err)
		}
		if _, err := dao.SysMenu.Ctx(ctx).Unscoped().WhereIn(dao.SysMenu.Columns().Id, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup menu rows: %v", err)
		}
	}
}

// mustCountRole counts visible role rows for one role ID.
func mustCountRole(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: roleID}).Count()
	if err != nil {
		t.Fatalf("count role: %v", err)
	}
	return count
}

// mustCountRoleMenu counts role-menu rows for one role ID.
func mustCountRoleMenu(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysRoleMenu.Ctx(ctx).Where(do.SysRoleMenu{RoleId: roleID}).Count()
	if err != nil {
		t.Fatalf("count role-menu relations: %v", err)
	}
	return count
}

// mustCountUserRole counts user-role rows for one role ID.
func mustCountUserRole(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{RoleId: roleID}).Count()
	if err != nil {
		t.Fatalf("count user-role relations: %v", err)
	}
	return count
}
