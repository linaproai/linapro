// This file verifies persistent scheduler registration, scope guards, and core execution semantics.

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/jobmgmt/internal/shellexec"
)

// fakeClusterService provides deterministic primary-node behavior for scheduler tests.
type fakeClusterService struct {
	enabled bool
	primary bool
	nodeID  string
}

// Start is a no-op for scheduler tests.
func (f fakeClusterService) Start(ctx context.Context) {}

// Stop is a no-op for scheduler tests.
func (f fakeClusterService) Stop(ctx context.Context) {}

// IsEnabled reports the configured cluster enablement state.
func (f fakeClusterService) IsEnabled() bool { return f.enabled }

// IsPrimary reports the configured primary-node state.
func (f fakeClusterService) IsPrimary() bool { return f.primary }

// NodeID returns the configured node identifier.
func (f fakeClusterService) NodeID() string {
	if f.nodeID == "" {
		return "test-node"
	}
	return f.nodeID
}

// schedulerTestCleaner satisfies host-handler registration for tests.
type schedulerTestCleaner struct{}

// CleanupDueLogs is a no-op for scheduler tests.
func (schedulerTestCleaner) CleanupDueLogs(ctx context.Context) (int64, error) { return 0, nil }

// fakeShellExecutor provides deterministic shell-execution behavior for scheduler tests.
type fakeShellExecutor struct {
	execute func(ctx context.Context, in shellexec.ExecuteInput) (*shellexec.ExecuteOutput, error)
}

// Execute delegates to the configured test callback.
func (f fakeShellExecutor) Execute(
	ctx context.Context,
	in shellexec.ExecuteInput,
) (*shellexec.ExecuteOutput, error) {
	return f.execute(ctx, in)
}

// testDefaultGroupID resolves the default job group ID for scheduler tests.
func testDefaultGroupID(t *testing.T, ctx context.Context) uint64 {
	t.Helper()

	var group *entity.SysJobGroup
	if err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{IsDefault: 1}).
		Scan(&group); err != nil {
		t.Fatalf("expected default group query to succeed, got error: %v", err)
	}
	if group == nil {
		t.Fatal("expected default scheduled job group to exist")
	}
	return group.Id
}

