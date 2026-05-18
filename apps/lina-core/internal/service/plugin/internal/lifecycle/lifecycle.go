// Package lifecycle implements dynamic plugin install, uninstall, and reconcile
// lifecycle flows together with helpers for resolving runtime plugin resources.
package lifecycle

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
)

// ReconcileProvider abstracts the runtime reconciler so lifecycle can trigger
// plugin convergence without importing the runtime sub-package directly.
type ReconcileProvider interface {
	// ReconcileDynamicPluginRequest submits a desired state transition to the reconciler.
	ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string) error
	// ShouldRefreshInstalledDynamicRelease reports whether the installed release is stale.
	ShouldRefreshInstalledDynamicRelease(ctx context.Context, registry interface{}, manifest *catalog.Manifest) bool
	// EnsureRuntimeArtifactAvailable ensures the WASM artifact is present for lifecycle actions.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
}

// TopologyProvider abstracts the cluster topology status needed by lifecycle flows.
type TopologyProvider interface {
	// IsPrimaryNode reports whether this host instance is the primary cluster node.
	IsPrimaryNode() bool
}

// Service defines the lifecycle service contract.
type Service interface {
	// SetReconciler wires the runtime package's reconcile provider.
	SetReconciler(r ReconcileProvider)
	// SetTopology wires the cluster topology provider.
	SetTopology(t TopologyProvider)
	// Install executes the install lifecycle for a discovered dynamic plugin.
	// Repeated installs are treated as idempotent unless the same version needs a refresh.
	Install(ctx context.Context, pluginID string) error
	// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
	Uninstall(ctx context.Context, pluginID string) error
	// ExecuteManifestSQLFiles executes plugin manifest SQL files and records every attempt
	// in sys_plugin_migration. The mock phase is intentionally excluded from this entry
	// point because mock data must be loaded transactionally via
	// ExecuteManifestMockSQLFilesInTx.
	ExecuteManifestSQLFiles(
		ctx context.Context,
		manifest *catalog.Manifest,
		direction catalog.MigrationDirection,
	) error
	// ExecuteManifestMockSQLFilesInTx executes a plugin's mock-data SQL files inside the
	// caller-supplied transaction (carried via ctx) and records each step in
	// sys_plugin_migration with phase=mock. Caller MUST run this inside
	// dao.SysPluginMigration.Transaction(...) so any failure rolls back the entire load.
	ExecuteManifestMockSQLFilesInTx(
		ctx context.Context,
		manifest *catalog.Manifest,
	) MockSQLExecutionResult
	// ResolveSQLAssets extracts lifecycle SQL either from embedded runtime artifact sections
	// or from source-style directory conventions, while preserving execution order.
	ResolveSQLAssets(
		manifest *catalog.Manifest,
		direction catalog.MigrationDirection,
	) ([]*SQLAsset, error)
	// ResolvePluginSQLAssets resolves SQL assets from the manifest and returns them as catalog.ArtifactSQLAsset
	// slices for callers that expect the catalog asset type rather than lifecycle.SQLAsset.
	ResolvePluginSQLAssets(manifest *catalog.Manifest, direction catalog.MigrationDirection) ([]*catalog.ArtifactSQLAsset, error)
}

// Ensure serviceImpl satisfies the lifecycle contract used by the root facade.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery and registry access.
	catalogSvc catalog.Service
	// reconciler triggers runtime convergence for desired state transitions.
	reconciler ReconcileProvider
	// topology provides cluster topology information.
	topology TopologyProvider
}

// New creates a new lifecycle Service with the given catalog service.
// Call SetReconciler and SetTopology after construction to wire runtime dependencies.
func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{catalogSvc: catalogSvc}
}
