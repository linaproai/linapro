// This file verifies coordination-backed online-session hot state.

package session

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/coordination"
	"lina-core/pkg/bizerr"
)

// TestCoordinationStoreWritesHotStateAndProjection verifies login-time session
// writes are dual-written to coordination KV and the PostgreSQL projection.
func TestCoordinationStoreWritesHotStateAndProjection(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), 2*time.Hour)
	tokenID := uniqueSessionTestToken("coord-dual-write")

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := store.Set(ctx, &Session{
		TokenId:  tokenID,
		TenantId: 7,
		UserId:   11,
		Username: "coord-user",
		Ip:       "127.0.0.1",
		Browser:  "go-test",
		Os:       "darwin",
	}); err != nil {
		t.Fatalf("set coordination session: %v", err)
	}

	hotKey := coordinationSessionKey(t, coordSvc, 7, tokenID)
	raw, ok, err := coordSvc.KV().Get(ctx, hotKey)
	if err != nil || !ok {
		t.Fatalf("expected coordination hot state, ok=%t err=%v", ok, err)
	}
	payload, err := decodeSessionHotState(raw)
	if err != nil {
		t.Fatalf("decode hot state: %v", err)
	}
	if payload.TokenID != tokenID || payload.TenantID != 7 || payload.UserID != 11 {
		t.Fatalf("unexpected hot-state payload: %#v", payload)
	}
	ttl, err := coordSvc.KV().TTL(ctx, hotKey)
	if err != nil {
		t.Fatalf("read hot-state ttl: %v", err)
	}
	if ttl <= 90*time.Minute || ttl > 2*time.Hour {
		t.Fatalf("expected hot-state ttl near 2h, got %s", ttl)
	}

	projected, err := store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get projected session: %v", err)
	}
	if projected == nil || projected.TenantId != 7 || projected.UserId != 11 {
		t.Fatalf("expected projected session, got %#v", projected)
	}
}

