// Package integration bridges pluginhost callback registrations and declared plugin
// configurations into the host route, menu, permission, and lifecycle integration flows.
package integration

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/pluginhost"
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

// ManagedJob describes one plugin-owned scheduled-job definition that the
// host can project into the unified scheduled-job management table.
type ManagedJob struct {
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
	Handler pluginhost.JobHandler
}

// sourceServicesForPlugin returns the source-plugin-only service view at the
// callback boundary after the common capability services are scoped.
func (s *serviceImpl) sourceServicesForPlugin(pluginID string) capability.Services {
	if s == nil || s.capabilities == nil {
		return nil
	}
	return capabilityowner.ServicesForPlugin(s.capabilities, pluginID)
}

// filterRuntime holds a snapshot of which plugins are currently enabled for use
// by menu and permission filters within a single request.
type filterRuntime struct {
	manifests   []*catalog.Manifest
	enabledByID map[string]bool
}

// sharedState stores process-wide integration caches used by source-plugin
// route guards and route-binding projections.
type sharedState struct {
	sourceRouteBindingsMu sync.RWMutex
	sourceRouteBindings   map[string][]pluginhost.SourceRouteBinding

	enabledSnapshotMu     sync.RWMutex
	enabledSnapshot       map[string]bool
	enabledSnapshotLoaded bool
}

// newSharedState creates an integration state holder for one host composition root.
func newSharedState() *sharedState {
	return &sharedState{
		sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
		enabledSnapshot:     make(map[string]bool),
	}
}

// Service defines the integration service contract used by the root plugin
// service, runtime reconciler, lifecycle orchestrator, and upgrade planner.
type Service interface {
	// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
	// It implements catalog.BackendConfigLoader.
	LoadPluginBackendConfig(manifest *catalog.Manifest) error

	// ListResourceRecords queries plugin-owned backend resource rows using the
	// generic plugin resource contract.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
	// ResolveResourcePermission resolves the permission required by the generic
	// resource list endpoint for one plugin-owned backend resource.
	ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error)

	// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
	ListSourceRouteBindings() []pluginhost.SourceRouteBinding
	// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
	RegisterHTTPRoutes(
		ctx context.Context,
		server *ghttp.Server,
		pluginGroup *ghttp.RouterGroup,
		middlewares pluginhost.RouteMiddlewares,
	) error
	// RegisterJobs registers callback-contributed scheduled jobs for source plugins.
	RegisterJobs(ctx context.Context) error
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

	// DispatchPluginHookEvent dispatches one named hook event to all enabled plugins.
	// It implements catalog hook dispatch and runtime.IntegrationService.
	DispatchPluginHookEvent(
		ctx context.Context,
		eventName pluginhost.ExtensionPoint,
		payload map[string]interface{},
	) error

	// FilterMenus filters disabled plugin menus by menu_key prefix "plugin:<plugin-id>".
	FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
	// FilterPermissionMenus filters permission menus based on plugin enablement and plugin-defined permission visibility.
	// It implements runtime.IntegrationService.
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

	// WithStartupDataSnapshot returns a child context carrying full-table
	// snapshots for small plugin integration tables during startup reconciliation.
	WithStartupDataSnapshot(ctx context.Context) (context.Context, error)

	// CanExposeBusinessEntries reports whether the plugin with the given ID can expose business entries.
	CanExposeBusinessEntries(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports whether pluginID is platform-enabled for capability provider use.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsInstalledEnabledForTenant reports whether the plugin is installed, enabled, and
	// available for the current tenant without applying business-entry upgrade gates.
	IsInstalledEnabledForTenant(ctx context.Context, pluginID string) bool
	// SetTenantPluginEnabledState persists one tenant-scoped plugin enablement row.
	SetTenantPluginEnabledState(ctx context.Context, pluginID string, tenantID int, enabled bool) error
	// RefreshEnabledSnapshot rebuilds the in-memory business-entry snapshot used by runtime guards.
	RefreshEnabledSnapshot(ctx context.Context) error
	// SetPluginEnabledState updates one plugin entry in the in-memory business-entry snapshot.
	SetPluginEnabledState(pluginID string, enabled bool)
	// DeletePluginEnabledState removes one plugin entry from the in-memory business-entry snapshot.
	DeletePluginEnabledState(pluginID string)

	// SyncPluginMenusAndPermissions reconciles all manifest menus and dynamic route permission
	// entries into sys_menu.
	// It implements runtime.IntegrationService and catalog.MenuSyncer.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginMenus reconciles only the manifest-declared menus, skipping route-permission entries.
	// Used during reconciler rollback to restore the previous menu state without touching permissions.
	// It implements runtime.IntegrationService.
	SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
	// It implements runtime.IntegrationService.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
	// ListPluginMenusByPlugin is the exported form of listPluginMenusByPlugin for cross-package access.
	ListPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error)

	// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the
	// current governance resource index derived from the given manifest.
	// It implements catalog.ResourceRefSyncer.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
	// ListPluginResourceRefs is the exported form of listPluginResourceRefs for cross-package access.
	ListPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error)
	// BuildResourceRefDescriptors is the exported form of buildPluginResourceRefDescriptors for cross-package access.
	BuildResourceRefDescriptors(manifest *catalog.Manifest) []*plugintypes.ResourceRefDescriptor
}

// Ensure serviceImpl satisfies the integration contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	catalogSvc catalog.Service
	storeSvc   store.Service

	bizCtxSvc bizctx.Service

	topology cluster.Service

	capabilities capability.Services

	orgSvc orgcap.Service

	runtimeSvc runtime.Service

	sharedState *sharedState
}

// New creates and returns a new integration Service.
func New(
	catalogSvc catalog.Service,
	storeSvc store.Service,
	bizCtxSvc bizctx.Service,
	topology cluster.Service,
	capabilities capability.Services,
	orgSvc orgcap.Service,
	runtimeSvc runtime.Service,
) Service {
	return &serviceImpl{
		catalogSvc:   catalogSvc,
		storeSvc:     storeSvc,
		bizCtxSvc:    bizCtxSvc,
		topology:     topology,
		capabilities: capabilities,
		orgSvc:       orgSvc,
		runtimeSvc:   runtimeSvc,
		sharedState:  newSharedState(),
	}
}
