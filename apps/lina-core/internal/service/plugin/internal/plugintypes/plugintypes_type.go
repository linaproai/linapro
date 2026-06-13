// This file defines the plugin type enumeration and normalization helpers.

package plugintypes

import "strings"

// PluginType defines the recognized plugin types.
type PluginType string

const (
	// TypeSource identifies a compiled-in source plugin.
	TypeSource PluginType = "source"
	// TypeDynamic identifies a runtime-loaded WASM plugin.
	TypeDynamic PluginType = "dynamic"
)

// String returns the canonical type value.
func (value PluginType) String() string { return string(value) }

// NormalizeType converts a raw type string to the canonical PluginType.
func NormalizeType(value string) PluginType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case TypeSource.String():
		return TypeSource
	case TypeDynamic.String():
		return TypeDynamic
	default:
		return PluginType(strings.TrimSpace(strings.ToLower(value)))
	}
}

// IsSupportedType reports whether the given type string is a recognized plugin type.
func IsSupportedType(value string) bool {
	pluginType := NormalizeType(value)
	return pluginType == TypeSource || pluginType == TypeDynamic
}
