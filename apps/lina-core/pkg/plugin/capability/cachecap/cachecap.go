// This file defines the source-plugin visible cache contract. The contract is
// intentionally narrower than the host kvcache service so plugins cannot
// control owner identities, backend providers, or internal cache keys.

package cachecap

import (
	"context"
	"time"
)

// Cache value kind constants describe the concrete payload representation
// stored in one plugin cache item.
const (
	// CacheValueKindString identifies string cache values.
	CacheValueKindString = 1
	// CacheValueKindInt identifies integer cache values.
	CacheValueKindInt = 2
	// MaxBatchKeys bounds one cache batch request by logical key count.
	MaxBatchKeys = 100
	// MaxKeyBytes bounds one cache logical key.
	MaxKeyBytes = 256
	// MaxBatchValueBytes bounds total string payload bytes in one cache batch write.
	MaxBatchValueBytes = 1 * 1024 * 1024
)

// CacheItem describes one source-plugin visible cache snapshot. Cache items
// are lossy runtime acceleration data and must not be used as authority for
// permissions, configuration, tenant boundaries, plugin state, or business
// records.
type CacheItem struct {
	// Key is the plugin-local logical cache key inside the namespace.
	Key string
	// ValueKind identifies whether this item stores a string or integer value.
	ValueKind int
	// Value is the string payload when ValueKind is CacheValueKindString.
	Value string
	// IntValue is the integer payload when ValueKind is CacheValueKindInt.
	IntValue int64
	// ExpireAt is the optional absolute expiration time; nil means no expiration.
	ExpireAt *time.Time
}

// GetManyInput carries one bounded multi-key cache read request.
type GetManyInput struct {
	// Namespace is the plugin-local cache namespace.
	Namespace string
	// Keys are plugin-local logical cache keys inside Namespace.
	Keys []string
}

// GetManyOutput carries one bounded multi-key cache read response.
type GetManyOutput struct {
	// Items contains found cache entries keyed by plugin-local logical key.
	Items map[string]*CacheItem
	// MissingKeys contains requested keys with no unexpired entry.
	MissingKeys []string
}

// SetManyItem describes one cache string value write.
type SetManyItem struct {
	// Key is the plugin-local logical cache key inside Namespace.
	Key string
	// Value is the string payload to store.
	Value string
	// TTL is the optional expiration duration. Zero means no expiration.
	TTL time.Duration
}

// SetManyInput carries one bounded multi-key cache write request.
type SetManyInput struct {
	// Namespace is the plugin-local cache namespace.
	Namespace string
	// Items are string values to store.
	Items []SetManyItem
}

// SetManyOutput carries one bounded multi-key cache write response.
type SetManyOutput struct {
	// Items contains latest cache entries keyed by plugin-local logical key.
	Items map[string]*CacheItem
}

// DeleteManyInput carries one bounded multi-key cache delete request.
type DeleteManyInput struct {
	// Namespace is the plugin-local cache namespace.
	Namespace string
	// Keys are plugin-local logical cache keys inside Namespace.
	Keys []string
}

// Service defines the governed cache operations published to source
// plugins. Implementations must bind calls to the current plugin ID and tenant
// scope before delegating to the host cache backend.
type Service interface {
	// Get returns one unexpired cache item from the plugin namespace.
	Get(ctx context.Context, namespace string, key string) (*CacheItem, bool, error)
	// GetMany returns unexpired cache items for an explicit bounded key set.
	GetMany(ctx context.Context, in GetManyInput) (*GetManyOutput, error)
	// Set stores a string value in the plugin namespace. ttl=0 means no expiration.
	Set(ctx context.Context, namespace string, key string, value string, ttl time.Duration) (*CacheItem, error)
	// SetMany stores string values in the plugin namespace. Per-item ttl=0 means no expiration.
	SetMany(ctx context.Context, in SetManyInput) (*SetManyOutput, error)
	// Delete removes one cache item. Deleting a missing item is a successful no-op.
	Delete(ctx context.Context, namespace string, key string) error
	// DeleteMany removes explicit cache keys. Deleting missing items is a successful no-op.
	DeleteMany(ctx context.Context, in DeleteManyInput) error
	// Incr increments one integer cache item by delta. ttl applies to new items
	// and preserves backend-specific existing expiration semantics otherwise.
	Incr(ctx context.Context, namespace string, key string, delta int64, ttl time.Duration) (*CacheItem, error)
	// Expire updates one cache item's expiration policy. ttl=0 clears expiration.
	Expire(ctx context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error)
}
