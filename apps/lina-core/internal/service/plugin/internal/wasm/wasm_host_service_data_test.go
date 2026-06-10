// This file tests structured data host service dispatch and authorization
// error handling.
package wasm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestHandleHostServiceInvokeDataLifecycle verifies governed data CRUD host calls.
func TestHandleHostServiceInvokeDataLifecycle(t *testing.T) {
	ctx := context.Background()
	table := "sys_plugin_node_state"
	pluginMarker := "test-wasm-data-lifecycle"
	cleanupWasmTestNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupWasmTestNodeStates(t, ctx, pluginMarker)
	})

	hcc := &hostCallContext{
		pluginID: "test-plugin-wasm-data",
		capabilities: map[string]struct{}{
			protocol.CapabilityDataRead:   {},
			protocol.CapabilityDataMutate: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceData,
				Methods: []string{
					protocol.HostServiceMethodDataList,
					protocol.HostServiceMethodDataGet,
					protocol.HostServiceMethodDataCreate,
					protocol.HostServiceMethodDataUpdate,
					protocol.HostServiceMethodDataDelete,
					protocol.HostServiceMethodDataTransaction,
				},
				Tables: []string{table},
			},
		},
		executionSource: protocol.ExecutionSourceRoute,
		identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
	}

	createResponse := invokeDataHostService(
		t,
		hcc,
		protocol.HostServiceMethodDataCreate,
		table,
		&protocol.HostServiceDataMutationRequest{
			RecordJSON: mustMarshalWasmJSON(t, map[string]any{
				"pluginId":     pluginMarker,
				"releaseId":    1,
				"nodeKey":      "node-wasm-1",
				"desiredState": "running",
				"currentState": "pending",
				"generation":   1,
				"errorMessage": "",
			}),
		},
	)
	if createResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("create expected success, got status=%d payload=%s", createResponse.Status, string(createResponse.Payload))
	}
	createPayload, err := protocol.UnmarshalHostServiceDataMutationResponse(createResponse.Payload)
	if err != nil {
		t.Fatalf("decode create payload failed: %v", err)
	}
	if len(createPayload.KeyJSON) == 0 {
		t.Fatalf("expected create response key, got %#v", createPayload)
	}

	listResponse := invokeDataHostService(
		t,
		hcc,
		protocol.HostServiceMethodDataList,
		table,
		&protocol.HostServiceDataListRequest{
			PlanJSON: mustMarshalWasmJSON(t, map[string]any{
				"table":  table,
				"action": "list",
				"filters": []map[string]any{
					{"field": "pluginId", "operator": "eq", "valueJson": mustMarshalWasmJSON(t, pluginMarker)},
				},
				"page": map[string]any{"pageNum": 1, "pageSize": 10},
			}),
		},
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("list expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	listPayload, err := protocol.UnmarshalHostServiceDataListResponse(listResponse.Payload)
	if err != nil {
		t.Fatalf("decode list payload failed: %v", err)
	}
	if listPayload.Total != 1 || len(listPayload.Records) != 1 {
		t.Fatalf("unexpected list payload: %#v", listPayload)
	}
	record := mustUnmarshalWasmRecord(t, listPayload.Records[0])
	if record["pluginId"] != pluginMarker {
		t.Fatalf("unexpected list record: %#v", record)
	}
}

// TestHandleHostServiceInvokeDataRejectsAnonymousRequestAccess verifies request-only data access needs identity.
func TestHandleHostServiceInvokeDataRejectsAnonymousRequestAccess(t *testing.T) {
	table := "sys_plugin_node_state"
	hcc := &hostCallContext{
		pluginID: "test-plugin-wasm-data",
		capabilities: map[string]struct{}{
			protocol.CapabilityDataRead: {},
		},
		hostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceData,
				Methods: []string{protocol.HostServiceMethodDataList},
				Tables:  []string{table},
			},
		},
		executionSource: protocol.ExecutionSourceRoute,
	}

	response := invokeDataHostService(
		t,
		hcc,
		protocol.HostServiceMethodDataList,
		table,
		&protocol.HostServiceDataListRequest{
			PlanJSON: mustMarshalWasmJSON(t, map[string]any{
				"table":  table,
				"action": "list",
				"page":   map[string]any{"pageNum": 1, "pageSize": 10},
			}),
		},
	)
	if response.Status == protocol.HostCallStatusSuccess {
		t.Fatal("expected anonymous request access to be rejected")
	}
	if !strings.Contains(string(response.Payload), "authenticated user") {
		t.Fatalf("expected denial reason to mention login context, got %s", string(response.Payload))
	}
}

// invokeDataHostService marshals and dispatches one data host service request.
func invokeDataHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	table string,
	request any,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	var payload []byte
	switch typedRequest := request.(type) {
	case *protocol.HostServiceDataListRequest:
		payload = protocol.MarshalHostServiceDataListRequest(typedRequest)
	case *protocol.HostServiceDataMutationRequest:
		payload = protocol.MarshalHostServiceDataMutationRequest(typedRequest)
	case *protocol.HostServiceDataGetRequest:
		payload = protocol.MarshalHostServiceDataGetRequest(typedRequest)
	case *protocol.HostServiceDataTransactionRequest:
		payload = protocol.MarshalHostServiceDataTransactionRequest(typedRequest)
	default:
		t.Fatalf("unsupported data host service request type: %T", request)
	}

	envelope := &protocol.HostServiceRequestEnvelope{
		Service: protocol.HostServiceData,
		Method:  method,
		Table:   table,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), hcc, protocol.MarshalHostServiceRequestEnvelope(envelope))
}

// cleanupWasmTestNodeStates removes sys_plugin_node_state rows created by wasm data tests.
func cleanupWasmTestNodeStates(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup wasm test node states for %s: %v", pluginID, err)
	}
}

// mustMarshalWasmJSON marshals test values and fails on error.
func mustMarshalWasmJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return data
}

// mustUnmarshalWasmRecord unmarshals JSON objects and fails on error.
func mustUnmarshalWasmRecord(t *testing.T, data []byte) map[string]any {
	t.Helper()
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return record
}
