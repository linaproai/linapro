// This file implements a coordination KV-backed kvcache provider through the
// shared coordination KVStore abstraction.

package coordkv

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache/internal/contract"
	"lina-core/pkg/bizerr"
)

// Coordination KV-backed kvcache constants.
const (
	coordKVValueKindString = "string"
	coordKVValueKindInt    = "int"
)

// coordKVProvider adapts coordination KVStore to kvcache Provider.
type coordKVProvider struct {
	coordinationSvc coordination.Service
}

// coordKVBackend adapts coordination KVStore to kvcache Backend.
type coordKVBackend struct {
	coordinationSvc coordination.Service
	kvStore         coordination.KVStore
	keyBuilder      *coordination.KeyBuilder
}

// coordKVValue stores one typed kvcache payload in coordination KV.
type coordKVValue struct {
	Kind     string `json:"kind"`
	Value    string `json:"value,omitempty"`
	IntValue int64  `json:"intValue,omitempty"`
}

// NewProvider returns a kvcache backend provider backed by coordination KVStore.
func NewProvider(coordinationSvc coordination.Service) contract.Provider {
	return &coordKVProvider{coordinationSvc: coordinationSvc}
}

// NewBackend returns one coordination KV kvcache backend instance.
func (p *coordKVProvider) NewBackend() contract.Backend {
	if p == nil || p.coordinationSvc == nil {
		return nil
	}
	return &coordKVBackend{
		coordinationSvc: p.coordinationSvc,
		kvStore:         p.coordinationSvc.KV(),
		keyBuilder:      p.coordinationSvc.KeyBuilder(),
	}
}

// Name returns the stable coordination KV backend name.
func (b *coordKVBackend) Name() contract.BackendName {
	return contract.BackendCoordinationKV
}

// RequiresExpiredCleanup reports that coordination KV expires entries natively.
func (b *coordKVBackend) RequiresExpiredCleanup() bool {
	return false
}

// Get returns the current cache entry snapshot identified by ownerType and cacheKey.
func (b *coordKVBackend) Get(
	ctx context.Context,
	ownerType contract.OwnerType,
	cacheKey string,
) (*contract.Item, bool, error) {
	identity, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return nil, false, err
	}
	rawValue, ok, err := b.kvStore.Get(ctx, backendKey)
	if err != nil || !ok {
		return nil, ok, err
	}
	payload, err := decodeCoordKVValue(rawValue)
	if err != nil {
		return nil, false, err
	}
	ttl, err := b.kvStore.TTL(ctx, backendKey)
	if err != nil {
		return nil, false, err
	}
	return coordKVPayloadToItem(identity.CacheKey(), payload, ttl), true, nil
}

// GetInt returns the current integer cache value identified by ownerType and cacheKey.
func (b *coordKVBackend) GetInt(ctx context.Context, ownerType contract.OwnerType, cacheKey string) (int64, bool, error) {
	item, ok, err := b.Get(ctx, ownerType, cacheKey)
	if err != nil || !ok {
		return 0, ok, err
	}
	if item.ValueKind != contract.ValueKindInt {
		return 0, false, bizerr.NewCode(contract.CodeKVCacheValueNotInteger)
	}
	return item.IntValue, true, nil
}

// Set stores or replaces a string cache value with a backend-neutral TTL.
func (b *coordKVBackend) Set(
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
	payload := coordKVValue{Kind: coordKVValueKindString, Value: value}
	encoded, err := encodeCoordKVValue(payload)
	if err != nil {
		return nil, err
	}
	if err = b.kvStore.Set(ctx, backendKey, encoded, ttl); err != nil {
		return nil, err
	}
	return &contract.Item{
		Key:       identity.CacheKey(),
		ValueKind: contract.ValueKindString,
		Value:     value,
		ExpireAt:  contract.ExpireAtFromTTL(ttl),
	}, nil
}

