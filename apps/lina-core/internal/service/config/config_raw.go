// This file exposes raw host configuration reads for trusted internal
// adapters that need business-neutral access without expanding Service.

package config

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// GetRaw returns the raw host configuration value for key. Empty key and "."
// follow GoFrame semantics and return the full static configuration snapshot.
// Tenant-visible sys_config keys return their runtime-effective values.
func (s *serviceImpl) GetRaw(ctx context.Context, key string) (*gvar.Var, error) {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == RuntimeParamKeyLogRetentionDays {
		value, err := s.getRequiredLogRetentionDaysValue(ctx)
		if err != nil {
			return nil, err
		}
		return gvar.New(value), nil
	}
	if normalizedKey != "" && normalizedKey != "." {
		value, ok, err := s.lookupRuntimeParamValue(ctx, normalizedKey)
		if err != nil {
			return nil, err
		}
		if ok {
			return gvar.New(value), nil
		}
	}
	if IsManagedSysConfigKey(normalizedKey) {
		value, err := s.getProtectedConfigValueOrDefault(ctx, normalizedKey)
		if err != nil {
			return nil, err
		}
		return gvar.New(value), nil
	}
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		return nil, gerror.Wrapf(err, "read host config key failed key=%s", key)
	}
	return value, nil
}

// Built-in runtime parameter keys stored in sys_config.
const (
	// RuntimeParamKeyJWTExpire stores the runtime JWT token lifetime.
	RuntimeParamKeyJWTExpire = "sys.jwt.expire"
	// RuntimeParamKeySessionTimeout stores the runtime online-session inactivity timeout.
	RuntimeParamKeySessionTimeout = "sys.session.timeout"
	// RuntimeParamKeyUploadMaxSize stores the runtime upload size ceiling in MB.
	RuntimeParamKeyUploadMaxSize = "sys.upload.maxSize"
	// RuntimeParamKeyLoginBlackIPList stores the runtime login IP blacklist.
	RuntimeParamKeyLoginBlackIPList = "sys.login.blackIPList"
	// RuntimeParamKeyLogRetentionDays stores the maximum log retention period in days.
	RuntimeParamKeyLogRetentionDays = "sys.log.retentionDays"
	// RuntimeParamKeyCronShellEnabled stores the global shell-job enable switch.
	RuntimeParamKeyCronShellEnabled = "sys.cron.shell.enabled"
	// RuntimeParamKeyCronLogRetention stores the default cron-log retention policy.
	RuntimeParamKeyCronLogRetention = "sys.cron.log.retention"
)

// RuntimeParamSpec describes one built-in runtime parameter managed through
// sys_config.
type RuntimeParamSpec struct {
	Key          string // Key is the sys_config key consumed by host runtime paths.
	DefaultValue string // DefaultValue is the host fallback value.
}

// runtimeParamSpecs lists all built-in runtime parameters backed by sys_config.
var runtimeParamSpecs = []RuntimeParamSpec{
	{
		Key:          RuntimeParamKeyJWTExpire,
		DefaultValue: "24h",
	},
	{
		Key:          RuntimeParamKeySessionTimeout,
		DefaultValue: "24h",
	},
	{
		Key:          RuntimeParamKeyUploadMaxSize,
		DefaultValue: "100",
	},
	{
		Key:          RuntimeParamKeyLoginBlackIPList,
		DefaultValue: "",
	},
	{
		Key:          RuntimeParamKeyLogRetentionDays,
		DefaultValue: "90",
	},
	{
		Key:          RuntimeParamKeyCronShellEnabled,
		DefaultValue: "true",
	},
	{
		Key:          RuntimeParamKeyCronLogRetention,
		DefaultValue: `{"mode":"days","value":30}`,
	},
}

