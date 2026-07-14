// This file implements short-lived single-use password-reset token storage.

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"lina-core/internal/service/kvcache"
)

const passwordResetTokenTTL time.Duration = 30 * time.Minute

const (
	passwordResetStoreNamespace   = "password-reset"
	passwordResetConsumeNamespace = "password-reset-consume"
	passwordResetValueSchema      = 1
	passwordResetTokenPrefix      = "rst_"
	passwordResetTokenBytes       = 24
)

// passwordResetStore defines the storage contract for password-reset tokens.
type passwordResetStore interface {
	// Create stores one short-lived password-reset token for a user.
	Create(ctx context.Context, record passwordResetRecord) (string, error)
	// Consume returns and deletes one password-reset token.
	Consume(ctx context.Context, token string) (passwordResetRecord, bool, error)
}

// passwordResetRecord stores the account targeted by one reset token.
type passwordResetRecord struct {
	UserID    int
	Email     string
	ExpiresAt time.Time
}

// storedPasswordResetRecord is the persisted JSON envelope for reset tokens.
type storedPasswordResetRecord struct {
	Schema   int                 `json:"schema"`
	Record   passwordResetRecord `json:"record"`
	StoredAt time.Time           `json:"storedAt"`
}

// kvPasswordResetStore stores password-reset tokens in the host KV cache.
type kvPasswordResetStore struct {
	cache kvcache.Service
}

// newKVPasswordResetStore creates a kvcache-backed password-reset token store.
func newKVPasswordResetStore(cache kvcache.Service) passwordResetStore {
	return &kvPasswordResetStore{cache: cache}
}

// Create stores one short-lived password-reset token.
func (s *kvPasswordResetStore) Create(ctx context.Context, record passwordResetRecord) (string, error) {
	token, err := generatePasswordResetToken()
	if err != nil {
		return "", err
	}
	record.ExpiresAt = time.Now().Add(passwordResetTokenTTL)
	payload, err := json.Marshal(storedPasswordResetRecord{
		Schema:   passwordResetValueSchema,
		Record:   record,
		StoredAt: time.Now(),
	})
	if err != nil {
		return "", err
	}
	_, err = s.cache.Set(ctx, kvcache.OwnerTypeModule, passwordResetCacheKey(token), string(payload), passwordResetTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Consume returns and deletes one password-reset token. Expired or already
// consumed tokens are treated as missing so callers get one stable error.
func (s *kvPasswordResetStore) Consume(ctx context.Context, token string) (passwordResetRecord, bool, error) {
	consumeItem, err := s.cache.Incr(ctx, kvcache.OwnerTypeModule, passwordResetConsumeCacheKey(token), 1, passwordResetTokenTTL)
	if err != nil {
		return passwordResetRecord{}, false, err
	}
	if consumeItem.IntValue != 1 {
		return passwordResetRecord{}, false, nil
	}

	item, ok, err := s.cache.Get(ctx, kvcache.OwnerTypeModule, passwordResetCacheKey(token))
	if err != nil {
		return passwordResetRecord{}, false, err
	}
	if !ok {
		return passwordResetRecord{}, false, nil
	}
	if err = s.cache.Delete(ctx, kvcache.OwnerTypeModule, passwordResetCacheKey(token)); err != nil {
		return passwordResetRecord{}, false, err
	}

	var payload storedPasswordResetRecord
	if err = json.Unmarshal([]byte(item.Value), &payload); err != nil {
		return passwordResetRecord{}, false, err
	}
	if payload.Schema != passwordResetValueSchema || time.Now().After(payload.Record.ExpiresAt) {
		return passwordResetRecord{}, false, nil
	}
	return payload.Record, true, nil
}

// generatePasswordResetToken returns a high-entropy opaque reset token.
func generatePasswordResetToken() (string, error) {
	buffer := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return passwordResetTokenPrefix + hex.EncodeToString(buffer), nil
}

// passwordResetCacheKey builds the scoped kvcache key for one reset token.
func passwordResetCacheKey(token string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, passwordResetStoreNamespace, token)
}

// passwordResetConsumeCacheKey builds the single-use consume marker key.
func passwordResetConsumeCacheKey(token string) string {
	return kvcache.BuildCacheKey(authTokenStoreOwner, passwordResetConsumeNamespace, token)
}
