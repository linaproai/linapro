// This file verifies scheduled-job execution-log cleanup supports both
// full-table cleanup and selected-row batch deletion.

package jobmgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/jobmeta"
)

// jobLogCleanupConfigStub overrides log-retention settings while inheriting
// the rest of the config contract from a real host config service.
type jobLogCleanupConfigStub struct {
	hostconfig.Service
	logRetentionDays int64
	cronRetention    *hostconfig.CronLogRetentionConfig
}

// GetLogRetentionDays returns the global maximum log retention period for tests.
func (s jobLogCleanupConfigStub) GetLogRetentionDays(context.Context) (int64, error) {
	return s.logRetentionDays, nil
}

// GetCronLogRetention returns the default cron-log retention policy for tests.
func (s jobLogCleanupConfigStub) GetCronLogRetention(context.Context) (*hostconfig.CronLogRetentionConfig, error) {
	if s.cronRetention != nil {
		return s.cronRetention, nil
	}
	return &hostconfig.CronLogRetentionConfig{
		Mode:  hostconfig.CronLogRetentionModeNone,
		Value: 0,
	}, nil
}

// insertLogCleanupTestJob creates one disabled handler job for execution-log tests.
func insertLogCleanupTestJob(t *testing.T, ctx context.Context) int64 {
	t.Helper()

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("job-log-cleanup"),
		TaskType:       string(jobmeta.TaskTypeHandler),
		HandlerRef:     "host:cleanup-job-logs",
		Params:         `{}`,
		TimeoutSeconds: 30,
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusDisabled),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected execution-log test job insert to succeed, got error: %v", err)
	}
	return int64(insertID)
}

// insertLogCleanupTestLog creates one persisted execution log for cleanup tests.
func insertLogCleanupTestLog(t *testing.T, ctx context.Context, jobID int64, suffix string) int64 {
	t.Helper()

	return insertLogCleanupTestLogAt(t, ctx, jobID, suffix, time.Now())
}

// insertLogCleanupTestLogAt creates one persisted execution log at a controlled
// start time for retention-boundary tests.
func insertLogCleanupTestLogAt(
	t *testing.T,
	ctx context.Context,
	jobID int64,
	suffix string,
	startAt time.Time,
) int64 {
	t.Helper()

	insertID, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:          jobID,
		JobSnapshot:    fmt.Sprintf(`{"name":"%s"}`, suffix),
		NodeId:         "test-node",
		Trigger:        string(jobmeta.TriggerTypeManual),
		ParamsSnapshot: `{}`,
		StartAt:        &startAt,
		EndAt:          &startAt,
		DurationMs:     1,
		Status:         string(jobmeta.LogStatusSuccess),
		ErrMsg:         "",
		ResultJson:     `{}`,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected execution-log insert to succeed, got error: %v", err)
	}
	return int64(insertID)
}

// listJobLogIDs returns the remaining execution-log IDs for assertions.
func listJobLogIDs(t *testing.T, ctx context.Context, jobID int64) []int64 {
	t.Helper()

	var rows []*entity.SysJobLog
	if err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{JobId: jobID}).
		OrderAsc(dao.SysJobLog.Columns().Id).
		Scan(&rows); err != nil {
		t.Fatalf("expected execution-log query to succeed, got error: %v", err)
	}

	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		result = append(result, row.Id)
	}
	return result
}