// newRegistryWithHandler creates one registry preloaded with the cleanup handler and one test handler.
func newRegistryWithHandler(
	t *testing.T,
	ref string,
	callback func(ctx context.Context, params json.RawMessage) (any, error),
) jobhandler.Registry {
	t.Helper()

	registry := jobhandler.New()
	if err := jobhandler.RegisterHostHandlers(registry, schedulerTestCleaner{}); err != nil {
		t.Fatalf("expected host handler registration to succeed, got error: %v", err)
	}
	if err := registry.Register(jobhandler.HandlerDef{
		Ref:          ref,
		DisplayName:  "Scheduler Test Handler",
		ParamsSchema: `{"type":"object","properties":{}}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke:       callback,
	}); err != nil {
		t.Fatalf("expected test handler registration to succeed, got error: %v", err)
	}
	return registry
}

// registerEnabledHostHandlersAsNoop installs no-op callbacks for any enabled
// host handler refs already persisted in sys_job so startup-load tests do not
// depend on the surrounding database being pristine.
func registerEnabledHostHandlersAsNoop(
	t *testing.T,
	ctx context.Context,
	registry jobhandler.Registry,
) {
	t.Helper()

	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Fields(dao.SysJob.Columns().HandlerRef).
		Where(do.SysJob{Status: string(jobmeta.JobStatusEnabled)}).
		Distinct().
		Scan(&jobs)
	if err != nil {
		t.Fatalf("expected enabled handler query to succeed, got error: %v", err)
	}

	for _, job := range jobs {
		if job == nil {
			continue
		}
		handlerRef := strings.TrimSpace(job.HandlerRef)
		if !strings.HasPrefix(handlerRef, "host:") {
			continue
		}
		if _, exists := registry.Lookup(handlerRef); exists {
			continue
		}
		err = registry.Register(jobhandler.HandlerDef{
			Ref:          handlerRef,
			DisplayName:  handlerRef,
			Description:  "scheduler test no-op host handler",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourceHost,
			Invoke: func(ctx context.Context, params json.RawMessage) (any, error) {
				return nil, nil
			},
		})
		if err != nil {
			t.Fatalf("expected no-op host handler registration to succeed for %s, got error: %v", handlerRef, err)
		}
	}
}

// insertTestJob stores one enabled job row for scheduler tests.
func insertTestJob(
	t *testing.T,
	ctx context.Context,
	handlerRef string,
	scope jobmeta.JobScope,
	concurrency jobmeta.JobConcurrency,
	maxConcurrency int,
	maxExecutions int,
) uint64 {
	t.Helper()

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        testDefaultGroupID(t, ctx),
		Name:           fmt.Sprintf("scheduler-test-%d", time.Now().UnixNano()),
		TaskType:       string(jobmeta.TaskTypeHandler),
		HandlerRef:     handlerRef,
		Params:         `{}`,
		TimeoutSeconds: 30,
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(scope),
		Concurrency:    string(concurrency),
		MaxConcurrency: maxConcurrency,
		MaxExecutions:  maxExecutions,
		Status:         string(jobmeta.JobStatusEnabled),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected scheduler test job insert to succeed, got error: %v", err)
	}
	return uint64(insertID)
}

// cleanupSchedulerJob removes scheduler test jobs, logs, and gcron registrations.
func cleanupSchedulerJob(t *testing.T, ctx context.Context, jobID uint64) {
	t.Helper()
	if jobID == 0 {
		return
	}
	gcron.Remove(jobEntryName(jobID))
	if _, err := dao.SysJobLog.Ctx(ctx).Where(do.SysJobLog{JobId: jobID}).Delete(); err != nil {
		t.Fatalf("expected scheduler test log cleanup to succeed, got error: %v", err)
	}
	if _, err := dao.SysJob.Ctx(ctx).Unscoped().Where(do.SysJob{Id: jobID}).Delete(); err != nil {
		t.Fatalf("expected scheduler test job cleanup to succeed, got error: %v", err)
	}
}

// waitForCondition polls until the provided callback returns true or the timeout expires.
func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("expected condition to become true before timeout")
}

// latestLogStatuses returns all statuses currently stored for the target job.
func latestLogStatuses(t *testing.T, ctx context.Context, jobID uint64) []string {
	t.Helper()
	var logs []*entity.SysJobLog
	if err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{JobId: jobID}).
		OrderAsc(dao.SysJobLog.Columns().Id).
		Scan(&logs); err != nil {
		t.Fatalf("expected scheduler test log query to succeed, got error: %v", err)
	}
	result := make([]string, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil {
			continue
		}
		result = append(result, logRow.Status)
	}
	return result
}

// latestLogs returns all persisted logs for one scheduler test job.
func latestLogs(t *testing.T, ctx context.Context, jobID uint64) []*entity.SysJobLog {
	t.Helper()

	var logs []*entity.SysJobLog
	if err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{JobId: jobID}).
		OrderAsc(dao.SysJobLog.Columns().Id).
		Scan(&logs); err != nil {
		t.Fatalf("expected scheduler test log query to succeed, got error: %v", err)
	}
	return logs
}

// TestNormalizeGcronPatternUsesHashPlaceholder verifies 5-field cron input is
// normalized with GoFrame's `#` seconds placeholder instead of a fixed zero.
func TestNormalizeGcronPatternUsesHashPlaceholder(t *testing.T) {
	pattern, err := normalizeGcronPattern("17 3 * * *")
	if err != nil {
		t.Fatalf("expected 5-field cron normalization to succeed, got error: %v", err)
	}
	if pattern != "# 17 3 * * *" {
		t.Fatalf("expected 5-field cron to normalize to '# 17 3 * * *', got %q", pattern)
	}
}

// TestRefreshRegistersAndRemoveUnregistersJob verifies the scheduler wires persistent jobs into gcron.
func TestRefreshRegistersAndRemoveUnregistersJob(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = newRegistryWithHandler(t, "host:scheduler-register", func(ctx context.Context, params json.RawMessage) (any, error) {
			return nil, nil
		})
		svc   = New(fakeClusterService{primary: true}, registry, nil).(*serviceImpl)
		jobID = insertTestJob(t, ctx, "host:scheduler-register", jobmeta.JobScopeMasterOnly, jobmeta.JobConcurrencySingleton, 1, 0)
	)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	if err := svc.Refresh(ctx, jobID); err != nil {
		t.Fatalf("expected scheduler refresh to succeed, got error: %v", err)
	}
	if entry := gcron.Search(jobEntryName(jobID)); entry == nil {
		t.Fatal("expected gcron entry to exist after refresh")
	}

	svc.Remove(jobID)
	if entry := gcron.Search(jobEntryName(jobID)); entry != nil {
		t.Fatalf("expected gcron entry to be removed, got %#v", entry)
	}
}

// TestRunCronJobSkipsOnNonPrimaryNode verifies master-only jobs emit skipped_not_primary logs on follower nodes.
func TestRunCronJobSkipsOnNonPrimaryNode(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = newRegistryWithHandler(t, "host:scheduler-skip", func(ctx context.Context, params json.RawMessage) (any, error) {
			return nil, nil
		})
		svc   = New(fakeClusterService{enabled: true, primary: false}, registry, nil).(*serviceImpl)
		jobID = insertTestJob(t, ctx, "host:scheduler-skip", jobmeta.JobScopeMasterOnly, jobmeta.JobConcurrencySingleton, 1, 0)
	)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	svc.runCronJob(ctx, jobID)
	waitForCondition(t, 2*time.Second, func() bool {
		statuses := latestLogStatuses(t, ctx, jobID)
		return len(statuses) == 1 && statuses[0] == string(jobmeta.LogStatusSkippedNotPrimary)
	})
}

// TestLoadAndRegisterPausesMissingPluginHandlerJobs verifies startup loading
// downgrades enabled plugin-handler jobs when their handler is unavailable.
func TestLoadAndRegisterPausesMissingPluginHandlerJobs(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = jobhandler.New()
		svc      *serviceImpl
	)
	if err := jobhandler.RegisterHostHandlers(registry, schedulerTestCleaner{}); err != nil {
		t.Fatalf("expected host handler registration to succeed, got error: %v", err)
	}
	registerEnabledHostHandlersAsNoop(t, ctx, registry)
	svc = New(fakeClusterService{primary: true}, registry, fakeShellExecutor{
		execute: func(ctx context.Context, in shellexec.ExecuteInput) (*shellexec.ExecuteOutput, error) {
			return &shellexec.ExecuteOutput{}, nil
		},
	}).(*serviceImpl)

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        testDefaultGroupID(t, ctx),
		Name:           fmt.Sprintf("scheduler-missing-plugin-%d", time.Now().UnixNano()),
		TaskType:       string(jobmeta.TaskTypeHandler),
		HandlerRef:     "plugin:test-missing/cleanup",
		Params:         `{}`,
		TimeoutSeconds: 30,
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusEnabled),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected missing plugin handler job insert to succeed, got error: %v", err)
	}
	jobID := uint64(insertID)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	if err = svc.LoadAndRegister(ctx); err != nil {
		t.Fatalf("expected startup load to downgrade missing plugin handler job, got error: %v", err)
	}

	var jobRow *entity.SysJob
	if err = dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil {
		t.Fatalf("expected downgraded job query to succeed, got error: %v", err)
	}
	if jobRow == nil {
		t.Fatal("expected downgraded job to remain present")
	}
	if got := jobmeta.NormalizeJobStatus(jobRow.Status); got != jobmeta.JobStatusPausedByPlugin {
		t.Fatalf("expected missing plugin handler job status paused_by_plugin, got %s", got)
	}
	if jobRow.StopReason != string(jobmeta.StopReasonPluginUnavailable) {
		t.Fatalf("expected missing plugin handler job stop_reason=%s, got %s", jobmeta.StopReasonPluginUnavailable, jobRow.StopReason)
	}
	if entry := gcron.Search(jobEntryName(jobID)); entry != nil {
		t.Fatalf("expected missing plugin handler job not to register into gcron, got %#v", entry)
	}
}

