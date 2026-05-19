// This file verifies scheduled-job group tenant isolation for list, create,
// update, delete, job-count, and deletion-migration paths.

package jobmgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

const (
	// jobGroupTenantA identifies one isolated tenant boundary for group tests.
	jobGroupTenantA = 930101
	// jobGroupTenantB identifies another isolated tenant boundary for group tests.
	jobGroupTenantB = 930102
)

// TestJobGroupListCreateAndCountsAreTenantScoped verifies tenant users only
// see and count current-tenant scheduled-job groups.
func TestJobGroupListCreateAndCountsAreTenantScoped(t *testing.T) {
	var (
		ctx           = context.Background()
		tenantACtx    = jobGroupTenantContext(ctx, jobGroupTenantA)
		tenantBCtx    = jobGroupTenantContext(ctx, jobGroupTenantB)
		sharedCode    = uniqueTestName("tenant-group-shared")
		tenantAOnly   = uniqueTestName("tenant-group-a-only")
		tenantAName   = uniqueTestName("tenant-a-group")
		tenantBName   = uniqueTestName("tenant-b-group")
		platformID    = insertJobGroupTenantFixture(t, ctx, 0, uniqueTestName("tenant-platform"), uniqueTestName("platform"), false)
		tenantAID     = insertJobGroupTenantFixture(t, ctx, jobGroupTenantA, sharedCode, tenantAName, false)
		tenantAOnlyID = insertJobGroupTenantFixture(
			t,
			ctx,
			jobGroupTenantA,
			tenantAOnly,
			uniqueTestName("tenant-a-only-group"),
			false,
		)
		tenantBID  = insertJobGroupTenantFixture(t, ctx, jobGroupTenantB, sharedCode, tenantBName, false)
		tenantAJob = insertJobTenantFixture(t, ctx, jobGroupTenantA, tenantAID, uniqueTestName("tenant-a-job"))
		tenantBJob = insertJobTenantFixture(t, ctx, jobGroupTenantB, tenantAID, uniqueTestName("tenant-b-job"))
	)
	t.Cleanup(func() {
		cleanupJobHard(t, ctx, tenantAJob)
		cleanupJobHard(t, ctx, tenantBJob)
		cleanupGroupHard(t, ctx, platformID)
		cleanupGroupHard(t, ctx, tenantAID)
		cleanupGroupHard(t, ctx, tenantAOnlyID)
		cleanupGroupHard(t, ctx, tenantBID)
	})

	svc := newTestService(t)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{TenantId: jobGroupTenantA, UserId: 1}})

	out, err := svc.ListGroups(tenantACtx, ListGroupsInput{PageNum: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("expected tenant group list to succeed, got error: %v", err)
	}
	groups := groupListByID(out.List)
	if _, ok := groups[tenantAID]; !ok {
		t.Fatalf("expected tenant A group %d in list ids=%v", tenantAID, groups)
	}
	if _, ok := groups[tenantBID]; ok {
		t.Fatalf("did not expect tenant B group %d in tenant A list ids=%v", tenantBID, groups)
	}
	if _, ok := groups[platformID]; ok {
		t.Fatalf("did not expect platform group %d in tenant A list ids=%v", platformID, groups)
	}
	if groups[tenantAID].JobCount != 1 {
		t.Fatalf("expected tenant A job count=1, got %d", groups[tenantAID].JobCount)
	}

	createdID, err := svc.CreateGroup(tenantACtx, SaveGroupInput{
		Code: sharedCode,
		Name: uniqueTestName("tenant-a-created"),
	})
	if !bizerr.Is(err, CodeJobGroupCodeExists) {
		t.Fatalf("expected same-tenant duplicate code to be rejected, got id=%d err=%v", createdID, err)
	}

	createdID, err = svc.CreateGroup(tenantBCtx, SaveGroupInput{
		Code: tenantAOnly,
		Name: uniqueTestName("tenant-b-created"),
	})
	if err != nil {
		t.Fatalf("expected tenant B group create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupGroupHard(t, ctx, createdID) })
	assertJobGroupTenant(t, ctx, createdID, jobGroupTenantB)
}

// TestJobGroupUpdateRejectsOutOfTenantScope verifies updates cannot target
// another tenant's group and same-code validation stays inside the tenant.
func TestJobGroupUpdateRejectsOutOfTenantScope(t *testing.T) {
	var (
		ctx        = context.Background()
		tenantACtx = jobGroupTenantContext(ctx, jobGroupTenantA)
		tenantBID  = insertJobGroupTenantFixture(t, ctx, jobGroupTenantB, uniqueTestName("tenant-b-update"), uniqueTestName("tenant-b"), false)
	)
	t.Cleanup(func() { cleanupGroupHard(t, ctx, tenantBID) })

	svc := newTestService(t)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{TenantId: jobGroupTenantA, UserId: 1}})

	err := svc.UpdateGroup(tenantACtx, UpdateGroupInput{
		ID: tenantBID,
		SaveGroupInput: SaveGroupInput{
			Code: uniqueTestName("tenant-a-update"),
			Name: uniqueTestName("tenant-a-update-name"),
		},
	})
	if !bizerr.Is(err, CodeJobGroupNotFound) {
		t.Fatalf("expected out-of-tenant update to return not found, got %v", err)
	}

	var group *entity.SysJobGroup
	if err = dao.SysJobGroup.Ctx(ctx).Where(do.SysJobGroup{Id: tenantBID}).Scan(&group); err != nil {
		t.Fatalf("expected tenant B group query to succeed, got error: %v", err)
	}
	if group == nil || group.TenantId != jobGroupTenantB {
		t.Fatalf("expected tenant B group to remain unchanged, got %#v", group)
	}
}