// TestCoordinationStoreValidatesTenantAndRefreshesTTL verifies request-path
// validation uses tenant/token hot state and refreshes Redis TTL.
func TestCoordinationStoreValidatesTenantAndRefreshesTTL(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	tokenID := uniqueSessionTestToken("coord-touch")

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := store.Set(ctx, &Session{TokenId: tokenID, TenantId: 21, UserId: 31, Username: "touch-user"}); err != nil {
		t.Fatalf("set coordination session: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 22, tokenID, time.Hour); err != nil || ok {
		t.Fatalf("expected tenant mismatch miss, ok=%t err=%v", ok, err)
	}

	hotKey := coordinationSessionKey(t, coordSvc, 21, tokenID)
	if _, err := coordSvc.KV().Expire(ctx, hotKey, 40*time.Millisecond); err != nil {
		t.Fatalf("shorten hot-state ttl: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 21, tokenID, 2*time.Hour); err != nil || !ok {
		t.Fatalf("expected matching session active, ok=%t err=%v", ok, err)
	}
	ttl, err := coordSvc.KV().TTL(ctx, hotKey)
	if err != nil {
		t.Fatalf("read refreshed ttl: %v", err)
	}
	if ttl <= time.Hour {
		t.Fatalf("expected refreshed ttl > 1h, got %s", ttl)
	}
}

// TestCoordinationStoreThrottlesProjectionLastActive verifies Redis hot state is
// refreshed on every touch while PostgreSQL projection writes stay throttled.
func TestCoordinationStoreThrottlesProjectionLastActive(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	tokenID := uniqueSessionTestToken("coord-throttle")
	recent := gtime.Now().Truncate(time.Second)

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := store.Set(ctx, &Session{
		TokenId:        tokenID,
		TenantId:       41,
		UserId:         51,
		Username:       "throttle-user",
		LoginTime:      recent,
		LastActiveTime: recent,
	}); err != nil {
		t.Fatalf("set coordination session: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 41, tokenID, time.Hour); err != nil || !ok {
		t.Fatalf("touch recent session: ok=%t err=%v", ok, err)
	}
	projected, err := store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get projected session: %v", err)
	}
	if projected == nil || projected.LastActiveTime.String() != recent.String() {
		t.Fatalf("expected throttled projection time %v, got %#v", recent, projected)
	}

	stale := gtime.Now().Add(-2 * sessionLastActiveUpdateWindow).Truncate(time.Second)
	if err = insertOrUpdateSessionProjection(ctx, &Session{
		TokenId:        tokenID,
		TenantId:       41,
		UserId:         51,
		Username:       "throttle-user",
		LoginTime:      stale,
		LastActiveTime: stale,
	}); err != nil {
		t.Fatalf("force stale projection: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 41, tokenID, time.Hour); err != nil || !ok {
		t.Fatalf("touch stale session: ok=%t err=%v", ok, err)
	}
	projected, err = store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get refreshed projection: %v", err)
	}
	if projected == nil || !projected.LastActiveTime.After(stale) {
		t.Fatalf("expected projection last-active to refresh after throttle window, got %#v", projected)
	}
}

// TestCoordinationStoreDeleteRemovesHotStateAndProjection verifies force-logout
// style deletion clears both Redis hot state and PostgreSQL projection.
func TestCoordinationStoreDeleteRemovesHotStateAndProjection(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	tokenID := uniqueSessionTestToken("coord-delete")

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := store.Set(ctx, &Session{TokenId: tokenID, TenantId: 61, UserId: 71, Username: "delete-user"}); err != nil {
		t.Fatalf("set coordination session: %v", err)
	}
	if err := store.Delete(ctx, tokenID); err != nil {
		t.Fatalf("delete coordination session: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 61, tokenID, time.Hour); err != nil || ok {
		t.Fatalf("expected deleted hot state miss, ok=%t err=%v", ok, err)
	}
	projected, err := store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get deleted projection: %v", err)
	}
	if projected != nil {
		t.Fatalf("expected projection deleted, got %#v", projected)
	}
}

// TestCoordinationStoreDeleteByUserIdUsesIndex verifies tenant/user deletion
// removes only indexed sessions for that tenant and user.
func TestCoordinationStoreDeleteByUserIdUsesIndex(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	firstToken := uniqueSessionTestToken("coord-user-delete-a")
	secondToken := uniqueSessionTestToken("coord-user-delete-b")
	otherTenantToken := uniqueSessionTestToken("coord-user-delete-c")

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, firstToken)
		cleanupSessionTestToken(t, ctx, secondToken)
		cleanupSessionTestToken(t, ctx, otherTenantToken)
	})
	for _, item := range []*Session{
		{TokenId: firstToken, TenantId: 81, UserId: 91, Username: "delete-user"},
		{TokenId: secondToken, TenantId: 81, UserId: 91, Username: "delete-user"},
		{TokenId: otherTenantToken, TenantId: 82, UserId: 91, Username: "delete-user"},
	} {
		if err := store.Set(ctx, item); err != nil {
			t.Fatalf("set session %s: %v", item.TokenId, err)
		}
	}
	if err := store.DeleteByUserId(ctx, 81, 91); err != nil {
		t.Fatalf("delete sessions by user: %v", err)
	}
	if ok, err := store.TouchOrValidate(ctx, 81, firstToken, time.Hour); err != nil || ok {
		t.Fatalf("expected first token deleted, ok=%t err=%v", ok, err)
	}
	if ok, err := store.TouchOrValidate(ctx, 81, secondToken, time.Hour); err != nil || ok {
		t.Fatalf("expected second token deleted, ok=%t err=%v", ok, err)
	}
	if ok, err := store.TouchOrValidate(ctx, 82, otherTenantToken, time.Hour); err != nil || !ok {
		t.Fatalf("expected other tenant token preserved, ok=%t err=%v", ok, err)
	}
}

// TestCoordinationStoreRedisReadFailureFailClosed verifies request validation
// returns a structured error when shared hot state cannot be read.
func TestCoordinationStoreRedisReadFailureFailClosed(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	writer := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	tokenID := uniqueSessionTestToken("coord-fail-closed")

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := writer.Set(ctx, &Session{TokenId: tokenID, TenantId: 101, UserId: 111, Username: "failure-user"}); err != nil {
		t.Fatalf("set coordination session: %v", err)
	}
	store := &CoordinationStore{
		kvStore:    failingSessionKVStore{},
		keyBuilder: coordSvc.KeyBuilder(),
		projection: NewDBStore(),
		defaultTTL: time.Hour,
	}
	if ok, err := store.TouchOrValidate(ctx, 101, tokenID, time.Hour); ok || !bizerr.Is(err, CodeSessionStateUnavailable) {
		t.Fatalf("expected fail-closed session state error, ok=%t err=%v", ok, err)
	}
}

