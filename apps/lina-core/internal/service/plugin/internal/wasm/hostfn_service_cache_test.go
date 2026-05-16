// This file tests cache host service dispatch, authorization, and size limits.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

// trackingCacheService records cache method calls while returning deterministic
// responses for shared-instance wiring tests.
type trackingCacheService struct {
	getCalls    int
	setCalls    int
	deleteCalls int
	incrCalls   int
	expireCalls int
	lastKey     string
	value       string
}

// BackendName returns a deterministic backend name for assertions.
func (s *trackingCacheService) BackendName() kvcache.BackendName {
	return kvcache.BackendName("tracking")
}

// RequiresExpiredCleanup reports no cleanup requirement for the fake backend.
func (s *trackingCacheService) RequiresExpiredCleanup() bool { return false }

// Get records one cache read.
func (s *trackingCacheService) Get(_ context.Context, _ kvcache.OwnerType, cacheKey string) (*kvcache.Item, bool, error) {
	s.getCalls++
	s.lastKey = cacheKey
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindString, Value: s.value}, true, nil
}

// GetInt records no dedicated behavior because host service dispatch uses Get.
func (s *trackingCacheService) GetInt(context.Context, kvcache.OwnerType, string) (int64, bool, error) {
	return 0, false, nil
}

// Set records one cache write.
func (s *trackingCacheService) Set(_ context.Context, _ kvcache.OwnerType, cacheKey string, value string, _ time.Duration) (*kvcache.Item, error) {
	s.setCalls++
	s.lastKey = cacheKey
	s.value = value
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindString, Value: value}, nil
}

// Delete records one cache delete.
func (s *trackingCacheService) Delete(_ context.Context, _ kvcache.OwnerType, cacheKey string) error {
	s.deleteCalls++
	s.lastKey = cacheKey
	return nil
}

// Incr records one cache increment.
func (s *trackingCacheService) Incr(_ context.Context, _ kvcache.OwnerType, cacheKey string, delta int64, _ time.Duration) (*kvcache.Item, error) {
	s.incrCalls++
	s.lastKey = cacheKey
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindInt, IntValue: delta}, nil
}

// Expire records one cache expiration update.
func (s *trackingCacheService) Expire(_ context.Context, _ kvcache.OwnerType, cacheKey string, _ time.Duration) (bool, *gtime.Time, error) {
	s.expireCalls++
	s.lastKey = cacheKey
	return true, nil, nil
}

// CleanupExpired records no behavior for the fake backend.
func (s *trackingCacheService) CleanupExpired(context.Context) error { return nil }

// createPluginKVCacheTableSQL prepares the governed plugin cache table for tests.
const createPluginKVCacheTableSQL = `
CREATE TABLE IF NOT EXISTS sys_kv_cache (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_type  VARCHAR(16) NOT NULL DEFAULT '',
    owner_key   VARCHAR(64) NOT NULL DEFAULT '',
    namespace   VARCHAR(64) NOT NULL DEFAULT '',
    cache_key   VARCHAR(128) NOT NULL DEFAULT '',
    value_kind  SMALLINT NOT NULL DEFAULT 1,
    value_bytes BYTEA NOT NULL,
    value_int   BIGINT NOT NULL DEFAULT 0,
    expire_at   TIMESTAMP NULL DEFAULT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_kv_cache_owner_namespace_key ON sys_kv_cache (owner_type, owner_key, namespace, cache_key);
CREATE INDEX IF NOT EXISTS idx_sys_kv_cache_expire_at ON sys_kv_cache (expire_at);
`

