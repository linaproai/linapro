// This file implements request envelope protobuf-wire encoding.

package codec

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// marshalRequestEnvelope encodes a bridge request envelope into protobuf wire
// fields without relying on generated message types.
func marshalRequestEnvelope(in *BridgeRequestEnvelopeV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.PluginID); value != "" {
		content = appendStringField(content, 1, value)
	}
	if in.Route != nil {
		content = appendBytesField(content, 2, marshalRouteSnapshot(in.Route))
	}
	if in.Request != nil {
		content = appendBytesField(content, 3, marshalRequestSnapshot(in.Request))
	}
	if in.Identity != nil {
		content = appendBytesField(content, 4, marshalIdentitySnapshot(in.Identity))
	}
	if value := strings.TrimSpace(in.RequestID); value != "" {
		content = appendStringField(content, 5, value)
	}
	return content
}

// unmarshalRequestEnvelope decodes protobuf wire fields into a bridge request
// envelope in-place.
func unmarshalRequestEnvelope(content []byte, out *BridgeRequestEnvelopeV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode bridge request tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode bridge request pluginId")
			}
			out.PluginID = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge request route")
			}
			out.Route = &RouteMatchSnapshotV1{}
			if err := unmarshalRouteSnapshot(value, out.Route); err != nil {
				return err
			}
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge request request")
			}
			out.Request = &HTTPRequestSnapshotV1{}
			if err := unmarshalRequestSnapshot(value, out.Request); err != nil {
				return err
			}
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge request identity")
			}
			out.Identity = &IdentitySnapshotV1{}
			if err := unmarshalIdentitySnapshot(value, out.Identity); err != nil {
				return err
			}
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode bridge request requestId")
			}
			out.RequestID = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown bridge request field")
			}
			content = content[size:]
		}
	}
	return nil
}
