// This file defines the source-plugin visible dynamic-route contract.

package contract

import "github.com/gogf/gf/v2/net/ghttp"

// DynamicRouteMetadata is the published projection of one matched dynamic route.
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
	// ResponseBody stores the raw bridge response body captured by the runtime dispatcher.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}

// RouteService defines dynamic-route context operations published to source plugins.
type RouteService interface {
	// DynamicRouteMetadata returns metadata attached to the current dynamic-route request.
	DynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata
}
