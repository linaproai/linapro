// This file defines dynamic route executors and runtime selection for active
// dynamic plugin releases.

package runtime

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dynamicRouteExecutor executes one encoded bridge request against one active runtime.
type dynamicRouteExecutor interface {
	// Execute runs one bridge request against the selected runtime implementation.
	Execute(ctx context.Context, manifest *catalog.Manifest, request *bridgecontract.BridgeRequestEnvelopeV1) (*bridgecontract.BridgeResponseEnvelopeV1, error)
}

// dynamicPlaceholderExecutor is the fallback executor returned when no bridge runtime
// is available for the current plugin release.
type dynamicPlaceholderExecutor struct{}

// Execute returns a bridge-not-implemented response for non-executable releases.
func (e *dynamicPlaceholderExecutor) Execute(
	_ context.Context,
	_ *catalog.Manifest,
	_ *bridgecontract.BridgeRequestEnvelopeV1,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	return bridgecodec.NewFailureResponse(
		http.StatusNotImplemented,
		"BRIDGE_NOT_IMPLEMENTED",
		"Dynamic route bridge is not executable for the active plugin release",
	), nil
}

// dynamicWasmRuntimeExecutor invokes the injected wasm runtime for the given
// plugin manifest.
type dynamicWasmRuntimeExecutor struct {
	runtime wasm.Runtime
}

// Execute encodes the bridge request and dispatches it into the active WASM guest.
func (e *dynamicWasmRuntimeExecutor) Execute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *bridgecontract.BridgeRequestEnvelopeV1,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	if e == nil || e.runtime == nil {
		return nil, gerror.New("dynamic wasm runtime is not configured")
	}
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return bridgecodec.NewInternalErrorResponse("Dynamic wasm executor: manifest or artifact is nil"), nil
	}
	requestContent, err := bridgecodec.EncodeRequestEnvelope(request)
	if err != nil {
		return nil, err
	}
	routePath := ""
	if request != nil && request.Route != nil {
		routePath = request.Route.RoutePath
	}
	return e.runtime.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:                  manifest.ID,
		ArtifactPath:              manifest.RuntimeArtifact.Path,
		BridgeSpec:                manifest.BridgeSpec,
		Capabilities:              manifest.HostCapabilities,
		HostServices:              manifest.HostServices,
		ArtifactDefaultConfig:     buildArtifactDefaultConfig(manifest),
		ArtifactManifestResources: buildArtifactManifestResources(manifest),
		ExecutionSource:           bridgecontract.ExecutionSourceRoute,
		RoutePath:                 routePath,
		RequestID:                 request.RequestID,
		Identity:                  request.Identity,
	}, requestContent)
}

// executeDynamicRoute selects and runs the appropriate executor for the given manifest.
func (s *serviceImpl) executeDynamicRoute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *bridgecontract.BridgeRequestEnvelopeV1,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	executor := s.selectDynamicRouteExecutor(manifest)
	return executor.Execute(ctx, manifest, request)
}

// ExecuteDynamicRoute is the exported form of executeDynamicRoute for cross-package access.
func (s *serviceImpl) ExecuteDynamicRoute(
	ctx context.Context,
	manifest *catalog.Manifest,
	request *bridgecontract.BridgeRequestEnvelopeV1,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	return s.executeDynamicRoute(ctx, manifest, request)
}

// selectDynamicRouteExecutor returns the executor appropriate for the manifest's bridge spec.
func (s *serviceImpl) selectDynamicRouteExecutor(manifest *catalog.Manifest) dynamicRouteExecutor {
	if manifest == nil || manifest.BridgeSpec == nil {
		return &dynamicPlaceholderExecutor{}
	}
	if manifest.BridgeSpec.RouteExecution && manifest.BridgeSpec.RuntimeKind == bridgecontract.RuntimeKindWasm {
		return &dynamicWasmRuntimeExecutor{runtime: s.wasmRuntime}
	}
	return &dynamicPlaceholderExecutor{}
}
