// This file exposes dynamic-plugin OpenAPI path and operation helper functions
// used by runtime route documentation.

package openapi

import (
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/pkg/pluginbridge"
)

// BuildRoutePublicPath returns the full public URL path for one dynamic plugin route.
func BuildRoutePublicPath(pluginID string, routePath string) string {
	return RoutePublicPrefix + "/" + strings.TrimSpace(pluginID) + NormalizeDynamicRoutePath(routePath)
}

// NormalizeDynamicRoutePath ensures a route path starts with "/" and has no trailing slash.
func NormalizeDynamicRoutePath(path string) string {
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

// BuildRouteOpenAPIOperation is the exported form of buildRouteOpenAPIOperation for cross-package access.
func BuildRouteOpenAPIOperation(pluginID string, route *pluginbridge.RouteContract, bridgeSpec *pluginbridge.BridgeSpec) *goai.Operation {
	return buildRouteOpenAPIOperation(pluginID, route, bridgeSpec)
}
