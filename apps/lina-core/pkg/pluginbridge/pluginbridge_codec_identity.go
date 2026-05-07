// This file implements identity snapshot protobuf-wire encoding.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// marshalIdentitySnapshot encodes authenticated identity context into protobuf
// wire fields.
func marshalIdentitySnapshot(in *IdentitySnapshotV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.TokenID); value != "" {
		content = appendStringField(content, 1, value)
	}
	if in.UserID != 0 {
		content = appendVarintField(content, 2, uint64(in.UserID))
	}
	if value := strings.TrimSpace(in.Username); value != "" {
		content = appendStringField(content, 3, value)
	}
	if in.Status != 0 {
		content = appendVarintField(content, 4, uint64(in.Status))
	}
	for _, permission := range in.Permissions {
		if normalized := strings.TrimSpace(permission); normalized != "" {
			content = appendStringField(content, 5, normalized)
		}
	}
	for _, roleName := range in.RoleNames {
		if normalized := strings.TrimSpace(roleName); normalized != "" {
			content = appendStringField(content, 6, normalized)
		}
	}
	if in.IsSuperAdmin {
		content = appendVarintField(content, 7, 1)
	}
	if in.DataScope != 0 {
		content = appendVarintField(content, 8, uint64(in.DataScope))
	}
	if in.DataScopeUnsupported {
		content = appendVarintField(content, 9, 1)
	}
	if in.UnsupportedDataScope != 0 {
		content = appendVarintField(content, 10, uint64(in.UnsupportedDataScope))
	}
	return content
}

// unmarshalIdentitySnapshot decodes authenticated identity context from
// protobuf wire fields.
func unmarshalIdentitySnapshot(content []byte, out *IdentitySnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode identity snapshot tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot tokenId")
			}
			out.TokenID = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot userId")
			}
			out.UserID = int32(value)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot username")
			}
			out.Username = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot status")
			}
			out.Status = int32(value)
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot permissions")
			}
			out.Permissions = append(out.Permissions, value)
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot roleNames")
			}
			out.RoleNames = append(out.RoleNames, value)
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot isSuperAdmin")
			}
			out.IsSuperAdmin = value > 0
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot dataScope")
			}
			out.DataScope = int32(value)
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot dataScopeUnsupported")
			}
			out.DataScopeUnsupported = value > 0
			content = content[size:]
		case 10:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode identity snapshot unsupportedDataScope")
			}
			out.UnsupportedDataScope = int32(value)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown identity snapshot field")
			}
			content = content[size:]
		}
	}
	return nil
}
