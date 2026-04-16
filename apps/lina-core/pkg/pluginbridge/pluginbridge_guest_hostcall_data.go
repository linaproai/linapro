//go:build wasip1

// This file provides guest-side helpers for the governed structured data host service.

package pluginbridge

import "encoding/json"

// DataHostService exposes the compatibility guest-side helpers for the governed
// structured data host service. New guest code should prefer plugindb.
type DataHostService interface {
	// List executes one governed structured data list request.
	List(table string, filters map[string]string, pageNum int32, pageSize int32) (*DataListResult, error)
	// ListRequest executes one governed structured data list request with the raw host-service request payload.
	ListRequest(table string, request *HostServiceDataListRequest) (*HostServiceDataListResponse, error)
	// Get reads one governed record by key from an authorized table.
	Get(table string, key any) (map[string]any, bool, error)
	// GetRequest executes one governed data get request with the raw host-service request payload.
	GetRequest(table string, request *HostServiceDataGetRequest) (*DataGetResult, error)
	// Create creates one governed record in an authorized table.
	Create(table string, record map[string]any) (*DataMutationResult, error)
	// Update updates one governed record in an authorized table.
	Update(table string, key any, record map[string]any) (*DataMutationResult, error)
	// Delete deletes one governed record in an authorized table.
	Delete(table string, key any) (*DataMutationResult, error)
	// Transaction executes one governed structured data transaction.
	Transaction(table string, operations []*DataTransactionInput) (*DataTransactionResult, error)
}

type dataHostService struct{}

var defaultDataHostService DataHostService = &dataHostService{}

// DataListResult is the decoded guest-side result of one data list request.
type DataListResult struct {
	// Records is the ordered JSON-decoded result set.
	Records []map[string]any
	// Total is the total number of matching rows before pagination.
	Total int32
}

// DataMutationResult is the decoded guest-side result of one data mutation.
type DataMutationResult struct {
	// AffectedRows is the number of rows affected by the mutation.
	AffectedRows int64
	// Key is the JSON-decoded resource key returned by the host when available.
	Key any
	// Record is the optional JSON-decoded record snapshot returned by the host.
	Record map[string]any
}

// DataTransactionInput describes one guest-side transaction step.
type DataTransactionInput struct {
	// Method is one structured mutation method such as create/update/delete.
	Method string
	// Key is the optional resource key used by update/delete.
	Key any
	// Record is the optional input document used by create/update.
	Record map[string]any
}

// DataTransactionResult is the decoded guest-side result of one data transaction.
type DataTransactionResult struct {
	// Results is the ordered list of per-step mutation results.
	Results []*DataMutationResult
	// AffectedRows is the aggregate affected row count across all steps.
	AffectedRows int64
}

// DataGetResult is the decoded guest-side result of one data get request.
type DataGetResult struct {
	// Found reports whether the requested record exists.
	Found bool
	// Record is the optional JSON-decoded record snapshot returned by the host.
	Record map[string]any
}

// Data returns the compatibility structured data host service guest client.
// New guest code should prefer plugindb.Open().
func Data() DataHostService {
	return defaultDataHostService
}

