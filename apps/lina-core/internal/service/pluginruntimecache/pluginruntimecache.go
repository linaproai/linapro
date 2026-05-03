// Package pluginruntimecache coordinates plugin runtime cache revisions across
// cluster nodes through cachecoord.
package pluginruntimecache

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
type Refresher func(ctx context.Context) error

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
			return refresher(ctx)
		}
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loaded && r.value == revision {
		return nil
	}
	if refresher != nil {
		if err := refresher(ctx); err != nil {
			return err
		}
	}
	r.value = revision
	r.loaded = true
	return nil
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

// configureRuntimeCacheDomain declares plugin-runtime consistency policy in
// the package that owns the plugin runtime cache semantics.
func configureRuntimeCacheDomain(clusterEnabled bool, cacheCoordSvc cachecoord.Service) {
	if !clusterEnabled || cacheCoordSvc == nil {
		return
	}
	if err := cacheCoordSvc.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           runtimeCacheDomain,
		AuthoritySource:  "plugin registry, active releases, plugin node state, and artifacts",
		ConsistencyModel: cachecoord.ConsistencySharedRevision,
		MaxStale:         runtimeCacheMaxStale,
		SyncMechanism:    "persistent MySQL sys_cache_revision plus runtime cache invalidation",
		FailureStrategy:  cachecoord.FailureStrategyConservativeHide,
	}); err != nil {
		panic(err)
	}
}

// EnsureFresh refreshes this process-local cache domain when cluster mode is
// enabled and cachecoord reports a newer plugin runtime revision.
func (c *Controller) EnsureFresh(ctx context.Context) error {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return nil
	}
	revision, err := c.CurrentRevision(ctx)
	if err != nil {
		return err
	}
	return c.observed.Ensure(ctx, revision, c.refresher)
}

// CurrentRevision returns the current shared revision for this controller.
func (c *Controller) CurrentRevision(ctx context.Context) (int64, error) {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return 0, nil
	}
	return c.cacheCoordSvc.CurrentRevision(
		ctx,
		runtimeCacheDomain,
		c.scope,
	)
}

// IsObserved reports whether this process-local domain has consumed revision.
func (c *Controller) IsObserved(revision int64) bool {
	if c == nil || !c.clusterEnabled {
		return true
	}
	return c.observed.Matches(revision)
}

// StoreObserved records that this process-local domain has consumed revision.
func (c *Controller) StoreObserved(revision int64) {
	if c == nil || !c.clusterEnabled {
		return
	}
	c.observed.Store(revision)
}

// MarkChanged publishes one plugin runtime cache mutation to other cluster
// nodes. Single-node deployments skip cachecoord and return revision 0.
func (c *Controller) MarkChanged(ctx context.Context) (int64, error) {
	return c.markChanged(ctx, true)
}

// PublishChanged publishes one shared revision without recording it as consumed
// by the local process. It is used by callers that need the background consumer
// on the same node to retry work if the foreground mutation fails afterward.
func (c *Controller) PublishChanged(ctx context.Context) (int64, error) {
	return c.markChanged(ctx, false)
}

// markChanged increments the shared revision and optionally records the
// returned value as already consumed by this local cache domain.
func (c *Controller) markChanged(ctx context.Context, storeObserved bool) (int64, error) {
	if c == nil || !c.clusterEnabled || c.cacheCoordSvc == nil {
		return 0, nil
	}
	revision, err := c.cacheCoordSvc.MarkChanged(
		ctx,
		runtimeCacheDomain,
		c.scope,
		c.changeReason,
	)
	if err != nil {
		return 0, err
	}
	if storeObserved {
		c.observed.Store(revision)
	}
	return revision, nil
}
