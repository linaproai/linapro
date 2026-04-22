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

// ListManagedCronJobs returns plugin-owned cron definitions for scheduled-job projection.
func (s *serviceImpl) ListManagedCronJobs(ctx context.Context) ([]ManagedCronJob, error) {
	return s.integrationSvc.ListManagedCronJobs(ctx)
}

// ListManagedCronJobsByPlugin returns plugin-owned cron definitions for one plugin.
func (s *serviceImpl) ListManagedCronJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error) {
	return s.integrationSvc.ListManagedCronJobsByPlugin(ctx, pluginID)
}

// DispatchHookEvent dispatches one named hook event to all enabled plugins.
func (s *serviceImpl) DispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	return s.integrationSvc.DispatchPluginHookEvent(ctx, event, values)
}

// FilterMenus filters disabled plugin menus from the given menu list.
func (s *serviceImpl) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return s.integrationSvc.FilterMenus(ctx, menus)
}

// FilterPermissionMenus filters permission menus based on plugin enablement.
func (s *serviceImpl) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return s.integrationSvc.FilterPermissionMenus(ctx, menus)
}

// ResolveResourcePermission resolves the plugin-scoped permission required by one plugin resource.
func (s *serviceImpl) ResolveResourcePermission(ctx context.Context, pluginID string, resourceID string) (string, error) {
	return s.integrationSvc.ResolveResourcePermission(ctx, pluginID, resourceID)
}

// ListResourceRecords queries plugin-owned backend resource rows.
func (s *serviceImpl) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	return s.integrationSvc.ListResourceRecords(ctx, in)
}
