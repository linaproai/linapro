// Package kvcache defines the host KV cache service facade and backend-agnostic
// service contract.
package kvcache

import (
	"context"
	"time"

	"lina-core/internal/service/kvcache/internal/contract"
)

// OwnerType identifies the business scope that owns one cache entry.
type OwnerType = contract.OwnerType

// Cache owner type constants identify the supported cache-entry ownership
// scopes.
const (
	// OwnerTypePlugin identifies dynamic plugin-owned cache entries.
	OwnerTypePlugin = contract.OwnerTypePlugin
	// OwnerTypeModule identifies host module-owned cache entries.
	OwnerTypeModule = contract.OwnerTypeModule
)

// Cache value kind constants describe whether one entry stores string or
// integer data.
const (
	// ValueKindString identifies string cache values.
	ValueKindString = contract.ValueKindString
	// ValueKindInt identifies integer cache values.
	ValueKindInt = contract.ValueKindInt
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
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - *Item: the cache entry snapshot when the entry exists, including value kind, value, and expiration time.
	// - bool: whether the unexpired cache entry exists.
	// - error: returned when the scoped cache key is invalid or the backend read fails.
	Get(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) (*Item, bool, error)
	// GetInt returns the current integer cache value identified by ownerType and
	// one scoped cache key built by BuildCacheKey.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - int64: the integer cache value when the entry exists and is stored as an integer.
	// - bool: whether the unexpired cache entry exists.
	// - error: returned when the scoped cache key is invalid, the existing entry is not stored
	// as an integer, or the backend read fails.
	GetInt(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) (int64, bool, error)
	// Set stores or replaces a string cache value for the specified scoped cache key.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - value: string payload to store in the cache entry.
	// - ttl: positive entry lifetime; zero or negative values are rejected so cache entries always expire.
	// Returns:
	// - *Item: the latest cache entry snapshot after the value has been written successfully.
	// - error: returned when the scoped cache key is invalid, the value exceeds the allowed size,
	// ttl is not positive or the backend write fails.
	Set(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		value string,
		ttl time.Duration,
	) (*Item, error)
	// Delete removes the cache entry identified by ownerType and one scoped cache key.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// Returns:
	// - error: returned when the scoped cache key is invalid or the backend delete fails.
	// Deleting a non-existent entry is treated as a successful no-op.
	Delete(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
	) error
	// Incr increments an integer cache value by delta and returns the updated entry snapshot.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - delta: increment amount added to the current integer value; when the entry does not exist, delta becomes the initial value.
	// - ttl: positive entry lifetime applied to the resulting integer entry; zero or negative values are rejected.
	// Returns:
	// - *Item: the latest cache entry snapshot after the increment succeeds.
	// - error: returned when the scoped cache key is invalid, ttl is not positive,
	// the existing entry is not stored as an integer, or any backend operation fails.
	Incr(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		delta int64,
		ttl time.Duration,
	) (*Item, error)
	// Expire updates the expiration policy of a cache entry without changing its value.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
	// - ttl: positive new lifetime; zero or negative values are rejected.
	// Returns:
	// - bool: whether an existing cache entry was found and updated.
	// - *time.Time: the normalized absolute expiration time when an existing entry is updated.
	// - error: returned when the scoped cache key is invalid, ttl is not positive,
	// or the backend expiration update fails.
	Expire(
		ctx context.Context,
		ownerType OwnerType,
		cacheKey string,
		ttl time.Duration,
	) (bool, *time.Time, error)
	// CleanupExpired asks backends that need external expiration cleanup to
	// remove one bounded batch of expired cache entries. Backends with native TTL
	// support treat this as a no-op.
	// Parameters:
	// - ctx: request-scoped context used for backend access, tracing, and cancellation.
	// Returns:
	// - error: returned when the backend cleanup fails. Native-TTL no-op backends
	// return nil.
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
type Item = contract.Item

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
		backend = NewMemoryProvider().NewBackend()
	}
	return &serviceImpl{backend: backend}
}

// TTLFromSeconds converts the plugin host-service wire TTL into the duration
// used by the backend-agnostic kvcache service contract.
func TTLFromSeconds(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}
