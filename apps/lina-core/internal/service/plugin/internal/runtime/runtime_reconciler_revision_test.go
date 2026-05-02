// This file verifies revision-gated background reconciler decisions.

package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/kvcache"
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

// reconcilerRevisionKV provides deterministic shared-KV behavior for runtime
// reconciler revision tests.
type reconcilerRevisionKV struct {
	revision  int64
	found     bool
	getCalls  int
	getKey    string
	incrCalls int
	incrKey   string
}

// Get is unused by reconciler revision tests.
func (f *reconcilerRevisionKV) Get(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

// GetInt returns the configured reconciler revision.
func (f *reconcilerRevisionKV) GetInt(
	_ context.Context,
	_ kvcache.OwnerType,
	cacheKey string,
) (int64, bool, error) {
	f.getCalls++
	f.getKey = cacheKey
	return f.revision, f.found, nil
}

// Set is unused by reconciler revision tests.
func (f *reconcilerRevisionKV) Set(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ int64,
) (*kvcache.Item, error) {
	return nil, nil
}

// Delete is unused by reconciler revision tests.
func (f *reconcilerRevisionKV) Delete(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) error {
	return nil
}

// Incr advances and returns the in-memory reconciler revision.
func (f *reconcilerRevisionKV) Incr(
	_ context.Context,
	_ kvcache.OwnerType,
	cacheKey string,
	delta int64,
	_ int64,
) (*kvcache.Item, error) {
	f.incrCalls++
	f.incrKey = cacheKey
	f.revision += delta
	f.found = true
	return &kvcache.Item{IntValue: f.revision}, nil
}

// Expire is unused by reconciler revision tests.
func (f *reconcilerRevisionKV) Expire(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired is unused by reconciler revision tests.
func (f *reconcilerRevisionKV) CleanupExpired(_ context.Context) error {
	return nil
}

// TestNextBackgroundReconcileDecisionUsesSharedRevision verifies the background
// loop skips full scans after it has consumed the current shared revision.
func TestNextBackgroundReconcileDecisionUsesSharedRevision(t *testing.T) {
	ctx := context.Background()
	fakeKV := &reconcilerRevisionKV{revision: 3, found: true}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: false,
			nodeID:  "node-a",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl: pluginruntimecache.NewControllerForKey(
			pluginruntimecache.ReconcilerRevisionCacheKey,
			true,
			fakeKV,
			observed,
			nil,
		),
		lastReconcilerSweepAt: time.Now(),
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

	fakeKV.revision = 4
	decision, err = service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("third reconcile decision failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 4 || decision.reason != "revision_changed" {
		t.Fatalf("expected revision-changed run for revision 4, got %+v", decision)
	}
	if fakeKV.getKey != pluginruntimecache.ReconcilerRevisionCacheKey {
		t.Fatalf("expected shared reconciler key %q, got %q", pluginruntimecache.ReconcilerRevisionCacheKey, fakeKV.getKey)
	}
}

// TestNextBackgroundReconcileDecisionAllowsSafetySweep verifies the reconciler
// still performs a low-frequency full scan when the revision is unchanged.
func TestNextBackgroundReconcileDecisionAllowsSafetySweep(t *testing.T) {
	ctx := context.Background()
	fakeKV := &reconcilerRevisionKV{revision: 9, found: true}
	observed := pluginruntimecache.NewObservedRevision()
	observed.Store(9)
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: true,
			nodeID:  "node-primary",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl: pluginruntimecache.NewControllerForKey(
			pluginruntimecache.ReconcilerRevisionCacheKey,
			true,
			fakeKV,
			observed,
			nil,
		),
		lastReconcilerSweepAt: time.Now().Add(-runtimeReconcilerSafetySweepInterval - time.Second),
	}

	decision, err := service.nextBackgroundReconcileDecision(ctx)
	if err != nil {
		t.Fatalf("safety sweep decision failed: %v", err)
	}
	if !decision.shouldRun || decision.revision != 9 || decision.reason != "safety_sweep" {
		t.Fatalf("expected safety-sweep run for revision 9, got %+v", decision)
	}
}

// TestNotifyReconcilerChangedUsesSharedReconcilerKey verifies dynamic runtime
// mutations publish wake-up revisions under the reconciler coordination domain.
func TestNotifyReconcilerChangedUsesSharedReconcilerKey(t *testing.T) {
	fakeKV := &reconcilerRevisionKV{revision: 12, found: true}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: false,
			nodeID:  "node-b",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl: pluginruntimecache.NewControllerForKey(
			pluginruntimecache.ReconcilerRevisionCacheKey,
			true,
			fakeKV,
			observed,
			nil,
		),
	}

	if err := service.notifyReconcilerChanged(context.Background(), "test_mutation"); err != nil {
		t.Fatalf("notify reconciler changed failed: %v", err)
	}
	if fakeKV.incrCalls != 1 {
		t.Fatalf("expected one shared revision increment, got %d", fakeKV.incrCalls)
	}
	if fakeKV.incrKey != pluginruntimecache.ReconcilerRevisionCacheKey {
		t.Fatalf("expected shared reconciler key %q, got %q", pluginruntimecache.ReconcilerRevisionCacheKey, fakeKV.incrKey)
	}
	if !service.reconcilerRevisionCtrl.IsObserved(13) {
		t.Fatalf("expected published revision 13 to be observed locally")
	}
}

// TestPublishReconcilerChangedCanLeaveLocalRevisionUnobserved verifies primary
// foreground requests can publish desired-state changes while preserving
// background retry behavior if the immediate convergence fails.
func TestPublishReconcilerChangedCanLeaveLocalRevisionUnobserved(t *testing.T) {
	fakeKV := &reconcilerRevisionKV{revision: 20, found: true}
	observed := pluginruntimecache.NewObservedRevision()
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{
			cluster: true,
			primary: true,
			nodeID:  "node-primary",
		},
		reconcilerRevisionObserved: observed,
		reconcilerRevisionCtrl: pluginruntimecache.NewControllerForKey(
			pluginruntimecache.ReconcilerRevisionCacheKey,
			true,
			fakeKV,
			observed,
			nil,
		),
		lastReconcilerSweepAt: time.Now(),
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
