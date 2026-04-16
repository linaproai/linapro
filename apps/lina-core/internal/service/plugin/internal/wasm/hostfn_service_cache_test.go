// This file tests cache host service dispatch, authorization, and size limits.

package wasm

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/pluginbridge"
)

const createPluginKVCacheTableSQL = `
CREATE TABLE IF NOT EXISTS sys_kv_cache (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    owner_type VARCHAR(16) NOT NULL DEFAULT '' COMMENT '所属类型：plugin=动态插件 module=宿主模块',
    owner_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属标识：插件ID或模块名',
    namespace VARCHAR(64) NOT NULL DEFAULT '' COMMENT '缓存命名空间，对应 host-cache 资源标识',
    cache_key VARCHAR(128) NOT NULL DEFAULT '' COMMENT '缓存键',
    value_kind TINYINT NOT NULL DEFAULT 1 COMMENT '值类型：1=字符串 2=整数',
    value_bytes VARBINARY(4096) NOT NULL COMMENT '缓存字节值，供 get/set 使用',
    value_int BIGINT NOT NULL DEFAULT 0 COMMENT '缓存整数值，供 incr 使用',
    expire_at DATETIME NULL DEFAULT NULL COMMENT '过期时间，NULL表示永不过期',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_owner_namespace_key (owner_type, owner_key, namespace, cache_key),
    KEY idx_expire_at (expire_at)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='宿主分布式KV缓存表';
`

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

func ensurePluginKVCacheTable(t *testing.T, ctx context.Context) {
	t.Helper()
	if _, err := g.DB().Exec(ctx, createPluginKVCacheTableSQL); err != nil {
		t.Fatalf("expected sys_kv_cache table to be created, got error: %v", err)
	}
}

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
