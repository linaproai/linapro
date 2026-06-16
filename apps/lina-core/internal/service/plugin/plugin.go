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
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/internal/service/session"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"

	"lina-core/internal/model/entity"

	"lina-core/pkg/plugin/capability"
	aitextsvc "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
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

	// DynamicRouteMetadata stores generic metadata for dynamic routes.
	DynamicRouteMetadata = runtime.DynamicRouteMetadata

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

// GetDynamicRouteMetadata returns generic dynamic-route metadata from the request.
// This package-level function is retained for callers that cannot import the runtime sub-package.
var GetDynamicRouteMetadata = runtime.GetDynamicRouteMetadata

// ScanRegisteredSourceManifests returns registered source-plugin manifests
// without synchronizing registry, release, menu, permission, or cache state.
func ScanRegisteredSourceManifests() ([]*SourceManifest, error) {
	return catalog.New(nil).ScanEmbeddedSourceManifests()
}

// AuthHookService defines auth-related plugin hook operations.
type AuthHookService interface {
	// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
	HandleAuthLoginSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
	// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
	HandleAuthLoginFailed(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
	// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
	HandleAuthLogoutSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error
}

// DataCommentService defines host data-table comment lookup operations.
type DataCommentService interface {
	// ResolveDataTableComments resolves host-side table comments for the given
	// data-table names. It degrades to an empty map when metadata lookup is
	// unavailable so plugin list APIs are not blocked by optional schema comments.
	ResolveDataTableComments(ctx context.Context, tables []string) map[string]string
}

// FrontendAssetService defines runtime frontend bundle and asset operations.
type FrontendAssetService interface {
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
}

// SourceIntegrationService defines host integration operations for source plugins.
type SourceIntegrationService interface {
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
		aiTextManager *aitextsvc.Manager,
	) error
	// ListExecutableJobs returns plugin-owned job definitions whose
	// handlers are safe to publish for execution. Dynamic plugins must be in
	// an enabled business-entry state; disabled, pending-upgrade, abnormal, and
	// failed-upgrade dynamic plugins are excluded. Use this only for runtime
	// handler publication, not for authorization previews or task-table
	// projection.
	ListExecutableJobs(ctx context.Context) ([]ManagedJob, error)
	// ListExecutableJobsByPlugin returns executable job definitions for
	// one plugin. It applies the same enablement and runtime-state rules as
	// ListExecutableJobs while narrowing discovery to pluginID, so callers
	// can register handlers during a plugin enable lifecycle without exposing
	// declarations that are not currently executable.
	ListExecutableJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error)
	// ListJobDeclarationsByPlugin returns declared job metadata for one
	// plugin without requiring the plugin business entry to be enabled. This is
	// intended for management review and host-service authorization previews,
	// including not-yet-installed dynamic plugins. Callers must not publish the
	// returned handlers directly because the plugin may not be executable.
	ListJobDeclarationsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error)
	// ListInstalledJobDeclarations returns declared job metadata for
	// installed plugins without requiring their business entries to be enabled.
	// Scheduled-job projection uses this to create or update task-table rows
	// for installed plugins while avoiding preview-only declarations from
	// uninstalled plugins.
	ListInstalledJobDeclarations(ctx context.Context) ([]ManagedJob, error)
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

// ResourceQueryService defines plugin-owned backend resource query operations.
type ResourceQueryService interface {
	// ResolveResourcePermission resolves the plugin-scoped permission required
	// by the generic resource list endpoint for one plugin-owned resource.
	ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error)
	// ListResourceRecords queries plugin-owned backend resource rows.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
}

