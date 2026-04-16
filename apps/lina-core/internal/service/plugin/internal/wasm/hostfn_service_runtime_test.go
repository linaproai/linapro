// This file tests runtime host service methods including log/state/info
// dispatch and capability validation.
package wasm

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/pluginbridge"
)

const createPluginStateTableSQL = `
CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id           INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    state_key    VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '状态键',
    state_value  LONGTEXT                          COMMENT '状态值（支持JSON）',
    created_at   DATETIME                          COMMENT '创建时间',
    updated_at   DATETIME                          COMMENT '更新时间',
    UNIQUE KEY uk_plugin_state (plugin_id, state_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件键值状态存储表';
`

func TestHandleHostServiceInvokeRuntimeStateLifecycle(t *testing.T) {
	ctx := context.Background()
	if _, err := g.DB().Exec(ctx, createPluginStateTableSQL); err != nil {
		t.Fatalf("expected plugin state table to be created, got error: %v", err)
	}

	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-state",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{
					pluginbridge.HostServiceMethodRuntimeStateGet,
					pluginbridge.HostServiceMethodRuntimeStateSet,
					pluginbridge.HostServiceMethodRuntimeStateDelete,
				},
			},
		},
	}
	cleanupRuntimeStateKey(t, ctx, hcc.pluginID, "demo")
	t.Cleanup(func() {
		cleanupRuntimeStateKey(t, ctx, hcc.pluginID, "demo")
	})

	setResponse := invokeRuntimeHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodRuntimeStateSet,
		pluginbridge.MarshalHostCallStateSetRequest(&pluginbridge.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-1",
		}),
	)
	if setResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected state.set success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}

	getResponse := invokeRuntimeHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodRuntimeStateGet,
		pluginbridge.MarshalHostCallStateGetRequest(&pluginbridge.HostCallStateGetRequest{Key: "demo"}),
	)
	if getResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected state.get success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := pluginbridge.UnmarshalHostCallStateGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("expected state.get payload decode to succeed, got error: %v", err)
	}
	if !getPayload.Found || getPayload.Value != "value-1" {
		t.Fatalf("expected stored state value to round-trip, got %#v", getPayload)
	}

	deleteResponse := invokeRuntimeHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodRuntimeStateDelete,
		pluginbridge.MarshalHostCallStateDeleteRequest(&pluginbridge.HostCallStateDeleteRequest{Key: "demo"}),
	)
	if deleteResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected state.delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
}

func TestHandleHostServiceInvokeRuntimeInfoNowAndNode(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-info",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceRuntime,
				Methods: []string{
					pluginbridge.HostServiceMethodRuntimeInfoNow,
					pluginbridge.HostServiceMethodRuntimeInfoNode,
				},
			},
		},
	}

	nowResponse := invokeRuntimeHostService(t, hcc, pluginbridge.HostServiceMethodRuntimeInfoNow, nil)
	if nowResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected info.now success, got status=%d payload=%s", nowResponse.Status, string(nowResponse.Payload))
	}
	nowPayload, err := pluginbridge.UnmarshalHostServiceValueResponse(nowResponse.Payload)
	if err != nil {
		t.Fatalf("expected info.now payload decode to succeed, got error: %v", err)
	}
	if strings.TrimSpace(nowPayload.Value) == "" {
		t.Fatal("expected info.now value to be non-empty")
	}

	nodeResponse := invokeRuntimeHostService(t, hcc, pluginbridge.HostServiceMethodRuntimeInfoNode, nil)
	if nodeResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected info.node success, got status=%d payload=%s", nodeResponse.Status, string(nodeResponse.Payload))
	}
	nodePayload, err := pluginbridge.UnmarshalHostServiceValueResponse(nodeResponse.Payload)
	if err != nil {
		t.Fatalf("expected info.node payload decode to succeed, got error: %v", err)
	}
	if strings.TrimSpace(nodePayload.Value) == "" {
		t.Fatal("expected info.node value to be non-empty")
	}
}

func invokeRuntimeHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceRuntime,
		Method:  method,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
}

func cleanupRuntimeStateKey(t *testing.T, ctx context.Context, pluginID string, key string) {
	t.Helper()
	if _, err := g.DB().Model(pluginStateTable).Ctx(ctx).
		Where("plugin_id", pluginID).
		Where("state_key", key).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup runtime state key %s/%s: %v", pluginID, key, err)
	}
}
