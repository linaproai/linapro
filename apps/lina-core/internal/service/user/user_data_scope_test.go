// This file verifies role data-scope enforcement for host user-management
// service operations.

package user

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/orgcap"
	"lina-core/pkg/bizerr"
)

// TestUserDataScopeListSelfShowsOnlyCurrentUser verifies self-scope list
// queries return only the current authenticated user.
func TestUserDataScopeListSelfShowsOnlyCurrentUser(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-self-current")
	otherUserID := insertUserDeleteTestUser(t, ctx, "scope-self-other")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-self-role", userDataScopeSelf, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list users with self data scope: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].SysUser.Id != currentUserID {
		t.Fatalf("expected only current user %d, got total=%d list=%v", currentUserID, out.Total, userDataScopeListIDs(out.List))
	}
}

// TestUserDataScopeListDeptUsesOrgCap verifies department-scope list queries
// use orgcap to inject a scoped database constraint.
func TestUserDataScopeListDeptUsesOrgCap(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-dept-current")
	deptMateUserID := insertUserDeleteTestUser(t, ctx, "scope-dept-mate")
	otherUserID := insertUserDeleteTestUser(t, ctx, "scope-dept-other")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-dept-role", userDataScopeDept, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, deptMateUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}
	svc.orgCapSvc = userDataScopeStaticOrgCap{
		enabled:     true,
		userDeptIDs: map[int][]int{currentUserID: {101}},
		deptUserIDs: map[int][]int{101: {currentUserID, deptMateUserID}},
	}

	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list users with dept data scope: %v", err)
	}
	actual := userDataScopeIDSet(out.List)
	if out.Total != 2 || len(out.List) != 2 {
		t.Fatalf("expected two dept users, got total=%d list=%v", out.Total, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[currentUserID]; !ok {
		t.Fatalf("expected current user %d in dept scope result, got %v", currentUserID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[deptMateUserID]; !ok {
		t.Fatalf("expected dept mate %d in dept scope result, got %v", deptMateUserID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[otherUserID]; ok {
		t.Fatalf("did not expect other user %d in dept scope result, got %v", otherUserID, userDataScopeListIDs(out.List))
	}
}

// TestUserDataScopeDeptUnavailableFallsBackToSelf verifies department scope
// falls back to self scope when organization capability is unavailable.
func TestUserDataScopeDeptUnavailableFallsBackToSelf(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-dept-disabled-current")
	otherUserID := insertUserDeleteTestUser(t, ctx, "scope-dept-disabled-other")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-dept-disabled-role", userDataScopeDept, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list users with unavailable dept data scope: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].SysUser.Id != currentUserID {
		t.Fatalf("expected self result when orgcap is unavailable, got total=%d list=%v", out.Total, userDataScopeListIDs(out.List))
	}
}

// TestUserDataScopeResolvesWidestEnabledRole verifies enabled role scopes are
// merged by the widest-scope-wins rule and disabled roles are ignored.
func TestUserDataScopeResolvesWidestEnabledRole(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-widest-current")
	selfRoleID := insertUserDataScopeTestRole(t, ctx, "scope-widest-self", userDataScopeSelf, 1)
	deptRoleID := insertUserDataScopeTestRole(t, ctx, "scope-widest-dept", userDataScopeDept, 1)
	disabledAllRoleID := insertUserDataScopeTestRole(t, ctx, "scope-widest-disabled-all", userDataScopeAll, 0)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{selfRoleID, deptRoleID, disabledAllRoleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, selfRoleID)
	insertUserDeleteTestUserRole(t, ctx, currentUserID, deptRoleID)
	insertUserDeleteTestUserRole(t, ctx, currentUserID, disabledAllRoleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	scope, userID, err := svc.currentUserDataScope(ctx)
	if err != nil {
		t.Fatalf("resolve current user data scope: %v", err)
	}
	if userID != currentUserID {
		t.Fatalf("expected current user ID %d, got %d", currentUserID, userID)
	}
	if scope != userDataScopeDept {
		t.Fatalf("expected dept scope after widest enabled merge, got %d", scope)
	}
}

// TestUserDataScopeRejectsInvisibleStatusUpdate verifies write operations fail
// before mutating target users outside the current data scope.
func TestUserDataScopeRejectsInvisibleStatusUpdate(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-update-current")
	targetUserID := insertUserDeleteTestUser(t, ctx, "scope-update-target")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-update-role", userDataScopeSelf, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, targetUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	err := svc.UpdateStatus(ctx, targetUserID, StatusDisabled)
	if err == nil {
		t.Fatal("expected invisible status update to be rejected")
	}
	if !bizerr.Is(err, CodeUserDataScopeDenied) {
		t.Fatalf("expected CodeUserDataScopeDenied, got %v", err)
	}
	if status := mustQueryUserStatus(t, ctx, targetUserID); status != int(StatusNormal) {
		t.Fatalf("expected target user status to remain normal, got %d", status)
	}
}

// TestUserDataScopeBatchDeleteRejectsInvisibleTarget verifies batch delete is
// rejected atomically when any target user is outside the data scope.
func TestUserDataScopeBatchDeleteRejectsInvisibleTarget(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-delete-current")
	deptMateUserID := insertUserDeleteTestUser(t, ctx, "scope-delete-mate")
	invisibleUserID := insertUserDeleteTestUser(t, ctx, "scope-delete-invisible")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-delete-role", userDataScopeDept, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, deptMateUserID, invisibleUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}
	svc.orgCapSvc = userDataScopeStaticOrgCap{
		enabled:     true,
		userDeptIDs: map[int][]int{currentUserID: {201}},
		deptUserIDs: map[int][]int{201: {currentUserID, deptMateUserID}},
	}

	err := svc.BatchDelete(ctx, []int{deptMateUserID, invisibleUserID})
	if err == nil {
		t.Fatal("expected invisible batch delete target to be rejected")
	}
	if !bizerr.Is(err, CodeUserDataScopeDenied) {
		t.Fatalf("expected CodeUserDataScopeDenied, got %v", err)
	}
	if count := mustCountUser(t, ctx, deptMateUserID); count != 1 {
		t.Fatalf("expected visible target to remain after rejected batch, count=%d", count)
	}
	if count := mustCountUser(t, ctx, invisibleUserID); count != 1 {
		t.Fatalf("expected invisible target to remain after rejected batch, count=%d", count)
	}
}

// TestUserDataScopeExportAllAppliesSelfScope verifies full export uses the same
// self-scope filter as user list queries.
func TestUserDataScopeExportAllAppliesSelfScope(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserDeleteTestUser(t, ctx, "scope-export-current")
	otherUserID := insertUserDeleteTestUser(t, ctx, "scope-export-other")
	roleID := insertUserDataScopeTestRole(t, ctx, "scope-export-role", userDataScopeSelf, 1)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := New(nil).(*serviceImpl)
	svc.bizCtxSvc = userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	data, err := svc.Export(ctx, ExportInput{})
	if err != nil {
		t.Fatalf("export users with self data scope: %v", err)
	}
	rows, err := readUserDataScopeExportRows(data)
	if err != nil {
		t.Fatalf("read exported workbook: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header and one self row, got %d rows: %v", len(rows), rows)
	}
	if len(rows[1]) == 0 || rows[1][0] == "" {
		t.Fatalf("expected exported self username in first column, got row %v", rows[1])
	}
	if rows[1][0] != mustQueryUserUsername(t, ctx, currentUserID) {
		t.Fatalf("expected exported username for current user, got row %v", rows[1])
	}
}

// insertUserDataScopeTestRole inserts one temporary role with a configurable
// data scope and status.
func insertUserDataScopeTestRole(t *testing.T, ctx context.Context, label string, scope userDataScope, status int) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: int(scope),
		Status:    status,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert data-scope test role: %v", err)
	}
	return int(id)
}

// mustQueryUserStatus returns one user's current visible status.
func mustQueryUserStatus(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	var user *modelUserStatusProjection
	if err := dao.SysUser.Ctx(ctx).
		Fields(dao.SysUser.Columns().Status).
		Where(do.SysUser{Id: userID}).
		Scan(&user); err != nil {
		t.Fatalf("query user status: %v", err)
	}
	if user == nil {
		t.Fatalf("expected user %d to exist", userID)
	}
	return user.Status
}

// mustQueryUserUsername returns one user's username.
func mustQueryUserUsername(t *testing.T, ctx context.Context, userID int) string {
	t.Helper()

	var user *modelUserUsernameProjection
	if err := dao.SysUser.Ctx(ctx).
		Fields(dao.SysUser.Columns().Username).
		Where(do.SysUser{Id: userID}).
		Scan(&user); err != nil {
		t.Fatalf("query user username: %v", err)
	}
	if user == nil {
		t.Fatalf("expected user %d to exist", userID)
	}
	return user.Username
}

// readUserDataScopeExportRows reads rows from a generated user export workbook.
func readUserDataScopeExportRows(data []byte) (rows [][]string, err error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	rows, err = f.GetRows("Sheet1")
	return rows, err
}

// userDataScopeListIDs returns user IDs from list output for clear assertions.
func userDataScopeListIDs(items []*ListOutputItem) []int {
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if item == nil || item.SysUser == nil {
			continue
		}
		ids = append(ids, item.SysUser.Id)
	}
	return ids
}

// userDataScopeIDSet returns a set of user IDs from list output.
func userDataScopeIDSet(items []*ListOutputItem) map[int]struct{} {
	result := make(map[int]struct{}, len(items))
	for _, id := range userDataScopeListIDs(items) {
		result[id] = struct{}{}
	}
	return result
}

// modelUserStatusProjection is a tiny user-status test projection.
type modelUserStatusProjection struct {
	Status int `json:"status"`
}

// modelUserUsernameProjection is a tiny username test projection.
type modelUserUsernameProjection struct {
	Username string `json:"username"`
}

// userDataScopeStaticOrgCap provides deterministic organization data for
// user-management data-scope tests.
type userDataScopeStaticOrgCap struct {
	enabled     bool
	userDeptIDs map[int][]int
	deptUserIDs map[int][]int
}

// Enabled reports whether the fake organization capability is enabled.
func (f userDataScopeStaticOrgCap) Enabled(context.Context) bool { return f.enabled }

// ListUserDeptAssignments returns no department projections for list rendering.
func (f userDataScopeStaticOrgCap) ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error) {
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// GetUserIDsByDept returns the configured user IDs for one department.
func (f userDataScopeStaticOrgCap) GetUserIDsByDept(_ context.Context, deptID int) ([]int, error) {
	return append([]int(nil), f.deptUserIDs[deptID]...), nil
}

// GetAllAssignedUserIDs returns an empty assignment collection.
func (f userDataScopeStaticOrgCap) GetAllAssignedUserIDs(context.Context) ([]int, error) {
	return []int{}, nil
}

// GetUserDeptInfo returns an empty department projection.
func (f userDataScopeStaticOrgCap) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 0, "", nil
}

// GetUserDeptName returns an empty department name.
func (f userDataScopeStaticOrgCap) GetUserDeptName(context.Context, int) (string, error) {
	return "", nil
}

// GetUserDeptIDs returns the configured department IDs for one user.
func (f userDataScopeStaticOrgCap) GetUserDeptIDs(_ context.Context, userID int) ([]int, error) {
	return append([]int(nil), f.userDeptIDs[userID]...), nil
}

// ApplyUserDeptScope injects a deterministic department constraint for tests.
func (f userDataScopeStaticOrgCap) ApplyUserDeptScope(_ context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error) {
	deptIDs := f.userDeptIDs[currentUserID]
	if len(deptIDs) == 0 {
		return model, true, nil
	}

	seen := make(map[int]struct{})
	userIDs := make([]int, 0)
	for _, deptID := range deptIDs {
		for _, userID := range f.deptUserIDs[deptID] {
			if _, ok := seen[userID]; ok {
				continue
			}
			seen[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}
	}
	if len(userIDs) == 0 {
		return model, true, nil
	}
	return model.WhereIn(userIDColumn, userIDs), false, nil
}

// BuildUserDeptScopeExists returns no standalone EXISTS query because this fake
// applies deterministic constraints directly.
func (f userDataScopeStaticOrgCap) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	return nil, true, nil
}

// GetUserPostIDs returns no post IDs.
func (f userDataScopeStaticOrgCap) GetUserPostIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ReplaceUserAssignments accepts assignment replacement without doing work.
func (f userDataScopeStaticOrgCap) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments accepts assignment cleanup without doing work.
func (f userDataScopeStaticOrgCap) CleanupUserAssignments(context.Context, int) error {
	return nil
}

// UserDeptTree returns no department tree nodes.
func (f userDataScopeStaticOrgCap) UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error) {
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns no post options.
func (f userDataScopeStaticOrgCap) ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error) {
	return []*orgcap.PostOption{}, nil
}
