// This file aliases shared protobuf-wire helpers used by bridge envelope codecs.

package codec

import (
	"lina-core/pkg/plugin/pluginbridge/internal/wire"
)

var (
	appendHeaderMap          = wire.AppendHeaderMap
	appendStringMap          = wire.AppendStringMap
	appendStringListMap      = wire.AppendStringListMap
	appendStringField        = wire.AppendStringField
	appendStringFieldContent = wire.AppendStringFieldContent
	appendBytesField         = wire.AppendBytesField
	appendVarintField        = wire.AppendVarintField
	unmarshalStringEntry     = wire.UnmarshalStringEntry
	unmarshalStringListEntry = wire.UnmarshalStringListEntry
	unmarshalHeaderEntry     = wire.UnmarshalHeaderEntry
)

// EncodeBodyBase64 returns a review-friendly body preview for tests and logs.
func EncodeBodyBase64(body []byte) string {
	return wire.EncodeBodyBase64(body)
}
