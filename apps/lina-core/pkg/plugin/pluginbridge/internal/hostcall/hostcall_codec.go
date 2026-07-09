// This file implements the protobuf-wire codec for host call request and
// response envelopes. Each opcode has its own message layout following the
// same hand-rolled protowire encoding used by the bridge codec in codec.go.

package hostcall

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"

	"lina-core/pkg/bizerr"
)

const (
	hostCallErrorCodeCapabilityDenied = "HOST_CALL_CAPABILITY_DENIED"
	hostCallErrorCodeNotFound         = "HOST_CALL_NOT_FOUND"
	hostCallErrorCodeInvalidRequest   = "HOST_CALL_INVALID_REQUEST"
	hostCallErrorCodeInternal         = "HOST_CALL_INTERNAL_ERROR"
)

// ---------------------------------------------------------------------------
// Generic host call response envelope
// ---------------------------------------------------------------------------

// HostCallResponseEnvelope wraps every host call response with a status code.
type HostCallResponseEnvelope struct {
	// Status indicates the outcome: 0=success, 1=capability_denied, 2=not_found,
	// 3=invalid_request, 4=internal_error.
	Status uint32 `json:"status"`
	// Payload carries opcode-specific response data on success, or a JSON
	// HostCallErrorPayload on failure.
	Payload []byte `json:"payload,omitempty"`
}

// HostCallErrorPayload carries structured, localizable error metadata returned
// by failed host calls without changing the outer ABI envelope.
type HostCallErrorPayload struct {
	// ErrorCode is a stable machine-readable error code.
	ErrorCode string `json:"errorCode"`
	// MessageKey is the runtime i18n key used by the guest or UI boundary.
	MessageKey string `json:"messageKey"`
	// MessageParams carries named parameters for localized rendering.
	MessageParams map[string]any `json:"messageParams,omitempty"`
	// Fallback is an English fallback message.
	Fallback string `json:"fallback"`
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
			return nil, gerror.New("failed to decode host call response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host call response status")
			}
			out.Status = uint32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host call response payload")
			}
			out.Payload = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown host call response field")
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
	return NewHostCallErrorPayloadResponse(status, HostCallErrorPayload{
		ErrorCode:  defaultHostCallErrorCode(status),
		MessageKey: defaultHostCallMessageKey(status),
		Fallback:   normalizeHostCallErrorFallback(status, message),
	})
}

// NewHostCallErrorResponseFromError builds an error response from a structured
// bizerr when present, falling back to status-scoped host-call metadata.
func NewHostCallErrorResponseFromError(status uint32, err error) *HostCallResponseEnvelope {
	if err == nil {
		return NewHostCallErrorResponse(status, "")
	}
	if messageErr, ok := bizerr.As(err); ok {
		return NewHostCallErrorPayloadResponse(status, HostCallErrorPayload{
			ErrorCode:     messageErr.RuntimeCode(),
			MessageKey:    messageErr.MessageKey(),
			MessageParams: messageErr.Params(),
			Fallback:      messageErr.Fallback(),
		})
	}
	return NewHostCallErrorResponse(status, err.Error())
}

// NewHostCallErrorPayloadResponse builds an error response with explicit
// structured payload metadata.
func NewHostCallErrorPayloadResponse(status uint32, payload HostCallErrorPayload) *HostCallResponseEnvelope {
	payload.ErrorCode = strings.TrimSpace(payload.ErrorCode)
	if payload.ErrorCode == "" {
		payload.ErrorCode = defaultHostCallErrorCode(status)
	}
	payload.MessageKey = strings.TrimSpace(payload.MessageKey)
	if payload.MessageKey == "" {
		payload.MessageKey = defaultHostCallMessageKey(status)
	}
	payload.Fallback = normalizeHostCallErrorFallback(status, payload.Fallback)
	content, err := json.Marshal(payload)
	if err != nil {
		content = []byte(`{"errorCode":"HOST_CALL_INTERNAL_ERROR","messageKey":"error.host_call.internal_error","fallback":"Host call failed"}`)
	}
	return &HostCallResponseEnvelope{Status: status, Payload: content}
}

// UnmarshalHostCallErrorPayload decodes the structured error payload returned
// by a failed host call.
func UnmarshalHostCallErrorPayload(data []byte) (*HostCallErrorPayload, error) {
	out := &HostCallErrorPayload{}
	if err := json.Unmarshal(data, out); err != nil {
		return nil, gerror.Wrap(err, "decode host call error payload failed")
	}
	return out, nil
}

func defaultHostCallErrorCode(status uint32) string {
	switch status {
	case HostCallStatusCapabilityDenied:
		return hostCallErrorCodeCapabilityDenied
	case HostCallStatusNotFound:
		return hostCallErrorCodeNotFound
	case HostCallStatusInvalidRequest:
		return hostCallErrorCodeInvalidRequest
	case HostCallStatusInternalError:
		return hostCallErrorCodeInternal
	default:
		return "HOST_CALL_FAILED"
	}
}

func defaultHostCallMessageKey(status uint32) string {
	switch status {
	case HostCallStatusCapabilityDenied:
		return "error.host_call.capability_denied"
	case HostCallStatusNotFound:
		return "error.host_call.not_found"
	case HostCallStatusInvalidRequest:
		return "error.host_call.invalid_request"
	case HostCallStatusInternalError:
		return "error.host_call.internal_error"
	default:
		return "error.host_call.failed"
	}
}

func normalizeHostCallErrorFallback(status uint32, fallback string) string {
	if value := strings.TrimSpace(fallback); value != "" {
		return value
	}
	switch status {
	case HostCallStatusCapabilityDenied:
		return "Host call capability denied"
	case HostCallStatusNotFound:
		return "Host call target not found"
	case HostCallStatusInvalidRequest:
		return "Invalid host call request"
	case HostCallStatusInternalError:
		return "Host call failed"
	default:
		return "Host call failed"
	}
}

// ---------------------------------------------------------------------------
// OpcodeLog: structured log request
// ---------------------------------------------------------------------------

// HostCallLogRequest carries a structured log entry from the guest.
type HostCallLogRequest struct {
	// Level is the guest log severity encoded as an integer level.
	Level int32 `json:"level"`
	// Message is the primary log message emitted by the guest.
	Message string `json:"message"`
	// Fields carries structured key-value log attributes attached to the entry.
	Fields map[string]string `json:"fields,omitempty"`
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
			return nil, gerror.New("failed to decode host call log request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host call log level")
			}
			out.Level = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host call log message")
			}
			out.Message = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host call log fields")
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
				return nil, gerror.New("failed to skip unknown host call log field")
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
	// Key is the plugin-scoped runtime state key to read.
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
			return nil, gerror.New("failed to decode state get request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state get key")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown state get field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// HostCallStateGetResponse carries the state value and existence flag.
type HostCallStateGetResponse struct {
	// Value stores the runtime state value when the key exists.
	Value string `json:"value"`
	// Found reports whether the requested runtime state key exists.
	Found bool `json:"found"`
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
			return nil, gerror.New("failed to decode state get response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state get value")
			}
			out.Value = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state get found")
			}
			out.Found = value != 0
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown state get response field")
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
	// Key is the plugin-scoped runtime state key to write.
	Key string `json:"key"`
	// Value is the runtime state payload stored under Key.
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
			return nil, gerror.New("failed to decode state set request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state set key")
			}
			out.Key = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state set value")
			}
			out.Value = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown state set field")
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
	// Key is the plugin-scoped runtime state key to delete.
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
			return nil, gerror.New("failed to decode state delete request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode state delete key")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown state delete field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
