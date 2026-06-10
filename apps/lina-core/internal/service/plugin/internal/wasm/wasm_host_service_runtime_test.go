// This file tests runtime host service methods including log/state/info
// dispatch and capability validation.

package wasm

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// createPluginStateTableSQL provisions the plugin runtime state table required
// by runtime host-service state lifecycle tests.
const createPluginStateTableSQL = `
CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plugin_id   VARCHAR(64) NOT NULL DEFAULT '',
    tenant_id   INT NOT NULL DEFAULT 0,
    state_key   VARCHAR(255) NOT NULL DEFAULT '',
    state_value TEXT,
    enabled     BOOL NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_state_plugin_tenant_key ON sys_plugin_state (plugin_id, tenant_id, state_key);
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
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeStateGet,
					protocol.HostServiceMethodRuntimeStateSet,
					protocol.HostServiceMethodRuntimeStateDelete,
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
		protocol.HostServiceMethodRuntimeStateSet,
		protocol.MarshalHostCallStateSetRequest(&protocol.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-1",
		}),
	)
	if setResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.set success, got status=%d payload=%s", setResponse.Status, string(setResponse.Payload))
	}

	getResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGet,
		protocol.MarshalHostCallStateGetRequest(&protocol.HostCallStateGetRequest{Key: "demo"}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.get success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := protocol.UnmarshalHostCallStateGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("expected state.get payload decode to succeed, got error: %v", err)
	}
	if !getPayload.Found || getPayload.Value != "value-1" {
		t.Fatalf("expected stored state value to round-trip, got %#v", getPayload)
	}

	updateResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateSet,
		protocol.MarshalHostCallStateSetRequest(&protocol.HostCallStateSetRequest{
			Key:   "demo",
			Value: "value-2",
		}),
	)
	if updateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected second state.set success, got status=%d payload=%s", updateResponse.Status, string(updateResponse.Payload))
	}
	getUpdatedResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateGet,
		protocol.MarshalHostCallStateGetRequest(&protocol.HostCallStateGetRequest{Key: "demo"}),
	)
	if getUpdatedResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected updated state.get success, got status=%d payload=%s", getUpdatedResponse.Status, string(getUpdatedResponse.Payload))
	}
	getUpdatedPayload, err := protocol.UnmarshalHostCallStateGetResponse(getUpdatedResponse.Payload)
	if err != nil {
		t.Fatalf("expected updated state.get payload decode to succeed, got error: %v", err)
	}
	if !getUpdatedPayload.Found || getUpdatedPayload.Value != "value-2" {
		t.Fatalf("expected updated state value to round-trip, got %#v", getUpdatedPayload)
	}

	deleteResponse := invokeRuntimeHostService(
		t,
		hcc,
		protocol.HostServiceMethodRuntimeStateDelete,
		protocol.MarshalHostCallStateDeleteRequest(&protocol.HostCallStateDeleteRequest{Key: "demo"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected state.delete success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}
}

// TestHandleHostServiceInvokeRuntimeInfoNowAndNode verifies runtime info
// methods return non-empty host metadata payloads.
func TestHandleHostServiceInvokeRuntimeInfoNowAndNode(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin-runtime-info",
		capabilities: map[string]struct{}{
			protocol.CapabilityRuntime: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeInfoNow,
					protocol.HostServiceMethodRuntimeInfoNode,
				},
			},
		},
	}

	beforeMillis := time.Now().Add(-1 * time.Second).UnixMilli()
	nowResponse := invokeRuntimeHostService(t, hcc, protocol.HostServiceMethodRuntimeInfoNow, nil)
	afterMillis := time.Now().Add(1 * time.Second).UnixMilli()
	if nowResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected info.now success, got status=%d payload=%s", nowResponse.Status, string(nowResponse.Payload))
	}
	nowPayload, err := protocol.UnmarshalHostServiceValueResponse(nowResponse.Payload)
	if err != nil {
		t.Fatalf("expected info.now payload decode to succeed, got error: %v", err)
	}
	if strings.TrimSpace(nowPayload.Value) == "" {
		t.Fatal("expected info.now value to be non-empty")
	}
	nowMillis, err := strconv.ParseInt(nowPayload.Value, 10, 64)
	if err != nil {
		t.Fatalf("expected info.now value to be Unix milliseconds, got %q: %v", nowPayload.Value, err)
	}
	if nowMillis < beforeMillis || nowMillis > afterMillis {
		t.Fatalf("expected info.now value within test window, got %d outside [%d,%d]", nowMillis, beforeMillis, afterMillis)
	}

	nodeResponse := invokeRuntimeHostService(t, hcc, protocol.HostServiceMethodRuntimeInfoNode, nil)
	if nodeResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected info.node success, got status=%d payload=%s", nodeResponse.Status, string(nodeResponse.Payload))
	}
	nodePayload, err := protocol.UnmarshalHostServiceValueResponse(nodeResponse.Payload)
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
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceRuntime,
		Method:  method,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), hcc, protocol.MarshalHostServiceRequestEnvelope(request))
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
