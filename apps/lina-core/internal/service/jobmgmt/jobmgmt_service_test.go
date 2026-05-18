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
	"lina-core/pkg/bizerr"
)

// assertBusinessCode verifies that err carries the expected structured
// business error code.
func assertBusinessCode(t *testing.T, err error, code *bizerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatal("expected structured business error")
	}
	actual, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %v", err)
	}
	if !actual.Matches(code) {
		t.Fatalf("expected business code %s, got %s", code.RuntimeCode(), actual.RuntimeCode())
	}
}

// TestDeleteGroupsMigratesJobsToDefault verifies non-default group deletion migrates jobs to the default group.
func TestDeleteGroupsMigratesJobsToDefault(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
		groupID   int64
		jobID     int64
		groupCode = uniqueTestName("test-job-group")
		groupName = uniqueTestName("测试任务分组")
		jobName   = uniqueTestName("测试任务")
	)

	groupID, err := svc.CreateGroup(ctx, SaveGroupInput{
		Code: groupCode,
		Name: groupName,
	})
	if err != nil {
		t.Fatalf("expected group create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupGroupHard(t, ctx, groupID) })

	insertedJobID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        groupID,
		Name:           jobName,
		Description:    "Temporary job used to verify group deletion migration.",
		TaskType:       jobmeta.TaskTypeShell,
		TimeoutSeconds: int64((5 * time.Minute).Seconds()),
		ShellCmd:       "printf 'group-migration'",
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		ExecutedCount:  0,
		StopReason:     "",
		Status:         jobmeta.JobStatusDisabled,
		IsBuiltin:      0,
		SeedVersion:    0,
		CreatedBy:      0,
		UpdatedBy:      0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected job fixture insert to succeed, got error: %v", err)
	}
	jobID = int64(insertedJobID)
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

	jobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("builtin-locked"),
		Description:    "Temporary built-in used to verify immutable field protection.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     uniqueTestName("host:builtin-locked"),
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "# 17 3 * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobmeta.JobStatusEnabled,
	})
	defer cleanupJobHard(t, ctx, jobID)

	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: jobID}).
		Scan(&job); err != nil {
		t.Fatalf("expected built-in job query to succeed, got error: %v", err)
	}
	if job == nil {
		t.Fatal("expected synced built-in job to exist")
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
		t.Fatal("expected built-in job update to be rejected")
	}
	assertBusinessCode(t, err, CodeJobBuiltinUpdateDenied)
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
		TaskType:       jobmeta.TaskTypeShell,
		Timeout:        0,
		ShellCmd:       "printf 'timeout'",
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
		TaskType:       jobmeta.TaskTypeShell,
		Timeout:        5 * time.Minute,
		ShellCmd:       "printf 'concurrency'",
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

// TestCreateJobRejectsInvalidCronAndManagedStatus verifies save-time validation
// rejects unsupported cron formats, managed status values, and invalid runtime fields.
func TestCreateJobRejectsInvalidCronAndManagedStatus(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
	)

	testCases := []struct {
		name     string
		input    SaveJobInput
		wantCode *bizerr.Code
	}{
		{
			name: "unsupported cron field count",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-cron-count"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'cron-count'",
				CronExpr:       "* * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobmeta.JobStatusDisabled,
			},
			wantCode: CodeJobCronFieldCountInvalid,
		},
		{
			name: "manual hash seconds placeholder",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-cron-hash"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'cron-hash'",
				CronExpr:       "# 17 3 * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobmeta.JobStatusDisabled,
			},
			wantCode: CodeJobCronSecondsRequired,
		},
		{
			name: "timezone must be valid",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-timezone"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'timezone'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Mars/Phobos",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobmeta.JobStatusDisabled,
			},
			wantCode: CodeJobTimezoneInvalid,
		},
		{
			name: "status is system managed",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-status"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'status'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobmeta.JobStatusPausedByPlugin,
			},
			wantCode: CodeJobStatusInvalid,
		},
		{
			name: "timeout must use whole seconds",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-timeout-seconds"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        1500 * time.Millisecond,
				ShellCmd:       "printf 'timeout-seconds'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobmeta.JobStatusDisabled,
			},
			wantCode: CodeJobTimeoutSecondAlignedRequired,
		},
		{
			name: "parallel max concurrency upper bound",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-max-concurrency"),
				TaskType:       jobmeta.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'max-concurrency'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeMasterOnly,
				Concurrency:    jobmeta.JobConcurrencyParallel,
				MaxConcurrency: 101,
				Status:         jobmeta.JobStatusDisabled,
			},
			wantCode: CodeJobMaxConcurrencyInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateJob(ctx, tc.input)
			if err == nil {
				t.Fatalf("expected CreateJob to reject %s", tc.name)
			}
			assertBusinessCode(t, err, tc.wantCode)
		})
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

// TestPreviewCronRejectsInvalidFormats verifies preview shares the strict cron validation rules.
func TestPreviewCronRejectsInvalidFormats(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
	)

	testCases := []struct {
		expr     string
		timezone string
		wantCode *bizerr.Code
	}{
		{
			expr:     "* * * *",
			timezone: "UTC",
			wantCode: CodeJobCronFieldCountInvalid,
		},
		{
			expr:     "# 17 3 * * *",
			timezone: "UTC",
			wantCode: CodeJobCronSecondsRequired,
		},
		{
			expr:     "17 3 * * *",
			timezone: "Invalid/Timezone",
			wantCode: CodeJobTimezoneInvalid,
		},
	}

	for _, tc := range testCases {
		_, err := svc.PreviewCron(ctx, tc.expr, tc.timezone)
		if err == nil {
			t.Fatalf("expected PreviewCron(%q, %q) to fail", tc.expr, tc.timezone)
		}
		assertBusinessCode(t, err, tc.wantCode)
	}
}
