// This file manages runtime i18n bundle cache invalidation hooks.

package i18n

import "lina-core/pkg/pluginhost"

func init() {
	pluginhost.RegisterSourcePluginRegistryListener(invalidateRuntimeBundleCache)
}

// InvalidateRuntimeBundleCache clears the cached runtime translation bundles.
func (s *serviceImpl) InvalidateRuntimeBundleCache() {
	invalidateRuntimeBundleCache()
}

// InvalidateContentCache clears cached sys_i18n_content lookup results.
func (s *serviceImpl) InvalidateContentCache() {
	invalidateContentCache()
}

// invalidateRuntimeBundleCache resets the in-memory runtime bundle cache.
func invalidateRuntimeBundleCache() {
	runtimeBundleCache.Lock()
	runtimeBundleCache.bundles = make(map[string]map[string]string)
	runtimeBundleCache.Unlock()

	runtimeLocaleCache.Lock()
	runtimeLocaleCache.loaded = false
	runtimeLocaleCache.locales = nil
	runtimeLocaleCache.Unlock()
}

// invalidateContentCache resets the in-memory business-content cache.
func invalidateContentCache() {
	runtimeContentCache.Lock()
	runtimeContentCache.variants = make(map[string]map[string]ContentVariant)
	runtimeContentCache.Unlock()
}
