// This file implements protected-config snapshot caching backed by one
// process-local gcache entry plus a deployment-aware revision source:
// single-node mode keeps the revision in process memory, while clustered mode
// synchronizes the revision through sys_kv_cache.

package config

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/logger"
)

// Runtime parameter snapshot cache keys and synchronization intervals.
const (
	runtimeParamRevisionOwnerKey     = "runtime-config"
	runtimeParamRevisionNamespace    = "sys-config"
	runtimeParamRevisionLogicalKey   = "revision"
	runtimeParamSnapshotCacheKey     = "runtime-param-snapshot"
	runtimeParamSnapshotCacheTTL     = time.Hour
	runtimeParamRevisionSyncInterval = 10 * time.Second
)

// runtimeParamRevisionCacheKey is the shared KV key used to synchronize the
// effective protected-config revision across nodes.
var runtimeParamRevisionCacheKey = kvcache.BuildCacheKey(
	runtimeParamRevisionOwnerKey,
	runtimeParamRevisionNamespace,
	runtimeParamRevisionLogicalKey,
)

// runtimeParamSnapshot stores one immutable parsed view of all protected
// runtime parameters for a single effective revision.
type runtimeParamSnapshot struct {
	revision              int64
	values                map[string]string
	durationValues        map[string]time.Duration
	int64Values           map[string]int64
	parseErrors           map[string]error
	loginBlackIPList      []string
	loginBlacklistMatcher *loginBlacklistMatcher
}

// cachedRuntimeParamSnapshot wraps one immutable runtime snapshot with local
// cache metadata used by the process cache layer.
type cachedRuntimeParamSnapshot struct {
	Revision    int64                 // Revision is the effective revision used for this cache entry.
	RefreshedAt time.Time             // RefreshedAt records when this cache entry was rebuilt from sys_config.
	Snapshot    *runtimeParamSnapshot // Snapshot is the immutable parsed runtime-parameter snapshot.
}

// runtimeParamSnapshotCache stores the process-local immutable snapshot cache
// for protected runtime parameters.
var runtimeParamSnapshotCache = gcache.New()

// runtimeParamRevisionState records the latest revision currently visible to
// this process so single-node mode can invalidate stale local snapshots
// without paying the distributed shared-KV coordination cost.
var runtimeParamRevisionState = struct {
	sync.RWMutex
	value       int64
	initialized bool
}{}

// RuntimeParamSnapshotSyncInterval returns the runtime-parameter watcher
// interval used when multi-node sync is enabled.
func RuntimeParamSnapshotSyncInterval() time.Duration {
	return runtimeParamRevisionSyncInterval
}

// MarkRuntimeParamsChanged bumps the effective runtime-parameter revision and
// clears the current process snapshot cache after one protected mutation.
func (s *serviceImpl) MarkRuntimeParamsChanged(ctx context.Context) error {
	revision, err := s.runtimeParamRevisionCtrl.MarkChanged(ctx)
	if err != nil {
		return err
	}

	// Drop the current process cache immediately after a successful write-side
	// revision bump so the mutating node never waits for the watcher cycle.
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
		logger.Warningf(ctx, "clear runtime param snapshot cache failed err=%v", removeErr)
	}
	logger.Debugf(ctx, "runtime param revision bumped revision=%d", revision)
	return nil
}

// NotifyRuntimeParamsChanged best-effort refreshes the effective runtime-parameter revision.
func (s *serviceImpl) NotifyRuntimeParamsChanged(ctx context.Context) {
	if err := s.MarkRuntimeParamsChanged(ctx); err != nil {
		logger.Warningf(ctx, "update runtime param revision failed: %v", err)
	}
}

