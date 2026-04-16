// Package datahost implements the governed host-side execution layer for
// structured dynamic-plugin data service requests.
package datahost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/plugindb/shared"
)

const (
	defaultDataListPageNum  = 1
	defaultDataListPageSize = 10
	maxDataListPageSize     = 100
)

type executionContext struct {
	pluginID        string
	table           string
	executionSource pluginbridge.ExecutionSource
	identity        *pluginbridge.IdentitySnapshotV1
}

type modelProvider interface {
	Model(tableNameOrStruct ...any) *gdb.Model
}

// ExecuteList executes one governed paged list against an authorized table.
func ExecuteList(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataListRequest,
) (*pluginbridge.HostServiceDataListResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataList); err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataList, false))

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}

	plan, err := decodeDataListPlan(table, request)
	if err != nil {
		return nil, err
	}
	model := buildResourceModel(db, ctx, resource)
	model, err = applyPlanFilters(model, resource, plan.Filters)
	if err != nil {
		return nil, err
	}
	model, err = applyResourceDataScope(ctx, model, resource, identity)
	if err != nil {
		return nil, err
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	response := &pluginbridge.HostServiceDataListResponse{
		Total: int32(total),
	}
	if plan.Action == shared.DataPlanActionCount {
		return response, nil
	}
	fieldArgs, err := buildPlanFieldArgs(resource, plan.Fields)
	if err != nil {
		return nil, err
	}
	orderBy, err := buildPlanOrderBy(resource, plan.Orders)
	if err != nil {
		return nil, err
	}
	page := plan.Page
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
		recordJSON, marshalErr := json.Marshal(buildResourceRecordWithSelection(record.Map(), resource, plan.Fields))
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
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataGetRequest,
) (*pluginbridge.HostServiceDataGetResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataGet); err != nil {
		return nil, err
	}
	plan, err := decodeDataGetPlan(table, request)
	if err != nil {
		return nil, err
	}
	keyValue, err := decodeJSONScalar(plan.KeyJSON)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataGet, false))

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}

	model := buildResourceModel(db, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, identity)
	if err != nil {
		return nil, err
	}

	fieldArgs, err := buildPlanFieldArgs(resource, plan.Fields)
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
		return &pluginbridge.HostServiceDataGetResponse{Found: false}, nil
	}
	recordJSON, err := json.Marshal(buildResourceRecordWithSelection(records[0].Map(), resource, plan.Fields))
	if err != nil {
		return nil, err
	}
	return &pluginbridge.HostServiceDataGetResponse{
		Found:      true,
		RecordJSON: recordJSON,
	}, nil
}

