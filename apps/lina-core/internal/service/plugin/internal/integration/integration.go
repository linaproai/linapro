// Package integration bridges pluginhost callback registrations and declared plugin
// configurations into the host route, menu, permission, and lifecycle integration flows.

package integration

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"

	"github.com/gogf/gf/v2/net/ghttp"
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

// isEnabled reports whether the plugin with the given ID is currently enabled.
func (r *filterRuntime) isEnabled(pluginID string) bool {
	if r == nil {
		return false
	}
	return r.enabledByID[strings.TrimSpace(pluginID)]
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
	// ListManagedCronJobs returns plugin-owned cron definitions for scheduled-job projection.
	ListManagedCronJobs(ctx context.Context) ([]ManagedCronJob, error)
	// ListManagedCronJobsByPlugin returns cron definitions owned by one plugin.
	ListManagedCronJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error)
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
}

// PluginStateService defines plugin enablement lookup operations.
type PluginStateService interface {
	// IsEnabled reports whether the plugin with the given ID is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// RefreshEnabledSnapshot rebuilds the in-memory enablement snapshot used by runtime guards.
	RefreshEnabledSnapshot(ctx context.Context) error
	// SetPluginEnabledState updates one plugin entry in the in-memory enablement snapshot.
	SetPluginEnabledState(pluginID string, enabled bool)
	// DeletePluginEnabledState removes one plugin entry from the in-memory enablement snapshot.
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

	sharedState *sharedState
}

// New creates and returns a new integration Service.
func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{
		catalogSvc:  catalogSvc,
		sharedState: defaultSharedState,
	}
}

// SetBizCtxProvider wires the business-context provider used by route handlers.
func (s *serviceImpl) SetBizCtxProvider(p BizCtxProvider) {
	s.bizCtxSvc = p
}

// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
func (s *serviceImpl) SetTopologyProvider(t TopologyProvider) {
	s.topology = t
}

// SetDynamicCronExecutor wires the runtime executor used by declared
// dynamic-plugin cron jobs.
func (s *serviceImpl) SetDynamicCronExecutor(executor DynamicCronExecutor) {
	s.dynamicCronExecutor = executor
}

// IsEnabled reports whether the plugin with the given ID is currently installed and enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled
}

// RefreshEnabledSnapshot rebuilds the in-memory enablement snapshot used by runtime guards.
func (s *serviceImpl) RefreshEnabledSnapshot(ctx context.Context) error {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return err
	}
	enabledByID, err := s.buildEnabledPluginMapFromCatalog(ctx, manifests, false)
	if err != nil {
		return err
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot = enabledByID
	s.sharedState.enabledSnapshotLoaded = true
	return nil
}

// SetPluginEnabledState updates one plugin entry in the in-memory enablement snapshot.
func (s *serviceImpl) SetPluginEnabledState(pluginID string, enabled bool) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot[normalizedPluginID] = enabled
	s.sharedState.enabledSnapshotLoaded = true
}

// DeletePluginEnabledState removes one plugin entry from the in-memory enablement snapshot.
func (s *serviceImpl) DeletePluginEnabledState(pluginID string) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	delete(s.sharedState.enabledSnapshot, normalizedPluginID)
	s.sharedState.enabledSnapshotLoaded = true
}

// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
func (s *serviceImpl) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	s.sharedState.sourceRouteBindingsMu.RLock()
	defer s.sharedState.sourceRouteBindingsMu.RUnlock()

	items := make([]pluginhost.SourceRouteBinding, 0)
	for _, bindings := range s.sharedState.sourceRouteBindings {
		items = append(items, pluginhost.CloneSourceRouteBindings(bindings)...)
	}
	return items
}

// buildFilterRuntime builds a filter runtime by scanning all manifests and loading
// the current enablement status for each discovered plugin.
func (s *serviceImpl) buildFilterRuntime(ctx context.Context) (*filterRuntime, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	return s.buildFilterRuntimeFromManifests(ctx, manifests)
}

