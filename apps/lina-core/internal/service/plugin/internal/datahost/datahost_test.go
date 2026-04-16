// This file tests governed data service execution, typed-plan handling, and
// transactional mutations.
package datahost

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	plugindbhost "lina-core/pkg/plugindb/host"
	"lina-core/pkg/plugindb/shared"
)

func TestExecuteCRUDLifecycle(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &pluginbridge.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		IsSuperAdmin: true,
	}
	pluginMarker := "test-datahost-crud"
	cleanupNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupNodeStates(t, ctx, pluginMarker)
	})

	createRecord := map[string]any{
		"pluginId":     pluginMarker,
		"releaseId":    1,
		"nodeKey":      "node-crud-1",
		"desiredState": "running",
		"currentState": "pending",
		"generation":   1,
		"errorMessage": "",
	}
	createResponse, err := ExecuteCreate(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataMutationRequest{
			RecordJSON: mustMarshalJSON(t, createRecord),
		},
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	keyValue := mustUnmarshalJSONValue(t, createResponse.KeyJSON)
	if keyValue == nil {
		t.Fatalf("expected create response key, got %#v", createResponse)
	}

	listResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataListRequest{
			Filters:  map[string]string{"pluginId": pluginMarker},
			PageNum:  1,
			PageSize: 10,
		},
	)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listResponse.Total != 1 || len(listResponse.Records) != 1 {
		t.Fatalf("unexpected list response: %#v", listResponse)
	}

	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataGetRequest{
			KeyJSON: createResponse.KeyJSON,
		},
	)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !getResponse.Found {
		t.Fatalf("expected get to find created row")
	}
	gotRecord := mustUnmarshalJSONRecord(t, getResponse.RecordJSON)
	if gotRecord["pluginId"] != pluginMarker {
		t.Fatalf("unexpected get record: %#v", gotRecord)
	}

	updateResponse, err := ExecuteUpdate(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataMutationRequest{
			KeyJSON: createResponse.KeyJSON,
			RecordJSON: mustMarshalJSON(t, map[string]any{
				"currentState": "running",
				"errorMessage": "updated",
			}),
		},
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updateResponse.AffectedRows != 1 {
		t.Fatalf("expected update affectedRows=1, got %#v", updateResponse)
	}

	deleteResponse, err := ExecuteDelete(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataMutationRequest{
			KeyJSON: createResponse.KeyJSON,
		},
	)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if deleteResponse.AffectedRows != 1 {
		t.Fatalf("expected delete affectedRows=1, got %#v", deleteResponse)
	}
}