// TestRunCronJobSingletonSkipsOverlap verifies singleton jobs skip overlapping cron ticks on the same node.
func TestRunCronJobSingletonSkipsOverlap(t *testing.T) {
	var (
		ctx         = context.Background()
		releaseCh   = make(chan struct{})
		releaseOnce sync.Once
		registry    = newRegistryWithHandler(t, "host:scheduler-singleton", func(ctx context.Context, params json.RawMessage) (any, error) {
			<-releaseCh
			return map[string]any{"ok": true}, nil
		})
		svc   = New(fakeClusterService{primary: true}, registry, nil).(*serviceImpl)
		jobID = insertTestJob(t, ctx, "host:scheduler-singleton", jobmeta.JobScopeMasterOnly, jobmeta.JobConcurrencySingleton, 1, 0)
	)
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(releaseCh) })
		cleanupSchedulerJob(t, ctx, jobID)
	})

	svc.runCronJob(ctx, jobID)
	waitForCondition(t, 2*time.Second, func() bool {
		svc.mu.Lock()
		defer svc.mu.Unlock()
		return svc.runningCounts[jobID] == 1
	})

	svc.runCronJob(ctx, jobID)
	releaseOnce.Do(func() { close(releaseCh) })

	waitForCondition(t, 3*time.Second, func() bool {
		statuses := latestLogStatuses(t, ctx, jobID)
		if len(statuses) != 2 {
			return false
		}
		return statuses[0] == string(jobmeta.LogStatusRunning) || statuses[1] == string(jobmeta.LogStatusSkippedSingleton)
	})
	waitForCondition(t, 3*time.Second, func() bool {
		statuses := latestLogStatuses(t, ctx, jobID)
		if len(statuses) != 2 {
			return false
		}
		foundSuccess := false
		foundSkip := false
		for _, status := range statuses {
			if status == string(jobmeta.LogStatusSuccess) {
				foundSuccess = true
			}
			if status == string(jobmeta.LogStatusSkippedSingleton) {
				foundSkip = true
			}
		}
		return foundSuccess && foundSkip
	})
}

