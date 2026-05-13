// This file verifies scheduled-job and log data-scope enforcement.

package jobmgmt

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// TestJobDataScopeFiltersUserJobsKeepsBuiltinsAndProtectsLogs verifies
// user-created jobs are scoped by owner while built-in jobs remain visible.
func TestJobDataScopeFiltersUserJobsKeepsBuiltinsAndProtectsLogs(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertJobScopeUser(t, ctx, "job-scope-current")
	otherUserID := insertJobScopeUser(t, ctx, "job-scope-other")
	roleID := insertJobScopeRole(t, ctx, "job-scope-self", 3)
	t.Cleanup(func() {
		cleanupJobScopeUsers(t, ctx, []int{currentUserID, otherUserID})
		cleanupJobScopeRoles(t, ctx, []int{roleID})
	})
	insertJobScopeUserRole(t, ctx, currentUserID, roleID)

	visibleJobID := insertJobScopeJob(t, ctx, currentUserID, "visible", 0)
	hiddenJobID := insertJobScopeJob(t, ctx, otherUserID, "hidden", 0)
	builtinJobID := insertJobScopeJob(t, ctx, otherUserID, "builtin", 1)
	hiddenLogID := insertJobScopeLog(t, ctx, hiddenJobID, "hidden-log")
	visibleLogID := insertJobScopeLog(t, ctx, visibleJobID, "visible-log")
	t.Cleanup(func() {
		cleanupJobHard(t, ctx, visibleJobID)
		cleanupJobHard(t, ctx, hiddenJobID)
		cleanupJobHard(t, ctx, builtinJobID)
	})

	svc := newTestService(t)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{UserId: currentUserID}})

	out, err := svc.ListJobs(ctx, ListJobsInput{PageNum: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	jobIDs := jobScopeListIDSet(out.List)
	if _, ok := jobIDs[visibleJobID]; !ok {
		t.Fatalf("expected visible job %d in list ids=%v", visibleJobID, jobIDs)
	}
	if _, ok := jobIDs[builtinJobID]; !ok {
		t.Fatalf("expected builtin job %d in list ids=%v", builtinJobID, jobIDs)
	}
	if _, ok := jobIDs[hiddenJobID]; ok {
		t.Fatalf("did not expect hidden job %d in list ids=%v", hiddenJobID, jobIDs)
	}

	keywordOut, err := svc.ListJobs(ctx, ListJobsInput{Keyword: "job-scope", PageNum: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list jobs with keyword: %v", err)
	}
	keywordJobIDs := jobScopeListIDSet(keywordOut.List)
	if _, ok := keywordJobIDs[visibleJobID]; !ok {
		t.Fatalf("expected visible job %d in keyword list ids=%v", visibleJobID, keywordJobIDs)
	}
	if _, ok := keywordJobIDs[builtinJobID]; !ok {
		t.Fatalf("expected builtin job %d in keyword list ids=%v", builtinJobID, keywordJobIDs)
	}
	if _, ok := keywordJobIDs[hiddenJobID]; ok {
		t.Fatalf("did not expect hidden job %d in keyword list ids=%v", hiddenJobID, keywordJobIDs)
	}

	if _, err = svc.GetJob(ctx, hiddenJobID); !bizerr.Is(err, CodeJobDataScopeDenied) {
		t.Fatalf("expected hidden job detail denied, got %v", err)
	}
	if _, err = svc.TriggerJob(ctx, hiddenJobID); !bizerr.Is(err, CodeJobDataScopeDenied) {
		t.Fatalf("expected hidden job trigger denied, got %v", err)
	}
	if err = svc.ClearLogs(ctx, nil, int64ToString(hiddenLogID)); !bizerr.Is(err, CodeJobDataScopeDenied) {
		t.Fatalf("expected hidden log clear denied, got %v", err)
	}

	logs, err := svc.ListLogs(ctx, ListLogsInput{PageNum: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	logIDs := jobScopeLogIDSet(logs.List)
	if _, ok := logIDs[visibleLogID]; !ok {
		t.Fatalf("expected visible log %d in list ids=%v", visibleLogID, logIDs)
	}
	if _, ok := logIDs[hiddenLogID]; ok {
		t.Fatalf("did not expect hidden log %d in list ids=%v", hiddenLogID, logIDs)
	}
}

// insertJobScopeUser inserts one temporary user.
func insertJobScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()

	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: uniqueTestName(prefix),
		Password: "hashed",
		Nickname: prefix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert job-scope user: %v", err)
	}
	return int(id)
}

// insertJobScopeRole inserts one temporary role.
func insertJobScopeRole(t *testing.T, ctx context.Context, prefix string, scope int) int {
	t.Helper()

	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      uniqueTestName(prefix),
		Key:       uniqueTestName(prefix + "-key"),
		Sort:      99,
		DataScope: scope,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert job-scope role: %v", err)
	}
	return int(id)
}

// insertJobScopeUserRole binds one user to one role.
func insertJobScopeUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()
	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{UserId: userID, RoleId: roleID}).Insert(); err != nil {
		t.Fatalf("insert job-scope user role: %v", err)
	}
}

// insertJobScopeJob inserts one temporary scheduled job.
func insertJobScopeJob(t *testing.T, ctx context.Context, ownerID int, suffix string, isBuiltin int) int64 {
	t.Helper()

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("job-scope-" + suffix),
		TaskType:       string(jobmeta.TaskTypeShell),
		HandlerRef:     "",
		Params:         `{}`,
		TimeoutSeconds: 30,
		ShellCmd:       "printf 'scope'",
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusDisabled),
		IsBuiltin:      isBuiltin,
		CreatedBy:      ownerID,
		UpdatedBy:      ownerID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert job-scope job: %v", err)
	}
	return int64(insertID)
}

// insertJobScopeLog inserts one temporary execution log.
func insertJobScopeLog(t *testing.T, ctx context.Context, jobID int64, name string) int64 {
	t.Helper()
	now := gtime.NewFromTime(time.Now())
	insertID, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:          jobID,
		JobSnapshot:    `{"name":"` + name + `"}`,
		NodeId:         "test",
		Trigger:        string(jobmeta.TriggerTypeManual),
		ParamsSnapshot: `{}`,
		StartAt:        now,
		EndAt:          now,
		DurationMs:     1,
		Status:         string(jobmeta.LogStatusSuccess),
		ResultJson:     `{}`,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert job-scope log: %v", err)
	}
	return int64(insertID)
}

// cleanupJobScopeUsers removes temporary users.
func cleanupJobScopeUsers(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, ids).Delete(); err != nil {
		t.Fatalf("cleanup job-scope user roles: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup job-scope users: %v", err)
	}
}

// cleanupJobScopeRoles removes temporary roles.
func cleanupJobScopeRoles(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, ids).Delete(); err != nil {
		t.Fatalf("cleanup job-scope user roles by role: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup job-scope roles: %v", err)
	}
}

// jobScopeListIDSet returns job IDs from list output.
func jobScopeListIDSet(items []*JobListItem) map[int64]struct{} {
	result := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item == nil || item.SysJob == nil {
			continue
		}
		result[item.SysJob.Id] = struct{}{}
	}
	return result
}

// jobScopeLogIDSet returns log IDs from list output.
func jobScopeLogIDSet(items []*LogListItem) map[int64]struct{} {
	result := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item == nil || item.SysJobLog == nil {
			continue
		}
		result[item.SysJobLog.Id] = struct{}{}
	}
	return result
}

// int64ToString formats one ID without importing strconv at call sites.
func int64ToString(id int64) string {
	return gconv.String(id)
}
