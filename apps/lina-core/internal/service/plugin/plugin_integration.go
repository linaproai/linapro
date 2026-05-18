// This file exposes host integration methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
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

// RegisterCrons registers callback-contributed cron jobs for source plugins.
func (s *serviceImpl) RegisterCrons(ctx context.Context) error {
	return s.integrationSvc.RegisterCrons(ctx)
}

// SetHostServices wires the host-published service directory used by source plugins.
func (s *serviceImpl) SetHostServices(services pluginhost.HostServices) {
	if s == nil || s.integrationSvc == nil {
		return
	}
	s.integrationSvc.SetHostServices(services)
}

// ListExecutableCronJobs returns plugin-owned cron definitions whose handlers
// are safe to publish for execution. Dynamic plugins must be installed, enabled
// for the current business-entry context, and free of runtime-upgrade blocking
// states; declarations from disabled or preview-only dynamic plugins are
// intentionally excluded. Use this method for executable handler publication,
// not for authorization previews or scheduled-job table projection.
func (s *serviceImpl) ListExecutableCronJobs(ctx context.Context) ([]ManagedCronJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListExecutableCronJobs(ctx)
}

// ListExecutableCronJobsByPlugin returns executable plugin-owned cron
// definitions for one plugin. The method applies the same runtime cache
// freshness, install, enablement, and runtime-state checks as
// ListExecutableCronJobs while narrowing discovery to pluginID. Job-handler
// lifecycle synchronization uses this path when an enabled plugin publishes its
// concrete handler references.
func (s *serviceImpl) ListExecutableCronJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListExecutableCronJobsByPlugin(ctx, pluginID)
}

// ListCronDeclarationsByPlugin returns plugin-owned cron declaration metadata
// for management review without requiring the plugin business entry to be
// enabled. This path is used by plugin list and authorization-preview screens,
// including before a dynamic plugin is installed. Returned items describe what
// the plugin declares; callers must not treat them as proof that handlers can be
// executed.
func (s *serviceImpl) ListCronDeclarationsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListCronDeclarationsByPlugin(ctx, pluginID)
}

// ListInstalledCronDeclarations returns declared cron metadata for installed
// plugins without requiring their business entries to be enabled. Scheduled-job
// projection uses this path so installed-but-disabled plugins can keep visible
// task rows, while uninstalled authorization-preview declarations stay out of
// the persistent task table.
func (s *serviceImpl) ListInstalledCronDeclarations(ctx context.Context) ([]ManagedCronJob, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.integrationSvc.ListInstalledCronDeclarations(ctx)
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
	readCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return err
	}
	return s.integrationSvc.DispatchPluginHookEvent(readCtx, event, values)
}

// FilterMenus filters disabled plugin menus from the given menu list.
func (s *serviceImpl) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "filter_menus")
	return s.integrationSvc.FilterMenus(ctx, menus)
}

// FilterPermissionMenus filters permission menus based on plugin enablement.
func (s *serviceImpl) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	s.ensureRuntimeCacheFreshBestEffort(ctx, "filter_permission_menus")
	return s.integrationSvc.FilterPermissionMenus(ctx, menus)
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
