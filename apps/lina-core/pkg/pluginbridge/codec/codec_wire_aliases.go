// This file aliases shared protobuf-wire helpers used by bridge envelope codecs.

package codec

import (
	"google.golang.org/protobuf/encoding/protowire"

	"lina-core/pkg/pluginbridge/internal/wire"
)

var (
	appendHeaderMap         = wire.AppendHeaderMap
	appendStringMap         = wire.AppendStringMap
	appendStringListMap     = wire.AppendStringListMap
	appendStringField       = wire.AppendStringField
	appendStringFieldContent = wire.AppendStringFieldContent
	appendBytesField        = wire.AppendBytesField
	appendVarintField       = wire.AppendVarintField
	unmarshalStringEntry    = wire.UnmarshalStringEntry
	unmarshalStringListEntry = wire.UnmarshalStringListEntry
	unmarshalHeaderEntry    = wire.UnmarshalHeaderEntry
)

// EncodeBodyBase64 returns a review-friendly body preview for tests and logs.
func EncodeBodyBase64(body []byte) string {
	return wire.EncodeBodyBase64(body)
}

// wireFieldNumber keeps protowire imported for documentation tooling in this
// alias file while concrete codecs continue to use protowire directly.
type wireFieldNumber = protowire.Number
