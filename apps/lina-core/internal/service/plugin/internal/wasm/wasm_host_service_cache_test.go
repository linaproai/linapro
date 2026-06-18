// This file tests cache host service dispatch, authorization, and size limits.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
func (s *trackingCacheService) Expire(_ context.Context, _ kvcache.OwnerType, cacheKey string, _ time.Duration) (bool, *time.Time, error) {
	s.expireCalls++
	s.lastKey = cacheKey
	return true, nil, nil
}

// CleanupExpired records no behavior for the fake backend.
func (s *trackingCacheService) CleanupExpired(context.Context) error { return nil }

// staticCacheService is stateless so concurrent wiring tests can exercise the
// host-service reference without introducing fake-service data races.
type staticCacheService struct{}

// BackendName returns a deterministic backend name for assertions.
func (staticCacheService) BackendName() kvcache.BackendName {
	return kvcache.BackendName("static")
}

// RequiresExpiredCleanup reports no cleanup requirement for the fake backend.
func (staticCacheService) RequiresExpiredCleanup() bool { return false }

// Get returns one deterministic cache item.
func (staticCacheService) Get(_ context.Context, _ kvcache.OwnerType, cacheKey string) (*kvcache.Item, bool, error) {
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindString, Value: "race-safe"}, true, nil
}

// GetInt returns no dedicated integer value because host service dispatch uses Get.
func (staticCacheService) GetInt(context.Context, kvcache.OwnerType, string) (int64, bool, error) {
	return 0, false, nil
}

// Set returns one deterministic string cache item.
func (staticCacheService) Set(_ context.Context, _ kvcache.OwnerType, cacheKey string, value string, _ time.Duration) (*kvcache.Item, error) {
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindString, Value: value}, nil
}

// Delete accepts one cache delete.
func (staticCacheService) Delete(context.Context, kvcache.OwnerType, string) error {
	return nil
}

// Incr returns one deterministic integer cache item.
func (staticCacheService) Incr(_ context.Context, _ kvcache.OwnerType, cacheKey string, delta int64, _ time.Duration) (*kvcache.Item, error) {
	return &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindInt, IntValue: delta}, nil
}

// Expire accepts one expiration update.
func (staticCacheService) Expire(context.Context, kvcache.OwnerType, string, time.Duration) (bool, *time.Time, error) {
	return true, nil, nil
}

// CleanupExpired records no behavior for the fake backend.
func (staticCacheService) CleanupExpired(context.Context) error { return nil }

// cacheDomainTestService adapts a shared kvcache.Service to cachecap.Service
// for WASM dispatcher tests without expanding capabilityhost's production API.
type cacheDomainTestService struct {
	pluginID string
	service  kvcache.Service
}

// Get returns one cache item from the plugin-scoped cache key.
func (s *cacheDomainTestService) Get(ctx context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	cacheKey := s.cacheKey(ctx, namespace, key)
	item, found, err := s.service.Get(ctx, kvcache.OwnerTypePlugin, cacheKey)
	return cacheItemFromKV(item, key), found, err
}

// GetMany returns explicit cache items from the plugin-scoped cache key set.
func (s *cacheDomainTestService) GetMany(ctx context.Context, in cachecap.GetManyInput) (*cachecap.GetManyOutput, error) {
	output := &cachecap.GetManyOutput{
		Items:       make(map[string]*cachecap.CacheItem, len(in.Keys)),
		MissingKeys: []string{},
	}
	for _, key := range in.Keys {
		item, found, err := s.Get(ctx, in.Namespace, key)
		if err != nil {
			return nil, err
		}
		if !found {
			output.MissingKeys = append(output.MissingKeys, key)
			continue
		}
		output.Items[key] = item
	}
	return output, nil
}

// Set writes one cache item to the plugin-scoped cache key.
func (s *cacheDomainTestService) Set(ctx context.Context, namespace string, key string, value string, ttl time.Duration) (*cachecap.CacheItem, error) {
	cacheKey := s.cacheKey(ctx, namespace, key)
	item, err := s.service.Set(ctx, kvcache.OwnerTypePlugin, cacheKey, value, ttl)
	return cacheItemFromKV(item, key), err
}

