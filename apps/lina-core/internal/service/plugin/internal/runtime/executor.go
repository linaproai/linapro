// This file defines dynamic route executors and runtime selection for active
// dynamic plugin releases.

package runtime

import (
	"context"
	"net/http"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/pluginbridge"
)

// dynamicRouteExecutor executes one encoded bridge request against one active runtime.
type dynamicRouteExecutor interface {
	Execute(ctx context.Context, manifest *catalog.Manifest, request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error)
}

// dynamicPlaceholderExecutor is the fallback executor returned when no bridge runtime
// is available for the current plugin release.
type dynamicPlaceholderExecutor struct{}

func (e *dynamicPlaceholderExecutor) Execute(
	_ context.Context,
	_ *catalog.Manifest,
	_ *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return pluginbridge.NewFailureResponse(
		http.StatusNotImplemented,
		"BRIDGE_NOT_IMPLEMENTED",
		"Dynamic route bridge is not executable for the active plugin release",
	), nil
}

// dynamicWasmExecutor invokes the wasm bridge for the given plugin manifest.
type dynamicWasmExecutor struct{}

func (e *dynamicWasmExecutor) Execute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return pluginbridge.NewInternalErrorResponse("Dynamic wasm executor: manifest or artifact is nil"), nil
	}
	requestContent, err := pluginbridge.EncodeRequestEnvelope(request)
	if err != nil {
		return nil, err
	}
	routePath := ""
	if request != nil && request.Route != nil {
		routePath = request.Route.RoutePath
	}
	return wasm.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:        manifest.ID,
		ArtifactPath:    manifest.RuntimeArtifact.Path,
		BridgeSpec:      manifest.BridgeSpec,
		Capabilities:    manifest.HostCapabilities,
		HostServices:    manifest.HostServices,
		ExecutionSource: pluginbridge.ExecutionSourceRoute,
		RoutePath:       routePath,
		RequestID:       request.RequestID,
		Identity:        request.Identity,
	}, requestContent)
}

// executeDynamicRoute selects and runs the appropriate executor for the given manifest.
func (s *serviceImpl) executeDynamicRoute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	executor := s.selectDynamicRouteExecutor(manifest)
	return executor.Execute(ctx, manifest, request)
}

// ExecuteDynamicRoute is the exported form of executeDynamicRoute for cross-package access.
func (s *serviceImpl) ExecuteDynamicRoute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return s.executeDynamicRoute(ctx, manifest, request)
}

// selectDynamicRouteExecutor returns the executor appropriate for the manifest's bridge spec.
func (s *serviceImpl) selectDynamicRouteExecutor(manifest *catalog.Manifest) dynamicRouteExecutor {
	if manifest == nil || manifest.BridgeSpec == nil {
		return &dynamicPlaceholderExecutor{}
	}
	if manifest.BridgeSpec.RouteExecution && manifest.BridgeSpec.RuntimeKind == pluginbridge.RuntimeKindWasm {
		return &dynamicWasmExecutor{}
	}
	return &dynamicPlaceholderExecutor{}
}
