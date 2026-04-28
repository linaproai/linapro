// This file defines the cache host service request and response codecs shared by
// guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// Cache value kind constants describe the concrete payload representation
// carried in cache response snapshots.
const (
	// HostServiceCacheValueKindString identifies string cache values.
	HostServiceCacheValueKindString = 1
	// HostServiceCacheValueKindInt identifies integer cache values.
	HostServiceCacheValueKindInt = 2
)

// HostServiceCacheValue describes one governed cache value snapshot.
type HostServiceCacheValue struct {
	// ValueKind identifies whether the cache value is string or integer based.
	ValueKind int32 `json:"valueKind"`
	// Value is the canonical string representation of the cache value.
	Value string `json:"value"`
	// IntValue is the integer payload when ValueKind is integer.
	IntValue int64 `json:"intValue,omitempty"`
	// ExpireAt is the optional expiration time.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceCacheGetRequest carries one cache read request.
type HostServiceCacheGetRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
}

// HostServiceCacheGetResponse carries one cache read response.
type HostServiceCacheGetResponse struct {
	// Found reports whether the cache entry exists.
	Found bool `json:"found"`
	// Value is the cache value snapshot when Found is true.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheSetRequest carries one cache write request.
type HostServiceCacheSetRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// Value is the string payload to store.
	Value string `json:"value"`
	// ExpireSeconds is the optional expiration duration in seconds. Zero means no expiration.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheSetResponse carries one cache write response.
type HostServiceCacheSetResponse struct {
	// Value is the resulting cache value snapshot.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheDeleteRequest carries one cache delete request.
type HostServiceCacheDeleteRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
}

// HostServiceCacheIncrRequest carries one cache integer increment request.
type HostServiceCacheIncrRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// Delta is the increment delta applied to the current integer value.
	Delta int64 `json:"delta,omitempty"`
	// ExpireSeconds is the optional expiration duration in seconds. Zero means keep the existing policy.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheIncrResponse carries one cache integer increment response.
type HostServiceCacheIncrResponse struct {
	// Value is the resulting cache value snapshot.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheExpireRequest carries one cache expiration update request.
type HostServiceCacheExpireRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// ExpireSeconds is the new expiration duration in seconds. Zero clears the expiration.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheExpireResponse carries one cache expiration update response.
type HostServiceCacheExpireResponse struct {
	// Found reports whether the cache entry exists.
	Found bool `json:"found"`
	// ExpireAt is the optional updated expiration time.
	ExpireAt string `json:"expireAt,omitempty"`
}

// MarshalHostServiceCacheGetRequest encodes one cache get request.
func MarshalHostServiceCacheGetRequest(req *HostServiceCacheGetRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Key != "" {
		content = appendStringField(content, 1, req.Key)
	}
	return content
}

// UnmarshalHostServiceCacheGetRequest decodes one cache get request.
func UnmarshalHostServiceCacheGetRequest(data []byte) (*HostServiceCacheGetRequest, error) {
	out := &HostServiceCacheGetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache get request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache get request key")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache get request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheGetResponse encodes one cache get response.
func MarshalHostServiceCacheGetResponse(resp *HostServiceCacheGetResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if resp.Value != nil {
		content = appendBytesField(content, 2, marshalHostServiceCacheValue(resp.Value))
	}
	return content
}

// UnmarshalHostServiceCacheGetResponse decodes one cache get response.
func UnmarshalHostServiceCacheGetResponse(data []byte) (*HostServiceCacheGetResponse, error) {
	out := &HostServiceCacheGetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache get response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache get response found")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache get response value")
			}
			cacheValue, err := unmarshalHostServiceCacheValue(value)
			if err != nil {
				return nil, err
			}
			out.Value = cacheValue
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache get response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheSetRequest encodes one cache set request.
func MarshalHostServiceCacheSetRequest(req *HostServiceCacheSetRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Key != "" {
		content = appendStringField(content, 1, req.Key)
	}
	if req.Value != "" {
		content = appendStringField(content, 2, req.Value)
	}
	if req.ExpireSeconds != 0 {
		content = appendVarintField(content, 3, uint64(req.ExpireSeconds))
	}
	return content
}

// UnmarshalHostServiceCacheSetRequest decodes one cache set request.
func UnmarshalHostServiceCacheSetRequest(data []byte) (*HostServiceCacheSetRequest, error) {
	out := &HostServiceCacheSetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache set request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache set request key")
			}
			out.Key = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache set request value")
			}
			out.Value = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache set request expireSeconds")
			}
			out.ExpireSeconds = int64(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache set request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheSetResponse encodes one cache set response.
func MarshalHostServiceCacheSetResponse(resp *HostServiceCacheSetResponse) []byte {
	var content []byte
	if resp == nil || resp.Value == nil {
		return content
	}
	return appendBytesField(content, 1, marshalHostServiceCacheValue(resp.Value))
}

// UnmarshalHostServiceCacheSetResponse decodes one cache set response.
func UnmarshalHostServiceCacheSetResponse(data []byte) (*HostServiceCacheSetResponse, error) {
	out := &HostServiceCacheSetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache set response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache set response value")
			}
			cacheValue, err := unmarshalHostServiceCacheValue(value)
			if err != nil {
				return nil, err
			}
			out.Value = cacheValue
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache set response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheDeleteRequest encodes one cache delete request.
func MarshalHostServiceCacheDeleteRequest(req *HostServiceCacheDeleteRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Key != "" {
		content = appendStringField(content, 1, req.Key)
	}
	return content
}

// UnmarshalHostServiceCacheDeleteRequest decodes one cache delete request.
func UnmarshalHostServiceCacheDeleteRequest(data []byte) (*HostServiceCacheDeleteRequest, error) {
	out := &HostServiceCacheDeleteRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache delete request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache delete request key")
			}
			out.Key = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache delete request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheIncrRequest encodes one cache incr request.
func MarshalHostServiceCacheIncrRequest(req *HostServiceCacheIncrRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Key != "" {
		content = appendStringField(content, 1, req.Key)
	}
	if req.Delta != 0 {
		content = appendVarintField(content, 2, uint64(req.Delta))
	}
	if req.ExpireSeconds != 0 {
		content = appendVarintField(content, 3, uint64(req.ExpireSeconds))
	}
	return content
}

// UnmarshalHostServiceCacheIncrRequest decodes one cache incr request.
func UnmarshalHostServiceCacheIncrRequest(data []byte) (*HostServiceCacheIncrRequest, error) {
	out := &HostServiceCacheIncrRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache incr request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache incr request key")
			}
			out.Key = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache incr request delta")
			}
			out.Delta = int64(value)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache incr request expireSeconds")
			}
			out.ExpireSeconds = int64(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache incr request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheIncrResponse encodes one cache incr response.
func MarshalHostServiceCacheIncrResponse(resp *HostServiceCacheIncrResponse) []byte {
	var content []byte
	if resp == nil || resp.Value == nil {
		return content
	}
	return appendBytesField(content, 1, marshalHostServiceCacheValue(resp.Value))
}

// UnmarshalHostServiceCacheIncrResponse decodes one cache incr response.
func UnmarshalHostServiceCacheIncrResponse(data []byte) (*HostServiceCacheIncrResponse, error) {
	out := &HostServiceCacheIncrResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache incr response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache incr response value")
			}
			cacheValue, err := unmarshalHostServiceCacheValue(value)
			if err != nil {
				return nil, err
			}
			out.Value = cacheValue
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache incr response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheExpireRequest encodes one cache expire request.
func MarshalHostServiceCacheExpireRequest(req *HostServiceCacheExpireRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Key != "" {
		content = appendStringField(content, 1, req.Key)
	}
	if req.ExpireSeconds != 0 {
		content = appendVarintField(content, 2, uint64(req.ExpireSeconds))
	}
	return content
}

// UnmarshalHostServiceCacheExpireRequest decodes one cache expire request.
func UnmarshalHostServiceCacheExpireRequest(data []byte) (*HostServiceCacheExpireRequest, error) {
	out := &HostServiceCacheExpireRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache expire request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache expire request key")
			}
			out.Key = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache expire request expireSeconds")
			}
			out.ExpireSeconds = int64(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache expire request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceCacheExpireResponse encodes one cache expire response.
func MarshalHostServiceCacheExpireResponse(resp *HostServiceCacheExpireResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if resp.ExpireAt != "" {
		content = appendStringField(content, 2, resp.ExpireAt)
	}
	return content
}

// UnmarshalHostServiceCacheExpireResponse decodes one cache expire response.
func UnmarshalHostServiceCacheExpireResponse(data []byte) (*HostServiceCacheExpireResponse, error) {
	out := &HostServiceCacheExpireResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache expire response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache expire response found")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache expire response expireAt")
			}
			out.ExpireAt = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache expire response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// marshalHostServiceCacheValue encodes one cache value snapshot into protobuf
// wire fields.
func marshalHostServiceCacheValue(value *HostServiceCacheValue) []byte {
	var content []byte
	if value == nil {
		return content
	}
	if value.ValueKind != 0 {
		content = appendVarintField(content, 1, uint64(value.ValueKind))
	}
	if value.Value != "" {
		content = appendStringField(content, 2, value.Value)
	}
	if value.IntValue != 0 {
		content = appendVarintField(content, 3, uint64(value.IntValue))
	}
	if value.ExpireAt != "" {
		content = appendStringField(content, 4, value.ExpireAt)
	}
	return content
}

// unmarshalHostServiceCacheValue decodes one cache value snapshot from
// protobuf wire fields.
func unmarshalHostServiceCacheValue(data []byte) (*HostServiceCacheValue, error) {
	out := &HostServiceCacheValue{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cache value tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache value kind")
			}
			out.ValueKind = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache value string")
			}
			out.Value = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache value intValue")
			}
			out.IntValue = int64(value)
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cache value expireAt")
			}
			out.ExpireAt = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cache value field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