// runtimeParamSpecByKey indexes runtimeParamSpecs by key for validation and
// lookup operations on built-in runtime settings.
var runtimeParamSpecByKey = func() map[string]RuntimeParamSpec {
	specByKey := make(map[string]RuntimeParamSpec, len(runtimeParamSpecs))
	for _, spec := range runtimeParamSpecs {
		specByKey[spec.Key] = spec
	}
	return specByKey
}()

// runtimeParamKeys preserves the deterministic built-in runtime-parameter key order.
var runtimeParamKeys = []string{
	RuntimeParamKeyJWTExpire,
	RuntimeParamKeySessionTimeout,
	RuntimeParamKeyUploadMaxSize,
	RuntimeParamKeyLoginBlackIPList,
	RuntimeParamKeyLogRetentionDays,
	RuntimeParamKeyCronShellEnabled,
	RuntimeParamKeyCronLogRetention,
}

// RuntimeParamSpecs returns all built-in runtime parameter specs.
func RuntimeParamSpecs() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(runtimeParamSpecs))
	copy(specs, runtimeParamSpecs)
	return specs
}

// LookupRuntimeParamSpec returns one built-in runtime parameter spec by key.
func LookupRuntimeParamSpec(key string) (RuntimeParamSpec, bool) {
	spec, ok := runtimeParamSpecByKey[strings.TrimSpace(key)]
	return spec, ok
}

// IsManagedRuntimeParamKey reports whether the key belongs to one built-in
// runtime parameter managed through sys_config by the host runtime.
func IsManagedRuntimeParamKey(key string) bool {
	_, ok := LookupRuntimeParamSpec(key)
	return ok
}

// ValidateRuntimeParamValue validates one built-in runtime parameter value.
func ValidateRuntimeParamValue(key string, value string) error {
	switch strings.TrimSpace(key) {
	case RuntimeParamKeyJWTExpire:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeySessionTimeout:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeyUploadMaxSize:
		_, err := validatePositiveInt64Value(key, value)
		return err

	case RuntimeParamKeyLoginBlackIPList:
		return validateIPBlacklistValue(key, value)

	case RuntimeParamKeyLogRetentionDays:
		_, err := validatePositiveInt64Value(key, value)
		return err

	case RuntimeParamKeyCronShellEnabled:
		_, err := parseStrictBoolValue(key, value)
		return err

	case RuntimeParamKeyCronLogRetention:
		return validateCronLogRetentionValue(key, value)
	}
	return nil
}

// lookupRuntimeParamValue reads one sys_config value from the current immutable
// snapshot.
func (s *serviceImpl) lookupRuntimeParamValue(ctx context.Context, key string) (value string, ok bool, err error) {
	snapshot, err := s.getRuntimeParamSnapshot(ctx)
	if err != nil || snapshot == nil {
		return "", false, err
	}
	value, ok = snapshot.lookupValue(key)
	return value, ok, nil
}

// resolveRuntimeDurationOverride returns one runtime duration override when the
// built-in parameter exists, or the current static value when it is absent.
func (s *serviceImpl) resolveRuntimeDurationOverride(
	ctx context.Context,
	key string,
	current time.Duration,
) (time.Duration, error) {
	snapshot, err := s.getRuntimeParamSnapshot(ctx)
	if err != nil {
		return 0, err
	}
	if snapshot == nil {
		return current, nil
	}
	duration, ok, err := snapshot.lookupDuration(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return current, nil
	}
	return duration, nil
}

// resolveRuntimeInt64Override returns one runtime integer override when the
// built-in parameter exists, or the current static value when it is absent.
func (s *serviceImpl) resolveRuntimeInt64Override(
	ctx context.Context,
	key string,
	current int64,
) (int64, error) {
	snapshot, err := s.getRuntimeParamSnapshot(ctx)
	if err != nil {
		return 0, err
	}
	if snapshot == nil {
		return current, nil
	}
	parsed, ok, err := snapshot.lookupInt64(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return current, nil
	}
	return parsed, nil
}