func TestExecuteTransactionAppliesMutationsAtomically(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResourceWithNodeKey()
	identity := &pluginbridge.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		IsSuperAdmin: true,
	}
	pluginMarker := "test-datahost-transaction"
	nodeKey := "node-transaction-1"
	cleanupNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupNodeStates(t, ctx, pluginMarker)
	})

	response, err := ExecuteTransaction(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataTransactionRequest{
			Operations: []*pluginbridge.HostServiceDataTransactionOperation{
				{
					Method: pluginbridge.HostServiceMethodDataCreate,
					RecordJSON: mustMarshalJSON(t, map[string]any{
						"pluginId":     pluginMarker,
						"releaseId":    1,
						"nodeKey":      nodeKey,
						"desiredState": "running",
						"currentState": "pending",
						"generation":   1,
						"errorMessage": "",
					}),
				},
				{
					Method:  pluginbridge.HostServiceMethodDataUpdate,
					KeyJSON: mustMarshalJSON(t, nodeKey),
					RecordJSON: mustMarshalJSON(t, map[string]any{
						"currentState": "running",
					}),
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}
	if response.AffectedRows != 2 || len(response.Results) != 2 {
		t.Fatalf("unexpected transaction response: %#v", response)
	}

	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataGetRequest{
			KeyJSON: mustMarshalJSON(t, nodeKey),
		},
	)
	if err != nil {
		t.Fatalf("get after transaction failed: %v", err)
	}
	record := mustUnmarshalJSONRecord(t, getResponse.RecordJSON)
	if record["currentState"] != "running" {
		t.Fatalf("expected currentState=running after transaction, got %#v", record)
	}
}

func TestExecuteListSupportsPlugindbPlan(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &pluginbridge.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		IsSuperAdmin: true,
	}
	pluginMarker := "test-datahost-plan-list"
	cleanupNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupNodeStates(t, ctx, pluginMarker)
	})

	for _, item := range []map[string]any{
		{
			"pluginId":     pluginMarker,
			"releaseId":    1,
			"nodeKey":      "adv-1",
			"desiredState": "running",
			"currentState": "pending",
			"generation":   1,
			"errorMessage": "",
		},
		{
			"pluginId":     pluginMarker,
			"releaseId":    1,
			"nodeKey":      "adv-2",
			"desiredState": "running",
			"currentState": "running",
			"generation":   2,
			"errorMessage": "",
		},
	} {
		if _, err := ExecuteCreate(
			ctx,
			"test-plugin-data",
			resource.Table,
			pluginbridge.ExecutionSourceRoute,
			identity,
			resource,
			&pluginbridge.HostServiceDataMutationRequest{RecordJSON: mustMarshalJSON(t, item)},
		); err != nil {
			t.Fatalf("ExecuteCreate failed: %v", err)
		}
	}

	planJSON, err := shared.MarshalQueryPlanJSON(&shared.DataQueryPlan{
		Table:  resource.Table,
		Action: shared.DataPlanActionList,
		Fields: []string{"nodeKey", "currentState"},
		Filters: []*shared.DataFilter{
			mustNewEQFilter(t, "pluginId", pluginMarker),
			mustNewINFilter(t, "currentState", []string{"pending", "running"}),
			mustNewLikeFilter(t, "nodeKey", "adv-"),
		},
		Orders: []*shared.DataOrder{shared.NewDESCOrder("nodeKey")},
		Page:   &shared.DataPagination{PageNum: 1, PageSize: 10},
	})
	if err != nil {
		t.Fatalf("MarshalQueryPlanJSON failed: %v", err)
	}

	listResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataListRequest{PlanJSON: planJSON},
	)
	if err != nil {
		t.Fatalf("ExecuteList failed: %v", err)
	}
	if listResponse.Total != 2 || len(listResponse.Records) != 2 {
		t.Fatalf("unexpected list response: %#v", listResponse)
	}
	firstRecord := mustUnmarshalJSONRecord(t, listResponse.Records[0])
	if firstRecord["nodeKey"] != "adv-2" || firstRecord["currentState"] != "running" {
		t.Fatalf("unexpected first record: %#v", firstRecord)
	}
	if _, exists := firstRecord["pluginId"]; exists {
		t.Fatalf("expected selected fields only, got %#v", firstRecord)
	}

	countPlanJSON, err := shared.MarshalQueryPlanJSON(&shared.DataQueryPlan{
		Table:  resource.Table,
		Action: shared.DataPlanActionCount,
		Filters: []*shared.DataFilter{
			mustNewEQFilter(t, "pluginId", pluginMarker),
		},
	})
	if err != nil {
		t.Fatalf("MarshalQueryPlanJSON count failed: %v", err)
	}
	countResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataListRequest{PlanJSON: countPlanJSON},
	)
	if err != nil {
		t.Fatalf("ExecuteList count failed: %v", err)
	}
	if countResponse.Total != 2 || len(countResponse.Records) != 0 {
		t.Fatalf("unexpected count response: %#v", countResponse)
	}
}

func TestExecuteGetSupportsPlugindbFieldSelection(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &pluginbridge.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		IsSuperAdmin: true,
	}
	pluginMarker := "test-datahost-plan-get"
	cleanupNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupNodeStates(t, ctx, pluginMarker)
	})

	createResponse, err := ExecuteCreate(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataMutationRequest{
			RecordJSON: mustMarshalJSON(t, map[string]any{
				"pluginId":     pluginMarker,
				"releaseId":    1,
				"nodeKey":      "adv-get-1",
				"desiredState": "running",
				"currentState": "pending",
				"generation":   1,
				"errorMessage": "",
			}),
		},
	)
	if err != nil {
		t.Fatalf("ExecuteCreate failed: %v", err)
	}

	planJSON, err := shared.MarshalQueryPlanJSON(&shared.DataQueryPlan{
		Table:   resource.Table,
		Action:  shared.DataPlanActionGet,
		Fields:  []string{"currentState"},
		KeyJSON: append([]byte(nil), createResponse.KeyJSON...),
	})
	if err != nil {
		t.Fatalf("MarshalQueryPlanJSON failed: %v", err)
	}
	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		pluginbridge.ExecutionSourceRoute,
		identity,
		resource,
		&pluginbridge.HostServiceDataGetRequest{
			KeyJSON:  append([]byte(nil), createResponse.KeyJSON...),
			PlanJSON: planJSON,
		},
	)
	if err != nil {
		t.Fatalf("ExecuteGet failed: %v", err)
	}
	if !getResponse.Found {
		t.Fatal("expected get response to find record")
	}
	record := mustUnmarshalJSONRecord(t, getResponse.RecordJSON)
	if len(record) != 1 || record["currentState"] != "pending" {
		t.Fatalf("unexpected selected record: %#v", record)
	}
}