// LifecycleManagementService defines plugin lifecycle and status management operations.
type LifecycleManagementService interface {
	// BootstrapAutoEnable synchronizes manifests and ensures every plugin listed
	// in plugin.autoEnable is installed and enabled before later host wiring runs.
	BootstrapAutoEnable(ctx context.Context) error
	// ReconcileAutoEnabledTenantPlugins applies startup auto-enable policy to
	// tenant-scoped plugins after source-plugin providers have registered.
	ReconcileAutoEnabledTenantPlugins(ctx context.Context) error
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
	// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled,
	// and optionally persists one host-confirmed host service authorization snapshot
	// before enabling a dynamic plugin.
	UpdateStatus(
		ctx context.Context,
		pluginID string,
		status int,
		authorization *HostServiceAuthorizationInput,
	) error
	// Enable enables the specified plugin.
	Enable(ctx context.Context, pluginID string) error
	// Disable disables the specified plugin.
	Disable(ctx context.Context, pluginID string) error
	// UpdateTenantProvisioningPolicy updates the platform-owned new-tenant plugin provisioning policy.
	UpdateTenantProvisioningPolicy(ctx context.Context, pluginID string, autoEnableForNewTenants bool) error
	// IsInstalled returns whether a plugin is installed.
	IsInstalled(ctx context.Context, pluginID string) bool
	// IsEnabled returns whether a plugin is enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled returns whether pluginID is platform-enabled for framework capability provider use.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// AITextProviderEnv returns typed, plugin-scoped text AI provider construction inputs.
	AITextProviderEnv(pluginID string) aitextsvc.ProviderEnv
	// OrgProviderEnv returns typed, plugin-scoped organization-provider construction inputs.
	OrgProviderEnv(pluginID string) orgspi.ProviderEnv
	// TenantProviderEnv returns typed, plugin-scoped tenant-provider construction inputs.
	TenantProviderEnv(pluginID string) tenantspi.ProviderEnv
	// IsEnabledAuthoritative returns whether pluginID is installed, enabled, and
	// allowed to expose business entries after forcing a persisted governance
	// read instead of reusing process-local platform snapshots. It preserves the
	// current tenant/request scope and returns false when authoritative state
	// cannot be resolved.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
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
	// ListEnabledPluginIDs returns the IDs of plugins that are currently
	// installed and enabled.
	ListEnabledPluginIDs(ctx context.Context) ([]string, error)
}

// SourceUpgradeGovernanceService defines source-plugin upgrade discovery and
// explicit effective-version switching operations.
type SourceUpgradeGovernanceService interface {
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
	// current effective version to the newer discovered source version.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades; list/runtime state exposes the result.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
	// ValidateStartupConsistency fails fast when persisted plugin and tenant
	// governance state is incoherent before routes are served.
	ValidateStartupConsistency(ctx context.Context) error
}

// RegistryQueryService defines manifest synchronization and plugin list query operations.
type RegistryQueryService interface {
	// WithStartupDataSnapshot returns a child context carrying plugin startup
	// snapshots shared by one host startup orchestration.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
	SyncSourcePlugins(ctx context.Context) error
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
	// PrewarmManagementList builds the plugin management summary list read model.
	PrewarmManagementList(ctx context.Context) error
	// PreviewRuntimeUpgrade returns a side-effect-free upgrade preview for one pending plugin.
	PreviewRuntimeUpgrade(ctx context.Context, pluginID string) (*RuntimeUpgradePreview, error)
	// ExecuteRuntimeUpgrade runs one explicit runtime upgrade after confirmation.
	ExecuteRuntimeUpgrade(
		ctx context.Context,
		pluginID string,
		options RuntimeUpgradeOptions,
	) (*RuntimeUpgradeResult, error)
}

// OpenAPIProjectionService defines plugin route projection into the host OpenAPI document.
type OpenAPIProjectionService interface {
	// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
}

// RuntimeManagementService defines dynamic plugin runtime reconciliation and state query operations.
type RuntimeManagementService interface {
	// StartRuntimeReconciler starts the background reconciler loop for dynamic plugins.
	StartRuntimeReconciler(ctx context.Context)
	// ReconcileRuntimePlugins runs one reconciliation pass for all dynamic plugins.
	ReconcileRuntimePlugins(ctx context.Context) error
	// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
	ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error)
}

// DynamicPackageService defines runtime WASM package upload operations.
type DynamicPackageService interface {
	// UploadDynamicPackage validates and stores a runtime WASM package.
	UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error)
}

// DynamicRouteService defines host-managed dynamic route middleware and dispatch registration operations.
type DynamicRouteService interface {
	// PrepareDynamicRouteMiddleware prepares dynamic route state before the main handler.
	PrepareDynamicRouteMiddleware(r *ghttp.Request)
	// AuthenticateDynamicRouteMiddleware authenticates JWT tokens for dynamic routes.
	AuthenticateDynamicRouteMiddleware(r *ghttp.Request)
	// RegisterDynamicRouteDispatcher binds the dynamic route catch-all handler to the group.
	RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup)
}

