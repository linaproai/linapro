// This file tests governed data service execution, typed-plan handling, and
// transactional mutations.
package datahost

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	recordstorehost "lina-core/internal/service/plugin/internal/datahost/internal/host"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestExecuteCRUDLifecycle verifies governed create, list, get, update, and delete flows.
func TestExecuteCRUDLifecycle(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &protocol.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		DataScope:    1,
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
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataMutationRequest{
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

	listPlanJSON := mustMarshalJSON(t, map[string]any{
		"table":  resource.Table,
		"action": "list",
		"filters": []map[string]any{
			{"field": "pluginId", "operator": "eq", "valueJson": mustMarshalJSON(t, pluginMarker)},
		},
		"page": map[string]any{"pageNum": 1, "pageSize": 10},
	})
	listResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataListRequest{
			PlanJSON: listPlanJSON,
		},
	)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listResponse.Total != 1 || len(listResponse.Records) != 1 {
		t.Fatalf("unexpected list response: %#v", listResponse)
	}

	getPlanJSON := mustMarshalJSON(t, map[string]any{
		"table":   resource.Table,
		"action":  "get",
		"keyJson": append([]byte(nil), createResponse.KeyJSON...),
	})
	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataGetRequest{
			PlanJSON: getPlanJSON,
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
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataMutationRequest{
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
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataMutationRequest{
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

// TestExecuteTransactionAppliesMutationsAtomically verifies transactional
// mutations commit as one atomic unit.
func TestExecuteTransactionAppliesMutationsAtomically(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResourceWithNodeKey()
	identity := &protocol.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		DataScope:    1,
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
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataTransactionRequest{
			Operations: []*protocol.HostServiceDataTransactionOperation{
				{
					Method: protocol.HostServiceMethodDataCreate,
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
					Method:  protocol.HostServiceMethodDataUpdate,
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

	getPlanJSON := mustMarshalJSON(t, map[string]any{
		"table":   resource.Table,
		"action":  "get",
		"keyJson": mustMarshalJSON(t, nodeKey),
	})
	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataGetRequest{
			PlanJSON: getPlanJSON,
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

// TestBuildAuthorizedTableContractSkipsPostgreSQLIdentityFields verifies live
// PostgreSQL identity primary keys stay outside guest writable fields.
func TestBuildAuthorizedTableContractSkipsPostgreSQLIdentityFields(t *testing.T) {
	ctx := context.Background()
	tableName := "test_datahost_identity_contract"
	dropIdentityContractTable(t, ctx, tableName)
	t.Cleanup(func() {
		dropIdentityContractTable(t, ctx, tableName)
	})
	if _, err := g.DB().Exec(ctx, `
CREATE TABLE test_datahost_identity_contract (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title      VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`); err != nil {
		t.Fatalf("failed to create identity contract table: %v", err)
	}

	resource, err := BuildAuthorizedTableContract(ctx, tableName, []string{
		protocol.HostServiceMethodDataCreate,
		protocol.HostServiceMethodDataList,
		protocol.HostServiceMethodDataUpdate,
	})
	if err != nil {
		t.Fatalf("build authorized table contract failed: %v", err)
	}
	if resource.KeyField != "id" {
		t.Fatalf("expected identity primary key to remain keyField=id, got %s", resource.KeyField)
	}
	if containsFieldName(resource.WritableFields, "id") {
		t.Fatalf("expected identity id field to be auto-managed, got writableFields=%#v", resource.WritableFields)
	}
	if containsFieldName(resource.WritableFields, "createdAt") {
		t.Fatalf("expected createdAt field to be auto-managed, got writableFields=%#v", resource.WritableFields)
	}
	if !containsFieldName(resource.WritableFields, "title") {
		t.Fatalf("expected title to remain writable, got writableFields=%#v", resource.WritableFields)
	}
}

// TestBuildCachedAuthorizedTableContractUsesMigrationFingerprint verifies schema
// metadata is reused until the plugin migration ledger changes.
func TestBuildCachedAuthorizedTableContractUsesMigrationFingerprint(t *testing.T) {
	ctx := context.Background()
	pluginID := "test-plugin-datahost-contract-cache-migration"
	tableName := "test_datahost_contract_cache_migration"
	dropIdentityContractTable(t, ctx, tableName)
	cleanupTableContractMigrations(t, ctx, pluginID)
	t.Cleanup(func() {
		dropIdentityContractTable(t, ctx, tableName)
		cleanupTableContractMigrations(t, ctx, pluginID)
		InvalidateTableContractCache(ctx, pluginID, "test_cleanup")
	})

	createTableContractCacheTable(t, ctx, tableName, "title VARCHAR(64) NOT NULL DEFAULT ''")
	insertTableContractMigration(t, ctx, pluginID, plugintypes.MigrationDirectionInstall, "install-step-001", "v1")

	input := AuthorizedTableContractInput{
		PluginID: pluginID,
		Table:    tableName,
		Methods:  []string{protocol.HostServiceMethodDataList, protocol.HostServiceMethodDataUpdate},
	}
	resource, err := BuildCachedAuthorizedTableContract(ctx, input)
	if err != nil {
		t.Fatalf("build cached table contract failed: %v", err)
	}
	if !contractContainsField(resource, "title") {
		t.Fatalf("expected initial contract to include title, got %#v", resource.Fields)
	}

	alterTableContractCacheTable(t, ctx, tableName, "body VARCHAR(64) NOT NULL DEFAULT ''")
	cached, err := BuildCachedAuthorizedTableContract(ctx, input)
	if err != nil {
		t.Fatalf("build cached table contract after ddl failed: %v", err)
	}
	if contractContainsField(cached, "body") {
		t.Fatalf("expected unchanged migration fingerprint to reuse cached contract, got %#v", cached.Fields)
	}

	insertTableContractMigration(t, ctx, pluginID, plugintypes.MigrationDirectionUpgrade, "upgrade-step-001", "v2")
	refreshed, err := BuildCachedAuthorizedTableContract(ctx, input)
	if err != nil {
		t.Fatalf("build refreshed table contract failed: %v", err)
	}
	if !contractContainsField(refreshed, "body") {
		t.Fatalf("expected migration fingerprint change to refresh contract, got %#v", refreshed.Fields)
	}
}

// TestInvalidateTableContractCacheRefreshesLiveSchema verifies plugin-scoped
// invalidation refreshes cached contracts even when the migration key is stable.
func TestInvalidateTableContractCacheRefreshesLiveSchema(t *testing.T) {
	ctx := context.Background()
	pluginID := "test-plugin-datahost-contract-cache-invalidate"
	tableName := "test_datahost_contract_cache_invalidate"
	dropIdentityContractTable(t, ctx, tableName)
	cleanupTableContractMigrations(t, ctx, pluginID)
	t.Cleanup(func() {
		dropIdentityContractTable(t, ctx, tableName)
		cleanupTableContractMigrations(t, ctx, pluginID)
		InvalidateTableContractCache(ctx, pluginID, "test_cleanup")
	})

	createTableContractCacheTable(t, ctx, tableName, "title VARCHAR(64) NOT NULL DEFAULT ''")
	insertTableContractMigration(t, ctx, pluginID, plugintypes.MigrationDirectionInstall, "install-step-001", "v1")

	input := AuthorizedTableContractInput{
		PluginID: pluginID,
		Table:    tableName,
		Methods:  []string{protocol.HostServiceMethodDataList},
	}
	if _, err := BuildCachedAuthorizedTableContract(ctx, input); err != nil {
		t.Fatalf("build initial cached table contract failed: %v", err)
	}
	alterTableContractCacheTable(t, ctx, tableName, "body VARCHAR(64) NOT NULL DEFAULT ''")
	InvalidateTableContractCache(ctx, pluginID, "test_schema_changed")

	refreshed, err := BuildCachedAuthorizedTableContract(ctx, input)
	if err != nil {
		t.Fatalf("build invalidated table contract failed: %v", err)
	}
	if !contractContainsField(refreshed, "body") {
		t.Fatalf("expected explicit invalidation to refresh live schema, got %#v", refreshed.Fields)
	}
}

// TestBuildCachedAuthorizedTableContractUsesAuthorizationFingerprint verifies
// method authorization shrinkage does not reuse a broader cached contract.
func TestBuildCachedAuthorizedTableContractUsesAuthorizationFingerprint(t *testing.T) {
	ctx := context.Background()
	pluginID := "test-plugin-datahost-contract-cache-auth"
	tableName := "test_datahost_contract_cache_auth"
	dropIdentityContractTable(t, ctx, tableName)
	cleanupTableContractMigrations(t, ctx, pluginID)
	t.Cleanup(func() {
		dropIdentityContractTable(t, ctx, tableName)
		cleanupTableContractMigrations(t, ctx, pluginID)
		InvalidateTableContractCache(ctx, pluginID, "test_cleanup")
	})

	createTableContractCacheTable(t, ctx, tableName, "title VARCHAR(64) NOT NULL DEFAULT ''")
	insertTableContractMigration(t, ctx, pluginID, plugintypes.MigrationDirectionInstall, "install-step-001", "v1")

	full, err := BuildCachedAuthorizedTableContract(ctx, AuthorizedTableContractInput{
		PluginID: pluginID,
		Table:    tableName,
		Methods:  []string{protocol.HostServiceMethodDataList, protocol.HostServiceMethodDataUpdate},
	})
	if err != nil {
		t.Fatalf("build broad cached table contract failed: %v", err)
	}
	if !resourceAllowsOperation(full, protocol.HostServiceMethodDataUpdate) {
		t.Fatalf("expected broad contract to allow update, got %#v", full.Operations)
	}

	shrunk, err := BuildCachedAuthorizedTableContract(ctx, AuthorizedTableContractInput{
		PluginID: pluginID,
		Table:    tableName,
		Methods:  []string{protocol.HostServiceMethodDataList},
	})
	if err != nil {
		t.Fatalf("build shrunk cached table contract failed: %v", err)
	}
	if resourceAllowsOperation(shrunk, protocol.HostServiceMethodDataUpdate) {
		t.Fatalf("expected authorization shrink to reject old update method, got %#v", shrunk.Operations)
	}
}

// TestExecuteListSupportsRecordStoreCapabilityPlan verifies typed record store capability
// list plans are honored.
func TestExecuteListSupportsRecordStoreCapabilityPlan(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &protocol.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		DataScope:    1,
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
			protocol.ExecutionSourceRoute,
			identity,
			nil,
			resource,
			&protocol.HostServiceDataMutationRequest{RecordJSON: mustMarshalJSON(t, item)},
		); err != nil {
			t.Fatalf("ExecuteCreate failed: %v", err)
		}
	}

	planJSON := mustMarshalJSON(t, map[string]any{
		"table":  resource.Table,
		"action": "list",
		"fields": []string{"nodeKey", "currentState"},
		"filters": []map[string]any{
			{"field": "pluginId", "operator": "eq", "valueJson": mustMarshalJSON(t, pluginMarker)},
			{"field": "currentState", "operator": "in", "valuesJson": [][]byte{mustMarshalJSON(t, "pending"), mustMarshalJSON(t, "running")}},
			{"field": "nodeKey", "operator": "like", "valueJson": mustMarshalJSON(t, "adv-")},
		},
		"orders": []map[string]any{
			{"field": "nodeKey", "direction": "desc"},
		},
		"page": map[string]any{"pageNum": 1, "pageSize": 10},
	})

	listResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataListRequest{PlanJSON: planJSON},
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

	countPlanJSON := mustMarshalJSON(t, map[string]any{
		"table":  resource.Table,
		"action": "count",
		"filters": []map[string]any{
			{"field": "pluginId", "operator": "eq", "valueJson": mustMarshalJSON(t, pluginMarker)},
		},
	})
	countResponse, err := ExecuteList(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataListRequest{PlanJSON: countPlanJSON},
	)
	if err != nil {
		t.Fatalf("ExecuteList count failed: %v", err)
	}
	if countResponse.Total != 2 || len(countResponse.Records) != 0 {
		t.Fatalf("unexpected count response: %#v", countResponse)
	}
}

// TestBuildResourceRecordWithSelectionFallsBackToColumnName verifies selected
// field projection still works when the driver returns physical column keys.
func TestBuildResourceRecordWithSelectionFallsBackToColumnName(t *testing.T) {
	t.Parallel()

	record := buildResourceRecordWithSelection(
		map[string]interface{}{
			"user_name":  "admin",
			"event_name": "Login successful",
		},
		&catalog.ResourceSpec{
			Fields: []*catalog.ResourceField{
				{Name: "userName", Column: "user_name"},
				{Name: "eventName", Column: "event_name"},
			},
		},
		[]string{"userName", "eventName"},
	)

	if record["userName"] != "admin" {
		t.Fatalf("expected userName to fall back to user_name column, got %#v", record)
	}
	if record["eventName"] != "Login successful" {
		t.Fatalf("expected eventName to fall back to event_name column, got %#v", record)
	}
}

// TestExecuteGetSupportsDataCapabilityFieldSelection verifies typed get plans can
// restrict the returned field selection.
func TestExecuteGetSupportsDataCapabilityFieldSelection(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &protocol.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		DataScope:    1,
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
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataMutationRequest{
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

	planJSON := mustMarshalJSON(t, map[string]any{
		"table":   resource.Table,
		"action":  "get",
		"fields":  []string{"currentState"},
		"keyJson": append([]byte(nil), createResponse.KeyJSON...),
	})
	getResponse, err := ExecuteGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataGetRequest{
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

// TestExecuteBatchGetSupportsFieldSelectionAndMissingKeys verifies batch_get
// uses one key-set query and keeps invisible or absent records opaque.
func TestExecuteBatchGetSupportsFieldSelectionAndMissingKeys(t *testing.T) {
	ctx := context.Background()
	resource := buildTestNodeStateResource()
	identity := &protocol.IdentitySnapshotV1{
		UserID:       1,
		Username:     "admin",
		DataScope:    1,
		IsSuperAdmin: true,
	}
	pluginMarker := "test-datahost-plan-batch-get"
	cleanupNodeStates(t, ctx, pluginMarker)
	t.Cleanup(func() {
		cleanupNodeStates(t, ctx, pluginMarker)
	})

	createResponse, err := ExecuteCreate(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataMutationRequest{
			RecordJSON: mustMarshalJSON(t, map[string]any{
				"pluginId":     pluginMarker,
				"releaseId":    1,
				"nodeKey":      "batch-get-1",
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

	response, err := ExecuteBatchGet(
		ctx,
		"test-plugin-data",
		resource.Table,
		protocol.ExecutionSourceRoute,
		identity,
		nil,
		resource,
		&protocol.HostServiceDataBatchGetRequest{
			KeyJSON: [][]byte{append([]byte(nil), createResponse.KeyJSON...), mustMarshalJSON(t, 999999)},
			Fields:  []string{"currentState"},
		},
	)
	if err != nil {
		t.Fatalf("ExecuteBatchGet failed: %v", err)
	}
	if len(response.Records) != 1 || len(response.MissingKeyJSON) != 1 {
		t.Fatalf("unexpected batch_get response: %#v", response)
	}
	record := mustUnmarshalJSONRecord(t, response.Records[0])
	if len(record) != 1 || record["currentState"] != "pending" {
		t.Fatalf("unexpected selected batch_get record: %#v", record)
	}
	if string(response.MissingKeyJSON[0]) != "999999" {
		t.Fatalf("unexpected missing key JSON: %s", string(response.MissingKeyJSON[0]))
	}
}

// TestPluginDataDBDoCommitRejectsUnauthorizedTable verifies unauthorized table
// writes are rejected by the governed host database wrapper.
func TestPluginDataDBDoCommitRejectsUnauthorizedTable(t *testing.T) {
	db, err := getPluginDataDB()
	if err != nil {
		t.Fatalf("getPluginDataDB failed: %v", err)
	}
	ctx := withPluginDataAudit(context.Background(), &recordstorehost.AuditMetadata{
		PluginID:      "test-plugin-data",
		Table:         "sys_plugin_node_state",
		Method:        protocol.HostServiceMethodDataDelete,
		ResourceTable: "sys_plugin_node_state",
	})
	_, err = db.Ctx(ctx).Exec(ctx, "DELETE FROM sys_plugin WHERE plugin_id = ?", "forbidden")
	if err == nil {
		t.Fatal("expected DoCommit to reject unauthorized table")
	}
	if !strings.Contains(err.Error(), "authorized table") {
		t.Fatalf("expected unauthorized table error, got %v", err)
	}
}

// buildTestNodeStateResource returns a broad CRUD resource contract for sys_plugin_node_state.
func buildTestNodeStateResource() *catalog.ResourceSpec {
	return &catalog.ResourceSpec{
		Key:   "nodeStates",
		Type:  plugintypes.ResourceSpecTypeTableList.String(),
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
			{Param: "pluginId", Column: "plugin_id", Operator: plugintypes.ResourceFilterOperatorEQ.String()},
		},
		OrderBy: catalog.ResourceOrderBySpec{
			Column:    "id",
			Direction: plugintypes.ResourceOrderDirectionASC.String(),
		},
		Operations: []string{
			protocol.HostServiceMethodDataList,
			protocol.HostServiceMethodDataGet,
			protocol.HostServiceMethodDataBatchGet,
			protocol.HostServiceMethodDataCreate,
			protocol.HostServiceMethodDataUpdate,
			protocol.HostServiceMethodDataDelete,
			protocol.HostServiceMethodDataTransaction,
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
		Access: plugintypes.ResourceAccessModeRequest.String(),
	}
}

// buildTestNodeStateResourceWithNodeKey uses nodeKey as the logical key field for tests.
func buildTestNodeStateResourceWithNodeKey() *catalog.ResourceSpec {
	resource := buildTestNodeStateResource()
	resource.KeyField = "nodeKey"
	return resource
}

// cleanupNodeStates removes test rows for the given plugin marker.
func cleanupNodeStates(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete plugin node states for %s: %v", pluginID, err)
	}
}

// dropIdentityContractTable removes the temporary identity table used by
// contract synthesis tests.
func dropIdentityContractTable(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()
	if !testTableIdentifierPattern.MatchString(tableName) {
		t.Fatalf("unsafe identity contract table name: %s", tableName)
	}
	if _, err := g.DB().Exec(ctx, "DROP TABLE IF EXISTS "+tableName); err != nil {
		t.Fatalf("failed to drop identity contract table %s: %v", tableName, err)
	}
}

// createTableContractCacheTable creates a temporary plugin-owned table for
// contract-cache tests.
func createTableContractCacheTable(t *testing.T, ctx context.Context, tableName string, columns string) {
	t.Helper()
	if !testTableIdentifierPattern.MatchString(tableName) {
		t.Fatalf("unsafe contract cache table name: %s", tableName)
	}
	if _, err := g.DB().Exec(ctx, fmt.Sprintf(`
CREATE TABLE %s (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    %s
)`, tableName, columns)); err != nil {
		t.Fatalf("failed to create contract cache table %s: %v", tableName, err)
	}
}

// alterTableContractCacheTable adds one column to a temporary cache test table.
func alterTableContractCacheTable(t *testing.T, ctx context.Context, tableName string, columnDDL string) {
	t.Helper()
	if !testTableIdentifierPattern.MatchString(tableName) {
		t.Fatalf("unsafe contract cache table name: %s", tableName)
	}
	if _, err := g.DB().Exec(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDDL)); err != nil {
		t.Fatalf("failed to alter contract cache table %s: %v", tableName, err)
	}
}

// insertTableContractMigration writes one lifecycle migration row used by the
// table-contract cache key.
func insertTableContractMigration(
	t *testing.T,
	ctx context.Context,
	pluginID string,
	phase plugintypes.MigrationDirection,
	migrationKey string,
	checksum string,
) {
	t.Helper()
	if _, err := dao.SysPluginMigration.Ctx(ctx).Data(do.SysPluginMigration{
		PluginId:       pluginID,
		ReleaseId:      1,
		Phase:          phase.String(),
		MigrationKey:   migrationKey,
		Checksum:       checksum,
		ExecutionOrder: 1,
		Status:         plugintypes.MigrationExecutionStatusSucceeded.String(),
	}).Insert(); err != nil {
		t.Fatalf("failed to insert migration row for %s: %v", pluginID, err)
	}
}

// cleanupTableContractMigrations removes test migration rows for one plugin ID.
func cleanupTableContractMigrations(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to cleanup table contract migrations for %s: %v", pluginID, err)
	}
}

// contractContainsField reports whether a synthesized contract exposes a field.
func contractContainsField(resource *catalog.ResourceSpec, fieldName string) bool {
	if resource == nil {
		return false
	}
	for _, field := range resource.Fields {
		if field != nil && field.Name == fieldName {
			return true
		}
	}
	return false
}

// testTableIdentifierPattern restricts temporary test table names before they
// are embedded in DDL statements.
var testTableIdentifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// containsFieldName reports whether one field list contains target.
func containsFieldName(fields []string, target string) bool {
	for _, field := range fields {
		if field == target {
			return true
		}
	}
	return false
}

// mustMarshalJSON marshals a value and fails the test on error.
func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return data
}

// mustUnmarshalJSONValue decodes arbitrary JSON and fails the test on error.
func mustUnmarshalJSONValue(t *testing.T, data []byte) any {
	t.Helper()
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return value
}

// mustUnmarshalJSONRecord decodes a JSON object and fails the test on error.
func mustUnmarshalJSONRecord(t *testing.T, data []byte) map[string]any {
	t.Helper()
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("json unmarshal record failed: %v", err)
	}
	return record
}
