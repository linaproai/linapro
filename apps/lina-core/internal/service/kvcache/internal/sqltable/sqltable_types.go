// This file defines shared cache models used by the SQL table kvcache
// implementation.

package sqltable

import "github.com/gogf/gf/v2/os/gtime"

// OwnerType identifies the business scope that owns one cache entry.
type OwnerType string

// Cache owner type constants identify the supported cache-entry ownership
// scopes.
const (
	// OwnerTypePlugin identifies dynamic plugin-owned cache entries.
	OwnerTypePlugin OwnerType = "plugin"
	// OwnerTypeModule identifies host module-owned cache entries.
	OwnerTypeModule OwnerType = "module"
)

// Cache value kind constants describe whether one entry stores string or
// integer data.
const (
	// ValueKindString identifies string cache values.
	ValueKindString = 1
	// ValueKindInt identifies integer cache values.
	ValueKindInt = 2
)

// Cache-size constants bound the persisted identity and payload lengths for KV
// cache entries.
const (
	maxOwnerTypeBytes = 16
	maxOwnerKeyBytes  = 64
	maxNamespaceBytes = 64
	maxCacheKeyBytes  = 128
	maxValueBytes     = 4096
)

// Item defines one cache entry snapshot.
type Item struct {
	// Key is the logical cache key inside the namespace.
	Key string
	// ValueKind identifies whether the entry stores a string or integer value.
	ValueKind int
	// Value is the string payload of the cache entry.
	Value string
	// IntValue is the integer payload of the cache entry.
	IntValue int64
	// ExpireAt is the optional expiration time.
	ExpireAt *gtime.Time
}

// String returns the canonical owner type value.
func (value OwnerType) String() string {
	return string(value)
}
