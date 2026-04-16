// Package openapi projects enabled dynamic plugin routes into the host OpenAPI model
// so generated API documentation reflects all active extension routes.

package openapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

// RoutePublicPrefix is the fixed URL prefix under which all dynamic plugin routes are served.
const RoutePublicPrefix = "/api/v1/extensions"

// Service defines the openapi service contract.
type Service interface {
	// ProjectDynamicRoutesToOpenAPI projects currently enabled dynamic plugin routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest scanning and active manifest lookup.
	catalogSvc catalog.Service
}

// New creates a new openapi Service backed by the given catalog service.
func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{catalogSvc: catalogSvc}
}

// ProjectDynamicRoutesToOpenAPI projects currently enabled dynamic plugin routes into the host OpenAPI paths.
func (s *serviceImpl) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}
	if paths == nil {
		return nil
	}

	runtime, err := buildFilterRuntime(ctx, manifests)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
			continue
		}
		if !runtime.isEnabled(manifest.ID) {
			continue
		}
		activeManifest, manifestErr := s.catalogSvc.GetActiveManifest(ctx, manifest.ID)
		if manifestErr != nil || activeManifest == nil {
			continue
		}
		for _, route := range activeManifest.Routes {
			if route == nil {
				continue
			}
			publicPath := BuildRoutePublicPath(activeManifest.ID, route.Path)
			pathItem, ok := paths[publicPath]
			if !ok {
				pathItem = goai.Path{}
			}
			operation := buildRouteOpenAPIOperation(activeManifest.ID, route, activeManifest.BridgeSpec)
			switch strings.ToUpper(strings.TrimSpace(route.Method)) {
			case http.MethodGet:
				pathItem.Get = operation
			case http.MethodPost:
				pathItem.Post = operation
			case http.MethodPut:
				pathItem.Put = operation
			case http.MethodDelete:
				pathItem.Delete = operation
			}
			paths[publicPath] = pathItem
		}
	}
	return nil
}

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

func buildRouteOpenAPIOperation(
	pluginID string,
	route *pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) *goai.Operation {
	if route == nil {
		return nil
	}
	operation := &goai.Operation{
		Tags:        append([]string(nil), route.Tags...),
		Summary:     route.Summary,
		Description: route.Description,
		OperationID: pluginID + "_" + strings.ToLower(route.Method) + "_" + strings.ReplaceAll(strings.Trim(route.Path, "/"), "/", "_"),
		Responses: goai.Responses{
			"500": goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route execution failed"}},
		},
	}
	if bridgeSpec != nil && bridgeSpec.RouteExecution {
		operation.Responses["200"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route response"}}
	} else {
		operation.Responses["501"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route bridge is not executable"}}
	}
	if route.Access == pluginbridge.AccessLogin {
		operation.Security = &goai.SecurityRequirements{{"BearerAuth": {}}}
	}
	return operation
}
