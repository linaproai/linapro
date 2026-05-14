// This file covers local token access-context cache behavior, invalidation,
// and cloning safety for request-scoped mutations.

package role

import (
	"context"
	"github.com/gogf/gf/v2/os/gcache"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/datascope"
)

// fakeRoleConfigService provides deterministic config values for access-cache tests.
type fakeRoleConfigService struct {
	clusterEnabled bool
	jwtExpire      time.Duration
	sessionTimeout time.Duration
}

// GetCluster returns the cluster config used by the test service.
func (f *fakeRoleConfigService) GetCluster(_ context.Context) *hostconfig.ClusterConfig {
	return &hostconfig.ClusterConfig{Enabled: f.clusterEnabled}
}

// GetClusterRedis returns empty Redis coordination settings for tests.
func (f *fakeRoleConfigService) GetClusterRedis(_ context.Context) *hostconfig.ClusterRedisConfig {
	return &hostconfig.ClusterRedisConfig{}
}

// IsClusterEnabled reports the configured cluster mode for the test service.
func (f *fakeRoleConfigService) IsClusterEnabled(_ context.Context) bool {
	return f.clusterEnabled
}

// OverrideClusterEnabledForDialect locks the fake cluster mode for dialect tests.
func (f *fakeRoleConfigService) OverrideClusterEnabledForDialect(value bool) {
	f.clusterEnabled = value
}

// GetJwt returns a JWT config assembled from the test service fields.
func (f *fakeRoleConfigService) GetJwt(_ context.Context) (*hostconfig.JwtConfig, error) {
	expire, err := f.GetJwtExpire(context.Background())
	if err != nil {
		return nil, err
	}
	return &hostconfig.JwtConfig{
		Secret: "test-secret",
		Expire: expire,
	}, nil
}

// GetJwtSecret returns the fixed JWT secret used in tests.
func (f *fakeRoleConfigService) GetJwtSecret(_ context.Context) string {
	return "test-secret"
}

// GetJwtExpire returns the configured JWT expiration or the default test value.
func (f *fakeRoleConfigService) GetJwtExpire(_ context.Context) (time.Duration, error) {
	if f.jwtExpire > 0 {
		return f.jwtExpire, nil
	}
	return 24 * time.Hour, nil
}

// GetPublicFrontend returns an empty frontend config for tests.
func (f *fakeRoleConfigService) GetPublicFrontend(_ context.Context) (*hostconfig.PublicFrontendConfig, error) {
	return &hostconfig.PublicFrontendConfig{}, nil
}

// GetI18n returns the default runtime i18n config used in access-cache tests.
func (f *fakeRoleConfigService) GetI18n(_ context.Context) *hostconfig.I18nConfig {
	return &hostconfig.I18nConfig{
		Default: "zh-CN",
		Enabled: true,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: "zh-CN", NativeName: "简体中文"},
			{Locale: "en-US", NativeName: "English"},
		},
	}
}

// GetLogin returns an empty login config for tests.
func (f *fakeRoleConfigService) GetLogin(_ context.Context) (*hostconfig.LoginConfig, error) {
	return &hostconfig.LoginConfig{}, nil
}

// GetCron returns default cron settings for tests.
func (f *fakeRoleConfigService) GetCron(_ context.Context) (*hostconfig.CronConfig, error) {
	return &hostconfig.CronConfig{
		Shell: hostconfig.CronShellConfig{
			Enabled:   false,
			Supported: true,
		},
		LogRetention: hostconfig.CronLogRetentionConfig{
			Mode:  hostconfig.CronLogRetentionModeDays,
			Value: 30,
		},
	}, nil
}

// IsCronShellEnabled always reports false in tests.
func (f *fakeRoleConfigService) IsCronShellEnabled(_ context.Context) (bool, error) {
	return false, nil
}

// GetCronLogRetention returns the default cron-log retention policy for tests.
func (f *fakeRoleConfigService) GetCronLogRetention(_ context.Context) (*hostconfig.CronLogRetentionConfig, error) {
	return &hostconfig.CronLogRetentionConfig{
		Mode:  hostconfig.CronLogRetentionModeDays,
		Value: 30,
	}, nil
}

// IsLoginIPBlacklisted always reports false in tests.
func (f *fakeRoleConfigService) IsLoginIPBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// GetServerExtensions returns an empty server extensions config for tests.
func (f *fakeRoleConfigService) GetServerExtensions(_ context.Context) *hostconfig.ServerExtensionsConfig {
	return &hostconfig.ServerExtensionsConfig{}
}

