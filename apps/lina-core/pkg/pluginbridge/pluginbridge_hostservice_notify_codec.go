// This file defines the notify host service request and response codecs shared by
// guest SDK helpers and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceNotifySendRequest carries one unified notification send request.
type HostServiceNotifySendRequest struct {
	// Title is the notification title.
	Title string `json:"title"`
	// Content is the notification body content.
	Content string `json:"content"`
	// SourceType is the optional source type override.
	SourceType string `json:"sourceType,omitempty"`
	// SourceID is the optional source record identifier.
	SourceID string `json:"sourceId,omitempty"`
	// CategoryCode is the optional inbox category override.
	CategoryCode string `json:"categoryCode,omitempty"`
	// RecipientUserIDs is the ordered recipient user identifier list.
	RecipientUserIDs []int64 `json:"recipientUserIds,omitempty"`
	// PayloadJSON is the optional JSON-encoded metadata payload.
	PayloadJSON []byte `json:"payloadJson,omitempty"`
}

// HostServiceNotifySendResponse carries one unified notification send result.
type HostServiceNotifySendResponse struct {
	// MessageID is the created notify message identifier.
	MessageID int64 `json:"messageId"`
	// DeliveryCount is the number of created delivery rows.
	DeliveryCount int32 `json:"deliveryCount"`
}

// MarshalHostServiceNotifySendRequest encodes one notify send request.
func MarshalHostServiceNotifySendRequest(req *HostServiceNotifySendRequest) []byte {
	var content []byte
	if req == nil {
		return content
	}
	if req.Title != "" {
		content = appendStringField(content, 1, req.Title)
	}
	if req.Content != "" {
		content = appendStringField(content, 2, req.Content)
	}
	if req.SourceType != "" {
		content = appendStringField(content, 3, req.SourceType)
	}
	if req.SourceID != "" {
		content = appendStringField(content, 4, req.SourceID)
	}
	if req.CategoryCode != "" {
		content = appendStringField(content, 5, req.CategoryCode)
	}
	for _, userID := range req.RecipientUserIDs {
		if userID > 0 {
			content = appendVarintField(content, 6, uint64(userID))
		}
	}
	if len(req.PayloadJSON) > 0 {
		content = appendBytesField(content, 7, req.PayloadJSON)
	}
	return content
}

// UnmarshalHostServiceNotifySendRequest decodes one notify send request.
func UnmarshalHostServiceNotifySendRequest(data []byte) (*HostServiceNotifySendRequest, error) {
	out := &HostServiceNotifySendRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 notify send request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request title 失败")
			}
			out.Title = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request content 失败")
			}
			out.Content = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request sourceType 失败")
			}
			out.SourceType = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request sourceId 失败")
			}
			out.SourceID = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request categoryCode 失败")
			}
			out.CategoryCode = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request recipientUserId 失败")
			}
			out.RecipientUserIDs = append(out.RecipientUserIDs, int64(value))
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send request payloadJson 失败")
			}
			out.PayloadJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 notify send request 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceNotifySendResponse encodes one notify send response.
func MarshalHostServiceNotifySendResponse(resp *HostServiceNotifySendResponse) []byte {
	var content []byte
	if resp == nil {
		return content
	}
	if resp.MessageID != 0 {
		content = appendVarintField(content, 1, uint64(resp.MessageID))
	}
	if resp.DeliveryCount != 0 {
		content = appendVarintField(content, 2, uint64(resp.DeliveryCount))
	}
	return content
}

// UnmarshalHostServiceNotifySendResponse decodes one notify send response.
func UnmarshalHostServiceNotifySendResponse(data []byte) (*HostServiceNotifySendResponse, error) {
	out := &HostServiceNotifySendResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("解析 notify send response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send response messageId 失败")
			}
			out.MessageID = int64(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("解析 notify send response deliveryCount 失败")
			}
			out.DeliveryCount = int32(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("跳过未知 notify send response 字段失败")
			}
			content = content[size:]
		}
	}
	return out, nil
}
