// This file tests runtime host service methods including log/state/info
// dispatch and capability validation.

package wasm

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

// createPluginStateTableSQL provisions the plugin runtime state table required
// by runtime host-service state lifecycle tests.
const createPluginStateTableSQL = `
CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plugin_id   VARCHAR(64) NOT NULL DEFAULT '',
    state_key   VARCHAR(255) NOT NULL DEFAULT '',
    state_value TEXT,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_state_plugin_key ON sys_plugin_state (plugin_id, state_key);
`

// TestHandleHostServiceInvokeRuntimeStateLifecycle verifies runtime state
// get/set/delete calls persist and remove plugin-scoped values correctly.
func TestHandleHostServiceInvokeRuntimeStateLifecycle(t *testing.T) {
	ctx := context.Background()
	for _, statement := range dialect.SplitSQLStatements(createPluginStateTableSQL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("expected plugin state table to be created, got error: %v\nSQL:\n%s", err, statement)
		}
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

	updateResponse := invokeRuntimeHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodRuntimeStateSet,
		pluginbridge.MarshalHostCallStateSetRequest(&pluginbridge.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-2",
		}),
	)
	if updateResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected second state.set success, got status=%d payload=%s", updateResponse.Status, string(updateResponse.Payload))
	}
	getUpdatedResponse := invokeRuntimeHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodRuntimeStateGet,
		pluginbridge.MarshalHostCallStateGetRequest(&pluginbridge.HostCallStateGetRequest{Key: "demo"}),
	)
	if getUpdatedResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected updated state.get success, got status=%d payload=%s", getUpdatedResponse.Status, string(getUpdatedResponse.Payload))
	}
	getUpdatedPayload, err := pluginbridge.UnmarshalHostCallStateGetResponse(getUpdatedResponse.Payload)
	if err != nil {
		t.Fatalf("expected updated state.get payload decode to succeed, got error: %v", err)
	}
	if !getUpdatedPayload.Found || getUpdatedPayload.Value != "value-2" {
		t.Fatalf("expected updated state value to round-trip, got %#v", getUpdatedPayload)
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

// TestHandleHostServiceInvokeRuntimeInfoNowAndNode verifies runtime info
// methods return non-empty host metadata payloads.
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

// invokeRuntimeHostService dispatches one runtime host-service request and
// returns the raw response envelope for assertions.
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

// cleanupRuntimeStateKey deletes one plugin runtime state row so lifecycle
// tests can run repeatedly without leftover state.
func cleanupRuntimeStateKey(t *testing.T, ctx context.Context, pluginID string, key string) {
	t.Helper()
	if _, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: pluginID, StateKey: key}).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup runtime state key %s/%s: %v", pluginID, key, err)
	}
}