// GetLogger returns an empty logger config for tests.
func (f *fakeRoleConfigService) GetLogger(_ context.Context) *hostconfig.LoggerConfig {
	return &hostconfig.LoggerConfig{}
}

// GetMetadata returns an empty metadata config for tests.
func (f *fakeRoleConfigService) GetMetadata(_ context.Context) *hostconfig.MetadataConfig {
	return &hostconfig.MetadataConfig{}
}

// GetMonitor returns an empty monitor config for tests.
func (f *fakeRoleConfigService) GetMonitor(_ context.Context) *hostconfig.MonitorConfig {
	return &hostconfig.MonitorConfig{}
}

// GetHealth returns an empty health config for tests.
func (f *fakeRoleConfigService) GetHealth(_ context.Context) *hostconfig.HealthConfig {
	return &hostconfig.HealthConfig{}
}

// GetShutdown returns an empty shutdown config for tests.
func (f *fakeRoleConfigService) GetShutdown(_ context.Context) *hostconfig.ShutdownConfig {
	return &hostconfig.ShutdownConfig{}
}

// GetScheduler returns an empty scheduler config for tests.
func (f *fakeRoleConfigService) GetScheduler(_ context.Context) *hostconfig.SchedulerConfig {
	return &hostconfig.SchedulerConfig{DefaultTimezone: "UTC"}
}

// GetSchedulerDefaultTimezone returns UTC for deterministic tests.
func (f *fakeRoleConfigService) GetSchedulerDefaultTimezone(_ context.Context) string {
	return "UTC"
}

// GetOpenApi returns an empty OpenAPI config for tests.
func (f *fakeRoleConfigService) GetOpenApi(_ context.Context) *hostconfig.OpenApiConfig {
	return &hostconfig.OpenApiConfig{}
}

// GetPlugin returns an empty plugin config for tests.
func (f *fakeRoleConfigService) GetPlugin(_ context.Context) *hostconfig.PluginConfig {
	return &hostconfig.PluginConfig{}
}

// GetPluginAutoEnable returns no startup auto-enable plugin IDs for tests.
func (f *fakeRoleConfigService) GetPluginAutoEnable(_ context.Context) []string {
	return nil
}

// GetPluginAutoEnableEntries returns no startup auto-enable entries for tests.
func (f *fakeRoleConfigService) GetPluginAutoEnableEntries(_ context.Context) []hostconfig.PluginAutoEnableEntry {
	return nil
}

// GetPluginDynamicStoragePath returns an empty dynamic storage path for tests.
func (f *fakeRoleConfigService) GetPluginDynamicStoragePath(_ context.Context) string {
	return ""
}

// GetSession returns a session config assembled from the test service fields.
func (f *fakeRoleConfigService) GetSession(_ context.Context) (*hostconfig.SessionConfig, error) {
	timeout, err := f.GetSessionTimeout(context.Background())
	if err != nil {
		return nil, err
	}
	return &hostconfig.SessionConfig{
		Timeout:         timeout,
		CleanupInterval: 5 * time.Minute,
	}, nil
}

// GetSessionTimeout returns the configured session timeout or the default test value.
func (f *fakeRoleConfigService) GetSessionTimeout(_ context.Context) (time.Duration, error) {
	if f.sessionTimeout > 0 {
		return f.sessionTimeout, nil
	}
	return 24 * time.Hour, nil
}

// GetUpload returns an empty upload config for tests.
func (f *fakeRoleConfigService) GetUpload(_ context.Context) (*hostconfig.UploadConfig, error) {
	return &hostconfig.UploadConfig{}, nil
}

// GetUploadPath returns an empty upload path for tests.
func (f *fakeRoleConfigService) GetUploadPath(_ context.Context) string {
	return ""
}

// GetUploadMaxSize returns zero because upload size is irrelevant to these tests.
func (f *fakeRoleConfigService) GetUploadMaxSize(_ context.Context) (int64, error) {
	return 0, nil
}

// MarkRuntimeParamsChanged is a no-op success stub for tests.
func (f *fakeRoleConfigService) MarkRuntimeParamsChanged(_ context.Context) error {
	return nil
}

// NotifyRuntimeParamsChanged is a no-op stub for tests.
func (f *fakeRoleConfigService) NotifyRuntimeParamsChanged(_ context.Context) {}

