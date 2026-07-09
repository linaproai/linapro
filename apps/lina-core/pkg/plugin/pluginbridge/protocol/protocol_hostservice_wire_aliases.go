// This file aliases shared protobuf-wire helpers used by host-service codecs.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/wire"

var (
	appendStringMap      = wire.AppendStringMap
	appendStringField    = wire.AppendStringField
	appendBytesField     = wire.AppendBytesField
	appendVarintField    = wire.AppendVarintField
	unmarshalStringEntry = wire.UnmarshalStringEntry
)