// TestJobGroupDeleteMigratesOnlyCurrentTenantJobs verifies delete migration
// cannot move jobs outside the current tenant boundary.
func TestJobGroupDeleteMigratesOnlyCurrentTenantJobs(t *testing.T) {
	var (
		ctx            = context.Background()
		tenantACtx     = jobGroupTenantContext(ctx, jobGroupTenantA)
		tenantADefault = insertJobGroupTenantFixture(t, ctx, jobGroupTenantA, uniqueTestName("tenant-a-default"), uniqueTestName("tenant-a-default"), true)
		tenantBDefault = insertJobGroupTenantFixture(t, ctx, jobGroupTenantB, uniqueTestName("tenant-b-default"), uniqueTestName("tenant-b-default"), true)
		tenantAGroup   = insertJobGroupTenantFixture(t, ctx, jobGroupTenantA, uniqueTestName("tenant-a-delete"), uniqueTestName("tenant-a-delete"), false)
		tenantBGroup   = insertJobGroupTenantFixture(t, ctx, jobGroupTenantB, uniqueTestName("tenant-b-delete"), uniqueTestName("tenant-b-delete"), false)
		tenantAJob     = insertJobTenantFixture(t, ctx, jobGroupTenantA, tenantAGroup, uniqueTestName("tenant-a-delete-job"))
		tenantBJob     = insertJobTenantFixture(t, ctx, jobGroupTenantB, tenantAGroup, uniqueTestName("tenant-b-same-group-id-job"))
	)
	t.Cleanup(func() {
		cleanupJobHard(t, ctx, tenantAJob)
		cleanupJobHard(t, ctx, tenantBJob)
		cleanupGroupHard(t, ctx, tenantADefault)
		cleanupGroupHard(t, ctx, tenantBDefault)
		cleanupGroupHard(t, ctx, tenantAGroup)
		cleanupGroupHard(t, ctx, tenantBGroup)
	})

	svc := newTestService(t)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{TenantId: jobGroupTenantA, UserId: 1}})

	if err := svc.DeleteGroups(tenantACtx, gconv.String(tenantBGroup)); !bizerr.Is(err, CodeJobGroupNotFound) {
		t.Fatalf("expected out-of-tenant delete to return not found, got %v", err)
	}

	if err := svc.DeleteGroups(tenantACtx, gconv.String(tenantAGroup)); err != nil {
		t.Fatalf("expected tenant A group delete to succeed, got error: %v", err)
	}

	assertJobGroupID(t, ctx, tenantAJob, tenantADefault)
	assertJobGroupID(t, ctx, tenantBJob, tenantAGroup)
	assertJobGroupTenant(t, ctx, tenantBGroup, jobGroupTenantB)
}

