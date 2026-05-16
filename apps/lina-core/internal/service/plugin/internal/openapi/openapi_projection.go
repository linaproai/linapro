// This file projects enabled dynamic-plugin route contracts into the host
// OpenAPI path model without mutating plugin runtime state.

package openapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

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

// buildRouteOpenAPIOperation converts one runtime route contract into a host
// OpenAPI operation while reflecting whether the bridge is executable.
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