// splitSemicolonValues splits one semicolon-delimited config value into
// trimmed non-empty items.
func splitSemicolonValues(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	values := make([]string, 0, len(parts))
	for _, item := range parts {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}

// validatePositiveDurationValue validates one duration-form runtime parameter.
func validatePositiveDurationValue(key string, value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeConfigParamDurationInvalid, bizerr.P("key", key))
	}
	if duration <= 0 {
		return 0, bizerr.NewCode(CodeConfigParamPositiveRequired, bizerr.P("key", key))
	}
	return duration, nil
}

// validatePositiveInt64Value validates one positive integer runtime parameter.
func validatePositiveInt64Value(key string, value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeConfigParamIntegerInvalid, bizerr.P("key", key))
	}
	if parsed <= 0 {
		return 0, bizerr.NewCode(CodeConfigParamPositiveRequired, bizerr.P("key", key))
	}
	return parsed, nil
}

// validateIPBlacklistValue validates one semicolon-delimited IP blacklist made
// of individual IPs or CIDR ranges.
func validateIPBlacklistValue(key string, value string) error {
	for _, item := range splitSemicolonValues(value) {
		if net.ParseIP(item) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(item); err == nil {
			continue
		}
		return bizerr.NewCode(
			CodeConfigParamIPCIDRInvalid,
			bizerr.P("key", key),
			bizerr.P("value", item),
		)
	}
	return nil
}

// Runtime-configuration cache coordination reasons.
const (
	// runtimeParamCacheDomain coordinates sys_config runtime snapshots.
	runtimeParamCacheDomain cachecoord.Domain = "runtime-config"
	// runtimeParamCacheChangeReason records sys_config mutations.
	runtimeParamCacheChangeReason cachecoord.ChangeReason = "runtime_params_changed"
	// runtimeParamCacheMaxStale is the runtime-config freshness budget.
	runtimeParamCacheMaxStale = 10 * time.Second
)

// runtimeParamRevisionController hides the single-node and clustered revision
// synchronization strategies behind one common contract.
type runtimeParamRevisionController interface {
	// CurrentRevision returns the effective revision currently visible to the process.
	CurrentRevision(ctx context.Context) (int64, error)
	// SyncRevision refreshes the process-local revision from the active source.
	SyncRevision(ctx context.Context) (int64, error)
	// MarkChanged records one sys_config mutation and returns the new revision.
	MarkChanged(ctx context.Context) (int64, error)
}

// localRuntimeParamRevisionController keeps revision ownership entirely inside
// the current process so single-node deployments avoid any cachecoord traffic.
type localRuntimeParamRevisionController struct{}

// clusterRuntimeParamRevisionController coordinates revision changes through
// cachecoord while still caching the last synchronized value in process memory.
type clusterRuntimeParamRevisionController struct {
	cacheCoordSvc cachecoord.Service
}

// newCacheCoordRuntimeParamRevisionController selects the deployment-specific
// revision strategy backed by cachecoord in cluster mode.
func newCacheCoordRuntimeParamRevisionController(clusterEnabled bool) runtimeParamRevisionController {
	if clusterEnabled {
		cacheCoordSvc := cachecoord.Default(cachecoord.NewStaticTopology(true))
		configureRuntimeParamCacheDomain(cacheCoordSvc)
		return &clusterRuntimeParamRevisionController{
			cacheCoordSvc: cacheCoordSvc,
		}
	}
	return &localRuntimeParamRevisionController{}
}

// configureRuntimeParamCacheDomain declares the runtime-config consistency
// contract without making cachecoord own a global domain registry.
func configureRuntimeParamCacheDomain(cacheCoordSvc cachecoord.Service) {
	if cacheCoordSvc == nil {
		return
	}
	if err := cacheCoordSvc.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           runtimeParamCacheDomain,
		AuthoritySource:  "sys_config runtime configuration",
		ConsistencyModel: cachecoord.ConsistencySharedRevision,
		MaxStale:         runtimeParamCacheMaxStale,
		SyncMechanism:    "persistent sys_cache_revision plus request or watcher refresh",
		FailureStrategy:  cachecoord.FailureStrategyReturnVisibleError,
	}); err != nil {
		panic(err)
	}
}

