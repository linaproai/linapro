// This file verifies scheduled-job execution-log cleanup supports both
// full-table cleanup and selected-row batch deletion.

package jobmgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
)

// insertLogCleanupTestJob creates one disabled handler job for execution-log tests.
func insertLogCleanupTestJob(t *testing.T, ctx context.Context) uint64 {
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
	return uint64(insertID)
}

// insertLogCleanupTestLog creates one persisted execution log for cleanup tests.
func insertLogCleanupTestLog(t *testing.T, ctx context.Context, jobID uint64, suffix string) uint64 {
	t.Helper()

	startAt := gtime.NewFromTime(time.Now())
	insertID, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:          jobID,
		JobSnapshot:    fmt.Sprintf(`{"name":"%s"}`, suffix),
		NodeId:         "test-node",
		Trigger:        string(jobmeta.TriggerTypeManual),
		ParamsSnapshot: `{}`,
		StartAt:        startAt,
		EndAt:          startAt,
		DurationMs:     1,
		Status:         string(jobmeta.LogStatusSuccess),
		ErrMsg:         "",
		ResultJson:     `{}`,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected execution-log insert to succeed, got error: %v", err)
	}
	return uint64(insertID)
}

// listJobLogIDs returns the remaining execution-log IDs for assertions.
func listJobLogIDs(t *testing.T, ctx context.Context, jobID uint64) []uint64 {
	t.Helper()

	var rows []*entity.SysJobLog
	if err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{JobId: jobID}).
		OrderAsc(dao.SysJobLog.Columns().Id).
		Scan(&rows); err != nil {
		t.Fatalf("expected execution-log query to succeed, got error: %v", err)
	}

	result := make([]uint64, 0, len(rows))
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
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	const rollbackMessage = "rollback execution-log cleanup test transaction"
	err := dao.SysJobLog.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		firstLogID := insertLogCleanupTestLog(t, ctx, jobID, "first")
		secondLogID := insertLogCleanupTestLog(t, ctx, jobID, "second")

		if err := svc.ClearLogs(ctx, nil, fmt.Sprintf("%d", firstLogID)); err != nil {
			t.Fatalf("expected selected execution-log delete to succeed, got error: %v", err)
		}

		remainingIDs := listJobLogIDs(t, ctx, jobID)
		if len(remainingIDs) != 1 || remainingIDs[0] != secondLogID {
			t.Fatalf("expected only second log to remain after selected delete, got %#v", remainingIDs)
		}

		if err := svc.ClearLogs(ctx, nil, ""); err != nil {
			t.Fatalf("expected full execution-log cleanup to succeed, got error: %v", err)
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
