// This file defines the structured data host service request and response
// codecs shared by guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"

	"lina-core/pkg/plugindb/shared"
)

// HostServiceDataListRequest carries one governed paged list request.
type HostServiceDataListRequest struct {
	// Filters contains declared query filter values keyed by API field name.
	Filters map[string]string `json:"filters,omitempty"`
	// PageNum is the 1-based requested page number.
	PageNum int32 `json:"pageNum,omitempty"`
	// PageSize is the requested page size.
	PageSize int32 `json:"pageSize,omitempty"`
	// PlanJSON is the optional JSON-encoded typed query plan used by plugindb.
	PlanJSON []byte `json:"planJson,omitempty"`
}

// HostServiceDataListResponse carries one governed paged list response.
type HostServiceDataListResponse struct {
	// Records contains one JSON document per returned row.
	Records [][]byte `json:"records,omitempty"`
	// Total is the total number of matching rows before pagination.
	Total int32 `json:"total,omitempty"`
}

// HostServiceDataGetRequest carries one governed detail query by key.
type HostServiceDataGetRequest struct {
	// KeyJSON is the JSON-encoded key value matching the resource keyField.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// PlanJSON is the optional JSON-encoded typed query plan used by plugindb.
	PlanJSON []byte `json:"planJson,omitempty"`
}

// HostServiceDataGetResponse carries one governed detail response.
type HostServiceDataGetResponse struct {
	// Found reports whether one matching row exists inside the current governance boundary.
	Found bool `json:"found"`
	// RecordJSON is the JSON-encoded record when Found is true.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataMutationRequest carries one governed create/update/delete request.
type HostServiceDataMutationRequest struct {
	// KeyJSON is the JSON-encoded key value for update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the JSON-encoded input document for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataMutationResponse carries one governed mutation response.
type HostServiceDataMutationResponse struct {
	// AffectedRows is the number of rows affected by the mutation.
	AffectedRows int64 `json:"affectedRows,omitempty"`
	// KeyJSON is the JSON-encoded resource key returned after create/update when available.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the optional JSON-encoded record snapshot returned by the host.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataTransactionOperation carries one structured mutation step inside a transaction.
type HostServiceDataTransactionOperation struct {
	// Method is one structured data mutation method such as create/update/delete.
	Method string `json:"method"`
	// KeyJSON is the JSON-encoded resource key used by update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the JSON-encoded input document used by create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataTransactionRequest carries one governed transaction request.
type HostServiceDataTransactionRequest struct {
	// Operations is the ordered list of mutation steps executed atomically.
	Operations []*HostServiceDataTransactionOperation `json:"operations,omitempty"`
}

// HostServiceDataTransactionResponse carries one governed transaction result summary.
type HostServiceDataTransactionResponse struct {
	// Results is the ordered list of per-step mutation results.
	Results []*HostServiceDataMutationResponse `json:"results,omitempty"`
	// AffectedRows is the aggregate affected row count across all steps.
	AffectedRows int64 `json:"affectedRows,omitempty"`
}

// MarshalHostServiceDataListRequest encodes one data list request.
func MarshalHostServiceDataListRequest(req *HostServiceDataListRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if len(req.Filters) > 0 {
		content = appendStringMap(content, 1, req.Filters)
	}
	if req.PageNum > 0 {
		content = appendVarintField(content, 2, uint64(req.PageNum))
	}
	if req.PageSize > 0 {
		content = appendVarintField(content, 3, uint64(req.PageSize))
	}
	if len(req.PlanJSON) > 0 {
		content = appendBytesField(content, 4, req.PlanJSON)
	}
	return content
}

// UnmarshalHostServiceDataListRequest decodes one data list request.
func UnmarshalHostServiceDataListRequest(data []byte) (*HostServiceDataListRequest, error) {
	out := &HostServiceDataListRequest{
		Filters: make(map[string]string),
	}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data list request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data list request filters 失败")
			}
			if err := unmarshalStringEntry(value, out.Filters); err != nil {
				return nil, err
			}
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data list request pageNum 失败")
			}
			out.PageNum = int32(value)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data list request pageSize 失败")
			}
			out.PageSize = int32(value)
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data list request planJson 失败")
			}
			out.PlanJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data list request 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Filters) == 0 {
		out.Filters = nil
	}
	return out, nil
}

