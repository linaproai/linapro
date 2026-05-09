// This file verifies shared data-scope resolution independent of resource services.

package datascope

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/orgcap"
)

// TestCurrentResolvesWidestScope verifies super-admin, enabled-role merging,
// disabled-role exclusion, no-role denial, and missing-context handling.
func TestCurrentResolvesWidestScope(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name       string
		bizCtx     *model.Context
		roleReader *dataScopeRoleReader
		wantScope  Scope
		wantErr    *bizerr.Code
	}{
		{
			name:   "super admin receives all",
			bizCtx: &model.Context{UserId: 1},
			roleReader: &dataScopeRoleReader{
				snapshots: map[int]*AccessSnapshot{
					1: {UserID: 1, Scope: ScopeAll, IsSuperAdmin: true},
				},
			},
			wantScope: ScopeAll,
		},
		{
			name:   "widest enabled role wins",
			bizCtx: &model.Context{UserId: 2},
			roleReader: &dataScopeRoleReader{
				snapshots: map[int]*AccessSnapshot{
					2: {UserID: 2, Scope: ScopeDept},
				},
			},
			wantScope: ScopeDept,
		},
		{
			name:       "no enabled roles denies",
			bizCtx:     &model.Context{UserId: 3},
			roleReader: &dataScopeRoleReader{},
			wantScope:  ScopeNone,
		},
		{
			name:    "missing context fails",
			wantErr: CodeDataScopeNotAuthenticated,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := New(Dependencies{
				BizCtxSvc: dataScopeStaticBizCtx{ctx: tc.bizCtx},
				RoleSvc:   tc.roleReader,
			})
			actual, err := svc.Current(ctx)
			if tc.wantErr != nil {
				if !bizerr.Is(err, tc.wantErr) {
					t.Fatalf("expected %s, got %v", tc.wantErr.RuntimeCode(), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve current data scope: %v", err)
			}
			if actual == nil || actual.Scope != tc.wantScope {
				t.Fatalf("expected scope %d, got %#v", tc.wantScope, actual)
			}
			if tc.roleReader != nil && tc.roleReader.calls != 1 {
				t.Fatalf("expected one data-scope snapshot read, got %d", tc.roleReader.calls)
			}
		})
	}
}

// TestCurrentRejectsUnsupportedScope verifies invalid enabled role data-scope
// values fail with a structured error.
func TestCurrentRejectsUnsupportedScope(t *testing.T) {
	svc := New(Dependencies{
		BizCtxSvc: dataScopeStaticBizCtx{ctx: &model.Context{UserId: 5}},
		RoleSvc: &dataScopeRoleReader{
			err: bizerr.NewCode(CodeDataScopeUnsupported, bizerr.P("scope", 99)),
		},
	})

	_, err := svc.Current(context.Background())
	if !bizerr.Is(err, CodeDataScopeUnsupported) {
		t.Fatalf("expected unsupported data-scope error, got %v", err)
	}
}

// TestApplyUserScopeUsesOrgCapabilityForDepartment verifies department scope is
// delegated to the optional organization provider instead of materializing
// visible user IDs in the shared data-scope service.
func TestApplyUserScopeUsesOrgCapabilityForDepartment(t *testing.T) {
	ctx := context.Background()
	queryModel := dao.SysUser.Ctx(ctx)
	orgCapSvc := &dataScopeTrackingOrgCap{enabled: true}
	svc := New(Dependencies{
		BizCtxSvc: dataScopeStaticBizCtx{ctx: &model.Context{UserId: 11}},
		RoleSvc: &dataScopeRoleReader{
			snapshots: map[int]*AccessSnapshot{
				11: {UserID: 11, Scope: ScopeDept},
			},
		},
		OrgCapSvc: orgCapSvc,
	})

	if _, _, err := svc.ApplyUserScope(ctx, queryModel, "sys_user.id"); err != nil {
		t.Fatalf("apply department data scope: %v", err)
	}
	if orgCapSvc.applyCalls != 1 || orgCapSvc.applyCurrentUserID != 11 || orgCapSvc.applyUserIDColumn != "sys_user.id" {
		t.Fatalf("expected orgcap ApplyUserDeptScope call, got calls=%d userID=%d column=%q", orgCapSvc.applyCalls, orgCapSvc.applyCurrentUserID, orgCapSvc.applyUserIDColumn)
	}
}

