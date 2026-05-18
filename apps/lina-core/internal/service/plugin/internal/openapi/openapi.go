// Package openapi projects enabled dynamic plugin routes into the host OpenAPI model
// so generated API documentation reflects all active extension routes.
package openapi

import (
	"context"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/catalog"
)

// RoutePublicPrefix is the fixed URL prefix under which all dynamic plugin routes are served.
const RoutePublicPrefix = "/api/v1/extensions"

// Service defines the openapi service contract.
type Service interface {
	// ProjectDynamicRoutesToOpenAPI projects currently enabled dynamic plugin routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
}

// Ensure serviceImpl satisfies the dynamic-route OpenAPI projection contract.
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
