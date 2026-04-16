// This file implements token-scoped access-context caching and permission
// topology revision synchronization for declarative interface authorization.

package role

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/service/kvcache"
	"lina-core/pkg/logger"
)

const (
	accessCacheKeyPrefix    = "role:user-access:"
	accessRevisionOwnerKey  = "authz"
	accessRevisionNamespace = "permission-access"
	accessRevisionCacheKey  = "topology-revision"
	// Refresh the shared revision infrequently because permission topology changes
	// are rare, while local invalidation still takes effect immediately on writes.
	accessRevisionRefreshInterval = 3 * time.Second
)

type cachedUserAccessContext struct {
	UserID   int                // UserID owns the cached access context.
	Revision int64              // Revision is the permission topology version used to build the cache entry.
	Access   *UserAccessContext // Access is the effective access snapshot for the token.
}

var accessCacheState = struct {
	sync.RWMutex
	tokenUsers map[string]int
	userTokens map[int]map[string]struct{}
}{
	tokenUsers: map[string]int{},
	userTokens: map[int]map[string]struct{}{},
}

var accessRevisionState = struct {
	sync.RWMutex
	value    int64
	expireAt time.Time
}{}

// PrimeTokenAccessContext preloads the access context cache for one freshly issued login token.
func (s *serviceImpl) PrimeTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
) (*UserAccessContext, error) {
	if tokenID == "" || userID <= 0 {
		return nil, nil
	}
	return s.getTokenAccessContext(ctx, tokenID, userID)
}

// InvalidateTokenAccessContext removes the cached access context bound to one token.
func (s *serviceImpl) InvalidateTokenAccessContext(ctx context.Context, tokenID string) {
	if tokenID == "" {
		return
	}

	if _, err := gcache.Remove(ctx, accessCacheKey(tokenID)); err != nil {
		logger.Warningf(ctx, "remove token access cache failed tokenID=%s err=%v", tokenID, err)
	}
	s.removeIndexedToken(tokenID)
}

// InvalidateUserAccessContexts removes all cached access contexts bound to one user.
func (s *serviceImpl) InvalidateUserAccessContexts(ctx context.Context, userID int) {
	if userID <= 0 {
		return
	}

	var tokenIDs []string
	accessCacheState.Lock()
	if boundTokens, ok := accessCacheState.userTokens[userID]; ok {
		tokenIDs = make([]string, 0, len(boundTokens))
		for tokenID := range boundTokens {
			tokenIDs = append(tokenIDs, tokenID)
			delete(accessCacheState.tokenUsers, tokenID)
		}
		delete(accessCacheState.userTokens, userID)
	}
	accessCacheState.Unlock()

	if len(tokenIDs) == 0 {
		return
	}

	keys := make([]any, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		keys = append(keys, accessCacheKey(tokenID))
	}
	if err := gcache.Removes(ctx, keys); err != nil {
		logger.Warningf(ctx, "remove user access caches failed userID=%d err=%v", userID, err)
	}
}

// MarkAccessTopologyChanged bumps the shared permission topology revision and clears local token caches.
func (s *serviceImpl) MarkAccessTopologyChanged(ctx context.Context) error {
	s.clearLocalAccessCache(ctx)
	s.clearLocalAccessRevision()

	item, err := s.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		accessRevisionOwnerKey,
		accessRevisionNamespace,
		accessRevisionCacheKey,
		1,
		0,
	)
	if err != nil {
		return err
	}

	s.storeLocalAccessRevision(item.IntValue)
	return nil
}

// NotifyAccessTopologyChanged best-effort refreshes the shared permission topology revision.
func (s *serviceImpl) NotifyAccessTopologyChanged(ctx context.Context) {
	if err := s.MarkAccessTopologyChanged(ctx); err != nil {
		logger.Warningf(ctx, "update access topology revision failed: %v", err)
	}
}

