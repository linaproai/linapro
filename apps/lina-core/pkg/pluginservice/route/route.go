// Package route exposes host dynamic-route context lookups to source plugins
// without publishing any plugin-owned persistence contract.
package route

import (
	"github.com/gogf/gf/v2/net/ghttp"

	internalplugin "lina-core/internal/service/plugin"
)

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

// Service defines dynamic-route context operations published to source plugins.
type Service interface {
	// DynamicRouteMetadata returns metadata attached to the current dynamic-route request.
	DynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata
}

// serviceAdapter bridges internal dynamic-route helpers into the published contract.
type serviceAdapter struct{}

// New creates and returns the published dynamic-route service adapter.
func New() Service {
	return &serviceAdapter{}
}

// DynamicRouteMetadata returns metadata attached to the current dynamic-route request.
func (s *serviceAdapter) DynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata {
	metadata := internalplugin.GetDynamicRouteMetadata(request)
	if metadata == nil {
		return nil
	}
	return &DynamicRouteMetadata{
		PluginID:            metadata.PluginID,
		Method:              metadata.Method,
		PublicPath:          metadata.PublicPath,
		Tags:                append([]string(nil), metadata.Tags...),
		Summary:             metadata.Summary,
		Meta:                cloneStringMap(metadata.Meta),
		ResponseBody:        metadata.ResponseBody,
		ResponseContentType: metadata.ResponseContentType,
	}
}

// cloneStringMap returns a shallow copy of one string map.
func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
