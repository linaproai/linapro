// Package upgrade owns plugin runtime-upgrade planning and execution
// orchestration for source and dynamic plugins. It keeps the root plugin facade
// focused on public contracts and governance guards while this package owns the
// upgrade state model, preview read paths, execution strategy dispatch, and
// cache publication boundary.
package upgrade

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
)

type (
	// RuntimeUpgradeManifestSnapshot aliases the review-friendly manifest snapshot model.
	RuntimeUpgradeManifestSnapshot = store.ManifestSnapshot

	// RuntimeUpgradeSQLSummary summarizes manifest SQL assets visible to preview.
	RuntimeUpgradeSQLSummary = SQLSummary

	// RuntimeUpgradeHostServicesDiff summarizes service-level hostServices drift.
	RuntimeUpgradeHostServicesDiff = HostServicesDiff

	// RuntimeUpgradeHostServiceChange summarizes one service-level hostServices change.
	RuntimeUpgradeHostServiceChange = HostServiceChange

	// RuntimeUpgradeState identifies whether discovered plugin files match the effective state.
	RuntimeUpgradeState = plugintypes.RuntimeUpgradeState

	// RuntimeUpgradeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
	RuntimeUpgradeAbnormalReason = plugintypes.RuntimeUpgradeAbnormalReason

	// DependencyCheckResult aliases the management-facing dependency status snapshot.
	DependencyCheckResult = plugindep.CheckProjection
)

const (
	// RuntimeUpgradeStateNormal means the effective and discovered metadata are aligned.
	RuntimeUpgradeStateNormal = plugintypes.RuntimeUpgradeStateNormal
	// RuntimeUpgradeStatePendingUpgrade means discovered plugin files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade = plugintypes.RuntimeUpgradeStatePendingUpgrade
	// RuntimeUpgradeStateAbnormal means discovered plugin files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal = plugintypes.RuntimeUpgradeStateAbnormal
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is currently reconciling.
	RuntimeUpgradeStateUpgradeRunning = plugintypes.RuntimeUpgradeStateUpgradeRunning
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed = plugintypes.RuntimeUpgradeStateUpgradeFailed
	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than the DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective = plugintypes.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is not semver-compatible.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed = plugintypes.RuntimeUpgradeAbnormalReasonVersionCompareFailed
)

// RuntimeUpgradePreview describes the side-effect-free plan shown before a
// runtime plugin upgrade is confirmed.
type RuntimeUpgradePreview struct {
	// PluginID is the target plugin identifier.
	PluginID string
	// RuntimeState is the current runtime-upgrade state re-read by the host.
	RuntimeState RuntimeUpgradeState
	// EffectiveVersion is the database-effective version before upgrade.
	EffectiveVersion string
	// DiscoveredVersion is the file-discovered target version.
	DiscoveredVersion string
	// FromManifest is the current effective manifest snapshot.
	FromManifest *RuntimeUpgradeManifestSnapshot
	// ToManifest is the target manifest snapshot discovered from files.
	ToManifest *RuntimeUpgradeManifestSnapshot
	// DependencyCheck contains install and reverse-dependency precheck results.
	DependencyCheck *DependencyCheckResult
	// SQLSummary summarizes target manifest SQL assets without executing them.
	SQLSummary RuntimeUpgradeSQLSummary
	// HostServicesDiff summarizes requested host service changes.
	HostServicesDiff RuntimeUpgradeHostServicesDiff
	// RiskHints lists stable operator-facing risk hint keys.
	RiskHints []string
}

// RuntimeUpgradeOptions captures explicit management confirmations for a
// runtime plugin upgrade request.
type RuntimeUpgradeOptions struct {
	// Confirmed must be true before the host performs upgrade side effects.
	Confirmed bool
	// Authorization optionally carries the hostServices authorization snapshot
	// confirmed for the target dynamic release before it becomes effective.
	Authorization *store.HostServiceAuthorizationInput
}

// RuntimeUpgradeResult describes one completed explicit runtime upgrade action.
type RuntimeUpgradeResult struct {
	// PluginID is the upgraded plugin identifier.
	PluginID string
	// RuntimeState is the post-upgrade runtime state.
	RuntimeState RuntimeUpgradeState
	// EffectiveVersion is the database-effective version after the request.
	EffectiveVersion string
	// DiscoveredVersion is the currently discovered version after the request.
	DiscoveredVersion string
	// FromVersion is the effective version observed before upgrade side effects.
	FromVersion string
	// ToVersion is the target discovered version requested by the operator.
	ToVersion string
	// Executed reports whether the service performed upgrade side effects.
	Executed bool
}

// RuntimeCacheFreshener refreshes process-local plugin runtime caches before
// read-only upgrade status and preview paths consume derived state.
type RuntimeCacheFreshener interface {
	// EnsureRuntimeCacheFresh synchronizes local runtime caches with the shared revision.
	EnsureRuntimeCacheFresh(ctx context.Context) error
}

