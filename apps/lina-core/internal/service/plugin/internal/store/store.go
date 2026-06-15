// Package store owns plugin governance persistence for registry, release,
// authorization, node-state, and review projections. Its public contract returns
// store-owned projections and must not expose generated DAO/DO/Entity models.
package store

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// ManifestCatalog provides read-only manifest helpers needed to derive persisted
// governance snapshots from already-discovered manifests.
type ManifestCatalog interface {
	catalog.SQLAssetCatalog
	catalog.FrontendAssetCatalog

	// BuildRegistryChecksum returns the persisted checksum for one manifest.
	BuildRegistryChecksum(manifest *catalog.Manifest) string
	// BuildPackagePath returns the canonical package path for a manifest release row.
	BuildPackagePath(manifest *catalog.Manifest) string
	// RuntimeStorageDir returns the absolute path of the runtime WASM storage directory.
	RuntimeStorageDir(ctx context.Context) (string, error)
	// LoadManifestFromArtifactPath loads and validates a dynamic plugin manifest from an artifact path.
	LoadManifestFromArtifactPath(artifactPath string) (*catalog.Manifest, error)
}

// NodeIDProvider exposes the current host node identity used for node-state projections.
type NodeIDProvider interface {
	// CurrentNodeID returns the stable identifier of the current host node.
	CurrentNodeID() string
}

// RuntimeStatePatch carries registry runtime-state updates without exposing DO models.
type RuntimeStatePatch struct {
	// DesiredState optionally updates the desired host lifecycle state.
	DesiredState string
	// CurrentState optionally updates the current host lifecycle state.
	CurrentState string
}

// GovernanceSnapshot aggregates the review-oriented governance data shown in the plugin management UI.
type GovernanceSnapshot struct {
	// ReleaseVersion is the version string of the currently active release.
	ReleaseVersion string
	// LifecycleState is the derived lifecycle state key.
	LifecycleState string
	// NodeState is the current node state projection.
	NodeState string
	// ResourceCount is the number of resource reference rows for the active release.
	ResourceCount int
	// MigrationState is the review-friendly migration state key.
	MigrationState string
}

