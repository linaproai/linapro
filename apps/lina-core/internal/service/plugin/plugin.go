// Package plugin implements plugin manifest discovery, lifecycle orchestration,
// governance metadata synchronization, and host integration for Lina plugins.
package plugin

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cachecoord/revisionctrl"
	"lina-core/internal/service/cluster"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/management"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/upgrade"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"

	"lina-core/internal/model/entity"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/pluginhost"
)

type (
	// SourceManifest aliases the framework plugin manifest model discovered
	// from registered source plugins.
	SourceManifest = catalog.Manifest

	// DynamicUploadInput defines input for uploading a runtime WASM package.
	DynamicUploadInput = runtime.DynamicUploadInput

	// DynamicUploadOutput defines output for uploading a runtime WASM package.
	DynamicUploadOutput = runtime.DynamicUploadOutput

	// RuntimeStateListOutput defines output for public runtime state queries.
	RuntimeStateListOutput = runtime.RuntimeStateListOutput

	// RuntimeUpgradeFailure defines latest runtime-upgrade failure details.
	RuntimeUpgradeFailure = runtime.RuntimeUpgradeFailure

	// RuntimeUpgradeManifestSnapshot aliases the review-friendly manifest snapshot model.
	RuntimeUpgradeManifestSnapshot = upgrade.RuntimeUpgradeManifestSnapshot

	// RuntimeUpgradeSQLSummary summarizes manifest SQL assets visible to preview.
	RuntimeUpgradeSQLSummary = upgrade.RuntimeUpgradeSQLSummary

	// RuntimeUpgradeHostServicesDiff summarizes service-level hostServices drift.
	RuntimeUpgradeHostServicesDiff = upgrade.RuntimeUpgradeHostServicesDiff

	// RuntimeUpgradeHostServiceChange summarizes one service-level hostServices change.
	RuntimeUpgradeHostServiceChange = upgrade.RuntimeUpgradeHostServiceChange

	// RuntimeUpgradeState aliases the plugin runtime-upgrade state enum.
	RuntimeUpgradeState = upgrade.RuntimeUpgradeState

	// RuntimeUpgradeAbnormalReason aliases the plugin runtime-upgrade abnormal reason enum.
	RuntimeUpgradeAbnormalReason = upgrade.RuntimeUpgradeAbnormalReason

	// RuntimeUpgradePreview describes the side-effect-free plan shown before a
	// runtime plugin upgrade is confirmed.
	RuntimeUpgradePreview = upgrade.RuntimeUpgradePreview

	// RuntimeUpgradeOptions captures explicit management confirmations for a
	// runtime plugin upgrade request.
	RuntimeUpgradeOptions = upgrade.RuntimeUpgradeOptions

	// RuntimeUpgradeResult describes one completed explicit runtime upgrade action.
	RuntimeUpgradeResult = upgrade.RuntimeUpgradeResult

	// SourceUpgradeStatus aliases the unified upgrade source-plugin upgrade status contract.
	SourceUpgradeStatus = upgrade.SourceUpgradeStatus

	// SourceUpgradeResult aliases the unified upgrade explicit source-plugin upgrade result contract.
	SourceUpgradeResult = upgrade.SourceUpgradeResult

	// ResourceListInput defines input for querying a plugin-owned backend resource.
	ResourceListInput = integration.ResourceListInput

	// ResourceListOutput defines output for querying a plugin-owned backend resource.
	ResourceListOutput = integration.ResourceListOutput

	// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
	RuntimeFrontendAssetOutput = frontend.RuntimeFrontendAssetOutput

	// Metadata stores generic metadata for dynamic routes.
	Metadata = runtime.Metadata

	// PluginDynamicStateItem represents public runtime state of one plugin.
	PluginDynamicStateItem = runtime.PluginDynamicStateItem

	// HostServiceAuthorizationInput defines one install/enable authorization confirmation payload.
	HostServiceAuthorizationInput = store.HostServiceAuthorizationInput

	// InstallOptions captures the per-request install decoration that callers can opt into.
	// All fields default to the zero value, which preserves the original install behavior
	// (no mock data and no host-service authorization snapshot).
	InstallOptions struct {
		// Authorization optionally carries a host-service authorization snapshot for
		// dynamic plugins that require explicit confirmation before install.
		Authorization *HostServiceAuthorizationInput
		// InstallMode optionally carries the platform operator's explicit tenant
		// governance selection. Empty means use the plugin manifest default.
		InstallMode string
		// InstallMockData enables the optional mock-data load phase. When true the host
		// scans manifest/sql/mock-data/ and executes those SQL files inside a single
		// database transaction; any failure rolls back only the mock load and leaves
		// the install SQL phase results intact.
		InstallMockData bool
	}

	// HostServiceAuthorizationDecision narrows one authorized service snapshot.
	HostServiceAuthorizationDecision = store.HostServiceAuthorizationDecision

	// ManagedJob describes one plugin-owned scheduled-job definition that
	// the host can project into the unified scheduled-job management table.
	ManagedJob = integration.ManagedJob

	// DependencyFrameworkCheck exposes framework compatibility for management clients.
	DependencyFrameworkCheck = plugindep.FrameworkProjection

	// DependencyPluginCheck exposes one plugin dependency edge.
	DependencyPluginCheck = plugindep.PluginProjection

	// DependencyBlocker exposes one hard dependency failure.
	DependencyBlocker = plugindep.BlockerProjection

	// DependencyReverseDependent exposes one installed downstream hard dependency.
	DependencyReverseDependent = plugindep.ReverseDependentProjection

	// DependencyOwnerHostServiceSummary exposes owner-aware host service summaries for reverse diagnostics.
	DependencyOwnerHostServiceSummary = plugindep.OwnerHostServiceProjection

	// DependencyCheckResult is the management-facing dependency status snapshot.
	DependencyCheckResult = plugindep.CheckProjection

	// PluginItem is the display-ready projection of one plugin entry.
	PluginItem = management.PluginItem

	// ListOutput defines output for plugin list query.
	ListOutput = management.ListOutput

	// ListInput defines input for plugin list query.
	ListInput = management.ListInput
)

