// This file implements token-scoped access-context caching and permission
// topology revision synchronization for declarative interface authorization.

package role

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/service/kvcache"
	"lina-core/pkg/logger"
)

const (
	accessCacheKeyPrefix       = "role:user-access:"
	accessRevisionOwnerKey     = "authz"
	accessRevisionNamespace    = "permission-access"
	accessRevisionLogicalKey   = "topology-revision"
	accessRevisionSyncInterval = 3 * time.Second
	// Refresh the shared revision infrequently because permission topology changes
	// are rare, while local invalidation still takes effect immediately on writes.
	accessRevisionRefreshInterval = accessRevisionSyncInterval
)

var accessRevisionCacheKey = kvcache.BuildCacheKey(
	accessRevisionOwnerKey,
	accessRevisionNamespace,
	accessRevisionLogicalKey,
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

var accessContextCache = gcache.New()

// AccessRevisionSyncInterval returns the watcher interval used to synchronize
// process-local permission topology revision state on every node.
func AccessRevisionSyncInterval() time.Duration {
	return accessRevisionSyncInterval
}

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

	s.evictTokenAccessContext(ctx, tokenID)
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
	if err := accessContextCache.Removes(ctx, keys); err != nil {
		logger.Warningf(ctx, "remove user access caches failed userID=%d err=%v", userID, err)
	}
}

// MarkAccessTopologyChanged bumps the effective permission-topology revision and clears local token caches.
func (s *serviceImpl) MarkAccessTopologyChanged(ctx context.Context) error {
	s.clearLocalAccessCache(ctx)
	_, err := s.accessRevisionCtrl.MarkChanged(ctx)
	return err
}

// NotifyAccessTopologyChanged best-effort refreshes the effective permission-topology revision.
func (s *serviceImpl) NotifyAccessTopologyChanged(ctx context.Context) {
	if err := s.MarkAccessTopologyChanged(ctx); err != nil {
		logger.Warningf(ctx, "update access topology revision failed: %v", err)
	}
}

// SyncAccessTopologyRevision synchronizes the process-local permission
// topology revision and clears stale token snapshots after cross-node changes.
func (s *serviceImpl) SyncAccessTopologyRevision(ctx context.Context) error {
	_, err := s.accessRevisionCtrl.SyncRevision(ctx, func() {
		// The watcher only needs to clear token-scoped access snapshots when another
		// node has already bumped the shared revision. Once the local revision catches
		// up, requests can keep reading process memory until the next sync window.
		s.clearLocalAccessCache(ctx)
	})
	return err
}

// getTokenAccessContext returns one token-scoped access snapshot that stays
// valid only while the effective topology revision matches the cached entry.
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
	return s.loadTokenAccessContextWithCacheLock(
		ctx,
		tokenID,
		userID,
		revision,
		func(ctx context.Context) (*UserAccessContext, error) {
			return s.loadUserAccessContext(ctx, userID)
		},
	)
}

// getCachedTokenAccessContext returns one cached snapshot only when the token,
// user, and topology revision all still point to the same effective grants.
func (s *serviceImpl) getCachedTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
	revision int64,
) *UserAccessContext {
	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey(tokenID))
	if err != nil || cachedVar == nil {
		return nil
	}

	cached := extractCachedUserAccessContext(cachedVar.Val())
	if cached == nil {
		// Remove corrupted cache payloads eagerly so later requests rebuild a
		// clean token snapshot instead of repeatedly re-reading the bad entry.
		s.evictTokenAccessContext(ctx, tokenID)
		return nil
	}
	if cached.UserID != userID || cached.Revision != revision {
		// A token can only reuse the cached snapshot while both the owner and the
		// topology revision stay aligned. Any mismatch means the cache entry is
		// now stale for this request and should be discarded immediately.
		s.evictTokenAccessContext(ctx, tokenID)
		return nil
	}
	s.indexAccessToken(tokenID, userID)
	return cloneUserAccessContext(cached.Access)
}

