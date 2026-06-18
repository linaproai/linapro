// Package datahost implements the governed host-side execution layer for
// structured dynamic-plugin data service requests.
package datahost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginbridge/recordstore"
)

// Default pagination limits for governed list operations.
const (
	defaultDataListPageNum  = 1
	defaultDataListPageSize = 10
	maxDataListPageSize     = 100
	maxDataBatchGetKeys     = 100
)

// executionContext stores plugin, table, source, and identity data for one request.
type executionContext struct {
	pluginID        string
	table           string
	executionSource protocol.ExecutionSource
	identity        *protocol.IdentitySnapshotV1
	orgSvc          orgcap.Service
}

// modelProvider abstracts gdb.DB and gdb.TX so mutation helpers can share logic.
type modelProvider interface {
	// Model returns one GoFrame model bound to the requested table or struct.
	Model(tableNameOrStruct ...any) *gdb.Model
}

// ExecuteList executes one governed paged list against an authorized table.
func ExecuteList(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataListRequest,
) (*protocol.HostServiceDataListResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataList); err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataList, false))

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}

	requestPlan, err := decodeDataListPlan(table, request)
	if err != nil {
		return nil, err
	}
	model := buildResourceModel(db, ctx, resource)
	model, err = applyPlanFilters(model, resource, requestPlan.Filters)
	if err != nil {
		return nil, err
	}
	model, err = applyResourceDataScope(ctx, model, resource, identity, orgSvc)
	if err != nil {
		return nil, err
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	response := &protocol.HostServiceDataListResponse{
		Total: int32(total),
	}
	if requestPlan.Action == recordstore.PlanActionCount {
		return response, nil
	}
	fieldArgs, err := buildPlanFieldArgs(resource, requestPlan.Fields)
	if err != nil {
		return nil, err
	}
	orderBy, err := buildPlanOrderBy(resource, requestPlan.Orders)
	if err != nil {
		return nil, err
	}
	page := requestPlan.Page
	records, err := model.
		Fields(fieldArgs...).
		Page(int(page.PageNum), int(page.PageSize)).
		Order(orderBy).
		All()
	if err != nil {
		return nil, err
	}
	response.Records = make([][]byte, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordJSON, marshalErr := json.Marshal(buildResourceRecordWithSelection(record.Map(), resource, requestPlan.Fields))
		if marshalErr != nil {
			return nil, marshalErr
		}
		response.Records = append(response.Records, recordJSON)
	}
	return response, nil
}

// ExecuteGet executes one governed detail lookup against an authorized table.
func ExecuteGet(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataGetRequest,
) (*protocol.HostServiceDataGetResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataGet); err != nil {
		return nil, err
	}
	requestPlan, err := decodeDataGetPlan(table, request)
	if err != nil {
		return nil, err
	}
	keyValue, err := decodeJSONScalar(requestPlan.KeyJSON)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataGet, false))

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}

	model := buildResourceModel(db, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, identity, orgSvc)
	if err != nil {
		return nil, err
	}

	fieldArgs, err := buildPlanFieldArgs(resource, requestPlan.Fields)
	if err != nil {
		return nil, err
	}
	records, err := model.
		Fields(fieldArgs...).
		Limit(1).
		All()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &protocol.HostServiceDataGetResponse{Found: false}, nil
	}
	recordJSON, err := json.Marshal(buildResourceRecordWithSelection(records[0].Map(), resource, requestPlan.Fields))
	if err != nil {
		return nil, err
	}
	return &protocol.HostServiceDataGetResponse{
		Found:      true,
		RecordJSON: recordJSON,
	}, nil
}