// TestHandleHostServiceInvokeCacheLifecycle verifies cache get/set/incr/expire/delete flows.
func TestHandleHostServiceInvokeCacheLifecycle(t *testing.T) {
	ctx := context.Background()
	ensurePluginKVCacheTable(t, ctx)

	pluginID := "test-plugin-cache"
	namespace := "orders-cache"
	cleanupPluginCacheNamespace(t, ctx, pluginID, namespace)
	t.Cleanup(func() {
		cleanupPluginCacheNamespace(t, ctx, pluginID, namespace)
	})

	hcc := newCacheHostCallContext(pluginID, namespace)

	setResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheSet,
		namespace,
		pluginbridge.MarshalHostServiceCacheSetRequest(&pluginbridge.HostServiceCacheSetRequest{
			Key:           "profile",
			Value:         `{"enabled":true}`,
			ExpireSeconds: 60,
		}),
	)
	if setResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("set: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	setPayload, err := pluginbridge.UnmarshalHostServiceCacheSetResponse(setResponse.Payload)
	if err != nil {
		t.Fatalf("set payload decode failed: %v", err)
	}
	if setPayload.Value == nil || setPayload.Value.Value != `{"enabled":true}` {
		t.Fatalf("set payload: got %#v", setPayload.Value)
	}

	getResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheGet,
		namespace,
		pluginbridge.MarshalHostServiceCacheGetRequest(&pluginbridge.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("get: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := pluginbridge.UnmarshalHostServiceCacheGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("get payload decode failed: %v", err)
	}
	if !getPayload.Found || getPayload.Value == nil || getPayload.Value.Value != `{"enabled":true}` {
		t.Fatalf("get payload: got %#v", getPayload)
	}

	incrResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheIncr,
		namespace,
		pluginbridge.MarshalHostServiceCacheIncrRequest(&pluginbridge.HostServiceCacheIncrRequest{
			Key:           "counter",
			Delta:         2,
			ExpireSeconds: 60,
		}),
	)
	if incrResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("incr: expected success, got status=%d payload=%s", incrResponse.Status, string(incrResponse.Payload))
	}
	incrPayload, err := pluginbridge.UnmarshalHostServiceCacheIncrResponse(incrResponse.Payload)
	if err != nil {
		t.Fatalf("incr payload decode failed: %v", err)
	}
	if incrPayload.Value == nil || incrPayload.Value.IntValue != 2 || incrPayload.Value.ValueKind != pluginbridge.HostServiceCacheValueKindInt {
		t.Fatalf("incr payload: got %#v", incrPayload.Value)
	}

	expireResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheExpire,
		namespace,
		pluginbridge.MarshalHostServiceCacheExpireRequest(&pluginbridge.HostServiceCacheExpireRequest{
			Key:           "profile",
			ExpireSeconds: 120,
		}),
	)
	if expireResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expire: expected success, got status=%d payload=%s", expireResponse.Status, string(expireResponse.Payload))
	}
	expirePayload, err := pluginbridge.UnmarshalHostServiceCacheExpireResponse(expireResponse.Payload)
	if err != nil {
		t.Fatalf("expire payload decode failed: %v", err)
	}
	if !expirePayload.Found || strings.TrimSpace(expirePayload.ExpireAt) == "" {
		t.Fatalf("expire payload: got %#v", expirePayload)
	}

	deleteResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheDelete,
		namespace,
		pluginbridge.MarshalHostServiceCacheDeleteRequest(&pluginbridge.HostServiceCacheDeleteRequest{Key: "profile"}),
	)
	if deleteResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("delete: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}

	getDeletedResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheGet,
		namespace,
		pluginbridge.MarshalHostServiceCacheGetRequest(&pluginbridge.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getDeletedResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("get after delete: expected success, got status=%d payload=%s", getDeletedResponse.Status, string(getDeletedResponse.Payload))
	}
	getDeletedPayload, err := pluginbridge.UnmarshalHostServiceCacheGetResponse(getDeletedResponse.Payload)
	if err != nil {
		t.Fatalf("get after delete payload decode failed: %v", err)
	}
	if getDeletedPayload.Found {
		t.Fatalf("expected deleted cache entry to be missing, got %#v", getDeletedPayload)
	}
}