// loadTokenAccessContextWithCacheLock serializes same-token cold loads so
// concurrent protected requests do not rebuild the same access snapshot.
func (s *serviceImpl) loadTokenAccessContextWithCacheLock(
	ctx context.Context,
	tokenID string,
	userID int,
	revision int64,
	loader func(context.Context) (*UserAccessContext, error),
) (*UserAccessContext, error) {
	cachedVar, err := accessContextCache.GetOrSetFuncLock(
		ctx,
		accessCacheKey(tokenID),
		func(ctx context.Context) (value any, err error) {
			// The loader runs under one cache-key write lock, so concurrent first
			// requests for the same token share a single access-context rebuild.
			loaded, loadErr := loader(ctx)
			if loadErr != nil {
				return nil, loadErr
			}

			cached := buildCachedUserAccessContext(userID, revision, loaded)
			if cached == nil {
				return nil, gerror.New("token access context loader returned empty snapshot")
			}
			return cached, nil
		},
		s.resolveAccessCacheTTL(ctx),
	)
	if err != nil {
		return nil, err
	}
	if cachedVar == nil {
		return nil, gerror.New("token access context cache returned empty result")
	}

	cached := extractCachedUserAccessContext(cachedVar.Val())
	latestRevision, revisionErr := s.getAccessRevision(ctx)
	if revisionErr == nil {
		revision = latestRevision
	}
	if cached == nil || cached.UserID != userID || cached.Revision != revision {
		// The shared lock prevents duplicate cold loads for one token, but the
		// entry can still become stale before this goroutine resumes, for example
		// when a concurrent topology write clears local token snapshots. Rebuild
		// once more on the current revision so the caller always gets a fresh view.
		s.evictTokenAccessContext(ctx, tokenID)
		return s.rebuildTokenAccessContext(ctx, tokenID, userID, revision, loader)
	}

	s.indexAccessToken(tokenID, userID)
	return cloneUserAccessContext(cached.Access), nil
}

// rebuildTokenAccessContext loads one fresh snapshot directly and writes it back
// to cache after invalid or stale cache entries have been discarded.
func (s *serviceImpl) rebuildTokenAccessContext(
	ctx context.Context,
	tokenID string,
	userID int,
	revision int64,
	loader func(context.Context) (*UserAccessContext, error),
) (*UserAccessContext, error) {
	loaded, err := loader(ctx)
	if err != nil {
		return nil, err
	}

	s.cacheTokenAccessContext(ctx, tokenID, userID, revision, loaded)
	return cloneUserAccessContext(loaded), nil
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

	cached := buildCachedUserAccessContext(userID, revision, access)
	if cached == nil {
		return
	}
	if err := accessContextCache.Set(
		ctx, accessCacheKey(tokenID), cached, s.resolveAccessCacheTTL(ctx),
	); err != nil {
		logger.Warningf(ctx, "set token access cache failed tokenID=%s err=%v", tokenID, err)
		return
	}
	s.indexAccessToken(tokenID, userID)
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
	if err := accessContextCache.Removes(ctx, keys); err != nil {
		logger.Warningf(ctx, "clear local access cache failed err=%v", err)
	}
}

// evictTokenAccessContext removes one token snapshot from the local cache and
// clears the reverse index so later bulk invalidation stays accurate.
func (s *serviceImpl) evictTokenAccessContext(ctx context.Context, tokenID string) {
	if tokenID == "" {
		return
	}

	if _, err := accessContextCache.Remove(ctx, accessCacheKey(tokenID)); err != nil {
		logger.Warningf(ctx, "remove token access cache failed tokenID=%s err=%v", tokenID, err)
	}
	s.removeIndexedToken(tokenID)
}