// getTokenAccessContext returns one token-scoped access snapshot that stays
// valid only while the shared topology revision matches the cached entry.
func (s *serviceImpl) getTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
) (*UserAccessContext, error) {
	revision, err := s.getAccessRevision(ctx)
	if err != nil {
		return nil, err
	}

	if cached := s.getCachedTokenAccessContext(ctx, tokenID, userID, revision); cached != nil {
		return cached, nil
	}

	loaded, err := s.loadUserAccessContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cacheTokenAccessContext(ctx, tokenID, userID, revision, loaded)
	return cloneUserAccessContext(loaded), nil
}

// getCachedTokenAccessContext returns one cached snapshot only when the token,
// user, and topology revision all still point to the same effective grants.
func (s *serviceImpl) getCachedTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
	revision int64,
) *UserAccessContext {
	cachedVar, err := gcache.Get(ctx, accessCacheKey(tokenID))
	if err != nil || cachedVar == nil {
		return nil
	}

	cached, ok := cachedVar.Val().(*cachedUserAccessContext)
	if !ok || cached == nil || cached.Access == nil {
		return nil
	}
	if cached.UserID != userID || cached.Revision != revision {
		return nil
	}
	return cloneUserAccessContext(cached.Access)
}

// cacheTokenAccessContext stores one detached access snapshot and indexes the
// token so later logout or user-level invalidation can remove all bound entries.
func (s *serviceImpl) cacheTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
	revision int64,
	access *UserAccessContext,
) {
	if tokenID == "" || userID <= 0 || access == nil {
		return
	}

	cached := &cachedUserAccessContext{
		UserID:   userID,
		Revision: revision,
		Access:   cloneUserAccessContext(access),
	}
	if err := gcache.Set(ctx, accessCacheKey(tokenID), cached, s.resolveAccessCacheTTL(ctx)); err != nil {
		logger.Warningf(ctx, "set token access cache failed tokenID=%s err=%v", tokenID, err)
	}

	accessCacheState.Lock()
	accessCacheState.tokenUsers[tokenID] = userID
	if _, ok := accessCacheState.userTokens[userID]; !ok {
		accessCacheState.userTokens[userID] = make(map[string]struct{})
	}
	accessCacheState.userTokens[userID][tokenID] = struct{}{}
	accessCacheState.Unlock()
}

// clearLocalAccessCache drops all token snapshots held by the current process
// after one topology mutation so subsequent requests rebuild fresh grants.
func (s *serviceImpl) clearLocalAccessCache(ctx context.Context) {
	var tokenIDs []string

	accessCacheState.Lock()
	tokenIDs = make([]string, 0, len(accessCacheState.tokenUsers))
	for tokenID := range accessCacheState.tokenUsers {
		tokenIDs = append(tokenIDs, tokenID)
	}
	accessCacheState.tokenUsers = map[string]int{}
	accessCacheState.userTokens = map[int]map[string]struct{}{}
	accessCacheState.Unlock()

	if len(tokenIDs) == 0 {
		return
	}

	keys := make([]any, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		keys = append(keys, accessCacheKey(tokenID))
	}
	if err := gcache.Removes(ctx, keys); err != nil {
		logger.Warningf(ctx, "clear local access cache failed err=%v", err)
	}
}

// removeIndexedToken removes one token from the local reverse indexes that map
// token IDs back to their owning user for bulk invalidation.
func (s *serviceImpl) removeIndexedToken(tokenID string) {
	accessCacheState.Lock()
	defer accessCacheState.Unlock()

	userID, ok := accessCacheState.tokenUsers[tokenID]
	if !ok {
		return
	}
	delete(accessCacheState.tokenUsers, tokenID)

	boundTokens := accessCacheState.userTokens[userID]
	if boundTokens == nil {
		return
	}
	delete(boundTokens, tokenID)
	if len(boundTokens) == 0 {
		delete(accessCacheState.userTokens, userID)
	}
}

