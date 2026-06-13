// This file exposes host integration and hook dispatch methods on the root
// plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/pkg/plugin/capability"
	aitextsvc "lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// English fallback messages published with host authentication lifecycle events.
const (
	// AuthEventMessageLoginSuccessful is the English fallback for successful login messages.
	AuthEventMessageLoginSuccessful = "Login successful"
	// AuthEventMessageLoginFailed is the English fallback for generic failed login messages.
	AuthEventMessageLoginFailed = "Login failed"
	// AuthEventMessageLogoutSuccessful is the English fallback for successful logout messages.
	AuthEventMessageLogoutSuccessful = "Logout successful"
	// AuthEventMessageInvalidCredentials is the English fallback for invalid credential messages.
	AuthEventMessageInvalidCredentials = "Invalid username or password"
	// AuthEventMessageUserDisabled is the English fallback for disabled account messages.
	AuthEventMessageUserDisabled = "User account is disabled"
	// AuthEventMessageIPBlacklisted is the English fallback for blocked login IP messages.
	AuthEventMessageIPBlacklisted = "Login IP is blacklisted"
)

// sourceServicesProvider stores the startup-owned capability directory and
// returns plugin-scoped source-plugin service views for integration callbacks.
type sourceServicesProvider struct {
	capabilities capability.Services
}

// SourceServicesForPlugin returns a plugin-scoped source-plugin service view.
func (p *sourceServicesProvider) SourceServicesForPlugin(pluginID string) pluginhost.Services {
	if p == nil {
		return nil
	}
	capabilities := p.capabilities
	if capabilities == nil {
		return nil
	}
	services := capability.ServicesForPlugin(capabilities, pluginID)
	if sourceServices, ok := services.(pluginhost.Services); ok {
		return sourceServices
	}
	return nil
}

// StorageCleanupServices returns the startup-owned shared capability directory
// for runtime dynamic-plugin storage cleanup.
func (p *sourceServicesProvider) StorageCleanupServices() capability.Services {
	if p == nil {
		return nil
	}
	return p.capabilities
}

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

// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginSucceeded,
		input,
		pluginhost.AuthHookReasonLoginSuccessful,
		AuthEventMessageLoginSuccessful,
	)
}

// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginFailed(ctx context.Context, input pluginhost.AuthHookPayloadInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginFailed,
		input,
		pluginhost.AuthHookReasonLoginFailed,
		AuthEventMessageLoginFailed,
	)
}

// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLogoutSucceeded(ctx context.Context, input pluginhost.AuthHookPayloadInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		input,
		pluginhost.AuthHookReasonLogoutSuccessful,
		AuthEventMessageLogoutSuccessful,
	)
}

// AITextProviderEnv returns typed, plugin-scoped text AI provider construction inputs.
func (s *serviceImpl) AITextProviderEnv(pluginID string) aitextsvc.ProviderEnv {
	env := aitextsvc.ProviderEnv{PluginID: pluginID}
	if s == nil || s.capabilities == nil {
		return env
	}
	services := capability.ServicesForPlugin(s.capabilities, pluginID)
	if services == nil {
		return env
	}
	env.BizCtx = services.BizCtx()
	env.Cache = services.Cache()
	return env
}

// OrgProviderEnv returns typed, plugin-scoped organization-provider construction inputs.
func (s *serviceImpl) OrgProviderEnv(pluginID string) orgspi.ProviderEnv {
	env := orgspi.ProviderEnv{PluginID: pluginID}
	if s == nil || s.capabilities == nil {
		return env
	}
	services := capability.ServicesForPlugin(s.capabilities, pluginID)
	if services == nil {
		return env
	}
	sourceServices, ok := services.(interface {
		TenantFilter() tenantspi.PluginTableFilterService
	})
	if !ok {
		return env
	}
	env.TenantFilter = sourceServices.TenantFilter()
	env.Users = services.Users()
	return env
}

