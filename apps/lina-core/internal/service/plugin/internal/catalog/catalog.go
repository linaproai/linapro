// Package catalog provides plugin manifest discovery, validation, and
// manifest-owned asset access for the Lina host plugin system.
package catalog

import (
	"context"
	"sync"
)

// ConfigProvider abstracts the configuration dependency needed for manifest scanning.
type ConfigProvider interface {
	// GetPluginDynamicStoragePath returns the filesystem path where runtime wasm
	// artifacts are stored.
	GetPluginDynamicStoragePath(ctx context.Context) string
}

// ManifestReader covers manifest discovery, loading, parsing, and validation.
// Callers that only need to inspect manifests (without touching the registry,
// release rows, or asset paths) should depend on this narrower interface.
type ManifestReader interface {
	// ScanEmbeddedSourceManifests discovers manifests from all registered embedded source plugins.
	ScanEmbeddedSourceManifests() ([]*Manifest, error)
	// ScanManifests merges source-plugin discovery and runtime-wasm discovery
	// into one normalized manifest list used by lifecycle and governance services.
	ScanManifests() ([]*Manifest, error)
	// ScanManifestsByID returns the latest discovered manifests keyed by plugin ID
	// for list and projection paths that need to reuse one discovery result.
	ScanManifestsByID() (map[string]*Manifest, error)
	// ReadSourcePluginManifestContent reads the raw manifest content from an embedded or
	// filesystem-backed source plugin.
	ReadSourcePluginManifestContent(manifest *Manifest) ([]byte, error)
	// ReadSourcePluginAssetContent reads one asset relative path from an embedded or filesystem source plugin.
	ReadSourcePluginAssetContent(manifest *Manifest, relativePath string) (string, error)
	// LoadManifestFromYAML parses a plugin.yaml file at the given path into a Manifest.
	LoadManifestFromYAML(filePath string, manifest *Manifest) error
	// LoadManifestFromArtifactPath loads and validates a dynamic plugin manifest from
	// the given absolute WASM artifact file path.
	LoadManifestFromArtifactPath(artifactPath string) (*Manifest, error)
	// InvalidateManifestCache removes cached manifest projections for one plugin
	// ID. Empty pluginID clears all dynamic manifest cache entries.
	InvalidateManifestCache(pluginID string)
	// GetDesiredManifest returns the latest discovered manifest for the given plugin ID.
	// For dynamic plugins this is the mutable staging artifact stored at the configured
	// runtime storage path. Changes here do not take effect until the reconciler archives
	// the artifact as an active release.
	GetDesiredManifest(pluginID string) (*Manifest, error)
	// ValidateManifest validates required fields and structural constraints in a plugin manifest.
	// For source plugins it additionally checks for go.mod and backend/plugin.go.
	// For dynamic plugins it validates the runtime artifact when present.
	ValidateManifest(manifest *Manifest, filePath string) error
	// ValidateUploadedRuntimeManifest validates the identity fields extracted from a WASM artifact manifest.
	ValidateUploadedRuntimeManifest(manifest *Manifest) error
}

// SQLAssetCatalog covers plugin SQL file path listings across the install,
// uninstall, and mock-data directions plus the corresponding low-level
// directory-scan helpers shared with build tooling.
type SQLAssetCatalog interface {
	// ListInstallSQLPaths returns the ordered install SQL file paths for a source plugin manifest.
	ListInstallSQLPaths(manifest *Manifest) []string
	// ListUninstallSQLPaths returns the ordered uninstall SQL file paths for a source plugin manifest.
	ListUninstallSQLPaths(manifest *Manifest) []string
	// ListMockSQLPaths returns the ordered mock-data SQL file paths for a source plugin manifest.
	// Mock SQL is only loaded when the operator explicitly opts in at install time.
	ListMockSQLPaths(manifest *Manifest) []string
	// HasMockSQLData reports whether the manifest carries any mock-data SQL assets.
	// Used by the management API and frontend to decide whether to expose the
	// "Install mock data" option for the plugin.
	HasMockSQLData(manifest *Manifest) bool
	// DiscoverSQLPaths discovers plugin SQL files by directory convention.
	DiscoverSQLPaths(rootDir string, uninstall bool) []string
	// DiscoverMockSQLPaths discovers plugin mock-data SQL files by directory convention.
	DiscoverMockSQLPaths(rootDir string) []string
}

// FrontendAssetCatalog covers plugin frontend asset path listings (pages and
// slots) plus the corresponding low-level directory-scan helpers.
type FrontendAssetCatalog interface {
	// ListFrontendPagePaths returns the frontend page source paths for a source plugin manifest.
	ListFrontendPagePaths(manifest *Manifest) []string
	// ListFrontendSlotPaths returns the frontend slot source paths for a source plugin manifest.
	ListFrontendSlotPaths(manifest *Manifest) []string
	// DiscoverPagePaths discovers plugin page source files by directory convention.
	DiscoverPagePaths(rootDir string) []string
	// DiscoverSlotPaths discovers plugin slot source files by directory convention.
	DiscoverSlotPaths(rootDir string) []string
}

// ManifestMetadata covers manifest-derived metadata helpers used by the store
// when deriving governance projections from already-discovered manifests.
type ManifestMetadata interface {
	// BuildRegistryChecksum returns a review-friendly checksum derived from the manifest source.
	// For dynamic plugins, the artifact checksum is returned directly. For source plugins the
	// manifest YAML bytes are hashed using SHA-256.
	BuildRegistryChecksum(manifest *Manifest) string
	// BuildPackagePath returns the canonical package path for a manifest used in release rows.
	BuildPackagePath(manifest *Manifest) string
	// RuntimeStorageDir returns the absolute path of the runtime WASM storage directory
	// configured in plugin.dynamic.storagePath.
	RuntimeStorageDir(ctx context.Context) (string, error)
}

// Service composes catalog-owned manifest and manifest-asset capabilities.
type Service interface {
	ManifestReader
	SQLAssetCatalog
	FrontendAssetCatalog
	ManifestMetadata
}

// Ensure serviceImpl satisfies the catalog contract used across plugin sub-packages.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// configSvc provides plugin configuration values.
	configSvc ConfigProvider
	// cacheMu protects immutable manifest read-model entries.
	cacheMu sync.RWMutex
	// sourceManifestCache stores source manifests by plugin ID with manifest-file stat guards.
	sourceManifestCache map[string]*sourceManifestCacheEntry
	// runtimeArtifactCache stores dynamic manifest projections by absolute artifact path.
	runtimeArtifactCache map[string]*runtimeArtifactCacheEntry
	// runtimePluginArtifactIndex maps plugin ID to its latest discovered artifact path.
	runtimePluginArtifactIndex map[string]string
	// parseCounts tracks runtime artifact parse counts for cache-boundary tests.
	parseCounts map[string]int
}

// New creates a new catalog Service with the given configuration provider.
func New(configSvc ConfigProvider) Service {
	return &serviceImpl{
		configSvc:                  configSvc,
		sourceManifestCache:        make(map[string]*sourceManifestCacheEntry),
		runtimeArtifactCache:       make(map[string]*runtimeArtifactCacheEntry),
		runtimePluginArtifactIndex: make(map[string]string),
		parseCounts:                make(map[string]int),
	}
}