const (
	// RuntimeUpgradeStateNormal means the effective and discovered metadata are aligned.
	RuntimeUpgradeStateNormal = upgrade.RuntimeUpgradeStateNormal
	// RuntimeUpgradeStatePendingUpgrade means discovered plugin files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade = upgrade.RuntimeUpgradeStatePendingUpgrade
	// RuntimeUpgradeStateAbnormal means discovered plugin files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal = upgrade.RuntimeUpgradeStateAbnormal
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is currently reconciling.
	RuntimeUpgradeStateUpgradeRunning = upgrade.RuntimeUpgradeStateUpgradeRunning
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed = upgrade.RuntimeUpgradeStateUpgradeFailed
	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than the DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective = upgrade.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is not semver-compatible.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed = upgrade.RuntimeUpgradeAbnormalReasonVersionCompareFailed
)

// UninstallOptions defines one plugin uninstall policy snapshot.
type UninstallOptions struct {
	// PurgeStorageData reports whether uninstall should also clear plugin-owned
	// table data and stored files.
	PurgeStorageData bool
	// Force reports whether an authorized caller requested precondition veto bypass.
	Force bool
}

// UpdateStatusOptions defines one plugin status transition request.
type UpdateStatusOptions struct {
	// Status is the target enabled-state value, where 1 means enabled and 0 means disabled.
	Status int
	// Authorization optionally carries one host-confirmed host-service authorization
	// snapshot before enabling a dynamic plugin.
	Authorization *HostServiceAuthorizationInput
}

// ManagedJobQuery defines one plugin-owned scheduled-job discovery request.
type ManagedJobQuery struct {
	// PluginID optionally narrows discovery to one plugin. Empty means all matching plugins.
	PluginID string
	// ExecutableOnly requires installed, enabled, runtime-safe job handlers that
	// may be published for execution.
	ExecutableOnly bool
	// InstalledOnly skips preview declarations from uninstalled plugins while
	// still returning disabled-plugin declarations for management projection.
	InstalledOnly bool
	// IncludeHandlers keeps executable handler functions in the returned jobs.
	// Management and projection callers should leave this false.
	IncludeHandlers bool
}

// Service defines the composed plugin service contract.
type Service interface {
	managementService
	startupService
	runtimeHTTPService
	integrationService
	jobService
	stateService
	capabilityEnvService
	tenantLifecycleService
}

