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

// invalidateRuntimeBundleCache resets the in-memory runtime bundle cache.
func invalidateRuntimeBundleCache() {
	runtimeBundleCache.Lock()
	defer runtimeBundleCache.Unlock()

	runtimeBundleCache.bundles = make(map[string]map[string]string)
}
