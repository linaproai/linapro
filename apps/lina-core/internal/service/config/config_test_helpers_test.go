// This file keeps static configuration mutation helpers scoped to tests.

package config

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// hostConfigDefaultSpecsCopy returns a detached copy of static default specs for tests.
func hostConfigDefaultSpecsCopy() []hostConfigDefaultSpec {
	specs := make([]hostConfigDefaultSpec, len(staticHostConfigDefaultSpecs))
	copy(specs, staticHostConfigDefaultSpecs)
	return specs
}

// setPluginDynamicStoragePathOverride overrides the dynamic-plugin storage path for tests.
func setPluginDynamicStoragePathOverride(path string) {
	pluginDynamicStoragePathOverride.Store(strings.TrimSpace(path))
}

// setPluginAutoEnableOverride overrides the startup auto-enable plugin IDs for tests.
func setPluginAutoEnableOverride(pluginIDs []string) {
	if len(pluginIDs) == 0 {
		pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{})
		return
	}
	entries := make([]PluginAutoEnableEntry, 0, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		entries = append(entries, PluginAutoEnableEntry{ID: pluginID})
	}
	normalized, err := normalizePluginAutoEnableEntries(entries)
	if err != nil {
		panic(gerror.Wrap(err, "setPluginAutoEnableOverride received invalid plugin IDs"))
	}
	pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{
		set:   true,
		value: normalized,
	})
}

// setPluginAllowForceUninstallOverride overrides plugin.allowForceUninstall for tests.
func setPluginAllowForceUninstallOverride(value *bool) {
	if value == nil {
		pluginAllowForceUninstallOverride.Store(pluginAllowForceUninstallOverrideState{})
		return
	}
	pluginAllowForceUninstallOverride.Store(pluginAllowForceUninstallOverrideState{
		set:   true,
		value: *value,
	})
}

// publicFrontendSettingSpecsCopy returns public frontend setting specs for tests.
func publicFrontendSettingSpecsCopy() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(publicFrontendSettingSpecs))
	copy(specs, publicFrontendSettingSpecs)
	return specs
}

// runtimeParamSpecsCopy returns all built-in runtime parameter specs for tests.
func runtimeParamSpecsCopy() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(runtimeParamSpecs))
	copy(specs, runtimeParamSpecs)
	return specs
}

// runtimeParamSnapshotSyncIntervalForTest returns the multi-node watcher interval.
func runtimeParamSnapshotSyncIntervalForTest() time.Duration {
	return runtimeParamRevisionSyncInterval
}

// clearLocalRuntimeParamRevision removes the process-local revision marker for tests.
func clearLocalRuntimeParamRevision() {
	runtimeParamRevisionState.Lock()
	runtimeParamRevisionState.value = 0
	runtimeParamRevisionState.initialized = false
	runtimeParamRevisionState.Unlock()
}

// resetStaticConfigCaches drops all once guards and cached static config objects for tests.
func resetStaticConfigCaches() {
	processStaticConfigCaches = newStaticConfigCaches()
}
