// This file implements the protobuf-wire codec for host call request and
// response envelopes. Each opcode has its own message layout following the
// same hand-rolled protowire encoding used by the bridge codec in codec.go.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// ---------------------------------------------------------------------------
// Generic host call response envelope
// ---------------------------------------------------------------------------

// HostCallResponseEnvelope wraps every host call response with a status code.
type HostCallResponseEnvelope struct {
	// Status indicates the outcome: 0=success, 1=capability_denied, 2=not_found,
	// 3=invalid_request, 4=internal_error.
	Status uint32 `json:"status"`
	// Payload carries opcode-specific response data on success, or an error
	// message string on failure.
	Payload []byte `json:"payload,omitempty"`
}

// MarshalHostCallResponse encodes a host call response envelope.
func MarshalHostCallResponse(resp *HostCallResponseEnvelope) []byte {
	var content []byte
	if resp.Status != 0 {
		content = appendVarintField(content, 1, uint64(resp.Status))
	}
	if len(resp.Payload) > 0 {
		content = appendBytesField(content, 2, resp.Payload)
	}
	return content
}

// UnmarshalHostCallResponse decodes a host call response envelope.
func UnmarshalHostCallResponse(data []byte) (*HostCallResponseEnvelope, error) {
	out := &HostCallResponseEnvelope{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 host call response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 host call response status 失败")
			}
			out.Status = uint32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 host call response payload 失败")
			}
			out.Payload = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 host call response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// NewHostCallSuccessResponse builds a success response with the given payload.
func NewHostCallSuccessResponse(payload []byte) *HostCallResponseEnvelope {
	return &HostCallResponseEnvelope{Status: HostCallStatusSuccess, Payload: payload}
}

// NewHostCallEmptySuccessResponse builds a success response with no payload.
func NewHostCallEmptySuccessResponse() *HostCallResponseEnvelope {
	return &HostCallResponseEnvelope{Status: HostCallStatusSuccess}
}

// NewHostCallErrorResponse builds an error response with the given status and message.
func NewHostCallErrorResponse(status uint32, message string) *HostCallResponseEnvelope {
	return &HostCallResponseEnvelope{Status: status, Payload: []byte(message)}
}

// ---------------------------------------------------------------------------
// OpcodeLog: structured log request
// ---------------------------------------------------------------------------

// HostCallLogRequest carries a structured log entry from the guest.
type HostCallLogRequest struct {
	Level   int32             `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// MarshalHostCallLogRequest encodes a log request.
func MarshalHostCallLogRequest(req *HostCallLogRequest) []byte {
	var content []byte
	if req.Level != 0 {
		content = appendVarintField(content, 1, uint64(req.Level))
	}
	if value := strings.TrimSpace(req.Message); value != "" {
		content = appendStringField(content, 2, value)
	}
	if len(req.Fields) > 0 {
		content = appendStringMap(content, 3, req.Fields)
	}
	return content
}

// UnmarshalHostCallLogRequest decodes a log request.
func UnmarshalHostCallLogRequest(data []byte) (*HostCallLogRequest, error) {
	out := &HostCallLogRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 host call log request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 host call log level 失败")
			}
			out.Level = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host call log message 失败")
			}
			out.Message = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 host call log fields 失败")
			}
			if out.Fields == nil {
				out.Fields = make(map[string]string)
			}
			if err := unmarshalStringEntry(value, out.Fields); err != nil {
				return nil, err
			}
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 host call log 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// OpcodeStateGet: state read request / response
// ---------------------------------------------------------------------------

// HostCallStateGetRequest carries a state read key.
type HostCallStateGetRequest struct {
	Key string `json:"key"`
}

// MarshalHostCallStateGetRequest encodes a state get request.
func MarshalHostCallStateGetRequest(req *HostCallStateGetRequest) []byte {
	var content []byte
	if value := strings.TrimSpace(req.Key); value != "" {
		content = appendStringField(content, 1, value)
	}
	return content
}

// UnmarshalHostCallStateGetRequest decodes a state get request.
func UnmarshalHostCallStateGetRequest(data []byte) (*HostCallStateGetRequest, error) {
	out := &HostCallStateGetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 state get request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 state get key 失败")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 state get 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// HostCallStateGetResponse carries the state value and existence flag.
type HostCallStateGetResponse struct {
	Value string `json:"value"`
	Found bool   `json:"found"`
}

// MarshalHostCallStateGetResponse encodes a state get response.
func MarshalHostCallStateGetResponse(resp *HostCallStateGetResponse) []byte {
	var content []byte
	if resp.Value != "" {
		content = appendStringField(content, 1, resp.Value)
	}
	if resp.Found {
		content = appendVarintField(content, 2, 1)
	}
	return content
}

// UnmarshalHostCallStateGetResponse decodes a state get response.
func UnmarshalHostCallStateGetResponse(data []byte) (*HostCallStateGetResponse, error) {
	out := &HostCallStateGetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 state get response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 state get value 失败")
			}
			out.Value = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 state get found 失败")
			}
			out.Found = value != 0
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 state get response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// OpcodeStateSet: state write request
// ---------------------------------------------------------------------------

// HostCallStateSetRequest carries a state write key-value pair.
type HostCallStateSetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MarshalHostCallStateSetRequest encodes a state set request.
func MarshalHostCallStateSetRequest(req *HostCallStateSetRequest) []byte {
	var content []byte
	if value := strings.TrimSpace(req.Key); value != "" {
		content = appendStringField(content, 1, value)
	}
	if req.Value != "" {
		content = appendStringField(content, 2, req.Value)
	}
	return content
}

// UnmarshalHostCallStateSetRequest decodes a state set request.
func UnmarshalHostCallStateSetRequest(data []byte) (*HostCallStateSetRequest, error) {
	out := &HostCallStateSetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 state set request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 state set key 失败")
			}
			out.Key = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 state set value 失败")
			}
			out.Value = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 state set 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// OpcodeStateDelete: state delete request
// ---------------------------------------------------------------------------

// HostCallStateDeleteRequest carries a state delete key.
type HostCallStateDeleteRequest struct {
	Key string `json:"key"`
}

// MarshalHostCallStateDeleteRequest encodes a state delete request.
func MarshalHostCallStateDeleteRequest(req *HostCallStateDeleteRequest) []byte {
	var content []byte
	if value := strings.TrimSpace(req.Key); value != "" {
		content = appendStringField(content, 1, value)
	}
	return content
}

// UnmarshalHostCallStateDeleteRequest decodes a state delete request.
func UnmarshalHostCallStateDeleteRequest(data []byte) (*HostCallStateDeleteRequest, error) {
	out := &HostCallStateDeleteRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 state delete request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 state delete key 失败")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 state delete 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
