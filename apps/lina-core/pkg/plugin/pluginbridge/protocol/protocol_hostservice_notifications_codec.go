// This file defines notifications host service request and response codecs
// shared by guest SDK helpers and the host-side Wasm dispatcher.

package protocol

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceNotificationsSendRequest carries one unified notification send request.
type HostServiceNotificationsSendRequest struct {
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

// HostServiceNotificationsSendResponse carries one unified notification send result.
type HostServiceNotificationsSendResponse struct {
	// MessageID is the created notify message identifier.
	MessageID int64 `json:"messageId"`
	// DeliveryCount is the number of created delivery rows.
	DeliveryCount int32 `json:"deliveryCount"`
}

// MarshalHostServiceNotificationsSendRequest encodes one notifications send request.
func MarshalHostServiceNotificationsSendRequest(req *HostServiceNotificationsSendRequest) []byte {
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

// UnmarshalHostServiceNotificationsSendRequest decodes one notifications send request.
func UnmarshalHostServiceNotificationsSendRequest(data []byte) (*HostServiceNotificationsSendRequest, error) {
	out := &HostServiceNotificationsSendRequest{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode notifications send request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request title")
			}
			out.Title = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request content")
			}
			out.Content = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request sourceType")
			}
			out.SourceType = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request sourceId")
			}
			out.SourceID = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request categoryCode")
			}
			out.CategoryCode = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request recipientUserId")
			}
			out.RecipientUserIDs = append(out.RecipientUserIDs, int64(value))
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send request payloadJson")
			}
			out.PayloadJSON = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown notifications send request field")
			}
			content = content[size:]
		}
	}
	return out, nil
}

// MarshalHostServiceNotificationsSendResponse encodes one notifications send response.
func MarshalHostServiceNotificationsSendResponse(resp *HostServiceNotificationsSendResponse) []byte {
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

// UnmarshalHostServiceNotificationsSendResponse decodes one notifications send response.
func UnmarshalHostServiceNotificationsSendResponse(data []byte) (*HostServiceNotificationsSendResponse, error) {
	out := &HostServiceNotificationsSendResponse{}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode notifications send response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send response messageId")
			}
			out.MessageID = int64(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode notifications send response deliveryCount")
			}
			out.DeliveryCount = int32(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown notifications send response field")
			}
			content = content[size:]
		}
	}
	return out, nil
}