// TestHandleHostServiceInvokeCacheRejectsOversizedValue verifies platform cache limits are enforced.
func TestHandleHostServiceInvokeCacheRejectsOversizedValue(t *testing.T) {
	ctx := context.Background()
	ensurePluginKVCacheTable(t, ctx)

	hcc := newCacheHostCallContext("test-plugin-cache-limit", "orders-cache")
	response := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheSet,
		"orders-cache",
		pluginbridge.MarshalHostServiceCacheSetRequest(&pluginbridge.HostServiceCacheSetRequest{
			Key:   "oversized",
			Value: strings.Repeat("a", 4097),
		}),
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for oversized cache value, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeCacheRejectsUnauthorizedNamespace verifies unauthorized namespaces are rejected.
func TestHandleHostServiceInvokeCacheRejectsUnauthorizedNamespace(t *testing.T) {
	hcc := newCacheHostCallContext("test-plugin-cache-denied", "orders-cache")
	response := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheGet,
		"other-cache",
		pluginbridge.MarshalHostServiceCacheGetRequest(&pluginbridge.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized cache namespace, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeCacheUsesConfiguredSharedService verifies cache
// host service dispatch reuses the explicitly configured shared instance.
func TestHandleHostServiceInvokeCacheUsesConfiguredSharedService(t *testing.T) {
	cacheSvc := &trackingCacheService{}
	previousCacheSvc := cacheHostService
	if err := ConfigureCacheHostService(cacheSvc); err != nil {
		t.Fatalf("configure cache host service failed: %v", err)
	}
	t.Cleanup(func() {
		cacheHostService = previousCacheSvc
	})

	hcc := newTenantCacheHostCallContext("test-plugin-cache-shared", "orders-cache", 77)
	setResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheSet,
		"orders-cache",
		pluginbridge.MarshalHostServiceCacheSetRequest(&pluginbridge.HostServiceCacheSetRequest{
			Key:   "profile",
			Value: "shared",
		}),
	)
	if setResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("set through shared cache: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	getResponse := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheGet,
		"orders-cache",
		pluginbridge.MarshalHostServiceCacheGetRequest(&pluginbridge.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("get through shared cache: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	if cacheSvc.setCalls != 1 || cacheSvc.getCalls != 1 {
		t.Fatalf("expected shared cache to receive one set and one get, got set=%d get=%d", cacheSvc.setCalls, cacheSvc.getCalls)
	}
	expectedKey := kvcache.BuildTenantCacheKey(77, "plugin-cache", hcc.pluginID, "orders-cache", "profile")
	if cacheSvc.lastKey != expectedKey {
		t.Fatalf("expected tenant-scoped cache key %q, got %q", expectedKey, cacheSvc.lastKey)
	}
}

// TestHandleHostServiceInvokeCacheUsesCoordinationKVAndTenantIsolation verifies
// the host cache service can run on coordination KV and keeps tenant keys apart.
func TestHandleHostServiceInvokeCacheUsesCoordinationKVAndTenantIsolation(t *testing.T) {
	cacheSvc := kvcache.New(kvcache.WithProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil))))
	previousCacheSvc := cacheHostService
	if err := ConfigureCacheHostService(cacheSvc); err != nil {
		t.Fatalf("configure cache host service failed: %v", err)
	}
	t.Cleanup(func() {
		cacheHostService = previousCacheSvc
	})

	if cacheSvc.BackendName() != kvcache.BackendCoordinationKV {
		t.Fatalf("expected coordination KV backend, got %q", cacheSvc.BackendName())
	}

	pluginID := "test-plugin-cache-tenant"
	namespace := "orders-cache"
	tenantOne := newTenantCacheHostCallContext(pluginID, namespace, 11)
	tenantTwo := newTenantCacheHostCallContext(pluginID, namespace, 22)

	setTenantCacheValue(t, tenantOne, namespace, "profile", "tenant-one")
	setTenantCacheValue(t, tenantTwo, namespace, "profile", "tenant-two")

	assertTenantCacheValue(t, tenantOne, namespace, "profile", "tenant-one")
	assertTenantCacheValue(t, tenantTwo, namespace, "profile", "tenant-two")
}

// TestConfigureCacheHostServiceRejectsNil verifies missing runtime cache
// injection returns an error instead of silently constructing an isolated backend.
func TestConfigureCacheHostServiceRejectsNil(t *testing.T) {
	if err := ConfigureCacheHostService(nil); err == nil {
		t.Fatal("expected nil cache host service to return an error")
	}
}

// ensurePluginKVCacheTable creates the plugin cache table needed by cache host call tests.
func ensurePluginKVCacheTable(t *testing.T, ctx context.Context) {
	t.Helper()
	for _, statement := range dialect.SplitSQLStatements(createPluginKVCacheTableSQL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("expected sys_kv_cache table to be created, got error: %v\nSQL:\n%s", err, statement)
		}
	}
}

