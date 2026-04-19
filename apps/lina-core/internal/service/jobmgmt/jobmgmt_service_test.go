// This file verifies scheduled-job management validation and migration behaviors.

package jobmgmt

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
)

// TestDeleteGroupsMigratesJobsToDefault verifies non-default group deletion migrates jobs to the default group.
func TestDeleteGroupsMigratesJobsToDefault(t *testing.T) {
	var (
		ctx          = context.Background()
		svc          = newTestService(t)
		defaultID    = defaultGroupID(t, ctx)
		groupID      uint64
		jobID        uint64
		groupCode    = uniqueTestName("test-job-group")
		groupName    = uniqueTestName("测试任务分组")
		jobName      = uniqueTestName("测试任务")
	)

	groupID, err := svc.CreateGroup(ctx, SaveGroupInput{
		Code: groupCode,
		Name: groupName,
	})
	if err != nil {
		t.Fatalf("expected group create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupGroupHard(t, ctx, groupID) })

	jobID, err = svc.CreateJob(ctx, SaveJobInput{
		GroupID:        groupID,
		Name:           jobName,
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:cleanup-job-logs",
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusDisabled,
	})
	if err != nil {
		t.Fatalf("expected job create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	if err = svc.DeleteGroups(ctx, gconv.String(groupID)); err != nil {
		t.Fatalf("expected group delete to succeed, got error: %v", err)
	}

	var jobRow *entity.SysJob
	if err = dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil {
		t.Fatalf("expected migrated job query to succeed, got error: %v", err)
	}
	if jobRow == nil {
		t.Fatal("expected migrated job to remain present after group deletion")
	}
	if jobRow.GroupId != defaultID {
		t.Fatalf("expected migrated job group_id=%d, got %d", defaultID, jobRow.GroupId)
	}
}

// TestUpdateBuiltInJobRejectsLockedFields verifies built-in job immutable fields stay protected.
func TestUpdateBuiltInJobRejectsLockedFields(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
		job *entity.SysJob
	)

	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1, HandlerRef: "host:cleanup-job-logs"}).
		Scan(&job); err != nil {
		t.Fatalf("expected built-in job query to succeed, got error: %v", err)
	}
	if job == nil {
		t.Fatal("expected built-in cleanup-job-logs seed to exist")
	}

	err := svc.UpdateJob(ctx, UpdateJobInput{
		ID: job.Id,
		SaveJobInput: SaveJobInput{
			GroupID:              job.GroupId,
			Name:                 job.Name,
			Description:          job.Description,
			TaskType:             jobmeta.NormalizeTaskType(job.TaskType),
			HandlerRef:           "host:another-handler",
			Params:               decodeJobParams(job.Params),
			Timeout:              time.Duration(job.TimeoutSeconds) * time.Second,
			CronExpr:             job.CronExpr,
			Timezone:             job.Timezone,
			Scope:                jobmeta.NormalizeJobScope(job.Scope),
			Concurrency:          jobmeta.NormalizeJobConcurrency(job.Concurrency),
			MaxConcurrency:       job.MaxConcurrency,
			MaxExecutions:        job.MaxExecutions,
			Status:               jobmeta.NormalizeJobStatus(job.Status),
			LogRetentionOverride: retentionOverrideFromJob(job.LogRetentionOverride),
		},
	})
	if err == nil {
		t.Fatal("expected built-in job update to reject handler_ref mutation")
	}
}

// TestCreateJobValidatesTimeoutAndConcurrency verifies core runtime validation rejects invalid settings.
func TestCreateJobValidatesTimeoutAndConcurrency(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
	)

	_, err := svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultID,
		Name:           uniqueTestName("invalid-timeout"),
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:cleanup-job-logs",
		Params:         map[string]any{},
		Timeout:        0,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusDisabled,
	})
	if err == nil {
		t.Fatal("expected zero timeout to fail validation")
	}

	_, err = svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultID,
		Name:           uniqueTestName("invalid-concurrency"),
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     "host:cleanup-job-logs",
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencyParallel,
		MaxConcurrency: 0,
		Status:         jobmeta.JobStatusDisabled,
	})
	if err == nil {
		t.Fatal("expected zero maxConcurrency to fail validation")
	}
}

// TestPreviewCronSupportsFiveFieldAndTimezone verifies cron preview accepts 5-field expressions and applies the requested timezone.
func TestPreviewCronSupportsFiveFieldAndTimezone(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
	)

	times, err := svc.PreviewCron(ctx, "17 3 * * *", "UTC")
	if err != nil {
		t.Fatalf("expected cron preview to succeed, got error: %v", err)
	}
	if len(times) != 5 {
		t.Fatalf("expected 5 preview times, got %d", len(times))
	}
	for i, item := range times {
		if got := item.Location().String(); got != "UTC" {
			t.Fatalf("expected preview time %d to use UTC, got %s", i, got)
		}
		if item.Minute() != 17 || item.Hour() != 3 || item.Second() != 0 {
			t.Fatalf("expected preview time %d to be 03:17:00 UTC, got %s", i, item.Format(time.RFC3339))
		}
		if i > 0 && !item.After(times[i-1]) {
			t.Fatalf("expected preview times to be strictly increasing, got %s then %s", times[i-1], item)
		}
	}
}
