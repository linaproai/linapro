// This file implements protected-config snapshot caching backed by one
// process-local gcache entry and a shared revision stored in sys_kv_cache.

package config

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/logger"
)

const (
	runtimeParamRevisionOwnerKey     = "runtime-config"
	runtimeParamRevisionNamespace    = "sys-config"
	runtimeParamRevisionCacheKey     = "revision"
	runtimeParamSnapshotCacheKey     = "runtime-param-snapshot"
	runtimeParamSnapshotCacheTTL     = time.Hour
	runtimeParamRevisionSyncInterval = 10 * time.Second
)

type runtimeParamSnapshot struct {
	revision              int64
	values                map[string]string
	durationValues        map[string]time.Duration
	int64Values           map[string]int64
	parseErrors           map[string]error
	loginBlackIPList      []string
	loginBlacklistMatcher *loginBlacklistMatcher
}

type cachedRuntimeParamSnapshot struct {
	Revision    int64                 // Revision is the shared revision used for this cache entry.
	RefreshedAt time.Time             // RefreshedAt records when this cache entry was rebuilt from sys_config.
	Snapshot    *runtimeParamSnapshot // Snapshot is the immutable parsed runtime-parameter snapshot.
}

var runtimeParamSnapshotCache = gcache.New()

// RuntimeParamSnapshotSyncInterval returns the shared runtime-parameter cache
// watcher interval used by all nodes.
func RuntimeParamSnapshotSyncInterval() time.Duration {
	return runtimeParamRevisionSyncInterval
}

// MarkRuntimeParamsChanged bumps the shared revision and clears the current
// process snapshot cache after one protected runtime parameter mutation.
func (s *serviceImpl) MarkRuntimeParamsChanged(ctx context.Context) error {
	item, err := s.kvCacheSvc.Incr(
		ctx,
		kvcache.OwnerTypeModule,
		runtimeParamRevisionOwnerKey,
		runtimeParamRevisionNamespace,
		runtimeParamRevisionCacheKey,
		1,
		0,
	)
	if err != nil {
		return err
	}

	// Drop the current process cache immediately after a successful write-side
	// revision bump so the mutating node never waits for the watcher cycle.
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
		logger.Warningf(ctx, "clear runtime param snapshot cache failed err=%v", removeErr)
	}
	logger.Debugf(ctx, "runtime param revision bumped revision=%d", item.IntValue)
	return nil
}

// NotifyRuntimeParamsChanged best-effort refreshes the shared runtime-parameter revision.
func (s *serviceImpl) NotifyRuntimeParamsChanged(ctx context.Context) {
	if err := s.MarkRuntimeParamsChanged(ctx); err != nil {
		logger.Warningf(ctx, "update runtime param revision failed: %v", err)
	}
}

// SyncRuntimeParamSnapshot synchronizes the process-local runtime-parameter
// snapshot cache with the latest shared revision.
func (s *serviceImpl) SyncRuntimeParamSnapshot(ctx context.Context) error {
	revision, err := s.getRuntimeParamRevision(ctx)
	if err != nil {
		return err
	}

	cached := s.getCachedRuntimeParamSnapshot(ctx)
	if cached != nil && cached.Snapshot != nil && cached.Revision == revision {
		// When watcher and shared revision are still aligned, only extend the
		// local hard TTL to keep the hot path on process memory.
		_, err = runtimeParamSnapshotCache.UpdateExpire(
			ctx, runtimeParamSnapshotCacheKey, runtimeParamSnapshotCacheTTL,
		)
		return err
	}

	loaded, err := s.loadCachedRuntimeParamSnapshot(ctx, revision)
	if err != nil {
		return err
	}
	return runtimeParamSnapshotCache.Set(
		ctx, runtimeParamSnapshotCacheKey, loaded, runtimeParamSnapshotCacheTTL,
	)
}

