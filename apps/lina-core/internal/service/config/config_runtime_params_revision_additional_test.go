// This file verifies clustered runtime-parameter revision synchronization and
// cache helper edge cases.

package config

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/bizerr"
)

// fakeClusterRevisionCacheCoordService provides deterministic cachecoord behavior
// for clustered runtime-parameter revision tests.
type fakeClusterRevisionCacheCoordService struct {
	currentRevision int64
	currentErr      error
	currentCalls    int32
	markRevision    int64
	markErr         error
	markCalls       int32
}

// ConfigureDomain is a no-op because these tests configure domain metadata elsewhere.
func (f *fakeClusterRevisionCacheCoordService) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged returns the configured changed revision or error.
func (f *fakeClusterRevisionCacheCoordService) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
	_ cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// MarkTenantChanged returns the same configured revision as the global change
// path because runtime-parameter tests only verify revision coordination.
func (f *fakeClusterRevisionCacheCoordService) MarkTenantChanged(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	_ cachecoord.InvalidationScope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	return f.MarkChanged(ctx, domain, scope, reason)
}

// EnsureFresh runs the refresher against the configured current revision.
func (f *fakeClusterRevisionCacheCoordService) EnsureFresh(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	refresher cachecoord.Refresher,
) (int64, error) {
	revision, err := f.CurrentRevision(ctx, domain, scope)
	if err != nil {
		return 0, err
	}
	if refresher != nil {
		if err = refresher(ctx, revision); err != nil {
			return 0, err
		}
	}
	return revision, nil
}

// CurrentRevision returns the configured revision value or the configured error.
func (f *fakeClusterRevisionCacheCoordService) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
) (int64, error) {
	atomic.AddInt32(&f.currentCalls, 1)
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.currentRevision, nil
}

// Snapshot is unused by runtime-parameter revision tests.
func (f *fakeClusterRevisionCacheCoordService) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
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

