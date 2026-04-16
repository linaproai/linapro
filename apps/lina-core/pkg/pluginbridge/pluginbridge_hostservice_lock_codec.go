// This file defines the lock host service request and response codecs shared by
// guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceLockAcquireRequest carries one distributed lock acquire request.
type HostServiceLockAcquireRequest struct {
	// LeaseMillis is the requested lease duration in milliseconds.
	LeaseMillis int64 `json:"leaseMillis,omitempty"`
}

// HostServiceLockAcquireResponse carries one distributed lock acquire response.
type HostServiceLockAcquireResponse struct {
	// Acquired reports whether the lock was acquired successfully.
	Acquired bool `json:"acquired"`
	// Ticket is the opaque lock ticket when Acquired is true.
	Ticket string `json:"ticket,omitempty"`
	// ExpireAt is the next expiration time of the held lock.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceLockRenewRequest carries one distributed lock renew request.
type HostServiceLockRenewRequest struct {
	// Ticket is the opaque lock ticket issued at acquire time.
	Ticket string `json:"ticket"`
}

// HostServiceLockRenewResponse carries one distributed lock renew response.
type HostServiceLockRenewResponse struct {
	// ExpireAt is the next expiration time of the renewed lock.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceLockReleaseRequest carries one distributed lock release request.
type HostServiceLockReleaseRequest struct {
	// Ticket is the opaque lock ticket issued at acquire time.
	Ticket string `json:"ticket"`
}

// MarshalHostServiceLockAcquireRequest encodes one lock acquire request.
func MarshalHostServiceLockAcquireRequest(req *HostServiceLockAcquireRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.LeaseMillis != 0 {
		content = appendVarintField(content, 1, uint64(req.LeaseMillis))
	}
	return content
}

// UnmarshalHostServiceLockAcquireRequest decodes one lock acquire request.
func UnmarshalHostServiceLockAcquireRequest(data []byte) (*HostServiceLockAcquireRequest, error) {
	out := &HostServiceLockAcquireRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 lock acquire request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 lock acquire request leaseMillis 失败")
			}
			out.LeaseMillis = int64(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 lock acquire request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceLockAcquireResponse encodes one lock acquire response.
func MarshalHostServiceLockAcquireResponse(resp *HostServiceLockAcquireResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.Acquired {
		content = appendVarintField(content, 1, 1)
	}
	if resp.Ticket != "" {
		content = appendStringField(content, 2, resp.Ticket)
	}
	if resp.ExpireAt != "" {
		content = appendStringField(content, 3, resp.ExpireAt)
	}
	return content
}

// UnmarshalHostServiceLockAcquireResponse decodes one lock acquire response.
func UnmarshalHostServiceLockAcquireResponse(data []byte) (*HostServiceLockAcquireResponse, error) {
	out := &HostServiceLockAcquireResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 lock acquire response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 lock acquire response acquired 失败")
			}
			out.Acquired = value != 0
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 lock acquire response ticket 失败")
			}
			out.Ticket = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 lock acquire response expireAt 失败")
			}
			out.ExpireAt = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 lock acquire response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceLockRenewRequest encodes one lock renew request.
func MarshalHostServiceLockRenewRequest(req *HostServiceLockRenewRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Ticket != "" {
		content = appendStringField(content, 1, req.Ticket)
	}
	return content
}

// UnmarshalHostServiceLockRenewRequest decodes one lock renew request.
func UnmarshalHostServiceLockRenewRequest(data []byte) (*HostServiceLockRenewRequest, error) {
	out := &HostServiceLockRenewRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 lock renew request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 lock renew request ticket 失败")
			}
			out.Ticket = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 lock renew request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceLockRenewResponse encodes one lock renew response.
func MarshalHostServiceLockRenewResponse(resp *HostServiceLockRenewResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.ExpireAt != "" {
		content = appendStringField(content, 1, resp.ExpireAt)
	}
	return content
}

// UnmarshalHostServiceLockRenewResponse decodes one lock renew response.
func UnmarshalHostServiceLockRenewResponse(data []byte) (*HostServiceLockRenewResponse, error) {
	out := &HostServiceLockRenewResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 lock renew response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 lock renew response expireAt 失败")
			}
			out.ExpireAt = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 lock renew response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceLockReleaseRequest encodes one lock release request.
func MarshalHostServiceLockReleaseRequest(req *HostServiceLockReleaseRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Ticket != "" {
		content = appendStringField(content, 1, req.Ticket)
	}
	return content
}

// UnmarshalHostServiceLockReleaseRequest decodes one lock release request.
func UnmarshalHostServiceLockReleaseRequest(data []byte) (*HostServiceLockReleaseRequest, error) {
	out := &HostServiceLockReleaseRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 lock release request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 lock release request ticket 失败")
			}
			out.Ticket = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 lock release request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
