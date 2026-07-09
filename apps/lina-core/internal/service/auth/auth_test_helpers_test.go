// This file keeps tenant-auth token and pre-token storage helpers scoped to tests.

package auth

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model/entity"
)

// generateToken generates one test access JWT without creating a refresh token.
func (s *serviceImpl) generateToken(ctx context.Context, user *entity.SysUser, tenantID int, clientType ClientType) (string, string, error) {
	tokenID := guid.S()
	token, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindAccess, clientType, false, 0)
	if err != nil {
		return "", "", err
	}
	return token, tokenID, nil
}

// memoryPreTokenStore keeps pre-login tokens in memory for isolated tests.
type memoryPreTokenStore struct {
	mu      sync.Mutex
	records map[string]preTokenRecord
}

// newMemoryPreTokenStore creates an empty in-memory pre-login token store.
func newMemoryPreTokenStore() *memoryPreTokenStore {
	return &memoryPreTokenStore{records: make(map[string]preTokenRecord)}
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

// Consume returns and deletes one pre-login token.
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
