// This file implements the host:log capability handler that forwards
// structured log entries from the WASM guest to the host logger.

package wasm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"lina-core/pkg/logger"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
)

// handleHostLog processes OpcodeLog requests from the guest.
func handleHostLog(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallLogRequest(reqBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
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
	case bridgehostcall.LogLevelDebug:
		logger.Debug(ctx, message)
	case bridgehostcall.LogLevelInfo:
		logger.Info(ctx, message)
	case bridgehostcall.LogLevelWarning:
		logger.Warning(ctx, message)
	case bridgehostcall.LogLevelError:
		logger.Error(ctx, message)
	default:
		logger.Info(ctx, message)
	}

	return bridgehostcall.NewHostCallEmptySuccessResponse()
}
