// This file verifies user deletion transaction and batch-delete protections.

package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/orgcap"
	"lina-core/pkg/bizerr"
)

// TestDeleteRollsBackWhenOrgCleanupFails verifies user soft deletion is
// rolled back when organization cleanup reports an error inside the transaction.
func TestDeleteRollsBackWhenOrgCleanupFails(t *testing.T) {
	ctx := context.Background()
	userID := insertUserDeleteTestUser(t, ctx, "delete-rollback")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})

	expectedErr := errors.New("cleanup failed")
	svc := New(nil).(*serviceImpl)
	svc.orgCapSvc = userDeleteFailingOrgCap{cleanupErr: expectedErr}
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx)}}

	err := svc.Delete(ctx, userID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected cleanup error, got %v", err)
	}
	if count := mustCountUser(t, ctx, userID); count != 1 {
		t.Fatalf("expected user soft delete to be rolled back, visible count=%d", count)
	}
}

// TestBatchDeleteRejectsCurrentUserAtomically verifies current-user protection
// rejects the whole batch before deleting any selected users.
func TestBatchDeleteRejectsCurrentUserAtomically(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "current-user")
	otherUserID := insertUserDeleteTestUser(t, ctx, "other-user")
	roleID := insertUserDeleteTestRole(t, ctx, "current-user-role")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	err := svc.BatchDelete(ctx, []int{otherUserID, currentUserID})
	if err == nil {
		t.Fatal("expected current user batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserCurrentDeleteDenied) {
		t.Fatalf("expected CodeUserCurrentDeleteDenied, got %v", err)
	}
	if count := mustCountUser(t, ctx, otherUserID); count != 1 {
		t.Fatalf("expected other user to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRemovesUsersAndAssociations verifies batch deletion soft
// deletes users and clears user-role associations in one service call.
func TestBatchDeleteRemovesUsersAndAssociations(t *testing.T) {
	ctx := context.Background()
	userIDs := []int{
		insertUserDeleteTestUser(t, ctx, "batch-delete-a"),
		insertUserDeleteTestUser(t, ctx, "batch-delete-b"),
	}
	roleID := insertUserDeleteTestRole(t, ctx, "batch-delete-role")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, userIDs)
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})

	for _, userID := range userIDs {
		if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			UserId: userID,
			RoleId: roleID,
		}).Insert(); err != nil {
			t.Fatalf("insert user-role relation: %v", err)
		}
	}

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx)}}
	if err := svc.BatchDelete(ctx, userIDs); err != nil {
		t.Fatalf("batch delete users: %v", err)
	}
	for _, userID := range userIDs {
		if count := mustCountUser(t, ctx, userID); count != 0 {
			t.Fatalf("expected user %d to be soft-deleted, visible count=%d", userID, count)
		}
		if count := mustCountUserRoles(t, ctx, userID); count != 0 {
			t.Fatalf("expected user-role rows for user %d to be deleted, count=%d", userID, count)
		}
	}
}

// TestBatchDeleteRejectsBuiltinAdminAtomically verifies built-in administrator
// protection rejects the whole batch before deleting any selected users.
func TestBatchDeleteRejectsBuiltinAdminAtomically(t *testing.T) {
	ctx := context.Background()
	otherUserID := insertUserDeleteTestUser(t, ctx, "other-admin-guard")
	adminUserID := mustQueryBuiltinAdminUserID(t, ctx)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{otherUserID})
	})

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: adminUserID}}
	err := svc.BatchDelete(ctx, []int{otherUserID, adminUserID})
	if err == nil {
		t.Fatal("expected builtin admin batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserBuiltinAdminDeleteDenied) {
		t.Fatalf("expected CodeUserBuiltinAdminDeleteDenied, got %v", err)
	}
	if count := mustCountUser(t, ctx, otherUserID); count != 1 {
		t.Fatalf("expected other user to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRejectsEmptyList verifies empty batch deletes return a stable
// bizerr code before touching the database.
func TestBatchDeleteRejectsEmptyList(t *testing.T) {
	err := New(nil).BatchDelete(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserDeleteIdsRequired) {
		t.Fatalf("expected CodeUserDeleteIdsRequired, got %v", err)
	}
}

// insertUserDeleteTestUser inserts a temporary user row for delete tests.
func insertUserDeleteTestUser(t *testing.T, ctx context.Context, label string) int {
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

// insertUserDeleteTestRole inserts one temporary role row for user deletion tests.
func insertUserDeleteTestRole(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: int(userDataScopeAll),
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test role: %v", err)
	}
	return int(id)
}

// insertUserDeleteTestUserRole inserts one temporary user-role relation.
func insertUserDeleteTestUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: userID,
		RoleId: roleID,
	}).Insert(); err != nil {
		t.Fatalf("insert test user-role relation: %v", err)
	}
}

