// Package listorder defines canonical list-query order directions shared by
// public API DTOs and backend query adapters. Keep this package free of
// database dependencies so API contracts can reuse the values without importing
// storage helpers.
package listorder

import "strings"

// Direction represents one normalized list-query sort direction.
type Direction string

// Supported list-query order direction constants.
const (
	// ASC sorts records in ascending order.
	ASC Direction = "asc"
	// DESC sorts records in descending order.
	DESC Direction = "desc"
)

// String returns the canonical serialized value.
func (direction Direction) String() string {
	return string(direction)
}

// IsSupported reports whether direction is one of the published values.
func (direction Direction) IsSupported() bool {
	return direction == ASC || direction == DESC
}

// Normalize converts raw caller input into one canonical direction value.
func Normalize(value string) Direction {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ASC.String():
		return ASC
	case DESC.String():
		return DESC
	default:
		return ""
	}
}

// NormalizeOrDefault converts raw caller input into one canonical value and
// returns fallback when the input is empty or unsupported.
func NormalizeOrDefault(value string, fallback Direction) Direction {
	normalized := Normalize(value)
	if normalized.IsSupported() {
		return normalized
	}
	if fallback.IsSupported() {
		return fallback
	}
	return DESC
}
