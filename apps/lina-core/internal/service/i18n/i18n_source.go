// This file resolves source-aware runtime i18n bundles for diagnostics and maintenance APIs.

package i18n

import "context"

// loadRawLocaleBundleWithSources returns the merged flat catalog for one locale
// together with per-key source descriptors. The values come straight from the
// layered cache, so admin diagnostics share the same merge ordering and source
// attribution as the runtime translation hot path.
func (s *serviceImpl) loadRawLocaleBundleWithSources(ctx context.Context, locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	resolvedLocale := s.ResolveLocale(ctx, locale)

	// Trigger build through the standard path so any newly populated sectors
	// stay in cache for subsequent translation reads.
	merged := s.ensureMergedCatalog(ctx, resolvedLocale)

	lc := runtimeBundleCache.getOrCreate(resolvedLocale)
	lc.mu.RLock()
	sources := cloneSourceDescriptorMap(lc.sources)
	lc.mu.RUnlock()

	return cloneFlatMessageMap(merged), sources
}

// cloneSourceDescriptorMap copies one source descriptor map so callers can
// safely retain it while concurrent invalidations replace the cache state.
func cloneSourceDescriptorMap(src map[string]MessageSourceDescriptor) map[string]MessageSourceDescriptor {
	if len(src) == 0 {
		return map[string]MessageSourceDescriptor{}
	}
	dst := make(map[string]MessageSourceDescriptor, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