// cleanupPluginCacheNamespace removes cache rows for the plugin namespace used in tests.
func cleanupPluginCacheNamespace(t *testing.T, ctx context.Context, pluginID string, namespace string) {
	t.Helper()
	if _, err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: kvcache.OwnerTypePlugin.String(),
		OwnerKey:  pluginID,
		Namespace: namespace,
	}).Delete(); err != nil {
		t.Fatalf("failed to cleanup plugin cache namespace %s/%s: %v", pluginID, namespace, err)
	}
}

// newCacheHostCallContext builds a host call context authorized for one cache namespace.
func newCacheHostCallContext(pluginID string, namespace string) *hostCallContext {
	return &hostCallContext{
		pluginID: pluginID,
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityCache: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceCache,
			Methods: []string{
				pluginbridge.HostServiceMethodCacheDelete,
				pluginbridge.HostServiceMethodCacheExpire,
				pluginbridge.HostServiceMethodCacheGet,
				pluginbridge.HostServiceMethodCacheIncr,
				pluginbridge.HostServiceMethodCacheSet,
			},
			Resources: []*pluginbridge.HostServiceResourceSpec{
				{Ref: namespace},
			},
		}},
	}
}

// newTenantCacheHostCallContext builds a cache host call context with a tenant identity.
func newTenantCacheHostCallContext(pluginID string, namespace string, tenantID int32) *hostCallContext {
	hcc := newCacheHostCallContext(pluginID, namespace)
	hcc.identity = &pluginbridge.IdentitySnapshotV1{TenantId: tenantID, UserID: 1, Username: "admin"}
	return hcc
}

// setTenantCacheValue writes one cache value through the host service dispatcher.
func setTenantCacheValue(t *testing.T, hcc *hostCallContext, namespace string, key string, value string) {
	t.Helper()
	response := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheSet,
		namespace,
		pluginbridge.MarshalHostServiceCacheSetRequest(&pluginbridge.HostServiceCacheSetRequest{
			Key:   key,
			Value: value,
		}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("set cache value: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// assertTenantCacheValue verifies one cache value through the host service dispatcher.
func assertTenantCacheValue(t *testing.T, hcc *hostCallContext, namespace string, key string, expected string) {
	t.Helper()
	response := invokeCacheHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodCacheGet,
		namespace,
		pluginbridge.MarshalHostServiceCacheGetRequest(&pluginbridge.HostServiceCacheGetRequest{Key: key}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("get cache value: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := pluginbridge.UnmarshalHostServiceCacheGetResponse(response.Payload)
	if err != nil {
		t.Fatalf("decode cache get payload failed: %v", err)
	}
	if !payload.Found || payload.Value == nil || payload.Value.Value != expected {
		t.Fatalf("expected cache value %q, got %#v", expected, payload)
	}
}

// invokeCacheHostService marshals and dispatches one cache host service request.
func invokeCacheHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	namespace string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceCache,
		Method:      method,
		ResourceRef: namespace,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		hcc,
		pluginbridge.MarshalHostServiceRequestEnvelope(request),
	)
}