// CachePublisher publishes plugin-scoped cache changes after upgrade state changes.
type CachePublisher interface {
	// PublishPluginChange publishes a plugin-scoped mutation reason.
	PublishPluginChange(ctx context.Context, pluginID string, pluginType string, reason string) error
	// SyncEnabledSnapshotAndPublishRuntimeChange refreshes local enablement and publishes a scoped mutation.
	SyncEnabledSnapshotAndPublishRuntimeChange(ctx context.Context, pluginID string, reason string) error
}

// MetadataReader reads framework delivery metadata needed by dependency checks.
type MetadataReader interface {
	// GetMetadata returns the current framework delivery metadata.
	GetMetadata(ctx context.Context) *configsvc.MetadataConfig
}

// I18nService localizes upgrade messages and invalidates runtime bundles.
type I18nService interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// Topology reports cluster mode and node identity for upgrade coordination.
type Topology interface {
	// IsEnabled reports whether clustered coordination is enabled.
	IsEnabled() bool
	// NodeID returns the stable identifier of the current node.
	NodeID() string
}

// Service defines unified plugin upgrade governance operations.
type Service interface {
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
	// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one pending plugin.
	PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error)
	// UpgradeSourcePlugin applies one explicit source-plugin runtime upgrade
	// through the unified source strategy and publishes scoped cache changes.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// ExecuteSourcePluginUpgrade runs the source strategy body without public
	// facade governance guards or cache publication. Runtime upgrade dispatchers
	// use it after the root facade has already guarded and published the outer
	// runtime-upgrade result.
	ExecuteSourcePluginUpgrade(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after the root
	// facade has applied platform governance guards.
	ExecuteRuntimeUpgrade(ctx context.Context, pluginID string, options RuntimeUpgradeOptions) (*RuntimeUpgradeResult, error)
}

// Ensure serviceImpl satisfies Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery and manifest asset access.
	catalogSvc catalog.Service
	// storeSvc owns plugin governance persistence and stable projections.
	storeSvc store.Service
	// lifecycleSvc provides install/uninstall and lifecycle precondition orchestration.
	lifecycleSvc lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// migrationSvc executes plugin SQL lifecycle phases and migration ledger writes.
	migrationSvc migration.Service
	// dependencyResolver evaluates install and reverse-dependency decisions.
	dependencyResolver *plugindep.Resolver
	// i18nSvc localizes upgrade result messages.
	i18nSvc I18nService
	// runtimeUpgradeLockStore coordinates explicit runtime upgrades across cluster nodes.
	runtimeUpgradeLockStore coordination.LockStore
	// cachePublisher publishes upgrade-related plugin cache changes through the root facade.
	cachePublisher CachePublisher
	// cacheFreshener refreshes local runtime caches before read-only upgrade paths.
	cacheFreshener RuntimeCacheFreshener
	// topology reports cluster mode and node identity.
	topology Topology
	// metadataSvc reads framework version metadata for dependency checks.
	metadataSvc MetadataReader
	// runtimeUpgradeLocksMu protects process-local runtime-upgrade locks.
	runtimeUpgradeLocksMu sync.Mutex
	// runtimeUpgradeLocks serializes explicit runtime upgrades per plugin in the current process.
	runtimeUpgradeLocks map[string]*sync.Mutex
}

// New creates a unified plugin upgrade governance service with explicit
// startup-owned dependencies. runtimeUpgradeLockStore may be nil in single-node
// deployments; all other dependencies are required so missing wiring fails at
// composition time instead of during an upgrade request.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	lifecycleSvc lifecycle.Service,
	runtimeSvc runtime.Service,
	integrationSvc integration.Service,
	migrationSvc migration.Service,
	dependencyResolver *plugindep.Resolver,
	i18nSvc I18nService,
	runtimeUpgradeLockStore coordination.LockStore,
	cachePublisher CachePublisher,
	cacheFreshener RuntimeCacheFreshener,
	topology Topology,
	metadataSvc MetadataReader,
) (Service, error) {
	if catalogSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil catalog service")
	}
	if storeSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil store service")
	}
	if lifecycleSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil lifecycle service")
	}
	if runtimeSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil runtime service")
	}
	if integrationSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil integration service")
	}
	if migrationSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil migration service")
	}
	if dependencyResolver == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil dependency resolver")
	}
	if i18nSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil i18n service")
	}
	if cachePublisher == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil cache publisher")
	}
	if cacheFreshener == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil cache freshener")
	}
	if metadataSvc == nil {
		return nil, gerror.New("plugin upgrade service requires a non-nil metadata reader")
	}
	return &serviceImpl{
		catalogSvc:              catalogSvc,
		storeSvc:                storeSvc,
		lifecycleSvc:            lifecycleSvc,
		runtimeSvc:              runtimeSvc,
		integrationSvc:          integrationSvc,
		migrationSvc:            migrationSvc,
		dependencyResolver:      dependencyResolver,
		i18nSvc:                 i18nSvc,
		runtimeUpgradeLockStore: runtimeUpgradeLockStore,
		cachePublisher:          cachePublisher,
		cacheFreshener:          cacheFreshener,
		topology:                topology,
		metadataSvc:             metadataSvc,
		runtimeUpgradeLocks:     make(map[string]*sync.Mutex),
	}, nil
}
