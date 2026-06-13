// Package lifecycle owns plugin lifecycle orchestration. During the C-stage
// refactor it is the target owner for install, uninstall, status, source
// lifecycle, startup auto-enable, and tenant lifecycle flows; the root plugin
// facade should keep only governance guards and delegation once each flow moves.
package lifecycle

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginhost"
)

// RuntimeOrchestrator defines the runtime capabilities lifecycle needs while
// migrating dynamic install, uninstall, status, tenant, and startup flows.
type RuntimeOrchestrator interface {
	// ReconcileDynamicPluginRequest submits a desired state transition to the reconciler.
	ReconcileDynamicPluginRequest(
		ctx context.Context,
		pluginID string,
		desiredState string,
		options runtime.DynamicReconcileOptions,
	) error
	// ShouldRefreshInstalledDynamicRelease reports whether the installed release is stale.
	ShouldRefreshInstalledDynamicRelease(ctx context.Context, registry interface{}, manifest *catalog.Manifest) bool
	// EnsureRuntimeArtifactAvailable ensures the WASM artifact is present for lifecycle actions.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
	// CheckIsInstalled reports whether one plugin is installed after runtime reconciliation.
	CheckIsInstalled(ctx context.Context, pluginID string) (bool, error)
	// UninstallWithOptions executes dynamic uninstall with an explicit cleanup policy.
	UninstallWithOptions(ctx context.Context, pluginID string, purgeStorageData bool) error
	// ForceUninstallMissingArtifact clears governance for unreadable dynamic artifacts.
	ForceUninstallMissingArtifact(ctx context.Context, registry *store.PluginRecord) error
	// LoadActiveDynamicPluginManifest loads the active dynamic manifest from the stable release.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *store.PluginRecord) (*catalog.Manifest, error)
	// RunDynamicLifecyclePrecondition executes one dynamic Before* lifecycle handler.
	RunDynamicLifecyclePrecondition(
		ctx context.Context,
		manifest *catalog.Manifest,
		input runtime.DynamicLifecycleInput,
	) (*runtime.DynamicLifecycleDecision, error)
	// RunDynamicLifecycleCallback executes one dynamic After* lifecycle handler.
	RunDynamicLifecycleCallback(
		ctx context.Context,
		manifest *catalog.Manifest,
		input runtime.DynamicLifecycleInput,
	) (*runtime.DynamicLifecycleDecision, error)
	// RefreshInstalledRuntimePluginReleases repairs stale installed dynamic releases during startup.
	RefreshInstalledRuntimePluginReleases(ctx context.Context) error
}

