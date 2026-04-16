package config

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
)

var pluginDynamicStoragePathOverride atomic.Value

// PluginConfig holds plugin-related host configuration.
type PluginConfig struct {
	Dynamic PluginDynamicConfig `json:"dynamic"` // Dynamic contains dynamic plugin storage settings.
	Runtime PluginDynamicConfig `json:"runtime"` // Runtime keeps legacy config compatibility for older runtime keys.
}

// PluginDynamicConfig holds dynamic plugin storage configuration.
type PluginDynamicConfig struct {
	StoragePath string `json:"storagePath"` // StoragePath is the directory used to discover and store dynamic wasm packages.
}

// GetPlugin reads plugin config from configuration file.
func (s *serviceImpl) GetPlugin(ctx context.Context) *PluginConfig {
	cfg := &PluginConfig{
		Dynamic: PluginDynamicConfig{
			StoragePath: "temp/output",
		},
	}
	mustScanConfig(ctx, "plugin", cfg)

	cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Dynamic.StoragePath)
	if cfg.Dynamic.StoragePath == "" {
		cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Runtime.StoragePath)
	}
	if cfg.Dynamic.StoragePath == "" {
		cfg.Dynamic.StoragePath = "temp/output"
	}
	return cfg
}

// GetPluginDynamicStoragePath returns the normalized dynamic wasm storage directory.
func (s *serviceImpl) GetPluginDynamicStoragePath(ctx context.Context) string {
	if override := getPluginDynamicStoragePathOverride(); override != "" {
		return override
	}
	return filepath.Clean(s.GetPlugin(ctx).Dynamic.StoragePath)
}

// SetPluginDynamicStoragePathOverride overrides the dynamic-plugin storage path.
// Tests use this to isolate runtime artifact discovery from the shared workspace.
func SetPluginDynamicStoragePathOverride(path string) {
	pluginDynamicStoragePathOverride.Store(strings.TrimSpace(path))
}

func getPluginDynamicStoragePathOverride() string {
	value := pluginDynamicStoragePathOverride.Load()
	if value == nil {
		return ""
	}
	path, ok := value.(string)
	if !ok {
		return ""
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}
