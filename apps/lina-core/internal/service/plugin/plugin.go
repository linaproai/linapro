// Package plugin implements plugin manifest discovery, lifecycle orchestration,
// governance metadata synchronization, and host integration for Lina plugins.
package plugin

import (
	"context"
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"

	// Re-export sub-package types that are referenced by controllers and callers.
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	// PluginItem is the display-ready projection of one plugin entry.
	"lina-core/pkg/pluginhost"
)

type (
	PluginItem = runtime.PluginItem

	// DynamicUploadInput defines input for uploading a runtime WASM package.
	DynamicUploadInput = runtime.DynamicUploadInput

	// DynamicUploadOutput defines output for uploading a runtime WASM package.
	DynamicUploadOutput = runtime.DynamicUploadOutput

	// RuntimeStateListOutput defines output for public runtime state queries.
	RuntimeStateListOutput = runtime.RuntimeStateListOutput

	// ResourceListInput defines input for querying a plugin-owned backend resource.
	ResourceListInput = integration.ResourceListInput

	// ResourceListOutput defines output for querying a plugin-owned backend resource.
	ResourceListOutput = integration.ResourceListOutput

	// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
	RuntimeFrontendAssetOutput = frontend.RuntimeFrontendAssetOutput

	// DynamicRouteOperLogMetadata stores operation-log metadata for dynamic routes.
	DynamicRouteOperLogMetadata = runtime.DynamicRouteOperLogMetadata

	// PluginDynamicStateItem represents public runtime state of one plugin.
	PluginDynamicStateItem = runtime.PluginDynamicStateItem

	// HostServiceAuthorizationInput defines one install/enable authorization confirmation payload.
	HostServiceAuthorizationInput = catalog.HostServiceAuthorizationInput

	// HostServiceAuthorizationDecision narrows one authorized service snapshot.
	HostServiceAuthorizationDecision = catalog.HostServiceAuthorizationDecision
)

// GetDynamicRouteOperLogMetadata returns dynamic-route operation-log metadata from the request.
// This package-level function is retained for callers that cannot import the runtime sub-package.
var GetDynamicRouteOperLogMetadata = runtime.GetDynamicRouteOperLogMetadata

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
}

// Service defines the plugin service contract.
type Service interface {
	// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
	HandleAuthLoginSucceeded(ctx context.Context, input AuthLoginSucceededInput) error
	// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
	HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error
	// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
	HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error
	// ResolveDataTableComments resolves host-side table comments for the given
	// data-table names. It degrades to an empty map when metadata lookup is
	// unavailable so plugin list APIs are not blocked by optional schema comments.
	ResolveDataTableComments(ctx context.Context, tables []string) map[string]string
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
	// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
	RegisterHTTPRoutes(
		ctx context.Context,
		pluginGroup *ghttp.RouterGroup,
		middlewares pluginhost.RouteMiddlewares,
	) error
	// RegisterCrons registers callback-contributed cron jobs for source plugins.
	RegisterCrons(ctx context.Context) error
	// DispatchAfterAuthRequest dispatches callback-style after-auth request handlers.
	DispatchAfterAuthRequest(ctx context.Context, input pluginhost.AfterAuthInput)
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
	// ListResourceRecords queries plugin-owned backend resource rows.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
	// Install executes the install lifecycle and optionally persists one host-confirmed
	// host service authorization snapshot when the target is a dynamic plugin.
	Install(
		ctx context.Context,
		pluginID string,
		authorization *HostServiceAuthorizationInput,
	) error
	// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
	Uninstall(ctx context.Context, pluginID string) error
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
	// IsInstalled returns whether a plugin is installed.
	IsInstalled(ctx context.Context, pluginID string) bool
	// IsEnabled returns whether a plugin is enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
	SyncSourcePlugins(ctx context.Context) error
	// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
	// returns the combined list of source and dynamic plugin items.
	SyncAndList(ctx context.Context) (*ListOutput, error)
	// List returns the plugin list with optional in-memory filtering applied.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
	// StartRuntimeReconciler starts the background reconciler loop for dynamic plugins.
	StartRuntimeReconciler(ctx context.Context)
	// ReconcileRuntimePlugins runs one reconciliation pass for all dynamic plugins.
	ReconcileRuntimePlugins(ctx context.Context) error
	// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
	ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error)
	// UploadDynamicPackage validates and stores a runtime WASM package.
	UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error)
	// PrepareDynamicRouteMiddleware prepares dynamic route state before the main handler.
	PrepareDynamicRouteMiddleware(r *ghttp.Request)
	// AuthenticateDynamicRouteMiddleware authenticates JWT tokens for dynamic routes.
	AuthenticateDynamicRouteMiddleware(r *ghttp.Request)
	// RegisterDynamicRouteDispatcher binds the dynamic route catch-all handler to the group.
	RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery, registry, and release governance.
	catalogSvc catalog.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// frontendSvc manages in-memory frontend bundles for dynamic plugins.
	frontendSvc frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc openapi.Service
}

// New creates and returns a new plugin Service. An optional Topology may be
// provided for cluster-aware deployments; single-node mode is the default.
func New(topologies ...Topology) Service {
	var topo Topology = singleNodeTopology{}
	if len(topologies) > 0 && topologies[0] != nil {
		topo = topologies[0]
	}

	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc)
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

	runtimeSvc.SetTopology(&runtimeTopologyAdapter{topo})
	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetAfterAuthDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{bizCtxProvider})

	return &serviceImpl{
		catalogSvc:     catalogSvc,
		lifecycleSvc:   lifecycleSvc,
		runtimeSvc:     runtimeSvc,
		integrationSvc: integrationSvc,
		frontendSvc:    frontendSvc,
		openapiSvc:     openapiSvc,
	}
}
