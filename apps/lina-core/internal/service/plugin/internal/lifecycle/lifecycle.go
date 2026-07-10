// Package lifecycle owns plugin lifecycle orchestration. During the C-stage
// refactor it is the target owner for install, uninstall, status, source
// lifecycle, startup auto-enable, and tenant lifecycle flows; the root plugin
// facade should keep only governance guards and delegation once each flow moves.
package lifecycle

import (
	"context"

	"lina-core/internal/service/cluster"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// Service defines the lifecycle service contract.
type Service interface {
	// BootstrapAutoEnable ensures configured startup plugins are installed and
	// enabled after the root facade has synchronized discovered manifests.
	BootstrapAutoEnable(ctx context.Context, options BootstrapAutoEnableOptions) error
	// BootstrapBuiltinPlugins ensures discovered built-in source plugins are
	// installed and enabled during startup without using plugin.autoEnable.
	BootstrapBuiltinPlugins(ctx context.Context, options BootstrapBuiltinOptions) error
	// ReconcileAutoEnabledTenantPlugins reconciles auto-enabled tenant-scoped
	// plugins into platform new-tenant provisioning policy.
	ReconcileAutoEnabledTenantPlugins(ctx context.Context, entries []AutoEnableEntry) error
	// Install executes the install lifecycle for a discovered source or dynamic
	// plugin and returns the dependency check result produced before side effects.
	Install(ctx context.Context, pluginID string, options InstallOptions) (*plugindep.CheckProjection, error)
	// Uninstall executes the full uninstall lifecycle for source or dynamic
	// plugins after the root facade has applied platform governance guards.
	Uninstall(ctx context.Context, pluginID string, options UninstallOptions) error
	// UpdateStatus executes enable or disable lifecycle orchestration for source
	// and dynamic plugins after the root facade has applied governance guards.
	UpdateStatus(ctx context.Context, pluginID string, status int, options UpdateStatusOptions) error
	// RegisterLifecycleObserver subscribes one synchronous lifecycle observer
	// and returns its unsubscribe function.
	RegisterLifecycleObserver(observer LifecycleObserver) func()
	// BindUpgrade attaches the composition-root upgrade service after shared
	// cache adapters become available. The root facade must call this during
	// startup before serving upgrade traffic.
	BindUpgrade(upgradeSvc upgrade.Service) error
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*upgrade.SourceUpgradeStatus, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
	// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one pending plugin.
	PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*upgrade.RuntimeUpgradePreview, error)
	// UpgradeSourcePlugin applies one explicit source-plugin runtime upgrade.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*upgrade.SourceUpgradeResult, error)
	// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
	ExecuteRuntimeUpgrade(ctx context.Context, pluginID string, options upgrade.RuntimeUpgradeOptions) (*upgrade.RuntimeUpgradeResult, error)
	// LifecycleService exposes tenant lifecycle governance through the existing
	// plugin capability contract instead of repeating its method set here.
	plugincap.LifecycleService
}

// Ensure serviceImpl satisfies the lifecycle contract used by the root facade.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery and manifest asset access.
	catalogSvc catalog.Service
	// storeSvc owns plugin governance persistence.
	storeSvc store.Service
	// runtimeSvc owns dynamic runtime convergence and lifecycle callback execution.
	runtimeSvc runtime.Service
	// integrationSvc owns menus, resource references, hook dispatch, and enabled snapshots.
	integrationSvc integration.Service
	// migrationSvc executes lifecycle SQL phases and migration ledger writes.
	migrationSvc migration.Service
	// dependencyResolver evaluates lifecycle dependency blockers.
	dependencyResolver *plugindep.Resolver
	// i18nSvc localizes lifecycle veto summaries and diagnostics.
	i18nSvc i18nsvc.Service
	// cachePublisher publishes runtime cache changes after successful lifecycle writes.
	cachePublisher runtime.CacheChangeNotifier
	// topology provides cluster topology information.
	topology cluster.Service
	// tenantSvc provides tenant-governance operations needed by lifecycle flows.
	tenantSvc tenantspi.Service
	// capabilities resolves plugin-scoped host capabilities for source-plugin callbacks.
	capabilities capability.Services
	// lifecycleObservers stores synchronous lifecycle observers for this lifecycle instance.
	lifecycleObservers *lifecycleObserverRegistry
	// upgradeSvc owns source and dynamic upgrade planning/execution after BindUpgrade.
	upgradeSvc upgrade.Service
}

// New creates a new lifecycle Service with explicit runtime dependencies.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	runtimeSvc runtime.Service,
	integrationSvc integration.Service,
	migrationSvc migration.Service,
	dependencyResolver *plugindep.Resolver,
	i18nSvc i18nsvc.Service,
	cachePublisher runtime.CacheChangeNotifier,
	topology cluster.Service,
	tenantSvc tenantspi.Service,
	capabilities capability.Services,
) Service {
	return &serviceImpl{
		catalogSvc:         catalogSvc,
		storeSvc:           storeSvc,
		runtimeSvc:         runtimeSvc,
		integrationSvc:     integrationSvc,
		migrationSvc:       migrationSvc,
		dependencyResolver: dependencyResolver,
		i18nSvc:            i18nSvc,
		cachePublisher:     cachePublisher,
		topology:           topology,
		tenantSvc:          tenantSvc,
		capabilities:       capabilities,
		lifecycleObservers: newLifecycleObserverRegistry(),
	}
}
