// Package revisionctrl coordinates runtime cache revisions across cluster nodes
// through cachecoord for plugin runtime and i18n cache consumers.
package revisionctrl

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/cachecoord"
)

// Plugin runtime cache coordination reasons.
const (
	// runtimeCacheDomain coordinates plugin runtime, frontend, i18n, and Wasm derived caches.
	runtimeCacheDomain cachecoord.Domain = "plugin-runtime"
	// RuntimeCacheChangeReason records normal plugin runtime derived-cache invalidation.
	RuntimeCacheChangeReason cachecoord.ChangeReason = "plugin_runtime_changed"
	// ReconcilerCacheChangeReason records dynamic reconciler wake-up changes.
	ReconcilerCacheChangeReason cachecoord.ChangeReason = "plugin_reconciler_changed"
	// runtimeCacheMaxStale is the plugin-runtime freshness budget.
	runtimeCacheMaxStale = 5 * time.Second
)

// Refresher rebuilds or invalidates one process-local plugin runtime cache
// domain after another cluster node publishes a newer shared revision.
type Refresher func(ctx context.Context, revision int64) error

// ObservedRevision records the latest shared revision consumed by one local
// cache domain.
type ObservedRevision struct {
	mu     sync.Mutex
	value  int64
	loaded bool
}

// NewObservedRevision creates an empty local revision marker for one cache domain.
func NewObservedRevision() *ObservedRevision {
	return &ObservedRevision{}
}

// Controller hides the cluster switch and cachecoord protocol for one local
// plugin runtime cache domain.
type Controller struct {
	clusterEnabled bool
	cacheCoordSvc  cachecoord.Service
	observed       *ObservedRevision
	refresher      Refresher
	scope          cachecoord.Scope
	changeReason   cachecoord.ChangeReason
	tenantScope    cachecoord.InvalidationScope
}

// NewControllerWithCoordinator creates a controller backed by the unified
// cachecoord service.
func NewControllerWithCoordinator(
	clusterEnabled bool,
	cacheCoordSvc cachecoord.Service,
	observed *ObservedRevision,
	refresher Refresher,
) *Controller {
	return NewControllerForScopeWithCoordinator(
		cachecoord.ScopeGlobal,
		RuntimeCacheChangeReason,
		clusterEnabled,
		cacheCoordSvc,
		observed,
		refresher,
	)
}

// NewControllerForScopeWithCoordinator creates a cachecoord-backed controller
// for one explicit plugin-runtime scope.
func NewControllerForScopeWithCoordinator(
	scope cachecoord.Scope,
	reason cachecoord.ChangeReason,
	clusterEnabled bool,
	cacheCoordSvc cachecoord.Service,
	observed *ObservedRevision,
	refresher Refresher,
) *Controller {
	if observed == nil {
		observed = NewObservedRevision()
	}
	if scope == "" {
		scope = cachecoord.ScopeGlobal
	}
	if reason == "" {
		reason = RuntimeCacheChangeReason
	}
	configureRuntimeCacheDomain(clusterEnabled, cacheCoordSvc)
	return &Controller{
		clusterEnabled: clusterEnabled,
		cacheCoordSvc:  cacheCoordSvc,
		observed:       observed,
		refresher:      refresher,
		scope:          scope,
		changeReason:   reason,
	}
}

// Store records that the cache domain has consumed the specified revision.
func (r *ObservedRevision) Store(revision int64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loaded && revision < r.value {
		return
	}
	r.value = revision
	r.loaded = true
}

// Matches reports whether the cache domain has already consumed the specified revision.
func (r *ObservedRevision) Matches(revision int64) bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loaded && r.value == revision
}

// Ensure checks the observed revision and runs refresher exactly once for the
// current caller when the shared revision has advanced.
func (r *ObservedRevision) Ensure(
	ctx context.Context,
	revision int64,
	refresher Refresher,
) error {
	if r == nil {
		if refresher != nil {
			return refresher(ctx, revision)
		}
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loaded && r.value == revision {
		return nil
	}
	if refresher != nil {
		if err := refresher(ctx, revision); err != nil {
			return err
		}
	}
	r.value = revision
	r.loaded = true
	return nil
}
