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
		authCtrl       = auth.NewV1(runtime.authSvc, runtime.bizCtxSvc)
		configCtrl     = configctrl.NewV1(runtime.sysConfigSvc)
		dictCtrl       = dict.NewV1(runtime.dictSvc)
		fileCtrl       = filectrl.NewV1(runtime.fileSvc)
		healthCtrl     = healthctrl.NewV1(runtime.configSvc, runtime.clusterSvc)
		i18nCtrl       = i18nctrl.NewV1(runtime.i18nSvc)
		pluginCtrl     = pluginctrl.NewV1(runtime.pluginSvc, runtime.bizCtxSvc, runtime.configSvc, runtime.i18nSvc, runtime.roleSvc)
		publicCfgCtrl  = publicconfigctrl.NewV1(runtime.configSvc, runtime.i18nSvc)
		menuCtrl       = menu.NewV1(runtime.menuSvc, runtime.roleSvc, runtime.bizCtxSvc)
		roleCtrl       = role.NewV1(runtime.roleSvc)
		userCtrl       = user.NewV1(runtime.userSvc, runtime.roleSvc, runtime.menuSvc, runtime.orgCapSvc, runtime.i18nSvc)
		userMsgCtrl    = usermsg.NewV1(runtime.userMsgSvc)
		jobCtrl        = jobctrl.NewV1(runtime.jobMgmtSvc)
		jobGroupCtrl   = jobgroupctrl.NewV1(runtime.jobMgmtSvc)
		jobLogCtrl     = joblogctrl.NewV1(runtime.bizCtxSvc, runtime.jobMgmtSvc, runtime.roleSvc)
		jobHandlerCtrl = jobhandlerctrl.NewV1(runtime.jobRegistry, runtime.i18nSvc)
		middlewareSvc  = runtime.middlewareSvc
		sysInfoCtrl    = sysinfo.NewV1(runtime.sysInfoSvc, runtime.i18nSvc)
	)

	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
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
			authCtrl.RefreshToken,
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
			userCtrl,
			dictCtrl,
			menuCtrl,
			roleCtrl,
			userMsgCtrl,
			sysInfoCtrl,
			fileCtrl.Delete,
			fileCtrl.Detail,
			fileCtrl.Download,
			fileCtrl.InfoByIds,
			fileCtrl.List,
			fileCtrl.FileSuffixes,
			fileCtrl.Upload,
			fileCtrl.UsageScenes,
			configCtrl,
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