// TestClearLogsSupportsDeleteAllAndSelectedIDs verifies full cleanup no longer
// trips the ORM delete guard and selected rows can be batch-deleted.
func TestClearLogsSupportsDeleteAllAndSelectedIDs(t *testing.T) {
	var (
		ctx   = context.Background()
		svc   = newTestService(t)
		jobID = insertLogCleanupTestJob(t, ctx)
	)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{UserId: 1}})
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	const rollbackMessage = "rollback execution-log cleanup test transaction"
	err := dao.SysJobLog.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		firstLogID := insertLogCleanupTestLog(t, ctx, jobID, "first")
		secondLogID := insertLogCleanupTestLog(t, ctx, jobID, "second")

		deleted, err := svc.ClearLogs(ctx, ClearLogsInput{IDs: fmt.Sprintf("%d", firstLogID)})
		if err != nil {
			t.Fatalf("expected selected execution-log delete to succeed, got error: %v", err)
		}
		if deleted != 1 {
			t.Fatalf("expected selected execution-log delete count 1, got %d", deleted)
		}

		remainingIDs := listJobLogIDs(t, ctx, jobID)
		if len(remainingIDs) != 1 || remainingIDs[0] != secondLogID {
			t.Fatalf("expected only second log to remain after selected delete, got %#v", remainingIDs)
		}

		rangeLogID := insertLogCleanupTestLog(t, ctx, jobID, "range")
		deleted, err = svc.ClearLogs(ctx, ClearLogsInput{
			JobID:     &jobID,
			BeginTime: "2000-01-01",
			EndTime:   time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		})
		if err != nil {
			t.Fatalf("expected ranged execution-log cleanup to succeed, got error: %v", err)
		}
		if deleted != 2 {
			t.Fatalf("expected ranged execution-log cleanup count 2, got %d", deleted)
		}

		if remainingIDs = listJobLogIDs(t, ctx, jobID); len(remainingIDs) != 0 {
			t.Fatalf("expected ranged cleanup to remove second and range logs including %d, got %#v", rangeLogID, remainingIDs)
		}

		insertLogCleanupTestLog(t, ctx, jobID, "full")
		deleted, err = svc.ClearLogs(ctx, ClearLogsInput{})
		if err != nil {
			t.Fatalf("expected full execution-log cleanup to succeed, got error: %v", err)
		}
		if deleted < 1 {
			t.Fatalf("expected full execution-log cleanup to delete at least one row, got %d", deleted)
		}

		if remainingIDs = listJobLogIDs(t, ctx, jobID); len(remainingIDs) != 0 {
			t.Fatalf("expected all execution logs to be removed after full cleanup, got %#v", remainingIDs)
		}
		return gerror.New(rollbackMessage)
	})
	if err == nil || err.Error() != rollbackMessage {
		t.Fatalf("expected transaction rollback marker %q, got %v", rollbackMessage, err)
	}
}

// TestCleanupDueLogsAppliesGlobalRetentionBeforeJobPolicy verifies the
// system-wide maximum retention deletes old execution logs even when the
// default cron-specific policy is disabled.
func TestCleanupDueLogsAppliesGlobalRetentionBeforeJobPolicy(t *testing.T) {
	var (
		ctx   = context.Background()
		svc   = newTestService(t)
		jobID = insertLogCleanupTestJob(t, ctx)
	)
	setJobMgmtTestBizCtx(svc, jobmgmtStaticBizCtx{ctx: &model.Context{UserId: 1}})
	svc.configSvc = jobLogCleanupConfigStub{
		Service:          hostconfig.New(),
		logRetentionDays: 7,
		cronRetention: &hostconfig.CronLogRetentionConfig{
			Mode:  hostconfig.CronLogRetentionModeNone,
			Value: 0,
		},
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	const rollbackMessage = "rollback global execution-log retention test transaction"
	err := dao.SysJobLog.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		oldLogID := insertLogCleanupTestLogAt(t, ctx, jobID, "old", time.Now().AddDate(0, 0, -10))
		newLogID := insertLogCleanupTestLogAt(t, ctx, jobID, "new", time.Now().AddDate(0, 0, -3))

		deleted, err := svc.CleanupDueLogs(ctx)
		if err != nil {
			t.Fatalf("expected cleanup due logs to succeed, got error: %v", err)
		}
		if deleted < 1 {
			t.Fatalf("expected at least the old fixture log to be deleted, got %d", deleted)
		}

		remainingIDs := listJobLogIDs(t, ctx, jobID)
		if len(remainingIDs) != 1 || remainingIDs[0] != newLogID {
			t.Fatalf("expected only new log %d to remain after global cleanup, old=%d got %#v", newLogID, oldLogID, remainingIDs)
		}
		return gerror.New(rollbackMessage)
	})
	if err == nil || err.Error() != rollbackMessage {
		t.Fatalf("expected transaction rollback marker %q, got %v", rollbackMessage, err)
	}
}