// TestCoordinationStoreCleanupInactiveKeepsProjectionCleanup verifies the
// PostgreSQL projection cleanup path is retained with coordination hot state.
func TestCoordinationStoreCleanupInactiveKeepsProjectionCleanup(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	store := NewCoordinationStoreWithDefaultTTL(coordSvc, NewDBStore(), time.Hour)
	tokenID := uniqueSessionTestToken("coord-cleanup")
	stale := gtime.Now().Add(-2 * time.Hour)

	t.Cleanup(func() {
		cleanupSessionTestToken(t, ctx, tokenID)
	})
	if err := store.Set(ctx, &Session{
		TokenId:        tokenID,
		TenantId:       121,
		UserId:         131,
		Username:       "cleanup-user",
		LoginTime:      stale,
		LastActiveTime: stale,
	}); err != nil {
		t.Fatalf("set stale coordination session: %v", err)
	}
	cleaned, err := store.CleanupInactive(ctx, time.Hour)
	if err != nil {
		t.Fatalf("cleanup inactive sessions: %v", err)
	}
	if cleaned <= 0 {
		t.Fatalf("expected cleanup to remove projection row, cleaned=%d", cleaned)
	}
	projected, err := store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get cleanup projection: %v", err)
	}
	if projected != nil {
		t.Fatalf("expected projection cleanup, got %#v", projected)
	}
}

// uniqueSessionTestToken creates a unique token ID for session tests.
func uniqueSessionTestToken(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// coordinationSessionKey builds the backend key for a tenant/token hot state.
func coordinationSessionKey(t *testing.T, coordSvc coordination.Service, tenantID int, tokenID string) string {
	t.Helper()

	key, err := coordSvc.KeyBuilder().RawKVKey(sessionHotStateComponent, fmt.Sprintf("%d", tenantID), tokenID)
	if err != nil {
		t.Fatalf("build coordination session key: %v", err)
	}
	return key
}

// cleanupSessionTestToken removes one online-session projection row.
func cleanupSessionTestToken(t *testing.T, ctx context.Context, tokenID string) {
	t.Helper()

	if tokenID == "" {
		return
	}
	if err := (&DBStore{}).Delete(ctx, tokenID); err != nil {
		t.Fatalf("cleanup session test token %s: %v", tokenID, err)
	}
}

// insertOrUpdateSessionProjection writes a DB-only projection row for throttle
// tests without touching coordination hot state.
func insertOrUpdateSessionProjection(ctx context.Context, item *Session) error {
	return (&DBStore{}).Set(ctx, item)
}

// failingSessionKVStore simulates a Redis/coordination read outage for
// request-path fail-closed tests.
type failingSessionKVStore struct{}

// Get returns a deterministic coordination failure.
func (f failingSessionKVStore) Get(context.Context, string) (string, bool, error) {
	return "", false, errors.New("coordination kv unavailable")
}

// Set returns a deterministic coordination failure.
func (f failingSessionKVStore) Set(context.Context, string, string, time.Duration) error {
	return errors.New("coordination kv unavailable")
}

// SetNX returns a deterministic coordination failure.
func (f failingSessionKVStore) SetNX(context.Context, string, string, time.Duration) (bool, error) {
	return false, errors.New("coordination kv unavailable")
}

// Delete returns a deterministic coordination failure.
func (f failingSessionKVStore) Delete(context.Context, string) error {
	return errors.New("coordination kv unavailable")
}

// CompareAndDelete returns a deterministic coordination failure.
func (f failingSessionKVStore) CompareAndDelete(context.Context, string, string) (bool, error) {
	return false, errors.New("coordination kv unavailable")
}

// IncrBy returns a deterministic coordination failure.
func (f failingSessionKVStore) IncrBy(context.Context, string, int64, time.Duration) (int64, error) {
	return 0, errors.New("coordination kv unavailable")
}

// Expire returns a deterministic coordination failure.
func (f failingSessionKVStore) Expire(context.Context, string, time.Duration) (bool, error) {
	return false, errors.New("coordination kv unavailable")
}

// TTL returns a deterministic coordination failure.
func (f failingSessionKVStore) TTL(context.Context, string) (time.Duration, error) {
	return 0, errors.New("coordination kv unavailable")
}