// TestCreateJobWritesCurrentTenant verifies UI-created jobs inherit the
// current tenant so group counts and job lists use the same tenant boundary.
func TestCreateJobWritesCurrentTenant(t *testing.T) {
	var (
		ctx        = context.Background()
		tenantACtx = jobGroupTenantContext(ctx, jobGroupTenantA)
		groupID    = insertJobGroupTenantFixture(t, ctx, jobGroupTenantA, uniqueTestName("tenant-a-create-job"), uniqueTestName("tenant-a-create-job"), false)
	)
	t.Cleanup(func() { cleanupGroupHard(t, ctx, groupID) })

	svc := newTestService(t)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{TenantId: jobGroupTenantA, UserId: 1}})

	jobID, err := svc.CreateJob(tenantACtx, SaveJobInput{
		GroupID:        groupID,
		Name:           uniqueTestName("tenant-created-job"),
		TaskType:       jobmeta.TaskTypeShell,
		Timeout:        jobGroupTestTimeout,
		ShellCmd:       "printf 'tenant'",
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusDisabled,
	})
	if err != nil {
		t.Fatalf("expected tenant job create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	var job *entity.SysJob
	if err = dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&job); err != nil {
		t.Fatalf("expected created job query to succeed, got error: %v", err)
	}
	if job == nil || job.TenantId != jobGroupTenantA {
		t.Fatalf("expected created job tenant_id=%d, got %#v", jobGroupTenantA, job)
	}
}

// jobGroupTestTimeout keeps tenant job creation fixtures aligned with service validation.
const jobGroupTestTimeout = 30 * time.Second

// jobGroupTenantContext injects one tenant for service-layer tenant scoping.
func jobGroupTenantContext(ctx context.Context, tenantID int) context.Context {
	return datascope.WithTenantForTest(ctx, tenantID)
}

// insertJobGroupTenantFixture inserts one tenant-owned group fixture.
func insertJobGroupTenantFixture(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	code string,
	name string,
	isDefault bool,
) int64 {
	t.Helper()

	defaultValue := 0
	if isDefault {
		defaultValue = 1
	}
	insertID, err := dao.SysJobGroup.Ctx(ctx).Data(do.SysJobGroup{
		TenantId:  tenantID,
		Code:      code,
		Name:      name,
		Remark:    fmt.Sprintf("tenant group fixture %d", tenantID),
		SortOrder: 99,
		IsDefault: defaultValue,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected job group fixture insert to succeed, got error: %v", err)
	}
	return int64(insertID)
}

// insertJobTenantFixture inserts one tenant-owned scheduled-job fixture.
func insertJobTenantFixture(t *testing.T, ctx context.Context, tenantID int, groupID int64, name string) int64 {
	t.Helper()

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		TenantId:       tenantID,
		GroupId:        groupID,
		Name:           name,
		Description:    "Tenant job-group isolation fixture.",
		TaskType:       string(jobmeta.TaskTypeShell),
		HandlerRef:     "",
		Params:         `{}`,
		TimeoutSeconds: 30,
		ShellCmd:       "printf 'tenant'",
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusDisabled),
		IsBuiltin:      0,
		SeedVersion:    0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected tenant job fixture insert to succeed, got error: %v", err)
	}
	return int64(insertID)
}

// groupListByID indexes group list rows by group ID.
func groupListByID(items []*GroupListItem) map[int64]*GroupListItem {
	result := make(map[int64]*GroupListItem, len(items))
	for _, item := range items {
		if item == nil || item.SysJobGroup == nil {
			continue
		}
		result[item.SysJobGroup.Id] = item
	}
	return result
}

// assertJobGroupTenant verifies one group still belongs to the expected tenant.
func assertJobGroupTenant(t *testing.T, ctx context.Context, groupID int64, tenantID int) {
	t.Helper()

	var group *entity.SysJobGroup
	if err := dao.SysJobGroup.Ctx(ctx).Where(do.SysJobGroup{Id: groupID}).Scan(&group); err != nil {
		t.Fatalf("expected job group query to succeed, got error: %v", err)
	}
	if group == nil {
		t.Fatalf("expected job group %d to exist", groupID)
	}
	if group.TenantId != tenantID {
		t.Fatalf("expected job group %d tenant_id=%d, got %d", groupID, tenantID, group.TenantId)
	}
}

// assertJobGroupID verifies one job's owning group.
func assertJobGroupID(t *testing.T, ctx context.Context, jobID int64, groupID int64) {
	t.Helper()

	var job *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&job); err != nil {
		t.Fatalf("expected job query to succeed, got error: %v", err)
	}
	if job == nil {
		t.Fatalf("expected job %d to exist", jobID)
	}
	if job.GroupId != groupID {
		t.Fatalf("expected job %d group_id=%d, got %d", jobID, groupID, job.GroupId)
	}
}
