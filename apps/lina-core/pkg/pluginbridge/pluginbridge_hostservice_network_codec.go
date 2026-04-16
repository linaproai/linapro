// This file defines the network host service request and response codecs
// shared by guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceNetworkRequest carries one governed outbound HTTP request.
type HostServiceNetworkRequest struct {
	// Method is the HTTP verb requested by the guest.
	Method string `json:"method"`
	// Headers is the guest-controlled request header map.
	Headers map[string]string `json:"headers,omitempty"`
	// Body is the optional raw request body.
	Body []byte `json:"body,omitempty"`
}

// HostServiceNetworkResponse carries one governed outbound HTTP response.
type HostServiceNetworkResponse struct {
	// StatusCode is the upstream HTTP status code.
	StatusCode int32 `json:"statusCode"`
	// Headers is the response header map flattened to first values.
	Headers map[string]string `json:"headers,omitempty"`
	// Body is the raw response body.
	Body []byte `json:"body,omitempty"`
	// ContentType is the normalized response MIME type.
	ContentType string `json:"contentType,omitempty"`
}

// MarshalHostServiceNetworkRequest encodes one network request.
func MarshalHostServiceNetworkRequest(req *HostServiceNetworkRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Method != "" {
		content = appendStringField(content, 1, req.Method)
	}
	if len(req.Headers) > 0 {
		content = appendStringMap(content, 2, req.Headers)
	}
	if len(req.Body) > 0 {
		content = appendBytesField(content, 3, req.Body)
	}
	return content
}

// UnmarshalHostServiceNetworkRequest decodes one network request.
func UnmarshalHostServiceNetworkRequest(data []byte) (*HostServiceNetworkRequest, error) {
	out := &HostServiceNetworkRequest{
		Headers: make(map[string]string),
	}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 network request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 network request method 失败")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 network request headers 失败")
			}
			if err := unmarshalStringEntry(value, out.Headers); err != nil {
				return nil, err
			}
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 network request body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 network request 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Headers) == 0 {
		out.Headers = nil
	}
	return out, nil
}

// MarshalHostServiceNetworkResponse encodes one network response.
func MarshalHostServiceNetworkResponse(resp *HostServiceNetworkResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.StatusCode > 0 {
		content = appendVarintField(content, 1, uint64(resp.StatusCode))
	}
	if len(resp.Headers) > 0 {
		content = appendStringMap(content, 2, resp.Headers)
	}
	if len(resp.Body) > 0 {
		content = appendBytesField(content, 3, resp.Body)
	}
	if resp.ContentType != "" {
		content = appendStringField(content, 4, resp.ContentType)
	}
	return content
}

// UnmarshalHostServiceNetworkResponse decodes one network response.
func UnmarshalHostServiceNetworkResponse(data []byte) (*HostServiceNetworkResponse, error) {
	out := &HostServiceNetworkResponse{
		Headers: make(map[string]string),
	}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 network response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 network response statusCode 失败")
			}
			out.StatusCode = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 network response headers 失败")
			}
			if err := unmarshalStringEntry(value, out.Headers); err != nil {
				return nil, err
			}
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 network response body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 network response contentType 失败")
			}
			out.ContentType = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 network response 字段失败")
			}
			content = content[size:]
		}
	}
	if len(out.Headers) == 0 {
		out.Headers = nil
	}
	return out, nil
}