// buildFilterRuntimeFromManifests builds a filter runtime for the given manifest list.
func (s *serviceImpl) buildFilterRuntimeFromManifests(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (*filterRuntime, error) {
	enabledByID, err := s.buildEnabledPluginMap(ctx, manifests)
	if err != nil {
		return nil, err
	}
	return &filterRuntime{
		manifests:   manifests,
		enabledByID: enabledByID,
	}, nil
}

// buildEnabledPluginMap queries the registry table for the installed/enabled state
// of each plugin in the manifest list.
func (s *serviceImpl) buildEnabledPluginMap(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (map[string]bool, error) {
	return s.buildEnabledPluginMapFromCatalog(ctx, manifests, true)
}

// buildEnabledPluginMapFromCatalog queries or reuses registry state for the
// supplied manifests. Refresh callers can disable snapshot reuse to rebuild the
// process-wide view after lifecycle changes.
func (s *serviceImpl) buildEnabledPluginMapFromCatalog(
	ctx context.Context,
	manifests []*catalog.Manifest,
	allowLoadedSnapshot bool,
) (map[string]bool, error) {
	var (
		enabledByID = make(map[string]bool, len(manifests))
		pluginIDs   = make([]string, 0, len(manifests))
	)
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		if _, ok := enabledByID[pluginID]; ok {
			continue
		}
		enabledByID[pluginID] = false
		pluginIDs = append(pluginIDs, pluginID)
	}
	if len(pluginIDs) == 0 {
		return enabledByID, nil
	}
	if allowLoadedSnapshot && s.applyLoadedEnabledSnapshot(enabledByID) {
		return enabledByID, nil
	}

	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	s.storeLoadedEnabledSnapshot(registries)

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if _, ok := enabledByID[pluginID]; !ok {
			continue
		}
		enabledByID[pluginID] = registry.Installed == catalog.InstalledYes &&
			registry.Status == catalog.StatusEnabled
	}
	return enabledByID, nil
}

// storeLoadedEnabledSnapshot refreshes the process-local enablement snapshot
// from one registry read so later filters in the same process can reuse it.
func (s *serviceImpl) storeLoadedEnabledSnapshot(registries []*entity.SysPlugin) {
	if s == nil || s.sharedState == nil {
		return
	}
	snapshot := make(map[string]bool, len(registries))
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if pluginID == "" {
			continue
		}
		snapshot[pluginID] = registry.Installed == catalog.InstalledYes &&
			registry.Status == catalog.StatusEnabled
	}
	s.sharedState.enabledSnapshotMu.Lock()
	defer s.sharedState.enabledSnapshotMu.Unlock()
	s.sharedState.enabledSnapshot = snapshot
	s.sharedState.enabledSnapshotLoaded = true
}

// applyLoadedEnabledSnapshot copies the process-local enablement snapshot into
// the requested plugin map when a lifecycle path has already warmed it.
func (s *serviceImpl) applyLoadedEnabledSnapshot(enabledByID map[string]bool) bool {
	if s == nil || s.sharedState == nil || len(enabledByID) == 0 {
		return false
	}
	s.sharedState.enabledSnapshotMu.RLock()
	defer s.sharedState.enabledSnapshotMu.RUnlock()
	if !s.sharedState.enabledSnapshotLoaded {
		return false
	}
	for pluginID := range enabledByID {
		enabledByID[pluginID] = s.sharedState.enabledSnapshot[pluginID]
	}
	return true
}

// buildBackgroundEnabledChecker returns a PluginEnabledChecker for use in source plugin
// route and cron registrars that need to guard runtime access.
func (s *serviceImpl) buildBackgroundEnabledChecker() pluginhost.PluginEnabledChecker {
	return func(pluginID string) bool {
		normalizedPluginID := strings.TrimSpace(pluginID)
		if normalizedPluginID == "" {
			return false
		}

		s.sharedState.enabledSnapshotMu.RLock()
		enabled, ok := s.sharedState.enabledSnapshot[normalizedPluginID]
		loaded := s.sharedState.enabledSnapshotLoaded
		s.sharedState.enabledSnapshotMu.RUnlock()
		if ok || loaded {
			return enabled
		}
		return s.IsEnabled(context.Background(), normalizedPluginID)
	}
}

// buildPrimaryNodeChecker returns a PrimaryNodeChecker for use in source plugin cron registrars.
func (s *serviceImpl) buildPrimaryNodeChecker() pluginhost.PrimaryNodeChecker {
	return func() bool {
		if s.topology == nil {
			return false
		}
		return s.topology.IsPrimaryNode()
	}
}

// setSourceRouteBindings stores the latest host-captured route bindings for one
// source plugin after registration completes.
func (s *serviceImpl) setSourceRouteBindings(pluginID string, bindings []pluginhost.SourceRouteBinding) {
	s.sharedState.sourceRouteBindingsMu.Lock()
	defer s.sharedState.sourceRouteBindingsMu.Unlock()
	s.sharedState.sourceRouteBindings[strings.TrimSpace(pluginID)] = pluginhost.CloneSourceRouteBindings(bindings)
}
