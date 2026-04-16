// This file defines the plugin type enumeration and normalization helpers.

package catalog

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
func (t PluginType) String() string { return string(t) }

// NormalizeType converts a raw type string to the canonical PluginType.
// Unknown values default to TypeSource for backward compatibility.
func NormalizeType(value string) PluginType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case TypeDynamic.String():
		return TypeDynamic
	default:
		return TypeSource
	}
}

// IsSupportedType reports whether the given type string is a recognized plugin type.
func IsSupportedType(value string) bool {
	t := NormalizeType(value)
	return t == TypeSource || t == TypeDynamic
}
