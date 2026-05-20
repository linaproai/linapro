// This file defines the source-plugin visible cache contract. The contract is
// intentionally narrower than the host kvcache service so plugins cannot
// control owner identities, backend providers, or internal cache keys.

package contract

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

// CacheService defines the governed cache operations published to source
// plugins. Implementations must bind calls to the current plugin ID and tenant
// scope before delegating to the host cache backend.
type CacheService interface {
	// Get returns one unexpired cache item from the plugin namespace.
	Get(ctx context.Context, namespace string, key string) (*CacheItem, bool, error)
	// Set stores a string value in the plugin namespace. ttl=0 means no expiration.
	Set(ctx context.Context, namespace string, key string, value string, ttl time.Duration) (*CacheItem, error)
	// Delete removes one cache item. Deleting a missing item is a successful no-op.
	Delete(ctx context.Context, namespace string, key string) error
	// Incr increments one integer cache item by delta. ttl applies to new items
	// and preserves backend-specific existing expiration semantics otherwise.
	Incr(ctx context.Context, namespace string, key string, delta int64, ttl time.Duration) (*CacheItem, error)
	// Expire updates one cache item's expiration policy. ttl=0 clears expiration.
	Expire(ctx context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error)
}