// getRuntimeParamSnapshot returns the latest runtime-parameter snapshot visible
// to the current process. It prefers the local cache entry and falls back to a
// single-flight cold load when the cache is empty.
func (s *serviceImpl) getRuntimeParamSnapshot(ctx context.Context) (snapshot *runtimeParamSnapshot) {
	defer func() {
		if recover() != nil {
			// Keep runtime-parameter reads best-effort so config-only tests and
			// degraded environments can still fall back to config.yaml/defaults.
			if cached := s.getCachedRuntimeParamSnapshot(ctx); cached != nil {
				snapshot = cached.Snapshot
			}
		}
	}()

	// Validate any existing local entry before entering GetOrSetFuncLock so
	// corrupted values can be removed and rebuilt within the same request.
	if cached := s.getCachedRuntimeParamSnapshot(ctx); cached != nil {
		return cached.Snapshot
	}

	cachedVar, err := runtimeParamSnapshotCache.GetOrSetFuncLock(
		ctx,
		runtimeParamSnapshotCacheKey,
		func(ctx context.Context) (value any, err error) {
			// This callback runs under the cache write lock, which suppresses
			// same-process cache stampedes when the snapshot is cold or invalidated.
			revision, revisionErr := s.getRuntimeParamRevision(ctx)
			if revisionErr != nil {
				return nil, revisionErr
			}
			return s.loadCachedRuntimeParamSnapshot(ctx, revision)
		},
		runtimeParamSnapshotCacheTTL,
	)
	if err != nil {
		logger.Warningf(ctx, "load runtime param snapshot fallback failed: %v", err)
		return nil
	}

	cached := extractCachedRuntimeParamSnapshot(cachedVar.Val())
	if cached == nil {
		if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
			logger.Warningf(ctx, "remove invalid runtime param snapshot cache failed err=%v", removeErr)
		}
		return nil
	}
	return cached.Snapshot
}

// getRuntimeParamRevision returns the latest shared revision visible to the current node.
func (s *serviceImpl) getRuntimeParamRevision(ctx context.Context) (int64, error) {
	revision, _, err := s.kvCacheSvc.GetInt(
		ctx,
		kvcache.OwnerTypeModule,
		runtimeParamRevisionOwnerKey,
		runtimeParamRevisionNamespace,
		runtimeParamRevisionCacheKey,
	)
	return revision, err
}

// loadCachedRuntimeParamSnapshot rebuilds one immutable runtime-parameter
// snapshot from sys_config and wraps it with cache metadata for local storage.
func (s *serviceImpl) loadCachedRuntimeParamSnapshot(
	ctx context.Context,
	revision int64,
) (*cachedRuntimeParamSnapshot, error) {
	loaded, err := s.loadRuntimeParamSnapshot(ctx, revision)
	if err != nil {
		return nil, err
	}
	return &cachedRuntimeParamSnapshot{
		Revision:    revision,
		RefreshedAt: time.Now(),
		Snapshot:    loaded,
	}, nil
}

// loadRuntimeParamSnapshot rebuilds one immutable protected-config snapshot
// from sys_config for the specified shared revision.
func (s *serviceImpl) loadRuntimeParamSnapshot(ctx context.Context, revision int64) (*runtimeParamSnapshot, error) {
	cols := dao.SysConfig.Columns()

	var rows []*entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		WhereIn(cols.Key, protectedConfigKeys).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	snapshot := &runtimeParamSnapshot{
		revision:         revision,
		values:           make(map[string]string, len(protectedConfigKeys)),
		durationValues:   make(map[string]time.Duration, 2),
		int64Values:      make(map[string]int64, 1),
		parseErrors:      make(map[string]error, 3),
		loginBlackIPList: nil,
	}
	for _, row := range rows {
		if row == nil {
			continue
		}

		key := strings.TrimSpace(row.Key)
		snapshot.values[key] = row.Value
		switch key {
		case RuntimeParamKeyJWTExpire, RuntimeParamKeySessionTimeout:
			if strings.TrimSpace(row.Value) == "" {
				continue
			}
			duration, parseErr := validatePositiveDurationValue(key, row.Value)
			if parseErr != nil {
				snapshot.parseErrors[key] = parseErr
				continue
			}
			snapshot.durationValues[key] = duration

		case RuntimeParamKeyUploadMaxSize:
			if strings.TrimSpace(row.Value) == "" {
				continue
			}
			value, parseErr := validatePositiveInt64Value(key, row.Value)
			if parseErr != nil {
				snapshot.parseErrors[key] = parseErr
				continue
			}
			snapshot.int64Values[key] = value

		case RuntimeParamKeyLoginBlackIPList:
			// Pre-parse blacklist rules once while rebuilding the snapshot so
			// login requests only need to parse the caller IP itself.
			snapshot.loginBlackIPList = splitSemicolonValues(row.Value)
			snapshot.loginBlacklistMatcher = newLoginBlacklistMatcher(snapshot.loginBlackIPList)
		}
	}

	return snapshot, nil
}

