// This file implements the read-only config host service for dynamic plugins.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/plugincap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// configHostServiceFactory is the shared plugin-scoped configuration factory
// used by wasm host calls.
var configHostServiceFactory plugincap.ConfigServiceFactory

// ConfigureConfigHostService replaces the plugin-scoped configuration factory
// used by wasm host calls. The factory must be non-nil.
func ConfigureConfigHostService(factory plugincap.ConfigServiceFactory) error {
	if factory == nil {
		return gerror.New("wasm config host service requires a non-nil config factory")
	}
	configHostServiceFactory = factory
	return nil
}

// dispatchConfigHostService routes config host service methods to the generic
// read-only plugin configuration service.
func dispatchConfigHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if configHostServiceFactory == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "config host service is not configured")
	}

	switch method {
	case bridgehostservice.HostServiceMethodConfigGet:
		factory := configHostServiceFactory
		if len(hcc.artifactDefaultConfig) > 0 {
			factory = factory.WithArtifactConfig(hcc.pluginID, hcc.artifactDefaultConfig)
		}
		return handleConfigGet(ctx, factory.ForPlugin(hcc.pluginID), request.Key)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported config host service method: "+method,
		)
	}
}

// handleConfigGet reads one raw configuration value and returns its JSON representation.
func handleConfigGet(ctx context.Context, reader plugincap.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	if reader == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "config host service is not scoped")
	}
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

// configValueResponse wraps one config value response in a host-call success envelope.
func configValueResponse(value string, found bool) *bridgehostcall.HostCallResponseEnvelope {
	payload := bridgehostservice.MarshalHostServiceConfigValueResponse(&bridgehostservice.HostServiceConfigValueResponse{
		Value: value,
		Found: found,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
