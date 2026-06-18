// This file tests structured data host service dispatch and authorization
// error handling.
package wasm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestHandleHostServiceInvokeDataLifecycle verifies governed data CRUD host calls.
func TestHandleHostServiceInvokeDataLifecycle(t *testing.T) {
	ctx := context.Background()
	table := "plugin_test_plugin_wasm_data_records"
	createWasmDataRecordsTable(t, ctx, table)
	t.Cleanup(func() {
		dropWasmDataRecordsTable(t, ctx, table)
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
					protocol.HostServiceMethodDataBatchGet,
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
				"pluginMarker": "test-wasm-data-lifecycle",
				"nodeKey":      "node-wasm-1",
				"currentState": "pending",
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
					{"field": "pluginMarker", "operator": "eq", "valueJson": mustMarshalWasmJSON(t, "test-wasm-data-lifecycle")},
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
	if record["pluginMarker"] != "test-wasm-data-lifecycle" {
		t.Fatalf("unexpected list record: %#v", record)
	}

	batchGetResponse := invokeDataHostService(
		t,
		hcc,
		protocol.HostServiceMethodDataBatchGet,
		table,
		&protocol.HostServiceDataBatchGetRequest{
			KeyJSON: [][]byte{append([]byte(nil), createPayload.KeyJSON...), mustMarshalWasmJSON(t, 999999)},
			Fields:  []string{"currentState"},
		},
	)
	if batchGetResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("batch_get expected success, got status=%d payload=%s", batchGetResponse.Status, string(batchGetResponse.Payload))
	}
	batchGetPayload, err := protocol.UnmarshalHostServiceDataBatchGetResponse(batchGetResponse.Payload)
	if err != nil {
		t.Fatalf("decode batch_get payload failed: %v", err)
	}
	if len(batchGetPayload.Records) != 1 || len(batchGetPayload.MissingKeyJSON) != 1 {
		t.Fatalf("unexpected batch_get payload: %#v", batchGetPayload)
	}
	batchRecord := mustUnmarshalWasmRecord(t, batchGetPayload.Records[0])
	if len(batchRecord) != 1 || batchRecord["currentState"] != "pending" {
		t.Fatalf("unexpected batch_get record projection: %#v", batchRecord)
	}
}

// TestHandleHostServiceInvokeDataRejectsAnonymousRequestAccess verifies request-only data access needs identity.
func TestHandleHostServiceInvokeDataRejectsAnonymousRequestAccess(t *testing.T) {
	ctx := context.Background()
	table := "plugin_test_plugin_wasm_data_records"
	createWasmDataRecordsTable(t, ctx, table)
	t.Cleanup(func() {
		dropWasmDataRecordsTable(t, ctx, table)
	})

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
	case *protocol.HostServiceDataBatchGetRequest:
		payload = protocol.MarshalHostServiceDataBatchGetRequest(typedRequest)
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
	return handleHostServiceInvoke(context.Background(), withTestHostCallRuntime(t, hcc), protocol.MarshalHostServiceRequestEnvelope(envelope))
}

func createWasmDataRecordsTable(t *testing.T, ctx context.Context, table string) {
	t.Helper()
	dropWasmDataRecordsTable(t, ctx, table)
	if _, err := g.DB().Exec(ctx, `
CREATE TABLE plugin_test_plugin_wasm_data_records (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plugin_marker VARCHAR(64) NOT NULL DEFAULT '',
    node_key VARCHAR(64) NOT NULL DEFAULT '',
    current_state VARCHAR(32) NOT NULL DEFAULT ''
)`); err != nil {
		t.Fatalf("failed to create wasm data records table %s: %v", table, err)
	}
}

func dropWasmDataRecordsTable(t *testing.T, ctx context.Context, table string) {
	t.Helper()
	if table != "plugin_test_plugin_wasm_data_records" {
		t.Fatalf("unsafe wasm data records table name: %s", table)
	}
	if _, err := g.DB().Exec(ctx, "DROP TABLE IF EXISTS plugin_test_plugin_wasm_data_records"); err != nil {
		t.Fatalf("failed to drop wasm data records table %s: %v", table, err)
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
