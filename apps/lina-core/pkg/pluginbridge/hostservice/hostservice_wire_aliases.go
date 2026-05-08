// This file aliases shared protobuf-wire helpers used by host-service codecs.

package hostservice

import "lina-core/pkg/pluginbridge/internal/wire"

var (
	appendStringMap      = wire.AppendStringMap
	appendStringField    = wire.AppendStringField
	appendBytesField     = wire.AppendBytesField
	appendVarintField    = wire.AppendVarintField
	unmarshalStringEntry = wire.UnmarshalStringEntry
)
