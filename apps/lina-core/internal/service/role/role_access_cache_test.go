// This file covers local token access-context cache behavior, invalidation,
// and cloning safety for request-scoped mutations.

package role

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gtime"

	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/kvcache"
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

// IsClusterEnabled reports the configured cluster mode for the test service.
func (f *fakeRoleConfigService) IsClusterEnabled(_ context.Context) bool {
	return f.clusterEnabled
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
			{Locale: "zh-TW", NativeName: "繁體中文"},
		},
	}
}

// GetLogin returns an empty login config for tests.
func (f *fakeRoleConfigService) GetLogin(_ context.Context) *hostconfig.LoginConfig {
	return &hostconfig.LoginConfig{}
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
func (f *fakeRoleConfigService) IsLoginIPBlacklisted(_ context.Context, _ string) bool {
	return false
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

// fakeKVCacheService tracks minimal cache interactions needed by access-cache tests.
type fakeKVCacheService struct {
	getIntValue int64
	getIntErr   error
	getIntCalls int32
	incrCalls   int32
}

// Get returns no cached string item because these tests only exercise int revision paths.
func (f *fakeKVCacheService) Get(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (*kvcache.Item, bool, error) {
	return nil, false, nil
}

// GetInt returns the configured revision value and tracks read calls.
func (f *fakeKVCacheService) GetInt(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) (int64, bool, error) {
	atomic.AddInt32(&f.getIntCalls, 1)
	return f.getIntValue, true, f.getIntErr
}

// Set is a no-op success stub for tests that do not inspect string values.
func (f *fakeKVCacheService) Set(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ string,
	_ int64,
) (*kvcache.Item, error) {
	return nil, nil
}

// Delete is a no-op success stub for tests.
func (f *fakeKVCacheService) Delete(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
) error {
	return nil
}

// Incr returns the configured integer value and tracks increment calls.
func (f *fakeKVCacheService) Incr(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
	_ int64,
) (*kvcache.Item, error) {
	atomic.AddInt32(&f.incrCalls, 1)
	return &kvcache.Item{IntValue: f.getIntValue}, nil
}

// Expire is a no-op stub because expiration is not exercised by these tests.
func (f *fakeKVCacheService) Expire(
	_ context.Context,
	_ kvcache.OwnerType,
	_ string,
	_ int64,
) (bool, *gtime.Time, error) {
	return false, nil, nil
}

// CleanupExpired is a no-op success stub for tests.
func (f *fakeKVCacheService) CleanupExpired(_ context.Context) error {
	return nil
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

// setAccessRevisionControllerForTest swaps the service revision controller with
// one built from the supplied test configuration.
func setAccessRevisionControllerForTest(
	svc *serviceImpl,
	clusterEnabled bool,
	kvCacheSvc kvcache.Service,
) {
	svc.accessRevisionCtrl = newAccessRevisionController(clusterEnabled, kvCacheSvc)
}

// TestNewAccessRevisionControllerSelectsByClusterMode verifies controller
// selection switches between local and cluster implementations.
func TestNewAccessRevisionControllerSelectsByClusterMode(t *testing.T) {
	if _, ok := newAccessRevisionController(
		false,
		&fakeKVCacheService{},
	).(*localAccessRevisionController); !ok {
		t.Fatal("expected single-node mode to use local access revision controller")
	}

	if _, ok := newAccessRevisionController(
		true,
		&fakeKVCacheService{},
	).(*clusterAccessRevisionController); !ok {
		t.Fatal("expected cluster mode to use shared access revision controller")
	}
}

// TestTokenAccessContextCacheLifecycle verifies token cache reads, invalidation,
// and revision mismatch eviction.
func TestTokenAccessContextCacheLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	tokenID := "token-cache-lifecycle"
	userID := 101
	access := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "editor"},
		MenuIds:      []int{10, 11},
		Permissions:  []string{"system:user:query", "system:user:edit"},
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
	svc := New(nil).(*serviceImpl)
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

// TestCloneUserAccessContextCopiesSlices verifies cloned access contexts do not
// share slice backing storage with the original value.
func TestCloneUserAccessContextCopiesSlices(t *testing.T) {
	original := &UserAccessContext{
		RoleIds:      []int{1, 2},
		RoleNames:    []string{"admin", "ops"},
		MenuIds:      []int{10, 20},
		Permissions:  []string{"user:list", "user:update"},
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
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeKV := &fakeKVCacheService{getIntValue: 9}
	svc.kvCacheSvc = fakeKV
	setAccessRevisionControllerForTest(svc, true, fakeKV)

	revision, err := svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 1 {
		t.Fatalf("expected exactly one GetInt call, got %d", atomic.LoadInt32(&fakeKV.getIntCalls))
	}
	if atomic.LoadInt32(&fakeKV.incrCalls) != 0 {
		t.Fatalf("expected no Incr calls for read path, got %d", atomic.LoadInt32(&fakeKV.incrCalls))
	}

	revision, err = svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("second get access revision failed: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected cached revision 9, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 1 {
		t.Fatalf("expected cached local revision to avoid extra GetInt calls, got %d", atomic.LoadInt32(&fakeKV.getIntCalls))
	}
}

// TestSyncAccessTopologyRevisionKeepsCacheWhenRevisionUnchanged verifies cache
// contents survive synchronization when the shared revision is unchanged.
func TestSyncAccessTopologyRevisionKeepsCacheWhenRevisionUnchanged(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeKV := &fakeKVCacheService{getIntValue: 7}
	svc.kvCacheSvc = fakeKV
	setAccessRevisionControllerForTest(svc, true, fakeKV)
	storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-same-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey("sync-same-revision"))
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
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	fakeKV := &fakeKVCacheService{getIntValue: 8}
	svc.kvCacheSvc = fakeKV
	setAccessRevisionControllerForTest(svc, true, fakeKV)
	storeLocalAccessRevision(7)
	svc.cacheTokenAccessContext(ctx, "sync-new-revision", 1, 7, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err := svc.SyncAccessTopologyRevision(ctx); err != nil {
		t.Fatalf("sync access topology revision failed: %v", err)
	}

	cachedVar, err := accessContextCache.Get(ctx, accessCacheKey("sync-new-revision"))
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

// TestSingleNodeAccessRevisionStaysLocal verifies single-node mode maintains
// revisions locally without calling the shared KV cache.
func TestSingleNodeAccessRevisionStaysLocal(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	fakeKV := &fakeKVCacheService{getIntValue: 9}
	svc.kvCacheSvc = fakeKV
	setAccessRevisionControllerForTest(svc, false, fakeKV)

	revision, err := svc.getAccessRevision(ctx)
	if err != nil {
		t.Fatalf("get single-node access revision failed: %v", err)
	}
	if revision != 1 {
		t.Fatalf("expected initial single-node revision 1, got %d", revision)
	}
	if atomic.LoadInt32(&fakeKV.getIntCalls) != 0 {
		t.Fatalf("expected single-node getAccessRevision to avoid GetInt, got %d calls", atomic.LoadInt32(&fakeKV.getIntCalls))
	}

	svc.cacheTokenAccessContext(ctx, "single-node-token", 1, revision, &UserAccessContext{
		Permissions: []string{"system:user:list"},
	})

	if err = svc.MarkAccessTopologyChanged(ctx); err != nil {
		t.Fatalf("mark single-node access topology changed failed: %v", err)
	}
	if atomic.LoadInt32(&fakeKV.incrCalls) != 0 {
		t.Fatalf("expected single-node topology change to avoid Incr, got %d calls", atomic.LoadInt32(&fakeKV.incrCalls))
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

// TestLoadTokenAccessContextWithCacheLockSuppressesDuplicateLoads verifies one
// cold load is shared across concurrent cache misses.
func TestLoadTokenAccessContextWithCacheLockSuppressesDuplicateLoads(t *testing.T) {
	ctx := context.Background()
	svc := New(nil).(*serviceImpl)
	resetRoleAccessCacheTestState(t, svc)

	svc.configSvc = &fakeRoleConfigService{clusterEnabled: true}
	svc.kvCacheSvc = &fakeKVCacheService{getIntValue: 3}
	setAccessRevisionControllerForTest(svc, true, svc.kvCacheSvc)

	var loadCalls atomic.Int32
	loader := func(context.Context) (*UserAccessContext, error) {
		loadCalls.Add(1)
		time.Sleep(30 * time.Millisecond)
		return &UserAccessContext{
			RoleIds:      []int{1},
			RoleNames:    []string{"admin"},
			MenuIds:      []int{101},
			Permissions:  []string{"system:user:list"},
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
	}
	if count != workers {
		t.Fatalf("expected %d access results, got %d", workers, count)
	}
}
