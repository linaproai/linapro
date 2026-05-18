// Package integration bridges pluginhost callback registrations and declared plugin
// configurations into the host route, menu, permission, and lifecycle integration flows.
package integration

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

const (
	// pluginTenantEnablementStateKey is the sys_plugin_state state_key used for
	// tenant-scoped plugin enablement rows.
	pluginTenantEnablementStateKey = "__tenant_enabled__"
	// pluginTenantEnabledValue stores the string value for an enabled tenant state.
	pluginTenantEnabledValue = "enabled"
	// pluginTenantDisabledValue stores the string value for a disabled tenant state.
	pluginTenantDisabledValue = "disabled"
)

// ManagedCronJob describes one plugin-owned scheduled-job definition that the
// host can project into the unified scheduled-job management table.
type ManagedCronJob struct {
	// PluginID identifies the owning plugin.
	PluginID string
	// Name is the stable plugin-local job name.
	Name string
	// DisplayName is the human-readable name shown in the UI.
	DisplayName string
	// Description explains the job purpose in the UI.
	Description string
	// Pattern stores the raw gcron pattern declared by the plugin.
	Pattern string
	// Timezone stores the UI display timezone when the pattern is cron-based.
	Timezone string
	// Scope selects master-only or all-node execution.
	Scope jobmeta.JobScope
	// Concurrency selects the overlap policy.
	Concurrency jobmeta.JobConcurrency
	// MaxConcurrency caps parallel overlap when Concurrency=parallel.
	MaxConcurrency int
	// Timeout bounds each execution.
	Timeout time.Duration
	// Handler executes the plugin-owned scheduled job.
	Handler pluginhost.CronJobHandler
}

// DynamicCronExecutor executes one dynamic-plugin declared cron job through the
// active runtime bridge.
type DynamicCronExecutor interface {
	// DiscoverCronContracts collects all dynamic-plugin cron declarations from
	// the plugin runtime's reserved registration entry point.
	DiscoverCronContracts(
		ctx context.Context,
		manifest *catalog.Manifest,
	) ([]*pluginbridge.CronContract, error)
	// ExecuteDeclaredCronJob runs one declared dynamic-plugin cron job against
	// the active manifest/runtime.
	ExecuteDeclaredCronJob(
		ctx context.Context,
		manifest *catalog.Manifest,
		contract *pluginbridge.CronContract,
	) error
}

// BizCtxProvider abstracts the business context dependency for data-scope queries.
type BizCtxProvider interface {
	// GetUserId returns the user ID stored in the current request business context.
	GetUserId(ctx context.Context) int
	// GetDataScope returns the effective role data-scope stored in the current request business context.
	GetDataScope(ctx context.Context) int
	// GetDataScopeUnsupported returns the unsupported data-scope state stored in the current request business context.
	GetDataScopeUnsupported(ctx context.Context) (bool, int)
}

// TopologyProvider abstracts cluster topology for primary-node routing decisions.
type TopologyProvider interface {
	// IsPrimaryNode reports whether this host instance is the designated primary node.
	IsPrimaryNode() bool
}

// filterRuntime holds a snapshot of which plugins are currently enabled for use
// by menu and permission filters within a single request.
type filterRuntime struct {
	manifests   []*catalog.Manifest
	enabledByID map[string]bool
}

// BackendConfigService defines manifest backend-declaration loading operations.
type BackendConfigService interface {
	// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
	// It implements catalog.BackendConfigLoader.
	LoadPluginBackendConfig(manifest *catalog.Manifest) error
}

// ResourceQueryService defines plugin-owned backend resource query operations.
type ResourceQueryService interface {
	// ListResourceRecords queries plugin-owned backend resource rows using the
	// generic plugin resource contract.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
	// ResolveResourcePermission resolves the permission required by the generic
	// resource list endpoint for one plugin-owned backend resource.
	ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error)
}

// SourceRegistrationService defines source-plugin route and cron registration operations.
type SourceRegistrationService interface {
	// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
	ListSourceRouteBindings() []pluginhost.SourceRouteBinding
	// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
	RegisterHTTPRoutes(
		ctx context.Context,
		server *ghttp.Server,
		pluginGroup *ghttp.RouterGroup,
		middlewares pluginhost.RouteMiddlewares,
	) error
	// RegisterCrons registers callback-contributed cron jobs for source plugins.
	RegisterCrons(ctx context.Context) error
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
}

