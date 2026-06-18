//go:build wasip1

// This file implements the governed record store capability execution path for wasm guests.

package recordstore

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
	dataplan "lina-core/pkg/plugin/pluginbridge/recordstore/internal/plan"
)

// ensureExecutionReady validates the accumulated query plan before one governed
// guest execution call.
func (q *Query) ensureExecutionReady(action dataplan.PlanAction) error {
	if q == nil {
		return gerror.New("record store capability query is nil")
	}
	if q.err != nil {
		return q.err
	}
	if strings.TrimSpace(q.table) == "" {
		return gerror.New("record store capability table cannot be empty")
	}
	for _, filter := range q.plan.Filters {
		if err := dataplan.ValidateFilter(filter); err != nil {
			return err
		}
	}
	for _, order := range q.plan.Orders {
		if err := dataplan.ValidateOrder(order); err != nil {
			return err
		}
	}
	q.plan.Action = action
	if err := dataplan.ValidateQueryPlan(q.plan); err != nil {
		return err
	}
	return nil
}

// One executes one governed single-record lookup.
func (q *Query) One() (map[string]any, bool, error) {
	if err := q.ensureExecutionReady(dataplan.PlanActionGet); err != nil {
		return nil, false, err
	}
	if len(q.plan.KeyJSON) > 0 {
		planJSON, err := dataplan.MarshalQueryPlanJSON(q.plan)
		if err != nil {
			return nil, false, err
		}
		responsePayload, err := invokeDataHostServiceGet(q.invoker, q.table, &protocol.HostServiceDataGetRequest{
			PlanJSON: planJSON,
		})
		if err != nil {
			return nil, false, err
		}
		record, err := decodeJSONRecord(responsePayload.RecordJSON)
		if err != nil {
			return nil, false, err
		}
		return record, responsePayload.Found, nil
	}
	records, _, err := q.All()
	if err != nil {
		return nil, false, err
	}
	if len(records) == 0 {
		return nil, false, nil
	}
	return records[0], true, nil
}

// BatchGet executes one governed multi-record lookup by primary keys.
func (q *Query) BatchGet(keys []any) ([]map[string]any, [][]byte, error) {
	if q == nil {
		return nil, nil, gerror.New("record store capability query is nil")
	}
	if q.err != nil {
		return nil, nil, q.err
	}
	if strings.TrimSpace(q.table) == "" {
		return nil, nil, gerror.New("record store capability table cannot be empty")
	}
	keyJSON := make([][]byte, 0, len(keys))
	for _, key := range keys {
		encodedKey, err := marshalJSONValue(key)
		if err != nil {
			return nil, nil, err
		}
		keyJSON = append(keyJSON, encodedKey)
	}
	responsePayload, err := invokeDataHostServiceBatchGet(q.invoker, q.table, &protocol.HostServiceDataBatchGetRequest{
		KeyJSON: keyJSON,
		Fields:  append([]string(nil), q.plan.Fields...),
	})
	if err != nil {
		return nil, nil, err
	}
	records, err := decodeJSONRecordList(responsePayload.Records)
	if err != nil {
		return nil, nil, err
	}
	return records, responsePayload.MissingKeyJSON, nil
}

// All executes one governed paged list query.
func (q *Query) All() ([]map[string]any, int32, error) {
	if err := q.ensureExecutionReady(dataplan.PlanActionList); err != nil {
		return nil, 0, err
	}
	planJSON, err := dataplan.MarshalQueryPlanJSON(q.plan)
	if err != nil {
		return nil, 0, err
	}
	result, err := invokeDataHostServiceList(q.invoker, q.table, &protocol.HostServiceDataListRequest{
		PlanJSON: planJSON,
	})
	if err != nil {
		return nil, 0, err
	}
	if result == nil {
		return nil, 0, nil
	}
	records, err := decodeJSONRecordList(result.Records)
	if err != nil {
		return nil, 0, err
	}
	return records, result.Total, nil
}

