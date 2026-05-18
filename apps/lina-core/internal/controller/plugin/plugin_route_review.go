// This file projects dynamic plugin route contracts into API review models used
// by plugin install and enable dialogs.

package plugin

import (
	"strings"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

// buildPluginRouteReviewItems converts current release dynamic route contracts
// into API review items that expose the host-visible public path and key access
// metadata used by governance dialogs.
func buildPluginRouteReviewItems(
	pluginID string,
	routes []*pluginbridge.RouteContract,
) []*v1.PluginRouteReviewItem {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || len(routes) == 0 {
		return nil
	}

	items := make([]*v1.PluginRouteReviewItem, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		items = append(items, &v1.PluginRouteReviewItem{
			Method:      strings.ToUpper(strings.TrimSpace(route.Method)),
			PublicPath:  pluginsvc.BuildDynamicRoutePublicPath(normalizedPluginID, route.Path),
			Access:      strings.TrimSpace(route.Access),
			Permission:  strings.TrimSpace(route.Permission),
			Summary:     strings.TrimSpace(route.Summary),
			Description: strings.TrimSpace(route.Description),
		})
	}
	if len(items) == 0 {
		return nil
	}
	return items
}