// TestApplyUserScopeFallsBackToSelfWhenOrgDisabled verifies department scope
// safely degrades to current-user filtering when organization capability is not
// enabled.
func TestApplyUserScopeFallsBackToSelfWhenOrgDisabled(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertDataScopeUser(t, ctx, "datascope-current")
	otherUserID := insertDataScopeUser(t, ctx, "datascope-other")
	t.Cleanup(func() { cleanupDataScopeUsers(t, ctx, []int{currentUserID, otherUserID}) })
	queryModel := dao.SysUser.Ctx(ctx).WhereIn(dao.SysUser.Columns().Id, []int{currentUserID, otherUserID})
	orgCapSvc := &dataScopeTrackingOrgCap{enabled: false}
	svc := New(Dependencies{
		BizCtxSvc: dataScopeStaticBizCtx{ctx: &model.Context{UserId: currentUserID}},
		RoleSvc: &dataScopeRoleReader{
			snapshots: map[int]*AccessSnapshot{
				currentUserID: {UserID: currentUserID, Scope: ScopeDept},
			},
		},
		OrgCapSvc: orgCapSvc,
	})

	scopedModel, empty, err := svc.ApplyUserScope(ctx, queryModel, "sys_user.id")
	if err != nil {
		t.Fatalf("apply department fallback data scope: %v", err)
	}
	if empty {
		t.Fatal("expected self fallback to produce a non-empty scoped model")
	}
	if orgCapSvc.applyCalls != 0 {
		t.Fatalf("expected disabled orgcap not to be called, got %d calls", orgCapSvc.applyCalls)
	}
	count, err := scopedModel.Count()
	if err != nil {
		t.Fatalf("count scoped fallback users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected self fallback to include only current user, got count=%d", count)
	}
}

// TestApplyUserScopeWithBypassComposesDeptExists verifies bypass-aware scopes
// compose department filtering through the provider-built EXISTS subquery.
func TestApplyUserScopeWithBypassComposesDeptExists(t *testing.T) {
	ctx := context.Background()
	queryModel := dao.SysJob.Ctx(ctx)
	orgCapSvc := &dataScopeTrackingOrgCap{enabled: true}
	svc := New(Dependencies{
		BizCtxSvc: dataScopeStaticBizCtx{ctx: &model.Context{UserId: 13}},
		RoleSvc: &dataScopeRoleReader{
			snapshots: map[int]*AccessSnapshot{
				13: {UserID: 13, Scope: ScopeDept},
			},
		},
		OrgCapSvc: orgCapSvc,
	})

	if _, _, err := svc.ApplyUserScopeWithBypass(ctx, queryModel, "sys_job.created_by", "sys_job.is_builtin", 1); err != nil {
		t.Fatalf("apply bypass department scope: %v", err)
	}
	if orgCapSvc.existsCalls != 1 || orgCapSvc.existsCurrentUserID != 13 || orgCapSvc.existsUserIDColumn != "sys_job.created_by" {
		t.Fatalf("expected orgcap BuildUserDeptScopeExists call, got calls=%d userID=%d column=%q", orgCapSvc.existsCalls, orgCapSvc.existsCurrentUserID, orgCapSvc.existsUserIDColumn)
	}
}