// TestRunCronJobMaxExecutionsDisablesJob verifies the scheduler auto-disables exhausted jobs.
func TestRunCronJobMaxExecutionsDisablesJob(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = newRegistryWithHandler(t, "host:scheduler-max", func(ctx context.Context, params json.RawMessage) (any, error) {
			return map[string]any{"ok": true}, nil
		})
		svc   = New(fakeClusterService{primary: true}, registry, nil).(*serviceImpl)
		jobID = insertTestJob(t, ctx, "host:scheduler-max", jobmeta.JobScopeMasterOnly, jobmeta.JobConcurrencySingleton, 1, 1)
	)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	svc.runCronJob(ctx, jobID)
	waitForCondition(t, 3*time.Second, func() bool {
		var jobRow *entity.SysJob
		if err := dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil || jobRow == nil {
			return false
		}
		return jobRow.Status == string(jobmeta.JobStatusDisabled) &&
			jobRow.StopReason == string(jobmeta.StopReasonMaxExecutionsReached) &&
			jobRow.ExecutedCount == 1
	})
}

// TestRunCronJobUnlimitedExecutionsStillAccumulatesCount verifies cron-triggered
// runs still increment executed_count when max_executions is unlimited.
func TestRunCronJobUnlimitedExecutionsStillAccumulatesCount(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = newRegistryWithHandler(t, "host:scheduler-unlimited", func(ctx context.Context, params json.RawMessage) (any, error) {
			return map[string]any{"ok": true}, nil
		})
		svc   = New(fakeClusterService{primary: true}, registry, nil).(*serviceImpl)
		jobID = insertTestJob(t, ctx, "host:scheduler-unlimited", jobmeta.JobScopeMasterOnly, jobmeta.JobConcurrencySingleton, 1, 0)
	)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	svc.runCronJob(ctx, jobID)
	waitForCondition(t, 3*time.Second, func() bool {
		var jobRow *entity.SysJob
		if err := dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil || jobRow == nil {
			return false
		}
		return jobRow.Status == string(jobmeta.JobStatusEnabled) &&
			jobRow.StopReason == "" &&
			jobRow.ExecutedCount == 1
	})
}