// Count executes one governed count query.
func (q *Query) Count() (int32, error) {
	if q == nil {
		return 0, gerror.New("record store capability query is nil")
	}
	if err := q.ensureExecutionReady(dataplan.PlanActionCount); err != nil {
		return 0, err
	}
	planJSON, err := dataplan.MarshalQueryPlanJSON(q.plan)
	if err != nil {
		return 0, err
	}
	result, err := invokeDataHostServiceList(q.invoker, q.table, &protocol.HostServiceDataListRequest{
		PlanJSON: planJSON,
	})
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, nil
	}
	return result.Total, nil
}

// Insert executes one governed insert mutation.
func (q *Query) Insert(record map[string]any) (*MutationResult, error) {
	if err := q.ensureExecutionReady(dataplan.PlanActionCreate); err != nil {
		return nil, err
	}
	result, err := invokeDataHostServiceMutation(q.invoker, q.table, protocol.HostServiceMethodDataCreate, nil, record)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result)
}

// Update executes one governed update mutation.
func (q *Query) Update(record map[string]any) (*MutationResult, error) {
	if err := q.ensureExecutionReady(dataplan.PlanActionUpdate); err != nil {
		return nil, err
	}
	if len(q.plan.KeyJSON) == 0 {
		return nil, gerror.New("record store capability update requires WhereKey")
	}
	key, err := dataplan.UnmarshalValueJSON(q.plan.KeyJSON)
	if err != nil {
		return nil, err
	}
	result, err := invokeDataHostServiceMutation(q.invoker, q.table, protocol.HostServiceMethodDataUpdate, key, record)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result)
}

// Delete executes one governed delete mutation.
func (q *Query) Delete() (*MutationResult, error) {
	if err := q.ensureExecutionReady(dataplan.PlanActionDelete); err != nil {
		return nil, err
	}
	if len(q.plan.KeyJSON) == 0 {
		return nil, gerror.New("record store capability delete requires WhereKey")
	}
	key, err := dataplan.UnmarshalValueJSON(q.plan.KeyJSON)
	if err != nil {
		return nil, err
	}
	result, err := invokeDataHostServiceMutation(q.invoker, q.table, protocol.HostServiceMethodDataDelete, key, nil)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result)
}

// Transaction executes one governed structured mutation transaction.
func (db *DB) Transaction(fn func(tx *Tx) error) error {
	if fn == nil {
		return gerror.New("record store capability transaction callback cannot be nil")
	}
	var invoker HostServiceInvoker
	if db != nil {
		invoker = db.invoker
	}
	tx := &Tx{invoker: invoker}
	if err := fn(tx); err != nil {
		return err
	}
	if tx.err != nil {
		return tx.err
	}
	if strings.TrimSpace(tx.table) == "" {
		return gerror.New("record store capability transaction table cannot be empty")
	}
	operations := make([]*protocol.HostServiceDataTransactionOperation, 0, len(tx.operations))
	for _, operation := range tx.operations {
		if operation == nil {
			continue
		}
		key, err := dataplan.UnmarshalValueJSON(operation.KeyJSON)
		if err != nil {
			return err
		}
		recordValue, err := dataplan.UnmarshalValueJSON(operation.RecordJSON)
		if err != nil {
			return err
		}
		record, _ := recordValue.(map[string]any)
		keyJSON, err := marshalJSONValue(key)
		if err != nil {
			return err
		}
		recordJSON, err := marshalJSONValue(record)
		if err != nil {
			return err
		}
		operations = append(operations, &protocol.HostServiceDataTransactionOperation{
			Method:     operation.Action.String(),
			KeyJSON:    keyJSON,
			RecordJSON: recordJSON,
		})
	}
	_, err := invokeDataHostServiceTransaction(tx.invoker, tx.table, &protocol.HostServiceDataTransactionRequest{
		Operations: operations,
	})
	return err
}

// decodeMutationResult maps the host bridge mutation result into the guest
// facade result type.
func decodeMutationResult(result *protocol.HostServiceDataMutationResponse) (*MutationResult, error) {
	if result == nil {
		return &MutationResult{}, nil
	}
	key, err := dataplan.UnmarshalValueJSON(result.KeyJSON)
	if err != nil {
		return nil, err
	}
	recordValue, err := dataplan.UnmarshalValueJSON(result.RecordJSON)
	if err != nil {
		return nil, err
	}
	record, _ := recordValue.(map[string]any)
	return &MutationResult{AffectedRows: result.AffectedRows, Key: key, Record: record}, nil
}

