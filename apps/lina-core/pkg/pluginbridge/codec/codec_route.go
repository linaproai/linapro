// This file implements route snapshot protobuf-wire encoding.

package codec

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// marshalRouteSnapshot encodes matched route metadata into protobuf wire
// fields.
func marshalRouteSnapshot(in *RouteMatchSnapshotV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.Method); value != "" {
		content = appendStringField(content, 1, value)
	}
	if value := strings.TrimSpace(in.PublicPath); value != "" {
		content = appendStringField(content, 2, value)
	}
	if value := strings.TrimSpace(in.InternalPath); value != "" {
		content = appendStringField(content, 3, value)
	}
	if value := strings.TrimSpace(in.RoutePath); value != "" {
		content = appendStringField(content, 4, value)
	}
	if value := strings.TrimSpace(in.Access); value != "" {
		content = appendStringField(content, 5, value)
	}
	if value := strings.TrimSpace(in.Permission); value != "" {
		content = appendStringField(content, 6, value)
	}
	if value := strings.TrimSpace(in.RequestType); value != "" {
		content = appendStringField(content, 7, value)
	}
	if len(in.PathParams) > 0 {
		content = appendStringMap(content, 8, in.PathParams)
	}
	if len(in.QueryValues) > 0 {
		content = appendStringListMap(content, 9, in.QueryValues)
	}
	return content
}

// unmarshalRouteSnapshot decodes matched route metadata from protobuf wire
// fields into the output structure.
func unmarshalRouteSnapshot(content []byte, out *RouteMatchSnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode route snapshot tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot method")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot publicPath")
			}
			out.PublicPath = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot internalPath")
			}
			out.InternalPath = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot routePath")
			}
			out.RoutePath = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot access")
			}
			out.Access = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot permission")
			}
			out.Permission = value
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot requestType")
			}
			out.RequestType = value
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot pathParams")
			}
			if out.PathParams == nil {
				out.PathParams = make(map[string]string)
			}
			if err := unmarshalStringEntry(value, out.PathParams); err != nil {
				return err
			}
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode route snapshot queryValues")
			}
			if out.QueryValues == nil {
				out.QueryValues = make(map[string][]string)
			}
			if err := unmarshalStringListEntry(value, out.QueryValues); err != nil {
				return err
			}
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown route snapshot field")
			}
			content = content[size:]
		}
	}
	return nil
}
