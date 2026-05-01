// This file covers built-in scheduled-job reconciliation for renamed and removed definitions.

package jobmgmt

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
)

// snapshotExistingBuiltinDefs converts current persisted built-ins into
// reconciliation inputs so tests can preserve global seeds while exercising
// pruning logic for their own temporary rows.
func snapshotExistingBuiltinDefs(t *testing.T, ctx context.Context) []BuiltinJobDef {
	t.Helper()

	var jobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1}).
		Scan(&jobs); err != nil {
		t.Fatalf("expected builtin job snapshot query to succeed, got error: %v", err)
	}

	result := make([]BuiltinJobDef, 0, len(jobs))
	for _, job := range jobs {
		if job == nil || job.Id == 0 {
			continue
		}

		var group *entity.SysJobGroup
		if err := dao.SysJobGroup.Ctx(ctx).
			Where(do.SysJobGroup{Id: job.GroupId}).
			Scan(&group); err != nil {
			t.Fatalf("expected builtin job group query to succeed, got error: %v", err)
		}
		if group == nil {
			t.Fatalf("expected builtin job %d group to exist", job.Id)
		}

		result = append(result, BuiltinJobDef{
			GroupCode:       group.Code,
			Name:            job.Name,
			Description:     job.Description,
			TaskType:        jobmeta.NormalizeTaskType(job.TaskType),
			HandlerRef:      job.HandlerRef,
			Params:          decodeJobParams(job.Params),
			Timeout:         time.Duration(job.TimeoutSeconds) * time.Second,
			Pattern:         job.CronExpr,
			Timezone:        job.Timezone,
			Scope:           jobmeta.NormalizeJobScope(job.Scope),
			Concurrency:     jobmeta.NormalizeJobConcurrency(job.Concurrency),
			MaxConcurrency:  job.MaxConcurrency,
			MaxExecutions:   job.MaxExecutions,
			Status:          jobmeta.NormalizeJobStatus(job.Status),
			LogRetentionRaw: job.LogRetentionOverride,
		})
	}
	return result
}

// TestSyncBuiltinJobsReusesBuiltinRowByGroupAndName verifies renamed handler
// refs reuse the existing code-owned row when the stable group/name identity is unchanged.
func TestSyncBuiltinJobsReusesBuiltinRowByGroupAndName(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	def := BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("builtin-same-name"),
		Description:    "Legacy handler ref used to verify row reuse by group/name.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:legacy-builtin-row-reuse",
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "@every 1m",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusEnabled,
	}
	jobID := syncBuiltinHandlerJob(t, ctx, svc, def)
	defer cleanupJobHard(t, ctx, jobID)

	def.HandlerRef = "plugin:monitor-server/cron:" + def.Name
	if _, err := svc.SyncBuiltinJobs(ctx, []BuiltinJobDef{def}); err != nil {
		t.Fatalf("expected builtin job rename sync to succeed, got error: %v", err)
	}

	var current *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: jobID}).
		Scan(&current); err != nil {
		t.Fatalf("expected current builtin job query to succeed, got error: %v", err)
	}
	if current == nil {
		t.Fatal("expected reused builtin row to remain present")
	}
	if current.HandlerRef != def.HandlerRef {
		t.Fatalf("expected handler ref to update to %s, got %s", def.HandlerRef, current.HandlerRef)
	}
	if got := jobmeta.NormalizeJobStatus(current.Status); got != jobmeta.JobStatusPausedByPlugin {
		t.Fatalf("expected missing plugin handler to downgrade to paused_by_plugin, got %s", got)
	}

	count, err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{GroupId: current.GroupId, Name: def.Name}).
		Count()
	if err != nil {
		t.Fatalf("expected builtin job count query to succeed, got error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one builtin row for reused group/name identity, got %d", count)
	}
}

// TestSyncBuiltinJobsPrunesRemovedBuiltins verifies removed code-owned jobs are
// hard-deleted together with their logs before new built-ins are inserted.
func TestSyncBuiltinJobsPrunesRemovedBuiltins(t *testing.T) {
	ctx := context.Background()
	scheduler := &trackingScheduler{}
	svc := newTestServiceWithRegistry(t, jobhandler.New(), scheduler)
	existingDefs := snapshotExistingBuiltinDefs(t, ctx)

	obsolete := BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("builtin-obsolete"),
		Description:    "Obsolete builtin used to verify reconciliation cleanup.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:obsolete-builtin-job",
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "@every 1m",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusEnabled,
	}
	obsoleteJobID := syncBuiltinHandlerJob(t, ctx, svc, obsolete)

	if _, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:      obsoleteJobID,
		Status:     string(jobmeta.LogStatusSuccess),
		Trigger:    string(jobmeta.TriggerTypeCron),
		NodeId:     "test-node",
		DurationMs: 1,
	}).Insert(); err != nil {
		t.Fatalf("expected obsolete builtin log insert to succeed, got error: %v", err)
	}

	current := BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("builtin-current"),
		Description:    "Current builtin kept after reconciliation.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:current-builtin-job",
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "@every 5m",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusEnabled,
	}
	desired := append(existingDefs, current)
	if _, err := svc.ReconcileBuiltinJobs(ctx, desired); err != nil {
		t.Fatalf("expected builtin job reconciliation to succeed, got error: %v", err)
	}

	var currentRow *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1, HandlerRef: current.HandlerRef}).
		Scan(&currentRow); err != nil {
		t.Fatalf("expected current builtin job query to succeed, got error: %v", err)
	}
	if currentRow == nil || currentRow.Id == 0 {
		t.Fatal("expected reconciled current builtin job to exist")
	}
	currentJobID := currentRow.Id
	defer cleanupJobHard(t, ctx, currentJobID)

	if removed := scheduler.removedIDs(); len(removed) == 0 || removed[0] != obsoleteJobID {
		t.Fatalf("expected scheduler to remove obsolete builtin job %d, got %#v", obsoleteJobID, removed)
	}

	var obsoleteRow *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Unscoped().
		Where(do.SysJob{Id: obsoleteJobID}).
		Scan(&obsoleteRow); err != nil {
		t.Fatalf("expected obsolete builtin unscoped query to succeed, got error: %v", err)
	}
	if obsoleteRow != nil {
		t.Fatalf("expected obsolete builtin job %d to be hard-deleted, got %#v", obsoleteJobID, obsoleteRow)
	}

	logCount, err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{JobId: obsoleteJobID}).
		Count()
	if err != nil {
		t.Fatalf("expected obsolete builtin log count query to succeed, got error: %v", err)
	}
	if logCount != 0 {
		t.Fatalf("expected obsolete builtin logs to be deleted, got %d", logCount)
	}
}