// MarshalHostServiceDataListResponse encodes one data list response.
func MarshalHostServiceDataListResponse(resp *HostServiceDataListResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	for _, record := range resp.Records {
		if len(record) > 0 {
			content = appendBytesField(content, 1, record)
		}
	}
	if resp.Total > 0 {
		content = appendVarintField(content, 2, uint64(resp.Total))
	}
	return content
}

// UnmarshalHostServiceDataListResponse decodes one data list response.
func UnmarshalHostServiceDataListResponse(data []byte) (*HostServiceDataListResponse, error) {
	out := &HostServiceDataListResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data list response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data list response record 失败")
			}
			out.Records = append(out.Records, append([]byte(nil), value...))
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data list response total 失败")
			}
			out.Total = int32(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data list response 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Records) == 0 {
		out.Records = nil
	}
	return out, nil
}

// MarshalHostServiceDataGetRequest encodes one data get request.
func MarshalHostServiceDataGetRequest(req *HostServiceDataGetRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if len(req.KeyJSON) > 0 {
		content = appendBytesField(content, 1, req.KeyJSON)
	}
	if len(req.PlanJSON) > 0 {
		content = appendBytesField(content, 2, req.PlanJSON)
	}
	return content
}

// UnmarshalHostServiceDataGetRequest decodes one data get request.
func UnmarshalHostServiceDataGetRequest(data []byte) (*HostServiceDataGetRequest, error) {
	out := &HostServiceDataGetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data get request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data get request keyJson 失败")
			}
			out.KeyJSON = append([]byte(nil), value...)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data get request planJson 失败")
			}
			out.PlanJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data get request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceDataGetResponse encodes one data get response.
func MarshalHostServiceDataGetResponse(resp *HostServiceDataGetResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if len(resp.RecordJSON) > 0 {
		content = appendBytesField(content, 2, resp.RecordJSON)
	}
	return content
}

// UnmarshalHostServiceDataGetResponse decodes one data get response.
func UnmarshalHostServiceDataGetResponse(data []byte) (*HostServiceDataGetResponse, error) {
	out := &HostServiceDataGetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data get response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data get response found 失败")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data get response recordJson 失败")
			}
			out.RecordJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data get response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceDataMutationRequest encodes one data mutation request.
func MarshalHostServiceDataMutationRequest(req *HostServiceDataMutationRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if len(req.KeyJSON) > 0 {
		content = appendBytesField(content, 1, req.KeyJSON)
	}
	if len(req.RecordJSON) > 0 {
		content = appendBytesField(content, 2, req.RecordJSON)
	}
	return content
}

// UnmarshalHostServiceDataMutationRequest decodes one data mutation request.
func UnmarshalHostServiceDataMutationRequest(data []byte) (*HostServiceDataMutationRequest, error) {
	out := &HostServiceDataMutationRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data mutation request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data mutation request keyJson 失败")
			}
			out.KeyJSON = append([]byte(nil), value...)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data mutation request recordJson 失败")
			}
			out.RecordJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data mutation request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceDataMutationResponse encodes one data mutation response.
func MarshalHostServiceDataMutationResponse(resp *HostServiceDataMutationResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.AffectedRows > 0 {
		content = appendVarintField(content, 1, uint64(resp.AffectedRows))
	}
	if len(resp.KeyJSON) > 0 {
		content = appendBytesField(content, 2, resp.KeyJSON)
	}
	if len(resp.RecordJSON) > 0 {
		content = appendBytesField(content, 3, resp.RecordJSON)
	}
	return content
}