// CurrentRevision lazily initializes the local revision so a single-node
// process can invalidate snapshots without depending on any external store.
func (c *localRuntimeParamRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if revision, ok := getLocalRuntimeParamRevision(); ok {
		return revision, nil
	}
	revision := bumpLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// SyncRevision is equivalent to CurrentRevision in single-node mode because
// there is no remote source that can advance independently of this process.
func (c *localRuntimeParamRevisionController) SyncRevision(ctx context.Context) (int64, error) {
	return c.CurrentRevision(ctx)
}

// MarkChanged advances the in-process revision immediately after one sys_config
// write so subsequent reads rebuild against the new local version.
func (c *localRuntimeParamRevisionController) MarkChanged(_ context.Context) (int64, error) {
	revision := bumpLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// CurrentRevision verifies freshness through cachecoord on the request path so
// sys_config readers do not indefinitely trust a process-local revision.
func (c *clusterRuntimeParamRevisionController) CurrentRevision(ctx context.Context) (int64, error) {
	return c.ensureFresh(ctx)
}

// SyncRevision always refreshes from cachecoord because watcher-driven sync must
// observe cross-node writes even when this process already has a local copy.
func (c *clusterRuntimeParamRevisionController) SyncRevision(ctx context.Context) (int64, error) {
	return c.ensureFresh(ctx)
}

// MarkChanged publishes one cross-node revision bump and then mirrors the new
// value locally so the mutating node does not wait for the next watcher cycle.
func (c *clusterRuntimeParamRevisionController) MarkChanged(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.MarkChanged(
		ctx,
		runtimeParamCacheDomain,
		cachecoord.ScopeGlobal,
		runtimeParamCacheChangeReason,
	)
	if err != nil {
		return 0, err
	}

	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// ensureFresh confirms that the local sys_config snapshot has consumed
// the latest coordinated revision or returns a visible freshness error.
func (c *clusterRuntimeParamRevisionController) ensureFresh(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.EnsureFresh(
		ctx,
		runtimeParamCacheDomain,
		cachecoord.ScopeGlobal,
		func(_ context.Context, revision int64) error {
			storeLocalRuntimeParamRevision(revision)
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	storeLocalRuntimeParamRevision(revision)
	return revision, nil
}

// Runtime parameter snapshot cache keys and synchronization intervals.
const (
	runtimeParamSnapshotCacheKey     = "runtime-param-snapshot"
	runtimeParamSnapshotCacheTTL     = time.Hour
	runtimeParamRevisionSyncInterval = 10 * time.Second
)

// runtimeParamSnapshot stores one immutable parsed view of all sys_config
// values visible to one tenant scope for a single effective revision.
type runtimeParamSnapshot struct {
	revision              int64
	values                map[string]string
	durationValues        map[string]time.Duration
	int64Values           map[string]int64
	parseErrors           map[string]error
	loginBlackIPList      []string
	loginBlacklistMatcher *loginBlacklistMatcher
}

// cachedRuntimeParamSnapshot wraps one immutable sys_config snapshot with local
// cache metadata used by the process cache layer.
type cachedRuntimeParamSnapshot struct {
	Revision    int64                 // Revision is the effective revision used for this cache entry.
	RefreshedAt time.Time             // RefreshedAt records when this cache entry was rebuilt from sys_config.
	Snapshot    *runtimeParamSnapshot // Snapshot is the immutable parsed sys_config snapshot.
}

// runtimeParamSnapshotCache stores the process-local immutable snapshot cache
// for tenant-visible sys_config values.
var runtimeParamSnapshotCache = gcache.New()

// runtimeParamRevisionState records the latest revision currently visible to
// this process so single-node mode can invalidate stale local snapshots
// without paying the distributed cachecoord coordination cost.
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

// MarkRuntimeParamsChanged bumps the effective sys_config revision and clears
// the current process snapshot cache after one configuration mutation.
func (s *serviceImpl) MarkRuntimeParamsChanged(ctx context.Context) error {
	revision, err := s.runtimeParamRevisionCtrl.MarkChanged(ctx)
	if err != nil {
		return err
	}

	// Drop the current process cache immediately after a successful write-side
	// revision bump so the mutating node never waits for the watcher cycle.
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, scopedRuntimeParamSnapshotCacheKey(ctx)); removeErr != nil {
		logger.Warningf(ctx, "clear runtime param snapshot cache failed err=%v", removeErr)
	}
	logger.Debugf(ctx, "runtime param revision bumped revision=%d", revision)
	return nil
}

// NotifyRuntimeParamsChanged best-effort refreshes the effective sys_config revision.
func (s *serviceImpl) NotifyRuntimeParamsChanged(ctx context.Context) {
	if err := s.MarkRuntimeParamsChanged(ctx); err != nil {
		logger.Warningf(ctx, "update runtime param revision failed: %v", err)
	}
}

// SyncRuntimeParamSnapshot synchronizes the process-local sys_config
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
		_, err = runtimeParamSnapshotCache.UpdateExpire(ctx, scopedRuntimeParamSnapshotCacheKey(ctx), runtimeParamSnapshotCacheTTL)
		return err
	}

	loaded, err := s.loadCachedRuntimeParamSnapshot(ctx, revision)
	if err != nil {
		return err
	}
	return runtimeParamSnapshotCache.Set(
		ctx, scopedRuntimeParamSnapshotCacheKey(ctx), loaded, runtimeParamSnapshotCacheTTL,
	)
}

// getRuntimeParamSnapshot returns the latest sys_config snapshot visible
// to the current process. It verifies the effective revision before accepting a
// local cache entry and propagates freshness or cold-load failures to callers.
func (s *serviceImpl) getRuntimeParamSnapshot(ctx context.Context) (*runtimeParamSnapshot, error) {
	revision, err := s.getRuntimeParamRevision(ctx)
	if err != nil {
		logger.Warningf(ctx, "get runtime param revision failed: %v", err)
		return nil, err
	}

	// Validate any existing local entry before entering GetOrSetFuncLock so
	// corrupted values can be removed and rebuilt within the same request.
	if cached := s.getCachedRuntimeParamSnapshot(ctx, revision); cached != nil {
		return cached.Snapshot, nil
	}

	cacheKey := scopedRuntimeParamSnapshotCacheKey(ctx)
	cachedVar, err := runtimeParamSnapshotCache.GetOrSetFuncLock(
		ctx,
		cacheKey,
		func(ctx context.Context) (value any, err error) {
			// This callback runs under the cache write lock, which suppresses
			// same-process cache stampedes when the snapshot is cold or invalidated.
			return s.loadCachedRuntimeParamSnapshot(ctx, revision)
		},
		runtimeParamSnapshotCacheTTL,
	)
	if err != nil {
		logger.Warningf(ctx, "load runtime param snapshot fallback failed: %v", err)
		return nil, err
	}

	cached := extractCachedRuntimeParamSnapshot(cachedVar.Val())
	if cached == nil {
		if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, cacheKey); removeErr != nil {
			logger.Warningf(ctx, "remove invalid runtime param snapshot cache failed err=%v", removeErr)
		}
		return nil, nil
	}
	latestRevision, revisionErr := s.getRuntimeParamRevision(ctx)
	if revisionErr != nil {
		logger.Warningf(ctx, "refresh runtime param revision failed: %v", revisionErr)
		return nil, revisionErr
	}
	if cached.Revision != latestRevision {
		if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, cacheKey); removeErr != nil {
			logger.Warningf(ctx, "remove stale runtime param snapshot cache failed err=%v", removeErr)
		}
		reloaded, loadErr := s.loadCachedRuntimeParamSnapshot(ctx, latestRevision)
		if loadErr != nil {
			logger.Warningf(ctx, "reload stale runtime param snapshot failed: %v", loadErr)
			return nil, loadErr
		}
		if err = runtimeParamSnapshotCache.Set(
			ctx, cacheKey, reloaded, runtimeParamSnapshotCacheTTL,
		); err != nil {
			logger.Warningf(ctx, "store refreshed runtime param snapshot failed: %v", err)
		}
		return reloaded.Snapshot, nil
	}
	return cached.Snapshot, nil
}

