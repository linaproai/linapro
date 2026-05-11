// This file keeps shared scheduled-job management test helpers.

package jobmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"sync"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
)

// noopScheduler keeps job-management unit tests focused on validation and persistence.
type noopScheduler struct{}

// LoadAndRegister is a no-op for validation-focused unit tests.
func (noopScheduler) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh is a no-op for validation-focused unit tests.
func (noopScheduler) Refresh(ctx context.Context, jobID int64) error { return nil }

// RegisterJobSnapshot is a no-op for validation-focused unit tests.
func (noopScheduler) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error { return nil }

// Remove is a no-op for validation-focused unit tests.
func (noopScheduler) Remove(jobID int64) {}

// Trigger is unsupported in validation-focused unit tests.
func (noopScheduler) Trigger(ctx context.Context, jobID int64) (int64, error) { return 0, nil }

// CancelLog is unsupported in validation-focused unit tests.
func (noopScheduler) CancelLog(ctx context.Context, logID int64) error { return nil }

// jobmgmtStaticBizCtx returns a fixed request business context for service tests.
type jobmgmtStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by service tests because they inject context directly.
func (s jobmgmtStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s jobmgmtStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetTenant is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}


// SetUserAccess is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// trackingScheduler captures refresh and remove calls for registry-cascade tests.
type trackingScheduler struct {
	mu        sync.Mutex
	refreshed []int64
	removed   []int64
}

// LoadAndRegister is a no-op for registry-cascade tests.
func (s *trackingScheduler) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh records refreshed job IDs.
func (s *trackingScheduler) Refresh(ctx context.Context, jobID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = append(s.refreshed, jobID)
	return nil
}

// RegisterJobSnapshot records refreshed job IDs for declaration-driven tests.
func (s *trackingScheduler) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return nil
	}
	return s.Refresh(ctx, job.Id)
}

// Remove records removed job IDs.
func (s *trackingScheduler) Remove(jobID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removed = append(s.removed, jobID)
}

// Trigger is unsupported in registry-cascade tests.
func (s *trackingScheduler) Trigger(ctx context.Context, jobID int64) (int64, error) { return 0, nil }

// CancelLog is unsupported in registry-cascade tests.
func (s *trackingScheduler) CancelLog(ctx context.Context, logID int64) error { return nil }

// reset clears recorded scheduler calls.
func (s *trackingScheduler) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = nil
	s.removed = nil
}

// refreshedIDs returns one copy of all recorded refresh calls.
func (s *trackingScheduler) refreshedIDs() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.refreshed...)
}

// removedIDs returns one copy of all recorded remove calls.
func (s *trackingScheduler) removedIDs() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.removed...)
}

// containsJobID reports whether one scheduler call snapshot contains the given job ID.
func containsJobID(jobIDs []int64, target int64) bool {
	for _, jobID := range jobIDs {
		if jobID == target {
			return true
		}
	}
	return false
}

// noopCleaner keeps host-handler registration lightweight for unit tests.
type noopCleaner struct{}

// CleanupDueLogs is a no-op for host handler registration tests.
func (noopCleaner) CleanupDueLogs(ctx context.Context) (int64, error) { return 0, nil }

// newTestService constructs one DB-backed job-management service with host handlers registered.
func newTestService(t *testing.T) *serviceImpl {
	t.Helper()

	registry := jobhandler.New()
	if err := jobhandler.RegisterHostHandlers(registry, noopCleaner{}); err != nil {
		t.Fatalf("expected host handler registration to succeed, got error: %v", err)
	}
	return New(nil, registry, noopScheduler{}).(*serviceImpl)
}

// newTestServiceWithRegistry constructs one DB-backed job-management service with
// explicit registry and scheduler dependencies for lifecycle-cascade tests.
func newTestServiceWithRegistry(
	t *testing.T,
	registry jobhandler.Registry,
	scheduler *trackingScheduler,
) *serviceImpl {
	t.Helper()

	if registry == nil {
		registry = jobhandler.New()
	}
	if scheduler == nil {
		scheduler = &trackingScheduler{}
	}
	return New(nil, registry, scheduler).(*serviceImpl)
}

// defaultGroupID resolves the current default job group ID for tests.
func defaultGroupID(t *testing.T, ctx context.Context) int64 {
	t.Helper()

	var group *entity.SysJobGroup
	if err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{IsDefault: 1}).
		Scan(&group); err != nil {
		t.Fatalf("expected default job group query to succeed, got error: %v", err)
	}
	if group == nil {
		t.Fatal("expected default scheduled job group to exist")
	}
	return group.Id
}

// cleanupJobHard removes one job and its logs using hard-delete semantics.
func cleanupJobHard(t *testing.T, ctx context.Context, jobID int64) {
	t.Helper()
	if jobID == 0 {
		return
	}
	if _, err := dao.SysJobLog.Ctx(ctx).Where(do.SysJobLog{JobId: jobID}).Delete(); err != nil {
		t.Fatalf("expected job-log cleanup to succeed, got error: %v", err)
	}
	if _, err := dao.SysJob.Ctx(ctx).Unscoped().Where(do.SysJob{Id: jobID}).Delete(); err != nil {
		t.Fatalf("expected job cleanup to succeed, got error: %v", err)
	}
}

// cleanupGroupHard removes one group using hard-delete semantics.
func cleanupGroupHard(t *testing.T, ctx context.Context, groupID int64) {
	t.Helper()
	if groupID == 0 {
		return
	}
	if _, err := dao.SysJobGroup.Ctx(ctx).Unscoped().Where(do.SysJobGroup{Id: groupID}).Delete(); err != nil {
		t.Fatalf("expected group cleanup to succeed, got error: %v", err)
	}
}

// decodeJobParams converts one persisted params JSON string back to a map for tests.
func decodeJobParams(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		panic(fmt.Sprintf("invalid persisted job params JSON: %v", err))
	}
	return result
}

// retentionOverrideFromJob converts one persisted override JSON string to a typed option for tests.
func retentionOverrideFromJob(raw string) *jobmeta.RetentionOption {
	option, err := jobmeta.ParseRetentionOption(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid persisted job retention override JSON: %v", err))
	}
	return option
}

// syncBuiltinHandlerJob projects one handler-based code-owned job into sys_job
// and returns the persisted row ID for assertions.
func syncBuiltinHandlerJob(
	t *testing.T,
	ctx context.Context,
	svc *serviceImpl,
	def BuiltinJobDef,
) int64 {
	t.Helper()

	if svc == nil {
		t.Fatal("expected service to be initialized")
	}
	if _, err := svc.SyncBuiltinJobs(ctx, []BuiltinJobDef{def}); err != nil {
		t.Fatalf("expected builtin job sync to succeed, got error: %v", err)
	}

	var job *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1, HandlerRef: def.HandlerRef}).
		Scan(&job); err != nil {
		t.Fatalf("expected builtin job query to succeed, got error: %v", err)
	}
	if job == nil || job.Id == 0 {
		t.Fatalf("expected builtin job %s to exist after sync", def.HandlerRef)
	}
	return job.Id
}

// uniqueTestName returns one collision-resistant identifier for DB-backed tests.
func uniqueTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
