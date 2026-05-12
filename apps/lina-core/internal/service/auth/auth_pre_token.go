// This file implements short-lived single-use pre-login token storage.

package auth

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/kvcache"
)

const preTokenTTL time.Duration = time.Minute

const (
	authTokenStoreOwner      = "auth"
	preTokenStoreNamespace   = "pre-token"
	preTokenConsumeNamespace = "pre-token-consume"
	revokeStoreNamespace     = "jwt-revoke"
	preTokenValueSchema      = 1
	preTokenGeneratedPrefix  = "pre_"
)

// preTokenStore defines the storage contract for pre-login token handoff.
type preTokenStore interface {
	// Create stores one short-lived pre-login token.
	Create(ctx context.Context, record preTokenRecord) (string, error)
	// Consume returns and deletes one pre-login token.
	Consume(ctx context.Context, token string) (preTokenRecord, bool, error)
}

// preTokenRecord stores the authenticated user identity between password
// verification and tenant selection.
type preTokenRecord struct {
	UserID    int
	Username  string
	Status    int
	ExpiresAt time.Time
}

// storedPreTokenRecord is the persisted JSON envelope for pre-login tokens.
type storedPreTokenRecord struct {
	Schema   int            `json:"schema"`
	Record   preTokenRecord `json:"record"`
	StoredAt time.Time      `json:"storedAt"`
}

// kvPreTokenStore stores pre-login tokens in the host KV cache so all cluster
// nodes see the same single-use handoff state.
type kvPreTokenStore struct {
	cache kvcache.Service
}

// memoryPreTokenStore keeps pre-login tokens in memory for isolated tests.
type memoryPreTokenStore struct {
	mu      sync.Mutex
	records map[string]preTokenRecord
}

// newPreTokenStore creates the default shared pre-login token store.
func newPreTokenStore() preTokenStore {
	return newKVPreTokenStore(kvcache.Instance())
}

// newKVPreTokenStore creates a kvcache-backed pre-login token store.
func newKVPreTokenStore(cache kvcache.Service) preTokenStore {
	return &kvPreTokenStore{cache: cache}
}

// newMemoryPreTokenStore creates an empty in-memory pre-login token store.
func newMemoryPreTokenStore() *memoryPreTokenStore {
	return &memoryPreTokenStore{records: make(map[string]preTokenRecord)}
}

// Create stores one short-lived pre-login token.
func (s *kvPreTokenStore) Create(ctx context.Context, record preTokenRecord) (string, error) {
	token := preTokenGeneratedPrefix + guid.S()
	record.ExpiresAt = time.Now().Add(preTokenTTL)
	payload, err := json.Marshal(storedPreTokenRecord{
		Schema:   preTokenValueSchema,
		Record:   record,
		StoredAt: time.Now(),
	})
	if err != nil {
		return "", err
	}
	_, err = s.cache.Set(ctx, kvcache.OwnerTypeModule, preTokenCacheKey(token), string(payload), preTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Consume returns and deletes one pre-login token. Expired tokens are treated
// the same as missing tokens so callers get one stable user-facing error.
func (s *kvPreTokenStore) Consume(ctx context.Context, token string) (preTokenRecord, bool, error) {
	consumeItem, err := s.cache.Incr(ctx, kvcache.OwnerTypeModule, preTokenConsumeCacheKey(token), 1, preTokenTTL)
	if err != nil {
		return preTokenRecord{}, false, err
	}
	if consumeItem.IntValue != 1 {
		return preTokenRecord{}, false, nil
	}

	item, ok, err := s.cache.Get(ctx, kvcache.OwnerTypeModule, preTokenCacheKey(token))
	if err != nil {
		return preTokenRecord{}, false, err
	}
	if !ok {
		return preTokenRecord{}, false, nil
	}
	if err = s.cache.Delete(ctx, kvcache.OwnerTypeModule, preTokenCacheKey(token)); err != nil {
		return preTokenRecord{}, false, err
	}

	var payload storedPreTokenRecord
	if err = json.Unmarshal([]byte(item.Value), &payload); err != nil {
		return preTokenRecord{}, false, err
	}
	if payload.Schema != preTokenValueSchema || time.Now().After(payload.Record.ExpiresAt) {
		return preTokenRecord{}, false, nil
	}
	return payload.Record, true, nil
}

// Create stores one short-lived pre-login token.
func (s *memoryPreTokenStore) Create(_ context.Context, record preTokenRecord) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := preTokenGeneratedPrefix + guid.S()
	record.ExpiresAt = time.Now().Add(preTokenTTL)
	s.records[token] = record
	return token, nil
}

// Consume returns and deletes one pre-login token. Expired tokens are treated
// the same as missing tokens so callers get one stable user-facing error.
func (s *memoryPreTokenStore) Consume(_ context.Context, token string) (preTokenRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[token]
	if !ok {
		return preTokenRecord{}, false, nil
	}
	delete(s.records, token)
	if time.Now().After(record.ExpiresAt) {
		return preTokenRecord{}, false, nil
	}
	return record, true, nil
}

// preTokenCacheKey builds the scoped kvcache key for one pre-login token.
func preTokenCacheKey(token string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, preTokenStoreNamespace, token)
}

// preTokenConsumeCacheKey builds the scoped kvcache key for the single-use
// consume marker of one pre-login token.
func preTokenConsumeCacheKey(token string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, preTokenConsumeNamespace, token)
}
