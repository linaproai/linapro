// This file defines storage host service request and response codecs shared by
// guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceStorageObject describes one governed storage object snapshot.
type HostServiceStorageObject struct {
	// Path is the logical object path relative to the storage resource root.
	Path string `json:"path"`
	// Size is the current object size in bytes.
	Size int64 `json:"size"`
	// ContentType is the normalized MIME type for the object.
	ContentType string `json:"contentType,omitempty"`
	// UpdatedAt is the host-side last update timestamp.
	UpdatedAt string `json:"updatedAt,omitempty"`
	// Visibility records the configured resource visibility policy.
	Visibility string `json:"visibility,omitempty"`
}

// HostServiceStoragePutRequest carries one governed storage write request.
type HostServiceStoragePutRequest struct {
	// Path is the logical target path relative to the resource root.
	Path string `json:"path"`
	// Body is the raw object payload.
	Body []byte `json:"body,omitempty"`
	// ContentType is the optional MIME type hint supplied by the guest.
	ContentType string `json:"contentType,omitempty"`
	// Overwrite requests replacement of an existing object at the same path.
	Overwrite bool `json:"overwrite,omitempty"`
}

// HostServiceStoragePutResponse carries storage metadata after a successful write.
type HostServiceStoragePutResponse struct {
	// Object is the resulting object metadata snapshot.
	Object *HostServiceStorageObject `json:"object,omitempty"`
}

// HostServiceStorageGetRequest carries one governed storage read request.
type HostServiceStorageGetRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageGetResponse carries one governed storage read response.
type HostServiceStorageGetResponse struct {
	// Found reports whether the requested object currently exists.
	Found bool `json:"found"`
	// Object is the metadata snapshot when the object exists.
	Object *HostServiceStorageObject `json:"object,omitempty"`
	// Body is the raw object payload when the object exists.
	Body []byte `json:"body,omitempty"`
}

// HostServiceStorageDeleteRequest carries one governed storage delete request.
type HostServiceStorageDeleteRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageListRequest carries one governed storage list request.
type HostServiceStorageListRequest struct {
	// Prefix restricts the result set to one logical object prefix.
	Prefix string `json:"prefix,omitempty"`
	// Limit caps the number of returned objects.
	Limit uint32 `json:"limit,omitempty"`
}

// HostServiceStorageListResponse carries one governed storage list response.
type HostServiceStorageListResponse struct {
	// Objects is the ordered list of visible object metadata snapshots.
	Objects []*HostServiceStorageObject `json:"objects,omitempty"`
}

// HostServiceStorageStatRequest carries one governed storage stat request.
type HostServiceStorageStatRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageStatResponse carries one governed storage stat response.
type HostServiceStorageStatResponse struct {
	// Found reports whether the requested object currently exists.
	Found bool `json:"found"`
	// Object is the metadata snapshot when the object exists.
	Object *HostServiceStorageObject `json:"object,omitempty"`
}

// MarshalHostServiceStoragePutRequest encodes one storage put request.
func MarshalHostServiceStoragePutRequest(req *HostServiceStoragePutRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Path != "" {
		content = appendStringField(content, 1, req.Path)
	}
	if len(req.Body) > 0 {
		content = appendBytesField(content, 2, req.Body)
	}
	if req.ContentType != "" {
		content = appendStringField(content, 3, req.ContentType)
	}
	if req.Overwrite {
		content = appendVarintField(content, 4, 1)
	}
	return content
}

// UnmarshalHostServiceStoragePutRequest decodes one storage put request.
func UnmarshalHostServiceStoragePutRequest(data []byte) (*HostServiceStoragePutRequest, error) {
	out := &HostServiceStoragePutRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage put request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage put request path 失败")
			}
			out.Path = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage put request body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage put request contentType 失败")
			}
			out.ContentType = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 storage put request overwrite 失败")
			}
			out.Overwrite = value != 0
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage put request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStoragePutResponse encodes one storage put response.
func MarshalHostServiceStoragePutResponse(resp *HostServiceStoragePutResponse) []byte {
	var content []byte
	if resp == nil || resp.Object == nil {
		return content
	}
	return appendBytesField(content, 1, marshalHostServiceStorageObject(resp.Object))
}

// UnmarshalHostServiceStoragePutResponse decodes one storage put response.
func UnmarshalHostServiceStoragePutResponse(data []byte) (*HostServiceStoragePutResponse, error) {
	out := &HostServiceStoragePutResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage put response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage put response object 失败")
			}
			object, err := unmarshalHostServiceStorageObject(value)
			if err != nil {
				return nil, err
			}
			out.Object = object
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage put response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageGetRequest encodes one storage get request.
func MarshalHostServiceStorageGetRequest(req *HostServiceStorageGetRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Path != "" {
		content = appendStringField(content, 1, req.Path)
	}
	return content
}

// UnmarshalHostServiceStorageGetRequest decodes one storage get request.
func UnmarshalHostServiceStorageGetRequest(data []byte) (*HostServiceStorageGetRequest, error) {
	out := &HostServiceStorageGetRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage get request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage get request path 失败")
			}
			out.Path = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage get request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageGetResponse encodes one storage get response.
