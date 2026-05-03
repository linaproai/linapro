// This file tests built-in kvcache expired-entry cleanup job projection.

package cron

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
)

// fakeKVCacheService captures cleanup calls for cron tests.
type fakeKVCacheService struct {
	backendName     kvcache.BackendName
	cleanupRequired bool
	cleanupCalls    int
}

// BackendName returns the fake backend name.
func (f *fakeKVCacheService) BackendName() kvcache.BackendName { return f.backendName }

// RequiresExpiredCleanup returns the configured cleanup requirement.
func (f *fakeKVCacheService) RequiresExpiredCleanup() bool { return f.cleanupRequired }

// Get is unused by cron cleanup tests.
func (f *fakeKVCacheService) Get(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

// GetInt is unused by cron cleanup tests.
func (f *fakeKVCacheService) GetInt(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
) (int64, bool, error) {
	return 0, false, nil
}

// Set is unused by cron cleanup tests.
func (f *fakeKVCacheService) Set(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*kvcache.Item, error) {
	return nil, nil
}

// Delete is unused by cron cleanup tests.
func (f *fakeKVCacheService) Delete(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
) error {
	return nil
}

// Incr is unused by cron cleanup tests.
func (f *fakeKVCacheService) Incr(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*kvcache.Item, error) {
	return nil, nil
}

// Expire is unused by cron cleanup tests.
func (f *fakeKVCacheService) Expire(
	ctx context.Context,
	ownerType kvcache.OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired records one cleanup invocation.
func (f *fakeKVCacheService) CleanupExpired(ctx context.Context) error {
	f.cleanupCalls++
	return nil
}

// TestKVCacheCleanupJobProjectedForCleanupBackend verifies cleanup-capable
// backends get an hourly built-in scheduled job and registered handler.
func TestKVCacheCleanupJobProjectedForCleanupBackend(t *testing.T) {
	ctx := context.Background()
	cacheSvc := &fakeKVCacheService{
		backendName:     kvcache.BackendMySQLMemory,
		cleanupRequired: true,
	}
	registry := jobhandler.New()
	svc := &serviceImpl{
		configSvc:  hostconfig.New(),
		kvCacheSvc: cacheSvc,
		registry:   registry,
	}

	if err := svc.registerManagedHandlers(); err != nil {
		t.Fatalf("expected managed handler registration to succeed, got error: %v", err)
	}
	if _, ok := registry.Lookup("host:kvcache-cleanup-expired"); !ok {
		t.Fatal("expected kvcache cleanup handler to be registered")
	}

	var cleanupJob *jobmgmtsvc.BuiltinJobDef
	for _, item := range svc.buildHostBuiltinJobs(ctx) {
		item := item
		if item.HandlerRef == "host:kvcache-cleanup-expired" {
			cleanupJob = &item
			break
		}
	}
	if cleanupJob == nil {
		t.Fatal("expected kvcache cleanup job to be projected")
	}
	if cleanupJob.Pattern != "@every 1h0m0s" {
		t.Fatalf("expected hourly cleanup pattern, got %s", cleanupJob.Pattern)
	}
	if cleanupJob.Scope != jobmeta.JobScopeMasterOnly {
		t.Fatalf("expected master-only cleanup scope, got %s", cleanupJob.Scope)
	}
}

// TestInvokeKVCacheExpiredCleanupDelegatesToCacheService verifies the handler
// calls CleanupExpired and returns backend diagnostics.
func TestInvokeKVCacheExpiredCleanupDelegatesToCacheService(t *testing.T) {
	cacheSvc := &fakeKVCacheService{
		backendName:     kvcache.BackendMySQLMemory,
		cleanupRequired: true,
	}
	svc := &serviceImpl{kvCacheSvc: cacheSvc}

	result, err := svc.invokeKVCacheExpiredCleanup(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("expected cleanup handler to succeed, got error: %v", err)
	}
	if cacheSvc.cleanupCalls != 1 {
		t.Fatalf("expected one cleanup call, got %d", cacheSvc.cleanupCalls)
	}
	resultMap, ok := result.(map[string]any)
	if !ok || resultMap["backend"] != string(kvcache.BackendMySQLMemory) || resultMap["cleaned"] != true {
		t.Fatalf("unexpected cleanup result: %#v", result)
	}
}
