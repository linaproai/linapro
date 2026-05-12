// This file implements JWT revocation tracking.

package auth

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/kvcache"
)

// revokeStore defines the JWT revocation storage contract.
type revokeStore interface {
	// Add marks a token as revoked until expiresAt.
	Add(ctx context.Context, tokenID string, expiresAt time.Time) error
	// Revoked reports whether a token is currently revoked.
	Revoked(ctx context.Context, tokenID string) (bool, error)
}

// layeredRevokeStore keeps a process-local revocation cache in front of the
// shared cache-backed store so the current node rejects revoked tokens
// immediately while other nodes converge through the shared cache state.
type layeredRevokeStore struct {
	local  *memoryRevokeStore
	shared revokeStore
}

// kvRevokeStore keeps revoked token IDs in the host KV cache until their JWT
// expiry so all cluster nodes reject the same revoked token IDs.
type kvRevokeStore struct {
	cache kvcache.Service
}

// memoryRevokeStore keeps revoked token IDs in memory for isolated tests.
type memoryRevokeStore struct {
	mu      sync.Mutex
	records map[string]time.Time
}

// newRevokeList creates the default shared token revocation store.
func newRevokeList() revokeStore {
	return newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(kvcache.Instance()))
}

// newLayeredRevokeStore composes process-local and shared revocation stores.
func newLayeredRevokeStore(local *memoryRevokeStore, shared revokeStore) revokeStore {
	return &layeredRevokeStore{local: local, shared: shared}
}

// newKVRevokeStore creates a kvcache-backed token revocation store.
func newKVRevokeStore(cache kvcache.Service) revokeStore {
	return &kvRevokeStore{cache: cache}
}

// newMemoryRevokeStore creates an empty in-memory token revocation store.
func newMemoryRevokeStore() *memoryRevokeStore {
	return &memoryRevokeStore{records: make(map[string]time.Time)}
}

// Add marks a token as revoked in local memory and the shared store.
func (s *layeredRevokeStore) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	if s == nil || tokenID == "" {
		return nil
	}
	if s.local != nil {
		if err := s.local.Add(ctx, tokenID, expiresAt); err != nil {
			return err
		}
	}
	if s.shared == nil {
		return nil
	}
	return s.shared.Add(ctx, tokenID, expiresAt)
}

// Revoked reports whether a token is revoked locally or in the shared store.
func (s *layeredRevokeStore) Revoked(ctx context.Context, tokenID string) (bool, error) {
	if s == nil || tokenID == "" {
		return false, nil
	}
	if s.local != nil {
		revoked, err := s.local.Revoked(ctx, tokenID)
		if err != nil || revoked {
			return revoked, err
		}
	}
	if s.shared == nil {
		return false, nil
	}
	revoked, err := s.shared.Revoked(ctx, tokenID)
	if err != nil || !revoked || s.local == nil {
		return revoked, err
	}
	if expiresAt, ok := sharedRevokeExpiresAt(ctx, s.shared, tokenID); ok {
		if err = s.local.Add(ctx, tokenID, expiresAt); err != nil {
			return false, err
		}
	}
	return true, nil
}

// Add marks a token as revoked until expiresAt.
func (s *kvRevokeStore) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	if tokenID == "" {
		return nil
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	_, err := s.cache.Set(ctx, kvcache.OwnerTypeModule, revokeCacheKey(tokenID), "1", ttl)
	return err
}

// Revoked reports whether a token is currently revoked.
func (s *kvRevokeStore) Revoked(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}
	_, ok, err := s.cache.Get(ctx, kvcache.OwnerTypeModule, revokeCacheKey(tokenID))
	return ok, err
}

// sharedRevokeExpiresAt returns the shared revocation expiration when the
// shared store exposes one through the cache item metadata.
func sharedRevokeExpiresAt(ctx context.Context, shared revokeStore, tokenID string) (time.Time, bool) {
	kvStore, ok := shared.(*kvRevokeStore)
	if !ok || kvStore == nil || kvStore.cache == nil || tokenID == "" {
		return time.Time{}, false
	}
	item, ok, err := kvStore.cache.Get(ctx, kvcache.OwnerTypeModule, revokeCacheKey(tokenID))
	if err != nil || !ok || item == nil || item.ExpireAt == nil {
		return time.Time{}, false
	}
	return item.ExpireAt.Time, true
}

// Add marks a token as revoked until expiresAt.
func (s *memoryRevokeStore) Add(_ context.Context, tokenID string, expiresAt time.Time) error {
	if tokenID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[tokenID] = expiresAt
	return nil
}

// Revoked reports whether a token is currently revoked.
func (s *memoryRevokeStore) Revoked(_ context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt, ok := s.records[tokenID]
	if !ok {
		return false, nil
	}
	if time.Now().After(expiresAt) {
		delete(s.records, tokenID)
		return false, nil
	}
	return true, nil
}

// revokeCacheKey builds the scoped kvcache key for one revoked JWT ID.
func revokeCacheKey(tokenID string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, revokeStoreNamespace, tokenID)
}
