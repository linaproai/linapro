// This file defines plugin storage-path configuration loading and test-time
// storage-path overrides.

package config

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"
)

// pluginDynamicStoragePathOverride stores an optional process-wide test
// override for the dynamic plugin storage root.
var pluginDynamicStoragePathOverride atomic.Value

// pluginAutoEnableOverrideState stores one optional process-wide test override
// for the startup auto-enable plugin ID list.
type pluginAutoEnableOverrideState struct {
	set   bool
	value []string
}

// pluginAutoEnableOverride stores an optional process-wide test override for
// startup auto-enable plugin IDs.
var pluginAutoEnableOverride atomic.Value

// PluginConfig holds plugin-related host configuration.
type PluginConfig struct {
	Dynamic    PluginDynamicConfig `json:"dynamic"`    // Dynamic contains dynamic plugin storage settings.
	Runtime    PluginDynamicConfig `json:"runtime"`    // Runtime keeps legacy config compatibility for older runtime keys.
	AutoEnable []string            `json:"autoEnable"` // AutoEnable lists plugin IDs that must be auto-enabled during host startup.
}

// PluginDynamicConfig holds dynamic plugin storage configuration.
type PluginDynamicConfig struct {
	StoragePath string `json:"storagePath"` // StoragePath is the directory used to discover and store dynamic wasm packages.
}

// GetPlugin reads plugin config from configuration file.
func (s *serviceImpl) GetPlugin(ctx context.Context) *PluginConfig {
	return clonePluginConfig(processStaticConfigCaches.plugin.load(func() *PluginConfig {
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
		cfg.AutoEnable = normalizePluginAutoEnableIDs(cfg.AutoEnable)
		if override, ok := getPluginAutoEnableOverride(); ok {
			cfg.AutoEnable = override
		}
		return cfg
	}))
}

// GetPluginAutoEnable returns the configured startup auto-enable plugin IDs.
func (s *serviceImpl) GetPluginAutoEnable(ctx context.Context) []string {
	cfg := s.GetPlugin(ctx)
	if cfg == nil || len(cfg.AutoEnable) == 0 {
		return nil
	}
	return append([]string(nil), cfg.AutoEnable...)
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

// SetPluginAutoEnableOverride overrides the startup auto-enable plugin IDs.
// Tests use this to isolate startup bootstrap behavior from shared config
// adapter content.
func SetPluginAutoEnableOverride(pluginIDs []string) {
	if len(pluginIDs) == 0 {
		pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{})
		return
	}
	pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{
		set:   true,
		value: normalizePluginAutoEnableIDs(pluginIDs),
	})
}

// getPluginDynamicStoragePathOverride returns the normalized test override when set.
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

// getPluginAutoEnableOverride returns the normalized test override when set.
func getPluginAutoEnableOverride() ([]string, bool) {
	value := pluginAutoEnableOverride.Load()
	if value == nil {
		return nil, false
	}
	state, ok := value.(pluginAutoEnableOverrideState)
	if !ok || !state.set {
		return nil, false
	}
	if len(state.value) == 0 {
		return []string{}, true
	}
	return append([]string(nil), state.value...), true
}

// normalizePluginAutoEnableIDs trims, validates, and de-duplicates startup
// auto-enable plugin IDs while preserving declaration order.
func normalizePluginAutoEnableIDs(pluginIDs []string) []string {
	if len(pluginIDs) == 0 {
		return nil
	}

	var (
		normalized = make([]string, 0, len(pluginIDs))
		seen       = make(map[string]struct{}, len(pluginIDs))
	)
	for index, pluginID := range pluginIDs {
		trimmedID := strings.TrimSpace(pluginID)
		if trimmedID == "" {
			panic(gerror.Newf("配置 plugin.autoEnable 第 %d 项不能为空", index+1))
		}
		if _, ok := seen[trimmedID]; ok {
			continue
		}
		seen[trimmedID] = struct{}{}
		normalized = append(normalized, trimmedID)
	}
	return normalized
}
