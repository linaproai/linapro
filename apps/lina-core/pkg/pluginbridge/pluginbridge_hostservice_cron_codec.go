// This file defines the cron registration host service request codec shared by
// the guest SDK and the host-side Wasm dispatcher.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// HostServiceCronRegisterRequest carries one dynamic-plugin cron declaration
// submitted through the governed cron host service.
type HostServiceCronRegisterRequest struct {
	// Contract is the cron declaration being registered for host-side discovery.
	Contract *CronContract `json:"contract,omitempty"`
}

// MarshalHostServiceCronRegisterRequest encodes one cron registration request.
func MarshalHostServiceCronRegisterRequest(req *HostServiceCronRegisterRequest) []byte {
	var content []byte
	if req == nil || req.Contract == nil {
		return content
	}
	if req.Contract.Name != "" {
		content = appendStringField(content, 1, req.Contract.Name)
	}
	if req.Contract.DisplayName != "" {
		content = appendStringField(content, 2, req.Contract.DisplayName)
	}
	if req.Contract.Description != "" {
		content = appendStringField(content, 3, req.Contract.Description)
	}
	if req.Contract.Pattern != "" {
		content = appendStringField(content, 4, req.Contract.Pattern)
	}
	if req.Contract.Timezone != "" {
		content = appendStringField(content, 5, req.Contract.Timezone)
	}
	if req.Contract.Scope != "" {
		content = appendStringField(content, 6, req.Contract.Scope.String())
	}
	if req.Contract.Concurrency != "" {
		content = appendStringField(content, 7, req.Contract.Concurrency.String())
	}
	if req.Contract.MaxConcurrency > 0 {
		content = appendVarintField(content, 8, uint64(req.Contract.MaxConcurrency))
	}
	if req.Contract.TimeoutSeconds > 0 {
		content = appendVarintField(content, 9, uint64(req.Contract.TimeoutSeconds))
	}
	if req.Contract.RequestType != "" {
		content = appendStringField(content, 10, req.Contract.RequestType)
	}
	if req.Contract.InternalPath != "" {
		content = appendStringField(content, 11, req.Contract.InternalPath)
	}
	return content
}

// UnmarshalHostServiceCronRegisterRequest decodes one cron registration request.
func UnmarshalHostServiceCronRegisterRequest(data []byte) (*HostServiceCronRegisterRequest, error) {
	out := &HostServiceCronRegisterRequest{
		Contract: &CronContract{},
	}
	content := data
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return nil, gerror.New("failed to decode cron register request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request name")
			}
			out.Contract.Name = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request displayName")
			}
			out.Contract.DisplayName = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request description")
			}
			out.Contract.Description = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request pattern")
			}
			out.Contract.Pattern = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request timezone")
			}
			out.Contract.Timezone = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request scope")
			}
			out.Contract.Scope = CronScope(value)
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request concurrency")
			}
			out.Contract.Concurrency = CronConcurrency(value)
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request maxConcurrency")
			}
			out.Contract.MaxConcurrency = int(value)
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request timeoutSeconds")
			}
			out.Contract.TimeoutSeconds = int(value)
			content = content[size:]
		case 10:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request requestType")
			}
			out.Contract.RequestType = value
			content = content[size:]
		case 11:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return nil, gerror.New("failed to decode cron register request internalPath")
			}
			out.Contract.InternalPath = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return nil, gerror.New("failed to skip unknown cron register request field")
			}
			content = content[size:]
		}
	}
	if out.Contract.Name == "" && out.Contract.Pattern == "" && out.Contract.InternalPath == "" && out.Contract.RequestType == "" {
		out.Contract = nil
	}
	return out, nil
}
