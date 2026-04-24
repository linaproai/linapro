// This file exposes OpenAPI projection methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/openapi"
)

// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
func (s *serviceImpl) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	return s.openapiSvc.ProjectDynamicRoutesToOpenAPI(ctx, paths)
}

// BuildDynamicRoutePublicPath returns the host-visible public path for one
// dynamic plugin route contract.
func BuildDynamicRoutePublicPath(pluginID string, routePath string) string {
	return openapi.BuildRoutePublicPath(pluginID, routePath)
}
