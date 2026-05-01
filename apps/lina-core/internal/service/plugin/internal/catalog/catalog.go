// Package catalog provides plugin manifest discovery, registry management,
// release tracking, and governance queries for the Lina host plugin system.
package catalog

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// ConfigProvider abstracts the configuration dependency needed for manifest scanning.
type ConfigProvider interface {
	// GetPluginDynamicStoragePath returns the filesystem path where runtime wasm
	// artifacts are stored.
	GetPluginDynamicStoragePath(ctx context.Context) string
}

// BackendConfigLoader loads plugin backend hook/resource declarations into a manifest.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle (integration → catalog → integration).
type BackendConfigLoader interface {
	// LoadPluginBackendConfig populates Hooks and BackendResources on the given manifest.
	LoadPluginBackendConfig(manifest *Manifest) error
}

// ArtifactParser parses a runtime WASM artifact file and extracts its embedded sections.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle (runtime → catalog → runtime).
type ArtifactParser interface {
	// ParseRuntimeWasmArtifact reads and validates the WASM file at filePath.
	ParseRuntimeWasmArtifact(filePath string) (*ArtifactSpec, error)
	// ParseRuntimeWasmArtifactContent parses a WASM artifact from an in-memory byte slice.
	ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*ArtifactSpec, error)
	// ValidateRuntimeArtifact validates a dynamic plugin source-tree artifact against manifest.
	ValidateRuntimeArtifact(manifest *Manifest, rootDir string) error
}

// DynamicManifestLoader loads the currently active manifest for an installed dynamic plugin.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle.
type DynamicManifestLoader interface {
	// LoadActiveDynamicPluginManifest returns the manifest backed by the active archived release.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*Manifest, error)
}

// NodeStateSyncer synchronizes node-level plugin state records.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle (runtime → catalog → runtime).
type NodeStateSyncer interface {
	// SyncPluginNodeState upserts the node state record for a plugin lifecycle event.
	SyncPluginNodeState(ctx context.Context, pluginID, version string, installed, enabled int, message string) error
	// GetPluginNodeState returns the current node state record for one plugin on one node.
	GetPluginNodeState(ctx context.Context, pluginID, nodeID string) (*entity.SysPluginNodeState, error)
	// CurrentNodeID returns the cluster node identifier for the running host.
	CurrentNodeID() string
}

// MenuSyncer synchronizes plugin-declared menus into the host menu table.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle (integration → catalog → integration).
type MenuSyncer interface {
	// SyncPluginMenusAndPermissions reconciles manifest menus into sys_menu.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *Manifest) error
}

// ResourceRefSyncer synchronizes plugin resource reference records.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle.
type ResourceRefSyncer interface {
	// SyncPluginResourceReferences persists resource reference rows for governance review.
	SyncPluginResourceReferences(ctx context.Context, manifest *Manifest) error
}

// ReleaseStateSyncer synchronizes the active runtime state of a plugin release.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle.
type ReleaseStateSyncer interface {
	// SyncPluginReleaseRuntimeState updates the active release row to reflect registry state.
	SyncPluginReleaseRuntimeState(ctx context.Context, registry *entity.SysPlugin) error
}

// HookDispatcher dispatches plugin lifecycle events to registered hook handlers.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle.
type HookDispatcher interface {
	// DispatchPluginHookEvent fires a lifecycle hook event with the given payload.
	DispatchPluginHookEvent(ctx context.Context, event pluginhost.ExtensionPoint, values map[string]interface{}) error
}

