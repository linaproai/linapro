// This file coordinates cluster-wide wake-up revisions for the dynamic-plugin
// reconciler so the background loop avoids scanning every registry row on each tick.

package runtime

import (
	"context"
	"time"

	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/pluginruntimecache"
	"lina-core/pkg/logger"
)

// runtimeReconcilerSafetySweepInterval is the low-frequency fallback full-scan
// cadence used when no shared revision changes are observed.
const runtimeReconcilerSafetySweepInterval = 5 * time.Minute

// backgroundReconcileDecision captures whether one reconciler tick should run a
// full registry convergence pass and which shared revision it will consume.
type backgroundReconcileDecision struct {
	// revision is the current shared reconciler wake-up revision.
	revision int64
	// shouldRun reports whether the tick should execute a full convergence pass.
	shouldRun bool
	// reason identifies why the tick needs to run.
	reason string
}

// configureReconcilerRevisionController wires the cluster-aware controller
// after topology has been injected.
func (s *serviceImpl) configureReconcilerRevisionController() {
	if s.reconcilerRevisionObserved == nil {
		s.reconcilerRevisionObserved = pluginruntimecache.NewObservedRevision()
	}
	s.reconcilerRevisionCtrl = pluginruntimecache.NewControllerForKey(
		pluginruntimecache.ReconcilerRevisionCacheKey,
		s.isClusterModeEnabled(),
		kvcache.New(),
		s.reconcilerRevisionObserved,
		nil,
	)
}

// ensureReconcilerRevisionController returns the configured controller, creating
// a topology-aware no-op controller when tests or direct constructors skipped
// setter injection.
func (s *serviceImpl) ensureReconcilerRevisionController() *pluginruntimecache.Controller {
	if s.reconcilerRevisionCtrl == nil {
		s.configureReconcilerRevisionController()
	}
	return s.reconcilerRevisionCtrl
}

// notifyReconcilerChanged publishes one desired-state or runtime-state mutation
// that may require another node to run a dynamic-plugin convergence pass.
func (s *serviceImpl) notifyReconcilerChanged(ctx context.Context, reason string) error {
	return s.publishReconcilerChanged(ctx, reason, true)
}

// publishReconcilerChanged publishes one reconciler wake-up revision and
// controls whether the local runtime marks that revision as already consumed.
func (s *serviceImpl) publishReconcilerChanged(ctx context.Context, reason string, storeObserved bool) error {
	controller := s.ensureReconcilerRevisionController()
	if controller == nil {
		return nil
	}
	var (
		revision int64
		err      error
	)
	if storeObserved {
		revision, err = controller.MarkChanged(ctx)
	} else {
		revision, err = controller.PublishChanged(ctx)
	}
	if err != nil {
		return err
	}
	if revision > 0 {
		logger.Debugf(ctx, "dynamic plugin reconciler revision bumped reason=%s revision=%d", reason, revision)
	}
	return nil
}

// nextBackgroundReconcileDecision checks the shared revision and safety sweep
// timer to decide whether the current tick should perform a full scan.
func (s *serviceImpl) nextBackgroundReconcileDecision(ctx context.Context) (backgroundReconcileDecision, error) {
	if !s.isClusterModeEnabled() {
		return backgroundReconcileDecision{
			shouldRun: true,
			reason:    "single_node",
		}, nil
	}

	controller := s.ensureReconcilerRevisionController()
	revision, err := controller.CurrentRevision(ctx)
	if err != nil {
		return backgroundReconcileDecision{}, err
	}
	if !controller.IsObserved(revision) {
		return backgroundReconcileDecision{
			revision:  revision,
			shouldRun: true,
			reason:    "revision_changed",
		}, nil
	}
	if s.shouldRunReconcilerSafetySweep(time.Now()) {
		return backgroundReconcileDecision{
			revision:  revision,
			shouldRun: true,
			reason:    "safety_sweep",
		}, nil
	}
	return backgroundReconcileDecision{
		revision: revision,
		reason:   "revision_unchanged",
	}, nil
}

// shouldRunReconcilerSafetySweep reports whether the fallback full-scan interval
// has elapsed since the last successful background convergence pass.
func (s *serviceImpl) shouldRunReconcilerSafetySweep(now time.Time) bool {
	s.reconcilerSafetyMu.Lock()
	defer s.reconcilerSafetyMu.Unlock()
	if s.lastReconcilerSweepAt.IsZero() {
		return true
	}
	return now.Sub(s.lastReconcilerSweepAt) >= runtimeReconcilerSafetySweepInterval
}

// markBackgroundReconcileObserved records that a successful full-scan pass has
// consumed the shared reconciler revision.
func (s *serviceImpl) markBackgroundReconcileObserved(revision int64, now time.Time) {
	controller := s.ensureReconcilerRevisionController()
	if controller != nil {
		controller.StoreObserved(revision)
	}

	s.reconcilerSafetyMu.Lock()
	defer s.reconcilerSafetyMu.Unlock()
	s.lastReconcilerSweepAt = now
}
