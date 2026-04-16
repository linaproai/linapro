// This file implements the structured host service invocation codec shared by
// guest helpers and the host service dispatcher.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceRequestEnvelope carries one structured host service invocation.
type HostServiceRequestEnvelope struct {
	// Service identifies the logical host service.
	Service string `json:"service"`
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
	return content
}

// UnmarshalHostServiceRequestEnvelope decodes one structured host service invocation.
func UnmarshalHostServiceRequestEnvelope(data []byte) (*HostServiceRequestEnvelope, error) {
	out := &HostServiceRequestEnvelope{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 host service request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host service request service 失败")
			}
			out.Service = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host service request method 失败")
			}
			out.Method = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host service request resourceRef 失败")
			}
			out.ResourceRef = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host service request table 失败")
			}
			out.Table = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 host service request payload 失败")
			}
			out.Payload = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 host service request 字段失败")
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
			return nil, gerror.New("解析 host service value response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 host service value response value 失败")
			}
			out.Value = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 host service value response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
