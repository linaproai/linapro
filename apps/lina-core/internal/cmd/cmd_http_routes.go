// This file binds host API routes and source-plugin HTTP routes.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/controller/auth"
	configctrl "lina-core/internal/controller/config"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
	healthctrl "lina-core/internal/controller/health"
	i18nctrl "lina-core/internal/controller/i18n"
	jobctrl "lina-core/internal/controller/job"
	jobgroupctrl "lina-core/internal/controller/jobgroup"
	jobhandlerctrl "lina-core/internal/controller/jobhandler"
	joblogctrl "lina-core/internal/controller/joblog"
	"lina-core/internal/controller/menu"
	pluginctrl "lina-core/internal/controller/plugin"
	publicconfigctrl "lina-core/internal/controller/publicconfig"
	"lina-core/internal/controller/role"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// bindHostAPIRoutes registers the versioned host API tree, including public,
// protected, and dynamic plugin dispatch routes.
func bindHostAPIRoutes(_ context.Context, server *ghttp.Server, runtime *httpRuntime) {
	var (
		authCtrl       = auth.NewV1()
		healthCtrl     = healthctrl.NewV1(runtime.configSvc, runtime.clusterSvc)
		i18nCtrl       = i18nctrl.NewV1()
		pluginCtrl     = pluginctrl.NewV1(runtime.clusterSvc)
		publicCfgCtrl  = publicconfigctrl.NewV1()
		jobCtrl        = jobctrl.NewV1(runtime.jobMgmtSvc)
		jobGroupCtrl   = jobgroupctrl.NewV1(runtime.jobMgmtSvc)
		jobLogCtrl     = joblogctrl.NewV1(runtime.jobMgmtSvc)
		jobHandlerCtrl = jobhandlerctrl.NewV1(runtime.jobRegistry)
		middlewareSvc  = runtime.middlewareSvc
	)

	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		fileCtrl := filectrl.NewV1()
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			middlewareSvc.Response,
			middlewareSvc.CORS,
			middlewareSvc.RequestBodyLimit,
			middlewareSvc.Ctx,
		)

		bindPublicStaticAPIRoutes(
			group,
			healthCtrl.Get,
			authCtrl.Login,
			i18nCtrl.RuntimeLocales,
			i18nCtrl.RuntimeMessages,
			pluginCtrl.DynamicList,
			publicCfgCtrl.Frontend,
			fileCtrl.Access,
		)
		bindProtectedStaticAPIRoutes(
			group,
			middlewareSvc,
			authCtrl.Logout,
			i18nCtrl.ExportMessages,
			i18nCtrl.MissingMessages,
			i18nCtrl.DiagnoseMessages,
			user.NewV1(),
			dict.NewV1(),
			menu.NewV1(),
			role.NewV1(),
			usermsg.NewV1(),
			sysinfo.NewV1(),
			fileCtrl.Delete,
			fileCtrl.Detail,
			fileCtrl.Download,
			fileCtrl.InfoByIds,
			fileCtrl.List,
			fileCtrl.FileSuffixes,
			fileCtrl.Upload,
			fileCtrl.UsageScenes,
			configctrl.NewV1(),
			jobCtrl,
			jobGroupCtrl,
			jobLogCtrl,
			jobHandlerCtrl,
			pluginCtrl.List,
			pluginCtrl.Sync,
			pluginCtrl.Install,
			pluginCtrl.UploadDynamicPackage,
			pluginCtrl.Enable,
			pluginCtrl.Disable,
			pluginCtrl.Uninstall,
			pluginCtrl.UpdateTenantProvisioningPolicy,
			pluginCtrl.ResourceList,
		)
		bindDynamicPluginAPIRoutes(group, runtime.pluginSvc)
	})
}

// bindPublicStaticAPIRoutes binds endpoints that must be reachable before the
// caller has a JWT, such as login, runtime locales, and public shell config.
func bindPublicStaticAPIRoutes(parent *ghttp.RouterGroup, handlers ...any) {
	parent.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(handlers...)
	})
}

// bindProtectedStaticAPIRoutes binds host endpoints that require both
// authentication and declarative permission checks.
func bindProtectedStaticAPIRoutes(
	parent *ghttp.RouterGroup,
	middlewareSvc middleware.Service,
	handlers ...any,
) {
	parent.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewareSvc.Auth,
			middlewareSvc.Tenancy,
			middlewareSvc.Permission,
		)
		group.Bind(handlers...)
	})
}

// bindDynamicPluginAPIRoutes registers the host-owned dynamic plugin API
// dispatcher under the versioned extension namespace.
func bindDynamicPluginAPIRoutes(parent *ghttp.RouterGroup, pluginSvc pluginsvc.Service) {
	parent.Group("/extensions", func(group *ghttp.RouterGroup) {
		// Dynamic plugin routes reuse the standard RouterGroup + Middleware flow,
		// while their route matching and governance remain host-owned.
		group.Middleware(
			pluginSvc.PrepareDynamicRouteMiddleware,
			pluginSvc.AuthenticateDynamicRouteMiddleware,
		)
		pluginSvc.RegisterDynamicRouteDispatcher(group)
	})
}

// bindSourcePluginHTTPRoutes lets source plugins contribute host routes and
// starts dynamic-runtime background work that depends on route publication.
func bindSourcePluginHTTPRoutes(
	ctx context.Context,
	backgroundCtx context.Context,
	server *ghttp.Server,
	runtime *httpRuntime,
) {
	var pluginGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		pluginGroup = group
	})
	if err := runtime.pluginSvc.RegisterHTTPRoutes(
		ctx,
		server,
		pluginGroup,
		runtime.middlewareSvc.PublishedRouteMiddlewares(),
	); err != nil {
		logger.Panicf(ctx, "register plugin routes failed: %v", err)
	}
	if err := runtime.pluginSvc.PrewarmRuntimeFrontendBundles(ctx); err != nil {
		logger.Warningf(ctx, "prewarm runtime frontend bundles failed: %v", err)
	}
	runtime.pluginSvc.StartRuntimeReconciler(backgroundCtx)
}
