// This file owns direct dynamic route dispatch for tests and internal callers
// that execute the fixed-prefix routing pipeline without a live RouterGroup.

package runtime

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// DispatchDynamicRoute dispatches one public-prefix request into the active release
// of one dynamic plugin. Matching always happens against the archived active manifest
// so staged uploads cannot affect live traffic before reconcile.
func (s *serviceImpl) DispatchDynamicRoute(
	ctx context.Context,
	in *DynamicRouteDispatchInput,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	if in == nil || in.Request == nil {
		return bridgecodec.NewBadRequestResponse("Dynamic route request is missing"), nil
	}

	runtimeState, failure, err := s.prepareDynamicRouteRuntime(ctx, in.Request)
	if err != nil {
		return nil, err
	}
	if failure != nil {
		return failure, nil
	}
	identity, failure, err := s.authorizeDynamicRouteRequest(ctx, runtimeState, in.Request)
	if err != nil {
		return nil, err
	}
	if failure != nil {
		return failure, nil
	}
	return s.executePreparedDynamicRoute(ctx, runtimeState, identity, in.Request)
}

// executePreparedDynamicRoute builds the bridge request envelope and invokes
// the runtime executor for the matched active route.
func (s *serviceImpl) executePreparedDynamicRoute(
	ctx context.Context,
	runtimeState *dynamicRouteRuntimeState,
	identity *bridgecontract.IdentitySnapshotV1,
	request *ghttp.Request,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		return bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is incomplete"), nil
	}

	requestEnvelope, err := s.buildDynamicRouteRequestEnvelopeWithIdentity(
		runtimeState.Match,
		request,
		identity,
	)
	if err != nil {
		return nil, err
	}
	if runtimeState.Manifest.BridgeSpec == nil || !runtimeState.Manifest.BridgeSpec.RouteExecution {
		return bridgecodec.NewFailureResponse(
			http.StatusNotImplemented,
			"BRIDGE_NOT_IMPLEMENTED",
			"Dynamic route bridge is not executable for the active plugin release",
		), nil
	}
	return s.executeDynamicRoute(ctx, runtimeState.Manifest, requestEnvelope)
}
