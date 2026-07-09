// This file defines the fixed-prefix dynamic route dispatcher entrypoints and
// core execution pipeline used by dynamic plugin REST execution.

package runtime

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// Request-context keys and sentinel values used by the dynamic route pipeline.
const (
	dynamicRouteCtxVarState    gctx.StrKey = "plugin_dynamic_route_state"
	dynamicRouteCtxVarIdentity gctx.StrKey = "plugin_dynamic_route_identity"
	dynamicRouteCtxVarMetadata gctx.StrKey = "plugin_dynamic_route_metadata"
)

// DynamicRouteDispatchInput describes one host-side dynamic route dispatch call.
type DynamicRouteDispatchInput struct {
	// Request is the original GoFrame request that entered a dynamic plugin public prefix.
	Request *ghttp.Request
}

// dynamicRouteRuntimeState stores the active manifest and route match cached on the request.
type dynamicRouteRuntimeState struct {
	Manifest *catalog.Manifest
	Match    *dynamicRouteMatch
}

// DynamicRouteRuntimeState is the exported form of dynamicRouteRuntimeState for cross-package access.
type DynamicRouteRuntimeState = dynamicRouteRuntimeState

// RegisterDynamicRouteDispatcher binds the fixed-prefix dispatcher into one host
// router group so dynamic routes reuse the standard RouterGroup registration flow.
func (s *serviceImpl) RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup) {
	if group == nil {
		return
	}
	group.ALL("/*dynamicPath", func(r *ghttp.Request) {
		s.handleDynamicRouteRequest(r)
	})
}

// PrepareDynamicRouteMiddleware resolves the active dynamic route contract and
// caches host-owned runtime state on the request before later middlewares run.
func (s *serviceImpl) PrepareDynamicRouteMiddleware(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState, failure, err := s.prepareDynamicRouteRuntime(r.Context(), r)
	if err != nil {
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse(err.Error()))
		r.ExitAll()
		return
	}
	if failure != nil {
		s.writeDynamicRouteResponse(r, failure)
		r.ExitAll()
		return
	}
	setDynamicRouteRuntimeState(r, runtimeState)
	setMetadata(r, buildMetadata(runtimeState))
	r.Middleware.Next()
}

// AuthenticateDynamicRouteMiddleware applies host-owned login and permission
// governance for the matched dynamic route before bridge execution starts.
func (s *serviceImpl) AuthenticateDynamicRouteMiddleware(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState := getDynamicRouteRuntimeState(r)
	if runtimeState == nil {
		s.writeDynamicRouteResponse(
			r,
			bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is missing"),
		)
		r.ExitAll()
		return
	}

	identity, failure, err := s.authorizeDynamicRouteRequest(r.Context(), runtimeState, r)
	if err != nil {
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse(err.Error()))
		r.ExitAll()
		return
	}
	if failure != nil {
		s.writeDynamicRouteResponse(r, failure)
		r.ExitAll()
		return
	}
	if identity != nil {
		setDynamicRouteIdentitySnapshot(r, identity)
	}
	r.Middleware.Next()
}

// handleDynamicRouteRequest executes the prepared dynamic route after earlier
// middleware stages cached route and identity state on the request.
func (s *serviceImpl) handleDynamicRouteRequest(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState := getDynamicRouteRuntimeState(r)
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is missing"))
		r.ExitAll()
		return
	}

	response, err := s.executePreparedDynamicRoute(
		r.Context(),
		runtimeState,
		getDynamicRouteIdentitySnapshot(r),
		r,
	)
	if err != nil {
		response = bridgecodec.NewInternalErrorResponse(err.Error())
	}
	if response == nil {
		response = bridgecodec.NewInternalErrorResponse("Dynamic route dispatcher returned nil response")
	}
	s.writeDynamicRouteResponse(r, response)
	r.ExitAll()
}

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

// prepareDynamicRouteRuntime resolves the active manifest and matched route for
// one incoming fixed-prefix request.
func (s *serviceImpl) prepareDynamicRouteRuntime(
	ctx context.Context,
	request *ghttp.Request,
) (*dynamicRouteRuntimeState, *bridgecontract.BridgeResponseEnvelopeV1, error) {
	if request == nil {
		return nil, bridgecodec.NewBadRequestResponse("Dynamic route request is missing"), nil
	}

	match, err := s.matchDynamicRoute(ctx, request)
	if err != nil {
		return nil, bridgecodec.NewBadRequestResponse(err.Error()), nil
	}
	if match == nil || match.Route == nil {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic route not found"), nil
	}

	manifest := match.Manifest
	if manifest == nil {
		var err error
		manifest, err = s.resolveActiveOrDesiredManifest(ctx, match.PluginID)
		if err != nil {
			return nil, bridgecodec.NewNotFoundResponse(err.Error()), nil
		}
	}
	registry, err := s.storeSvc.GetRegistry(ctx, match.PluginID)
	if err != nil {
		return nil, nil, err
	}
	if registry == nil || registry.Installed != statusflag.Installed.Int() || registry.Status != statusflag.EnabledValue.Int() {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	runtimeState, err := s.storeSvc.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		return nil, nil, err
	}
	if !store.RuntimeStateAllowsBusinessEntry(runtimeState.State) {
		message := bizerr.Format(
			CodePluginRuntimeUpgradeRequired.Fallback(),
			map[string]any{"pluginId": match.PluginID},
		)
		return nil, bridgecodec.NewFailureResponse(
			http.StatusConflict,
			CodePluginRuntimeUpgradeRequired.RuntimeCode(),
			message,
		), nil
	}
	if s.integrationSvc != nil && !s.integrationSvc.CanExposeBusinessEntries(ctx, match.PluginID) {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	return &dynamicRouteRuntimeState{
		Manifest: manifest,
		Match:    match,
	}, nil, nil
}