// SyncRuntimeParamSnapshot is a no-op success stub for tests.
func (f *fakeRoleConfigService) SyncRuntimeParamSnapshot(_ context.Context) error {
	return nil
}

// fakeRoleCacheCoordService provides deterministic cachecoord behavior for
// access revision controller tests.
type fakeRoleCacheCoordService struct {
	revision     int64
	currentErr   error
	markErr      error
	currentCalls int32
	markCalls    int32
	currentScope cachecoord.Scope
	markScope    cachecoord.Scope
	markTenant   cachecoord.InvalidationScope
}

// ConfigureDomain is a no-op because these tests configure domain metadata elsewhere.
func (f *fakeRoleCacheCoordService) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged returns the configured shared revision and tracks publish calls.
func (f *fakeRoleCacheCoordService) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
	_ cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	f.markScope = scope
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.revision, nil
}

// MarkTenantChanged mirrors MarkChanged for tenant-scoped revision tests.
func (f *fakeRoleCacheCoordService) MarkTenantChanged(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
	tenantScope cachecoord.InvalidationScope,
	_ cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	f.markScope = scope
	f.markTenant = tenantScope
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.revision, nil
}

// EnsureFresh runs the refresher against the configured revision.
func (f *fakeRoleCacheCoordService) EnsureFresh(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	refresher cachecoord.Refresher,
) (int64, error) {
	revision, err := f.CurrentRevision(ctx, domain, scope)
	if err != nil {
		return 0, err
	}
	if refresher != nil {
		if err = refresher(ctx, revision); err != nil {
			return 0, err
		}
	}
	return revision, nil
}

// CurrentRevision returns the configured shared revision and tracks read calls.
func (f *fakeRoleCacheCoordService) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	scope cachecoord.Scope,
) (int64, error) {
	atomic.AddInt32(&f.currentCalls, 1)
	f.currentScope = scope
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.revision, nil
}

// Snapshot is unused by access revision tests.
func (f *fakeRoleCacheCoordService) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
}

// fakeAccessRevisionController provides deterministic revision behavior for
// service-level access-cache tests that do not need the concrete controller.
type fakeAccessRevisionController struct {
	revision     int64
	currentErr   error
	syncErr      error
	markErr      error
	currentCalls int32
	syncCalls    int32
	markCalls    int32
}

// CurrentRevision returns the configured current revision.
func (f *fakeAccessRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	atomic.AddInt32(&f.currentCalls, 1)
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.revision, nil
}

// SyncRevision returns the configured revision after optionally invoking the callback.
func (f *fakeAccessRevisionController) SyncRevision(_ context.Context, onRevisionChange func()) (int64, error) {
	atomic.AddInt32(&f.syncCalls, 1)
	if f.syncErr != nil {
		return 0, f.syncErr
	}
	if onRevisionChange != nil {
		onRevisionChange()
	}
	return f.revision, nil
}

// MarkChanged returns the configured changed revision.
func (f *fakeAccessRevisionController) MarkChanged(_ context.Context) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.revision, nil
}

// resetRoleAccessCacheTestState clears process-local access cache state before
// and after each test case.
func resetRoleAccessCacheTestState(t *testing.T, svc *serviceImpl) {
	t.Helper()

	ctx := context.Background()
	accessContextCache = gcache.New()
	svc.clearLocalAccessCache(ctx)
	clearLocalAccessRevision()
	t.Cleanup(func() {
		accessContextCache = gcache.New()
		svc.clearLocalAccessCache(ctx)
		clearLocalAccessRevision()
	})
}

// TestNewCacheCoordAccessRevisionControllerSelectsByClusterMode verifies controller
// selection switches between local and cluster implementations.
func TestNewCacheCoordAccessRevisionControllerSelectsByClusterMode(t *testing.T) {
	if _, ok := newCacheCoordAccessRevisionController(false).(*localAccessRevisionController); !ok {
		t.Fatal("expected single-node mode to use local access revision controller")
	}

	controller, ok := newCacheCoordAccessRevisionController(true).(*clusterAccessRevisionController)
	if !ok {
		t.Fatal("expected cluster mode to use shared access revision controller")
	}
	if controller.cacheCoordSvc == nil {
		t.Fatal("expected cluster access revision controller to use cachecoord")
	}
}