func TestPluginDataDBDoCommitRejectsUnauthorizedTable(t *testing.T) {
	db, err := getPluginDataDB()
	if err != nil {
		t.Fatalf("getPluginDataDB failed: %v", err)
	}
	ctx := withPluginDataAudit(context.Background(), &plugindbhost.AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "sys_plugin_node_state",
		Method:        pluginbridge.HostServiceMethodDataDelete,
		ResourceTable: "sys_plugin_node_state",
	})
	_, err = db.Ctx(ctx).Exec(ctx, "DELETE FROM sys_plugin WHERE plugin_id = ?", "forbidden")
	if err == nil {
		t.Fatal("expected DoCommit to reject unauthorized table")
	}
	if !strings.Contains(err.Error(), "授权表") {
		t.Fatalf("expected unauthorized table error, got %v", err)
	}
}

func buildTestNodeStateResource() *catalog.ResourceSpec {
	return &catalog.ResourceSpec{
		Key:   "nodeStates",
		Type:  catalog.ResourceSpecTypeTableList.String(),
		Table: "sys_plugin_node_state",
		Fields: []*catalog.ResourceField{
			{Name: "id", Column: "id"},
			{Name: "pluginId", Column: "plugin_id"},
			{Name: "releaseId", Column: "release_id"},
			{Name: "nodeKey", Column: "node_key"},
			{Name: "desiredState", Column: "desired_state"},
			{Name: "currentState", Column: "current_state"},
			{Name: "generation", Column: "generation"},
			{Name: "errorMessage", Column: "error_message"},
		},
		Filters: []*catalog.ResourceQuery{
			{Param: "pluginId", Column: "plugin_id", Operator: catalog.ResourceFilterOperatorEQ.String()},
		},
		OrderBy: catalog.ResourceOrderBySpec{
			Column:    "id",
			Direction: catalog.ResourceOrderDirectionASC.String(),
		},
		Operations: []string{
			pluginbridge.HostServiceMethodDataList,
			pluginbridge.HostServiceMethodDataGet,
			pluginbridge.HostServiceMethodDataCreate,
			pluginbridge.HostServiceMethodDataUpdate,
			pluginbridge.HostServiceMethodDataDelete,
			pluginbridge.HostServiceMethodDataTransaction,
		},
		KeyField: "id",
		WritableFields: []string{
			"pluginId",
			"releaseId",
			"nodeKey",
			"desiredState",
			"currentState",
			"generation",
			"errorMessage",
		},
		Access: catalog.ResourceAccessModeRequest.String(),
	}
}

func buildTestNodeStateResourceWithNodeKey() *catalog.ResourceSpec {
	resource := buildTestNodeStateResource()
	resource.KeyField = "nodeKey"
	return resource
}

func cleanupNodeStates(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete plugin node states for %s: %v", pluginID, err)
	}
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return data
}

func mustUnmarshalJSONValue(t *testing.T, data []byte) any {
	t.Helper()
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return value
}

func mustUnmarshalJSONRecord(t *testing.T, data []byte) map[string]any {
	t.Helper()
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("json unmarshal record failed: %v", err)
	}
	return record
}

func mustNewEQFilter(t *testing.T, field string, value any) *shared.DataFilter {
	t.Helper()
	filter, err := shared.NewEQFilter(field, value)
	if err != nil {
		t.Fatalf("build eq filter failed: %v", err)
	}
	return filter
}

func mustNewINFilter(t *testing.T, field string, values any) *shared.DataFilter {
	t.Helper()
	filter, err := shared.NewINFilter(field, values)
	if err != nil {
		t.Fatalf("build in filter failed: %v", err)
	}
	return filter
}

func mustNewLikeFilter(t *testing.T, field string, value any) *shared.DataFilter {
	t.Helper()
	filter, err := shared.NewLikeFilter(field, value)
	if err != nil {
		t.Fatalf("build like filter failed: %v", err)
	}
	return filter
}