// Wiring covers the post-construction setter methods that connect the catalog
// with sibling packages. Wiring is intentionally separated so test fixtures
// can construct a catalog with only the wiring shape they need to override
// without depending on the full Service surface.
type Wiring interface {
	// SetBackendLoader wires the integration package's backend config loader.
	SetBackendLoader(loader BackendConfigLoader)
	// SetArtifactParser wires the runtime package's WASM artifact parser.
	SetArtifactParser(parser ArtifactParser)
	// SetDynamicManifestLoader wires the runtime package's active manifest loader.
	SetDynamicManifestLoader(loader DynamicManifestLoader)
	// SetNodeStateSyncer wires the runtime package's node state syncer.
	SetNodeStateSyncer(syncer NodeStateSyncer)
	// SetMenuSyncer wires the integration package's menu syncer.
	SetMenuSyncer(syncer MenuSyncer)
	// SetResourceRefSyncer wires the integration package's resource reference syncer.
	SetResourceRefSyncer(syncer ResourceRefSyncer)
	// SetReleaseStateSyncer wires the runtime package's release state syncer.
	SetReleaseStateSyncer(syncer ReleaseStateSyncer)
	// SetHookDispatcher wires the integration package's hook event dispatcher.
	SetHookDispatcher(dispatcher HookDispatcher)
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
	// LoadReleaseManifest loads the dynamic plugin manifest from a persisted release artifact.
	// The package path stored in the release row is resolved to an absolute host path before parsing.
	LoadReleaseManifest(ctx context.Context, release *entity.SysPluginRelease) (*Manifest, error)
	// GetDesiredManifest returns the latest discovered manifest for the given plugin ID.
	// For dynamic plugins this is the mutable staging artifact stored at the configured
	// runtime storage path. Changes here do not take effect until the reconciler archives
	// the artifact as an active release.
	GetDesiredManifest(pluginID string) (*Manifest, error)
	// GetActiveManifest returns the manifest currently in use by the host for serving.
	// For dynamic plugins this reloads from the archived active release so live traffic
	// sees the stable version while staging changes accumulate. Source plugins always
	// return the discovered manifest directly.
	GetActiveManifest(ctx context.Context, pluginID string) (*Manifest, error)
	// ValidateManifest validates required fields and structural constraints in a plugin manifest.
	// For source plugins it additionally checks for go.mod and backend/plugin.go.
	// For dynamic plugins it optionally validates the runtime artifact via ArtifactParser.
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

// Registry covers sys_plugin registry row reads and the lifecycle-state writes
// that orchestrate post-install/post-enable governance projection.
type Registry interface {
	// WithStartupDataSnapshot returns a child context carrying full-table
	// snapshots for small plugin catalog tables during startup reconciliation.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// GetRegistry returns the sys_plugin row for the given plugin ID, or nil if not found.
	GetRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error)
	// ListAllRegistries returns all sys_plugin rows ordered by plugin_id.
	ListAllRegistries(ctx context.Context) ([]*entity.SysPlugin, error)
	// SyncManifest creates or updates the registry row for a discovered manifest and
	// then synchronizes the release metadata snapshot and node state record.
	SyncManifest(ctx context.Context, manifest *Manifest) (*entity.SysPlugin, error)
	// SetPluginStatus updates the enabled flag on a plugin registry row and fires the
	// matching lifecycle hook event, then syncs release state and node state records.
	SetPluginStatus(ctx context.Context, pluginID string, enabled int) error
	// SetPluginInstalled updates the installed flag and derived lifecycle states for one plugin registry row.
	SetPluginInstalled(ctx context.Context, pluginID string, installed int) error
	// BuildPluginStatusKey returns the display key for a plugin's status record.
	BuildPluginStatusKey(pluginID string) string
	// SyncRegistryReleaseReference is the exported form of syncRegistryReleaseReference for
	// use by runtime-level callers that cannot call the private method directly.
	SyncRegistryReleaseReference(
		ctx context.Context,
		registry *entity.SysPlugin,
		manifest *Manifest,
	) (*entity.SysPlugin, error)
	// SyncMetadata orchestrates release metadata, resource reference, and node state
	// synchronization after a manifest or lifecycle change. It is the exported form
	// used by the runtime package after reconciler state transitions.
	SyncMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin, message string) error
}

// ReleaseStore covers sys_plugin_release row reads, writes, and the manifest
// snapshot helpers that bridge in-memory manifests with persisted release rows.
type ReleaseStore interface {
	// GetRelease returns the sys_plugin_release row for a plugin ID + version pair.
	GetRelease(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error)
	// GetReleaseByID returns the sys_plugin_release row with the given primary key.
	GetReleaseByID(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error)
	// GetRegistryRelease returns the active release row for a registry entry, preferring
	// the ReleaseId pointer and falling back to a version lookup.
	GetRegistryRelease(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPluginRelease, error)
	// GetActiveRelease returns the currently active release row for one plugin.
	GetActiveRelease(ctx context.Context, pluginID string) (*entity.SysPluginRelease, error)
	// UpdateReleaseState transitions a release row to the given status and optionally
	// updates its package path.
	UpdateReleaseState(ctx context.Context, releaseID int, status ReleaseStatus, packagePath string) error
	// SyncReleaseMetadata is the exported form of syncReleaseMetadata for runtime callers.
	SyncReleaseMetadata(ctx context.Context, manifest *Manifest, registry *entity.SysPlugin) error
	// BuildManifestSnapshot is the exported form of buildManifestSnapshot for cross-package access.
	BuildManifestSnapshot(manifest *Manifest) (string, error)
	// ParseManifestSnapshot unmarshals one persisted release manifest snapshot.
	ParseManifestSnapshot(content string) (*ManifestSnapshot, error)
	// PersistReleaseHostServiceAuthorization writes the current requested and
	// authorized host service snapshot into the matching release row.
	PersistReleaseHostServiceAuthorization(
		ctx context.Context,
		manifest *Manifest,
		input *HostServiceAuthorizationInput,
	) (*ManifestSnapshot, error)
	// PersistReleaseUninstallPurgePolicy writes one host-confirmed uninstall
	// cleanup policy snapshot into the given release row.
	PersistReleaseUninstallPurgePolicy(
		ctx context.Context,
		release *entity.SysPluginRelease,
		purgeStorageData bool,
	) (*ManifestSnapshot, error)
}

