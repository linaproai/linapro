// This file implements the single-node in-process kvcache backend using
// GoFrame gcache. The backend is intentionally lossy: process restart or TTL
// expiration turns entries into cache misses rather than restoring state from
// the database.

package memory

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/service/kvcache/internal/contract"
	"lina-core/pkg/bizerr"
)

// memoryProvider creates in-process memory backends for single-node topology.
type memoryProvider struct{}

// memoryBackend adapts GoFrame gcache to the kvcache memory Backend contract.
type memoryBackend struct {
	cache *gcache.Cache
	mu    sync.Mutex
}

// memoryValue stores one typed kvcache value in the in-process cache.
type memoryValue struct {
	Kind     int
	Value    string
	IntValue int64
	Key      string
	ExpireAt *time.Time
}

// NewProvider returns the single-node in-process kvcache provider.
func NewProvider() contract.Provider {
	return memoryProvider{}
}

// NewBackend creates one memory backend instance.
func (p memoryProvider) NewBackend() contract.Backend {
	return &memoryBackend{cache: gcache.New()}
}

// Name returns the stable memory backend name.
func (b *memoryBackend) Name() contract.BackendName {
	return contract.BackendMemory
}

// RequiresExpiredCleanup reports that the memory backend expires entries natively.
func (b *memoryBackend) RequiresExpiredCleanup() bool {
	return false
}

// Get returns one unexpired cache item from process memory.
func (b *memoryBackend) Get(
	ctx context.Context,
	ownerType contract.OwnerType,
	cacheKey string,
) (*contract.Item, bool, error) {
	identity, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return nil, false, err
	}
	value, ok, err := b.getValue(ctx, backendKey)
	if err != nil || !ok {
		return nil, ok, err
	}
	return memoryValueToItem(identity.CacheKey(), value), true, nil
}

// GetInt returns one integer cache value from process memory.
func (b *memoryBackend) GetInt(ctx context.Context, ownerType contract.OwnerType, cacheKey string) (int64, bool, error) {
	item, ok, err := b.Get(ctx, ownerType, cacheKey)
	if err != nil || !ok {
		return 0, ok, err
	}
	if item.ValueKind != contract.ValueKindInt {
		return 0, false, bizerr.NewCode(contract.CodeKVCacheValueNotInteger)
	}
	return item.IntValue, true, nil
}

// Set stores or replaces one string cache value in process memory.
func (b *memoryBackend) Set(
	ctx context.Context,
	ownerType contract.OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*contract.Item, error) {
	identity, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return nil, err
	}
	if err = contract.ValidatePositiveTTL(ttl); err != nil {
		return nil, err
	}
	if err = contract.ValidateMaxByteLength("value", value, contract.MaxValueBytes); err != nil {
		return nil, err
	}
	payload := &memoryValue{
		Kind:     contract.ValueKindString,
		Value:    value,
		Key:      identity.CacheKey(),
		ExpireAt: contract.ExpireAtFromTTL(ttl),
	}
	if err = b.cache.Set(ctx, backendKey, payload, ttl); err != nil {
		return nil, err
	}
	return memoryValueToItem(identity.CacheKey(), payload), nil
}

// Delete removes one cache entry from process memory.
func (b *memoryBackend) Delete(ctx context.Context, ownerType contract.OwnerType, cacheKey string) error {
	_, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return err
	}
	_, err = b.cache.Remove(ctx, backendKey)
	return err
}

// Incr increments one integer cache value atomically within the current process.
func (b *memoryBackend) Incr(
	ctx context.Context,
	ownerType contract.OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*contract.Item, error) {
	identity, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return nil, err
	}
	if err = contract.ValidatePositiveTTL(ttl); err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	value, ok, err := b.getValue(ctx, backendKey)
	if err != nil {
		return nil, err
	}
	if ok && value.Kind != contract.ValueKindInt {
		return nil, bizerr.NewCode(contract.CodeKVCacheIncrementValueNotInteger)
	}

	next := delta
	expireAt := contract.ExpireAtFromTTL(ttl)
	if ok {
		next = value.IntValue + delta
	}

	payload := &memoryValue{
		Kind:     contract.ValueKindInt,
		IntValue: next,
		Key:      identity.CacheKey(),
		ExpireAt: expireAt,
	}
	if err = b.cache.Set(ctx, backendKey, payload, ttl); err != nil {
		return nil, err
	}
	return memoryValueToItem(identity.CacheKey(), payload), nil
}

// Expire updates the expiration policy of one cache entry.
func (b *memoryBackend) Expire(
	ctx context.Context,
	ownerType contract.OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *time.Time, error) {
	_, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return false, nil, err
	}
	if err = contract.ValidatePositiveTTL(ttl); err != nil {
		return false, nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	value, ok, err := b.getValue(ctx, backendKey)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}
	expireAt := contract.ExpireAtFromTTL(ttl)
	next := *value
	next.ExpireAt = expireAt
	if err = b.cache.Set(ctx, backendKey, &next, ttl); err != nil {
		return false, nil, err
	}
	return true, expireAt, nil
}

// CleanupExpired is a no-op because the memory backend handles expiration internally.
func (b *memoryBackend) CleanupExpired(_ context.Context) error {
	return nil
}

// resolveBackendKey validates one public cache key and maps it to a unique
// process-local key.
func (b *memoryBackend) resolveBackendKey(
	ownerType contract.OwnerType,
	cacheKey string,
) (*contract.Identity, string, error) {
	if b == nil || b.cache == nil {
		return nil, "", bizerr.NewCode(contract.CodeKVCacheKeyInvalid)
	}
	identity, err := contract.ResolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, "", err
	}
	return identity, ownerType.String() + "." + cacheKey, nil
}

// getValue loads and type-checks one memory payload.
func (b *memoryBackend) getValue(ctx context.Context, backendKey string) (*memoryValue, bool, error) {
	raw, err := b.cache.Get(ctx, backendKey)
	if err != nil || raw == nil {
		return nil, false, err
	}
	value, ok := raw.Val().(*memoryValue)
	if !ok || value == nil {
		return nil, false, bizerr.NewCode(contract.CodeKVCacheKeyInvalid)
	}
	return value, true, nil
}

// memoryValueToItem maps one memory payload into a public cache item.
func memoryValueToItem(key string, value *memoryValue) *contract.Item {
	item := &contract.Item{
		Key:      key,
		ExpireAt: contract.CloneTime(value.ExpireAt),
	}
	if value.Kind == contract.ValueKindInt {
		item.ValueKind = contract.ValueKindInt
		item.IntValue = value.IntValue
		return item
	}
	item.ValueKind = contract.ValueKindString
	item.Value = value.Value
	return item
}