// getAccessRevision returns the current permission-topology revision. It first
// uses the short-lived local copy and falls back to the shared KV row when needed.
func (s *serviceImpl) getAccessRevision(ctx context.Context) (int64, error) {
	if revision, ok := s.getLocalAccessRevision(); ok {
		return revision, nil
	}

	// delta=0 means "read-or-initialize" for the shared integer key without
	// bumping the revision, so readers can observe a stable cross-instance value.
	item, err := s.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		accessRevisionOwnerKey,
		accessRevisionNamespace,
		accessRevisionCacheKey,
		0,
		0,
	)
	if err != nil {
		if revision, ok := s.getLocalAccessRevisionForce(); ok {
			return revision, nil
		}
		return 0, err
	}

	s.storeLocalAccessRevision(item.IntValue)
	return item.IntValue, nil
}

// getLocalAccessRevision returns the process-local revision only while its
// refresh window is still valid.
func (s *serviceImpl) getLocalAccessRevision() (int64, bool) {
	accessRevisionState.RLock()
	defer accessRevisionState.RUnlock()

	if accessRevisionState.expireAt.IsZero() || time.Now().After(accessRevisionState.expireAt) {
		return 0, false
	}
	return accessRevisionState.value, true
}

// getLocalAccessRevisionForce returns the last known local revision even after
// the refresh window expires so transient shared-cache failures can degrade softly.
func (s *serviceImpl) getLocalAccessRevisionForce() (int64, bool) {
	accessRevisionState.RLock()
	defer accessRevisionState.RUnlock()

	if accessRevisionState.expireAt.IsZero() {
		return 0, false
	}
	return accessRevisionState.value, true
}

// storeLocalAccessRevision records the shared revision in process memory so hot
// permission checks do not hit the shared KV cache on every request.
func (s *serviceImpl) storeLocalAccessRevision(revision int64) {
	accessRevisionState.Lock()
	accessRevisionState.value = revision
	accessRevisionState.expireAt = time.Now().Add(accessRevisionRefreshInterval)
	accessRevisionState.Unlock()
}

// clearLocalAccessRevision drops the process-local revision so the next read
// must resynchronize after a local topology write.
func (s *serviceImpl) clearLocalAccessRevision() {
	accessRevisionState.Lock()
	accessRevisionState.value = 0
	accessRevisionState.expireAt = time.Time{}
	accessRevisionState.Unlock()
}

// resolveAccessTokenID extracts the current login token ID from the business
// context so access snapshots can be cached per issued session.
func (s *serviceImpl) resolveAccessTokenID(ctx context.Context) string {
	if s == nil || s.bizCtxSvc == nil {
		return ""
	}
	businessCtx := s.bizCtxSvc.Get(ctx)
	if businessCtx == nil {
		return ""
	}
	return businessCtx.TokenId
}

// resolveAccessCacheTTL keeps token snapshots no longer than either the JWT or
// online-session lifetime because either expiry makes the cache unreachable.
func (s *serviceImpl) resolveAccessCacheTTL(ctx context.Context) time.Duration {
	if s == nil || s.configSvc == nil {
		return 24 * time.Hour
	}

	var (
		jwtTTL     = s.configSvc.GetJwt(ctx).Expire
		sessionTTL = s.configSvc.GetSession(ctx).Timeout
	)
	if sessionTTL < jwtTTL {
		return sessionTTL
	}
	return jwtTTL
}

// accessCacheKey builds the token-scoped cache key used by gcache.
func accessCacheKey(tokenID string) string {
	return accessCacheKeyPrefix + tokenID
}

// cloneUserAccessContext returns a deep-enough copy so request-scoped mutation
// never leaks back into the shared token snapshot.
func cloneUserAccessContext(access *UserAccessContext) *UserAccessContext {
	if access == nil {
		return nil
	}
	return &UserAccessContext{
		RoleIds:      append([]int(nil), access.RoleIds...),
		RoleNames:    append([]string(nil), access.RoleNames...),
		MenuIds:      append([]int(nil), access.MenuIds...),
		Permissions:  append([]string(nil), access.Permissions...),
		IsSuperAdmin: access.IsSuperAdmin,
	}
}
