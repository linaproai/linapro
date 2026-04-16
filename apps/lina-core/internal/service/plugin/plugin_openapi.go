// This file exposes OpenAPI projection methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/goai"
)

// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
func (s *serviceImpl) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	return s.openapiSvc.ProjectDynamicRoutesToOpenAPI(ctx, paths)
}