// UnmarshalHostServiceDataMutationResponse decodes one data mutation response.
func UnmarshalHostServiceDataMutationResponse(data []byte) (*HostServiceDataMutationResponse, error) {
	out := &HostServiceDataMutationResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data mutation response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data mutation response affectedRows 失败")
			}
			out.AffectedRows = int64(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data mutation response keyJson 失败")
			}
			out.KeyJSON = append([]byte(nil), value...)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data mutation response recordJson 失败")
			}
			out.RecordJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data mutation response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceDataTransactionRequest encodes one data transaction request.
func MarshalHostServiceDataTransactionRequest(req *HostServiceDataTransactionRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	for _, operation := range req.Operations {
		entry := marshalHostServiceDataTransactionOperation(operation)
		if len(entry) > 0 {
			content = appendBytesField(content, 1, entry)
		}
	}
	return content
}

// UnmarshalHostServiceDataTransactionRequest decodes one data transaction request.
func UnmarshalHostServiceDataTransactionRequest(data []byte) (*HostServiceDataTransactionRequest, error) {
	out := &HostServiceDataTransactionRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data transaction request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction request operation 失败")
			}
			operation, err := unmarshalHostServiceDataTransactionOperation(value)
			if err != nil {
				return nil, err
			}
			out.Operations = append(out.Operations, operation)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data transaction request 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Operations) == 0 {
		out.Operations = nil
	}
	return out, nil
}

// MarshalHostServiceDataTransactionResponse encodes one data transaction response.
func MarshalHostServiceDataTransactionResponse(resp *HostServiceDataTransactionResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	for _, result := range resp.Results {
		entry := MarshalHostServiceDataMutationResponse(result)
		if len(entry) > 0 {
			content = appendBytesField(content, 1, entry)
		}
	}
	if resp.AffectedRows > 0 {
		content = appendVarintField(content, 2, uint64(resp.AffectedRows))
	}
	return content
}

// UnmarshalHostServiceDataTransactionResponse decodes one data transaction response.
func UnmarshalHostServiceDataTransactionResponse(data []byte) (*HostServiceDataTransactionResponse, error) {
	out := &HostServiceDataTransactionResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data transaction response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction response result 失败")
			}
			result, err := UnmarshalHostServiceDataMutationResponse(value)
			if err != nil {
				return nil, err
			}
			out.Results = append(out.Results, result)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction response affectedRows 失败")
			}
			out.AffectedRows = int64(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data transaction response 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Results) == 0 {
		out.Results = nil
	}
	return out, nil
}

// DecodeHostServiceDataListPlan decodes the optional typed query plan embedded
// in one list request.
func DecodeHostServiceDataListPlan(request *HostServiceDataListRequest) (*shared.DataQueryPlan, error) {
	if request == nil || len(request.PlanJSON) == 0 {
		return nil, nil
	}
	return shared.UnmarshalQueryPlanJSON(request.PlanJSON)
}

// DecodeHostServiceDataGetPlan decodes the optional typed query plan embedded
// in one get request.
func DecodeHostServiceDataGetPlan(request *HostServiceDataGetRequest) (*shared.DataQueryPlan, error) {
	if request == nil || len(request.PlanJSON) == 0 {
		return nil, nil
	}
	return shared.UnmarshalQueryPlanJSON(request.PlanJSON)
}

func marshalHostServiceDataTransactionOperation(operation *HostServiceDataTransactionOperation) []byte {
	var content []byte
	if operation == nil {
		return content
	}
	if operation.Method != "" {
		content = appendStringField(content, 1, operation.Method)
	}
	if len(operation.KeyJSON) > 0 {
		content = appendBytesField(content, 2, operation.KeyJSON)
	}
	if len(operation.RecordJSON) > 0 {
		content = appendBytesField(content, 3, operation.RecordJSON)
	}
	return content
}

func unmarshalHostServiceDataTransactionOperation(data []byte) (*HostServiceDataTransactionOperation, error) {
	out := &HostServiceDataTransactionOperation{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 data transaction operation tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction operation method 失败")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction operation keyJson 失败")
			}
			out.KeyJSON = append([]byte(nil), value...)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 data transaction operation recordJson 失败")
			}
			out.RecordJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 data transaction operation 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