// TestClusterRuntimeParamRevisionControllerCurrentRevisionEnsuresFreshValue verifies
// request-path revision checks consult cachecoord instead of indefinitely
// trusting the process-local copy.
func TestClusterRuntimeParamRevisionControllerCurrentRevisionEnsuresFreshValue(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeCoord := &fakeClusterRevisionCacheCoordService{currentRevision: 7}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

	revision, err := controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("load shared revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected shared revision 7, got %d", revision)
	}

	revision, err = controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("reload shared revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected refreshed revision 7, got %d", revision)
	}
	if calls := atomic.LoadInt32(&fakeCoord.currentCalls); calls != 2 {
		t.Fatalf("expected two cachecoord read calls, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState verifies
// explicit sync always refreshes from cachecoord and replaces the local revision.
func TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState(t *testing.T) {
	clearLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(3)
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeCoord := &fakeClusterRevisionCacheCoordService{currentRevision: 9}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

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

	fakeCoord := &fakeClusterRevisionCacheCoordService{markRevision: 11}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

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
	if calls := atomic.LoadInt32(&fakeCoord.markCalls); calls != 1 {
		t.Fatalf("expected one cachecoord publish call, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerPropagatesCacheCoordErrors verifies
// cachecoord read and publish failures surface to callers.
func TestClusterRuntimeParamRevisionControllerPropagatesCacheCoordErrors(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	readErr := errors.New("read revision failed")
	readCoord := &fakeClusterRevisionCacheCoordService{currentErr: readErr}
	readController := &clusterRuntimeParamRevisionController{cacheCoordSvc: readCoord}
	if _, err := readController.CurrentRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected CurrentRevision error %v, got %v", readErr, err)
	}
	if _, err := readController.SyncRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected SyncRevision error %v, got %v", readErr, err)
	}

	writeErr := errors.New("increment revision failed")
	writeCoord := &fakeClusterRevisionCacheCoordService{markErr: writeErr}
	writeController := &clusterRuntimeParamRevisionController{cacheCoordSvc: writeCoord}
	if _, err := writeController.MarkChanged(context.Background()); !errors.Is(err, writeErr) {
		t.Fatalf("expected MarkChanged error %v, got %v", writeErr, err)
	}
}

// TestClusterRuntimeParamRevisionControllerConsumesCrossInstanceRevision
// verifies a second controller instance can observe a revision published by
// another clustered writer through the persistent coordination row.
func TestClusterRuntimeParamRevisionControllerConsumesCrossInstanceRevision(t *testing.T) {
	ctx := context.Background()
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)
	coordSvc := coordination.NewMemory(nil)

	publisher := &clusterRuntimeParamRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}
	consumer := &clusterRuntimeParamRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}

	revision, err := publisher.MarkChanged(ctx)
	if err != nil {
		t.Fatalf("publish runtime-param revision failed: %v", err)
	}
	clearLocalRuntimeParamRevision()

	observed, err := consumer.SyncRevision(ctx)
	if err != nil {
		t.Fatalf("consume runtime-param revision from second controller failed: %v", err)
	}
	if observed != revision {
		t.Fatalf("expected consumer revision %d, got %d", revision, observed)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != revision {
		t.Fatalf("expected local runtime-param revision %d, got value=%d ok=%t", revision, local, ok)
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

// TestRuntimeParamSnapshotReturnsControllerUnavailableError verifies malformed
// service construction no longer relies on recover to hide missing revision wiring.
func TestRuntimeParamSnapshotReturnsControllerUnavailableError(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	snapshot, err := (&serviceImpl{}).getRuntimeParamSnapshot(ctx)
	if err == nil {
		t.Fatal("expected missing runtime-param revision controller to return an error")
	}
	if snapshot != nil {
		t.Fatal("expected no snapshot when revision controller is unavailable")
	}
	if !bizerr.Is(err, CodeConfigRuntimeParamRevisionUnavailable) {
		t.Fatalf("expected runtime-param revision unavailable error, got %v", err)
	}
}

// TestRuntimeParamSnapshotPropagatesRevisionErrorWithCachedSnapshot verifies a
// stale process-local snapshot is not reused when freshness cannot be confirmed.
func TestRuntimeParamSnapshotPropagatesRevisionErrorWithCachedSnapshot(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	currentErr := errors.New("current revision failed")
	svc := &serviceImpl{
		runtimeParamRevisionCtrl: &fakeRuntimeParamRevisionController{currentErr: currentErr},
	}
	cached := &cachedRuntimeParamSnapshot{
		Revision:    3,
		RefreshedAt: time.Now(),
		Snapshot: &runtimeParamSnapshot{
			revision:       3,
			values:         map[string]string{RuntimeParamKeyJWTExpire: "12h"},
			durationValues: map[string]time.Duration{RuntimeParamKeyJWTExpire: 12 * time.Hour},
			int64Values:    map[string]int64{},
			parseErrors:    map[string]error{},
		},
	}
	if err := runtimeParamSnapshotCache.Set(
		ctx,
		runtimeParamSnapshotCacheKey,
		cached,
		runtimeParamSnapshotCacheTTL,
	); err != nil {
		t.Fatalf("seed runtime param snapshot cache: %v", err)
	}

	snapshot, err := svc.getRuntimeParamSnapshot(ctx)
	if !errors.Is(err, currentErr) {
		t.Fatalf("expected revision error %v, got %v", currentErr, err)
	}
	if snapshot != nil {
		t.Fatal("expected no snapshot when revision freshness check fails")
	}
}

// cleanupRuntimeConfigRevision removes the shared runtime-config revision row
// used by cross-instance tests.
func cleanupRuntimeConfigRevision(t *testing.T, ctx context.Context) {
	t.Helper()

	cleanup := func() {
		if _, err := dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
			Domain: runtimeParamCacheDomain,
			Scope:  cachecoord.ScopeGlobal,
		}).Delete(); err != nil {
			t.Fatalf("cleanup runtime-config revision failed: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
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