// cleanupUserDeleteTestRows removes temporary user rows and their role bindings.
func cleanupUserDeleteTestRows(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, userIDs).Delete(); err != nil {
		t.Fatalf("cleanup user-role rows: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, userIDs).Delete(); err != nil {
		t.Fatalf("cleanup user rows: %v", err)
	}
}

// cleanupUserDeleteTestRoles removes temporary role rows and their bindings.
func cleanupUserDeleteTestRoles(t *testing.T, ctx context.Context, roleIDs []int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, roleIDs).Delete(); err != nil {
		t.Fatalf("cleanup user-role role rows: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, roleIDs).Delete(); err != nil {
		t.Fatalf("cleanup role rows: %v", err)
	}
}

// mustCountUser counts visible user rows for one user ID.
func mustCountUser(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	count, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: userID}).Count()
	if err != nil {
		t.Fatalf("count user: %v", err)
	}
	return count
}

// mustCountUserRoles counts user-role rows for one user ID.
func mustCountUserRoles(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{UserId: userID}).Count()
	if err != nil {
		t.Fatalf("count user-role rows: %v", err)
	}
	return count
}

// mustQueryBuiltinAdminUserID resolves the built-in admin user ID for tests.
func mustQueryBuiltinAdminUserID(t *testing.T, ctx context.Context) int {
	t.Helper()

	var adminUser *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Username: "admin"}).Scan(&adminUser); err != nil {
		t.Fatalf("query built-in admin user: %v", err)
	}
	if adminUser == nil {
		t.Fatal("expected built-in admin user to exist")
	}
	return adminUser.Id
}

// userDeleteStaticBizCtx returns a fixed business context for current-user tests.
type userDeleteStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by service tests because they inject context directly.
func (s userDeleteStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s userDeleteStaticBizCtx) Get(context.Context) *model.Context {
	return s.ctx
}

// SetLocale is unused by delete tests.
func (s userDeleteStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by delete tests.
func (s userDeleteStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetTenant is unused by delete tests.
func (s userDeleteStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by delete tests.
func (s userDeleteStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}


// SetUserAccess is unused by delete tests.
func (s userDeleteStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// userDeleteFailingOrgCap fails cleanup while otherwise behaving as disabled.
type userDeleteFailingOrgCap struct {
	cleanupErr error
}

// Enabled reports the optional organization capability as disabled.
func (f userDeleteFailingOrgCap) Enabled(context.Context) bool { return false }

// ListUserDeptAssignments returns no organization assignments.
func (f userDeleteFailingOrgCap) ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error) {
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// GetUserIDsByDept returns no users for a department.
func (f userDeleteFailingOrgCap) GetUserIDsByDept(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// GetAllAssignedUserIDs returns no assigned users.
func (f userDeleteFailingOrgCap) GetAllAssignedUserIDs(context.Context) ([]int, error) {
	return []int{}, nil
}

// GetUserDeptInfo returns an empty department projection.
func (f userDeleteFailingOrgCap) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 0, "", nil
}

// GetUserDeptName returns an empty department name.
func (f userDeleteFailingOrgCap) GetUserDeptName(context.Context, int) (string, error) {
	return "", nil
}

// GetUserDeptIDs returns no department IDs.
func (f userDeleteFailingOrgCap) GetUserDeptIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ApplyUserDeptScope reports an empty department scope for delete tests.
func (f userDeleteFailingOrgCap) ApplyUserDeptScope(_ context.Context, model *gdb.Model, _ string, _ int) (*gdb.Model, bool, error) {
	return model, true, nil
}

// BuildUserDeptScopeExists reports an empty department scope for delete tests.
func (f userDeleteFailingOrgCap) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	return nil, true, nil
}

// GetUserPostIDs returns no post IDs.
func (f userDeleteFailingOrgCap) GetUserPostIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ReplaceUserAssignments accepts assignment replacement without doing work.
func (f userDeleteFailingOrgCap) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments returns the configured cleanup error.
func (f userDeleteFailingOrgCap) CleanupUserAssignments(context.Context, int) error {
	return f.cleanupErr
}

// UserDeptTree returns no department tree nodes.
func (f userDeleteFailingOrgCap) UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error) {
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns no post options.
func (f userDeleteFailingOrgCap) ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error) {
	return []*orgcap.PostOption{}, nil
}
