// This file verifies plugin runtime cache revision coordination behavior.

package pluginruntimecache

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/cachecoord"
)

// fakePluginRuntimeCacheCoordService provides deterministic cachecoord behavior
// for revision tests.
type fakePluginRuntimeCacheCoordService struct {
	currentRevision int64
	currentErr      error
	currentCalls    int32
	currentScope    cachecoord.Scope
	markRevision    int64
	markErr         error
	markCalls       int32
	markScope       cachecoord.Scope
	markReason      cachecoord.ChangeReason
	markTenantScope cachecoord.InvalidationScope
}

// ConfigureDomain is a no-op because these tests configure domain metadata elsewhere.
func (f *fakePluginRuntimeCacheCoordService) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged returns the configured changed revision and tracks publish metadata.
func (f *fakePluginRuntimeCacheCoordService) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	f.markScope = scope
	f.markReason = reason
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// MarkTenantChanged returns the configured changed revision and tracks tenant-scoped publish metadata.
func (f *fakePluginRuntimeCacheCoordService) MarkTenantChanged(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
	tenantScope cachecoord.InvalidationScope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	f.markScope = scope
	f.markReason = reason
	f.markTenantScope = tenantScope
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// EnsureFresh runs the refresher against the configured current revision.
func (f *fakePluginRuntimeCacheCoordService) EnsureFresh(
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

// CurrentRevision returns the configured shared revision.
func (f *fakePluginRuntimeCacheCoordService) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
) (int64, error) {
	atomic.AddInt32(&f.currentCalls, 1)
	f.currentScope = scope
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.currentRevision, nil
}

// Snapshot is unused by revision tests.
func (f *fakePluginRuntimeCacheCoordService) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
}

// TestControllerSingleNodeNoops verifies single-node deployments avoid
// cachecoord reads, writes, and refresh callbacks.
func TestControllerSingleNodeNoops(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{currentRevision: 3, markRevision: 4}
	var refreshCalls int32
	controller := NewControllerWithCoordinator(false, fakeCoord, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("single-node ensure fresh failed: %v", err)
	}
	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("single-node mark changed failed: %v", err)
	}
	if revision != 0 {
		t.Fatalf("expected single-node revision 0, got %d", revision)
	}
	if atomic.LoadInt32(&fakeCoord.currentCalls) != 0 || atomic.LoadInt32(&fakeCoord.markCalls) != 0 {
		t.Fatalf("expected no cachecoord traffic, got current=%d mark=%d", fakeCoord.currentCalls, fakeCoord.markCalls)
	}
	if atomic.LoadInt32(&refreshCalls) != 0 {
		t.Fatalf("expected no refresh callback, got %d", refreshCalls)
	}
}

// TestControllerEnsureFreshRefreshesOnRevisionChange verifies each cache domain
// refreshes once per newly observed shared revision.
func TestControllerEnsureFreshRefreshesOnRevisionChange(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{currentRevision: 5}
	var refreshCalls int32
	controller := NewControllerWithCoordinator(true, fakeCoord, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("first ensure fresh failed: %v", err)
	}
	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("second ensure fresh failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected one refresh for revision 5, got %d", refreshCalls)
	}

	fakeCoord.currentRevision = 6
	if err := controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("third ensure fresh failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 2 {
		t.Fatalf("expected second refresh for revision 6, got %d", refreshCalls)
	}
}

// TestControllerMarkChangedStoresReturnedRevision verifies the mutating node
// records the revision it published so its next read path does not refresh again.
func TestControllerMarkChangedStoresReturnedRevision(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{currentRevision: 9, markRevision: 9}
	var refreshCalls int32
	controller := NewControllerWithCoordinator(true, fakeCoord, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("mark changed failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected revision 9, got %d", revision)
	}
	if err = controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("ensure after mark failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 0 {
		t.Fatalf("expected no refresh after local mark, got %d", refreshCalls)
	}
}

// TestControllerPublishChangedLeavesRevisionUnobserved verifies callers can
// publish a revision that the same local process should still consume later.
func TestControllerPublishChangedLeavesRevisionUnobserved(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{currentRevision: 10, markRevision: 10}
	var refreshCalls int32
	controller := NewControllerWithCoordinator(true, fakeCoord, NewObservedRevision(), func(_ context.Context) error {
		atomic.AddInt32(&refreshCalls, 1)
		return nil
	})

	revision, err := controller.PublishChanged(context.Background())
	if err != nil {
		t.Fatalf("publish changed failed: %v", err)
	}
	if revision != 10 {
		t.Fatalf("expected revision 10, got %d", revision)
	}
	if err = controller.EnsureFresh(context.Background()); err != nil {
		t.Fatalf("ensure after publish failed: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected refresh after unobserved publish, got %d", refreshCalls)
	}
}