// SetMany writes explicit cache items to the plugin-scoped cache key set.
func (s *cacheDomainTestService) SetMany(ctx context.Context, in cachecap.SetManyInput) (*cachecap.SetManyOutput, error) {
	output := &cachecap.SetManyOutput{Items: make(map[string]*cachecap.CacheItem, len(in.Items))}
	for _, item := range in.Items {
		written, err := s.Set(ctx, in.Namespace, item.Key, item.Value, item.TTL)
		if err != nil {
			return nil, err
		}
		output.Items[item.Key] = written
	}
	return output, nil
}

// Delete removes one cache item from the plugin-scoped cache key.
func (s *cacheDomainTestService) Delete(ctx context.Context, namespace string, key string) error {
	return s.service.Delete(ctx, kvcache.OwnerTypePlugin, s.cacheKey(ctx, namespace, key))
}

// DeleteMany removes explicit cache items from the plugin-scoped cache key set.
func (s *cacheDomainTestService) DeleteMany(ctx context.Context, in cachecap.DeleteManyInput) error {
	for _, key := range in.Keys {
		if err := s.Delete(ctx, in.Namespace, key); err != nil {
			return err
		}
	}
	return nil
}

// Incr increments one integer cache item at the plugin-scoped cache key.
func (s *cacheDomainTestService) Incr(ctx context.Context, namespace string, key string, delta int64, ttl time.Duration) (*cachecap.CacheItem, error) {
	cacheKey := s.cacheKey(ctx, namespace, key)
	item, err := s.service.Incr(ctx, kvcache.OwnerTypePlugin, cacheKey, delta, ttl)
	return cacheItemFromKV(item, key), err
}

// Expire updates one cache item's expiration policy.
func (s *cacheDomainTestService) Expire(ctx context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error) {
	return s.service.Expire(ctx, kvcache.OwnerTypePlugin, s.cacheKey(ctx, namespace, key), ttl)
}

// cacheKey builds the same plugin/tenant logical key shape as the host adapter.
func (s *cacheDomainTestService) cacheKey(ctx context.Context, namespace string, key string) string {
	if current := bizctxcap.CurrentFromContext(ctx); current.TenantID > 0 {
		return kvcache.BuildTenantCacheKey(
			tenantcap.TenantID(current.TenantID),
			"plugin-cache",
			s.pluginID,
			namespace,
			key,
		)
	}
	return kvcache.BuildCacheKey(s.pluginID, namespace, key)
}