// invokeDataHostServiceList dispatches one governed record store list request
// through the structured data host-service protocol.
func invokeDataHostServiceList(
	invoker HostServiceInvoker,
	table string,
	request *protocol.HostServiceDataListRequest,
) (*protocol.HostServiceDataListResponse, error) {
	payload, err := invokeRecordStoreHostService(
		invoker,
		protocol.HostServiceData,
		protocol.HostServiceMethodDataList,
		"",
		table,
		protocol.MarshalHostServiceDataListRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceDataListResponse(payload)
}

// invokeDataHostServiceGet dispatches one governed record store detail request
// through the structured data host-service protocol.
func invokeDataHostServiceGet(
	invoker HostServiceInvoker,
	table string,
	request *protocol.HostServiceDataGetRequest,
) (*protocol.HostServiceDataGetResponse, error) {
	payload, err := invokeRecordStoreHostService(
		invoker,
		protocol.HostServiceData,
		protocol.HostServiceMethodDataGet,
		"",
		table,
		protocol.MarshalHostServiceDataGetRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceDataGetResponse(payload)
}

// invokeDataHostServiceBatchGet dispatches one governed record store batch_get
// request through the structured data host-service protocol.
func invokeDataHostServiceBatchGet(
	invoker HostServiceInvoker,
	table string,
	request *protocol.HostServiceDataBatchGetRequest,
) (*protocol.HostServiceDataBatchGetResponse, error) {
	payload, err := invokeRecordStoreHostService(
		invoker,
		protocol.HostServiceData,
		protocol.HostServiceMethodDataBatchGet,
		"",
		table,
		protocol.MarshalHostServiceDataBatchGetRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceDataBatchGetResponse(payload)
}

// invokeDataHostServiceMutation dispatches one governed record store mutation
// request through the structured data host-service protocol.
func invokeDataHostServiceMutation(
	invoker HostServiceInvoker,
	table string,
	method string,
	key any,
	record map[string]any,
) (*protocol.HostServiceDataMutationResponse, error) {
	keyJSON, err := marshalJSONValue(key)
	if err != nil {
		return nil, err
	}
	recordJSON, err := marshalJSONValue(record)
	if err != nil {
		return nil, err
	}
	payload, err := invokeRecordStoreHostService(
		invoker,
		protocol.HostServiceData,
		method,
		"",
		table,
		protocol.MarshalHostServiceDataMutationRequest(&protocol.HostServiceDataMutationRequest{
			KeyJSON:    keyJSON,
			RecordJSON: recordJSON,
		}),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceDataMutationResponse(payload)
}

// invokeDataHostServiceTransaction dispatches one governed record store
// transaction request through the structured data host-service protocol.
func invokeDataHostServiceTransaction(
	invoker HostServiceInvoker,
	table string,
	request *protocol.HostServiceDataTransactionRequest,
) (*protocol.HostServiceDataTransactionResponse, error) {
	payload, err := invokeRecordStoreHostService(
		invoker,
		protocol.HostServiceData,
		protocol.HostServiceMethodDataTransaction,
		"",
		table,
		protocol.MarshalHostServiceDataTransactionRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceDataTransactionResponse(payload)
}

// invokeRecordStoreHostService dispatches one record store host-service call
// through the transport injected by pluginbridge.
func invokeRecordStoreHostService(
	invoker HostServiceInvoker,
	service string,
	method string,
	resourceRef string,
	table string,
	payload []byte,
) ([]byte, error) {
	if invoker == nil {
		return nil, gerror.New("record store capability host-service invoker is not configured")
	}
	return invoker(service, method, resourceRef, table, payload)
}

// marshalJSONValue encodes one arbitrary JSON-compatible value for host-service
// payload transport.
func marshalJSONValue(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

// decodeJSONRecord decodes one JSON-encoded record returned by the host bridge.
func decodeJSONRecord(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return record, nil
}

// decodeJSONRecordList decodes the JSON-encoded record list returned by the
// host bridge list endpoint.
func decodeJSONRecordList(items [][]byte) ([]map[string]any, error) {
	if len(items) == 0 {
		return []map[string]any{}, nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		record, err := decodeJSONRecord(item)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}
