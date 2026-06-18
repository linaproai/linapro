// This file implements the manifest host service for dynamic plugins.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/manifestcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchManifestHostService routes manifest.get calls to the scoped manifest reader.
func dispatchManifestHostService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil || hcc.runtime == nil || hcc.runtime.manifestFactory == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "manifest host service is not configured")
	}
	factory := hcc.runtime.manifestFactory
	if len(hcc.artifactManifestResources) > 0 {
		factory = factory.WithArtifactResources(hcc.pluginID, hcc.artifactManifestResources)
	}
	reader := factory.ForPlugin(hcc.pluginID)

	switch method {
	case bridgehostservice.HostServiceMethodManifestGet:
		request, err := bridgehostservice.UnmarshalHostServiceManifestGetRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if err = validateManifestRequestPath(resourceRef, request.Path); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return handleManifestGet(ctx, reader, request.Path)
	case bridgehostservice.HostServiceMethodManifestGetMany:
		var request manifestGetManyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		for _, path := range request.Paths {
			if err := validateManifestRequestPath(resourceRef, path); err != nil {
				return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
			}
		}
		return domainCapabilityResult(reader.GetMany(ctx, manifestcap.GetManyInput{Paths: request.Paths}))
	case bridgehostservice.HostServiceMethodManifestList:
		var request manifestListRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		if err := validateManifestRequestPath(resourceRef, manifestListResourcePath(request.Prefix)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return domainCapabilityResult(reader.List(ctx, manifestcap.ListInput{Prefix: request.Prefix, Limit: request.Limit}))
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

func validateManifestRequestPath(resourceRef string, path string) error {
	if strings.TrimSpace(resourceRef) == "" {
		return nil
	}
	if matchAuthorizedManifestPath([]string{resourceRef}, path) {
		return nil
	}
	return gerror.New("manifest request target mismatch")
}

func manifestListResourcePath(prefix string) string {
	trimmed := strings.Trim(strings.ReplaceAll(strings.TrimSpace(prefix), "\\", "/"), "/")
	if trimmed == "" {
		return ".manifest-list-probe"
	}
	return trimmed + "/.manifest-list-probe"
}

type manifestGetManyRequest struct {
	Paths []string `json:"paths"`
}

type manifestListRequest struct {
	Prefix string `json:"prefix"`
	Limit  int    `json:"limit,omitempty"`
}
