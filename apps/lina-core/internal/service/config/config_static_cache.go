// This file provides process-local static config object caches so config.yaml
// sections are parsed once and then served by cloned objects.

package config

import "sync"

// staticConfigBox owns one process-local static config object. The once guard
// ensures each config.yaml section is parsed at most once during one process
// lifetime unless tests explicitly reset the cache box set.
type staticConfigBox[T any] struct {
	once  sync.Once
	value *T
}

// load returns the cached config object, initializing it lazily through the
// provided loader on first access only.
func (box *staticConfigBox[T]) load(loader func() *T) *T {
	box.once.Do(func() {
		box.value = loader()
	})
	return box.value
}

// staticConfigCaches groups all process-local static config boxes so
// configuration loading and test resets can be managed in one place.
type staticConfigCaches struct {
	cluster          staticConfigBox[ClusterConfig]
	jwt              staticConfigBox[JwtConfig]
	logger           staticConfigBox[LoggerConfig]
	metadata         staticConfigBox[MetadataConfig]
	monitor          staticConfigBox[MonitorConfig]
	plugin           staticConfigBox[PluginConfig]
	serverExtensions staticConfigBox[ServerExtensionsConfig]
	session          staticConfigBox[SessionConfig]
	upload           staticConfigBox[UploadConfig]
}

// processStaticConfigCaches is the singleton cache registry used by the config
// service to reuse static config.yaml sections across requests.
var processStaticConfigCaches = newStaticConfigCaches()

// newStaticConfigCaches allocates one empty cache registry. Production code
// uses it once during startup, while tests reuse it to clear once state.
func newStaticConfigCaches() *staticConfigCaches {
	return &staticConfigCaches{}
}

// resetStaticConfigCaches drops all once guards and cached objects. Tests call
// this after mutating config adapter content so later reads observe new data.
func resetStaticConfigCaches() {
	processStaticConfigCaches = newStaticConfigCaches()
}

// cloneClusterConfig returns a detached copy so callers cannot mutate the
// shared cached cluster config instance in process memory.
func cloneClusterConfig(cfg *ClusterConfig) *ClusterConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneJwtConfig returns a detached copy so runtime override logic can modify
// the effective values without polluting the static cache.
func cloneJwtConfig(cfg *JwtConfig) *JwtConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneLoggerConfig returns a detached copy of the cached logger config.
func cloneLoggerConfig(cfg *LoggerConfig) *LoggerConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneServerExtensionsConfig returns a detached copy of the cached server extension config.
func cloneServerExtensionsConfig(cfg *ServerExtensionsConfig) *ServerExtensionsConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneMetadataConfig deep-copies slice fields because metadata is shared by
// OpenAPI and system-info callers that must not mutate the cached backing
// slices.
func cloneMetadataConfig(cfg *MetadataConfig) *MetadataConfig {
	if cfg == nil {
		return nil
	}

	cloned := &MetadataConfig{
		Framework: cfg.Framework,
		OpenApi:   cfg.OpenApi,
	}
	if len(cfg.Backend) > 0 {
		cloned.Backend = append([]MetadataComponentInfo(nil), cfg.Backend...)
	}
	if len(cfg.Frontend) > 0 {
		cloned.Frontend = append([]MetadataComponentInfo(nil), cfg.Frontend...)
	}
	return cloned
}

// cloneMonitorConfig returns a detached copy of the cached monitor config.
func cloneMonitorConfig(cfg *MonitorConfig) *MonitorConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneOpenApiConfig returns a detached copy of the cached OpenAPI metadata.
func cloneOpenApiConfig(cfg *OpenApiConfig) *OpenApiConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// clonePluginConfig returns a detached copy of the cached plugin config.
func clonePluginConfig(cfg *PluginConfig) *PluginConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneSessionConfig returns a detached copy so runtime timeout overrides do
// not mutate the cached cleanup interval and base timeout values.
func cloneSessionConfig(cfg *SessionConfig) *SessionConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

// cloneUploadConfig returns a detached copy so runtime max-size overrides stay
// request-local and never mutate the cached upload path/default size.
func cloneUploadConfig(cfg *UploadConfig) *UploadConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}
