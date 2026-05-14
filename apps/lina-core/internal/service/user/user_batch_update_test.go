// This file verifies atomic user batch-update behavior.

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// TestBatchUpdateReplacesStatusAndRoles verifies selected users receive the
// requested status and role set in one service call.
func TestBatchUpdateReplacesStatusAndRoles(t *testing.T) {
	ctx := context.Background()
	userIDs := []int{
		insertUserDeleteTestUser(t, ctx, "batch-update-a"),
		insertUserDeleteTestUser(t, ctx, "batch-update-b"),
	}
	oldRoleID := insertUserDeleteTestRole(t, ctx, "batch-update-old-role")
	newRoleID := insertUserDeleteTestRole(t, ctx, "batch-update-new-role")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, userIDs)
		cleanupUserDeleteTestRoles(t, ctx, []int{oldRoleID, newRoleID})
	})
	for _, userID := range userIDs {
		insertUserDeleteTestUserRole(t, ctx, userID, oldRoleID)
	}

	disabledStatus := int(StatusDisabled)
	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx), TenantId: 0, DataScope: 1}})
	if err := svc.BatchUpdate(ctx, BatchUpdateInput{
		Ids:          userIDs,
		UpdateStatus: true,
		Status:       &disabledStatus,
		UpdateRoles:  true,
		RoleIds:      []int{newRoleID},
	}); err != nil {
		t.Fatalf("batch update users: %v", err)
	}

	for _, userID := range userIDs {
		if status := mustQueryUserBatchUpdateStatus(t, ctx, userID); status != disabledStatus {
			t.Fatalf("expected user %d status %d, got %d", userID, disabledStatus, status)
		}
		if count := mustCountUserRole(t, ctx, userID, oldRoleID, 0); count != 0 {
			t.Fatalf("expected old role removed for user %d, count=%d", userID, count)
		}
		if count := mustCountUserRole(t, ctx, userID, newRoleID, 0); count != 1 {
			t.Fatalf("expected new role assigned for user %d, count=%d", userID, count)
		}
	}
}

// TestBatchUpdateRejectsCurrentUserAtomically verifies current-user protection
// rejects the whole batch before changing any selected row.
func TestBatchUpdateRejectsCurrentUserAtomically(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "batch-update-current")
	otherUserID := insertUserDeleteTestUser(t, ctx, "batch-update-other")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
	})

	disabledStatus := int(StatusDisabled)
	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID, TenantId: 0, DataScope: 1}})
	err := svc.BatchUpdate(ctx, BatchUpdateInput{
		Ids:          []int{otherUserID, currentUserID},
		UpdateStatus: true,
		Status:       &disabledStatus,
	})
	if err == nil {
		t.Fatal("expected current-user batch update to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserCurrentEditDenied) {
		t.Fatalf("expected CodeUserCurrentEditDenied, got %v", err)
	}
	if status := mustQueryUserBatchUpdateStatus(t, ctx, otherUserID); status != int(StatusNormal) {
		t.Fatalf("expected other user to remain enabled, got status=%d", status)
	}
}

// TestBatchUpdateReplacesTenantMemberships verifies platform batch tenant
// assignment rewrites active memberships and primary tenant IDs atomically.
func TestBatchUpdateReplacesTenantMemberships(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "batch-update-tenant-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "batch-update-tenant-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "batch-update-tenant-operator", 0)
	userIDs := []int{
		insertUserTenantMembershipTestUser(t, ctx, fmt.Sprintf("batch-update-tenant-u1-%d", time.Now().UnixNano()), tenantAID),
		insertUserTenantMembershipTestUser(t, ctx, fmt.Sprintf("batch-update-tenant-u2-%d", time.Now().UnixNano()), tenantAID),
	}
	roleID := insertUserDataScopeTestRole(t, ctx, "batch-update-tenant-operator-role", userDataScopeAll, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, append([]int{operatorID}, userIDs...))
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, append([]int{operatorID}, userIDs...))
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, operatorID, roleID)
	for _, userID := range userIDs {
		insertUserTenantMembershipTestMembership(t, ctx, userID, tenantAID, userTenantMembershipTestActive)
	}

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 1}})
	if err := svc.BatchUpdate(ctx, BatchUpdateInput{
		Ids:          userIDs,
		UpdateTenant: true,
		TenantIds:    []int{tenantBID},
	}); err != nil {
		t.Fatalf("batch update user tenants: %v", err)
	}

	for _, userID := range userIDs {
		if tenantID := mustQueryUserTenantID(t, ctx, userID); tenantID != tenantBID {
			t.Fatalf("expected user %d primary tenant %d, got %d", userID, tenantBID, tenantID)
		}
		if count := mustCountUserTenantMembership(t, ctx, userID, tenantAID); count != 0 {
			t.Fatalf("expected user %d tenant A membership removed, got %d", userID, count)
		}
		if count := mustCountUserTenantMembership(t, ctx, userID, tenantBID); count != 1 {
			t.Fatalf("expected user %d tenant B membership retained, got %d", userID, count)
		}
	}
}

// TestBatchUpdateRejectsRoleTenantCombinedPatch verifies role assignments and
// tenant membership rewrites cannot be combined under one current-tenant role
// selection.
func TestBatchUpdateRejectsRoleTenantCombinedPatch(t *testing.T) {
	ctx := context.Background()
	userID := insertUserDeleteTestUser(t, ctx, "batch-update-role-tenant-conflict")
	roleID := insertUserDeleteTestRole(t, ctx, "batch-update-role-tenant-conflict-role")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})

	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx), TenantId: 0, DataScope: 1}})
	err := svc.BatchUpdate(ctx, BatchUpdateInput{
		Ids:          []int{userID},
		UpdateRoles:  true,
		RoleIds:      []int{roleID},
		UpdateTenant: true,
		TenantIds:    []int{62081},
	})
	if err == nil {
		t.Fatal("expected role and tenant batch update to be rejected")
	}
	if !bizerr.Is(err, CodeUserBatchUpdateRoleTenantConflict) {
		t.Fatalf("expected CodeUserBatchUpdateRoleTenantConflict, got %v", err)
	}
	if count := mustCountUserRole(t, ctx, userID, roleID, 0); count != 0 {
		t.Fatalf("expected no role relation after rejected batch, got count=%d", count)
	}
}

// mustQueryUserBatchUpdateStatus returns one user's visible status.
func mustQueryUserBatchUpdateStatus(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	var user *modelUserBatchUpdateStatus
	if err := dao.SysUser.Ctx(ctx).Fields(dao.SysUser.Columns().Status).Where(do.SysUser{Id: userID}).Scan(&user); err != nil {
		t.Fatalf("query user status: %v", err)
	}
	if user == nil {
		t.Fatalf("expected user %d to exist", userID)
	}
	return user.Status
}

// modelUserBatchUpdateStatus is a tiny test projection for sys_user.status.
type modelUserBatchUpdateStatus struct {
	Status int `json:"status" orm:"status"`
}

// mustCountUserRole counts one user-role relation in a tenant boundary.
func mustCountUserRole(t *testing.T, ctx context.Context, userID int, roleID int, tenantID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: userID, RoleId: roleID, TenantId: tenantID}).
		Count()
	if err != nil {
		t.Fatalf("count user role: %v", err)
	}
	return count
}
