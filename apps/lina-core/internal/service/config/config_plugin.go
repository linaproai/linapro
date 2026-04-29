// This file defines plugin storage-path configuration loading and test-time
// storage-path overrides.

package config

import (
	"context"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Plugin config defaults used when config.yaml omits plugin.dynamic.
const defaultPluginDynamicStoragePath = "temp/output"

// pluginDynamicStoragePathOverride stores an optional process-wide test
// override for the dynamic plugin storage root.
var pluginDynamicStoragePathOverride atomic.Value

// pluginAutoEnableOverrideState stores one optional process-wide test override
// for the startup auto-enable plugin entry list.
type pluginAutoEnableOverrideState struct {
	set   bool
	value []PluginAutoEnableEntry
}

// pluginAutoEnableOverride stores an optional process-wide test override for
// startup auto-enable entries.
var pluginAutoEnableOverride atomic.Value

// PluginAutoEnableEntry represents one normalized entry of plugin.autoEnable.
// The YAML schema enforces a single structured object form per entry:
//
//	- id: "plugin-demo-source"
//	  withMockData: false
//
// withMockData defaults to false when omitted. Operators must explicitly opt
// in to mock data per plugin so auto-installed plugins do not accidentally
// ship demo data into production environments. Bare-string entries are
// rejected at config load to keep the host configuration shape uniform and
// the operator intent explicit.
type PluginAutoEnableEntry struct {
	// ID is the plugin identifier the host should ensure is installed and enabled.
	ID string `json:"id"`
	// WithMockData enables the optional mock-data load phase during the
	// startup auto-install flow for this entry. Defaults to false.
	WithMockData bool `json:"withMockData"`
}

// PluginConfig holds plugin-related host configuration.
type PluginConfig struct {
	Dynamic PluginDynamicConfig `json:"dynamic"` // Dynamic contains dynamic plugin storage settings.
	Runtime PluginDynamicConfig `json:"runtime"` // Runtime keeps legacy config compatibility for older runtime keys.
	// AutoEnable lists plugin entries the host must auto-install and enable
	// during startup. Populated manually from g.Cfg() rather than via the
	// generic scan pipeline because the YAML schema accepts a mix of bare
	// string IDs and {id, withMockData} objects per entry. The struct tags
	// instruct gconv to skip this field during automatic scan.
	AutoEnable []PluginAutoEnableEntry `c:"-" gconv:"-" json:"-"`
}

// PluginDynamicConfig holds dynamic plugin storage configuration.
type PluginDynamicConfig struct {
	StoragePath string `json:"storagePath"` // StoragePath is the directory used to discover and store dynamic wasm packages.
}

// GetPlugin reads plugin config from configuration file. Sub-keys are scanned
// individually so the autoEnable union schema (mixed string and object items)
// can be parsed by readRawPluginAutoEnableEntries without fighting the
// generic scan pipeline's strict type expectations.
//
// Validation errors from autoEnable parsing fail-fast at the cache-load boundary:
// helpers return errors and this function panics ONCE so the startup process
// terminates with a clear message before dependent components run. Sibling
// helpers stay error-returning so they remain unit-testable.
func (s *serviceImpl) GetPlugin(ctx context.Context) *PluginConfig {
	cfg := clonePluginConfig(processStaticConfigCaches.plugin.load(func() *PluginConfig {
		cfg := &PluginConfig{
			Dynamic: PluginDynamicConfig{
				StoragePath: defaultPluginDynamicStoragePath,
			},
		}
		mustScanConfig(ctx, "plugin.dynamic", &cfg.Dynamic)
		mustScanConfig(ctx, "plugin.runtime", &cfg.Runtime)

		cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Dynamic.StoragePath)
		if cfg.Dynamic.StoragePath == "" {
			cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Runtime.StoragePath)
		}
		if cfg.Dynamic.StoragePath == "" {
			cfg.Dynamic.StoragePath = defaultPluginDynamicStoragePath
		}
		rawEntries, err := readRawPluginAutoEnableEntries(ctx)
		if err != nil {
			panic(gerror.Wrap(err, "load config plugin.autoEnable failed"))
		}
		normalized, err := normalizePluginAutoEnableEntries(rawEntries)
		if err != nil {
			panic(gerror.Wrap(err, "normalize config plugin.autoEnable failed"))
		}
		cfg.AutoEnable = normalized
		return cfg
	}))
	if override, ok := getPluginAutoEnableOverride(); ok {
		cfg.AutoEnable = override
	}
	return cfg
}