// dataScopeStaticBizCtx returns a fixed request business context.
type dataScopeStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by data-scope tests.
func (s dataScopeStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s dataScopeStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by data-scope tests.
func (s dataScopeStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by data-scope tests.
func (s dataScopeStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetUserAccess is unused by data-scope tests.
func (s dataScopeStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// dataScopeRoleReader supplies deterministic access snapshots for data-scope tests.
type dataScopeRoleReader struct {
	snapshots map[int]*AccessSnapshot
	err       error
	calls     int
}

// GetUserDataScopeSnapshot returns configured effective role data-scope snapshots.
func (r *dataScopeRoleReader) GetUserDataScopeSnapshot(_ context.Context, userID int) (*AccessSnapshot, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if snapshot := r.snapshots[userID]; snapshot != nil {
		return snapshot, nil
	}
	return &AccessSnapshot{UserID: userID, Scope: ScopeNone}, nil
}

// insertDataScopeUser inserts one temporary user for data-scope integration tests.
func insertDataScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()
	insertID, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()),
		Password: "hashed",
		Nickname: prefix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert data-scope user: %v", err)
	}
	return int(insertID)
}

// cleanupDataScopeUsers removes temporary users for data-scope integration tests.
func cleanupDataScopeUsers(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()
	if len(userIDs) == 0 {
		return
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, userIDs).Delete(); err != nil {
		t.Fatalf("cleanup data-scope users: %v", err)
	}
}

// dataScopeTrackingOrgCap records organization capability calls.
type dataScopeTrackingOrgCap struct {
	enabled             bool
	applyCalls          int
	applyCurrentUserID  int
	applyUserIDColumn   string
	existsCalls         int
	existsCurrentUserID int
	existsUserIDColumn  string
	existsModel         *gdb.Model
}

// Enabled returns the configured organization capability state.
func (f *dataScopeTrackingOrgCap) Enabled(context.Context) bool { return f.enabled }

// ListUserDeptAssignments returns no department projections.
func (f *dataScopeTrackingOrgCap) ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error) {
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// GetUserIDsByDept returns no users.
func (f *dataScopeTrackingOrgCap) GetUserIDsByDept(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// GetAllAssignedUserIDs returns no assigned users.
func (f *dataScopeTrackingOrgCap) GetAllAssignedUserIDs(context.Context) ([]int, error) {
	return []int{}, nil
}

// GetUserDeptInfo returns no department projection.
func (f *dataScopeTrackingOrgCap) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 0, "", nil
}

// GetUserDeptName returns no department name.
func (f *dataScopeTrackingOrgCap) GetUserDeptName(context.Context, int) (string, error) {
	return "", nil
}

// GetUserDeptIDs returns no department IDs.
func (f *dataScopeTrackingOrgCap) GetUserDeptIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ApplyUserDeptScope records the department-scope invocation.
func (f *dataScopeTrackingOrgCap) ApplyUserDeptScope(_ context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error) {
	f.applyCalls++
	f.applyCurrentUserID = currentUserID
	f.applyUserIDColumn = userIDColumn
	return model.Where(userIDColumn, currentUserID), false, nil
}

// BuildUserDeptScopeExists records the EXISTS-builder invocation.
func (f *dataScopeTrackingOrgCap) BuildUserDeptScopeExists(_ context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error) {
	f.existsCalls++
	f.existsCurrentUserID = currentUserID
	f.existsUserIDColumn = userIDColumn
	if f.existsModel == nil {
		return nil, true, nil
	}
	return f.existsModel, false, nil
}

// GetUserPostIDs returns no post IDs.
func (f *dataScopeTrackingOrgCap) GetUserPostIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ReplaceUserAssignments accepts assignment replacement.
func (f *dataScopeTrackingOrgCap) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments accepts assignment cleanup.
func (f *dataScopeTrackingOrgCap) CleanupUserAssignments(context.Context, int) error {
	return nil
}

// UserDeptTree returns no department tree.
func (f *dataScopeTrackingOrgCap) UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error) {
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns no post options.
func (f *dataScopeTrackingOrgCap) ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error) {
	return []*orgcap.PostOption{}, nil
}