// Service defines the plugin service contract by composing plugin sub-capabilities.
type Service interface {
	AuthHookService
	DataCommentService
	FrontendAssetService
	SourceIntegrationService
	ResourceQueryService
	LifecycleManagementService
	SourceUpgradeGovernanceService
	RegistryQueryService
	OpenAPIProjectionService
	RuntimeManagementService
	DynamicPackageService
	DynamicRouteService
	LifecycleObserverRegistrar
}

// Ensure serviceImpl satisfies the composed plugin facade contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// configSvc reads host startup configuration such as plugin.autoEnable.
	configSvc configsvc.Service
	// topology reports whether the current host instance should execute shared
	// lifecycle actions or wait for another primary node to converge them.
	topology Topology
	// catalogSvc provides manifest discovery, validation, and manifest asset access.
	catalogSvc catalog.Service
	// storeSvc owns plugin governance persistence and stable projections.
	storeSvc store.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// migrationSvc executes plugin SQL lifecycle phases and migration ledger writes.
	migrationSvc migration.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// upgradeSvc provides unified source and dynamic upgrade planning and execution.
	upgradeSvc upgrade.Service
	// frontendSvc manages in-memory frontend bundles for dynamic plugins.
	frontendSvc frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
	// capabilities exposes runtime-owned adapters for lazy provider construction.
	capabilities capability.Services
	// sourceServices resolves plugin-scoped source-plugin services for integration callbacks.
	sourceServices *sourceServicesProvider
	// orgDeptProvider resolves organization departments for plugin resource data-scope filters.
	orgDeptProvider *organizationDeptProvider
	// i18nSvc localizes plugin lifecycle messages and invalidates runtime
	// translation bundles after plugin lifecycle mutations.
	i18nSvc pluginI18nService
	// runtimeCacheRevisionCtrl coordinates process-local runtime caches in cluster deployments.
	runtimeCacheRevisionCtrl *revisionctrl.Controller
	// activeDynamicArtifactSnapshotMu protects activeDynamicArtifactSnapshot.
	activeDynamicArtifactSnapshotMu sync.Mutex
	// activeDynamicArtifactSnapshot stores the last reconciled active dynamic artifact path by plugin ID.
	activeDynamicArtifactSnapshot map[string]string
	// wasmRuntime owns dynamic-plugin WASM execution and host-call dependencies.
	wasmRuntime *wasmRuntimeProvider
	// lifecycleObservers stores transitional root-level observers for flows not yet migrated to lifecycle.
	lifecycleObservers *lifecycleObserverRegistry
	// runtimeUpgradeLockStore coordinates explicit runtime upgrades across cluster nodes.
	runtimeUpgradeLockStore coordination.LockStore
	// managementListCache stores the plugin-management summary read model.
	managementListCache *management.ListCache
	// tenantStartup validates tenant-governance startup state through a narrow tenant capability.
	tenantStartup pluginTenantStartupCapability
	// tenantProvisioning provisions tenant-scoped auto-enabled plugins after startup policy convergence.
	tenantProvisioning tenantspi.PluginProvisioningService
	// tenantGovernance guards platform plugin-governance writes in HTTP paths.
	tenantGovernance platformGovernanceTenantCapability
}

