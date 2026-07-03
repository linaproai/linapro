// This file exposes dynamic-plugin OpenAPI public path helpers used by runtime
// route documentation.

package openapi

import (
	"strings"

	"lina-core/pkg/plugin/pluginhost"
)

// BuildRoutePublicPath returns the full public URL path for one dynamic plugin route.
func BuildRoutePublicPath(pluginID string, routePath string) string {
	return pluginhost.PluginAPINamespacePrefix + "/" + strings.TrimSpace(pluginID) + normalizeDynamicRoutePath(routePath)
}

// normalizeDynamicRoutePath ensures a route path starts with "/" and has no trailing slash.
func normalizeDynamicRoutePath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	if len(normalized) > 1 {
		normalized = strings.TrimSuffix(normalized, "/")
	}
	return normalized
}