// indexAccessToken records the token-to-user relation for one cached access
// snapshot so logout and user-level invalidation can remove all bound entries.
func (s *serviceImpl) indexAccessToken(tokenID string, userID int) {
	if tokenID == "" || userID <= 0 {
		return
	}

	accessCacheState.Lock()
	accessCacheState.tokenUsers[tokenID] = userID
	if _, ok := accessCacheState.userTokens[userID]; !ok {
		accessCacheState.userTokens[userID] = make(map[string]struct{})
	}
	accessCacheState.userTokens[userID][tokenID] = struct{}{}
	accessCacheState.Unlock()
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

// getAccessRevision delegates deployment-specific revision lookup to the
// constructor-selected controller so permission reads keep one consistent path.
func (s *serviceImpl) getAccessRevision(ctx context.Context) (int64, error) {
	return s.accessRevisionCtrl.CurrentRevision(ctx)
}

// getLocalAccessRevision returns the process-local revision only while its
// refresh window is still valid.
func getLocalAccessRevision() (int64, bool) {
	accessRevisionState.RLock()
	defer accessRevisionState.RUnlock()

	if accessRevisionState.expireAt.IsZero() || time.Now().After(accessRevisionState.expireAt) {
		return 0, false
	}
	return accessRevisionState.value, true
}

// getLocalAccessRevisionForce returns the last known local revision even after
// the refresh window expires so transient shared-cache failures can degrade softly.
func getLocalAccessRevisionForce() (int64, bool) {
	accessRevisionState.RLock()
	defer accessRevisionState.RUnlock()

	if accessRevisionState.expireAt.IsZero() {
		return 0, false
	}
	return accessRevisionState.value, true
}

// storeLocalAccessRevision records the shared revision in process memory so hot
// permission checks do not hit the shared KV cache on every request.
func storeLocalAccessRevision(revision int64) {
	accessRevisionState.Lock()
	accessRevisionState.value = revision
	accessRevisionState.expireAt = time.Now().Add(accessRevisionRefreshInterval)
	accessRevisionState.Unlock()
}

// bumpLocalAccessRevision advances the process-local revision while preserving
// the same refresh TTL semantics used by clustered local snapshots.
func bumpLocalAccessRevision() int64 {
	accessRevisionState.Lock()
	defer accessRevisionState.Unlock()

	if accessRevisionState.expireAt.IsZero() {
		accessRevisionState.value = 1
	} else {
		accessRevisionState.value++
	}
	accessRevisionState.expireAt = time.Now().Add(accessRevisionRefreshInterval)
	return accessRevisionState.value
}

// clearLocalAccessRevision drops the process-local revision so the next read
// must resynchronize after a local topology write.
func clearLocalAccessRevision() {
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
		jwtTTL     = s.configSvc.GetJwtExpire(ctx)
		sessionTTL = s.configSvc.GetSessionTimeout(ctx)
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

// buildCachedUserAccessContext detaches one access snapshot from request-local
// slices before the cache stores it for token reuse.
func buildCachedUserAccessContext(
	userID int,
	revision int64,
	access *UserAccessContext,
) *cachedUserAccessContext {
	if userID <= 0 || access == nil {
		return nil
	}
	return &cachedUserAccessContext{
		UserID:   userID,
		Revision: revision,
		Access:   cloneUserAccessContext(access),
	}
}

// extractCachedUserAccessContext keeps cache reads defensive so stale or
// unexpected cache values do not crash permission checks.
func extractCachedUserAccessContext(value any) *cachedUserAccessContext {
	cached, ok := value.(*cachedUserAccessContext)
	if !ok || cached == nil || cached.Access == nil {
		return nil
	}
	return cached
}

// cloneSliceWithCopy allocates the exact target length once and then copies the
// slice content so hot-path access context cloning does not rely on append's
// growth logic for every slice field.
func cloneSliceWithCopy[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]T, len(values))
	copy(cloned, values)
	return cloned
}

// cloneUserAccessContext returns a deep-enough copy so request-scoped mutation
// never leaks back into the shared token snapshot.
func cloneUserAccessContext(access *UserAccessContext) *UserAccessContext {
	if access == nil {
		return nil
	}
	return &UserAccessContext{
		RoleIds:      cloneSliceWithCopy(access.RoleIds),
		RoleNames:    cloneSliceWithCopy(access.RoleNames),
		MenuIds:      cloneSliceWithCopy(access.MenuIds),
		Permissions:  cloneSliceWithCopy(access.Permissions),
		IsSuperAdmin: access.IsSuperAdmin,
	}
}
