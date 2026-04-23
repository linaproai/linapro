// This file verifies clustered runtime-parameter revision synchronization and
// cache helper edge cases.

package config

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/kvcache"
)

// fakeClusterRevisionKVCacheService provides deterministic shared-KV behavior
// for clustered runtime-parameter revision tests.
type fakeClusterRevisionKVCacheService struct {
	getIntValue int64
	getIntErr   error
	getIntCalls int32
	incrValue   int64
	incrErr     error
	incrCalls   int32
}

// Get returns no string cache item because these tests only exercise integer revisions.
func (f *fakeClusterRevisionKVCacheService) Get(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

// GetInt returns the configured revision value or the configured error.
func (f *fakeClusterRevisionKVCacheService) GetInt(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (int64, bool, error) {
	atomic.AddInt32(&f.getIntCalls, 1)
	if f.getIntErr != nil {
		return 0, false, f.getIntErr
	}
	return f.getIntValue, true, nil
}

// Set is a no-op success stub because these tests never write string cache values.
func (f *fakeClusterRevisionKVCacheService) Set(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ int64,
) (*kvcache.Item, error) {
	return nil, nil
}

// Delete is a no-op success stub because these tests never delete shared KV keys.
func (f *fakeClusterRevisionKVCacheService) Delete(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) error {
	return nil
}

// Incr returns the configured incremented revision value or the configured error.
func (f *fakeClusterRevisionKVCacheService) Incr(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
	_ int64,
) (*kvcache.Item, error) {
	atomic.AddInt32(&f.incrCalls, 1)
	if f.incrErr != nil {
		return nil, f.incrErr
	}
	return &kvcache.Item{IntValue: f.incrValue}, nil
}

// Expire is a no-op stub because expiration is irrelevant to revision tests.
func (f *fakeClusterRevisionKVCacheService) Expire(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired is a no-op success stub because expiration cleanup is not part of these tests.
func (f *fakeClusterRevisionKVCacheService) CleanupExpired(_ context.Context) error {
	return nil
}

// fakeRuntimeParamRevisionController provides deterministic revision behavior
// for service-level cache helper tests.
type fakeRuntimeParamRevisionController struct {
	currentRevision int64
	syncRevision    int64
	markRevision    int64
	currentErr      error
	syncErr         error
	markErr         error
	markCalls       int32
}

// CurrentRevision returns the configured current revision or error.
func (f *fakeRuntimeParamRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.currentRevision, nil
}

// SyncRevision returns the configured synchronized revision or error.
func (f *fakeRuntimeParamRevisionController) SyncRevision(_ context.Context) (int64, error) {
	if f.syncErr != nil {
		return 0, f.syncErr
	}
	return f.syncRevision, nil
}

// MarkChanged returns the configured changed revision or error.
func (f *fakeRuntimeParamRevisionController) MarkChanged(_ context.Context) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// TestClusterRuntimeParamRevisionControllerCurrentRevisionCachesSharedValue verifies
// the clustered controller reads shared KV once and then serves the local copy.
func TestClusterRuntimeParamRevisionControllerCurrentRevisionCachesSharedValue(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeKV := &fakeClusterRevisionKVCacheService{getIntValue: 7}
	controller := &clusterRuntimeParamRevisionController{kvCacheSvc: fakeKV}

	revision, err := controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("load shared revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected shared revision 7, got %d", revision)
	}

	revision, err = controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("load cached local revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected cached revision 7, got %d", revision)
	}
	if calls := atomic.LoadInt32(&fakeKV.getIntCalls); calls != 1 {
		t.Fatalf("expected one shared GetInt call, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState verifies
// explicit sync always refreshes from shared KV and replaces the local revision.
func TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState(t *testing.T) {
	clearLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(3)
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeKV := &fakeClusterRevisionKVCacheService{getIntValue: 9}
	controller := &clusterRuntimeParamRevisionController{kvCacheSvc: fakeKV}

	revision, err := controller.SyncRevision(context.Background())
	if err != nil {
		t.Fatalf("sync shared revision: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected synced revision 9, got %d", revision)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != 9 {
		t.Fatalf("expected local revision to be updated to 9, got value=%d ok=%t", local, ok)
	}
}

// TestClusterRuntimeParamRevisionControllerMarkChangedStoresReturnedRevision verifies
// successful shared increments are mirrored locally for the writing node.
func TestClusterRuntimeParamRevisionControllerMarkChangedStoresReturnedRevision(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeKV := &fakeClusterRevisionKVCacheService{incrValue: 11}
	controller := &clusterRuntimeParamRevisionController{kvCacheSvc: fakeKV}

	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("increment shared revision: %v", err)
	}
	if revision != 11 {
		t.Fatalf("expected incremented revision 11, got %d", revision)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != 11 {
		t.Fatalf("expected local revision to be updated to 11, got value=%d ok=%t", local, ok)
	}
	if calls := atomic.LoadInt32(&fakeKV.incrCalls); calls != 1 {
		t.Fatalf("expected one shared Incr call, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerPropagatesSharedKVErrors verifies
// shared-KV read and increment failures surface to callers.
func TestClusterRuntimeParamRevisionControllerPropagatesSharedKVErrors(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	readErr := errors.New("read revision failed")
	readKV := &fakeClusterRevisionKVCacheService{getIntErr: readErr}
	readController := &clusterRuntimeParamRevisionController{kvCacheSvc: readKV}
	if _, err := readController.CurrentRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected CurrentRevision error %v, got %v", readErr, err)
	}
	if _, err := readController.SyncRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected SyncRevision error %v, got %v", readErr, err)
	}

	writeErr := errors.New("increment revision failed")
	writeKV := &fakeClusterRevisionKVCacheService{incrErr: writeErr}
	writeController := &clusterRuntimeParamRevisionController{kvCacheSvc: writeKV}
	if _, err := writeController.MarkChanged(context.Background()); !errors.Is(err, writeErr) {
		t.Fatalf("expected MarkChanged error %v, got %v", writeErr, err)
	}
}

// TestNotifyRuntimeParamsChangedSwallowsRevisionErrors verifies the helper is
// best-effort and never panics when revision publication fails.
func TestNotifyRuntimeParamsChangedSwallowsRevisionErrors(t *testing.T) {
	svc := &serviceImpl{
		runtimeParamRevisionCtrl: &fakeRuntimeParamRevisionController{markErr: errors.New("boom")},
	}

	svc.NotifyRuntimeParamsChanged(context.Background())
}

// TestRuntimeParamSnapshotSyncIntervalExposesWatcherInterval verifies the
// public helper returns the configured synchronization interval constant.
func TestRuntimeParamSnapshotSyncIntervalExposesWatcherInterval(t *testing.T) {
	if interval := RuntimeParamSnapshotSyncInterval(); interval != 10*time.Second {
		t.Fatalf("expected runtime snapshot sync interval 10s, got %s", interval)
	}
}

// TestGetCachedRuntimeParamSnapshotRemovesInvalidEntries verifies malformed or
// stale cache entries are dropped before later reads try to reuse them.
func TestGetCachedRuntimeParamSnapshotRemovesInvalidEntries(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	svc := New().(*serviceImpl)
	if err := runtimeParamSnapshotCache.Set(ctx, runtimeParamSnapshotCacheKey, "invalid", runtimeParamSnapshotCacheTTL); err != nil {
		t.Fatalf("seed invalid runtime snapshot cache entry: %v", err)
	}
	if cached := svc.getCachedRuntimeParamSnapshot(ctx, 1); cached != nil {
		t.Fatal("expected invalid cache entry to be rejected")
	}
	if cachedVar, err := runtimeParamSnapshotCache.Get(ctx, runtimeParamSnapshotCacheKey); err != nil {
		t.Fatalf("get invalid cache entry after removal: %v", err)
	} else if cachedVar != nil {
		t.Fatal("expected invalid cache entry to be removed")
	}

	valid := &cachedRuntimeParamSnapshot{
		Revision:    2,
		RefreshedAt: time.Now(),
		Snapshot: &runtimeParamSnapshot{
			revision:       2,
			values:         map[string]string{},
			durationValues: map[string]time.Duration{},
			int64Values:    map[string]int64{},
			parseErrors:    map[string]error{},
		},
	}
	if err := runtimeParamSnapshotCache.Set(ctx, runtimeParamSnapshotCacheKey, valid, runtimeParamSnapshotCacheTTL); err != nil {
		t.Fatalf("seed stale runtime snapshot cache entry: %v", err)
	}
	if cached := svc.getCachedRuntimeParamSnapshot(ctx, 1); cached != nil {
		t.Fatal("expected revision-mismatched cache entry to be rejected")
	}
}

// TestExtractCachedRuntimeParamSnapshotRejectsBrokenValues verifies defensive
// cache decoding ignores wrong types and nil snapshots.
func TestExtractCachedRuntimeParamSnapshotRejectsBrokenValues(t *testing.T) {
	if cached := extractCachedRuntimeParamSnapshot("invalid"); cached != nil {
		t.Fatal("expected string cache payload to be rejected")
	}
	if cached := extractCachedRuntimeParamSnapshot(&cachedRuntimeParamSnapshot{}); cached != nil {
		t.Fatal("expected cache payload without snapshot to be rejected")
	}
}
