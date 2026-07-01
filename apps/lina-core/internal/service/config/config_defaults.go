// This file centralizes host configuration default metadata used by raw
// HostConfig reads and default-aware typed helpers.

package config

import (
	"fmt"
	"strings"
)

// hostConfigDefaultSpec describes one host-owned default value by raw
// HostConfig key. The value mirrors the config.yaml representation whenever
// the typed static getter expects duration strings or paths.
type hostConfigDefaultSpec struct {
	Key   string // Key is the raw HostConfig key that owns the default.
	Value any    // Value is the default returned after runtime and static misses.
}

// staticHostConfigDefaultSpecs lists hard-coded static config defaults that
// already exist in typed host config getters.
var staticHostConfigDefaultSpecs = []hostConfigDefaultSpec{
	{Key: "jwt.expire", Value: "24h"},
	{Key: "session.timeout", Value: "24h"},
	{Key: "session.cleanupInterval", Value: "5m"},
	{Key: "upload.path", Value: defaultUploadPath},
	{Key: "upload.maxSize", Value: defaultUploadMaxSize},
	{Key: "workspace.basePath", Value: defaultWorkspaceBasePath},
	{Key: "cluster.enabled", Value: false},
	{Key: "cluster.election.lease", Value: "30s"},
	{Key: "cluster.election.renewInterval", Value: "10s"},
	{Key: "cluster.redis.connectTimeout", Value: "3s"},
	{Key: "cluster.redis.readTimeout", Value: "2s"},
	{Key: "cluster.redis.writeTimeout", Value: "2s"},
	{Key: "scheduler.defaultTimezone", Value: "UTC"},
	{Key: "logger.path", Value: ""},
	{Key: "logger.file", Value: defaultLoggerFilePattern},
	{Key: "logger.stdout", Value: true},
	{Key: "logger.extensions.structured", Value: false},
	{Key: "logger.extensions.traceIDEnabled", Value: false},
	{Key: "plugin.allowForceUninstall", Value: defaultPluginAllowForceUninstall},
	{Key: "plugin.dynamic.storagePath", Value: defaultPluginDynamicStoragePath},
	{Key: "server.extensions.apiDocPath", Value: defaultServerApiDocPath},
}

// hostConfigDefaultValuesByKey indexes every built-in host default that raw
// HostConfig reads may use after runtime and static config sources miss.
var hostConfigDefaultValuesByKey = func() map[string]any {
	defaults := make(
		map[string]any,
		len(runtimeParamSpecs)+len(publicFrontendSettingSpecs)+len(staticHostConfigDefaultSpecs),
	)
	for _, spec := range runtimeParamSpecs {
		registerHostConfigDefault(defaults, spec.Key, spec.DefaultValue)
	}
	for _, spec := range publicFrontendSettingSpecs {
		registerHostConfigDefault(defaults, spec.Key, spec.DefaultValue)
	}
	for _, spec := range staticHostConfigDefaultSpecs {
		registerHostConfigDefault(defaults, spec.Key, spec.Value)
	}
	return defaults
}()

// registerHostConfigDefault stores one normalized host config default key.
func registerHostConfigDefault(defaults map[string]any, key string, value any) {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return
	}
	defaults[normalizedKey] = value
}

// lookupHostConfigDefaultValue returns one host-owned default value by key.
func lookupHostConfigDefaultValue(key string) (any, bool) {
	value, ok := hostConfigDefaultValuesByKey[strings.TrimSpace(key)]
	return value, ok
}

// hostConfigDefaultSpecsCopy returns a detached copy of static default specs.
func hostConfigDefaultSpecsCopy() []hostConfigDefaultSpec {
	specs := make([]hostConfigDefaultSpec, len(staticHostConfigDefaultSpecs))
	copy(specs, staticHostConfigDefaultSpecs)
	return specs
}

// hostConfigDefaultValueString formats a metadata default for string-based
// protected config getters.
func hostConfigDefaultValueString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
