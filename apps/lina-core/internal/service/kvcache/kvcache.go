// This file defines the distributed KV cache service component and shared value models.

package kvcache

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"
	// OwnerType defines the supported cache owner types.
)

type OwnerType string

const (
	// OwnerTypePlugin identifies dynamic plugin-owned cache entries.
	OwnerTypePlugin OwnerType = "plugin"
	// OwnerTypeModule identifies host module-owned cache entries.
	OwnerTypeModule OwnerType = "module"
)

const (
	// ValueKindString identifies string cache values.
	ValueKindString = 1
	// ValueKindInt identifies integer cache values.
	ValueKindInt = 2
)

const (
	maxOwnerTypeBytes = 16
	maxOwnerKeyBytes  = 64
	maxNamespaceBytes = 64
	maxCacheKeyBytes  = 128
	maxValueBytes     = 4096
)

// Service defines the kvcache service contract.
type Service interface {
	// Get returns the current cache entry snapshot identified by ownerType, ownerKey,
	// namespace, and cacheKey.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
	// - namespace: logical group name used to organize related cache entries for the same owner.
	// - cacheKey: concrete key to read inside the namespace.
	// Returns:
	// - *Item: the cache entry snapshot when the entry exists, including value kind, value, and expiration time.
	// - bool: whether the cache entry exists after expired data has been cleaned up.
	// - error: returned when identity parameters are invalid, expired-entry cleanup fails, or the database query fails.
	Get(
		ctx context.Context,
		ownerType OwnerType,
		ownerKey string,
		namespace string,
		cacheKey string,
	) (*Item, bool, error)
	// Set stores or replaces a string cache value for the specified owner, namespace,
	// and cache key.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
	// - namespace: logical group name used to organize related cache entries for the same owner.
	// - cacheKey: concrete key to write inside the namespace.
	// - value: string payload to persist in the cache entry.
	// - expireSeconds: entry lifetime in seconds; 0 means never expire, and positive values create an absolute expiration time.
	// Returns:
	// - *Item: the latest cache entry snapshot after the value has been written successfully.
	// - error: returned when identity parameters are invalid, the value exceeds the allowed size,
	// expireSeconds is negative, expired-entry cleanup fails, or the upsert operation fails.
	Set(
		ctx context.Context,
		ownerType OwnerType,
		ownerKey string,
		namespace string,
		cacheKey string,
		value string,
		expireSeconds int64,
	) (*Item, error)
	// Delete removes the cache entry identified by ownerType, ownerKey, namespace,
	// and cacheKey.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
	// - namespace: logical group name used to organize related cache entries for the same owner.
	// - cacheKey: concrete key to delete inside the namespace.
	// Returns:
	// - error: returned when identity parameters are invalid or the delete statement fails.
	// Deleting a non-existent entry is treated as a successful no-op.
	Delete(
		ctx context.Context,
		ownerType OwnerType,
		ownerKey string,
		namespace string,
		cacheKey string,
	) error
	// Incr increments an integer cache value by delta and returns the updated entry snapshot.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
	// - namespace: logical group name used to organize related cache entries for the same owner.
	// - cacheKey: concrete key to increment inside the namespace.
	// - delta: increment amount added to the current integer value; when the entry does not exist, delta becomes the initial value.
	// - expireSeconds: new entry lifetime in seconds; 0 keeps the entry non-expiring when creating a new item and preserves the current expiration when updating an existing item.
	// Returns:
	// - *Item: the latest cache entry snapshot after the increment succeeds.
	// - error: returned when identity parameters are invalid, expireSeconds is negative,
	// expired-entry cleanup fails, the existing entry is not stored as an integer, or any database operation fails.
	Incr(
		ctx context.Context,
		ownerType OwnerType,
		ownerKey string,
		namespace string,
		cacheKey string,
		delta int64,
		expireSeconds int64,
	) (*Item, error)
	// Expire updates the expiration policy of a cache entry without changing its value.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// - ownerType: cache owner category, used to isolate entries across different business scopes.
	// - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
	// - namespace: logical group name used to organize related cache entries for the same owner.
	// - cacheKey: concrete key whose expiration policy should be updated.
	// - expireSeconds: new lifetime in seconds; 0 clears the expiration and makes the entry persistent.
	// Returns:
	// - bool: whether an existing cache entry was found and updated.
	// - *gtime.Time: the normalized absolute expiration time; nil means the entry will not expire.
	// - error: returned when identity parameters are invalid, expireSeconds is negative,
	// expired-entry cleanup fails, or the database update fails.
	Expire(
		ctx context.Context,
		ownerType OwnerType,
		ownerKey string,
		namespace string,
		cacheKey string,
		expireSeconds int64,
	) (bool, *gtime.Time, error)
	// CleanupExpired removes all cache entries whose expiration time is earlier than
	// the current time.
	// Parameters:
	// - ctx: request-scoped context used for database access, tracing, and cancellation.
	// Returns:
	// - error: returned when the cleanup delete statement fails. When no expired entries
	// exist, the method returns nil.
	CleanupExpired(ctx context.Context) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

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
func New() Service {
	return &serviceImpl{}
}

// String returns the canonical owner type value.
func (value OwnerType) String() string {
	return string(value)
}