// Governance covers review-friendly checksums, governance projection
// snapshots, and other read-only metadata derivations used by management UIs
// and audit pipelines.
type Governance interface {
	// BuildRegistryChecksum returns a review-friendly checksum derived from the manifest source.
	// For dynamic plugins, the artifact checksum is returned directly. For source plugins the
	// manifest YAML bytes are hashed using SHA-256.
	BuildRegistryChecksum(manifest *Manifest) string
	// BuildGovernanceSnapshot loads the current governance projection for one plugin version.
	BuildGovernanceSnapshot(
		ctx context.Context,
		pluginID string,
		version string,
		pluginType string,
		installed int,
		enabled int,
	) (*GovernanceSnapshot, error)
	// BuildPackagePath returns the canonical package path for a manifest used in release rows.
	BuildPackagePath(manifest *Manifest) string
	// RuntimeStorageDir returns the absolute path of the runtime WASM storage directory
	// configured in plugin.dynamic.storagePath.
	RuntimeStorageDir(ctx context.Context) (string, error)
}

// Service composes every catalog-owned capability into one full surface so the
// existing cross-package callers (`pluginSvc.catalogSvc`, the test harness in
// `testutil`) can keep using a single interface while new callers may depend
// on the narrower interfaces above per the Interface Segregation Principle.
//
// New code that only needs one concern (e.g., manifest reading or SQL asset
// listings) should declare a parameter of the corresponding sub-interface
// rather than the full Service composite.
type Service interface {
	Wiring
	ManifestReader
	SQLAssetCatalog
	FrontendAssetCatalog
	Registry
	ReleaseStore
	Governance
}

// Ensure serviceImpl satisfies the catalog contract used across plugin sub-packages.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// configSvc provides plugin configuration values.
	configSvc ConfigProvider
	// backendLoader loads backend hook/resource declarations into manifests.
	// Set via SetBackendLoader after construction to avoid import cycles.
	backendLoader BackendConfigLoader
	// artifactParser reads and validates WASM artifact files.
	// Set via SetArtifactParser after construction to avoid import cycles.
	artifactParser ArtifactParser
	// dynamicManifestLoader loads the active release manifest for dynamic plugins.
	// Set via SetDynamicManifestLoader after construction to avoid import cycles.
	dynamicManifestLoader DynamicManifestLoader
	// nodeStateSyncer syncs node state records for lifecycle events.
	// Set via SetNodeStateSyncer after construction to avoid import cycles.
	nodeStateSyncer NodeStateSyncer
	// menuSyncer syncs plugin menus into the host menu table.
	// Set via SetMenuSyncer after construction to avoid import cycles.
	menuSyncer MenuSyncer
	// resourceRefSyncer syncs plugin resource reference records.
	// Set via SetResourceRefSyncer after construction to avoid import cycles.
	resourceRefSyncer ResourceRefSyncer
	// releaseStateSyncer syncs the active runtime state of a plugin release.
	// Set via SetReleaseStateSyncer after construction to avoid import cycles.
	releaseStateSyncer ReleaseStateSyncer
	// hookDispatcher dispatches lifecycle hook events to registered handlers.
	// Set via SetHookDispatcher after construction to avoid import cycles.
	hookDispatcher HookDispatcher
}

// New creates a new catalog Service with the given configuration provider.
// Call the Set* methods after all sub-services are constructed to wire
// the cross-package dependencies.
func New(configSvc ConfigProvider) Service {
	return &serviceImpl{configSvc: configSvc}
}

// SetBackendLoader wires the integration package's backend config loader.
func (s *serviceImpl) SetBackendLoader(loader BackendConfigLoader) {
	s.backendLoader = loader
}

// SetArtifactParser wires the runtime package's WASM artifact parser.
func (s *serviceImpl) SetArtifactParser(parser ArtifactParser) {
	s.artifactParser = parser
}

// SetDynamicManifestLoader wires the runtime package's active manifest loader.
func (s *serviceImpl) SetDynamicManifestLoader(loader DynamicManifestLoader) {
	s.dynamicManifestLoader = loader
}

// SetNodeStateSyncer wires the runtime package's node state syncer.
func (s *serviceImpl) SetNodeStateSyncer(syncer NodeStateSyncer) {
	s.nodeStateSyncer = syncer
}

// SetMenuSyncer wires the integration package's menu syncer.
func (s *serviceImpl) SetMenuSyncer(syncer MenuSyncer) {
	s.menuSyncer = syncer
}

// SetResourceRefSyncer wires the integration package's resource reference syncer.
func (s *serviceImpl) SetResourceRefSyncer(syncer ResourceRefSyncer) {
	s.resourceRefSyncer = syncer
}

// SetReleaseStateSyncer wires the runtime package's release state syncer.
func (s *serviceImpl) SetReleaseStateSyncer(syncer ReleaseStateSyncer) {
	s.releaseStateSyncer = syncer
}

// SetHookDispatcher wires the integration package's hook event dispatcher.
func (s *serviceImpl) SetHookDispatcher(dispatcher HookDispatcher) {
	s.hookDispatcher = dispatcher
}
