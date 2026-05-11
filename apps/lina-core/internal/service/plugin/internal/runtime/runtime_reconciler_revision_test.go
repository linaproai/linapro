// This file verifies revision-gated background reconciler decisions.

package runtime

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/pluginruntimecache"
)

// reconcilerRevisionTestTopology provides deterministic cluster topology for
// revision-gated reconciler tests.
type reconcilerRevisionTestTopology struct {
	cluster bool
	primary bool
	nodeID  string
}

// IsClusterModeEnabled reports the configured cluster switch.
func (t reconcilerRevisionTestTopology) IsClusterModeEnabled() bool {
	return t.cluster
}

// IsPrimaryNode reports the configured primary flag.
func (t reconcilerRevisionTestTopology) IsPrimaryNode() bool {
	return t.primary
}

// CurrentNodeID returns the configured node identifier.
func (t reconcilerRevisionTestTopology) CurrentNodeID() string {
	return t.nodeID
}

// reconcilerRevisionCacheCoord provides deterministic cachecoord behavior for
// runtime reconciler revision tests.
type reconcilerRevisionCacheCoord struct {
	revision     int64
	currentCall  int
	currentScope cachecoord.Scope
	markCalls    int
	markScope    cachecoord.Scope
	markReason   cachecoord.ChangeReason
}

// ConfigureDomain is a no-op because these tests configure domain metadata elsewhere.
func (f *reconcilerRevisionCacheCoord) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged advances and returns the in-memory reconciler revision.
func (f *reconcilerRevisionCacheCoord) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	f.markCalls++
	f.markScope = scope
	f.markReason = reason
	f.revision++
	return f.revision, nil
}

// MarkTenantChanged advances and returns the in-memory tenant-scoped reconciler revision.
func (f *reconcilerRevisionCacheCoord) MarkTenantChanged(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	_ cachecoord.InvalidationScope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	return f.MarkChanged(ctx, domain, scope, reason)
}

// EnsureFresh runs the refresher against the configured revision.
func (f *reconcilerRevisionCacheCoord) EnsureFresh(
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

// CurrentRevision returns the configured reconciler revision.
func (f *reconcilerRevisionCacheCoord) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
) (int64, error) {
	f.currentCall++
	f.currentScope = scope
	return f.revision, nil
}

// Snapshot is unused by reconciler revision tests.
func (f *reconcilerRevisionCacheCoord) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
}

// newTestReconcilerRevisionController creates a cachecoord-backed reconciler
// revision controller for tests.
func newTestReconcilerRevisionController(
	fakeCoord cachecoord.Service,
	observed *pluginruntimecache.ObservedRevision,
) *pluginruntimecache.Controller {
	return pluginruntimecache.NewControllerForScopeWithCoordinator(
		cachecoord.ScopeReconciler,
		pluginruntimecache.ReconcilerCacheChangeReason,
		true,
		fakeCoord,
		observed,
		nil,
	)
}

// TestNextBackgroundReconcileDecisionUsesSharedRevision verifies the background
// loop skips full scans after it has consumed the current shared revision.
func TestNextBackgroundReconcileDecisionUsesSharedRevision(t *testing.T) {
	ctx := context.Background()
	fakeCoord := &reconcilerRevisionCacheCoord{revision: 3}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: false,
			nodeID:  "node-a",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl:     newTestReconcilerRevisionController(fakeCoord, observed),
		lastReconcilerSweepAt:      time.Now(),
	}

	decision, err := service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("first reconcile decision failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 3 || decision.reason != "revision_changed" {
		t.Fatalf("expected revision-changed run for revision 3, got %+v", decision)
	}
	service.markBackgroundReconcileObserved(decision.revision, time.Now())

	decision, err = service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("second reconcile decision failed: %v", err)
	}
	if decision.shouldRun {
		t.Fatalf("expected unchanged revision to skip full scan, got %+v", decision)
	}

	fakeCoord.revision = 4
	decision, err = service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("third reconcile decision failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 4 || decision.reason != "revision_changed" {
		t.Fatalf("expected revision-changed run for revision 4, got %+v", decision)
	}
	if fakeCoord.currentScope != cachecoord.ScopeReconciler {
		t.Fatalf("expected reconciler scope %q, got %q", cachecoord.ScopeReconciler, fakeCoord.currentScope)
	}
}

