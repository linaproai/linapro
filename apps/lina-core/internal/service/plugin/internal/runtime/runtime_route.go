// This file defines the fixed-prefix dynamic route dispatcher entrypoints and
// core execution pipeline used by dynamic plugin REST execution.

package runtime

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// Request-context keys and sentinel values used by the dynamic route pipeline.
const (
	dynamicRouteCtxVarState    gctx.StrKey = "plugin_dynamic_route_state"
	dynamicRouteCtxVarIdentity gctx.StrKey = "plugin_dynamic_route_identity"
	dynamicRouteCtxVarMetadata gctx.StrKey = "plugin_dynamic_route_metadata"

	// statusNormal represents the normal/enabled status for role and menu queries.
	statusNormal = 1
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
	setDynamicRouteMetadata(r, buildDynamicRouteMetadata(runtimeState))
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
	if registry == nil || registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
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
	if s.menuFilter != nil && !s.menuFilter.CanExposeBusinessEntries(ctx, match.PluginID) {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	return &dynamicRouteRuntimeState{
		Manifest: manifest,
		Match:    match,
	}, nil, nil
}