// RegisterSourcePluginProviderFactories registers compile-time source-plugin
// provider declarations into the startup-owned shared provider managers.
func (s *serviceImpl) RegisterSourcePluginProviderFactories(
	tenantManager *tenantspi.Manager,
	orgManager *orgspi.Manager,
	aiTextManager *aitextsvc.Manager,
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
		if factory := definition.GetAITextProviderFactory(); factory != nil {
			if aiTextManager == nil {
				return gerror.New("plugin service requires text AI provider manager")
			}
			if err := aiTextManager.RegisterFactory(pluginID, factory); err != nil {
				return err
			}
		}
	}
	return nil
}

// TenantProviderEnv returns typed, plugin-scoped tenant-provider construction inputs.
func (s *serviceImpl) TenantProviderEnv(pluginID string) tenantspi.ProviderEnv {
	env := tenantspi.ProviderEnv{PluginID: pluginID}
	if s == nil || s.capabilities == nil {
		return env
	}
	services := capability.ServicesForPlugin(s.capabilities, pluginID)
	if services == nil {
		return env
	}
	env.BizCtx = services.BizCtx()
	if plugins := services.Plugins(); plugins != nil {
		env.PluginLifecycle = plugins.Lifecycle()
	}
	env.Users = services.Users()
	env.Plugins = services.Plugins()
	if sourceServices, ok := services.(interface {
		Admin() capability.AdminServices
	}); ok {
		if admin := sourceServices.Admin(); admin != nil {
			env.PluginAdmin = admin.Plugins()
		}
	}
	return env
}

// ListExecutableJobs returns plugin-owned job definitions whose handlers
// are safe to publish for execution. Dynamic plugins must be installed, enabled
// for the current business-entry context, and free of runtime-upgrade blocking
// states; declarations from disabled or preview-only dynamic plugins are
// intentionally excluded. Use this method for executable handler publication,
// not for authorization previews or scheduled-job table projection.
func (s *serviceImpl) ListExecutableJobs(ctx context.Context) ([]ManagedJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListExecutableJobs(ctx)
}

// ListExecutableJobsByPlugin returns executable plugin-owned scheduled job
// definitions for one plugin. The method applies the same runtime cache
// freshness, install, enablement, and runtime-state checks as
// ListExecutableJobs while narrowing discovery to pluginID. Job-handler
// lifecycle synchronization uses this path when an enabled plugin publishes its
// concrete handler references.
func (s *serviceImpl) ListExecutableJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListExecutableJobsByPlugin(ctx, pluginID)
}

// ListJobDeclarationsByPlugin returns plugin-owned job declaration metadata
// for management review without requiring the plugin business entry to be
// enabled. This path is used by plugin list and authorization-preview screens,
// including before a dynamic plugin is installed. Returned items describe what
// the plugin declares; callers must not treat them as proof that handlers can be
// executed.
func (s *serviceImpl) ListJobDeclarationsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListJobDeclarationsByPlugin(ctx, pluginID)
}

// ListInstalledJobDeclarations returns declared job metadata for installed
// plugins without requiring their business entries to be enabled. Scheduled-job
// projection uses this path so installed-but-disabled plugins can keep visible
// task rows, while uninstalled authorization-preview declarations stay out of
// the persistent task table.
func (s *serviceImpl) ListInstalledJobDeclarations(ctx context.Context) ([]ManagedJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListInstalledJobDeclarations(ctx)
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

// dispatchAuthHookEvent normalizes common auth payload reason and message
// defaults before forwarding the event to the shared integration hook dispatcher.
func (s *serviceImpl) dispatchAuthHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	input pluginhost.AuthHookPayloadInput,
	defaultReason string,
	defaultMessage string,
) error {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return err
	}
	if input.Reason == "" {
		input.Reason = defaultReason
	}
	if input.Message == "" {
		input.Message = defaultMessage
	}
	return s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		event,
		pluginhost.BuildAuthHookPayloadValues(pluginhost.AuthHookPayloadInput{
			UserName:   input.UserName,
			Status:     input.Status,
			IP:         input.IP,
			ClientType: input.ClientType,
			Browser:    input.Browser,
			OS:         input.OS,
			Message:    input.Message,
			Reason:     input.Reason,
		}),
	)
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
