// This file implements the Redis/coordination-backed online-session hot state
// used by clustered deployments while PostgreSQL remains the management
// projection.

package session

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/datascope"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

var _ SessionConfigurableStore = (*CoordinationStore)(nil)

// CoordinationStore stores request-path session state in coordination KV and
// delegates management-list projections to PostgreSQL.
type CoordinationStore struct {
	coordinationSvc coordination.Service
	kvStore         coordination.KVStore
	keyBuilder      *coordination.KeyBuilder
	projection      Store
	defaultTTL      time.Duration
}

// sessionHotState is the JSON payload persisted in coordination KV for one
// active online session.
type sessionHotState struct {
	Schema         int       `json:"schema"`
	TokenID        string    `json:"tokenId"`
	TenantID       int       `json:"tenantId"`
	UserID         int       `json:"userId"`
	Username       string    `json:"username"`
	DeptName       string    `json:"deptName,omitempty"`
	IP             string    `json:"ip,omitempty"`
	Browser        string    `json:"browser,omitempty"`
	OS             string    `json:"os,omitempty"`
	LoginTime      time.Time `json:"loginTime"`
	LastActiveTime time.Time `json:"lastActiveTime"`
	StoredAt       time.Time `json:"storedAt"`
}

// sessionUserIndex records the active token IDs for a tenant/user pair so
// tenant-scoped force logout can delete Redis hot-state keys without scanning.
type sessionUserIndex struct {
	Schema  int       `json:"schema"`
	Tokens  []string  `json:"tokens"`
	Updated time.Time `json:"updated"`
}

// NewCoordinationStore creates a session store backed by coordination KV for
// hot state and PostgreSQL for management projections.
func NewCoordinationStore(coordinationSvc coordination.Service, projection Store) Store {
	if projection == nil {
		projection = &DBStore{}
	}
	if coordinationSvc == nil {
		return projection
	}
	if coordinationSvc.KV() == nil || coordinationSvc.KeyBuilder() == nil {
		return projection
	}
	return &CoordinationStore{
		coordinationSvc: coordinationSvc,
		kvStore:         coordinationSvc.KV(),
		keyBuilder:      coordinationSvc.KeyBuilder(),
		projection:      projection,
		defaultTTL:      24 * time.Hour,
	}
}

// NewCoordinationStoreWithDefaultTTL creates a coordination-backed store and
// applies the runtime-effective session timeout before callers write sessions.
func NewCoordinationStoreWithDefaultTTL(
	coordinationSvc coordination.Service,
	projection Store,
	ttl time.Duration,
) Store {
	store := NewCoordinationStore(coordinationSvc, projection)
	if configurable, ok := store.(SessionConfigurableStore); ok {
		configurable.SetDefaultTTL(ttl)
	}
	return store
}

// SetDefaultTTL updates the hot-state TTL used for login-time writes when the
// caller can provide the runtime-effective session timeout.
func (s *CoordinationStore) SetDefaultTTL(ttl time.Duration) {
	if s == nil || ttl <= 0 {
		return
	}
	s.defaultTTL = ttl
}

// Set writes one online session to coordination KV and PostgreSQL projection.
func (s *CoordinationStore) Set(ctx context.Context, session *Session) error {
	if s == nil || session == nil {
		return nil
	}
	if err := s.ensureReady(); err != nil {
		return err
	}
	now := time.Now()
	if session.LoginTime == nil {
		session.LoginTime = gtime.NewFromTime(now)
	}
	if session.LastActiveTime == nil {
		session.LastActiveTime = gtime.NewFromTime(now)
	}
	payload := sessionHotStateFromSession(session, now)
	encoded, err := encodeSessionHotState(payload)
	if err != nil {
		return err
	}
	ttl := s.effectiveTTL(0)
	backendKey, err := s.sessionKey(session.TenantId, session.TokenId)
	if err != nil {
		return err
	}
	if err = s.kvStore.Set(ctx, backendKey, encoded, ttl); err != nil {
		return bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if err = s.addUserIndexToken(ctx, session.TenantId, session.UserId, session.TokenId, ttl); err != nil {
		if cleanupErr := s.kvStore.Delete(ctx, backendKey); cleanupErr != nil {
			logger.Warningf(ctx, "cleanup session hot key after index write failure tokenId=%s tenantId=%d err=%v", session.TokenId, session.TenantId, cleanupErr)
		}
		return err
	}
	if err = s.projection.Set(ctx, session); err != nil {
		if cleanupErr := s.deleteHotState(ctx, session.TenantId, session.UserId, session.TokenId); cleanupErr != nil {
			logger.Warningf(ctx, "cleanup session hot state after projection write failure tokenId=%s tenantId=%d err=%v", session.TokenId, session.TenantId, cleanupErr)
		}
		return err
	}
	return nil
}

// Get returns one session projection by token ID.
func (s *CoordinationStore) Get(ctx context.Context, tokenId string) (*Session, error) {
	if s == nil || s.projection == nil {
		return nil, nil
	}
	return s.projection.Get(ctx, tokenId)
}

// Delete removes one session from coordination KV and PostgreSQL projection.
func (s *CoordinationStore) Delete(ctx context.Context, tokenId string) error {
	if s == nil {
		return nil
	}
	if err := s.ensureReady(); err != nil {
		return err
	}
	projection, err := s.projection.Get(ctx, tokenId)
	if err != nil {
		return err
	}
	if projection != nil {
		if err = s.deleteHotState(ctx, projection.TenantId, projection.UserId, tokenId); err != nil {
			return err
		}
	} else if err = s.deleteUnknownTenantHotState(ctx, tokenId); err != nil {
		return err
	}
	return s.projection.Delete(ctx, tokenId)
}

// DeleteByUserId removes all sessions belonging to a user in one tenant.
func (s *CoordinationStore) DeleteByUserId(ctx context.Context, tenantId int, userId int) error {
	if s == nil {
		return nil
	}
	if err := s.ensureReady(); err != nil {
		return err
	}
	tokens, err := s.userIndexTokens(ctx, tenantId, userId)
	if err != nil {
		return err
	}
	for _, tokenID := range tokens {
		if err = s.deleteHotState(ctx, tenantId, userId, tokenID); err != nil {
			return err
		}
	}
	return s.projection.DeleteByUserId(ctx, tenantId, userId)
}

// List returns all sessions matching the filter from PostgreSQL projection.
func (s *CoordinationStore) List(ctx context.Context, filter *ListFilter) ([]*Session, error) {
	return s.projection.List(ctx, filter)
}

// ListPage returns one paginated session list from PostgreSQL projection.
func (s *CoordinationStore) ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error) {
	return s.projection.ListPage(ctx, filter, pageNum, pageSize)
}