// getCachedRuntimeParamSnapshot returns the local cache entry only when the
// cached value still matches the expected snapshot wrapper type.
func (s *serviceImpl) getCachedRuntimeParamSnapshot(ctx context.Context) *cachedRuntimeParamSnapshot {
	cachedVar, err := runtimeParamSnapshotCache.Get(ctx, runtimeParamSnapshotCacheKey)
	if err != nil || cachedVar == nil {
		return nil
	}
	cached := extractCachedRuntimeParamSnapshot(cachedVar.Val())
	if cached != nil {
		return cached
	}
	// Remove the broken entry immediately so the next GetOrSetFuncLock call can
	// treat it as a real miss and rebuild the snapshot in the same request.
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
		logger.Warningf(ctx, "remove invalid runtime param snapshot cache failed err=%v", removeErr)
	}
	return nil
}

// extractCachedRuntimeParamSnapshot keeps cache reads defensive so future
// refactors or unexpected writes cannot poison the runtime-param hot path.
func extractCachedRuntimeParamSnapshot(value any) *cachedRuntimeParamSnapshot {
	cached, ok := value.(*cachedRuntimeParamSnapshot)
	if !ok || cached == nil || cached.Snapshot == nil {
		return nil
	}
	return cached
}

func (snapshot *runtimeParamSnapshot) lookupValue(key string) (string, bool) {
	if snapshot == nil {
		return "", false
	}

	value, ok := snapshot.values[strings.TrimSpace(key)]
	return value, ok
}

func (snapshot *runtimeParamSnapshot) lookupDuration(key string) (time.Duration, bool, error) {
	if snapshot == nil {
		return 0, false, nil
	}

	key = strings.TrimSpace(key)
	if err, ok := snapshot.parseErrors[key]; ok {
		return 0, false, err
	}
	value, ok := snapshot.durationValues[key]
	if !ok {
		return 0, false, nil
	}
	return value, true, nil
}

func (snapshot *runtimeParamSnapshot) lookupInt64(key string) (int64, bool, error) {
	if snapshot == nil {
		return 0, false, nil
	}

	key = strings.TrimSpace(key)
	if err, ok := snapshot.parseErrors[key]; ok {
		return 0, false, err
	}
	value, ok := snapshot.int64Values[key]
	if !ok {
		return 0, false, nil
	}
	return value, true, nil
}

func (snapshot *runtimeParamSnapshot) loginBlacklist() []string {
	if snapshot == nil || len(snapshot.loginBlackIPList) == 0 {
		return nil
	}
	values := make([]string, len(snapshot.loginBlackIPList))
	copy(values, snapshot.loginBlackIPList)
	return values
}

// isLoginIPBlacklisted checks the caller IP against the snapshot-level parsed
// matcher that was prepared when this runtime snapshot was loaded from sys_config.
func (snapshot *runtimeParamSnapshot) isLoginIPBlacklisted(ip string) bool {
	if snapshot == nil {
		return false
	}
	return snapshot.loginBlacklistMatcher.matches(ip)
}