// ExecuteBatchGet executes one governed detail batch lookup against an authorized table.
func ExecuteBatchGet(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataBatchGetRequest,
) (*protocol.HostServiceDataBatchGetResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
		orgSvc:          orgSvc,
	}
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataBatchGet); err != nil {
		return nil, err
	}
	keys, keyJSON, err := decodeDataBatchGetKeys(request)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return &protocol.HostServiceDataBatchGetResponse{}, nil
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataBatchGet, false))

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}
	model := buildResourceModel(db, ctx, resource).
		WhereIn(resolveResourceKeyColumn(resource), keys)
	model, err = applyResourceDataScope(ctx, model, resource, identity, orgSvc)
	if err != nil {
		return nil, err
	}
	queryFields := dataBatchGetQueryFields(resource, request.Fields)
	fieldArgs, err := buildPlanFieldArgs(resource, queryFields)
	if err != nil {
		return nil, err
	}
	records, err := model.Fields(fieldArgs...).All()
	if err != nil {
		return nil, err
	}

	keyField := findResourceField(resource, resource.KeyField)
	foundKeys := make(map[string]struct{}, len(records))
	response := &protocol.HostServiceDataBatchGetResponse{
		Records: make([][]byte, 0, len(records)),
	}
	for _, record := range records {
		if record == nil {
			continue
		}
		recordMap := record.Map()
		if keyField != nil {
			foundKeys[canonicalDataBatchKey(resolveResourceRecordValue(recordMap, keyField))] = struct{}{}
		}
		recordJSON, marshalErr := json.Marshal(buildResourceRecordWithSelection(recordMap, resource, request.Fields))
		if marshalErr != nil {
			return nil, marshalErr
		}
		response.Records = append(response.Records, recordJSON)
	}
	for index, key := range keys {
		if _, ok := foundKeys[canonicalDataBatchKey(key)]; ok {
			continue
		}
		response.MissingKeyJSON = append(response.MissingKeyJSON, append([]byte(nil), keyJSON[index]...))
	}
	return response, nil
}

// ExecuteCreate executes one governed record creation against an authorized table.
func ExecuteCreate(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
) (*protocol.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
		orgSvc:          orgSvc,
	}
	return executeCreateWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteUpdate executes one governed record update against an authorized table.
func ExecuteUpdate(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
) (*protocol.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
		orgSvc:          orgSvc,
	}
	return executeUpdateWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteDelete executes one governed record deletion against an authorized table.
func ExecuteDelete(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
) (*protocol.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
		orgSvc:          orgSvc,
	}
	return executeDeleteWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteTransaction executes one governed structured mutation transaction against an authorized table.
func ExecuteTransaction(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource protocol.ExecutionSource,
	identity *protocol.IdentitySnapshotV1,
	orgSvc orgcap.Service,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataTransactionRequest,
) (*protocol.HostServiceDataTransactionResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
		orgSvc:          orgSvc,
	}
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataTransaction); err != nil {
		return nil, err
	}
	if request == nil || len(request.Operations) == 0 {
		return nil, gerror.New("record store transaction requires at least one operation")
	}

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}
	txCtx := withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataTransaction, true))

	response := &protocol.HostServiceDataTransactionResponse{
		Results: make([]*protocol.HostServiceDataMutationResponse, 0, len(request.Operations)),
	}
	err = db.Transaction(txCtx, func(txExecCtx context.Context, tx gdb.TX) error {
		for _, operation := range request.Operations {
			if operation == nil {
				return gerror.New("record store transaction operation cannot be nil")
			}
			switch strings.ToLower(strings.TrimSpace(operation.Method)) {
			case protocol.HostServiceMethodDataCreate:
				result, createErr := executeCreateWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&protocol.HostServiceDataMutationRequest{RecordJSON: append([]byte(nil), operation.RecordJSON...)},
					true,
					tx,
				)
				if createErr != nil {
					return createErr
				}
				response.Results = append(response.Results, result)
				response.AffectedRows += result.AffectedRows
			case protocol.HostServiceMethodDataUpdate:
				result, updateErr := executeUpdateWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&protocol.HostServiceDataMutationRequest{
						KeyJSON:    append([]byte(nil), operation.KeyJSON...),
						RecordJSON: append([]byte(nil), operation.RecordJSON...),
					},
					true,
					tx,
				)
				if updateErr != nil {
					return updateErr
				}
				response.Results = append(response.Results, result)
				response.AffectedRows += result.AffectedRows
			case protocol.HostServiceMethodDataDelete:
				result, deleteErr := executeDeleteWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&protocol.HostServiceDataMutationRequest{KeyJSON: append([]byte(nil), operation.KeyJSON...)},
					true,
					tx,
				)
				if deleteErr != nil {
					return deleteErr
				}
				response.Results = append(response.Results, result)
				response.AffectedRows += result.AffectedRows
			default:
				return gerror.Newf("record store transaction operation is not supported: %s", operation.Method)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// executeCreateWithProvider inserts one governed record using either DB or TX providers.
func executeCreateWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*protocol.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataCreate); err != nil {
		return nil, err
	}
	data, keyValue, err := decodeMutationRecord(resource, request, false)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataCreate, inTransaction))

	if provider == nil {
		db, dbErr := getPluginDataDB()
		if dbErr != nil {
			return nil, dbErr
		}
		provider = db
	}
	result, err := buildResourceModel(provider, ctx, resource).Data(data).Insert()
	if err != nil {
		return nil, err
	}

	response := &protocol.HostServiceDataMutationResponse{AffectedRows: 1}
	if result != nil {
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return nil, rowsErr
		}
		response.AffectedRows = rowsAffected
		if keyValue == nil {
			lastInsertID, lastInsertErr := result.LastInsertId()
			if lastInsertErr != nil {
				return nil, lastInsertErr
			}
			if lastInsertID > 0 {
				keyValue = lastInsertID
			}
		}
	}
	response.KeyJSON, err = encodeJSONValue(keyValue)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// executeUpdateWithProvider updates one governed record using either DB or TX providers.
func executeUpdateWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*protocol.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataUpdate); err != nil {
		return nil, err
	}
	var keyJSON []byte
	if request != nil {
		keyJSON = request.KeyJSON
	}
	keyValue, err := decodeJSONScalar(keyJSON)
	if err != nil {
		return nil, err
	}
	data, _, err := decodeMutationRecord(resource, request, true)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataUpdate, inTransaction))

	if provider == nil {
		db, dbErr := getPluginDataDB()
		if dbErr != nil {
			return nil, dbErr
		}
		provider = db
	}
	model := buildResourceModel(provider, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, execCtx.identity, execCtx.orgSvc)
	if err != nil {
		return nil, err
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return nil, err
	}

	response := &protocol.HostServiceDataMutationResponse{}
	if result != nil {
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return nil, rowsErr
		}
		response.AffectedRows = rowsAffected
	}
	response.KeyJSON, err = encodeJSONValue(keyValue)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// executeDeleteWithProvider deletes one governed record using either DB or TX providers.
func executeDeleteWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*protocol.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, protocol.HostServiceMethodDataDelete); err != nil {
		return nil, err
	}
	var keyJSON []byte
	if request != nil {
		keyJSON = request.KeyJSON
	}
	keyValue, err := decodeJSONScalar(keyJSON)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, protocol.HostServiceMethodDataDelete, inTransaction))

	if provider == nil {
		db, dbErr := getPluginDataDB()
		if dbErr != nil {
			return nil, dbErr
		}
		provider = db
	}
	model := buildResourceModel(provider, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, execCtx.identity, execCtx.orgSvc)
	if err != nil {
		return nil, err
	}
	result, err := model.Delete()
	if err != nil {
		return nil, err
	}

	response := &protocol.HostServiceDataMutationResponse{}
	if result != nil {
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return nil, rowsErr
		}
		response.AffectedRows = rowsAffected
	}
	response.KeyJSON, err = encodeJSONValue(keyValue)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// validateExecutionAccess enforces declared operations and access mode before
// executing one governed data-table request.
func validateExecutionAccess(execCtx *executionContext, resource *catalog.ResourceSpec, method string) error {
	if execCtx == nil {
		return gerror.New("data service execution context is required")
	}
	if resource == nil {
		return gerror.New("data service table contract is required")
	}
	if !resourceAllowsOperation(resource, method) {
		return gerror.Newf("data table %s does not authorize method %s", resource.Table, method)
	}
	// Access mode combines the declared table contract with the current trigger
	// source and identity snapshot so request-bound tables cannot be reused by
	// anonymous or background execution paths.
	normalizedSource := protocol.NormalizeExecutionSource(execCtx.executionSource)
	switch plugintypes.NormalizeResourceAccessMode(resource.Access) {
	case plugintypes.ResourceAccessModeRequest:
		if normalizedSource != protocol.ExecutionSourceRoute {
			return gerror.Newf("data table %s only allows request-bound context", resource.Table)
		}
		if execCtx.identity == nil || execCtx.identity.UserID <= 0 {
			return gerror.Newf("data table %s requires authenticated user context", resource.Table)
		}
	case plugintypes.ResourceAccessModeSystem:
		return nil
	case plugintypes.ResourceAccessModeBoth:
		if normalizedSource == protocol.ExecutionSourceRoute && (execCtx.identity == nil || execCtx.identity.UserID <= 0) {
			return gerror.Newf("data table %s requires authenticated user in request-bound context", resource.Table)
		}
	default:
		return gerror.Newf("data table %s access configuration is invalid", resource.Table)
	}
	return nil
}

