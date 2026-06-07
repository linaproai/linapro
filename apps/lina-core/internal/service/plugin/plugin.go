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
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/runtimeupgrade"
	sourceupgradeinternal "lina-core/internal/service/plugin/internal/sourceupgrade"
	"lina-core/internal/service/plugin/runtimecache"
	"lina-core/internal/service/session"
	orgcapsvc "lina-core/pkg/plugin/capability/orgcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"

	"lina-core/internal/model/entity"

	"lina-core/pkg/plugin/capability"
	aitextsvc "lina-core/pkg/plugin/capability/aicap/aitext"
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
	RuntimeUpgradeManifestSnapshot = catalog.ManifestSnapshot

	// RuntimeUpgradeSQLSummary summarizes manifest SQL assets visible to preview.
	RuntimeUpgradeSQLSummary = runtimeupgrade.SQLSummary

	// RuntimeUpgradeHostServicesDiff summarizes service-level hostServices drift.
	RuntimeUpgradeHostServicesDiff = runtimeupgrade.HostServicesDiff

	// RuntimeUpgradeHostServiceChange summarizes one service-level hostServices change.
	RuntimeUpgradeHostServiceChange = runtimeupgrade.HostServiceChange

	// RuntimeUpgradeState aliases the plugin runtime-upgrade state enum.
	RuntimeUpgradeState = runtime.RuntimeUpgradeState

	// RuntimeUpgradeAbnormalReason aliases the plugin runtime-upgrade abnormal reason enum.
	RuntimeUpgradeAbnormalReason = runtime.RuntimeUpgradeAbnormalReason

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
	HostServiceAuthorizationInput = catalog.HostServiceAuthorizationInput

	// InstallOptions captures the per-request install decoration that callers can opt into.
	// All fields default to the zero value, which preserves the original install behavior
	// (no mock data, no host-service authorization snapshot, no startup context).
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
		// startupAutoEnable marks install requests initiated by plugin.autoEnable
		// startup bootstrap for the explicitly configured target plugin.
		startupAutoEnable bool
		// dependencyResult records the server-side dependency check produced
		// during this install request.
		dependencyResult *DependencyCheckResult
	}

	// RuntimeUpgradeOptions captures explicit management confirmations for a
	// runtime plugin upgrade request.
	RuntimeUpgradeOptions struct {
		// Confirmed must be true before the host performs upgrade side effects.
		Confirmed bool
		// Authorization optionally carries the hostServices authorization snapshot
		// confirmed for the target dynamic release before it becomes effective.
		Authorization *HostServiceAuthorizationInput
	}

	// HostServiceAuthorizationDecision narrows one authorized service snapshot.
	HostServiceAuthorizationDecision = catalog.HostServiceAuthorizationDecision

	// ManagedCronJob describes one plugin-owned scheduled-job definition that
	// the host can project into the unified scheduled-job management table.
	ManagedCronJob = integration.ManagedCronJob

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
	RuntimeUpgradeStateNormal = runtime.RuntimeUpgradeStateNormal
	// RuntimeUpgradeStatePendingUpgrade means discovered plugin files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade = runtime.RuntimeUpgradeStatePendingUpgrade
	// RuntimeUpgradeStateAbnormal means discovered plugin files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal = runtime.RuntimeUpgradeStateAbnormal
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is currently reconciling.
	RuntimeUpgradeStateUpgradeRunning = runtime.RuntimeUpgradeStateUpgradeRunning
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed = runtime.RuntimeUpgradeStateUpgradeFailed
	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than the DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective = runtime.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is not semver-compatible.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed = runtime.RuntimeUpgradeAbnormalReasonVersionCompareFailed
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
	// RegisterCrons registers callback-contributed cron jobs for source plugins.
	RegisterCrons(ctx context.Context) error
	// SetCapabilities wires the host-published capability services used by source plugins.
	SetCapabilities(capabilities capability.Services)
	// ListExecutableCronJobs returns plugin-owned cron definitions whose
	// handlers are safe to publish for execution. Dynamic plugins must be in
	// an enabled business-entry state; disabled, pending-upgrade, abnormal, and
	// failed-upgrade dynamic plugins are excluded. Use this only for runtime
	// handler publication, not for authorization previews or task-table
	// projection.
	ListExecutableCronJobs(ctx context.Context) ([]ManagedCronJob, error)
	// ListExecutableCronJobsByPlugin returns executable cron definitions for
	// one plugin. It applies the same enablement and runtime-state rules as
	// ListExecutableCronJobs while narrowing discovery to pluginID, so callers
	// can register handlers during a plugin enable lifecycle without exposing
	// declarations that are not currently executable.
	ListExecutableCronJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error)
	// ListCronDeclarationsByPlugin returns declared cron metadata for one
	// plugin without requiring the plugin business entry to be enabled. This is
	// intended for management review and host-service authorization previews,
	// including not-yet-installed dynamic plugins. Callers must not publish the
	// returned handlers directly because the plugin may not be executable.
	ListCronDeclarationsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error)
	// ListInstalledCronDeclarations returns declared cron metadata for
	// installed plugins without requiring their business entries to be enabled.
	// Scheduled-job projection uses this to create or update task-table rows
	// for installed plugins while avoiding preview-only declarations from
	// uninstalled plugins.
	ListInstalledCronDeclarations(ctx context.Context) ([]ManagedCronJob, error)
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
	OrgProviderEnv(pluginID string) orgcapsvc.ProviderEnv
	// TenantProviderEnv returns typed, plugin-scoped tenant-provider construction inputs.
	TenantProviderEnv(pluginID string) tenantcapsvc.ProviderEnv
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
	// SetTenantStartupCapability wires tenant provider availability and startup consistency checks.
	SetTenantStartupCapability(service pluginTenantStartupCapability)
	// SetTenantProvisioningCapability wires tenant plugin auto-provisioning.
	SetTenantProvisioningCapability(service tenantcapsvc.PluginProvisioningService)
	// SetTenantPlatformGovernanceCapability wires platform-scope plugin governance checks.
	SetTenantPlatformGovernanceCapability(service platformGovernanceTenantCapability)
	// SetOrganizationCapability wires the runtime-owned organization capability used by plugin resource scopes.
	SetOrganizationCapability(service orgcapsvc.Service)
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
	// catalogSvc provides manifest discovery, registry, and release governance.
	catalogSvc catalog.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// sourceUpgradeSvc provides source-plugin upgrade discovery, execution, and startup validation.
	sourceUpgradeSvc sourceupgradeinternal.Service
	// frontendSvc manages in-memory frontend bundles for dynamic plugins.
	frontendSvc frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
	// capabilities exposes runtime-owned adapters for lazy provider construction.
	capabilities capability.Services
	// i18nSvc localizes plugin lifecycle messages and invalidates runtime
	// translation bundles after plugin lifecycle mutations.
	i18nSvc pluginI18nService
	// runtimeCacheRevisionCtrl coordinates process-local runtime caches in cluster deployments.
	runtimeCacheRevisionCtrl *runtimecache.Controller
	// runtimeUpgradeLockStore coordinates explicit runtime upgrades across cluster nodes.
	runtimeUpgradeLockStore coordination.LockStore
	// managementListCache stores the plugin-management summary read model.
	managementListCache *management.ListCache
	// tenantStartup validates tenant-governance startup state through a narrow tenant capability.
	tenantStartup pluginTenantStartupCapability
	// tenantProvisioning provisions tenant-scoped auto-enabled plugins after startup policy convergence.
	tenantProvisioning tenantcapsvc.PluginProvisioningService
	// tenantGovernance guards platform plugin-governance writes in HTTP paths.
	tenantGovernance platformGovernanceTenantCapability
	// runtimeUpgradeLocksMu protects process-local runtime-upgrade locks.
	runtimeUpgradeLocksMu sync.Mutex
	// runtimeUpgradeLocks serializes explicit runtime upgrades per plugin in the current process.
	runtimeUpgradeLocks map[string]*sync.Mutex
}