// managementService defines plugin management and management-read contracts.
type managementService interface {
	// ResolveDataTableComments resolves host-side table comments for the given
	// data-table names. It degrades to an empty map when metadata lookup is
	// unavailable so plugin list APIs are not blocked by optional schema comments.
	ResolveDataTableComments(ctx context.Context, tables []string) map[string]string
	// Install executes the install lifecycle and returns the dependency plan/results
	// produced before the target plugin side effects. It optionally persists one
	// host-confirmed host service authorization snapshot when the target is a
	// dynamic plugin. When options.InstallMockData is true the optional mock-data
	// load phase runs inside one database transaction after install SQL completes;
	// any failure rolls back only the mock load and leaves the install results intact.
	Install(
		ctx context.Context,
		pluginID string,
		options InstallOptions,
	) (*DependencyCheckResult, error)
	// Uninstall executes the uninstall lifecycle with one explicit policy snapshot.
	Uninstall(ctx context.Context, pluginID string, options UninstallOptions) error
	// CheckPluginDependencies evaluates dependency status for plugin management UI.
	CheckPluginDependencies(ctx context.Context, pluginID string) (*DependencyCheckResult, error)
	// UpdateStatus updates plugin status and optionally persists one dynamic-plugin
	// host-service authorization snapshot before enabling.
	UpdateStatus(ctx context.Context, pluginID string, options UpdateStatusOptions) error
	// UpdateTenantProvisioningPolicy updates the platform-owned new-tenant plugin provisioning policy.
	UpdateTenantProvisioningPolicy(ctx context.Context, pluginID string, autoEnableForNewTenants bool) error
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
	// current effective version to the newer discovered source version.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// SyncSourcePluginsStrict synchronizes source plugins discovered by the
	// running host.
	SyncSourcePluginsStrict(ctx context.Context) (*ListOutput, error)
	// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
	// returns the combined list of source and dynamic plugin items.
	SyncAndList(ctx context.Context) (*ListOutput, error)
	// List returns the paginated read-only plugin summary list with optional filtering applied.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Get returns one read-only plugin detail projection by exact plugin ID.
	Get(ctx context.Context, pluginID string) (*PluginItem, error)
	// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one pending plugin.
	PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error)
	// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
	ExecuteRuntimeUpgrade(
		ctx context.Context,
		pluginID string,
		options RuntimeUpgradeOptions,
	) (*RuntimeUpgradeResult, error)
	// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
	ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error)
	// UploadDynamicPackage validates and stores a runtime WASM package.
	UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error)
	// ResolveResourcePermission resolves the plugin-scoped permission required
	// by the generic resource list endpoint for one plugin-owned resource.
	ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error)
	// ListResourceRecords queries plugin-owned backend resource rows.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
}

// runtimeHTTPService defines dynamic route and frontend asset contracts used by HTTP startup.
type runtimeHTTPService interface {
	// PrewarmRuntimeFrontendBundles preloads frontend bundles for enabled dynamic plugins.
	PrewarmRuntimeFrontendBundles(ctx context.Context) error
	// ResolveRuntimeFrontendAsset resolves one frontend asset for a dynamic plugin.
	ResolveRuntimeFrontendAsset(
		ctx context.Context,
		pluginID string,
		version string,
		relativePath string,
	) (*RuntimeFrontendAssetOutput, error)
	// BuildRuntimeFrontendPublicBaseURL returns the public base URL for a plugin's hosted frontend assets.
	BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string
	// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
	// StartRuntimeReconciler starts the background reconciler loop for dynamic plugins.
	StartRuntimeReconciler(ctx context.Context)
	// ReconcileRuntimePlugins runs one reconciliation pass for all dynamic plugins.
	ReconcileRuntimePlugins(ctx context.Context) error
	// PrepareDynamicRouteMiddleware prepares dynamic route state before the main handler.
	PrepareDynamicRouteMiddleware(r *ghttp.Request)
	// AuthenticateDynamicRouteMiddleware authenticates JWT tokens for dynamic routes.
	AuthenticateDynamicRouteMiddleware(r *ghttp.Request)
	// RegisterDynamicRouteDispatcher binds the dynamic route catch-all handler to the group.
	RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup)
}