// buildResourceModel builds the base safe model for the governed table resource.
func buildResourceModel(provider modelProvider, ctx context.Context, resource *catalog.ResourceSpec) *gdb.Model {
	return provider.Model(resource.Table).Safe().Ctx(ctx)
}

// buildResourceFieldArgs builds aliased select expressions for all declared fields.
func buildResourceFieldArgs(resource *catalog.ResourceSpec) []any {
	fields := make([]any, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		if field == nil {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s AS %s", field.Column, quoteResourceAlias(field.Name)))
	}
	return fields
}

// buildResourceOrderBy builds the fallback ORDER BY expression from the resource spec.
func buildResourceOrderBy(resource *catalog.ResourceSpec) string {
	if resource == nil {
		return ""
	}
	orderBy := strings.TrimSpace(resource.OrderBy.Column)
	if orderBy == "" {
		return ""
	}
	if plugintypes.NormalizeResourceOrderDirection(resource.OrderBy.Direction) == plugintypes.ResourceOrderDirectionDESC {
		return orderBy + " DESC"
	}
	return orderBy + " ASC"
}

// buildResourceRecord projects one database row into the logical resource field map.
func buildResourceRecord(recordMap map[string]interface{}, resource *catalog.ResourceSpec) map[string]interface{} {
	if len(recordMap) == 0 || resource == nil {
		return map[string]interface{}{}
	}
	row := make(map[string]interface{}, len(resource.Fields))
	for _, field := range resource.Fields {
		if field == nil {
			continue
		}
		row[field.Name] = normalizeResourceValue(resolveResourceRecordValue(recordMap, field))
	}
	return row
}

// quoteResourceAlias preserves logical camelCase field names across database
// drivers that fold unquoted aliases, such as PostgreSQL.
func quoteResourceAlias(alias string) string {
	return `"` + strings.ReplaceAll(alias, `"`, `""`) + `"`
}

// resolveResourceRecordValue reads projected rows from either the logical alias
// or the physical column name so live-schema table contracts work across drivers.
func resolveResourceRecordValue(recordMap map[string]interface{}, field *catalog.ResourceField) interface{} {
	if recordMap == nil || field == nil {
		return nil
	}
	if value, ok := recordMap[field.Name]; ok {
		return value
	}
	if value, ok := recordMap[field.Column]; ok {
		return value
	}
	return nil
}

// decodeMutationRecord decodes and validates one mutation payload against the
// resource writable-field contract.
func decodeMutationRecord(
	resource *catalog.ResourceSpec,
	request *protocol.HostServiceDataMutationRequest,
	forUpdate bool,
) (map[string]interface{}, interface{}, error) {
	var recordJSON []byte
	if request != nil {
		recordJSON = request.RecordJSON
	}
	record, err := decodeJSONObject(recordJSON)
	if err != nil {
		return nil, nil, err
	}
	if len(record) == 0 {
		return nil, nil, gerror.New("data mutation record cannot be empty")
	}

	data := make(map[string]interface{}, len(resource.WritableFields))
	var keyValue interface{}
	for _, writableField := range resource.WritableFields {
		value, ok := record[writableField]
		if !ok {
			continue
		}
		if forUpdate && writableField == resource.KeyField {
			return nil, nil, gerror.Newf("data update cannot modify keyField: %s", resource.KeyField)
		}
		column := resolveResourceFieldColumn(resource, writableField)
		if column == "" {
			return nil, nil, gerror.Newf("data table writableField is not mapped to a column: %s", writableField)
		}
		data[column] = value
		if writableField == resource.KeyField {
			keyValue = value
		}
	}
	if len(data) == 0 {
		return nil, nil, gerror.New("data mutation record contains no writable fields")
	}
	for fieldName := range record {
		if !resourceAllowsWritableField(resource, fieldName) {
			return nil, nil, gerror.Newf("data mutation field is not authorized: %s", fieldName)
		}
	}
	return data, keyValue, nil
}

// resolveResourceKeyColumn returns the physical column bound to the logical key field.
func resolveResourceKeyColumn(resource *catalog.ResourceSpec) string {
	return resolveResourceFieldColumn(resource, resource.KeyField)
}

// resolveResourceFieldColumn maps one logical field name to its table column.
func resolveResourceFieldColumn(resource *catalog.ResourceSpec, fieldName string) string {
	if resource == nil {
		return ""
	}
	targetFieldName := strings.TrimSpace(fieldName)
	for _, field := range resource.Fields {
		if field != nil && field.Name == targetFieldName {
			return field.Column
		}
	}
	return ""
}

// resourceAllowsWritableField reports whether the logical field is writable.
func resourceAllowsWritableField(resource *catalog.ResourceSpec, fieldName string) bool {
	if resource == nil {
		return false
	}
	targetFieldName := strings.TrimSpace(fieldName)
	for _, writableField := range resource.WritableFields {
		if writableField == targetFieldName {
			return true
		}
	}
	return false
}

// resourceAllowsOperation reports whether the resource authorizes the method.
func resourceAllowsOperation(resource *catalog.ResourceSpec, method string) bool {
	if resource == nil {
		return false
	}
	targetMethod := strings.ToLower(strings.TrimSpace(method))
	for _, operation := range resource.Operations {
		if operation == targetMethod {
			return true
		}
	}
	return false
}

// decodeJSONObject decodes a JSON object payload into a generic map.
func decodeJSONObject(content []byte) (map[string]interface{}, error) {
	if len(content) == 0 {
		return nil, nil
	}
	result := make(map[string]interface{})
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// decodeJSONScalar decodes a required scalar key payload.
func decodeJSONScalar(content []byte) (interface{}, error) {
	if len(content) == 0 {
		return nil, gerror.New("record store key cannot be empty")
	}
	var value interface{}
	if err := json.Unmarshal(content, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, gerror.New("record store key cannot be empty")
	}
	return value, nil
}

func decodeDataBatchGetKeys(request *protocol.HostServiceDataBatchGetRequest) ([]interface{}, [][]byte, error) {
	if request == nil || len(request.KeyJSON) == 0 {
		return nil, nil, gerror.New("record store batch_get keys cannot be empty")
	}
	if len(request.KeyJSON) > maxDataBatchGetKeys {
		return nil, nil, gerror.Newf("record store batch_get key count exceeds limit %d", maxDataBatchGetKeys)
	}
	keys := make([]interface{}, 0, len(request.KeyJSON))
	keyJSON := make([][]byte, 0, len(request.KeyJSON))
	seen := make(map[string]struct{}, len(request.KeyJSON))
	for _, rawKeyJSON := range request.KeyJSON {
		key, err := decodeJSONScalar(rawKeyJSON)
		if err != nil {
			return nil, nil, err
		}
		canonicalKey := canonicalDataBatchKey(key)
		if canonicalKey == "" {
			return nil, nil, gerror.New("record store batch_get key cannot be empty")
		}
		if _, ok := seen[canonicalKey]; ok {
			continue
		}
		seen[canonicalKey] = struct{}{}
		keys = append(keys, key)
		keyJSON = append(keyJSON, append([]byte(nil), rawKeyJSON...))
	}
	return keys, keyJSON, nil
}

func canonicalDataBatchKey(value interface{}) string {
	if value == nil {
		return ""
	}
	content, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(content)
}

func dataBatchGetQueryFields(resource *catalog.ResourceSpec, selected []string) []string {
	if len(selected) == 0 || resource == nil || strings.TrimSpace(resource.KeyField) == "" {
		return selected
	}
	fields := make([]string, 0, len(selected)+1)
	hasKey := false
	for _, field := range selected {
		normalized := strings.TrimSpace(field)
		if normalized == "" {
			fields = append(fields, field)
			continue
		}
		if normalized == resource.KeyField {
			hasKey = true
		}
		fields = append(fields, normalized)
	}
	if !hasKey {
		fields = append(fields, resource.KeyField)
	}
	return fields
}

// encodeJSONValue marshals one optional scalar or structured response value.
func encodeJSONValue(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

// normalizeResourceValue converts time values into JSON-safe strings.
func normalizeResourceValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case *time.Time:
		if typedValue == nil {
			return ""
		}
		return typedValue.Format(time.RFC3339Nano)
	case time.Time:
		return typedValue.Format(time.RFC3339Nano)
	case interface{ String() string }:
		return typedValue.String()
	default:
		return value
	}
}
