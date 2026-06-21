// Package contract defines the narrow shared kvcache backend contract used by
// the public service facade and internal backend implementations.
package contract

import (
	"context"
	"time"
)

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

// BackendName identifies one kvcache backend implementation.
type BackendName string

// Supported backend names for the kvcache service.
const (
	// BackendMemory stores cache entries in the host process memory.
	BackendMemory BackendName = "memory"
	// BackendCoordinationKV stores cache entries in the configured coordination KV backend.
	BackendCoordinationKV BackendName = "coordination-kv"
)

// Backend defines the backend-specific KV cache operations used by the public
// kvcache service facade.
type Backend interface {
	// Name returns the stable backend implementation name.
	Name() BackendName
	// RequiresExpiredCleanup reports whether the backend needs external cleanup
	// for expired entries.
	RequiresExpiredCleanup() bool
	// Get returns the current cache entry snapshot identified by ownerType and cacheKey.
	Get(ctx context.Context, ownerType OwnerType, cacheKey string) (*Item, bool, error)
	// GetInt returns the current integer cache value identified by ownerType and cacheKey.
	GetInt(ctx context.Context, ownerType OwnerType, cacheKey string) (int64, bool, error)
	// Set stores or replaces a string cache value with a backend-neutral TTL.
	Set(ctx context.Context, ownerType OwnerType, cacheKey string, value string, ttl time.Duration) (*Item, error)
	// Delete removes the cache entry identified by ownerType and cacheKey.
	Delete(ctx context.Context, ownerType OwnerType, cacheKey string) error
	// Incr increments an integer cache value by delta using a backend-neutral TTL.
	Incr(ctx context.Context, ownerType OwnerType, cacheKey string, delta int64, ttl time.Duration) (*Item, error)
	// Expire updates the expiration policy of a cache entry.
	Expire(ctx context.Context, ownerType OwnerType, cacheKey string, ttl time.Duration) (bool, *time.Time, error)
	// CleanupExpired removes expired entries when the backend needs external cleanup.
	CleanupExpired(ctx context.Context) error
}

// Provider creates one kvcache backend implementation.
type Provider interface {
	// NewBackend returns one backend instance for the provider implementation.
	NewBackend() Backend
}

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
	ExpireAt *time.Time
}
