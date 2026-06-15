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

// projectionCatalog narrows manifest discovery to the OpenAPI projection reads.
type projectionCatalog interface {
	// ScanManifests returns discovered plugin manifests.
	ScanManifests() ([]*catalog.Manifest, error)
	// GetDesiredManifest returns one discovered manifest.
	GetDesiredManifest(pluginID string) (*catalog.Manifest, error)
}

// projectionStore narrows the store dependency to the read-only projections
// needed for dynamic route documentation.
type projectionStore interface {
	// WithStartupDataSnapshot returns a context carrying a reusable store snapshot.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// ListAllRegistries returns current plugin registry projections.
	ListAllRegistries(ctx context.Context) ([]*store.PluginRecord, error)
	// GetRegistry returns one plugin registry projection.
	GetRegistry(ctx context.Context, pluginID string) (*store.PluginRecord, error)
	// GetRegistryRelease returns the active release projection for a registry row.
	GetRegistryRelease(ctx context.Context, registry *store.PluginRecord) (*store.ReleaseRecord, error)
	// LoadReleaseManifest loads one release manifest.
	LoadReleaseManifest(ctx context.Context, release *store.ReleaseRecord) (*catalog.Manifest, error)
}

// RevisionReader exposes the plugin-runtime cache revision used to partition
// dynamic route documentation projections.
type RevisionReader interface {
	// CurrentRevision returns the plugin-runtime revision visible to this process.
	CurrentRevision(ctx context.Context) (int64, error)
}

// LocaleBundleReader exposes the locale and runtime bundle version used to
// isolate OpenAPI plugin projections by request language.
type LocaleBundleReader interface {
	// GetLocale returns the effective request locale.
	GetLocale(ctx context.Context) string
	// BundleVersion returns the runtime translation bundle version for locale.
	BundleVersion(locale string) uint64
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
	catalogSvc projectionCatalog
	// storeSvc provides active release projections for dynamic route docs.
	storeSvc projectionStore
	// revisionReader provides the current plugin-runtime revision for cache keys.
	revisionReader RevisionReader
	// localeReader provides locale and bundle-version partitions for cache keys.
	localeReader LocaleBundleReader
	// cache stores dynamic-route OpenAPI path projections by runtime and locale.
	cache *projectionCache
}

// New creates a new openapi Service backed by shared catalog and store services.
func New(
	catalogSvc projectionCatalog,
	storeSvc projectionStore,
	revisionReader RevisionReader,
	localeReader LocaleBundleReader,
) Service {
	return &serviceImpl{
		catalogSvc:     catalogSvc,
		storeSvc:       storeSvc,
		revisionReader: revisionReader,
		localeReader:   localeReader,
		cache:          newProjectionCache(),
	}
}

// DeferredRevisionReader is a construction-time cycle breaker for root plugin
// services that own the runtime revision controller after sub-services exist.
type DeferredRevisionReader struct {
	mu      sync.RWMutex
	service RevisionReader
}

// NewDeferredRevisionReader creates an unbound revision reader.
func NewDeferredRevisionReader() *DeferredRevisionReader {
	return &DeferredRevisionReader{}
}

// Bind connects the runtime revision authority after root service creation.
func (r *DeferredRevisionReader) Bind(service RevisionReader) {
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

// DefaultLocaleBundleReader provides a stable fallback for tests or callers
// that do not need request-language partitioning.
type DefaultLocaleBundleReader struct{}

// GetLocale returns the host default locale.
func (DefaultLocaleBundleReader) GetLocale(context.Context) string {
	return i18nsvc.DefaultLocale
}

// BundleVersion returns zero because no runtime i18n service is configured.
func (DefaultLocaleBundleReader) BundleVersion(string) uint64 {
	return 0
}
