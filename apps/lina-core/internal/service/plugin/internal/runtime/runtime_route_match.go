// This file owns dynamic plugin public-path parsing and route-template
// matching. It keeps path governance separate from dispatch and auth logic.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/plugin/internal/catalog"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// dynamicRouteMatch stores the resolved plugin route and path parameters for one request.
type dynamicRouteMatch struct {
	PluginID     string
	PublicPath   string
	InternalPath string
	Route        *bridgecontract.RouteContract
	PathParams   map[string]string
	Manifest     *catalog.Manifest
}

// DynamicRouteMatch is the exported form of dynamicRouteMatch for cross-package access.
type DynamicRouteMatch = dynamicRouteMatch

// MatchDynamicRoutePath is the exported form of matchDynamicRoutePath for cross-package access.
func MatchDynamicRoutePath(routePath string, actualPath string) (map[string]string, bool) {
	return matchDynamicRoutePath(routePath, actualPath)
}

// matchDynamicRoute resolves `/x/{pluginId}/...` public paths to the
// plugin-declared internal route contract. The host owns only the `/x/{pluginId}`
// prefix; every following segment is plugin-defined route content.
func (s *serviceImpl) matchDynamicRoute(ctx context.Context, request *ghttp.Request) (*dynamicRouteMatch, error) {
	publicPath := strings.TrimSpace(request.URL.Path)
	if !strings.HasPrefix(publicPath, pluginhost.PluginAPINamespacePrefix+"/") {
		return nil, nil
	}
	pathSuffix := strings.TrimPrefix(publicPath, pluginhost.PluginAPINamespacePrefix+"/")
	segments := strings.Split(pathSuffix, "/")
	if len(segments) == 0 || strings.TrimSpace(segments[0]) == "" {
		return nil, gerror.New("dynamic plugin path is missing pluginId")
	}
	pluginID := strings.TrimSpace(segments[0])
	internalPath := "/"
	if len(segments) > 1 {
		internalPath = "/" + strings.Join(segments[1:], "/")
	}

	manifest, err := s.resolveActiveOrDesiredManifest(ctx, pluginID)
	if err != nil {
		return nil, nil
	}
	if manifest == nil || len(manifest.Routes) == 0 {
		return nil, nil
	}

	method := strings.ToUpper(strings.TrimSpace(request.Method))
	for _, route := range manifest.Routes {
		params, ok := matchDynamicRoutePath(route.Path, internalPath)
		if !ok {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(route.Method)) != method {
			continue
		}
		return &dynamicRouteMatch{
			PluginID:     pluginID,
			PublicPath:   publicPath,
			InternalPath: internalPath,
			Route:        route,
			PathParams:   params,
			Manifest:     manifest,
		}, nil
	}
	return nil, nil
}

// matchDynamicRoutePath compares one declared route template against the
// actual internal path and returns extracted path params when it matches.
func matchDynamicRoutePath(routePath string, actualPath string) (map[string]string, bool) {
	var (
		normalizedRoute  = normalizeDynamicRoutePath(routePath)
		normalizedActual = normalizeDynamicRoutePath(actualPath)
		routeSegments    = strings.Split(strings.TrimPrefix(normalizedRoute, "/"), "/")
		actualSegments   = strings.Split(strings.TrimPrefix(normalizedActual, "/"), "/")
	)
	if normalizedRoute == "/" {
		routeSegments = []string{}
	}
	if normalizedActual == "/" {
		actualSegments = []string{}
	}
	if len(routeSegments) != len(actualSegments) {
		return nil, false
	}

	params := make(map[string]string)
	for index, routeSegment := range routeSegments {
		actualSegment := actualSegments[index]
		if strings.HasPrefix(routeSegment, "{") && strings.HasSuffix(routeSegment, "}") {
			paramName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(routeSegment, "{"), "}"))
			if paramName == "" {
				return nil, false
			}
			params[paramName] = actualSegment
			continue
		}
		if routeSegment != actualSegment {
			return nil, false
		}
	}
	return params, true
}

// normalizeDynamicRoutePath canonicalizes route paths for matching.
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