// IntegrationOrchestrator defines integration side effects used by lifecycle
// flows after source or dynamic governance state changes.
type IntegrationOrchestrator interface {
	// SyncPluginMenusAndPermissions reconciles manifest menus and route permissions.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes plugin-owned menu rows.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginResourceReferences reconciles governance resource references.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
	// DispatchPluginHookEvent dispatches one plugin hook event.
	DispatchPluginHookEvent(ctx context.Context, event pluginhost.ExtensionPoint, values map[string]interface{}) error
	// RefreshEnabledSnapshot rebuilds the in-memory enabled snapshot.
	RefreshEnabledSnapshot(ctx context.Context) error
	// SetPluginEnabledState updates one plugin entry in the enabled snapshot.
	SetPluginEnabledState(pluginID string, enabled bool)
	// DeletePluginEnabledState removes one plugin entry from the enabled snapshot.
	DeletePluginEnabledState(pluginID string)
	// CanExposeBusinessEntries reports whether one plugin can expose business entries.
	CanExposeBusinessEntries(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports provider-level plugin availability.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
}

// DependencyResolver defines side-effect-free dependency checks used by
// lifecycle before install, uninstall, status, and upgrade side effects.
type DependencyResolver interface {
	// CheckInstall evaluates install dependency blockers for one target.
	CheckInstall(input plugindep.InstallCheckInput) *plugindep.InstallCheckResult
	// CheckReverse evaluates installed downstream blockers for one target.
	CheckReverse(input plugindep.ReverseCheckInput) *plugindep.ReverseCheckResult
}

// I18nTranslator localizes lifecycle veto and operator-facing diagnostics.
type I18nTranslator interface {
	// Translate resolves one runtime message key in the current request locale.
	Translate(ctx context.Context, key string, fallback string) string
}

// CachePublisher publishes lifecycle-affecting plugin runtime changes.
type CachePublisher interface {
	// MarkRuntimeCacheChanged records one cache-affecting plugin governance change.
	MarkRuntimeCacheChanged(ctx context.Context, reason string) error
	// PublishPluginChange records one plugin-scoped governance change.
	PublishPluginChange(ctx context.Context, pluginID string, pluginType string, reason string) error
}

// TopologyProvider abstracts the cluster topology status needed by lifecycle flows.
type TopologyProvider interface {
	// IsClusterModeEnabled reports whether this deployment runs multiple host nodes.
	IsClusterModeEnabled() bool
	// IsPrimaryNode reports whether this host instance is the primary cluster node.
	IsPrimaryNode() bool
}

// TenantProvisioningService provisions tenant-scoped auto-enabled plugins after
// startup policy convergence.
type TenantProvisioningService interface {
	// ProvisionAutoEnabledTenantPlugins applies platform auto-enable policy to existing tenants.
	ProvisionAutoEnabledTenantPlugins(ctx context.Context) error
}

// SourceServicesProvider resolves plugin-scoped source-plugin service
// directories for lifecycle callbacks that need host capabilities.
type SourceServicesProvider interface {
	// SourceServicesForPlugin returns the source-plugin service directory scoped to pluginID.
	SourceServicesForPlugin(pluginID string) pluginhost.Services
}

// Service defines the lifecycle service contract.
type Service interface {
	// BootstrapAutoEnable ensures configured startup plugins are installed and
	// enabled after the root facade has synchronized discovered manifests.
	BootstrapAutoEnable(ctx context.Context, options BootstrapAutoEnableOptions) error
	// ReconcileAutoEnabledTenantPlugins reconciles auto-enabled tenant-scoped
	// plugins into platform new-tenant provisioning policy.
	ReconcileAutoEnabledTenantPlugins(ctx context.Context, entries []AutoEnableEntry) error
	// Install executes the install lifecycle for a discovered source or dynamic
	// plugin and returns the dependency check result produced before side effects.
	Install(ctx context.Context, pluginID string, options InstallOptions) (*plugindep.CheckProjection, error)
	// InstallDynamic executes the low-level install lifecycle for a discovered
	// dynamic plugin. Repeated installs are idempotent unless the same version
	// needs a refresh.
	InstallDynamic(ctx context.Context, pluginID string) error
	// Uninstall executes the full uninstall lifecycle for source or dynamic
	// plugins after the root facade has applied platform governance guards.
	Uninstall(ctx context.Context, pluginID string, options UninstallOptions) error
	// UninstallDynamic executes the low-level uninstall lifecycle for an
	// installed dynamic plugin.
	UninstallDynamic(ctx context.Context, pluginID string) error
	// UpdateStatus executes enable or disable lifecycle orchestration for source
	// and dynamic plugins after the root facade has applied governance guards.
	UpdateStatus(ctx context.Context, pluginID string, status int, options UpdateStatusOptions) error
	// RegisterLifecycleObserver subscribes one synchronous lifecycle observer
	// and returns its unsubscribe function.
	RegisterLifecycleObserver(observer LifecycleObserver) func()
	// EnsureTenantPluginDisableAllowed runs plugin lifecycle preconditions
	// before tenant-scoped plugin disable.
	EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error
	// NotifyTenantPluginDisabled runs best-effort lifecycle notifications after
	// tenant-scoped plugin disable.
	NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int)
	// EnsureTenantDeleteAllowed runs plugin lifecycle preconditions before tenant deletion.
	EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error
	// NotifyTenantDeleted runs best-effort lifecycle notifications after tenant deletion.
	NotifyTenantDeleted(ctx context.Context, tenantID int)
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
	runtimeSvc RuntimeOrchestrator
	// integrationSvc owns menus, resource references, hook dispatch, and enabled snapshots.
	integrationSvc IntegrationOrchestrator
	// migrationSvc executes lifecycle SQL phases and migration ledger writes.
	migrationSvc migration.Service
	// dependencyResolver evaluates lifecycle dependency blockers.
	dependencyResolver DependencyResolver
	// i18nSvc localizes lifecycle veto summaries and diagnostics.
	i18nSvc I18nTranslator
	// cachePublisher publishes runtime cache changes after successful lifecycle writes.
	cachePublisher CachePublisher
	// topology provides cluster topology information.
	topology TopologyProvider
	// tenantProvisioning provisions tenant-scoped auto-enabled plugins after startup convergence.
	tenantProvisioning TenantProvisioningService
	// sourceServices resolves plugin-scoped host capabilities for source-plugin callbacks.
	sourceServices SourceServicesProvider
	// lifecycleObservers stores synchronous lifecycle observers for this lifecycle instance.
	lifecycleObservers *lifecycleObserverRegistry
}

// New creates a new lifecycle Service with explicit runtime dependencies.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	runtimeSvc RuntimeOrchestrator,
	integrationSvc IntegrationOrchestrator,
	migrationSvc migration.Service,
	dependencyResolver DependencyResolver,
	i18nSvc I18nTranslator,
	cachePublisher CachePublisher,
	topology TopologyProvider,
	tenantProvisioning TenantProvisioningService,
	sourceServices SourceServicesProvider,
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
		tenantProvisioning: tenantProvisioning,
		sourceServices:     sourceServices,
		lifecycleObservers: newLifecycleObserverRegistry(),
	}
}
