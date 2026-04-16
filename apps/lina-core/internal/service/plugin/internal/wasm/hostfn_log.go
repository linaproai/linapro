// This file implements the host:log capability handler that forwards
// structured log entries from the WASM guest to the host logger.

package wasm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
)

// handleHostLog processes OpcodeLog requests from the guest.
func handleHostLog(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallLogRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	message := fmt.Sprintf("[plugin:%s] %s", hcc.pluginID, strings.TrimSpace(req.Message))

	// Append structured fields as key=value pairs to the log message.
	if len(req.Fields) > 0 {
		keys := make([]string, 0, len(req.Fields))
		for k := range req.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([]string, 0, len(keys))
		for _, k := range keys {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, req.Fields[k]))
		}
		message = message + " " + strings.Join(pairs, " ")
	}

	switch req.Level {
	case pluginbridge.LogLevelDebug:
		logger.Debug(ctx, message)
	case pluginbridge.LogLevelInfo:
		logger.Info(ctx, message)
	case pluginbridge.LogLevelWarning:
		logger.Warning(ctx, message)
	case pluginbridge.LogLevelError:
		logger.Error(ctx, message)
	default:
		logger.Info(ctx, message)
	}

	return pluginbridge.NewHostCallEmptySuccessResponse()
}