// TestTokenAccessContextCacheLifecycle verifies token cache reads, invalidation,
// and revision mismatch eviction.
func TestTokenAccessContextCacheLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	tokenID := "token-cache-lifecycle"
	userID := 101
	access := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "editor"},
		MenuIds:      []int{10, 11},
		Permissions:  []string{"system:user:query", "system:user:edit"},
		DataScope:    datascope.ScopeDept,
		IsSuperAdmin: false,
	}

	svc.cacheTokenAccessContext(ctx, tokenID, userID, 7, access)

	cached := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7)
	if !reflect.DeepEqual(cached, access) {
		t.Fatalf("expected cached access %#v, got %#v", access, cached)
	}

	// Returned snapshots must be detached from the cached entry so request-level
	// mutations do not leak into the shared token cache.
	cached.Permissions[0] = "mutated"
	reloaded := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7)
	if reloaded == nil || reloaded.Permissions[0] != "system:user:query" {
		t.Fatalf("expected cached permissions to stay immutable, got %#v", reloaded)
	}

	if stale := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 8); stale != nil {
		t.Fatalf("expected revision mismatch to force cache miss, got %#v", stale)
	}

	svc.InvalidateTokenAccessContext(ctx, tokenID)
	if stale := svc.getCachedTokenAccessContext(ctx, tokenID, userID, 7); stale != nil {
		t.Fatalf("expected invalidated token cache to be empty, got %#v", stale)
	}
}

// TestInvalidateUserAccessContextsRemovesBoundTokensOnly verifies per-user
// invalidation clears only the target user's cached token entries.
func TestInvalidateUserAccessContextsRemovesBoundTokensOnly(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	sharedAccess := &UserAccessContext{
		Permissions: []string{"system:role:auth"},
	}

	svc.cacheTokenAccessContext(ctx, "user-1-token-a", 1, 3, sharedAccess)
	svc.cacheTokenAccessContext(ctx, "user-1-token-b", 1, 3, sharedAccess)
	svc.cacheTokenAccessContext(ctx, "user-2-token-a", 2, 3, sharedAccess)

	svc.InvalidateUserAccessContexts(ctx, 1)

	if access := svc.getCachedTokenAccessContext(ctx, "user-1-token-a", 1, 3); access != nil {
		t.Fatalf("expected first token for invalidated user to be removed, got %#v", access)
	}
	if access := svc.getCachedTokenAccessContext(ctx, "user-1-token-b", 1, 3); access != nil {
		t.Fatalf("expected second token for invalidated user to be removed, got %#v", access)
	}
	if access := svc.getCachedTokenAccessContext(ctx, "user-2-token-a", 2, 3); access == nil {
		t.Fatal("expected other users' cached tokens to remain available")
	}
}

// TestInvalidateUserAccessContextsIsTenantScoped verifies user invalidation
// removes only cache entries that belong to the current tenant bucket.
func TestInvalidateUserAccessContextsIsTenantScoped(t *testing.T) {
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	tenantOneCtx := datascope.WithTenantForTest(context.Background(), 1)
	tenantTwoCtx := datascope.WithTenantForTest(context.Background(), 2)
	sharedAccess := &UserAccessContext{
		Permissions: []string{"system:role:auth"},
	}

	svc.cacheTokenAccessContext(tenantOneCtx, "tenant-1-token", 1, 3, sharedAccess)
	svc.cacheTokenAccessContext(tenantTwoCtx, "tenant-2-token", 1, 3, sharedAccess)

	svc.InvalidateUserAccessContexts(tenantOneCtx, 1)

	if access := svc.getCachedTokenAccessContext(tenantOneCtx, "tenant-1-token", 1, 3); access != nil {
		t.Fatalf("expected tenant one token to be removed, got %#v", access)
	}
	if access := svc.getCachedTokenAccessContext(tenantTwoCtx, "tenant-2-token", 1, 3); access == nil {
		t.Fatal("expected tenant two token for same user to remain available")
	}
}

// TestAccessCacheKeyUsesTenantcapBucket verifies role access cache keys use the
// canonical tenantcap tenant/scope/key format.
func TestAccessCacheKeyUsesTenantcapBucket(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 42)
	key := accessCacheKey(ctx, "issued-token")

	if key != "tenant=42:scope=role:user-access:key=issued-token" {
		t.Fatalf("expected tenantcap cache key, got %q", key)
	}
}