// TestControllerPropagatesCacheCoordErrors verifies cachecoord failures are
// returned to callers that can fail closed.
func TestControllerPropagatesCacheCoordErrors(t *testing.T) {
	readErr := errors.New("read revision failed")
	readController := NewControllerWithCoordinator(
		true,
		&fakePluginRuntimeCacheCoordService{currentErr: readErr},
		NewObservedRevision(),
		nil,
	)
	if err := readController.EnsureFresh(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}

	writeErr := errors.New("write revision failed")
	writeController := NewControllerWithCoordinator(
		true,
		&fakePluginRuntimeCacheCoordService{markErr: writeErr},
		NewObservedRevision(),
		nil,
	)
	if _, err := writeController.MarkChanged(context.Background()); !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
}

// TestControllerForScopeUsesExplicitCacheCoordScope verifies non-default
// coordination scopes store revisions independently.
func TestControllerForScopeUsesExplicitCacheCoordScope(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{currentRevision: 11, markRevision: 12}
	controller := NewControllerForScopeWithCoordinator(
		cachecoord.ScopeReconciler,
		ReconcilerCacheChangeReason,
		true,
		fakeCoord,
		NewObservedRevision(),
		nil,
	)

	revision, err := controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("current revision failed: %v", err)
	}
	if revision != 11 {
		t.Fatalf("expected revision 11, got %d", revision)
	}
	if fakeCoord.currentScope != cachecoord.ScopeReconciler {
		t.Fatalf("expected current scope %q, got %q", cachecoord.ScopeReconciler, fakeCoord.currentScope)
	}

	revision, err = controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("mark changed failed: %v", err)
	}
	if revision != 12 {
		t.Fatalf("expected incremented revision 12, got %d", revision)
	}
	if fakeCoord.markScope != cachecoord.ScopeReconciler {
		t.Fatalf("expected mark scope %q, got %q", cachecoord.ScopeReconciler, fakeCoord.markScope)
	}
	if fakeCoord.markReason != ReconcilerCacheChangeReason {
		t.Fatalf("expected mark reason %q, got %q", ReconcilerCacheChangeReason, fakeCoord.markReason)
	}
}

// TestControllerMarkChangedCarriesTenantScope verifies plugin-runtime
// invalidation publishes tenant metadata instead of collapsing all tenants into
// one undifferentiated revision.
func TestControllerMarkChangedCarriesTenantScope(t *testing.T) {
	fakeCoord := &fakePluginRuntimeCacheCoordService{markRevision: 14}
	controller := NewControllerForScopeWithCoordinator(
		cachecoord.ScopeGlobal,
		RuntimeCacheChangeReason,
		true,
		fakeCoord,
		NewObservedRevision(),
		nil,
	).WithTenantScope(27, true)

	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("mark tenant-scoped plugin runtime changed failed: %v", err)
	}
	if revision != 14 {
		t.Fatalf("expected revision 14, got %d", revision)
	}
	if fakeCoord.markTenantScope.TenantID != 27 || !fakeCoord.markTenantScope.CascadeToTenants {
		t.Fatalf("expected tenant scoped publish metadata, got %#v", fakeCoord.markTenantScope)
	}
}

// TestControllerConsumesCrossInstancePluginRuntimeRevision verifies a second
// cache controller can observe plugin-runtime changes published by another
// instance through cachecoord.
func TestControllerConsumesCrossInstancePluginRuntimeRevision(t *testing.T) {
	ctx := context.Background()
	scope := cachecoord.Scope("unit-test-plugin-runtime-dual")
	cleanupPluginRuntimeRevision(t, ctx, scope)

	publisher := NewControllerForScopeWithCoordinator(
		scope,
		RuntimeCacheChangeReason,
		true,
		cachecoord.New(cachecoord.NewStaticTopology(true)),
		NewObservedRevision(),
		nil,
	)
	var refreshCalls int32
	consumer := NewControllerForScopeWithCoordinator(
		scope,
		RuntimeCacheChangeReason,
		true,
		cachecoord.New(cachecoord.NewStaticTopology(true)),
		NewObservedRevision(),
		func(_ context.Context) error {
			atomic.AddInt32(&refreshCalls, 1)
			return nil
		},
	)

	revision, err := publisher.MarkChanged(ctx)
	if err != nil {
		t.Fatalf("publish plugin-runtime revision failed: %v", err)
	}
	if err = consumer.EnsureFresh(ctx); err != nil {
		t.Fatalf("consume plugin-runtime revision from second controller failed: %v", err)
	}
	if !consumer.IsObserved(revision) {
		t.Fatalf("expected consumer to observe revision %d", revision)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Fatalf("expected one cache refresh after cross-instance revision, got %d", refreshCalls)
	}
}

// cleanupPluginRuntimeRevision removes one shared plugin-runtime revision row
// used by cross-instance tests.
func cleanupPluginRuntimeRevision(t *testing.T, ctx context.Context, scope cachecoord.Scope) {
	t.Helper()

	cleanup := func() {
		if _, err := dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
			Domain: runtimeCacheDomain,
			Scope:  scope,
		}).Delete(); err != nil {
			t.Fatalf("cleanup plugin-runtime revision failed: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
}
