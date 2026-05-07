// Package kvcache defines the host KV cache service facade and backend-agnostic
// service contract.
package kvcache

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
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

// Service defines the kvcache service contract.
type Service interface {
	// BackendName returns the concrete cache backend used by this service.
	BackendName() BackendName
	// RequiresExpiredCleanup reports whether the backend needs an external cleanup task.
	RequiresExpiredCleanup() bool
	// Get returns the current cache entry snapshot identified by ownerType and
	// one scoped cache key built by BuildCacheKey.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - *Item: the cache entry snapshot when the entry exists, including value kind, value, and expiration time.
	// - bool: whether the unexpired cache entry exists.
	// - error: returned when the scoped cache key is invalid or the database query fails.
	Get(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) (*Item, bool, error)
	// GetInt returns the current integer cache value identified by ownerType and
	// one scoped cache key built by BuildCacheKey.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - int64: the integer cache value when the entry exists and is stored as an integer.
	// - bool: whether the unexpired cache entry exists.
	// - error: returned when the scoped cache key is invalid, the existing entry is not stored
	// as an integer, or the database query fails.
	GetInt(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) (int64, bool, error)
	// Set stores or replaces a string cache value for the specified scoped cache key.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - value: string payload to persist in the cache entry.
	// - ttl: entry lifetime; 0 means never expire, and positive values create an absolute expiration time.
	// Returns:
	// - *Item: the latest cache entry snapshot after the value has been written successfully.
	// - error: returned when the scoped cache key is invalid, the value exceeds the allowed size,
	// ttl is negative or the upsert operation fails.
	Set(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		value string,
		ttl time.Duration,
	) (*Item, error)
	// Delete removes the cache entry identified by ownerType and one scoped cache key.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - error: returned when the scoped cache key is invalid or the delete statement fails.
	// Deleting a non-existent entry is treated as a successful no-op.
	Delete(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) error
	// Incr increments an integer cache value by delta and returns the updated entry snapshot.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - delta: increment amount added to the current integer value; when the entry does not exist, delta becomes the initial value.
	// - ttl: new entry lifetime; 0 keeps the entry non-expiring when creating a new item and preserves the current expiration when updating an existing item.
	// Returns:
	// - *Item: the latest cache entry snapshot after the increment succeeds.
	// - error: returned when the scoped cache key is invalid, ttl is negative,
	// expired-entry cleanup fails, the existing entry is not stored as an integer, or any database operation fails.
	Incr(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		delta int64,
		ttl time.Duration,
	) (*Item, error)
	// Expire updates the expiration policy of a cache entry without changing its value.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - ttl: new lifetime; 0 clears the expiration and makes the entry persistent.
	// Returns:
	// - bool: whether an existing cache entry was found and updated.
	// - *gtime.Time: the normalized absolute expiration time; nil means the entry will not expire.
	// - error: returned when the scoped cache key is invalid, ttl is negative,
	// expired-entry cleanup fails, or the database update fails.
	Expire(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		ttl time.Duration,
	) (bool, *gtime.Time, error)
	// CleanupExpired removes one bounded batch of cache entries whose expiration
	// time is earlier than the current time.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// Returns:
	// - error: returned when the cleanup delete statement fails. When no expired entries
	// exist, the method returns nil.
	CleanupExpired(ctx context.Context) error
}

// Interface compliance assertion for the default kvcache service
// implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	backend Backend
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
	ExpireAt *gtime.Time
}

// New creates and returns a new distributed KV cache service instance.
func New(options ...Option) Service {
	config := newServiceConfig()
	for _, option := range options {
		if option != nil {
			option(config)
		}
	}
	backend := config.backend
	if backend == nil && config.provider != nil {
		backend = config.provider.NewBackend()
	}
	if backend == nil {
		backend = NewSQLTableProvider().NewBackend()
	}
	return &serviceImpl{backend: backend}
}

// String returns the canonical owner type value.
func (value OwnerType) String() string {
	return string(value)
}

// TTLFromSeconds converts the plugin host-service wire TTL into the duration
// used by the backend-agnostic kvcache service contract.
func TTLFromSeconds(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}
