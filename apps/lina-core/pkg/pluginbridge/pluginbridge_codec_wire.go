// This file provides shared protobuf-wire helpers for bridge codecs.

package pluginbridge

import (
	"encoding/base64"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

// appendHeaderMap appends sorted header entries as repeated embedded messages.
func appendHeaderMap(content []byte, fieldNumber protowire.Number, values map[string][]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringListPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

// appendStringMap appends sorted string map entries as repeated embedded
// messages.
func appendStringMap(content []byte, fieldNumber protowire.Number, values map[string]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

// appendStringListMap appends sorted repeated-string map entries as repeated
// embedded messages.
func appendStringListMap(content []byte, fieldNumber protowire.Number, values map[string][]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringListPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

// marshalStringPair encodes one string map entry into protobuf wire fields.
func marshalStringPair(key string, value string) []byte {
	var content []byte
	content = appendStringField(content, 1, strings.TrimSpace(key))
	content = appendStringField(content, 2, strings.TrimSpace(value))
	return content
}

// marshalStringListPair encodes one repeated-string map entry into protobuf
// wire fields.
func marshalStringListPair(key string, values []string) []byte {
	var content []byte
	content = appendStringField(content, 1, strings.TrimSpace(key))
	for _, value := range values {
		content = appendStringField(content, 2, strings.TrimSpace(value))
	}
	return content
}

// appendStringField appends one string field to the protobuf payload.
func appendStringField(content []byte, fieldNumber protowire.Number, value string) []byte {
	return appendStringFieldContent(content, fieldNumber, value)
}

// appendStringFieldContent appends the provided string content as a protobuf
// bytes field.
func appendStringFieldContent(content []byte, fieldNumber protowire.Number, value string) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.BytesType)
	return protowire.AppendString(content, value)
}

// appendBytesField appends one bytes field to the protobuf payload.
func appendBytesField(content []byte, fieldNumber protowire.Number, value []byte) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.BytesType)
	return protowire.AppendBytes(content, value)
}

// appendVarintField appends one varint field to the protobuf payload.
func appendVarintField(content []byte, fieldNumber protowire.Number, value uint64) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.VarintType)
	return protowire.AppendVarint(content, value)
}

// unmarshalStringEntry decodes one embedded string map entry into the target
// map.
func unmarshalStringEntry(content []byte, output map[string]string) error {
	var (
		key   string
		value string
	)
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode string map entry tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode string map entry key")
			}
			key = item
			content = content[size:]
		case 2:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode string map entry value")
			}
			value = item
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown string map entry field")
			}
			content = content[size:]
		}
	}
	if strings.TrimSpace(key) != "" {
		output[key] = value
	}
	return nil
}

// unmarshalStringListEntry decodes one embedded repeated-string map entry into
// the target map.
func unmarshalStringListEntry(content []byte, output map[string][]string) error {
	var (
		key    string
		values []string
	)
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("failed to decode string list entry tag")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode string list entry key")
			}
			key = item
			content = content[size:]
		case 2:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("failed to decode string list entry value")
			}
			values = append(values, item)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("failed to skip unknown string list entry field")
			}
			content = content[size:]
		}
	}
	if strings.TrimSpace(key) != "" {
		output[key] = append([]string(nil), values...)
	}
	return nil
}

// unmarshalHeaderEntry decodes one header entry into the output header map.
func unmarshalHeaderEntry(content []byte, output map[string][]string) error {
	return unmarshalStringListEntry(content, output)
}

// sortedKeys returns map keys in ascending order so manual protobuf encoding
// stays deterministic.
func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// EncodeBodyBase64 returns a review-friendly body preview for tests and logs.
func EncodeBodyBase64(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(body)
}