// New creates and returns a new plugin Service.
// Pass a non-nil topology for cluster-aware deployments; pass nil to use the
// default single-node topology implementation.
// reconcilerLockSvc must be created by the startup composition root and shared
// with host-lock services that use the same deployment-selected locker backend.
// capabilityServices, tenant/org governance services, and WASM host-service
// factories must also be startup-owned shared instances; callers that need to
// break the host-service construction cycle should use RuntimeDelegate and bind
// it to the returned service before serving requests.
func New(
	topology Topology,
	configProvider configsvc.Service,
	bizCtxProvider bizctx.Service,
	cacheCoordSvc cachecoord.Service,
	i18nSvc i18nsvc.Service,
	sessionStore session.Store,
	roleAccess runtime.RoleAccessProjector,
	reconcilerLockSvc locker.Service,
	runtimeUpgradeLockStore coordination.LockStore,
	capabilityServices capability.Services,
	organizationSvc orgcapsvc.Service,
	tenantStartup pluginTenantStartupCapability,
	tenantProvisioning tenantspi.PluginProvisioningService,
	tenantGovernance platformGovernanceTenantCapability,
	pluginConfigFactory plugincap.ConfigServiceFactory,
	hostConfigSvc hostconfigcap.Service,
	pluginManifestFactory manifestcap.ServiceFactory,
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
		return nil, gerror.New("plugin service requires a non-nil role access projector")
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
	if tenantStartup == nil {
		return nil, gerror.New("plugin service requires a non-nil tenant startup capability")
	}
	if tenantProvisioning == nil {
		return nil, gerror.New("plugin service requires a non-nil tenant provisioning capability")
	}
	if tenantGovernance == nil {
		return nil, gerror.New("plugin service requires a non-nil tenant platform governance capability")
	}
	wasmRuntimeInstance, err := newWasmHostServiceRuntime(
		capabilityServices,
		pluginConfigFactory,
		hostConfigSvc,
		pluginManifestFactory,
	)
	if err != nil {
		return nil, err
	}

	var topo Topology = singleNodeTopology{}
	if topology != nil {
		topo = topology
	}

	var (
		catalogSvc           = catalog.New(configProvider)
		topologyProvider     = &runtimeTopologyAdapter{topo}
		storeSvc             = store.New(catalogSvc, topologyProvider)
		migrationSvc         = migration.New(catalogSvc, storeSvc)
		frontendSvc          = frontend.New(catalogSvc, storeSvc)
		openapiRevision      = openapi.NewDeferredRevisionReader()
		openapiSvc           = openapi.New(catalogSvc, storeSvc, openapiRevision, i18nSvc)
		sourceServices       = &sourceServicesProvider{capabilities: capabilityServices}
		orgDeptProvider      = &organizationDeptProvider{service: organizationSvc}
		integrationDelegates = &integrationDelegateProvider{}
		cacheChangeNotifier  = &runtimeCacheChangeNotifierProvider{}
		dependencyValidator  = &dependencyValidatorProvider{}
		wasmRuntime          = &wasmRuntimeProvider{runtime: wasmRuntimeInstance}
		lifecycleTopology    = &lifecycleTopologyAdapter{topo}
	)
	runtimeSvc := runtime.New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		frontendSvc,
		openapiSvc,
		i18nSvc,
		reconcilerLockSvc,
		topologyProvider,
		integrationDelegates,
		integrationDelegates,
		integrationDelegates,
		&jwtConfigAdapter{configProvider},
		&uploadSizeAdapter{configProvider},
		&userCtxAdapter{bizCtxProvider},
		sessionStore,
		roleAccess,
		integrationDelegates,
		cacheChangeNotifier,
		dependencyValidator,
		sourceServices,
		wasmRuntime,
	)
	integrationSvc := integration.New(
		catalogSvc,
		storeSvc,
		&bizCtxAdapter{bizCtxProvider},
		&integrationTopologyAdapter{topo},
		sourceServices,
		orgDeptProvider,
		runtimeSvc,
		integration.NewSharedState(),
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
		lifecycleTopology,
		tenantProvisioning,
		sourceServices,
	)

	service := &serviceImpl{
		configSvc:                     configProvider,
		topology:                      topo,
		catalogSvc:                    catalogSvc,
		storeSvc:                      storeSvc,
		lifecycleSvc:                  lifecycleSvc,
		migrationSvc:                  migrationSvc,
		runtimeSvc:                    runtimeSvc,
		integrationSvc:                integrationSvc,
		frontendSvc:                   frontendSvc,
		openapiSvc:                    openapiSvc,
		capabilities:                  capabilityServices,
		sourceServices:                sourceServices,
		orgDeptProvider:               orgDeptProvider,
		i18nSvc:                       i18nSvc,
		activeDynamicArtifactSnapshot: make(map[string]string),
		wasmRuntime:                   wasmRuntime,
		runtimeUpgradeLockStore:       runtimeUpgradeLockStore,
		managementListCache:           management.NewListCache(),
		tenantStartup:                 tenantStartup,
		tenantProvisioning:            tenantProvisioning,
		tenantGovernance:              tenantGovernance,
	}
	cacheChangeNotifier.BindService(service)
	dependencyValidator.BindService(service)
	service.runtimeCacheRevisionCtrl = newRuntimeCacheRevisionController(
		topo,
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
		lifecycleSvc,
		runtimeSvc,
		integrationSvc,
		migrationSvc,
		dependencyResolver,
		i18nSvc,
		runtimeUpgradeLockStore,
		upgradeCachePublisher{service: service},
		upgradeCacheFreshener{service: service},
		topo,
		configProvider,
	)
	if err != nil {
		return nil, err
	}
	service.upgradeSvc = upgradeSvc
	return service, nil
}
