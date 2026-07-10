// This file exposes host integration and hook dispatch methods on the root
// plugin facade.

package plugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
func (s *serviceImpl) RegisterHTTPRoutes(
	ctx context.Context,
	server *ghttp.Server,
	pluginGroup *ghttp.RouterGroup,
	middlewares pluginhost.RouteMiddlewares,
) error {
	return s.integrationSvc.RegisterHTTPRoutes(ctx, server, pluginGroup, middlewares)
}

// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
func (s *serviceImpl) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	return s.integrationSvc.ListSourceRouteBindings()
}

// RegisterJobs registers callback-contributed scheduled jobs for source plugins.
func (s *serviceImpl) RegisterJobs(ctx context.Context) error {
	return s.integrationSvc.RegisterJobs(ctx)
}

// OrgProviderEnv returns typed, plugin-scoped organization-provider construction inputs.
func (s *serviceImpl) OrgProviderEnv(_ context.Context, pluginID string) orgspi.ProviderEnv {
	env := orgspi.ProviderEnv{PluginID: pluginID}
	if s == nil || s.capabilities == nil {
		return env
	}
	services := capabilityowner.ServicesForPlugin(s.capabilities, pluginID)
	if services == nil {
		return env
	}
	env.Tenant = services.Tenant()
	env.Users = services.Users()
	return env
}

// RegisterSourcePluginProviderFactories registers compile-time source-plugin
// provider declarations into the startup-owned shared provider managers.
func (s *serviceImpl) RegisterSourcePluginProviderFactories(
	tenantManager *tenantspi.Manager,
	orgManager *orgspi.Manager,
) error {
	for _, definition := range pluginhost.ListSourcePlugins() {
		if definition == nil {
			continue
		}
		pluginID := definition.ID()
		if factory := definition.GetTenantProviderFactory(); factory != nil {
			if tenantManager == nil {
				return gerror.New("plugin service requires tenant provider manager")
			}
			if err := tenantManager.RegisterFactory(pluginID, factory); err != nil {
				return err
			}
		}
		if factory := definition.GetOrgProviderFactory(); factory != nil {
			if orgManager == nil {
				return gerror.New("plugin service requires organization provider manager")
			}
			if err := orgManager.RegisterFactory(pluginID, factory); err != nil {
				return err
			}
		}
	}
	return nil
}

// TenantProviderEnv returns typed, plugin-scoped tenant-provider construction inputs.
func (s *serviceImpl) TenantProviderEnv(_ context.Context, pluginID string) tenantspi.ProviderEnv {
	env := tenantspi.ProviderEnv{PluginID: pluginID}
	if s == nil || s.capabilities == nil {
		return env
	}
	services := capabilityowner.ServicesForPlugin(s.capabilities, pluginID)
	if services == nil {
		return env
	}
	env.BizCtx = services.BizCtx()
	env.PluginLifecycle = s.pluginLifecycleService
	if plugins := services.Plugins(); plugins != nil {
		if lifecycle := plugins.Lifecycle(); lifecycle != nil {
			env.PluginLifecycle = lifecycle
		}
	}
	env.Tenant = services.Tenant()
	env.Users = services.Users()
	env.Plugins = services.Plugins()
	return env
}

// ListManagedJobs returns plugin-owned job declarations or executable handlers
// according to the supplied query. Executable callers must opt in explicitly so
// management projections do not accidentally publish handler functions.
func (s *serviceImpl) ListManagedJobs(ctx context.Context, query ManagedJobQuery) ([]ManagedJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	pluginID := strings.TrimSpace(query.PluginID)
	var (
		items []ManagedJob
		err   error
	)
	switch {
	case query.ExecutableOnly && pluginID != "":
		items, err = s.integrationSvc.ListExecutableJobsByPlugin(ctx, pluginID)
	case query.ExecutableOnly:
		items, err = s.integrationSvc.ListExecutableJobs(ctx)
	case query.InstalledOnly:
		items, err = s.integrationSvc.ListInstalledJobDeclarations(ctx)
		if err == nil && pluginID != "" {
			items = filterManagedJobsByPlugin(items, pluginID)
		}
	case pluginID != "":
		items, err = s.integrationSvc.ListJobDeclarationsByPlugin(ctx, pluginID)
	default:
		return nil, gerror.New("plugin managed job query requires executable, installed, or plugin id scope")
	}
	if err != nil {
		return nil, err
	}
	if !query.IncludeHandlers {
		clearManagedJobHandlers(items)
	}
	return items, nil
}

// DispatchHookEvent dispatches one named hook event to all enabled plugins.
func (s *serviceImpl) DispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return err
	}
	readCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return err
	}
	return s.integrationSvc.DispatchPluginHookEvent(readCtx, event, values)
}

// FilterMenus filters disabled plugin menus from the given menu list.
func (s *serviceImpl) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "filter_menus")
	return s.integrationSvc.FilterMenus(integration.WithAuthoritativeEnablement(ctx), menus)
}

// FilterPermissionMenus filters permission menus based on plugin enablement.
func (s *serviceImpl) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "filter_permission_menus")
	return s.integrationSvc.FilterPermissionMenus(integration.WithAuthoritativeEnablement(ctx), menus)
}

// ResolveResourcePermission resolves the plugin-scoped permission required by one plugin resource.
func (s *serviceImpl) ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return "", err
	}
	return s.integrationSvc.ResolveResourcePermission(ctx, pluginID, resourceID)
}

// ListResourceRecords queries plugin-owned backend resource rows.
func (s *serviceImpl) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListResourceRecords(ctx, in)
}

// filterManagedJobsByPlugin keeps only jobs owned by pluginID in an already
// bounded result set.
func filterManagedJobsByPlugin(items []ManagedJob, pluginID string) []ManagedJob {
	if len(items) == 0 || strings.TrimSpace(pluginID) == "" {
		return items
	}
	filtered := make([]ManagedJob, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.PluginID) == pluginID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// clearManagedJobHandlers removes executable callbacks from management
// projections so declaration callers cannot accidentally publish handlers.
func clearManagedJobHandlers(items []ManagedJob) {
	for index := range items {
		items[index].Handler = nil
	}
}