// ListPageScoped returns one paginated session list constrained by data scope.
func (s *CoordinationStore) ListPageScoped(
	ctx context.Context,
	filter *ListFilter,
	pageNum, pageSize int,
	scopeSvc datascope.Service,
	tenantSvc tenantcapsvc.Service,
) (*ListResult, error) {
	return s.projection.ListPageScoped(ctx, filter, pageNum, pageSize, scopeSvc, tenantSvc)
}

// Count returns the total number of projected online sessions.
func (s *CoordinationStore) Count(ctx context.Context) (int, error) {
	return s.projection.Count(ctx)
}

// TouchOrValidate validates coordination hot state and refreshes its TTL.
func (s *CoordinationStore) TouchOrValidate(ctx context.Context, tenantId int, tokenId string, timeout time.Duration) (bool, error) {
	if s == nil {
		return false, nil
	}
	if err := s.ensureReady(); err != nil {
		return false, err
	}
	backendKey, err := s.sessionKey(tenantId, tokenId)
	if err != nil {
		return false, err
	}
	raw, ok, err := s.kvStore.Get(ctx, backendKey)
	if err != nil {
		return false, bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if !ok {
		return false, nil
	}
	payload, err := decodeSessionHotState(raw)
	if err != nil {
		return false, bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if payload.TenantID != tenantId || payload.TokenID != tokenId {
		return false, nil
	}
	now := time.Now()
	if timeout > 0 && !payload.LastActiveTime.IsZero() && !payload.LastActiveTime.After(now.Add(-timeout)) {
		if err = s.deleteHotState(ctx, payload.TenantID, payload.UserID, payload.TokenID); err != nil {
			return false, err
		}
		return false, nil
	}

	payload.LastActiveTime = now
	payload.StoredAt = now
	encoded, err := encodeSessionHotState(payload)
	if err != nil {
		return false, err
	}
	ttl := timeout
	if ttl <= 0 {
		ttl = s.effectiveTTL(timeout)
	}
	if err = s.kvStore.Set(ctx, backendKey, encoded, ttl); err != nil {
		return false, bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if err = s.updateProjectionLastActiveIfDue(ctx, tenantId, tokenId, gtime.NewFromTime(now)); err != nil {
		return false, err
	}
	return true, nil
}

// CleanupInactive removes inactive projection rows. Coordination KV expires hot
// state natively, so the cleanup remains focused on PostgreSQL projection rows.
func (s *CoordinationStore) CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	return s.projection.CleanupInactive(ctx, timeout)
}

// ensureReady verifies that the coordination store dependencies are available.
func (s *CoordinationStore) ensureReady() error {
	if s == nil || s.kvStore == nil || s.keyBuilder == nil {
		return bizerr.NewCode(CodeSessionStateUnavailable)
	}
	return nil
}

// effectiveTTL returns the runtime-provided session timeout or the store's
// default fallback.
func (s *CoordinationStore) effectiveTTL(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	if s != nil && s.defaultTTL > 0 {
		return s.defaultTTL
	}
	return 24 * time.Hour
}

// sessionHotStateFromSession converts one public session projection into the
// coordination hot-state payload.
func sessionHotStateFromSession(session *Session, now time.Time) sessionHotState {
	return sessionHotState{
		Schema:         sessionHotStateSchema,
		TokenID:        session.TokenId,
		TenantID:       session.TenantId,
		UserID:         session.UserId,
		Username:       session.Username,
		DeptName:       session.DeptName,
		IP:             session.Ip,
		Browser:        session.Browser,
		OS:             session.Os,
		LoginTime:      session.LoginTime.Time,
		LastActiveTime: session.LastActiveTime.Time,
		StoredAt:       now,
	}
}

// encodeSessionHotState serializes one hot-state payload.
func encodeSessionHotState(payload sessionHotState) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// decodeSessionHotState deserializes one hot-state payload.
func decodeSessionHotState(value string) (sessionHotState, error) {
	var payload sessionHotState
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return sessionHotState{}, err
	}
	if payload.Schema != sessionHotStateSchema {
		return sessionHotState{}, bizerr.NewCode(CodeSessionStateUnavailable)
	}
	return payload, nil
}

// sessionKey builds the tenant/token Redis hot-state key.
func (s *CoordinationStore) sessionKey(tenantID int, tokenID string) (string, error) {
	return s.keyBuilder.RawKVKey(
		sessionHotStateComponent,
		strconv.Itoa(tenantID),
		tokenID,
	)
}

// userIndexKey builds the tenant/user Redis index key.
func (s *CoordinationStore) userIndexKey(tenantID int, userID int) (string, error) {
	return s.keyBuilder.RawKVKey(
		sessionHotStateComponent,
		"user-index",
		strconv.Itoa(tenantID),
		strconv.Itoa(userID),
	)
}

// addUserIndexToken appends one token ID to the tenant/user index.
func (s *CoordinationStore) addUserIndexToken(ctx context.Context, tenantID int, userID int, tokenID string, ttl time.Duration) error {
	tokens, err := s.userIndexTokens(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	seen := false
	for _, existing := range tokens {
		if existing == tokenID {
			seen = true
			break
		}
	}
	if !seen {
		tokens = append(tokens, tokenID)
	}
	payload := sessionUserIndex{Schema: sessionUserIndexSchema, Tokens: tokens, Updated: time.Now()}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	indexKey, err := s.userIndexKey(tenantID, userID)
	if err != nil {
		return err
	}
	if err = s.kvStore.Set(ctx, indexKey, string(encoded), ttl); err != nil {
		return bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	return nil
}

// userIndexTokens reads token IDs from the tenant/user index.
func (s *CoordinationStore) userIndexTokens(ctx context.Context, tenantID int, userID int) ([]string, error) {
	indexKey, err := s.userIndexKey(tenantID, userID)
	if err != nil {
		return nil, err
	}
	raw, ok, err := s.kvStore.Get(ctx, indexKey)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if !ok {
		return nil, nil
	}
	var payload sessionUserIndex
	if err = json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if payload.Schema != sessionUserIndexSchema {
		return nil, bizerr.NewCode(CodeSessionStateUnavailable)
	}
	return append([]string(nil), payload.Tokens...), nil
}

// deleteHotState removes one tenant/token hot-state key and refreshes the
// tenant/user index without scanning Redis.
func (s *CoordinationStore) deleteHotState(ctx context.Context, tenantID int, userID int, tokenID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	backendKey, err := s.sessionKey(tenantID, tokenID)
	if err != nil {
		return err
	}
	if err = s.kvStore.Delete(ctx, backendKey); err != nil {
		return bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	if userID <= 0 {
		return nil
	}
	tokens, err := s.userIndexTokens(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	remaining := make([]string, 0, len(tokens))
	for _, existing := range tokens {
		if existing != tokenID {
			remaining = append(remaining, existing)
		}
	}
	indexKey, err := s.userIndexKey(tenantID, userID)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		if err = s.kvStore.Delete(ctx, indexKey); err != nil {
			return bizerr.WrapCode(err, CodeSessionStateUnavailable)
		}
		return nil
	}
	payload := sessionUserIndex{Schema: sessionUserIndexSchema, Tokens: remaining, Updated: time.Now()}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err = s.kvStore.Set(ctx, indexKey, string(encoded), s.effectiveTTL(0)); err != nil {
		return bizerr.WrapCode(err, CodeSessionStateUnavailable)
	}
	return nil
}

// deleteUnknownTenantHotState preserves token-ID-only deletion behavior when
// the PostgreSQL projection has already disappeared.
func (s *CoordinationStore) deleteUnknownTenantHotState(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return nil
	}
	return nil
}

// updateProjectionLastActiveIfDue refreshes PostgreSQL projection only outside
// the configured write-throttle window.
func (s *CoordinationStore) updateProjectionLastActiveIfDue(
	ctx context.Context,
	tenantID int,
	tokenID string,
	now *gtime.Time,
) error {
	cols := dao.SysOnlineSession.Columns()
	_, err := tenantSessionModel(ctx, tenantID, tokenID).
		WhereLT(cols.LastActiveTime, now.Add(-sessionLastActiveUpdateWindow)).
		Data(do.SysOnlineSession{LastActiveTime: now}).
		Update()
	return err
}
