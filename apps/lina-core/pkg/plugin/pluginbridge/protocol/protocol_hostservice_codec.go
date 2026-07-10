// This file implements the structured host service invocation codec shared by
// guest helpers and the host service dispatcher.

package protocol

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceRequestEnvelope carries one structured host service invocation.
type HostServiceRequestEnvelope struct {
	// Owner identifies the plugin-owned capability owner. Core-owned services
	// leave it empty.
	Owner string `json:"owner,omitempty"`
	// Service identifies the logical host service.
	Service string `json:"service"`
	// Version identifies the plugin-owned capability protocol version.
	Version string `json:"version,omitempty"`
	// Method identifies one method under the logical host service.
	Method string `json:"method"`
	// ResourceRef is the optional logical resource reference.
	ResourceRef string `json:"resourceRef,omitempty"`
	// Table is the optional authorized table name used by the data host service.
	Table string `json:"table,omitempty"`
	// Payload carries method-specific request bytes.
	Payload []byte `json:"payload,omitempty"`
}

// HostServiceValueResponse carries one string-based runtime info value.
type HostServiceValueResponse struct {
	// Value is the string representation returned by the host service.
	Value string `json:"value"`
}

// HostServiceJSONRequest carries one JSON-encoded ordinary host-service request.
type HostServiceJSONRequest struct {
	// Value is the compact JSON request owned by the target capability service.
	Value []byte `json:"value"`
}

// HostServiceJSONResponse carries one JSON-encoded ordinary host-service response.
type HostServiceJSONResponse struct {
	// Value is the compact JSON projection returned by the target capability service.
	Value []byte `json:"value"`
}

// MarshalHostServiceRequestEnvelope encodes one structured host service invocation.
func MarshalHostServiceRequestEnvelope(req *HostServiceRequestEnvelope) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if value := strings.TrimSpace(req.Service); value != "" {
		content = appendStringField(content, 1, value)
	}
	if value := strings.TrimSpace(req.Method); value != "" {
		content = appendStringField(content, 2, value)
	}
	if value := strings.TrimSpace(req.ResourceRef); value != "" {
		content = appendStringField(content, 3, value)
	}
	if value := strings.TrimSpace(req.Table); value != "" {
		content = appendStringField(content, 4, value)
	}
	if len(req.Payload) > 0 {
		content = appendBytesField(content, 5, req.Payload)
	}
	if value := strings.TrimSpace(req.Owner); value != "" {
		content = appendStringField(content, 6, value)
	}
	if value := strings.TrimSpace(req.Version); value != "" {
		content = appendStringField(content, 7, value)
	}
	return content
}

// UnmarshalHostServiceRequestEnvelope decodes one structured host service invocation.
func UnmarshalHostServiceRequestEnvelope(data []byte) (*HostServiceRequestEnvelope, error) {
	out := &HostServiceRequestEnvelope{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode host service request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request service")
			}
			out.Service = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request method")
			}
			out.Method = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request resourceRef")
			}
			out.ResourceRef = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request table")
			}
			out.Table = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request payload")
			}
			out.Payload = append([]byte(nil), value...)
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request owner")
			}
			out.Owner = value
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service request version")
			}
			out.Version = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown host service request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceValueResponse encodes one string value response.
func MarshalHostServiceValueResponse(resp *HostServiceValueResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if value := strings.TrimSpace(resp.Value); value != "" {
		content = appendStringField(content, 1, value)
	}
	return content
}

// UnmarshalHostServiceValueResponse decodes one string value response.
func UnmarshalHostServiceValueResponse(data []byte) (*HostServiceValueResponse, error) {
	out := &HostServiceValueResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode host service value response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service value response value")
			}
			out.Value = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown host service value response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceJSONRequest encodes one JSON value request.
func MarshalHostServiceJSONRequest(req *HostServiceJSONRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if len(req.Value) > 0 {
		content = appendBytesField(content, 1, req.Value)
	}
	return content
}

// UnmarshalHostServiceJSONRequest decodes one JSON value request.
func UnmarshalHostServiceJSONRequest(data []byte) (*HostServiceJSONRequest, error) {
	out := &HostServiceJSONRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode host service JSON request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service JSON request value")
			}
			out.Value = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown host service JSON request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceJSONResponse encodes one JSON value response.
func MarshalHostServiceJSONResponse(resp *HostServiceJSONResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if len(resp.Value) > 0 {
		content = appendBytesField(content, 1, resp.Value)
	}
	return content
}

// UnmarshalHostServiceJSONResponse decodes one JSON value response.
func UnmarshalHostServiceJSONResponse(data []byte) (*HostServiceJSONResponse, error) {
	out := &HostServiceJSONResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode host service JSON response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode host service JSON response value")
			}
			out.Value = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown host service JSON response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
