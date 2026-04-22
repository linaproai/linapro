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
			Ref:          "plugin:test-job-handler/cron:wait",
			DisplayName:  "Plugin Test Wait Handler",
			Description:  "Used to verify registry cascade behavior.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     "test-job-handler",
			Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
				return map[string]any{"ok": true}, nil
			},
		}
		disabledHandler = jobhandler.HandlerDef{
			Ref:          "plugin:test-job-handler/cron:wait-disabled",
			DisplayName:  "Plugin Disabled Test Handler",
			Description:  "Used to verify disabled builtin jobs stay untouched.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     "test-job-handler",
			Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
				return map[string]any{"ok": true}, nil
			},
		}
	)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}
	if err := registry.Register(disabledHandler); err != nil {
		t.Fatalf("expected disabled plugin handler registration to succeed, got error: %v", err)
	}

	enabledJobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("plugin-enabled-job"),
		Description:    "Enabled plugin builtin job.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusEnabled,
	})
	t.Cleanup(func() { cleanupJobHard(t, ctx, enabledJobID) })

	disabledJobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("plugin-disabled-job"),
		Description:    "Disabled plugin builtin job.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     disabledHandler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusDisabled,
	})
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
			Ref:          "plugin:test-job-handler/cron:restore",
			DisplayName:  "Plugin Restore Handler",
			Description:  "Used to verify paused job restoration.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourcePlugin,
			PluginID:     "test-job-handler",
			Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
				return map[string]any{"restored": true}, nil
			},
		}
	)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}

	jobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("plugin-restorable-job"),
		Description:    "Restorable plugin builtin job.",
		TaskType:       jobmeta.TaskTypeHandler,
		HandlerRef:     handler.Ref,
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobmeta.JobScopeMasterOnly,
		Concurrency:    jobmeta.JobConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobmeta.JobStatusEnabled,
	})
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	registry.Unregister(handler.Ref)
	scheduler.reset()

	if err := registry.Register(handler); err != nil {
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
