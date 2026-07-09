// Package apitime contains HTTP API response timestamp projection helpers for
// host and source-plugin DTO mapping. It keeps internal GoFrame time values out
// of public JSON contracts while preserving nil semantics for optional fields.
package apitime

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
)

// Milli converts a standard-library time pointer to a Unix timestamp in milliseconds.
// Nil or zero values are projected as nil so API responses can distinguish
// absent timestamps from the Unix epoch.
func Milli(value *time.Time) *int64 {
	if value == nil || value.IsZero() {
		return nil
	}
	millis := value.UnixMilli()
	return &millis
}

// MilliFromTime converts a standard-library time value to a Unix timestamp in
// milliseconds. Zero values are projected as nil for optional response fields.
func MilliFromTime(value time.Time) *int64 {
	if value.IsZero() {
		return nil
	}
	millis := value.UnixMilli()
	return &millis
}

// MilliFromString converts a stored textual timestamp into a Unix timestamp in
// milliseconds. Empty or unparsable values are projected as nil.
func MilliFromString(value string) *int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := gtime.StrToTime(trimmed)
	if err != nil {
		return nil
	}
	return MilliFromTime(parsed.Time)
}