// TestTokenAccessContextCacheIsTenantBucketed verifies same token values in
// different tenants do not collide in the process-local role access cache.
func TestTokenAccessContextCacheIsTenantBucketed(t *testing.T) {
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	tenantOneCtx := datascope.WithTenantForTest(context.Background(), 1)
	tenantTwoCtx := datascope.WithTenantForTest(context.Background(), 2)
	tokenID := "shared-token-id"

	svc.cacheTokenAccessContext(tenantOneCtx, tokenID, 11, 5, &UserAccessContext{
		Permissions: []string{"system:tenant-cache:one"},
	})
	svc.cacheTokenAccessContext(tenantTwoCtx, tokenID, 22, 5, &UserAccessContext{
		Permissions: []string{"system:tenant-cache:two"},
	})

	tenantOneAccess := svc.getCachedTokenAccessContext(tenantOneCtx, tokenID, 11, 5)
	if tenantOneAccess == nil || len(tenantOneAccess.Permissions) != 1 || tenantOneAccess.Permissions[0] != "system:tenant-cache:one" {
		t.Fatalf("expected tenant one cached access, got %#v", tenantOneAccess)
	}

	tenantTwoAccess := svc.getCachedTokenAccessContext(tenantTwoCtx, tokenID, 22, 5)
	if tenantTwoAccess == nil || len(tenantTwoAccess.Permissions) != 1 || tenantTwoAccess.Permissions[0] != "system:tenant-cache:two" {
		t.Fatalf("expected tenant two cached access, got %#v", tenantTwoAccess)
	}
}

// TestCloneUserAccessContextCopiesSlices verifies cloned access contexts do not
// share slice backing storage with the original value.
func TestCloneUserAccessContextCopiesSlices(t *testing.T) {
	original := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "ops"},
		MenuIds:      []int{10, 20},
		Permissions:  []string{"user:list", "user:update"},
		DataScope:    datascope.ScopeDept,
		IsSuperAdmin: true,
	}

	cloned := cloneUserAccessContext(original)
	if cloned == nil {
		t.Fatal("expected cloned access context")
	}

	cloned.RoleIds[0] = 99
	cloned.RoleNames[0] = "guest"
	cloned.MenuIds[0] = 88
	cloned.Permissions[0] = "guest:list"
	cloned.DataScope = datascope.ScopeSelf
	cloned.IsSuperAdmin = false

	if original.RoleIds[0] != 1 {
		t.Fatalf("expected original RoleIds to stay unchanged, got %v", original.RoleIds)
	}
	if original.RoleNames[0] != "admin" {
		t.Fatalf("expected original RoleNames to stay unchanged, got %v", original.RoleNames)
	}
	if original.MenuIds[0] != 10 {
		t.Fatalf("expected original MenuIds to stay unchanged, got %v", original.MenuIds)
	}
	if original.Permissions[0] != "user:list" {
		t.Fatalf("expected original Permissions to stay unchanged, got %v", original.Permissions)
	}
	if original.DataScope != datascope.ScopeDept {
		t.Fatalf("expected original DataScope to stay unchanged, got %d", original.DataScope)
	}
	if !original.IsSuperAdmin {
		t.Fatal("expected original IsSuperAdmin to stay unchanged")
	}
}

// TestCloneSliceWithCopyPreservesNilAndValues verifies the generic slice clone
// helper preserves nil semantics and copies values to a new backing array.
func TestCloneSliceWithCopyPreservesNilAndValues(t *testing.T) {
	if cloned := cloneSliceWithCopy[int](nil); cloned != nil {
		t.Fatalf("expected nil clone for nil slice, got %#v", cloned)
	}

	values := []string{"a", "b"}
	cloned := cloneSliceWithCopy(values)
	if len(cloned) != len(values) {
		t.Fatalf("expected cloned length %d, got %d", len(values), len(cloned))
	}
	if &cloned[0] == &values[0] {
		t.Fatal("expected cloned slice to have independent backing array")
	}
}