// TestNextBackgroundReconcileDecisionAllowsSafetySweep verifies the reconciler
// still performs a low-frequency full scan when the revision is unchanged.
func TestNextBackgroundReconcileDecisionAllowsSafetySweep(t *testing.T) {
	ctx := context.Background()
	fakeCoord := &reconcilerRevisionCacheCoord{revision: 9}
	observed := pluginruntimecache.NewObservedRevision()
	observed.Store(9)
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: true,
			nodeID:  "node-primary",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl:     newTestReconcilerRevisionController(fakeCoord, observed),
		lastReconcilerSweepAt:      time.Now().Add(-runtimeReconcilerSafetySweepInterval - time.Second),
	}

	decision, err := service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("safety sweep decision failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 9 || decision.reason != "safety_sweep" {
		t.Fatalf("expected safety-sweep run for revision 9, got %+v", decision)
	}
}

// TestNotifyReconcilerChangedUsesReconcilerScope verifies dynamic runtime
// mutations publish wake-up revisions under the reconciler coordination domain.
func TestNotifyReconcilerChangedUsesReconcilerScope(t *testing.T) {
	fakeCoord := &reconcilerRevisionCacheCoord{revision: 12}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: false,
			nodeID:  "node-b",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl:     newTestReconcilerRevisionController(fakeCoord, observed),
	}

	if err := service.notifyReconcilerChanged(context.Background(), "test_mutation"); err != nil {
		t.Fatalf("notify reconciler changed failed: %v", err)
	}
	if fakeCoord.markCalls != 1 {
		t.Fatalf("expected one shared revision publish, got %d", fakeCoord.markCalls)
	}
	if fakeCoord.markScope != cachecoord.ScopeReconciler {
		t.Fatalf("expected reconciler scope %q, got %q", cachecoord.ScopeReconciler, fakeCoord.markScope)
	}
	if fakeCoord.markReason != pluginruntimecache.ReconcilerCacheChangeReason {
		t.Fatalf("expected reconciler reason %q, got %q", pluginruntimecache.ReconcilerCacheChangeReason, fakeCoord.markReason)
	}
	if !service.reconcilerRevisionCtrl.IsObserved(13) {
		t.Fatalf("expected published revision 13 to be observed locally")
	}
}

// TestPublishReconcilerChangedCanLeaveLocalRevisionUnobserved verifies primary
// foreground requests can publish desired-state changes while preserving
// background retry behavior if the immediate convergence fails.
func TestPublishReconcilerChangedCanLeaveLocalRevisionUnobserved(t *testing.T) {
	fakeCoord := &reconcilerRevisionCacheCoord{revision: 20}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: true,
			nodeID:  "node-primary",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl:     newTestReconcilerRevisionController(fakeCoord, observed),
		lastReconcilerSweepAt:      time.Now(),
	}

	if err := service.publishReconcilerChanged(context.Background(), "desired_state_changed", false); err != nil {
		t.Fatalf("publish reconciler changed failed: %v", err)
	}
	if service.reconcilerRevisionCtrl.IsObserved(21) {
		t.Fatalf("expected published revision 21 to remain unobserved locally")
	}

	decision, err := service.nextBackgroundReconcileDecision(context.Background())
	if err != nil {
		t.Fatalf("decision after unobserved publish failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 21 || decision.reason != "revision_changed" {
		t.Fatalf("expected background retry for unobserved revision 21, got %+v", decision)
	}
}
