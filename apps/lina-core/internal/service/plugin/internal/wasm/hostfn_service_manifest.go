// This file implements the manifest host service for dynamic plugins.

package wasm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/manifestcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// manifestHostServiceFactory is the shared plugin-scoped manifest service factory.
var manifestHostServiceFactory manifestcap.ServiceFactory

// ConfigureManifestHostService replaces the plugin-scoped manifest service
// factory used by wasm host calls. The factory must be non-nil.
func ConfigureManifestHostService(factory manifestcap.ServiceFactory) error {
	if factory == nil {
		return gerror.New("wasm manifest host service requires a non-nil manifest factory")
	}
	manifestHostServiceFactory = factory
	return nil
}

// dispatchManifestHostService routes manifest.get calls to the scoped manifest reader.
func dispatchManifestHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceManifestGetRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if manifestHostServiceFactory == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "manifest host service is not configured")
	}

	switch method {
	case bridgehostservice.HostServiceMethodManifestGet:
		factory := manifestHostServiceFactory
		if len(hcc.artifactManifestResources) > 0 {
			factory = factory.WithArtifactResources(hcc.pluginID, hcc.artifactManifestResources)
		}
		return handleManifestGet(ctx, factory.ForPlugin(hcc.pluginID), request.Path)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported manifest host service method: "+method,
		)
	}
}

// handleManifestGet reads one manifest resource and returns its bytes.
func handleManifestGet(ctx context.Context, reader manifestcap.Service, resourcePath string) *bridgehostcall.HostCallResponseEnvelope {
	if reader == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "manifest host service is not scoped")
	}
	content, err := reader.Get(ctx, resourcePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	payload := bridgehostservice.MarshalHostServiceManifestGetResponse(&bridgehostservice.HostServiceManifestGetResponse{
		Found: len(content) > 0,
		Body:  content,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