// cacheItemFromKV maps one internal cache item into the domain cache DTO.
func cacheItemFromKV(item *kvcache.Item, logicalKey string) *cachecap.CacheItem {
	if item == nil {
		return nil
	}
	return &cachecap.CacheItem{
		Key:       logicalKey,
		ValueKind: item.ValueKind,
		Value:     item.Value,
		IntValue:  item.IntValue,
		ExpireAt:  item.ExpireAt,
	}
}

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
	configureCacheDomainServiceForTest(t, kvcache.New())

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
		protocol.HostServiceMethodCacheSet,
		namespace,
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:           "profile",
			Value:         `{"enabled":true}`,
			ExpireSeconds: 60,
		}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("set: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	setPayload, err := protocol.UnmarshalHostServiceCacheSetResponse(setResponse.Payload)
	if err != nil {
		t.Fatalf("set payload decode failed: %v", err)
	}
	if setPayload.Value == nil || setPayload.Value.Value != `{"enabled":true}` {
		t.Fatalf("set payload: got %#v", setPayload.Value)
	}
	if _, err := time.Parse(time.RFC3339Nano, setPayload.Value.ExpireAt); err != nil {
		t.Fatalf("set expireAt should use RFC3339Nano wire format, got %q: %v", setPayload.Value.ExpireAt, err)
	}

	getResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		namespace,
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := protocol.UnmarshalHostServiceCacheGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("get payload decode failed: %v", err)
	}
	if !getPayload.Found || getPayload.Value == nil || getPayload.Value.Value != `{"enabled":true}` {
		t.Fatalf("get payload: got %#v", getPayload)
	}

	incrResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheIncr,
		namespace,
		protocol.MarshalHostServiceCacheIncrRequest(&protocol.HostServiceCacheIncrRequest{
			Key:           "counter",
			Delta:         2,
			ExpireSeconds: 60,
		}),
	)
	if incrResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("incr: expected success, got status=%d payload=%s", incrResponse.Status, string(incrResponse.Payload))
	}
	incrPayload, err := protocol.UnmarshalHostServiceCacheIncrResponse(incrResponse.Payload)
	if err != nil {
		t.Fatalf("incr payload decode failed: %v", err)
	}
	if incrPayload.Value == nil || incrPayload.Value.IntValue != 2 || incrPayload.Value.ValueKind != protocol.HostServiceCacheValueKindInt {
		t.Fatalf("incr payload: got %#v", incrPayload.Value)
	}
	if _, err := time.Parse(time.RFC3339Nano, incrPayload.Value.ExpireAt); err != nil {
		t.Fatalf("incr expireAt should use RFC3339Nano wire format, got %q: %v", incrPayload.Value.ExpireAt, err)
	}

	expireResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheExpire,
		namespace,
		protocol.MarshalHostServiceCacheExpireRequest(&protocol.HostServiceCacheExpireRequest{
			Key:           "profile",
			ExpireSeconds: 120,
		}),
	)
	if expireResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expire: expected success, got status=%d payload=%s", expireResponse.Status, string(expireResponse.Payload))
	}
	expirePayload, err := protocol.UnmarshalHostServiceCacheExpireResponse(expireResponse.Payload)
	if err != nil {
		t.Fatalf("expire payload decode failed: %v", err)
	}
	if !expirePayload.Found || strings.TrimSpace(expirePayload.ExpireAt) == "" {
		t.Fatalf("expire payload: got %#v", expirePayload)
	}
	if _, err := time.Parse(time.RFC3339Nano, expirePayload.ExpireAt); err != nil {
		t.Fatalf("expire expireAt should use RFC3339Nano wire format, got %q: %v", expirePayload.ExpireAt, err)
	}

	deleteResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheDelete,
		namespace,
		protocol.MarshalHostServiceCacheDeleteRequest(&protocol.HostServiceCacheDeleteRequest{Key: "profile"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("delete: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}

	getDeletedResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		namespace,
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getDeletedResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get after delete: expected success, got status=%d payload=%s", getDeletedResponse.Status, string(getDeletedResponse.Payload))
	}
	getDeletedPayload, err := protocol.UnmarshalHostServiceCacheGetResponse(getDeletedResponse.Payload)
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
	configureCacheDomainServiceForTest(t, kvcache.New())

	hcc := newCacheHostCallContext("test-plugin-cache-limit", "orders-cache")
	response := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheSet,
		"orders-cache",
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:   "oversized",
			Value: strings.Repeat("a", 4097),
		}),
	)
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for oversized cache value, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeCacheBatchMethods verifies cache multi-key
// operations use the shared domain service and JSON envelopes.
func TestHandleHostServiceInvokeCacheBatchMethods(t *testing.T) {
	ctx := context.Background()
	ensurePluginKVCacheTable(t, ctx)
	configureCacheDomainServiceForTest(t, kvcache.New())

	pluginID := "test-plugin-cache-batch"
	namespace := "orders-cache-batch"
	cleanupPluginCacheNamespace(t, ctx, pluginID, namespace)
	t.Cleanup(func() {
		cleanupPluginCacheNamespace(t, ctx, pluginID, namespace)
	})
	hcc := newCacheHostCallContext(pluginID, namespace)

	setResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheSetMany,
		namespace,
		protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: []byte(`{"items":[{"key":"profile","value":"enabled","expireSeconds":60},{"key":"theme","value":"dark"}]}`)}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("set_many: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}

	getResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGetMany,
		namespace,
		protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: []byte(`{"keys":["profile","missing"]}`)}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get_many: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceCapabilityJSONResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("decode get_many JSON envelope: %v", err)
	}
	if !strings.Contains(string(payload.Value), `"profile"`) ||
		!strings.Contains(string(payload.Value), `"missingKeys":["missing"]`) {
		t.Fatalf("unexpected get_many JSON response: %s", string(payload.Value))
	}

	deleteResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheDeleteMany,
		namespace,
		protocol.MarshalHostServiceCapabilityJSONRequest(&protocol.HostServiceCapabilityJSONRequest{Value: []byte(`{"keys":["profile","theme"]}`)}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("delete_many: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
}

// TestHandleHostServiceInvokeCacheRejectsUnauthorizedNamespace verifies unauthorized namespaces are rejected.
func TestHandleHostServiceInvokeCacheRejectsUnauthorizedNamespace(t *testing.T) {
	hcc := newCacheHostCallContext("test-plugin-cache-denied", "orders-cache")
	response := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		"other-cache",
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized cache namespace, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeCacheRequiresScopedDomainService verifies missing
// cache domain wiring fails explicitly instead of using a package default.
func TestHandleHostServiceInvokeCacheRequiresScopedDomainService(t *testing.T) {
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{})
	hcc := newCacheHostCallContext("test-plugin-cache-unconfigured", "orders-cache")
	response := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		"orders-cache",
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if response.Status != protocol.HostCallStatusInternalError {
		t.Fatalf("expected internal error for unconfigured cache service, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeCacheUsesConfiguredSharedService verifies cache
// host service dispatch reuses the explicitly configured shared backend.
func TestHandleHostServiceInvokeCacheUsesConfiguredSharedService(t *testing.T) {
	cacheSvc := &trackingCacheService{}
	configureCacheDomainServiceForTest(t, cacheSvc)

	hcc := newTenantCacheHostCallContext("test-plugin-cache-shared", "orders-cache", 77)
	setResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheSet,
		"orders-cache",
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:   "profile",
			Value: "shared",
		}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("set through shared cache: expected success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}
	getResponse := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		"orders-cache",
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: "profile"}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
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
	configureCacheDomainServiceForTest(t, cacheSvc)

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

// configureCacheDomainServiceForTest installs one explicit shared cache
// backend through the unified capability directory.
func configureCacheDomainServiceForTest(t *testing.T, service kvcache.Service) {
	t.Helper()
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		cache: cacheDomainForKV(service),
	})
}

// cacheDomainForKV adapts the shared KV cache backend to the cachecap domain
// contract while preserving the source/dynamic plugin scoping path.
func cacheDomainForKV(service kvcache.Service) cachecap.Service {
	return &cacheDomainTestService{service: service}
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
			protocol.CapabilityCache: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceCache,
			Methods: []string{
				protocol.HostServiceMethodCacheDelete,
				protocol.HostServiceMethodCacheDeleteMany,
				protocol.HostServiceMethodCacheExpire,
				protocol.HostServiceMethodCacheGet,
				protocol.HostServiceMethodCacheGetMany,
				protocol.HostServiceMethodCacheIncr,
				protocol.HostServiceMethodCacheSet,
				protocol.HostServiceMethodCacheSetMany,
			},
			Resources: []*protocol.HostServiceResourceSpec{
				{Ref: namespace},
			},
		}},
	}
}

