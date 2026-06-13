// This file stores and reads dynamic route state from the GoFrame request
// context so middleware stages can share resolved route, identity, and metadata.

package runtime

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
)

// DynamicRouteMetadata stores generic metadata synthesized from one matched
// dynamic route for downstream source-plugin middleware.
type DynamicRouteMetadata struct {
	// PluginID is the dynamic plugin that owns the matched route.
	PluginID string
	// Method is the declared dynamic route HTTP method.
	Method string
	// PublicPath is the public host path matched by the request.
	PublicPath string
	// Tags are the route tags declared by the dynamic plugin manifest.
	Tags []string
	// Summary is the route summary declared by the dynamic plugin manifest.
	Summary string
	// Meta contains additional route declaration metadata by source tag name.
	Meta map[string]string
	// ResponseBody stores the raw bridge response body for middleware-side logging.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}

// BuildDynamicRouteMetadata is the exported form of buildDynamicRouteMetadata for cross-package access.
func BuildDynamicRouteMetadata(runtimeState *DynamicRouteRuntimeState) *DynamicRouteMetadata {
	return buildDynamicRouteMetadata(runtimeState)
}

// setDynamicRouteRuntimeState stores the resolved runtime state on the request context.
func setDynamicRouteRuntimeState(request *ghttp.Request, runtimeState *dynamicRouteRuntimeState) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarState, runtimeState)
}

// getDynamicRouteRuntimeState reads the cached runtime state from the request context.
func getDynamicRouteRuntimeState(request *ghttp.Request) *dynamicRouteRuntimeState {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarState).Val()
	if value == nil {
		return nil
	}
	runtimeState, _ := value.(*dynamicRouteRuntimeState)
	return runtimeState
}

// setDynamicRouteIdentitySnapshot stores the resolved identity snapshot on the request.
func setDynamicRouteIdentitySnapshot(request *ghttp.Request, identity *bridgecontract.IdentitySnapshotV1) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarIdentity, identity)
}

// getDynamicRouteIdentitySnapshot reads the cached identity snapshot from the request.
func getDynamicRouteIdentitySnapshot(request *ghttp.Request) *bridgecontract.IdentitySnapshotV1 {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarIdentity).Val()
	if value == nil {
		return nil
	}
	identity, _ := value.(*bridgecontract.IdentitySnapshotV1)
	return identity
}

// setDynamicRouteMetadata stores generic dynamic-route metadata on the request context.
func setDynamicRouteMetadata(request *ghttp.Request, metadata *DynamicRouteMetadata) {
	if request == nil || metadata == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarMetadata, metadata)
}

// buildDynamicRouteMetadata maps matched route declarations into generic
// request metadata for source-plugin middleware.
func buildDynamicRouteMetadata(runtimeState *dynamicRouteRuntimeState) *DynamicRouteMetadata {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Match.Route == nil {
		return nil
	}
	metadata := &DynamicRouteMetadata{
		PluginID:   strings.TrimSpace(runtimeState.Match.PluginID),
		Method:     strings.TrimSpace(runtimeState.Match.Route.Method),
		PublicPath: strings.TrimSpace(runtimeState.Match.PublicPath),
		Tags:       append([]string(nil), runtimeState.Match.Route.Tags...),
		Summary:    strings.TrimSpace(runtimeState.Match.Route.Summary),
		Meta:       cloneStringMap(runtimeState.Match.Route.Meta),
	}
	return metadata
}

// GetDynamicRouteMetadata returns generic dynamic-route metadata attached
// during the host middleware chain.
func GetDynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarMetadata).Val()
	if value == nil {
		return nil
	}
	metadata, _ := value.(*DynamicRouteMetadata)
	return metadata
}