// New creates and returns a new plugin Service.
// Pass a non-nil topology for cluster-aware deployments; pass nil to use the
// default single-node topology implementation.
// reconcilerLockSvc must be created by the startup composition root and shared
// with host-lock services that use the same deployment-selected locker backend.
func New(
	topology Topology,
	configProvider configsvc.Service,
	bizCtxProvider bizctx.Service,
	cacheCoordSvc cachecoord.Service,
	i18nSvc i18nsvc.Service,
	sessionStore session.Store,
	reconcilerLockSvc locker.Service,
	runtimeUpgradeLockStore coordination.LockStore,
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
	if reconcilerLockSvc == nil {
		return nil, gerror.New("plugin service requires a non-nil reconciler lock service")
	}

	var topo Topology = singleNodeTopology{}
	if topology != nil {
		topo = topology
	}

	var (
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc, i18nSvc, reconcilerLockSvc)
		integrationSvc = integration.New(catalogSvc)
	)

	// Wire cross-package dependencies via setter injection so each sub-package
	// can be constructed independently without circular imports.
	catalogSvc.SetBackendLoader(integrationSvc)
	catalogSvc.SetArtifactParser(runtimeSvc)
	catalogSvc.SetDynamicManifestLoader(runtimeSvc)
	catalogSvc.SetNodeStateSyncer(runtimeSvc)
	catalogSvc.SetMenuSyncer(integrationSvc)
	catalogSvc.SetResourceRefSyncer(integrationSvc)
	catalogSvc.SetReleaseStateSyncer(runtimeSvc)
	catalogSvc.SetHookDispatcher(integrationSvc)

	lifecycleSvc.SetReconciler(runtimeSvc)
	lifecycleSvc.SetTopology(&lifecycleTopologyAdapter{topo})

	integrationSvc.SetBizCtxProvider(&bizCtxAdapter{bizCtxProvider})
	integrationSvc.SetTopologyProvider(&integrationTopologyAdapter{topo})
	integrationSvc.SetDynamicCronExecutor(runtimeSvc)

	runtimeSvc.SetTopology(&runtimeTopologyAdapter{topo})
	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{configProvider})
	runtimeSvc.SetUploadSizeProvider(&uploadSizeAdapter{configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{bizCtxProvider})
	runtimeSvc.SetSessionStore(sessionStore)

	service := &serviceImpl{
		configSvc:               configProvider,
		topology:                topo,
		catalogSvc:              catalogSvc,
		lifecycleSvc:            lifecycleSvc,
		runtimeSvc:              runtimeSvc,
		integrationSvc:          integrationSvc,
		frontendSvc:             frontendSvc,
		openapiSvc:              openapiSvc,
		i18nSvc:                 i18nSvc,
		runtimeUpgradeLockStore: runtimeUpgradeLockStore,
		managementListCache:     management.NewListCache(),
		runtimeUpgradeLocks:     make(map[string]*sync.Mutex),
	}
	service.runtimeCacheRevisionCtrl = newRuntimeCacheRevisionController(
		topo,
		cacheCoordSvc,
		integrationSvc,
		frontendSvc,
		i18nSvc,
		service,
	)
	runtimeSvc.SetRuntimeCacheChangeNotifier(service)
	runtimeSvc.SetDependencyValidator(service)
	if err := runtimeSvc.ValidateRequiredDependencies(); err != nil {
		return nil, gerror.Wrap(err, "plugin runtime wiring validation failed")
	}
	service.sourceUpgradeSvc = sourceupgradeinternal.New(catalogSvc, lifecycleSvc, runtimeSvc, integrationSvc, i18nSvc, service)
	return service, nil
}
