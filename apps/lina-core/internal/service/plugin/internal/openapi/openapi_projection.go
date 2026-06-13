// This file projects enabled dynamic-plugin route contracts into the host
// OpenAPI path model without mutating plugin runtime state.

package openapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
		if manifest == nil || plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic {
			continue
		}
		if !runtime.isEnabled(manifest.ID) {
			continue
		}
		activeManifest, manifestErr := s.resolveActiveOrDesiredManifest(ctx, manifest.ID)
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

// resolveActiveOrDesiredManifest loads the active release manifest for installed
// dynamic plugins and falls back to the discovered manifest for inactive rows.
func (s *serviceImpl) resolveActiveOrDesiredManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry != nil &&
		plugintypes.NormalizeType(registry.Type) == plugintypes.TypeDynamic &&
		registry.Installed == plugintypes.InstalledYes &&
		registry.ReleaseId > 0 {
		release, releaseErr := s.storeSvc.GetRegistryRelease(ctx, registry)
		if releaseErr != nil || release == nil {
			return nil, releaseErr
		}
		return s.storeSvc.LoadReleaseManifest(ctx, release)
	}
	return s.catalogSvc.GetDesiredManifest(pluginID)
}

// buildRouteOpenAPIOperation converts one runtime route contract into a host
// OpenAPI operation while reflecting whether the bridge is executable.
func buildRouteOpenAPIOperation(
	pluginID string,
	route *protocol.RouteContract,
	bridgeSpec *protocol.BridgeSpec,
) *goai.Operation {
	if route == nil {
		return nil
	}
	operation := &goai.Operation{
		Tags:        append([]string(nil), route.Tags...),
		Summary:     route.Summary,
		Description: route.Description,
		Responses: goai.Responses{
			"500": goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route execution failed"}},
		},
	}
	if bridgeSpec != nil && bridgeSpec.RouteExecution {
		operation.Responses["200"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route response"}}
	} else {
		operation.Responses["501"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route bridge is not executable"}}
	}
	if route.Access == protocol.AccessLogin {
		operation.Security = &goai.SecurityRequirements{{"BearerAuth": {}}}
	}
	return operation
}