// newTenantCacheHostCallContext builds a cache host call context with a tenant identity.
func newTenantCacheHostCallContext(pluginID string, namespace string, tenantID int32) *hostCallContext {
	hcc := newCacheHostCallContext(pluginID, namespace)
	hcc.identity = &protocol.IdentitySnapshotV1{TenantId: tenantID, UserID: 1, Username: "admin"}
	return hcc
}

// setTenantCacheValue writes one cache value through the host service dispatcher.
func setTenantCacheValue(t *testing.T, hcc *hostCallContext, namespace string, key string, value string) {
	t.Helper()
	response := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheSet,
		namespace,
		protocol.MarshalHostServiceCacheSetRequest(&protocol.HostServiceCacheSetRequest{
			Key:   key,
			Value: value,
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("set cache value: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// assertTenantCacheValue verifies one cache value through the host service dispatcher.
func assertTenantCacheValue(t *testing.T, hcc *hostCallContext, namespace string, key string, expected string) {
	t.Helper()
	response := invokeCacheHostService(
		t,
		hcc,
		protocol.HostServiceMethodCacheGet,
		namespace,
		protocol.MarshalHostServiceCacheGetRequest(&protocol.HostServiceCacheGetRequest{Key: key}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get cache value: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceCacheGetResponse(response.Payload)
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
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	return dispatchCacheHostServiceRequest(t, hcc, method, namespace, payload)
}

// dispatchCacheHostServiceRequest dispatches one cache host service request.
func dispatchCacheHostServiceRequest(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	namespace string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceCache,
		Method:      method,
		ResourceRef: namespace,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(request),
	)
}
