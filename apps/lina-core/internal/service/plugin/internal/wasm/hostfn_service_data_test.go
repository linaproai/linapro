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
	"lina-core/pkg/pluginbridge"
)

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
			pluginbridge.CapabilityDataRead:   {},
			pluginbridge.CapabilityDataMutate: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceData,
				Methods: []string{
					pluginbridge.HostServiceMethodDataList,
					pluginbridge.HostServiceMethodDataGet,
					pluginbridge.HostServiceMethodDataCreate,
					pluginbridge.HostServiceMethodDataUpdate,
					pluginbridge.HostServiceMethodDataDelete,
					pluginbridge.HostServiceMethodDataTransaction,
				},
				Tables: []string{table},
			},
		},
		executionSource: pluginbridge.ExecutionSourceRoute,
		identity: &pluginbridge.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			IsSuperAdmin: true,
		},
	}

	createResponse := invokeDataHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodDataCreate,
		table,
		&pluginbridge.HostServiceDataMutationRequest{
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
	if createResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("create expected success, got status=%d payload=%s", createResponse.Status, string(createResponse.Payload))
	}
	createPayload, err := pluginbridge.UnmarshalHostServiceDataMutationResponse(createResponse.Payload)
	if err != nil {
		t.Fatalf("decode create payload failed: %v", err)
	}
	if len(createPayload.KeyJSON) == 0 {
		t.Fatalf("expected create response key, got %#v", createPayload)
	}

	listResponse := invokeDataHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodDataList,
		table,
		&pluginbridge.HostServiceDataListRequest{
			Filters:  map[string]string{"pluginId": pluginMarker},
			PageNum:  1,
			PageSize: 10,
		},
	)
	if listResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("list expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	listPayload, err := pluginbridge.UnmarshalHostServiceDataListResponse(listResponse.Payload)
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

func TestHandleHostServiceInvokeDataRejectsAnonymousRequestAccess(t *testing.T) {
	table := "sys_plugin_node_state"
	hcc := &hostCallContext{
		pluginID: "test-plugin-wasm-data",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityDataRead: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceData,
				Methods: []string{pluginbridge.HostServiceMethodDataList},
				Tables:  []string{table},
			},
		},
		executionSource: pluginbridge.ExecutionSourceRoute,
	}

	response := invokeDataHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodDataList,
		table,
		&pluginbridge.HostServiceDataListRequest{
			PageNum:  1,
			PageSize: 10,
		},
	)
	if response.Status == pluginbridge.HostCallStatusSuccess {
		t.Fatal("expected anonymous request access to be rejected")
	}
	if !strings.Contains(string(response.Payload), "登录用户") {
		t.Fatalf("expected denial reason to mention login context, got %s", string(response.Payload))
	}
}

func invokeDataHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	table string,
	request any,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	var payload []byte
	switch typedRequest := request.(type) {
	case *pluginbridge.HostServiceDataListRequest:
		payload = pluginbridge.MarshalHostServiceDataListRequest(typedRequest)
	case *pluginbridge.HostServiceDataMutationRequest:
		payload = pluginbridge.MarshalHostServiceDataMutationRequest(typedRequest)
	case *pluginbridge.HostServiceDataGetRequest:
		payload = pluginbridge.MarshalHostServiceDataGetRequest(typedRequest)
	case *pluginbridge.HostServiceDataTransactionRequest:
		payload = pluginbridge.MarshalHostServiceDataTransactionRequest(typedRequest)
	default:
		t.Fatalf("unsupported data host service request type: %T", request)
	}

	envelope := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceData,
		Method:  method,
		Table:   table,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), hcc, pluginbridge.MarshalHostServiceRequestEnvelope(envelope))
}

func cleanupWasmTestNodeStates(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup wasm test node states for %s: %v", pluginID, err)
	}
}

func mustMarshalWasmJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return data
}

func mustUnmarshalWasmRecord(t *testing.T, data []byte) map[string]any {
	t.Helper()
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return record
}