// TestGetAccessRevisionUsesPureReadPath verifies reading the cluster revision
// does not trigger mutation operations and benefits from local caching.
func TestGetAccessRevisionUsesPureReadPath(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeCoord := &fakeRoleCacheCoordService{revision: 9}
	svc.accessRevisionCtrl = &clusterAccessRevisionController{cacheCoordSvc: fakeCoord}

	revision, err := svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeCoord.currentCalls) != 1 {
		t.Fatalf("expected exactly one cachecoord read, got %d", atomic.LoadInt32(&fakeCoord.currentCalls))
	}
	if atomic.LoadInt32(&fakeCoord.markCalls) != 0 {
		t.Fatalf("expected no cachecoord writes for read path, got %d", atomic.LoadInt32(&fakeCoord.markCalls))
	}

	revision, err = svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("second get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected cached revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeCoord.currentCalls) != 1 {
		t.Fatalf("expected cached local revision to avoid extra cachecoord reads, got %d", atomic.LoadInt32(&fakeCoord.currentCalls))
	}
}

// TestSyncAccessTopologyRevisionKeepsCacheWhenRevisionUnchanged verifies cache
// contents survive synchronization when the shared revision is unchanged.
func TestSyncAccessTopologyRevisionKeepsCacheWhenRevisionUnchanged(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeCoord := &fakeRoleCacheCoordService{revision: 7}
	svc.accessRevisionCtrl = &clusterAccessRevisionController{cacheCoordSvc: fakeCoord}
	storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-same-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey(ctx, "sync-same-revision"))
	if err != nil {
		t.Fatalf("get cached access context after unchanged sync: %v", err)
	}
	if cachedVar == nil {
		t.Fatal("expected cached token access context to remain after unchanged revision sync")
	}
}

// TestSyncAccessTopologyRevisionClearsCacheWhenRevisionChanges verifies stale
// cached access contexts are evicted after a revision change is observed.
func TestSyncAccessTopologyRevisionClearsCacheWhenRevisionChanges(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeCoord := &fakeRoleCacheCoordService{revision: 8}
	svc.accessRevisionCtrl = &clusterAccessRevisionController{cacheCoordSvc: fakeCoord}
	storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-new-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey(ctx, "sync-new-revision"))
	if err != nil {
		t.Fatalf("get cached access context after changed sync: %v", err)
	}
	if cachedVar != nil {
		t.Fatal("expected stale token access context to be evicted after revision change")
	}

	revision, ok := getLocalAccessRevision()
	if !ok {
		t.Fatal("expected synced revision to remain locally cached")
	}
	if revision != 8 {
		t.Fatalf("expected local revision 8 after sync, got %d", revision)
	}
}

// TestClusterAccessRevisionControllerConsumesCrossInstanceRevision verifies a
// watcher-style controller instance consumes a revision published by another
// clustered writer and triggers stale token-cache eviction.
func TestClusterAccessRevisionControllerConsumesCrossInstanceRevision(t *testing.T) {
	ctx := context.Background()
	clearLocalAccessRevision()
	t.Cleanup(clearLocalAccessRevision)
	coordSvc := coordination.NewMemory(nil)

	publisher := &clusterAccessRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}
	consumer := &clusterAccessRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}

	revision, err := publisher.MarkChanged(ctx)
	if err != nil {
		t.Fatalf("publish permission-access revision failed: %v", err)
	}
	storeLocalAccessRevision(revision - 1)

	evicted := false
	observed, err := consumer.SyncRevision(ctx, func() {
		evicted = true
	})
	if err != nil {
		t.Fatalf("consume permission-access revision from second controller failed: %v", err)
	}
	if observed != revision {
		t.Fatalf("expected consumer revision %d, got %d", revision, observed)
	}
	if !evicted {
		t.Fatal("expected revision change to trigger stale token-cache eviction")
	}
}

// TestSingleNodeAccessRevisionStaysLocal verifies single-node mode maintains
// revisions locally without calling cachecoord.
func TestSingleNodeAccessRevisionStaysLocal(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	svc.accessRevisionCtrl = &localAccessRevisionController{}

	revision, err := svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get single-node access revision failed: %v", err)
	}
	if revision != 1 {
		t.Fatalf("expected initial single-node revision 1, got %d", revision)
	}
	svc.cacheTokenAccessContext(ctx, "single-node-token", 1, revision, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err = svc.MarkAccessTopologyChanged(ctx); err != nil {
		t.Fatalf("mark single-node access topology changed failed: %v", err)
	}
	revision, err = svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get single-node access revision after invalidation failed: %v", err)
	}
	if revision != 2 {
		t.Fatalf("expected single-node revision 2 after invalidation, got %d", revision)
	}
	if access := svc.getCachedTokenAccessContext(ctx, "single-node-token", 1, revision); access != nil {
		t.Fatalf("expected single-node token cache to be cleared after invalidation, got %#v", access)
	}
}

