// This file implements response and failure protobuf-wire encoding.

package codec

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// marshalResponseEnvelope encodes a bridge response envelope into protobuf
// wire fields.
func marshalResponseEnvelope(in *BridgeResponseEnvelopeV1) []byte {
	var content []byte
	if in.StatusCode != 0 {
		content = appendVarintField(content, 1, uint64(in.StatusCode))
	}
	if value := strings.TrimSpace(in.ContentType); value != "" {
		content = appendStringField(content, 2, value)
	}
	if len(in.Headers) > 0 {
		content = appendHeaderMap(content, 3, in.Headers)
	}
	if len(in.Body) > 0 {
		content = appendBytesField(content, 4, append([]byte(nil), in.Body...))
	}
	if in.Failure != nil {
		content = appendBytesField(content, 5, marshalFailure(in.Failure))
	}
	return content
}

// unmarshalResponseEnvelope decodes protobuf wire fields into a bridge
// response envelope in-place.
func unmarshalResponseEnvelope(content []byte, out *BridgeResponseEnvelopeV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode bridge response tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("failed to decode bridge response statusCode")
			}
			out.StatusCode = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode bridge response contentType")
			}
			out.ContentType = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge response headers")
			}
			if out.Headers == nil {
				out.Headers = make(map[string][]string)
			}
			if err := unmarshalHeaderEntry(value, out.Headers); err != nil {
				return err
			}
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge response body")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode bridge response failure")
			}
			out.Failure = &BridgeFailureV1{}
			if err := unmarshalFailure(value, out.Failure); err != nil {
				return err
			}
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown bridge response field")
			}
			content = content[size:]
		}
	}
	return nil
}

// marshalFailure encodes normalized failure metadata into protobuf wire
// fields.
func marshalFailure(in *BridgeFailureV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.Code); value != "" {
		content = appendStringField(content, 1, value)
	}
	if value := strings.TrimSpace(in.Message); value != "" {
		content = appendStringField(content, 2, value)
	}
	return content
}

// unmarshalFailure decodes normalized failure metadata from protobuf wire
// fields into the output structure.
func unmarshalFailure(content []byte, out *BridgeFailureV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode failure tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode failure code")
			}
			out.Code = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode failure message")
			}
			out.Message = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown failure field")
			}
			content = content[size:]
		}
	}
	return nil
}
