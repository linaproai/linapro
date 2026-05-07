// This file implements HTTP request snapshot protobuf-wire encoding.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// marshalRequestSnapshot encodes one sanitized HTTP request snapshot into
// protobuf wire fields.
func marshalRequestSnapshot(in *HTTPRequestSnapshotV1) []byte {
	var content []byte
	appendStringField := func(fieldNumber protowire.Number, value string) {
		if normalized := strings.TrimSpace(value); normalized != "" {
			content = appendStringFieldContent(content, fieldNumber, normalized)
		}
	}
	appendStringField(1, in.Method)
	appendStringField(2, in.PublicPath)
	appendStringField(3, in.InternalPath)
	appendStringField(4, in.RawPath)
	appendStringField(5, in.RawQuery)
	appendStringField(6, in.Host)
	appendStringField(7, in.Scheme)
	appendStringField(8, in.RemoteAddr)
	appendStringField(9, in.ClientIP)
	appendStringField(10, in.ContentType)
	if len(in.Headers) > 0 {
		content = appendHeaderMap(content, 11, in.Headers)
	}
	if len(in.Cookies) > 0 {
		content = appendStringMap(content, 12, in.Cookies)
	}
	if len(in.Body) > 0 {
		content = appendBytesField(content, 13, append([]byte(nil), in.Body...))
	}
	return content
}

// unmarshalRequestSnapshot decodes one sanitized HTTP request snapshot from
// protobuf wire fields.
func unmarshalRequestSnapshot(content []byte, out *HTTPRequestSnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode request snapshot tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot method")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot publicPath")
			}
			out.PublicPath = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot internalPath")
			}
			out.InternalPath = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot rawPath")
			}
			out.RawPath = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot rawQuery")
			}
			out.RawQuery = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot host")
			}
			out.Host = value
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot scheme")
			}
			out.Scheme = value
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot remoteAddr")
			}
			out.RemoteAddr = value
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot clientIp")
			}
			out.ClientIP = value
			content = content[size:]
		case 10:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot contentType")
			}
			out.ContentType = value
			content = content[size:]
		case 11:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot headers")
			}
			if out.Headers == nil {
				out.Headers = make(map[string][]string)
			}
			if err := unmarshalHeaderEntry(value, out.Headers); err != nil {
				return err
			}
			content = content[size:]
		case 12:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot cookies")
			}
			if out.Cookies == nil {
				out.Cookies = make(map[string]string)
			}
			if err := unmarshalStringEntry(value, out.Cookies); err != nil {
				return err
			}
			content = content[size:]
		case 13:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("failed to decode request snapshot body")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown request snapshot field")
			}
			content = content[size:]
		}
	}
	return nil
}