// GetPluginAutoEnable returns the configured startup auto-enable plugin IDs
// (without the mock-data opt-in flag). Used by callers that only need the ID
// list — e.g., the controller that builds the AutoEnableManaged map for the
// management UI's startup-managed badge.
func (s *serviceImpl) GetPluginAutoEnable(ctx context.Context) []string {
	entries := s.GetPluginAutoEnableEntries(ctx)
	if len(entries) == 0 {
		return nil
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids
}

// GetPluginAutoEnableEntries returns the full normalized startup auto-enable
// entries including the per-entry WithMockData opt-in flag. Used by the
// startup bootstrap flow to decide whether each plugin should also load its
// mock-data SQL during the startup auto-install.
func (s *serviceImpl) GetPluginAutoEnableEntries(ctx context.Context) []PluginAutoEnableEntry {
	cfg := s.GetPlugin(ctx)
	if cfg == nil || len(cfg.AutoEnable) == 0 {
		return nil
	}
	out := make([]PluginAutoEnableEntry, len(cfg.AutoEnable))
	copy(out, cfg.AutoEnable)
	return out
}

// GetPluginDynamicStoragePath returns the runtime-resolved dynamic wasm storage
// directory. Relative paths are anchored at the repository root when available.
func (s *serviceImpl) GetPluginDynamicStoragePath(ctx context.Context) string {
	if override := getPluginDynamicStoragePathOverride(); override != "" {
		return resolveRuntimePathWithDefault(override, defaultPluginDynamicStoragePath)
	}
	return resolveRuntimePathWithDefault(s.GetPlugin(ctx).Dynamic.StoragePath, defaultPluginDynamicStoragePath)
}

// SetPluginDynamicStoragePathOverride overrides the dynamic-plugin storage path.
// Tests use this to isolate runtime artifact discovery from the shared workspace.
func SetPluginDynamicStoragePathOverride(path string) {
	pluginDynamicStoragePathOverride.Store(strings.TrimSpace(path))
}

// SetPluginAutoEnableOverride overrides the startup auto-enable plugin IDs.
// Tests use this to isolate startup bootstrap behavior from shared config
// adapter content. Bare string IDs default to WithMockData=false to preserve
// the legacy override semantics from before the union schema was introduced.
//
// Tests pass already-validated IDs, so the underlying normalization should not
// fail; if it does, the test setup is itself broken and the panic from
// gerror.Must surfaces the same fail-fast behavior expected of test fixtures.
func SetPluginAutoEnableOverride(pluginIDs []string) {
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
		panic(gerror.Wrap(err, "SetPluginAutoEnableOverride received invalid plugin IDs"))
	}
	pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{
		set:   true,
		value: normalized,
	})
}

