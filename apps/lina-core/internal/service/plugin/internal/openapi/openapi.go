// Package openapi projects enabled dynamic plugin routes into the host OpenAPI model
// so generated API documentation reflects all active extension routes.
package openapi

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/net/goai"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/store"
)

// revisionReader exposes the plugin-runtime cache revision used to partition
// dynamic route documentation projections.
type revisionReader interface {
	// CurrentRevision returns the plugin-runtime revision visible to this process.
	CurrentRevision(ctx context.Context) (int64, error)
}

// Service defines the openapi service contract.
type Service interface {
	// ProjectDynamicRoutesToOpenAPI projects currently enabled dynamic plugin routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
	// InvalidateProjectionCache clears cached dynamic route OpenAPI projections.
	InvalidateProjectionCache(ctx context.Context, reason string)
}

// Ensure serviceImpl satisfies the dynamic-route OpenAPI projection contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest scanning and active manifest lookup.
	catalogSvc catalog.Service
	// storeSvc provides active release projections for dynamic route docs.
	storeSvc store.Service
	// revisionReader provides the current plugin-runtime revision for cache keys.
	revisionReader revisionReader
	// i18nSvc provides locale and bundle-version partitions for cache keys.
	i18nSvc i18nsvc.Service
	// cache stores dynamic-route OpenAPI path projections by runtime and locale.
	cache *projectionCache
}

// New creates a new openapi Service backed by shared catalog and store services.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	revisionReader revisionReader,
	i18nSvc i18nsvc.Service,
) Service {
	return &serviceImpl{
		catalogSvc:     catalogSvc,
		storeSvc:       storeSvc,
		revisionReader: revisionReader,
		i18nSvc:        i18nSvc,
		cache:          newProjectionCache(),
	}
}

// DeferredRevisionReader is a construction-time cycle breaker for root plugin
// services that own the runtime revision controller after sub-services exist.
type DeferredRevisionReader struct {
	mu      sync.RWMutex
	service revisionReader
}

// NewDeferredRevisionReader creates an unbound revision reader.
func NewDeferredRevisionReader() *DeferredRevisionReader {
	return &DeferredRevisionReader{}
}

// Bind connects the runtime revision authority after root service creation.
func (r *DeferredRevisionReader) Bind(service revisionReader) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.service = service
	r.mu.Unlock()
}

// CurrentRevision returns the bound revision, or zero before binding.
func (r *DeferredRevisionReader) CurrentRevision(ctx context.Context) (int64, error) {
	if r == nil {
		return 0, nil
	}
	r.mu.RLock()
	service := r.service
	r.mu.RUnlock()
	if service == nil {
		return 0, nil
	}
	return service.CurrentRevision(ctx)
}
