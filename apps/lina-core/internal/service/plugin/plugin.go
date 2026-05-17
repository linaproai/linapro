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
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	sourceupgradeinternal "lina-core/internal/service/plugin/internal/sourceupgrade"
	"lina-core/internal/service/pluginruntimecache"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"

	"lina-core/internal/model/entity"

	"lina-core/pkg/pluginhost"
	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

type (
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
	// (no mock data, no host-service authorization snapshot).
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
		// dependencyResult records the server-side dependency plan and automatic
		// installation result produced during this install request.
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

// PluginItem is the display-ready projection of one plugin entry.
type PluginItem struct {
	runtime.PluginItem
	// DependencyCheck carries server-side dependency status for management UIs.
	DependencyCheck *DependencyCheckResult
}

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

// RuntimeUpgradeSQLSummary summarizes manifest SQL assets visible to preview.
type RuntimeUpgradeSQLSummary struct {
	// InstallSQLCount is the number of target install/upgrade SQL assets.
	InstallSQLCount int
	// UninstallSQLCount is the number of target uninstall SQL assets.
	UninstallSQLCount int
	// MockSQLCount is the number of target mock SQL assets excluded from upgrade.
	MockSQLCount int
	// RuntimeSQLAssetCount is the dynamic artifact SQL section count when present.
	RuntimeSQLAssetCount int
}

// RuntimeUpgradeHostServicesDiff summarizes service-level hostServices drift.
type RuntimeUpgradeHostServicesDiff struct {
	// Added contains services declared by the target manifest but not by the effective snapshot.
	Added []*RuntimeUpgradeHostServiceChange
	// Removed contains services no longer requested by the target manifest.
	Removed []*RuntimeUpgradeHostServiceChange
	// Changed contains services whose methods or governed targets changed.
	Changed []*RuntimeUpgradeHostServiceChange
	// AuthorizationRequired reports whether the target manifest needs host confirmation.
	AuthorizationRequired bool
	// AuthorizationChanged reports whether requested host service scope changed.
	AuthorizationChanged bool
}

// RuntimeUpgradeHostServiceChange summarizes one service-level hostServices change.
type RuntimeUpgradeHostServiceChange struct {
	// Service is the logical host service identifier.
	Service string
	// FromMethods is the effective method set before upgrade.
	FromMethods []string
	// ToMethods is the target method set after upgrade.
	ToMethods []string
	// FromResourceCount is the number of governed targets before upgrade.
	FromResourceCount int
	// ToResourceCount is the number of governed targets after upgrade.
	ToResourceCount int
	// FromTables is the effective data-table set before upgrade.
	FromTables []string
	// ToTables is the target data-table set after upgrade.
	ToTables []string
	// FromPaths is the effective storage path set before upgrade.
	FromPaths []string
	// ToPaths is the target storage path set after upgrade.
	ToPaths []string
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

// ListOutput defines output for plugin list query.
type ListOutput struct {
	// List contains the filtered plugin list.
	List []*PluginItem
	// Total is the number of returned plugins.
	Total int
}

// ListInput defines input for plugin list query.
type ListInput struct {
	// ID filters by plugin identifier.
	ID string
	// Name filters by plugin display name.
	Name string
	// Type filters by normalized plugin type.
	Type string
	// Status filters by enabled flag.
	Status *int
	// Installed filters by installed flag.
	Installed *int
}

// AuthLoginSucceededInput defines input for auth hook events.
type AuthLoginSucceededInput struct {
	// UserName is the authenticated username.
	UserName string
	// Status is the login status code.
	Status int
	// Ip is the client IP address.
	Ip string
	// ClientType identifies the login client type.
	ClientType string
	// Browser is the detected browser description.
	Browser string
	// Os is the detected operating-system description.
	Os string
	// Message is the audit message delivered to plugins.
	Message string
	// Reason is the stable auth lifecycle reason code delivered to plugins.
	Reason string
}

// AuthHookService defines auth-related plugin hook operations.
type AuthHookService interface {
	// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
	HandleAuthLoginSucceeded(ctx context.Context, input AuthLoginSucceededInput) error
	// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
	HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error
	// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
	HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error
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
	// SetHostServices wires the host-published service directory used by source plugins.
	SetHostServices(services pluginhost.HostServices)
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
	ListSourceUpgradeStatuses(ctx context.Context) ([]*sourceupgradecontract.SourcePluginStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
	// current effective version to the newer discovered source version.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*sourceupgradecontract.SourcePluginUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades; list/runtime state exposes the result.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
	// ValidateStartupConsistency fails fast when persisted plugin and tenant
	// governance state is incoherent before routes are served.
	ValidateStartupConsistency(ctx context.Context) error
	// SetTenantCapability wires the runtime-owned tenant capability used by
	// startup consistency checks that span plugin and tenant governance.
	SetTenantCapability(service tenantcapsvc.Service)
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
	// List returns the read-only plugin list with optional in-memory filtering applied.
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
	// i18nSvc invalidates runtime translation bundles after plugin lifecycle mutations.
	i18nSvc runtimeBundleInvalidator
	// runtimeCacheRevisionCtrl coordinates process-local runtime caches in cluster deployments.
	runtimeCacheRevisionCtrl *pluginruntimecache.Controller
	// runtimeUpgradeLockStore coordinates explicit runtime upgrades across cluster nodes.
	runtimeUpgradeLockStore coordination.LockStore
	// tenantSvc validates tenant-governance startup state through the runtime-owned tenant capability.
	tenantSvc tenantcapsvc.Service
	// runtimeUpgradeLocksMu protects process-local runtime-upgrade locks.
	runtimeUpgradeLocksMu sync.Mutex
	// runtimeUpgradeLocks serializes explicit runtime upgrades per plugin in the current process.
	runtimeUpgradeLocks map[string]*sync.Mutex
}

// New creates and returns a new plugin Service.
// Pass a non-nil topology for cluster-aware deployments; pass nil to use the
// default single-node topology implementation.
func New(
	topology Topology,
	configProvider configsvc.Service,
	bizCtxProvider bizctx.Service,
	cacheCoordSvc cachecoord.Service,
	i18nSvc i18nsvc.Service,
	sessionStore session.Store,
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

	var topo Topology = singleNodeTopology{}
	if topology != nil {
		topo = topology
	}

	var (
		catalogSvc       = catalog.New(configProvider)
		lifecycleSvc     = lifecycle.New(catalogSvc)
		frontendSvc      = frontend.New(catalogSvc)
		openapiSvc       = openapi.New(catalogSvc)
		runtimeSvc       = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc, i18nSvc)
		integrationSvc   = integration.New(catalogSvc)
		cacheRevisionCtl = newRuntimeCacheRevisionController(
			topo,
			cacheCoordSvc,
			integrationSvc,
			frontendSvc,
			i18nSvc,
		)
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
		configSvc:                configProvider,
		topology:                 topo,
		catalogSvc:               catalogSvc,
		lifecycleSvc:             lifecycleSvc,
		runtimeSvc:               runtimeSvc,
		integrationSvc:           integrationSvc,
		frontendSvc:              frontendSvc,
		openapiSvc:               openapiSvc,
		i18nSvc:                  i18nSvc,
		runtimeCacheRevisionCtrl: cacheRevisionCtl,
		runtimeUpgradeLockStore:  runtimeUpgradeLockStore,
		runtimeUpgradeLocks:      make(map[string]*sync.Mutex),
	}
	runtimeSvc.SetRuntimeCacheChangeNotifier(service)
	runtimeSvc.SetDependencyValidator(service)
	service.sourceUpgradeSvc = sourceupgradeinternal.New(catalogSvc, lifecycleSvc, runtimeSvc, integrationSvc, i18nSvc, service)
	return service, nil
}