// TestClusterAccessRevisionUsesTenantScopedCacheCoord verifies clustered
// permission-access revisions are partitioned by tenant and platform changes
// request tenant cascade invalidation.
func TestClusterAccessRevisionUsesTenantScopedCacheCoord(t *testing.T) {
	fakeCoord := &fakeRoleCacheCoordService{revision: 8}
	controller := &clusterAccessRevisionController{cacheCoordSvc: fakeCoord}

	tenantCtx := datascope.WithTenantForTest(context.Background(), 44)
	if _, err := controller.CurrentRevision(tenantCtx); err != nil {
		t.Fatalf("read tenant access revision failed: %v", err)
	}
	expectedTenantScope := cachecoord.ScopedScope(
		cachecoord.ScopeGlobal,
		cachecoord.InvalidationScope{TenantID: 44},
	)
	if fakeCoord.currentScope != expectedTenantScope {
		t.Fatalf("expected tenant current scope %q, got %q", expectedTenantScope, fakeCoord.currentScope)
	}

	if _, err := controller.MarkChanged(tenantCtx); err != nil {
		t.Fatalf("mark tenant access revision failed: %v", err)
	}
	if fakeCoord.markTenant.TenantID != 44 || fakeCoord.markTenant.CascadeToTenants {
		t.Fatalf("expected tenant-only invalidation metadata, got %#v", fakeCoord.markTenant)
	}

	platformCtx := datascope.WithTenantForTest(context.Background(), datascope.PlatformTenantID)
	if _, err := controller.MarkChanged(platformCtx); err != nil {
		t.Fatalf("mark platform access revision failed: %v", err)
	}
	if fakeCoord.markTenant.TenantID != 0 || !fakeCoord.markTenant.CascadeToTenants {
		t.Fatalf("expected platform cascade invalidation metadata, got %#v", fakeCoord.markTenant)
	}
}

// cleanupPermissionAccessRevision removes the shared permission-access revision
// row used by cross-instance tests.
func cleanupPermissionAccessRevision(t *testing.T, ctx context.Context) {
	t.Helper()

	cleanup := func() {
		if _, err := dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
			Domain: accessTopologyCacheDomain,
			Scope:  cachecoord.ScopeGlobal,
		}).Delete(); err != nil {
			t.Fatalf("cleanup permission-access revision failed: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
}

// TestLoadTokenAccessContextWithCacheLockSuppressesDuplicateLoads verifies one
// cold load is shared across concurrent cache misses.
func TestLoadTokenAccessContextWithCacheLockSuppressesDuplicateLoads(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()
	resetRoleAccessCacheTestState(t, svc)

	svc.accessRevisionCtrl = &fakeAccessRevisionController{revision: 3}

	var loadCalls atomic.Int32
	loader := func(context.Context) (*UserAccessContext, error) {
		loadCalls.Add(1)
		time.Sleep(30 * time.Millisecond)
		return &UserAccessContext{
			RoleIds:      []int{1},
			RoleNames:    []string{"admin"},
			MenuIds:      []int{101},
			Permissions:  []string{"system:user:list"},
			DataScope:    datascope.ScopeAll,
			IsSuperAdmin: true,
		}, nil
	}

	const workers = 8
	results := make(chan *UserAccessContext, workers)
	errs := make(chan error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			access, err := svc.loadTokenAccessContextWithCacheLock(
				ctx,
				"concurrent-token",
				1,
				3,
				loader,
			)
			if err != nil {
				errs <- err
				return
			}
			results <- access
		}()
	}

	close(start)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("load token access context with cache lock failed: %v", err)
		}
	}
	if loadCalls.Load() != 1 {
		t.Fatalf("expected exactly one cold-load execution, got %d", loadCalls.Load())
	}

	count := 0
	for access := range results {
		count++
		if access == nil {
			t.Fatal("expected non-nil access context from cache lock loader")
		}
		if len(access.Permissions) != 1 || access.Permissions[0] != "system:user:list" {
			t.Fatalf("unexpected permissions from cached access context: %#v", access)
		}
		if access.DataScope != datascope.ScopeAll {
			t.Fatalf("unexpected data scope from cached access context: %#v", access)
		}
	}
	if count != workers {
		t.Fatalf("expected %d access results, got %d", workers, count)
	}
}
