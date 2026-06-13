// This file implements the manifest host service for dynamic plugins.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/manifestcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchManifestHostService routes manifest.get calls to the scoped manifest reader.
func dispatchManifestHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceManifestGetRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if hcc == nil || hcc.runtime == nil || hcc.runtime.manifestFactory == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "manifest host service is not configured")
	}

	switch method {
	case bridgehostservice.HostServiceMethodManifestGet:
		factory := hcc.runtime.manifestFactory
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
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	payload := bridgehostservice.MarshalHostServiceManifestGetResponse(&bridgehostservice.HostServiceManifestGetResponse{
		Found: len(content) > 0,
		Body:  content,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