// List executes one governed structured data list request.
func (s *dataHostService) List(
	table string,
	filters map[string]string,
	pageNum int32,
	pageSize int32,
) (*DataListResult, error) {
	response, err := s.ListRequest(table, &HostServiceDataListRequest{
		Filters:  filters,
		PageNum:  pageNum,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return &DataListResult{}, nil
	}
	records, err := unmarshalJSONRecordList(response.Records)
	if err != nil {
		return nil, err
	}
	return &DataListResult{
		Records: records,
		Total:   response.Total,
	}, nil
}

// ListRequest executes one governed structured data list request with the raw
// host-service request payload.
func (s *dataHostService) ListRequest(table string, request *HostServiceDataListRequest) (*HostServiceDataListResponse, error) {
	payload, err := invokeHostService(
		HostServiceData,
		HostServiceMethodDataList,
		"",
		table,
		MarshalHostServiceDataListRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceDataListResponse(payload)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Get reads one governed record by key from an authorized table.
func (s *dataHostService) Get(table string, key any) (map[string]any, bool, error) {
	keyJSON, err := marshalJSONValue(key)
	if err != nil {
		return nil, false, err
	}
	response, err := s.GetRequest(table, &HostServiceDataGetRequest{
		KeyJSON: keyJSON,
	})
	if err != nil {
		return nil, false, err
	}
	if response == nil || !response.Found {
		return nil, false, nil
	}
	return response.Record, true, nil
}

// GetRequest executes one governed data get request with the raw host-service
// request payload.
func (s *dataHostService) GetRequest(table string, request *HostServiceDataGetRequest) (*DataGetResult, error) {
	payload, err := invokeHostService(
		HostServiceData,
		HostServiceMethodDataGet,
		"",
		table,
		MarshalHostServiceDataGetRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceDataGetResponse(payload)
	if err != nil {
		return nil, err
	}
	result := &DataGetResult{}
	if response == nil || !response.Found {
		return result, nil
	}
	record, err := unmarshalJSONRecord(response.RecordJSON)
	if err != nil {
		return nil, err
	}
	result.Found = true
	result.Record = record
	return result, nil
}

// Create creates one governed record in an authorized table.
func (s *dataHostService) Create(table string, record map[string]any) (*DataMutationResult, error) {
	return s.mutate(table, HostServiceMethodDataCreate, nil, record)
}

// Update updates one governed record in an authorized table.
func (s *dataHostService) Update(table string, key any, record map[string]any) (*DataMutationResult, error) {
	return s.mutate(table, HostServiceMethodDataUpdate, key, record)
}

// Delete deletes one governed record in an authorized table.
func (s *dataHostService) Delete(table string, key any) (*DataMutationResult, error) {
	return s.mutate(table, HostServiceMethodDataDelete, key, nil)
}

// Transaction executes one governed structured data transaction.
func (s *dataHostService) Transaction(
	table string,
	operations []*DataTransactionInput,
) (*DataTransactionResult, error) {
	request := &HostServiceDataTransactionRequest{
		Operations: make([]*HostServiceDataTransactionOperation, 0, len(operations)),
	}
	for _, operation := range operations {
		if operation == nil {
			continue
		}
		keyJSON, err := marshalJSONValue(operation.Key)
		if err != nil {
			return nil, err
		}
		recordJSON, err := marshalJSONValue(operation.Record)
		if err != nil {
			return nil, err
		}
		request.Operations = append(request.Operations, &HostServiceDataTransactionOperation{
			Method:     operation.Method,
			KeyJSON:    keyJSON,
			RecordJSON: recordJSON,
		})
	}
	payload, err := invokeHostService(
		HostServiceData,
		HostServiceMethodDataTransaction,
		"",
		table,
		MarshalHostServiceDataTransactionRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceDataTransactionResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return &DataTransactionResult{}, nil
	}
	results := make([]*DataMutationResult, 0, len(response.Results))
	for _, result := range response.Results {
		decoded, decodeErr := decodeDataMutationResult(result)
		if decodeErr != nil {
			return nil, decodeErr
		}
		results = append(results, decoded)
	}
	return &DataTransactionResult{
		Results:      results,
		AffectedRows: response.AffectedRows,
	}, nil
}

func (s *dataHostService) mutate(
	table string,
	method string,
	key any,
	record map[string]any,
) (*DataMutationResult, error) {
	keyJSON, err := marshalJSONValue(key)
	if err != nil {
		return nil, err
	}
	recordJSON, err := marshalJSONValue(record)
	if err != nil {
		return nil, err
	}
	payload, err := invokeHostService(
		HostServiceData,
		method,
		"",
		table,
		MarshalHostServiceDataMutationRequest(&HostServiceDataMutationRequest{
			KeyJSON:    keyJSON,
			RecordJSON: recordJSON,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := UnmarshalHostServiceDataMutationResponse(payload)
	if err != nil {
		return nil, err
	}
	return decodeDataMutationResult(response)
}

func decodeDataMutationResult(response *HostServiceDataMutationResponse) (*DataMutationResult, error) {
	if response == nil {
		return &DataMutationResult{}, nil
	}
	key, err := unmarshalJSONValue(response.KeyJSON)
	if err != nil {
		return nil, err
	}
	record, err := unmarshalJSONRecord(response.RecordJSON)
	if err != nil {
		return nil, err
	}
	return &DataMutationResult{
		AffectedRows: response.AffectedRows,
		Key:          key,
		Record:       record,
	}, nil
}

func marshalJSONValue(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func unmarshalJSONValue(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func unmarshalJSONRecord(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	record := make(map[string]any)
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return record, nil
}

func unmarshalJSONRecordList(items [][]byte) ([]map[string]any, error) {
	if len(items) == 0 {
		return []map[string]any{}, nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		record, err := unmarshalJSONRecord(item)
		if err != nil {
			return nil, err
		}
		if record == nil {
			record = map[string]any{}
		}
		result = append(result, record)
	}
	return result, nil
}
