// This file aliases shared protobuf-wire helpers used by host-call codecs.

package hostcall

import "lina-core/pkg/pluginbridge/internal/wire"

var (
	appendStringMap      = wire.AppendStringMap
	appendStringField    = wire.AppendStringField
	appendBytesField     = wire.AppendBytesField
	appendVarintField    = wire.AppendVarintField
	unmarshalStringEntry = wire.UnmarshalStringEntry
)