// Service defines the plugin governance persistence contract.
type Service interface {
	// WithStartupDataSnapshot returns a child context carrying plugin startup snapshots.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// GetRegistry returns the plugin registry projection for pluginID, or nil when absent.
	GetRegistry(ctx context.Context, pluginID string) (*PluginRecord, error)
	// RefreshStartupRegistry reloads one registry row and refreshes the startup snapshot when present.
	RefreshStartupRegistry(ctx context.Context, pluginID string) (*PluginRecord, error)
	// ListAllRegistries returns all plugin registry projections ordered by plugin_id.
	ListAllRegistries(ctx context.Context) ([]*PluginRecord, error)
	// SyncManifest creates or updates registry and release rows for one discovered manifest.
	SyncManifest(ctx context.Context, manifest *catalog.Manifest) (*PluginRecord, error)
	// SetPluginStatus updates the enabled flag and stable host state for one plugin.
	SetPluginStatus(ctx context.Context, pluginID string, enabled int) error
	// SetPluginInstalled updates the installed flag and stable lifecycle state for one plugin.
	SetPluginInstalled(ctx context.Context, pluginID string, installed int) error
	// SetRegistryRuntimeState updates transient registry runtime-state fields.
	SetRegistryRuntimeState(ctx context.Context, pluginID string, patch RuntimeStatePatch) error
	// SetAutoEnableForNewTenants updates the platform-owned tenant provisioning policy.
	SetAutoEnableForNewTenants(ctx context.Context, pluginID string, enabled bool) error
	// BuildPluginStatusKey returns the display key for a plugin's status record.
	BuildPluginStatusKey(pluginID string) string
	// SyncRegistryReleaseReference links one registry row to the matching release row when applicable.
	SyncRegistryReleaseReference(ctx context.Context, registry *PluginRecord, manifest *catalog.Manifest) (*PluginRecord, error)
	// SyncMetadata synchronizes release metadata after a manifest or lifecycle change.
	SyncMetadata(ctx context.Context, manifest *catalog.Manifest, registry *PluginRecord, message string) error
	// PromoteSourceRelease switches a source plugin registry row to a discovered release.
	PromoteSourceRelease(ctx context.Context, registry *PluginRecord, manifest *catalog.Manifest, release *ReleaseRecord) (*PluginRecord, error)
	// GetRelease returns the release projection for a plugin ID and version.
	GetRelease(ctx context.Context, pluginID string, version string) (*ReleaseRecord, error)
	// GetReleaseByID returns the release projection with the given primary key.
	GetReleaseByID(ctx context.Context, releaseID int) (*ReleaseRecord, error)
	// RefreshStartupReleaseByID reloads one release row and refreshes the startup snapshot when present.
	RefreshStartupReleaseByID(ctx context.Context, releaseID int) (*ReleaseRecord, error)
	// GetRegistryRelease returns the active release projection for a registry entry.
	GetRegistryRelease(ctx context.Context, registry *PluginRecord) (*ReleaseRecord, error)
	// GetActiveRelease returns the currently active release projection for one plugin.
	GetActiveRelease(ctx context.Context, pluginID string) (*ReleaseRecord, error)
	// LoadReleaseManifest loads the dynamic plugin manifest from a persisted release artifact.
	LoadReleaseManifest(ctx context.Context, release *ReleaseRecord) (*catalog.Manifest, error)
	// UpdateReleaseState transitions a release row to the given status and optional package path.
	UpdateReleaseState(ctx context.Context, releaseID int, status plugintypes.ReleaseStatus, packagePath string) error
	// SyncReleaseMetadata synchronizes one manifest into the release table.
	SyncReleaseMetadata(ctx context.Context, manifest *catalog.Manifest, registry *PluginRecord) error
	// BuildManifestSnapshot returns the persisted manifest snapshot YAML for one manifest.
	BuildManifestSnapshot(manifest *catalog.Manifest) (string, error)
	// BuildPackagePath returns the canonical package path for a manifest release row.
	BuildPackagePath(manifest *catalog.Manifest) string
	// ParseManifestSnapshot unmarshals one persisted manifest snapshot.
	ParseManifestSnapshot(content string) (*ManifestSnapshot, error)
	// PersistReleaseHostServiceAuthorization writes requested and authorized host service snapshots.
	PersistReleaseHostServiceAuthorization(
		ctx context.Context,
		manifest *catalog.Manifest,
		input *HostServiceAuthorizationInput,
	) (*ManifestSnapshot, error)
	// PersistReleaseUninstallPurgePolicy writes one host-confirmed uninstall cleanup policy.
	PersistReleaseUninstallPurgePolicy(
		ctx context.Context,
		release *ReleaseRecord,
		purgeStorageData bool,
	) (*ManifestSnapshot, error)
	// BuildGovernanceSnapshot loads the current governance projection for one plugin version.
	BuildGovernanceSnapshot(
		ctx context.Context,
		pluginID string,
		version string,
		pluginType string,
		installed int,
		enabled int,
	) (*GovernanceSnapshot, error)
	// BuildRuntimeUpgradeState computes one plugin runtime-upgrade projection.
	BuildRuntimeUpgradeState(
		ctx context.Context,
		registry *PluginRecord,
		manifest *catalog.Manifest,
	) (plugintypes.RuntimeUpgradeProjection, error)
	// BuildRuntimeUpgradeFailureWithLatestMigration returns the latest failed upgrade phase for a target release.
	BuildRuntimeUpgradeFailureWithLatestMigration(
		ctx context.Context,
		release *ReleaseRecord,
	) (*plugintypes.RuntimeUpgradeFailure, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest-only helper projections used before persistence.
	catalogSvc ManifestCatalog
	// nodeIDProvider supplies the current node identity for governance projections.
	nodeIDProvider NodeIDProvider
	// cacheMu protects release and YAML snapshot read-model caches.
	cacheMu sync.RWMutex
	// releaseManifestCache stores parsed dynamic release manifests by immutable release identity.
	releaseManifestCache map[string]*releaseManifestCacheEntry
	// manifestSnapshotCache stores parsed YAML snapshots by content hash.
	manifestSnapshotCache map[string]*ManifestSnapshot
}

// New creates a plugin governance store with the given manifest catalog helper.
func New(catalogSvc ManifestCatalog, nodeIDProvider NodeIDProvider) Service {
	return &serviceImpl{
		catalogSvc:            catalogSvc,
		nodeIDProvider:        nodeIDProvider,
		releaseManifestCache:  make(map[string]*releaseManifestCacheEntry),
		manifestSnapshotCache: make(map[string]*ManifestSnapshot),
	}
}

// timePtr returns a pointer to value for generated DO time fields that preserve
// database NULL semantics with *time.Time.
func timePtr(value time.Time) *time.Time {
	return &value
}
