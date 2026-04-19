// This file keeps shared scheduled-job management test helpers.

package jobmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
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
func (noopScheduler) Refresh(ctx context.Context, jobID uint64) error { return nil }

// Remove is a no-op for validation-focused unit tests.
func (noopScheduler) Remove(jobID uint64) {}

// Trigger is unsupported in validation-focused unit tests.
func (noopScheduler) Trigger(ctx context.Context, jobID uint64) (uint64, error) { return 0, nil }

// CancelLog is unsupported in validation-focused unit tests.
func (noopScheduler) CancelLog(ctx context.Context, logID uint64) error { return nil }

// trackingScheduler captures refresh and remove calls for registry-cascade tests.
type trackingScheduler struct {
	mu        sync.Mutex
	refreshed []uint64
	removed   []uint64
}

// LoadAndRegister is a no-op for registry-cascade tests.
func (s *trackingScheduler) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh records refreshed job IDs.
func (s *trackingScheduler) Refresh(ctx context.Context, jobID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = append(s.refreshed, jobID)
	return nil
}

// Remove records removed job IDs.
func (s *trackingScheduler) Remove(jobID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removed = append(s.removed, jobID)
}

// Trigger is unsupported in registry-cascade tests.
func (s *trackingScheduler) Trigger(ctx context.Context, jobID uint64) (uint64, error) { return 0, nil }

// CancelLog is unsupported in registry-cascade tests.
func (s *trackingScheduler) CancelLog(ctx context.Context, logID uint64) error { return nil }

// reset clears recorded scheduler calls.
func (s *trackingScheduler) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = nil
	s.removed = nil
}

// refreshedIDs returns one copy of all recorded refresh calls.
func (s *trackingScheduler) refreshedIDs() []uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]uint64(nil), s.refreshed...)
}

// removedIDs returns one copy of all recorded remove calls.
func (s *trackingScheduler) removedIDs() []uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]uint64(nil), s.removed...)
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
func defaultGroupID(t *testing.T, ctx context.Context) uint64 {
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
func cleanupJobHard(t *testing.T, ctx context.Context, jobID uint64) {
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
func cleanupGroupHard(t *testing.T, ctx context.Context, groupID uint64) {
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
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}

// retentionOverrideFromJob converts one persisted override JSON string to a typed option for tests.
func retentionOverrideFromJob(raw string) *jobmeta.RetentionOption {
	option, _ := jobmeta.ParseRetentionOption(raw)
	return option
}

// uniqueTestName returns one collision-resistant identifier for DB-backed tests.
func uniqueTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