// HookDispatchService defines plugin hook dispatch operations.
type HookDispatchService interface {
	// DispatchPluginHookEvent dispatches one named hook event to all enabled plugins.
	// It implements catalog.HookDispatcher and runtime.HookDispatcher.
	DispatchPluginHookEvent(
		ctx context.Context,
		eventName pluginhost.ExtensionPoint,
		payload map[string]interface{},
	) error
}

// MenuFilterService defines menu filtering operations based on plugin state.
type MenuFilterService interface {
	// FilterMenus filters disabled plugin menus by menu_key prefix "plugin:<plugin-id>".
	FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
	// FilterPermissionMenus filters permission menus based on plugin enablement and plugin-defined permission visibility.
	// It implements runtime.PermissionMenuFilter.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
	// ShouldKeepPermission reports whether a permission should stay effective after plugin filtering.
	ShouldKeepPermission(ctx context.Context, menu *entity.SysMenu) bool
	// RunPluginDeclaredHook is the exported form of runPluginDeclaredHook for cross-package access.
	RunPluginDeclaredHook(
		ctx context.Context,
		pluginID string,
		hook *catalog.HookSpec,
		payload map[string]interface{},
	) error
}

// DependencyWiringService defines provider wiring operations required by integration runtime.
type DependencyWiringService interface {
	// WithStartupDataSnapshot returns a child context carrying full-table
	// snapshots for small plugin integration tables during startup reconciliation.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)
	// SetBizCtxProvider wires the business-context provider used by route handlers.
	SetBizCtxProvider(p BizCtxProvider)
	// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
	SetTopologyProvider(t TopologyProvider)
	// SetDynamicCronExecutor wires the runtime executor used by declared
	// dynamic-plugin cron jobs.
	SetDynamicCronExecutor(executor DynamicCronExecutor)
	// SetHostServices wires the host-published service directory used by source plugins.
	SetHostServices(services pluginhost.HostServices)
}

// PluginStateService defines plugin enablement lookup operations.
type PluginStateService interface {
	// IsEnabled reports whether the plugin with the given ID can expose business entries.
	IsEnabled(ctx context.Context, pluginID string) bool
	// SetTenantPluginEnabledState persists one tenant-scoped plugin enablement row.
	SetTenantPluginEnabledState(ctx context.Context, pluginID string, tenantID int, enabled bool) error
	// RefreshEnabledSnapshot rebuilds the in-memory business-entry snapshot used by runtime guards.
	RefreshEnabledSnapshot(ctx context.Context) error
	// SetPluginEnabledState updates one plugin entry in the in-memory business-entry snapshot.
	SetPluginEnabledState(pluginID string, enabled bool)
	// DeletePluginEnabledState removes one plugin entry from the in-memory business-entry snapshot.
	DeletePluginEnabledState(pluginID string)
}

// MenuSyncService defines plugin menu synchronization operations.
type MenuSyncService interface {
	// SyncPluginMenusAndPermissions reconciles all manifest menus and dynamic route permission
	// entries into sys_menu.
	// It implements runtime.MenuManager and catalog.MenuSyncer.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginMenus reconciles only the manifest-declared menus, skipping route-permission entries.
	// Used during reconciler rollback to restore the previous menu state without touching permissions.
	// It implements runtime.MenuManager.
	SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
	// It implements runtime.MenuManager.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
	// ListPluginMenusByPlugin is the exported form of listPluginMenusByPlugin for cross-package access.
	ListPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error)
}

// ResourceReferenceService defines plugin resource-reference synchronization operations.
type ResourceReferenceService interface {
	// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the
	// current governance resource index derived from the given manifest.
	// It implements catalog.ResourceRefSyncer.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
	// ListPluginResourceRefs is the exported form of listPluginResourceRefs for cross-package access.
	ListPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error)
	// BuildResourceRefDescriptors is the exported form of buildPluginResourceRefDescriptors for cross-package access.
	BuildResourceRefDescriptors(manifest *catalog.Manifest) []*catalog.ResourceRefDescriptor
}

// Service defines the integration service contract by composing integration sub-capabilities.
type Service interface {
	BackendConfigService
	ResourceQueryService
	SourceRegistrationService
	HookDispatchService
	MenuFilterService
	DependencyWiringService
	PluginStateService
	MenuSyncService
	ResourceReferenceService
}

// Ensure serviceImpl satisfies the composed integration contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	catalogSvc catalog.Service

	bizCtxSvc BizCtxProvider

	topology TopologyProvider

	dynamicCronExecutor DynamicCronExecutor

	hostServices pluginhost.HostServices

	sharedState *sharedState
}

// New creates and returns a new integration Service.
func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{
		catalogSvc:  catalogSvc,
		sharedState: defaultSharedState,
	}
}