// SetPluginAutoEnableEntriesOverride overrides the startup auto-enable plugin
// entries with the union schema's full per-entry payload. Tests that exercise
// the mock-data opt-in flow use this variant; tests that only care about ID
// normalization can keep using SetPluginAutoEnableOverride.
func SetPluginAutoEnableEntriesOverride(entries []PluginAutoEnableEntry) {
	if len(entries) == 0 {
		pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{})
		return
	}
	normalized, err := normalizePluginAutoEnableEntries(entries)
	if err != nil {
		panic(gerror.Wrap(err, "SetPluginAutoEnableEntriesOverride received invalid entries"))
	}
	pluginAutoEnableOverride.Store(pluginAutoEnableOverrideState{
		set:   true,
		value: normalized,
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
	return cleanConfigPath(path)
}

// getPluginAutoEnableOverride returns the normalized test override when set.
func getPluginAutoEnableOverride() ([]PluginAutoEnableEntry, bool) {
	value := pluginAutoEnableOverride.Load()
	if value == nil {
		return nil, false
	}
	state, ok := value.(pluginAutoEnableOverrideState)
	if !ok || !state.set {
		return nil, false
	}
	if len(state.value) == 0 {
		return []PluginAutoEnableEntry{}, true
	}
	return append([]PluginAutoEnableEntry(nil), state.value...), true
}

// readRawPluginAutoEnableEntries decodes plugin.autoEnable from the raw config
// value into the union-schema entry slice. Bare string elements normalize into
// {ID, WithMockData=false}; object elements scan their id and withMockData
// fields. Returns a typed error on shape violations so callers can decide how
// to react (the cache-load closure in GetPlugin promotes the error to a single
// startup-time panic; tests can recover from the panic to assert messages).
func readRawPluginAutoEnableEntries(ctx context.Context) ([]PluginAutoEnableEntry, error) {
	value := g.Cfg().MustGet(ctx, "plugin.autoEnable")
	if value == nil || value.IsEmpty() {
		return nil, nil
	}
	rawValue := value.Val()
	if rawValue == nil {
		return nil, nil
	}
	rawKind := reflect.TypeOf(rawValue).Kind()
	if rawKind != reflect.Slice && rawKind != reflect.Array {
		return nil, gerror.New("config plugin.autoEnable must be an array")
	}
	rawSlice := reflect.ValueOf(rawValue)
	entries := make([]PluginAutoEnableEntry, 0, rawSlice.Len())
	for index := 0; index < rawSlice.Len(); index++ {
		itemValue := rawSlice.Index(index)
		for itemValue.IsValid() && itemValue.Kind() == reflect.Interface {
			if itemValue.IsNil() {
				return nil, gerror.Newf("config plugin.autoEnable item %d cannot be nil", index+1)
			}
			itemValue = itemValue.Elem()
		}
		if !itemValue.IsValid() {
			return nil, gerror.Newf("config plugin.autoEnable item %d cannot be nil", index+1)
		}
		if itemValue.Kind() != reflect.Map {
			return nil, gerror.Newf(
				"config plugin.autoEnable item %d must be a {id, withMockData} object, got %s",
				index+1, itemValue.Kind().String(),
			)
		}
		entry, err := decodePluginAutoEnableEntryMap(index, itemValue)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// decodePluginAutoEnableEntryMap converts one object-form plugin.autoEnable
// item into the canonical PluginAutoEnableEntry. Keys are matched by their
// canonical lowercase YAML names ("id", "withmockdata"). Missing or extra
// keys, wrong types, or empty IDs surface as errors with the offending item
// index so callers can produce a clear failure message.
func decodePluginAutoEnableEntryMap(index int, value reflect.Value) (PluginAutoEnableEntry, error) {
	if value.Type().Key().Kind() != reflect.String && value.Type().Key().Kind() != reflect.Interface {
		return PluginAutoEnableEntry{}, gerror.Newf(
			"config plugin.autoEnable item %d object keys must be strings",
			index+1,
		)
	}
	entry := PluginAutoEnableEntry{}
	for _, keyValue := range value.MapKeys() {
		keyName := strings.ToLower(strings.TrimSpace(asKeyString(keyValue)))
		mapValue := value.MapIndex(keyValue)
		for mapValue.IsValid() && mapValue.Kind() == reflect.Interface {
			if mapValue.IsNil() {
				mapValue = reflect.Value{}
				break
			}
			mapValue = mapValue.Elem()
		}
		switch keyName {
		case "id":
			if !mapValue.IsValid() || mapValue.Kind() != reflect.String {
				return PluginAutoEnableEntry{}, gerror.Newf(
					"config plugin.autoEnable item %d field id must be a string",
					index+1,
				)
			}
			entry.ID = mapValue.String()
		case "withmockdata":
			if !mapValue.IsValid() || mapValue.Kind() != reflect.Bool {
				return PluginAutoEnableEntry{}, gerror.Newf(
					"config plugin.autoEnable item %d field withMockData must be a boolean",
					index+1,
				)
			}
			entry.WithMockData = mapValue.Bool()
		default:
			return PluginAutoEnableEntry{}, gerror.Newf(
				"config plugin.autoEnable item %d contains unsupported field %q (allowed: id, withMockData)",
				index+1, keyName,
			)
		}
	}
	if strings.TrimSpace(entry.ID) == "" {
		return PluginAutoEnableEntry{}, gerror.Newf(
			"config plugin.autoEnable item %d field id cannot be empty",
			index+1,
		)
	}
	return entry, nil
}

// asKeyString returns the string form of a reflect.Value used as a map key,
// supporting both reflect.String and reflect.Interface wrappers.
func asKeyString(keyValue reflect.Value) string {
	for keyValue.IsValid() && keyValue.Kind() == reflect.Interface {
		if keyValue.IsNil() {
			return ""
		}
		keyValue = keyValue.Elem()
	}
	if !keyValue.IsValid() {
		return ""
	}
	if keyValue.Kind() == reflect.String {
		return keyValue.String()
	}
	return ""
}

// normalizePluginAutoEnableEntries trims, validates, and de-duplicates startup
// auto-enable entries while preserving declaration order. The first occurrence
// of a given plugin ID wins; subsequent duplicates are silently dropped to
// match the legacy ID-only behavior. Returns an error when an entry has an
// empty ID so callers can decide how to surface the failure.
func normalizePluginAutoEnableEntries(entries []PluginAutoEnableEntry) ([]PluginAutoEnableEntry, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	var (
		normalized = make([]PluginAutoEnableEntry, 0, len(entries))
		seen       = make(map[string]struct{}, len(entries))
	)
	for index, entry := range entries {
		trimmedID := strings.TrimSpace(entry.ID)
		if trimmedID == "" {
			return nil, gerror.Newf("config plugin.autoEnable item %d cannot be empty", index+1)
		}
		if _, ok := seen[trimmedID]; ok {
			continue
		}
		seen[trimmedID] = struct{}{}
		normalized = append(normalized, PluginAutoEnableEntry{
			ID:           trimmedID,
			WithMockData: entry.WithMockData,
		})
	}
	return normalized, nil
}