// integrationService defines source-plugin route, hook, menu, and provider registration contracts.
type integrationService interface {
	// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
	RegisterHTTPRoutes(
		ctx context.Context,
		server *ghttp.Server,
		pluginGroup *ghttp.RouterGroup,
		middlewares pluginhost.RouteMiddlewares,
	) error
	// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
	ListSourceRouteBindings() []pluginhost.SourceRouteBinding
	// RegisterJobs registers callback-contributed scheduled jobs for source plugins.
	RegisterJobs(ctx context.Context) error
	// RegisterSourcePluginProviderFactories registers source-plugin provider declarations into shared managers.
	RegisterSourcePluginProviderFactories(
		tenantManager *tenantspi.Manager,
		orgManager *orgspi.Manager,
	) error
	// DispatchHookEvent dispatches one named hook event to all enabled plugins.
	DispatchHookEvent(
		ctx context.Context,
		event pluginhost.ExtensionPoint,
		values map[string]interface{},
	) error
	// FilterMenus filters disabled plugin menus from the given menu list.
	FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
	// FilterPermissionMenus filters permission menus based on plugin enablement.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

// startupService defines plugin startup and startup-read-model contracts.
type startupService interface {
	// BootstrapBuiltinPlugins synchronizes manifests and ensures project built-in
	// source plugins are installed, upgraded when needed, and enabled before
	// ordinary plugin.autoEnable startup reconciliation runs. The method is a
	// startup-only governance path and bypasses ordinary management write guards.
	BootstrapBuiltinPlugins(ctx context.Context) error
	// BootstrapAutoEnable synchronizes manifests and ensures every plugin listed
	// in plugin.autoEnable is installed and enabled before later host wiring runs.
	BootstrapAutoEnable(ctx context.Context) error
	// ReconcileAutoEnabledTenantPlugins applies startup auto-enable policy to
	// tenant-scoped plugins after source-plugin providers have registered.
	ReconcileAutoEnabledTenantPlugins(ctx context.Context) error
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades; list/runtime state exposes the result.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
	// ValidateStartupConsistency fails fast when persisted plugin and tenant
	// governance state is incoherent before routes are served.
	ValidateStartupConsistency(ctx context.Context) error
	// WithStartupDataSnapshot returns a child context carrying plugin startup
	// snapshots shared by one host startup orchestration.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// PrewarmManagementList builds the plugin management summary list read model.
	PrewarmManagementList(ctx context.Context) error
}

// stateService defines plugin state-read contracts used by guards and capability providers.
type stateService interface {
	// IsInstalled returns whether a plugin is installed.
	IsInstalled(ctx context.Context, pluginID string) bool
	// IsEnabled returns whether a plugin is enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled returns whether pluginID is platform-enabled for framework capability provider use.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsEnabledAuthoritative returns whether pluginID is installed, enabled, and
	// allowed to expose business entries after forcing a persisted governance
	// read instead of reusing process-local platform snapshots. It preserves the
	// current tenant/request scope and returns false when authoritative state
	// cannot be resolved.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}

// capabilityEnvService defines provider construction inputs scoped to one plugin.
type capabilityEnvService interface {
	// OrgProviderEnv returns typed, plugin-scoped organization-provider construction inputs.
	OrgProviderEnv(ctx context.Context, pluginID string) orgspi.ProviderEnv
	// TenantProviderEnv returns typed, plugin-scoped tenant-provider construction inputs.
	TenantProviderEnv(ctx context.Context, pluginID string) tenantspi.ProviderEnv
}

// tenantLifecycleService defines tenant-governance lifecycle preconditions and notifications.
type tenantLifecycleService interface {
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

// jobService defines plugin-owned scheduled-job and lifecycle observer contracts.
type jobService interface {
	// ListManagedJobs returns plugin-owned scheduled-job declarations or
	// executable handlers according to query. Callers requesting executable jobs
	// must set ExecutableOnly; management projection callers should leave
	// IncludeHandlers false so handler functions are not published accidentally.
	ListManagedJobs(ctx context.Context, query ManagedJobQuery) ([]ManagedJob, error)
	// ListEnabledPluginIDs returns the IDs of plugins that are currently
	// installed and enabled.
	ListEnabledPluginIDs(ctx context.Context) ([]string, error)
	// RegisterLifecycleObserver subscribes one synchronous lifecycle observer and
	// returns its unsubscribe function.
	RegisterLifecycleObserver(observer LifecycleObserver) func()
}

// Ensure serviceImpl satisfies the composed plugin facade contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// configSvc reads host startup configuration such as plugin.autoEnable.
	configSvc configsvc.Service
	// topology reports whether the current host instance should execute shared
	// lifecycle actions or wait for another primary node to converge them.
	topology cluster.Service
	// catalogSvc provides manifest discovery, validation, and manifest asset access.
	catalogSvc catalog.Service
	// storeSvc owns plugin governance persistence and stable projections.
	storeSvc store.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// pluginLifecycleService exposes host-internal lifecycle checks to provider
	// construction without publishing them through plugincap.Service.
	pluginLifecycleService plugincap.LifecycleService
	// migrationSvc executes plugin SQL lifecycle phases and migration ledger writes.
	migrationSvc migration.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// frontendSvc manages in-memory frontend bundles for dynamic plugins.
	frontendSvc frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
	// capabilities exposes runtime-owned adapters for lazy provider construction.
	capabilities capability.Services
	// i18nSvc localizes plugin lifecycle messages and invalidates runtime
	// translation bundles after plugin lifecycle mutations.
	i18nSvc i18nsvc.Service
	// runtimeCacheRevisionCtrl coordinates process-local runtime caches in cluster deployments.
	runtimeCacheRevisionCtrl *revisionctrl.Controller
	// activeDynamicArtifactSnapshotMu protects activeDynamicArtifactSnapshot.
	activeDynamicArtifactSnapshotMu sync.Mutex
	// activeDynamicArtifactSnapshot stores the last reconciled active dynamic artifact path by plugin ID.
	activeDynamicArtifactSnapshot map[string]string
	// wasmRuntime owns dynamic-plugin WASM execution and host-call dependencies.
	wasmRuntime *wasmRuntimeProvider
	// runtimeUpgradeLockStore coordinates explicit runtime upgrades across cluster nodes.
	runtimeUpgradeLockStore coordination.LockStore
	// managementListCache stores the plugin-management summary read model.
	managementListCache *management.ListCache
	// tenantSvc provides tenant governance, startup consistency, and tenant plugin provisioning.
	tenantSvc tenantspi.Service
}

// New creates and returns a new plugin Service.
// Pass a non-nil topology for cluster-aware deployments; pass nil to use the
// default single-node topology implementation.
// reconcilerLockSvc must be created by the startup composition root and shared
// with host-lock services that use the same deployment-selected locker backend.
// capabilityServices, tenant/org governance services, plugin config factory,
// and host config service must also be startup-owned shared instances. The
// plugin service owns the manifest resource factory used by the dynamic WASM
// host runtime. Callers that need to break the host-service construction cycle
// should use RuntimeDelegate and bind it to the returned service before serving
// requests.
func New(
	topology cluster.Service,
	configProvider configsvc.Service,
	bizCtxProvider bizctx.Service,
	cacheCoordSvc cachecoord.Service,
	i18nSvc i18nsvc.Service,
	sessionStore session.Store,
	roleAccess rolesvc.Service,
	reconcilerLockSvc locker.Service,
	runtimeUpgradeLockStore coordination.LockStore,
	capabilityServices capability.Services,
	organizationSvc orgcapsvc.Service,
	tenantSvc tenantspi.Service,
	pluginConfigFactory PluginConfigFactory,
	hostConfigSvc hostconfigcap.Service,
) (Service, error) {
	if configProvider == nil {
		return nil, gerror.New("plugin service requires a non-nil config service")
	}
	if bizCtxProvider == nil {
		return nil, gerror.New("plugin service requires a non-nil bizctx service")
	}
	if cacheCoordSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil cachecoord service")
	}
	if i18nSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil i18n service")
	}
	if sessionStore == nil {
		return nil, gerror.New("plugin service requires a non-nil session store")
	}
	if roleAccess == nil {
		return nil, gerror.New("plugin service requires a non-nil role service")
	}
	if reconcilerLockSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil reconciler lock service")
	}
	if capabilityServices == nil {
		return nil, gerror.New("plugin service requires non-nil capability services")
	}
	if organizationSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil organization capability service")
	}
	if tenantSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil tenant service")
	}
	ownerCapabilities, err := buildSourceCapabilityRegistry()
	if err != nil {
		return nil, gerror.Wrap(err, "build source capability registry failed")
	}
	wasmRuntimeInstance, err := newWasmHostServiceRuntime(
		capabilityServices,
		ownerCapabilities,
		pluginConfigFactory,
		hostConfigSvc,
		manifestresource.NewFactory(""),
	)
	if err != nil {
		return nil, err
	}

	topologySvc := topology
	if topologySvc == nil {
		topologySvc = cluster.New(nil)
	}

	var (
		catalogSvc           = catalog.New(configProvider)
		storeSvc             = store.New(catalogSvc, topologySvc)
		migrationSvc         = migration.New(catalogSvc, storeSvc)
		frontendSvc          = frontend.New(catalogSvc, storeSvc)
		openapiRevision      = openapi.NewDeferredRevisionReader()
		openapiSvc           = openapi.New(catalogSvc, storeSvc, openapiRevision, i18nSvc)
		integrationDelegates = &integrationDelegateProvider{}
		cacheChangeNotifier  = &runtimeCacheChangeNotifierProvider{}
		dependencyValidator  = &dependencyValidatorProvider{}
		wasmRuntime          = &wasmRuntimeProvider{runtime: wasmRuntimeInstance}
	)
	runtimeSvc := runtime.New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		frontendSvc,
		i18nSvc,
		reconcilerLockSvc,
		topologySvc,
		integrationDelegates,
		configProvider,
		bizCtxProvider,
		sessionStore,
		roleAccess,
		cacheChangeNotifier,
		dependencyValidator,
		capabilityServices,
		wasmRuntime,
	)
	integrationSvc := integration.New(
		catalogSvc,
		storeSvc,
		bizCtxProvider,
		topologySvc,
		capabilityServices,
		organizationSvc,
		runtimeSvc,
	)
	integrationDelegates.BindService(integrationSvc)
	dependencyResolver := plugindep.New()
	lifecycleSvc := lifecycle.New(
		catalogSvc,
		storeSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencyResolver,
		i18nSvc,
		cacheChangeNotifier,
		topologySvc,
		tenantSvc,
		capabilityServices,
	)

	service := &serviceImpl{
		configSvc:                     configProvider,
		topology:                      topologySvc,
		catalogSvc:                    catalogSvc,
		storeSvc:                      storeSvc,
		lifecycleSvc:                  lifecycleSvc,
		pluginLifecycleService:        plugincap.NewLifecycle(lifecycleSvc),
		migrationSvc:                  migrationSvc,
		runtimeSvc:                    runtimeSvc,
		integrationSvc:                integrationSvc,
		frontendSvc:                   frontendSvc,
		openapiSvc:                    openapiSvc,
		capabilities:                  capabilityServices,
		i18nSvc:                       i18nSvc,
		activeDynamicArtifactSnapshot: make(map[string]string),
		wasmRuntime:                   wasmRuntime,
		runtimeUpgradeLockStore:       runtimeUpgradeLockStore,
		managementListCache:           management.NewListCache(),
		tenantSvc:                     tenantSvc,
	}
	cacheChangeNotifier.BindService(service)
	dependencyValidator.BindService(service)
	service.runtimeCacheRevisionCtrl = newRuntimeCacheRevisionController(
		topologySvc,
		cacheCoordSvc,
		integrationSvc,
		frontendSvc,
		i18nSvc,
		service,
		openapiSvc,
		wasmRuntime,
		catalogSvc,
		storeSvc,
		service,
	)
	openapiRevision.Bind(service)
	upgradeSvc, err := upgrade.New(
		catalogSvc,
		storeSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencyResolver,
		i18nSvc,
		runtimeUpgradeLockStore,
		upgradeCachePublisher{service: service},
		upgradeCacheFreshener{service: service},
		topologySvc,
		configProvider,
	)
	if err != nil {
		return nil, err
	}
	if err = lifecycleSvc.BindUpgrade(upgradeSvc); err != nil {
		return nil, err
	}
	return service, nil
}
