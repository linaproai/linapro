// This file verifies handler-registry availability changes cascade into
// scheduled-job persistence state.

package jobmgmt

import (
	"context"
	"encoding/json"
	"testing"

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
	newTestServiceWithRegistry(t, registry, scheduler)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}
	if err := registry.Register(disabledHandler); err != nil {
		t.Fatalf("expected disabled plugin handler registration to succeed, got error: %v", err)
	}

	enabledJobID := insertRegistryHandlerJob(t, ctx, handler.Ref, jobmeta.JobStatusEnabled)
	t.Cleanup(func() { cleanupJobHard(t, ctx, enabledJobID) })

	disabledJobID := insertRegistryHandlerJob(t, ctx, disabledHandler.Ref, jobmeta.JobStatusDisabled)
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
	newTestServiceWithRegistry(t, registry, scheduler)

	if err := registry.Register(handler); err != nil {
		t.Fatalf("expected plugin handler registration to succeed, got error: %v", err)
	}

	jobID := insertRegistryHandlerJob(t, ctx, handler.Ref, jobmeta.JobStatusEnabled)
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

// insertRegistryHandlerJob stores one user-defined plugin handler job used by
// registry availability cascade tests.
func insertRegistryHandlerJob(
	t *testing.T,
	ctx context.Context,
	handlerRef string,
	status jobmeta.JobStatus,
) uint64 {
	t.Helper()

	insertID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        defaultGroupID(t, ctx),
		Name:           uniqueTestName("plugin-handler-job"),
		Description:    "Plugin handler cascade test job.",
		TaskType:       string(jobmeta.TaskTypeHandler),
		HandlerRef:     handlerRef,
		Params:         `{}`,
		TimeoutSeconds: 30,
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          string(jobmeta.JobScopeMasterOnly),
		Concurrency:    string(jobmeta.JobConcurrencySingleton),
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         string(status),
		IsBuiltin:      0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected registry cascade job insert to succeed, got error: %v", err)
	}
	return uint64(insertID)
}