// SyncRuntimeParamSnapshot synchronizes the process-local runtime-parameter
// snapshot cache with the latest effective revision.
func (s *serviceImpl) SyncRuntimeParamSnapshot(ctx context.Context) error {
	revision, err := s.runtimeParamRevisionCtrl.SyncRevision(ctx)
	if err != nil {
		return err
	}

	cached := s.getCachedRuntimeParamSnapshot(ctx, revision)
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
			if revision, err := s.getRuntimeParamRevision(ctx); err == nil {
				if cached := s.getCachedRuntimeParamSnapshot(ctx, revision); cached != nil {
					snapshot = cached.Snapshot
				}
			} else if cached := s.getCachedRuntimeParamSnapshot(ctx, 0); cached != nil {
				snapshot = cached.Snapshot
			}
		}
	}()

	revision, err := s.getRuntimeParamRevision(ctx)
	if err != nil {
		logger.Warningf(ctx, "get runtime param revision failed: %v", err)
		return nil
	}

	// Validate any existing local entry before entering GetOrSetFuncLock so
	// corrupted values can be removed and rebuilt within the same request.
	if cached := s.getCachedRuntimeParamSnapshot(ctx, revision); cached != nil {
		return cached.Snapshot
	}

	cachedVar, err := runtimeParamSnapshotCache.GetOrSetFuncLock(
		ctx,
		runtimeParamSnapshotCacheKey,
		func(ctx context.Context) (value any, err error) {
			// This callback runs under the cache write lock, which suppresses
			// same-process cache stampedes when the snapshot is cold or invalidated.
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
	latestRevision, revisionErr := s.getRuntimeParamRevision(ctx)
	if revisionErr != nil {
		logger.Warningf(ctx, "refresh runtime param revision failed: %v", revisionErr)
		return nil
	}
	if cached.Revision != latestRevision {
		if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
			logger.Warningf(ctx, "remove stale runtime param snapshot cache failed err=%v", removeErr)
		}
		reloaded, loadErr := s.loadCachedRuntimeParamSnapshot(ctx, latestRevision)
		if loadErr != nil {
			logger.Warningf(ctx, "reload stale runtime param snapshot failed: %v", loadErr)
			return nil
		}
		if err = runtimeParamSnapshotCache.Set(
			ctx, runtimeParamSnapshotCacheKey, reloaded, runtimeParamSnapshotCacheTTL,
		); err != nil {
			logger.Warningf(ctx, "store refreshed runtime param snapshot failed: %v", err)
		}
		return reloaded.Snapshot
	}
	return cached.Snapshot
}

// getRuntimeParamRevision delegates to the constructor-selected controller so
// callers do not need to know whether the revision lives locally or in shared KV.
func (s *serviceImpl) getRuntimeParamRevision(ctx context.Context) (int64, error) {
	return s.runtimeParamRevisionCtrl.CurrentRevision(ctx)
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
// from sys_config for the specified effective revision.
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
func (s *serviceImpl) getCachedRuntimeParamSnapshot(ctx context.Context, revision int64) *cachedRuntimeParamSnapshot {
	cachedVar, err := runtimeParamSnapshotCache.Get(ctx, runtimeParamSnapshotCacheKey)
	if err != nil || cachedVar == nil {
		return nil
	}
	cached := extractCachedRuntimeParamSnapshot(cachedVar.Val())
	if cached != nil && (revision <= 0 || cached.Revision == revision) {
		return cached
	}
	// Remove the broken entry immediately so the next GetOrSetFuncLock call can
	// treat it as a real miss and rebuild the snapshot in the same request. A
	// revision mismatch means a stale single-node rebuild or a watcher-detected
	// cluster refresh already made this entry obsolete.
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); removeErr != nil {
		logger.Warningf(ctx, "remove invalid runtime param snapshot cache failed err=%v", removeErr)
	}
	return nil
}

// getLocalRuntimeParamRevision returns the current process-local revision when initialized.
func getLocalRuntimeParamRevision() (int64, bool) {
	runtimeParamRevisionState.RLock()
	defer runtimeParamRevisionState.RUnlock()

	if !runtimeParamRevisionState.initialized {
		return 0, false
	}
	return runtimeParamRevisionState.value, true
}

// storeLocalRuntimeParamRevision persists one revision in process-local state.
func storeLocalRuntimeParamRevision(revision int64) {
	runtimeParamRevisionState.Lock()
	runtimeParamRevisionState.value = revision
	runtimeParamRevisionState.initialized = true
	runtimeParamRevisionState.Unlock()
}

// bumpLocalRuntimeParamRevision increments or initializes the process-local revision.
func bumpLocalRuntimeParamRevision() int64 {
	runtimeParamRevisionState.Lock()
	defer runtimeParamRevisionState.Unlock()

	if !runtimeParamRevisionState.initialized {
		runtimeParamRevisionState.value = 1
		runtimeParamRevisionState.initialized = true
		return runtimeParamRevisionState.value
	}
	runtimeParamRevisionState.value++
	return runtimeParamRevisionState.value
}

// clearLocalRuntimeParamRevision removes the process-local revision marker.
func clearLocalRuntimeParamRevision() {
	runtimeParamRevisionState.Lock()
	runtimeParamRevisionState.value = 0
	runtimeParamRevisionState.initialized = false
	runtimeParamRevisionState.Unlock()
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

// lookupValue returns the raw configured value for one protected parameter key.
func (snapshot *runtimeParamSnapshot) lookupValue(key string) (string, bool) {
	if snapshot == nil {
		return "", false
	}

	value, ok := snapshot.values[strings.TrimSpace(key)]
	return value, ok
}

// lookupDuration returns one parsed duration override or the parse error that
// was captured while building the snapshot.
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

// lookupInt64 returns one parsed integer override or the parse error that was
// captured while building the snapshot.
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

// loginBlacklist returns a detached copy of the parsed login blacklist rules.
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
