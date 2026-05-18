// This file implements the config host-service request and response codec.

package hostservice

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceConfigKeyRequest carries one config key. Empty key and "." mean
// the complete GoFrame configuration snapshot.
type HostServiceConfigKeyRequest struct {
	// Key is the requested configuration key.
	Key string `json:"key"`
}

// HostServiceConfigValueResponse carries one config value and existence flag.
type HostServiceConfigValueResponse struct {
	// Value is the returned value. The get method returns JSON; typed methods
	// return string representations.
	Value string `json:"value"`
	// Found reports whether the requested key exists.
	Found bool `json:"found"`
}

// MarshalHostServiceConfigKeyRequest encodes one config key request.
func MarshalHostServiceConfigKeyRequest(req *HostServiceConfigKeyRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if value := strings.TrimSpace(req.Key); value != "" {
		content = appendStringField(content, 1, value)
	}
	return content
}

// UnmarshalHostServiceConfigKeyRequest decodes one config key request.
func UnmarshalHostServiceConfigKeyRequest(data []byte) (*HostServiceConfigKeyRequest, error) {
	out := &HostServiceConfigKeyRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode config key request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode config key")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown config key request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceConfigValueResponse encodes one config value response.
func MarshalHostServiceConfigValueResponse(resp *HostServiceConfigValueResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Value != "" {
		content = appendStringField(content, 1, resp.Value)
	}
	if resp.Found {
		content = appendVarintField(content, 2, 1)
	}
	return content
}

// UnmarshalHostServiceConfigValueResponse decodes one config value response.
func UnmarshalHostServiceConfigValueResponse(data []byte) (*HostServiceConfigValueResponse, error) {
	out := &HostServiceConfigValueResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode config value response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode config value")
			}
			out.Value = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode config found flag")
			}
			out.Found = value != 0
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown config value response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