// getRuntimeParamRevision delegates to the constructor-selected controller so
// callers do not need to know whether the revision lives locally or in cachecoord.
func (s *serviceImpl) getRuntimeParamRevision(ctx context.Context) (int64, error) {
	if s == nil || s.runtimeParamRevisionCtrl == nil {
		return 0, bizerr.NewCode(CodeConfigRuntimeParamRevisionUnavailable)
	}
	return s.runtimeParamRevisionCtrl.CurrentRevision(ctx)
}

// loadCachedRuntimeParamSnapshot rebuilds one immutable sys_config
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

// loadRuntimeParamSnapshot rebuilds one immutable sys_config snapshot from all
// rows visible to the current tenant scope for the specified effective revision.
func (s *serviceImpl) loadRuntimeParamSnapshot(ctx context.Context, revision int64) (*runtimeParamSnapshot, error) {
	cols := dao.SysConfig.Columns()
	tenantID := datascope.CurrentTenantID(ctx)

	var rows []*entity.SysConfig
	model := dao.SysConfig.Ctx(ctx)
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
	} else {
		model = model.Where(cols.TenantId, datascope.PlatformTenantID)
	}
	err := model.Scan(&rows)
	if err != nil {
		return nil, err
	}

	effectiveRows := effectiveRuntimeParamRows(rows, tenantID)
	snapshot := &runtimeParamSnapshot{
		revision:         revision,
		values:           make(map[string]string, len(effectiveRows)),
		durationValues:   make(map[string]time.Duration, 2),
		int64Values:      make(map[string]int64, 2),
		parseErrors:      make(map[string]error, 3),
		loginBlackIPList: nil,
	}
	for _, row := range effectiveRows {
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

		case RuntimeParamKeyUploadMaxSize, RuntimeParamKeyLogRetentionDays:
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
	cacheKey := scopedRuntimeParamSnapshotCacheKey(ctx)
	cachedVar, err := runtimeParamSnapshotCache.Get(ctx, cacheKey)
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
	if _, removeErr := runtimeParamSnapshotCache.Remove(ctx, cacheKey); removeErr != nil {
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

// scopedRuntimeParamSnapshotCacheKey returns one process-local cache key for
// the tenant-visible sys_config snapshot.
func scopedRuntimeParamSnapshotCacheKey(ctx context.Context) string {
	return datascope.CacheKey(ctx, runtimeParamSnapshotCacheKey, "snapshot")
}

// effectiveRuntimeParamRows collapses platform and tenant sys_config rows into
// one effective view, preferring tenant rows when a tenant context is active.
func effectiveRuntimeParamRows(rows []*entity.SysConfig, tenantID int) []*entity.SysConfig {
	byKey := make(map[string]*entity.SysConfig, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := strings.TrimSpace(row.Key)
		if key == "" {
			continue
		}
		existing := byKey[key]
		if existing == nil || tenantID > datascope.PlatformTenantID &&
			existing.TenantId == datascope.PlatformTenantID && row.TenantId == tenantID {
			byKey[key] = row
		}
	}

	result := make([]*entity.SysConfig, 0, len(byKey))
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := strings.TrimSpace(row.Key)
		if selected := byKey[key]; selected != nil && selected.Id == row.Id {
			result = append(result, row)
		}
	}
	return result
}

// lookupValue returns the raw configured value for one sys_config key.
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
