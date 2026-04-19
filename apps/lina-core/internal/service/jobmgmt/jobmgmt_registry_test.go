// This file verifies handler-registry availability changes cascade into
// scheduled-job persistence state.

package jobmgmt

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
)

// TestHandlerUnregisterPausesEnabledJobs verifies handler disappearance pauses
// enabled jobs while leaving disabled jobs untouched.
func TestHandlerUnregisterPausesEnabledJobs(t *testing.T) {
	var (
		ctx       = context.Background()
		registry  = jobhandler.New()
		scheduler = &trackingScheduler{}
		svc       = newTestServiceWithRegistry(t, registry, scheduler)
		handler   = jobhandler.HandlerDef{
			Ref:          "plugin:test-job-handler/wait",
			DisplayName:  "Plugin Test Wait Handler",
			Description:  "Used to verify registry cascade behavior.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     "plugin:test-job-handler",
			Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
				return map[string]any{"ok": true}, nil
			},
		}
	)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}

	enabledJobID, err := svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("plugin-enabled-job"),
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusEnabled,
	})
	if err != nil {
		t.Fatalf("expected enabled job create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, enabledJobID) })

	disabledJobID, err := svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("plugin-disabled-job"),
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusDisabled,
	})
	if err != nil {
		t.Fatalf("expected disabled job create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, disabledJobID) })

	scheduler.reset()
	registry.Unregister(handler.Ref)

	enabledJob := mustLoadJobRow(t, ctx, enabledJobID)
	if got := jobmeta.NormalizeJobStatus(enabledJob.Status); got != jobmeta.JobStatusPausedByPlugin {
		t.Fatalf("expected enabled job to become paused_by_plugin, got %s", got)
	}
	if enabledJob.StopReason != string(jobmeta.StopReasonPluginUnavailable) {
		t.Fatalf("expected enabled job stop_reason=%s, got %s", jobmeta.StopReasonPluginUnavailable, enabledJob.StopReason)
	}

	disabledJob := mustLoadJobRow(t, ctx, disabledJobID)
	if got := jobmeta.NormalizeJobStatus(disabledJob.Status); got != jobmeta.JobStatusDisabled {
		t.Fatalf("expected disabled job to stay disabled, got %s", got)
	}

	removedIDs := scheduler.removedIDs()
	if len(removedIDs) != 1 || removedIDs[0] != enabledJobID {
		t.Fatalf("expected only enabled job to be removed from scheduler, got %#v", removedIDs)
	}
}

// TestHandlerRegisterRestoresPausedJobs verifies handler re-registration restores
// jobs paused because the plugin handler was unavailable.
func TestHandlerRegisterRestoresPausedJobs(t *testing.T) {
	var (
		ctx       = context.Background()
		registry  = jobhandler.New()
		scheduler = &trackingScheduler{}
		svc       = newTestServiceWithRegistry(t, registry, scheduler)
		handler   = jobhandler.HandlerDef{
			Ref:          "plugin:test-job-handler/restore",
			DisplayName:  "Plugin Restore Handler",
			Description:  "Used to verify paused job restoration.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     "plugin:test-job-handler",
			Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
				return map[string]any{"restored": true}, nil
			},
		}
	)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}

	jobID, err := svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("plugin-restorable-job"),
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusEnabled,
	})
	if err != nil {
		t.Fatalf("expected job create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	registry.Unregister(handler.Ref)
	scheduler.reset()

	if err = registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler re-registration to succeed, got error: %v", err)
	}

	jobRow := mustLoadJobRow(t, ctx, jobID)
	if got := jobmeta.NormalizeJobStatus(jobRow.Status); got != jobmeta.JobStatusEnabled {
		t.Fatalf("expected paused job to be restored to enabled, got %s", got)
	}
	if jobRow.StopReason != "" {
		t.Fatalf("expected restored job stop_reason to be cleared, got %s", jobRow.StopReason)
	}

	refreshedIDs := scheduler.refreshedIDs()
	if len(refreshedIDs) != 1 || refreshedIDs[0] != jobID {
		t.Fatalf("expected restored job to be refreshed in scheduler, got %#v", refreshedIDs)
	}
}

// mustLoadJobRow loads one persisted scheduled-job row for assertions.
func mustLoadJobRow(t *testing.T, ctx context.Context, jobID uint64) *entity.SysJob {
	t.Helper()

	var jobRow *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil {
		t.Fatalf("expected scheduled job query to succeed, got error: %v", err)
	}
	if jobRow == nil {
		t.Fatalf("expected scheduled job %d to exist", jobID)
	}
	return jobRow
}