// TestRunCronJobHandlerTimeoutMarksLogTimeout verifies handler executions that overrun their timeout persist timeout logs.
func TestRunCronJobHandlerTimeoutMarksLogTimeout(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = newRegistryWithHandler(t, "host:scheduler-timeout", func(ctx context.Context, params json.RawMessage) (any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		svc serviceImpl
	)

	svc = *(New(fakeClusterService{primary: true}, registry, nil).(*serviceImpl))

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        testDefaultGroupID(t, ctx),
		Name:           fmt.Sprintf("scheduler-timeout-%d", time.Now().UnixNano()),
		TaskType:       string(jobmeta.TaskTypeHandler),
		HandlerRef:     "host:scheduler-timeout",
		Params:         `{}`,
		TimeoutSeconds: 1,
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusEnabled),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected timeout test job insert to succeed, got error: %v", err)
	}
	jobID := uint64(insertID)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	svc.runCronJob(ctx, jobID)
	waitForCondition(t, 3*time.Second, func() bool {
		logs := latestLogs(t, ctx, jobID)
		return len(logs) == 1 && logs[0] != nil && logs[0].Status == string(jobmeta.LogStatusTimeout)
	})

	logs := latestLogs(t, ctx, jobID)
	if len(logs) != 1 || logs[0] == nil {
		t.Fatalf("expected one timeout log row, got %#v", logs)
	}
	if logs[0].ErrMsg == "" {
		t.Fatal("expected timeout log to keep an error message")
	}
	if !strings.Contains(logs[0].ErrMsg, "1s") {
		t.Fatalf("expected timeout log to include configured timeout, got %q", logs[0].ErrMsg)
	}
}

// TestCancelLogCancelsRunningShellExecution verifies scheduler-level cancellation updates the running log to cancelled.
func TestCancelLogCancelsRunningShellExecution(t *testing.T) {
	var (
		ctx      = context.Background()
		registry = jobhandler.New()
		svc      = New(
			fakeClusterService{primary: true},
			registry,
			fakeShellExecutor{
				execute: func(ctx context.Context, in shellexec.ExecuteInput) (*shellexec.ExecuteOutput, error) {
					<-ctx.Done()
					return &shellexec.ExecuteOutput{
						Cancelled: true,
						ExitCode:  -1,
					}, ctx.Err()
				},
			},
		).(*serviceImpl)
	)

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        testDefaultGroupID(t, ctx),
		Name:           fmt.Sprintf("scheduler-cancel-%d", time.Now().UnixNano()),
		TaskType:       string(jobmeta.TaskTypeShell),
		TimeoutSeconds: 60,
		ShellCmd:       "sleep 30",
		CronExpr:       "* * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(jobmeta.JobStatusEnabled),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected shell cancel test job insert to succeed, got error: %v", err)
	}
	jobID := uint64(insertID)
	t.Cleanup(func() { cleanupSchedulerJob(t, ctx, jobID) })

	svc.runCronJob(ctx, jobID)

	var logID uint64
	waitForCondition(t, 2*time.Second, func() bool {
		logs := latestLogs(t, ctx, jobID)
		if len(logs) != 1 || logs[0] == nil || logs[0].Status != string(jobmeta.LogStatusRunning) {
			return false
		}
		logID = logs[0].Id
		return logID > 0
	})

	if err = svc.CancelLog(ctx, logID); err != nil {
		t.Fatalf("expected scheduler cancel to succeed, got error: %v", err)
	}

	waitForCondition(t, 3*time.Second, func() bool {
		logs := latestLogs(t, ctx, jobID)
		return len(logs) == 1 && logs[0] != nil && logs[0].Status == string(jobmeta.LogStatusCancelled)
	})

	logs := latestLogs(t, ctx, jobID)
	if len(logs) != 1 || logs[0] == nil {
		t.Fatalf("expected one cancelled log row, got %#v", logs)
	}
	if logs[0].ResultJson == "" || !strings.Contains(logs[0].ResultJson, `"cancelled":true`) {
		t.Fatalf("expected cancelled log result_json to record cancellation, got %q", logs[0].ResultJson)
	}
}

// TestNormalizeGcronPatternSupportsFiveAndSixFields verifies stored cron expressions are normalized for gcron registration.
func TestNormalizeGcronPatternSupportsFiveAndSixFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "five fields",
			input: "17 3 * * *",
			want:  "# 17 3 * * *",
		},
		{
			name:  "six fields",
			input: "0 */5 * * * *",
			want:  "0 */5 * * * *",
		},
		{
			name:    "unsupported fields",
			input:   "* * * *",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeGcronPattern(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected normalizeGcronPattern(%q) to fail", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected normalizeGcronPattern(%q) to succeed, got error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("expected normalizeGcronPattern(%q)=%q, got %q", tc.input, tc.want, got)
			}
		})
	}
}
