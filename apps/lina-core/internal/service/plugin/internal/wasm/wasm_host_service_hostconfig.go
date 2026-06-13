// This file implements the hostConfig host service for dynamic plugins.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/plugin/capability/hostconfigcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchHostConfigService routes hostConfig.get calls to the host config reader.
func dispatchHostConfigService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if hcc == nil || hcc.runtime == nil || hcc.runtime.hostConfigService == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host config service is not configured")
	}

	switch method {
	case bridgehostservice.HostServiceMethodHostConfigGet:
		return handleHostConfigGet(ctx, hcc.runtime.hostConfigService, request.Key)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported host config service method: "+method,
		)
	}
}

// handleHostConfigGet reads one authorized host config value and returns JSON.
func handleHostConfigGet(ctx context.Context, reader hostconfigcap.Service, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return configValueResponse(string(encoded), true)
}