// Delete removes the cache entry identified by ownerType and cacheKey.
func (b *coordKVBackend) Delete(ctx context.Context, ownerType contract.OwnerType, cacheKey string) error {
	_, backendKey, err := b.resolveBackendKey(ownerType, cacheKey)
	if err != nil {
		return err
	}
	return b.kvStore.Delete(ctx, backendKey)
}

// Incr increments one integer cache value through coordination KV.
func (b *coordKVBackend) Incr(
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
	rawValue, ok, err := b.kvStore.Get(ctx, backendKey)
	if err != nil {
		return nil, err
	}
	if ok {
		payload, decodeErr := decodeCoordKVValue(rawValue)
		if decodeErr != nil {
			return nil, decodeErr
		}
		if payload.Kind != coordKVValueKindInt {
			return nil, bizerr.NewCode(contract.CodeKVCacheIncrementValueNotInteger)
		}
	}
	current, err := b.kvStore.IncrBy(ctx, backendKey, delta, ttl)
	if err != nil {
		return nil, err
	}
	remainingTTL, ttlErr := b.kvStore.TTL(ctx, backendKey)
	if ttlErr != nil {
		return nil, ttlErr
	}
	return &contract.Item{
		Key:       identity.CacheKey(),
		ValueKind: contract.ValueKindInt,
		IntValue:  current,
		ExpireAt:  contract.ExpireAtFromTTL(remainingTTL),
	}, nil
}

// Expire updates the expiration policy of a cache entry.
func (b *coordKVBackend) Expire(
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
	ok, err := b.kvStore.Expire(ctx, backendKey, ttl)
	if err != nil {
		return false, nil, err
	}
	return ok, contract.ExpireAtFromTTL(ttl), nil
}

// CleanupExpired is a no-op because coordination KV handles expiration.
func (b *coordKVBackend) CleanupExpired(_ context.Context) error {
	return nil
}

// resolveBackendKey validates a public cache key and maps it to a coordination key.
func (b *coordKVBackend) resolveBackendKey(
	ownerType contract.OwnerType,
	cacheKey string,
) (*contract.Identity, string, error) {
	if b == nil || b.kvStore == nil || b.keyBuilder == nil {
		return nil, "", bizerr.NewCode(contract.CodeKVCacheKeyInvalid)
	}
	identity, err := contract.ResolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, "", err
	}
	key, err := b.keyBuilder.KVKey(
		identity.TenantID(),
		ownerType.String(),
		identity.OwnerKey(),
		identity.Namespace(),
		identity.CacheKey(),
	)
	if err != nil {
		return nil, "", err
	}
	return identity, key, nil
}

// encodeCoordKVValue marshals one typed payload for coordination KV.
func encodeCoordKVValue(payload coordKVValue) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// decodeCoordKVValue unmarshals one typed payload from coordination KV.
func decodeCoordKVValue(rawValue string) (*coordKVValue, error) {
	var payload coordKVValue
	if err := json.Unmarshal([]byte(rawValue), &payload); err != nil {
		intValue, parseErr := strconv.ParseInt(rawValue, 10, 64)
		if parseErr != nil {
			return nil, err
		}
		return &coordKVValue{Kind: coordKVValueKindInt, IntValue: intValue}, nil
	}
	switch payload.Kind {
	case coordKVValueKindString, coordKVValueKindInt:
		return &payload, nil
	default:
		return nil, bizerr.NewCode(contract.CodeKVCacheKeyInvalid)
	}
}

// coordKVPayloadToItem maps a coordination KV payload into a public cache item.
func coordKVPayloadToItem(key string, payload *coordKVValue, ttl time.Duration) *contract.Item {
	item := &contract.Item{
		Key:      key,
		ExpireAt: contract.ExpireAtFromTTL(ttl),
	}
	if payload.Kind == coordKVValueKindInt {
		item.ValueKind = contract.ValueKindInt
		item.IntValue = payload.IntValue
		return item
	}
	item.ValueKind = contract.ValueKindString
	item.Value = payload.Value
	return item
}