// ExecuteCreate executes one governed record creation against an authorized table.
func ExecuteCreate(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	return executeCreateWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteUpdate executes one governed record update against an authorized table.
func ExecuteUpdate(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	return executeUpdateWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteDelete executes one governed record deletion against an authorized table.
func ExecuteDelete(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	return executeDeleteWithProvider(ctx, execCtx, resource, request, false, nil)
}

// ExecuteTransaction executes one governed structured mutation transaction against an authorized table.
func ExecuteTransaction(
	ctx context.Context,
	pluginID string,
	table string,
	executionSource pluginbridge.ExecutionSource,
	identity *pluginbridge.IdentitySnapshotV1,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataTransactionRequest,
) (*pluginbridge.HostServiceDataTransactionResponse, error) {
	execCtx := &executionContext{
		pluginID:        pluginID,
		table:           table,
		executionSource: executionSource,
		identity:        identity,
	}
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataTransaction); err != nil {
		return nil, err
	}
	if request == nil || len(request.Operations) == 0 {
		return nil, gerror.New("data transaction 至少需要一个操作")
	}

	db, err := getPluginDataDB()
	if err != nil {
		return nil, err
	}
	txCtx := withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataTransaction, true))

	response := &pluginbridge.HostServiceDataTransactionResponse{
		Results: make([]*pluginbridge.HostServiceDataMutationResponse, 0, len(request.Operations)),
	}
	err = db.Transaction(txCtx, func(txExecCtx context.Context, tx gdb.TX) error {
		for _, operation := range request.Operations {
			if operation == nil {
				return gerror.New("data transaction 操作不能为空")
			}
			switch strings.ToLower(strings.TrimSpace(operation.Method)) {
			case pluginbridge.HostServiceMethodDataCreate:
				result, createErr := executeCreateWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&pluginbridge.HostServiceDataMutationRequest{RecordJSON: append([]byte(nil), operation.RecordJSON...)},
					true,
					tx,
				)
				if createErr != nil {
					return createErr
				}
				response.Results = append(response.Results, result)
				response.AffectedRows += result.AffectedRows
			case pluginbridge.HostServiceMethodDataUpdate:
				result, updateErr := executeUpdateWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&pluginbridge.HostServiceDataMutationRequest{
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
			case pluginbridge.HostServiceMethodDataDelete:
				result, deleteErr := executeDeleteWithProvider(
					txExecCtx,
					execCtx,
					resource,
					&pluginbridge.HostServiceDataMutationRequest{KeyJSON: append([]byte(nil), operation.KeyJSON...)},
					true,
					tx,
				)
				if deleteErr != nil {
					return deleteErr
				}
				response.Results = append(response.Results, result)
				response.AffectedRows += result.AffectedRows
			default:
				return gerror.Newf("data transaction 不支持操作: %s", operation.Method)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func executeCreateWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataCreate); err != nil {
		return nil, err
	}
	data, keyValue, err := decodeMutationRecord(resource, request, false)
	if err != nil {
		return nil, err
	}
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataCreate, inTransaction))

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

	response := &pluginbridge.HostServiceDataMutationResponse{AffectedRows: 1}
	if result != nil {
		if rowsAffected, rowsErr := result.RowsAffected(); rowsErr == nil {
			response.AffectedRows = rowsAffected
		}
		if keyValue == nil {
			if lastInsertID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil && lastInsertID > 0 {
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

func executeUpdateWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataUpdate); err != nil {
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
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataUpdate, inTransaction))

	if provider == nil {
		db, dbErr := getPluginDataDB()
		if dbErr != nil {
			return nil, dbErr
		}
		provider = db
	}
	model := buildResourceModel(provider, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, execCtx.identity)
	if err != nil {
		return nil, err
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return nil, err
	}

	response := &pluginbridge.HostServiceDataMutationResponse{}
	if result != nil {
		if rowsAffected, rowsErr := result.RowsAffected(); rowsErr == nil {
			response.AffectedRows = rowsAffected
		}
	}
	response.KeyJSON, err = encodeJSONValue(keyValue)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func executeDeleteWithProvider(
	ctx context.Context,
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
	inTransaction bool,
	provider modelProvider,
) (*pluginbridge.HostServiceDataMutationResponse, error) {
	if err := validateExecutionAccess(execCtx, resource, pluginbridge.HostServiceMethodDataDelete); err != nil {
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
	ctx = withPluginDataAudit(ctx, buildPluginDataAuditMetadata(execCtx, resource, pluginbridge.HostServiceMethodDataDelete, inTransaction))

	if provider == nil {
		db, dbErr := getPluginDataDB()
		if dbErr != nil {
			return nil, dbErr
		}
		provider = db
	}
	model := buildResourceModel(provider, ctx, resource).
		Where(resolveResourceKeyColumn(resource), keyValue)
	model, err = applyResourceDataScope(ctx, model, resource, execCtx.identity)
	if err != nil {
		return nil, err
	}
	result, err := model.Delete()
	if err != nil {
		return nil, err
	}

	response := &pluginbridge.HostServiceDataMutationResponse{}
	if result != nil {
		if rowsAffected, rowsErr := result.RowsAffected(); rowsErr == nil {
			response.AffectedRows = rowsAffected
		}
	}
	response.KeyJSON, err = encodeJSONValue(keyValue)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func validateExecutionAccess(execCtx *executionContext, resource *catalog.ResourceSpec, method string) error {
	if execCtx == nil {
		return gerror.New("data service execution context is required")
	}
	if resource == nil {
		return gerror.New("data service table contract is required")
	}
	if !resourceAllowsOperation(resource, method) {
		return gerror.Newf("data table %s 未授权方法 %s", resource.Table, method)
	}
	// Access mode combines the declared table contract with the current trigger
	// source and identity snapshot so request-bound tables cannot be reused by
	// anonymous or background execution paths.
	normalizedSource := pluginbridge.NormalizeExecutionSource(execCtx.executionSource)
	switch catalog.NormalizeResourceAccessMode(resource.Access) {
	case catalog.ResourceAccessModeRequest:
		if normalizedSource != pluginbridge.ExecutionSourceRoute {
			return gerror.Newf("data table %s 仅允许请求型上下文", resource.Table)
		}
		if execCtx.identity == nil || execCtx.identity.UserID <= 0 {
			return gerror.Newf("data table %s 要求登录用户上下文", resource.Table)
		}
	case catalog.ResourceAccessModeSystem:
		return nil
	case catalog.ResourceAccessModeBoth:
		if normalizedSource == pluginbridge.ExecutionSourceRoute && (execCtx.identity == nil || execCtx.identity.UserID <= 0) {
			return gerror.Newf("data table %s 在请求型上下文要求登录用户", resource.Table)
		}
	default:
		return gerror.Newf("data table %s access 配置不合法", resource.Table)
	}
	return nil
}

func normalizeDataListRequest(request *pluginbridge.HostServiceDataListRequest) *pluginbridge.HostServiceDataListRequest {
	if request == nil {
		request = &pluginbridge.HostServiceDataListRequest{}
	}
	if request.PageNum <= 0 {
		request.PageNum = defaultDataListPageNum
	}
	if request.PageSize <= 0 {
		request.PageSize = defaultDataListPageSize
	}
	if request.PageSize > maxDataListPageSize {
		request.PageSize = maxDataListPageSize
	}
	return request
}

func buildResourceModel(provider modelProvider, ctx context.Context, resource *catalog.ResourceSpec) *gdb.Model {
	return provider.Model(resource.Table).Safe().Ctx(ctx)
}

func applyDeclaredFilters(model *gdb.Model, resource *catalog.ResourceSpec, filters map[string]string) (*gdb.Model, error) {
	if model == nil || resource == nil || len(filters) == 0 {
		return model, nil
	}

	declaredFilters := make(map[string]*catalog.ResourceQuery, len(resource.Filters))
	for _, filter := range resource.Filters {
		if filter == nil {
			continue
		}
		declaredFilters[filter.Param] = filter
	}
	for param := range filters {
		if _, ok := declaredFilters[param]; !ok {
			return nil, gerror.Newf("data list filter 未声明: %s", param)
		}
	}
	for _, filter := range resource.Filters {
		if filter == nil {
			continue
		}
		value := strings.TrimSpace(filters[filter.Param])
		if value == "" {
			continue
		}
		switch catalog.NormalizeResourceFilterOperator(filter.Operator) {
		case catalog.ResourceFilterOperatorEQ:
			model = model.Where(filter.Column, value)
		case catalog.ResourceFilterOperatorLike:
			model = model.WhereLike(filter.Column, "%"+value+"%")
		case catalog.ResourceFilterOperatorGTEDate:
			model = model.WhereGTE(filter.Column, value+" 00:00:00")
		case catalog.ResourceFilterOperatorLTEDate:
			model = model.WhereLTE(filter.Column, value+" 23:59:59")
		default:
			return nil, gerror.Newf("data list filter operator 不支持: %s", filter.Operator)
		}
	}
	return model, nil
}

func buildResourceFieldArgs(resource *catalog.ResourceSpec) []any {
	fields := make([]any, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		if field == nil {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s AS %s", field.Column, field.Name))
	}
	return fields
}

func buildResourceOrderBy(resource *catalog.ResourceSpec) string {
	if resource == nil {
		return ""
	}
	orderBy := strings.TrimSpace(resource.OrderBy.Column)
	if orderBy == "" {
		return ""
	}
	if catalog.NormalizeResourceOrderDirection(resource.OrderBy.Direction) == catalog.ResourceOrderDirectionDESC {
		return orderBy + " DESC"
	}
	return orderBy + " ASC"
}

func buildResourceRecord(recordMap map[string]interface{}, resource *catalog.ResourceSpec) map[string]interface{} {
	if len(recordMap) == 0 || resource == nil {
		return map[string]interface{}{}
	}
	row := make(map[string]interface{}, len(resource.Fields))
	for _, field := range resource.Fields {
		if field == nil {
			continue
		}
		row[field.Name] = normalizeResourceValue(recordMap[field.Name])
	}
	return row
}

func decodeMutationRecord(
	resource *catalog.ResourceSpec,
	request *pluginbridge.HostServiceDataMutationRequest,
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
		return nil, nil, gerror.New("data mutation record 不能为空")
	}

	data := make(map[string]interface{}, len(resource.WritableFields))
	var keyValue interface{}
	for _, writableField := range resource.WritableFields {
		value, ok := record[writableField]
		if !ok {
			continue
		}
		if forUpdate && writableField == resource.KeyField {
			return nil, nil, gerror.Newf("data update 不允许修改 keyField: %s", resource.KeyField)
		}
		column := resolveResourceFieldColumn(resource, writableField)
		if column == "" {
			return nil, nil, gerror.Newf("data table writableField 未映射字段: %s", writableField)
		}
		data[column] = value
		if writableField == resource.KeyField {
			keyValue = value
		}
	}
	if len(data) == 0 {
		return nil, nil, gerror.New("data mutation record 不包含可写字段")
	}
	for fieldName := range record {
		if !resourceAllowsWritableField(resource, fieldName) {
			return nil, nil, gerror.Newf("data mutation 字段未授权: %s", fieldName)
		}
	}
	return data, keyValue, nil
}

func resolveResourceKeyColumn(resource *catalog.ResourceSpec) string {
	return resolveResourceFieldColumn(resource, resource.KeyField)
}

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

func decodeJSONScalar(content []byte) (interface{}, error) {
	if len(content) == 0 {
		return nil, gerror.New("data key 不能为空")
	}
	var value interface{}
	if err := json.Unmarshal(content, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, gerror.New("data key 不能为空")
	}
	return value, nil
}

func encodeJSONValue(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func normalizeResourceValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case *gtime.Time:
		if typedValue == nil {
			return ""
		}
		return typedValue.String()
	case gtime.Time:
		return typedValue.String()
	default:
		return value
	}
}