func MarshalHostServiceStorageGetResponse(resp *HostServiceStorageGetResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if resp.Object != nil {
		content = appendBytesField(content, 2, marshalHostServiceStorageObject(resp.Object))
	}
	if len(resp.Body) > 0 {
		content = appendBytesField(content, 3, resp.Body)
	}
	return content
}

// UnmarshalHostServiceStorageGetResponse decodes one storage get response.
func UnmarshalHostServiceStorageGetResponse(data []byte) (*HostServiceStorageGetResponse, error) {
	out := &HostServiceStorageGetResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage get response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 storage get response found 失败")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage get response object 失败")
			}
			object, err := unmarshalHostServiceStorageObject(value)
			if err != nil {
				return nil, err
			}
			out.Object = object
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage get response body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage get response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageDeleteRequest encodes one storage delete request.
func MarshalHostServiceStorageDeleteRequest(req *HostServiceStorageDeleteRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Path != "" {
		content = appendStringField(content, 1, req.Path)
	}
	return content
}

// UnmarshalHostServiceStorageDeleteRequest decodes one storage delete request.
func UnmarshalHostServiceStorageDeleteRequest(data []byte) (*HostServiceStorageDeleteRequest, error) {
	out := &HostServiceStorageDeleteRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage delete request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage delete request path 失败")
			}
			out.Path = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage delete request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageListRequest encodes one storage list request.
func MarshalHostServiceStorageListRequest(req *HostServiceStorageListRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Prefix != "" {
		content = appendStringField(content, 1, req.Prefix)
	}
	if req.Limit > 0 {
		content = appendVarintField(content, 2, uint64(req.Limit))
	}
	return content
}

// UnmarshalHostServiceStorageListRequest decodes one storage list request.
func UnmarshalHostServiceStorageListRequest(data []byte) (*HostServiceStorageListRequest, error) {
	out := &HostServiceStorageListRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage list request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage list request prefix 失败")
			}
			out.Prefix = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 storage list request limit 失败")
			}
			out.Limit = uint32(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage list request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageListResponse encodes one storage list response.
func MarshalHostServiceStorageListResponse(resp *HostServiceStorageListResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	for _, object := range resp.Objects {
		if object == nil {
			continue
		}
		content = appendBytesField(content, 1, marshalHostServiceStorageObject(object))
	}
	return content
}

// UnmarshalHostServiceStorageListResponse decodes one storage list response.
func UnmarshalHostServiceStorageListResponse(data []byte) (*HostServiceStorageListResponse, error) {
	out := &HostServiceStorageListResponse{
		Objects: make([]*HostServiceStorageObject, 0),
	}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage list response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage list response object 失败")
			}
			object, err := unmarshalHostServiceStorageObject(value)
			if err != nil {
				return nil, err
			}
			out.Objects = append(out.Objects, object)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage list response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageStatRequest encodes one storage stat request.
func MarshalHostServiceStorageStatRequest(req *HostServiceStorageStatRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Path != "" {
		content = appendStringField(content, 1, req.Path)
	}
	return content
}

// UnmarshalHostServiceStorageStatRequest decodes one storage stat request.
func UnmarshalHostServiceStorageStatRequest(data []byte) (*HostServiceStorageStatRequest, error) {
	out := &HostServiceStorageStatRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage stat request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage stat request path 失败")
			}
			out.Path = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage stat request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceStorageStatResponse encodes one storage stat response.
func MarshalHostServiceStorageStatResponse(resp *HostServiceStorageStatResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Found {
		content = appendVarintField(content, 1, 1)
	}
	if resp.Object != nil {
		content = appendBytesField(content, 2, marshalHostServiceStorageObject(resp.Object))
	}
	return content
}

// UnmarshalHostServiceStorageStatResponse decodes one storage stat response.
func UnmarshalHostServiceStorageStatResponse(data []byte) (*HostServiceStorageStatResponse, error) {
	out := &HostServiceStorageStatResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage stat response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 storage stat response found 失败")
			}
			out.Found = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 storage stat response object 失败")
			}
			object, err := unmarshalHostServiceStorageObject(value)
			if err != nil {
				return nil, err
			}
			out.Object = object
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage stat response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

func marshalHostServiceStorageObject(object *HostServiceStorageObject) []byte {
	var content []byte
	if object == nil {
		return content
	}
	if object.Path != "" {
		content = appendStringField(content, 1, object.Path)
	}
	if object.Size > 0 {
		content = appendVarintField(content, 2, uint64(object.Size))
	}
	if object.ContentType != "" {
		content = appendStringField(content, 3, object.ContentType)
	}
	if object.UpdatedAt != "" {
		content = appendStringField(content, 4, object.UpdatedAt)
	}
	if object.Visibility != "" {
		content = appendStringField(content, 5, object.Visibility)
	}
	return content
}

func unmarshalHostServiceStorageObject(data []byte) (*HostServiceStorageObject, error) {
	out := &HostServiceStorageObject{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 storage object tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage object path 失败")
			}
			out.Path = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 storage object size 失败")
			}
			out.Size = int64(value)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage object contentType 失败")
			}
			out.ContentType = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage object updatedAt 失败")
			}
			out.UpdatedAt = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 storage object visibility 失败")
			}
			out.Visibility = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 storage object 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
