// Package integration bridges pluginhost callback registrations and declared plugin
// configurations into the host route, menu, permission, and lifecycle integration flows.

package integration

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginhost"

	"github.com/gogf/gf/v2/net/ghttp"
)

// BizCtxProvider abstracts the business context dependency for data-scope queries.
type BizCtxProvider interface {
	// GetUserId returns the user ID stored in the current request business context.
	GetUserId(ctx context.Context) int
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

// Service defines the integration service contract.
type Service interface {
	// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
	// It implements catalog.BackendConfigLoader.
	LoadPluginBackendConfig(manifest *catalog.Manifest) error
	// ListResourceRecords queries plugin-owned backend resource rows using the
	// generic plugin resource contract.
	ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
	// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
	RegisterHTTPRoutes(
		ctx context.Context,
		pluginGroup *ghttp.RouterGroup,
		middlewares pluginhost.RouteMiddlewares,
	) error
	// RegisterCrons registers callback-contributed cron jobs for source plugins.
	RegisterCrons(ctx context.Context) error
	// DispatchAfterAuth dispatches callback-style after-auth request handlers.
	// It implements runtime.AfterAuthDispatcher.
	DispatchAfterAuth(
		ctx context.Context,
		input pluginhost.AfterAuthInput,
	)
	// DispatchPluginHookEvent dispatches one named hook event to all enabled plugins.
	// It implements catalog.HookDispatcher and runtime.HookDispatcher.
	DispatchPluginHookEvent(
		ctx context.Context,
		eventName pluginhost.ExtensionPoint,
		payload map[string]interface{},
	) error
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
	// SetBizCtxProvider wires the business-context provider used by route handlers.
	SetBizCtxProvider(p BizCtxProvider)
	// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
	SetTopologyProvider(t TopologyProvider)
	// IsEnabled reports whether the plugin with the given ID is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// SyncPluginMenusAndPermissions reconciles all manifest menus and dynamic route permission
	// entries into sys_menu, then ensures the admin role has access to them.
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
	// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the
	// current governance resource index derived from the given manifest.
	// It implements catalog.ResourceRefSyncer.
	SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error
	// ListPluginResourceRefs is the exported form of listPluginResourceRefs for cross-package access.
	ListPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error)
	// BuildResourceRefDescriptors is the exported form of buildPluginResourceRefDescriptors for cross-package access.
	BuildResourceRefDescriptors(manifest *catalog.Manifest) []*catalog.ResourceRefDescriptor
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	catalogSvc catalog.Service

	bizCtxSvc BizCtxProvider

	topology TopologyProvider
}

func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{catalogSvc: catalogSvc}
}

// SetBizCtxProvider wires the business-context provider used by route handlers.
func (s *serviceImpl) SetBizCtxProvider(p BizCtxProvider) {
	s.bizCtxSvc = p
}

// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
func (s *serviceImpl) SetTopologyProvider(t TopologyProvider) {
	s.topology = t
}

// IsEnabled reports whether the plugin with the given ID is currently installed and enabled.
func (s *serviceImpl) IsEnabled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled
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
	enabledByID := make(map[string]bool, len(manifests))
	pluginIDs := make([]string, 0, len(manifests))
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

	var registries []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		WhereIn(dao.SysPlugin.Columns().PluginId, pluginIDs).
		Scan(&registries)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		enabledByID[pluginID] = registry.Installed == catalog.InstalledYes &&
			registry.Status == catalog.StatusEnabled
	}
	return enabledByID, nil
}

// buildBackgroundEnabledChecker returns a PluginEnabledChecker for use in source plugin
// route and cron registrars that need to guard runtime access.
func (s *serviceImpl) buildBackgroundEnabledChecker() pluginhost.PluginEnabledChecker {
	return func(pluginID string) bool {
		return s.IsEnabled(context.Background(), pluginID)
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
