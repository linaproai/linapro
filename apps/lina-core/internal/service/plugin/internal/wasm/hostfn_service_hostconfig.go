// This file implements the hostConfig host service for dynamic plugins.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/hostconfigcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// hostConfigService is the shared host config reader used by wasm host calls.
var hostConfigService hostconfigcap.Service

// ConfigureHostConfigService replaces the host config reader used by wasm host
// calls. The service must be non-nil.
func ConfigureHostConfigService(service hostconfigcap.Service) error {
	if service == nil {
		return gerror.New("wasm host config service requires a non-nil adapter")
	}
	hostConfigService = service
	return nil
}

// dispatchHostConfigService routes hostConfig.get calls to the host config reader.
func dispatchHostConfigService(
	ctx context.Context,
	_ *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if hostConfigService == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host config service is not configured")
	}

	switch method {
	case bridgehostservice.HostServiceMethodHostConfigGet:
		return handleHostConfigGet(ctx, hostConfigService, request.Key)
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
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(string(encoded), true)
}
