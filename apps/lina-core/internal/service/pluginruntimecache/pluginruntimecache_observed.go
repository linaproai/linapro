// This file implements process-local revision observation for plugin runtime
// cache domains.

package pluginruntimecache

import "context"

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
