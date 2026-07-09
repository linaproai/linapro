// This file implements manifest host-service request and response codecs.

package protocol

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceManifestGetRequest carries one manifest-relative resource path.
type HostServiceManifestGetRequest struct {
	// Path is the requested resource path relative to the plugin manifest root.
	Path string `json:"path"`
}

// HostServiceManifestGetResponse carries one manifest resource payload.
type HostServiceManifestGetResponse struct {
	// Found reports whether the requested manifest resource exists.
	Found bool `json:"found"`
	// Body is the manifest resource content when Found is true.
	Body []byte `json:"body,omitempty"`
}

// MarshalHostServiceManifestGetRequest encodes one manifest get request.
func MarshalHostServiceManifestGetRequest(req *HostServiceManifestGetRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Path != "" {
		content = appendStringField(content, 1, req.Path)
	}
	return content
}

// UnmarshalHostServiceManifestGetRequest decodes one manifest get request.
func UnmarshalHostServiceManifestGetRequest(data []byte) (*HostServiceManifestGetRequest, error) {
	out := &HostServiceManifestGetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode manifest get request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode manifest get request path")
			}
			out.Path = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown manifest get request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceManifestGetResponse encodes one manifest get response.
func MarshalHostServiceManifestGetResponse(resp *HostServiceManifestGetResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if len(resp.Body) > 0 {
		content = appendBytesField(content, 2, resp.Body)
	}
	return content
}

// UnmarshalHostServiceManifestGetResponse decodes one manifest get response.
func UnmarshalHostServiceManifestGetResponse(data []byte) (*HostServiceManifestGetResponse, error) {
	out := &HostServiceManifestGetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode manifest get response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode manifest get response found")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode manifest get response body")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown manifest get response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
