//go:build wasip1

// This file implements the governed plugindb execution path for wasm guests.

package plugindb

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/plugindb/shared"
)

func (q *Query) ensureExecutionReady(action shared.DataPlanAction) error {
	if q == nil {
		return gerror.New("plugindb query is nil")
	}
	if q.err != nil {
		return q.err
	}
	if strings.TrimSpace(q.table) == "" {
		return gerror.New("plugindb table cannot be empty")
	}
	for _, filter := range q.plan.Filters {
		if err := shared.ValidateDataFilter(filter); err != nil {
			return err
		}
	}
	for _, order := range q.plan.Orders {
		if err := shared.ValidateDataOrder(order); err != nil {
			return err
		}
	}
	q.plan.Action = action
	if err := shared.ValidateDataQueryPlan(q.plan); err != nil {
		return err
	}
	return nil
}

// One executes one governed single-record lookup.
func (q *Query) One() (map[string]any, bool, error) {
	if err := q.ensureExecutionReady(shared.DataPlanActionGet); err != nil {
		return nil, false, err
	}
	dataSvc := pluginbridge.Data()
	if len(q.plan.KeyJSON) > 0 {
		planJSON, err := shared.MarshalQueryPlanJSON(q.plan)
		if err != nil {
			return nil, false, err
		}
		responsePayload, err := dataSvc.GetRequest(q.table, &pluginbridge.HostServiceDataGetRequest{
			KeyJSON:  append([]byte(nil), q.plan.KeyJSON...),
			PlanJSON: planJSON,
		})
		if err != nil {
			return nil, false, err
		}
		return responsePayload.Record, responsePayload.Found, nil
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

// All executes one governed paged list query.
func (q *Query) All() ([]map[string]any, int32, error) {
	if err := q.ensureExecutionReady(shared.DataPlanActionList); err != nil {
		return nil, 0, err
	}
	planJSON, err := shared.MarshalQueryPlanJSON(q.plan)
	if err != nil {
		return nil, 0, err
	}
	result, err := pluginbridge.Data().ListRequest(q.table, &pluginbridge.HostServiceDataListRequest{
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
		return 0, gerror.New("plugindb query is nil")
	}
	if err := q.ensureExecutionReady(shared.DataPlanActionCount); err != nil {
		return 0, err
	}
	planJSON, err := shared.MarshalQueryPlanJSON(q.plan)
	if err != nil {
		return 0, err
	}
	result, err := pluginbridge.Data().ListRequest(q.table, &pluginbridge.HostServiceDataListRequest{
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
	if err := q.ensureExecutionReady(shared.DataPlanActionCreate); err != nil {
		return nil, err
	}
	result, err := pluginbridge.Data().Create(q.table, record)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result), nil
}

// Update executes one governed update mutation.
func (q *Query) Update(record map[string]any) (*MutationResult, error) {
	if err := q.ensureExecutionReady(shared.DataPlanActionUpdate); err != nil {
		return nil, err
	}
	if len(q.plan.KeyJSON) == 0 {
		return nil, gerror.New("plugindb update requires WhereKey")
	}
	key, err := shared.UnmarshalValueJSON(q.plan.KeyJSON)
	if err != nil {
		return nil, err
	}
	result, err := pluginbridge.Data().Update(q.table, key, record)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result), nil
}

// Delete executes one governed delete mutation.
func (q *Query) Delete() (*MutationResult, error) {
	if err := q.ensureExecutionReady(shared.DataPlanActionDelete); err != nil {
		return nil, err
	}
	if len(q.plan.KeyJSON) == 0 {
		return nil, gerror.New("plugindb delete requires WhereKey")
	}
	key, err := shared.UnmarshalValueJSON(q.plan.KeyJSON)
	if err != nil {
		return nil, err
	}
	result, err := pluginbridge.Data().Delete(q.table, key)
	if err != nil {
		return nil, err
	}
	return decodeMutationResult(result), nil
}

// Transaction executes one governed structured mutation transaction.
func (db *DB) Transaction(fn func(tx *Tx) error) error {
	if fn == nil {
		return gerror.New("plugindb transaction callback cannot be nil")
	}
	tx := &Tx{}
	if err := fn(tx); err != nil {
		return err
	}
	if tx.err != nil {
		return tx.err
	}
	if strings.TrimSpace(tx.table) == "" {
		return gerror.New("plugindb transaction table cannot be empty")
	}
	operations := make([]*pluginbridge.DataTransactionInput, 0, len(tx.operations))
	for _, operation := range tx.operations {
		if operation == nil {
			continue
		}
		key, err := shared.UnmarshalValueJSON(operation.KeyJSON)
		if err != nil {
			return err
		}
		recordValue, err := shared.UnmarshalValueJSON(operation.RecordJSON)
		if err != nil {
			return err
		}
		record, _ := recordValue.(map[string]any)
		operations = append(operations, &pluginbridge.DataTransactionInput{
			Method: operation.Action.String(),
			Key:    key,
			Record: record,
		})
	}
	_, err := pluginbridge.Data().Transaction(tx.table, operations)
	return err
}

func decodeMutationResult(result *pluginbridge.DataMutationResult) *MutationResult {
	if result == nil {
		return &MutationResult{}
	}
	return &MutationResult{AffectedRows: result.AffectedRows, Key: result.Key, Record: result.Record}
}

func decodeJSONRecordList(items [][]byte) ([]map[string]any, error) {
	if len(items) == 0 {
		return []map[string]any{}, nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if len(item) == 0 {
			result = append(result, map[string]any{})
			continue
		}
		record := make(map[string]any)
		if err := json.Unmarshal(item, &record); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}
