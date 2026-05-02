// Package pluginruntimecache coordinates plugin runtime cache revisions across
// cluster nodes through the host KV cache.
package pluginruntimecache

import (
	"context"
	"sync"

	"lina-core/internal/service/kvcache"
)

const (
	// revisionOwnerKey identifies the host module that owns plugin runtime revisions.
	revisionOwnerKey = "plugin-runtime"
	// revisionNamespace groups runtime cache coordination entries.
	revisionNamespace = "runtime-cache"
	// reconcilerRevisionNamespace groups dynamic runtime reconciler wake-up entries.
	reconcilerRevisionNamespace = "reconciler"
	// revisionLogicalKey stores the monotonic plugin runtime cache revision.
	revisionLogicalKey = "revision"
)

// RevisionCacheKey is the shared KV key used by all plugin runtime cache
// consumers. Each cache domain keeps its own observed revision while reading
// this same shared value.
var RevisionCacheKey = kvcache.BuildCacheKey(
	revisionOwnerKey,
	revisionNamespace,
	revisionLogicalKey,
)

// ReconcilerRevisionCacheKey is the shared KV key used to wake clustered
// dynamic-plugin reconcilers only when desired runtime state may have changed.
var ReconcilerRevisionCacheKey = kvcache.BuildCacheKey(
	revisionOwnerKey,
	reconcilerRevisionNamespace,
	revisionLogicalKey,
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

// Controller hides the cluster switch and shared KV protocol for one local
// plugin runtime cache domain.
type Controller struct {
	clusterEnabled bool
	kvCacheSvc     kvcache.Service
	observed       *ObservedRevision
	refresher      Refresher
	cacheKey       string
}

// NewController creates a runtime cache revision controller. When cluster mode
// is disabled, all methods become local no-ops so single-node deployments keep
// direct in-process invalidation behavior.
func NewController(
	clusterEnabled bool,
	kvCacheSvc kvcache.Service,
	observed *ObservedRevision,
	refresher Refresher,
) *Controller {
	return NewControllerForKey(
		RevisionCacheKey,
		clusterEnabled,
		kvCacheSvc,
		observed,
		refresher,
	)
}

// NewControllerForKey creates a revision controller for one explicit shared KV
// key. It is used when multiple plugin-runtime coordination domains need
// independent revisions, such as runtime cache invalidation and reconciler wake-up.
func NewControllerForKey(
	cacheKey string,
	clusterEnabled bool,
	kvCacheSvc kvcache.Service,
	observed *ObservedRevision,
	refresher Refresher,
) *Controller {
	if observed == nil {
		observed = NewObservedRevision()
	}
	if cacheKey == "" {
		cacheKey = RevisionCacheKey
	}
	return &Controller{
		clusterEnabled: clusterEnabled,
		kvCacheSvc:     kvCacheSvc,
		observed:       observed,
		refresher:      refresher,
		cacheKey:       cacheKey,
	}
}

// EnsureFresh refreshes this process-local cache domain when cluster mode is
// enabled and shared KV reports a newer plugin runtime revision.
func (c *Controller) EnsureFresh(ctx context.Context) error {
	if c == nil || !c.clusterEnabled || c.kvCacheSvc == nil {
		return nil
	}
	revision, err := c.CurrentRevision(ctx)
	if err != nil {
		return err
	}
	return c.observed.Ensure(ctx, revision, c.refresher)
}

// CurrentRevision returns the current shared revision for this controller. A
// missing shared entry is treated as revision 0.
func (c *Controller) CurrentRevision(ctx context.Context) (int64, error) {
	if c == nil || !c.clusterEnabled || c.kvCacheSvc == nil {
		return 0, nil
	}
	revision, found, err := c.kvCacheSvc.GetInt(
		ctx,
		kvcache.OwnerTypeModule,
		c.cacheKey,
	)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, nil
	}
	return revision, nil
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
// nodes. Single-node deployments skip shared KV and return revision 0.
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
	if c == nil || !c.clusterEnabled || c.kvCacheSvc == nil {
		return 0, nil
	}
	item, err := c.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		c.cacheKey,
		1,
		0,
	)
	if err != nil {
		return 0, err
	}
	if storeObserved {
		c.observed.Store(item.IntValue)
	}
	return item.IntValue, nil
}
