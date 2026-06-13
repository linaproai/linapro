// Package frontend manages plugin public asset resolution and in-memory dynamic
// frontend bundles. Runtime WASM assets stay cached in memory, but only
// plugin.yaml public_assets declarations are exposed through /x-assets.
package frontend

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/store"
)

// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
type RuntimeFrontendAssetOutput struct {
	// Content is the raw asset body.
	Content []byte
	// ContentType is the HTTP Content-Type header value returned to browsers.
	ContentType string
}

// Service defines the frontend service contract.
type Service interface {
	// EnsureBundleReader returns a BundleReader for the manifest, building and caching the bundle if needed.
	EnsureBundleReader(ctx context.Context, manifest *catalog.Manifest) (BundleReader, error)
	// ValidateRuntimeFrontendMenuBindings verifies that dynamic plugin menus only reference
	// hosted assets that exist in the plugin's in-memory bundle.
	ValidateRuntimeFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error
	// ValidateHostedMenuBindings is the exported form of validateHostedMenuBindings for cross-package access.
	ValidateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error
	// PrewarmRuntimeFrontendBundles rebuilds in-memory frontend bundles for all enabled
	// dynamic plugins during host startup. A single failed preload does not stop the host;
	// errors are collected and returned as one joined error.
	PrewarmRuntimeFrontendBundles(ctx context.Context) error
	// ResolveRuntimeFrontendAsset resolves one declared plugin public asset for public serving.
	ResolveRuntimeFrontendAsset(
		ctx context.Context,
		pluginID string,
		version string,
		relativePath string,
	) (*RuntimeFrontendAssetOutput, error)
	// BuildRuntimeFrontendPublicBaseURL returns the stable public base URL for plugin public assets.
	BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string
	// InvalidateBundle removes all cached bundle entries for the given plugin ID.
	InvalidateBundle(ctx context.Context, pluginID string, reason string)
	// InvalidateAllBundles removes every cached runtime frontend bundle.
	InvalidateAllBundles(ctx context.Context, reason string)
	// EnsureBundle guarantees an in-memory frontend bundle exists for the given manifest,
	// building and caching it if necessary. Returns the bundle for immediate use.
	// This is called by the runtime reconciler to pre-warm bundles after reconciliation.
	EnsureBundle(ctx context.Context, manifest *catalog.Manifest) error
}

// Ensure serviceImpl satisfies the runtime frontend asset contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest asset lookups for plugin public assets.
	catalogSvc catalog.Service
	// storeSvc provides registry and release projections for runtime frontend assets.
	storeSvc store.Service
}

// New creates a frontend Service backed by the shared plugin catalog and store.
func New(catalogSvc catalog.Service, storeSvc store.Service) Service {
	return &serviceImpl{catalogSvc: catalogSvc, storeSvc: storeSvc}
}
