// This file wires the host HTTP server, static API route groups, dynamic plugin
// routes, and frontend asset serving.

package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/controller/auth"
	configctrl "lina-core/internal/controller/config"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
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
	"lina-core/internal/model"
	"lina-core/internal/packed"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/cron"
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// HttpInput defines CLI input for the HTTP startup command.
type HttpInput struct {
	g.Meta `name:"http" brief:"start http server"`
}

// HttpOutput is the CLI output placeholder for the HTTP startup command.
type HttpOutput struct{}

// httpRuntime groups long-lived services that must be shared across HTTP
// startup phases without re-constructing them in each route binding helper.
type httpRuntime struct {
	clusterSvc    cluster.Service                // clusterSvc owns primary-election lifecycle for clustered deployments.
	pluginSvc     pluginsvc.Service              // pluginSvc owns plugin lifecycle, runtime assets, routes, and hooks.
	apiDocSvc     apidoc.Service                 // apiDocSvc builds the host-managed OpenAPI document.
	jobRegistry   jobhandlersvc.Registry         // jobRegistry stores host and plugin scheduled-job handlers.
	jobMgmtSvc    jobmgmtsvc.Service             // jobMgmtSvc backs scheduled-job management controllers and cron projection.
	middlewareSvc middleware.Service             // middlewareSvc publishes host middleware chains for static and plugin routes.
	cronSvc       cron.Service                   // cronSvc starts host-level and persistent scheduled jobs.
	serverCfg     *config.ServerExtensionsConfig // serverCfg contains host extension route settings such as API docs.
	uploadPath    string                         // uploadPath is the filesystem root for API-served uploaded files.
}

// Http bootstraps the host HTTP server, static API routes, plugin routes, and
// embedded frontend asset serving.
func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	s := g.Server()
	configSvc := config.New()
	if err = configureHTTPServer(ctx, s, configSvc); err != nil {
		return nil, err
	}

	runtime, err := newHTTPRuntime(ctx, configSvc)
	if err != nil {
		return nil, err
	}
	if err = startHTTPRuntime(ctx, runtime); err != nil {
		return nil, err
	}

	bindHostAPIRoutes(ctx, s, runtime)
	bindSourcePluginHTTPRoutes(ctx, s, runtime)
	if err = bindFrontendAssetRoutes(ctx, s, runtime.pluginSvc); err != nil {
		return nil, err
	}

	bindHostedOpenAPIDocs(ctx, s, runtime.apiDocSvc, runtime.serverCfg.ApiDocPath)
	dispatchSystemStartedHook(ctx, runtime.pluginSvc)

	s.Run()
	return
}

// configureHTTPServer applies process-level server configuration that must be
// in place before any route groups are bound.
func configureHTTPServer(
	ctx context.Context,
	server *ghttp.Server,
	configSvc config.Service,
) error {
	loggerCfg := configSvc.GetLogger(ctx)
	if err := logger.BindServer(server, logger.ServerOutputConfig{
		Path:   loggerCfg.Path,
		File:   loggerCfg.File,
		Stdout: loggerCfg.Stdout,
	}); err != nil {
		return err
	}

	// Request-size limits are enforced by host middleware so multipart uploads
	// can follow the runtime-effective sys.upload.maxSize value per request
	// instead of being clipped by GoFrame's static 8MB default at server entry.
	server.SetClientMaxBodySize(0)
	return nil
}

// newHTTPRuntime constructs the shared services used by the HTTP server and
// keeps their startup dependencies in one place.
func newHTTPRuntime(ctx context.Context, configSvc config.Service) (*httpRuntime, error) {
	var (
		clusterSvc    = cluster.New(configSvc.GetCluster(ctx))
		pluginSvc     = pluginsvc.New(clusterSvc)
		apiDocSvc     = apidoc.New(configSvc, pluginSvc)
		jobRegistry   = jobhandlersvc.New()
		jobScheduler  = jobmgmtsvc.NewScheduler(clusterSvc, jobRegistry, configSvc)
		jobMgmtSvc    = jobmgmtsvc.New(configSvc, jobRegistry, jobScheduler)
		middlewareSvc = middleware.New()
	)

	// Host-owned handler definitions are registered before cron startup so the
	// persistent scheduler can project and validate code-owned jobs immediately.
	if err := jobhandlersvc.RegisterHostHandlers(jobRegistry, jobMgmtSvc); err != nil {
		return nil, err
	}

	sessionCfg, err := configSvc.GetSession(ctx)
	if err != nil {
		return nil, err
	}

	return &httpRuntime{
		clusterSvc:    clusterSvc,
		pluginSvc:     pluginSvc,
		apiDocSvc:     apiDocSvc,
		jobRegistry:   jobRegistry,
		jobMgmtSvc:    jobMgmtSvc,
		middlewareSvc: middlewareSvc,
		cronSvc: cron.New(
			sessionCfg,
			middlewareSvc.SessionStore(),
			clusterSvc,
			jobRegistry,
			jobMgmtSvc,
			jobScheduler,
		),
		serverCfg:  configSvc.GetServerExtensions(ctx),
		uploadPath: configSvc.GetUploadPath(ctx),
	}, nil
}

// startHTTPRuntime starts cluster, plugin, and cron services in the order
// required for source-plugin handlers and dynamic runtime state to be ready.
func startHTTPRuntime(ctx context.Context, runtime *httpRuntime) error {
	runtime.clusterSvc.Start(ctx)

	// Auto-enable and source-upgrade validation run before plugin routes and
	// cron jobs are registered so stale plugin state fails the process early.
	if err := runtime.pluginSvc.BootstrapAutoEnable(ctx); err != nil {
		return err
	}
	if err := runtime.pluginSvc.ValidateSourcePluginUpgradeReadiness(ctx); err != nil {
		return err
	}
	if _, err := jobhandlersvc.AttachPluginLifecycle(
		ctx,
		runtime.jobRegistry,
		runtime.pluginSvc,
	); err != nil {
		return err
	}

	// Cron startup comes after plugin lifecycle wiring so plugin-owned scheduled
	// jobs are visible when the persistent scheduler loads enabled jobs.
	runtime.cronSvc.Start(ctx)
	return nil
}

// bindHostAPIRoutes registers the versioned host API tree, including public,
// protected, upload, and dynamic plugin dispatch routes.
func bindHostAPIRoutes(_ context.Context, server *ghttp.Server, runtime *httpRuntime) {
	var (
		authCtrl       = auth.NewV1()
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
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			middlewareSvc.Response,
			middlewareSvc.CORS,
			middlewareSvc.RequestBodyLimit,
			middlewareSvc.Ctx,
		)

		bindUploadRoutes(group, runtime.uploadPath)
		bindPublicStaticAPIRoutes(
			group,
			authCtrl.Login,
			i18nCtrl.RuntimeLocales,
			i18nCtrl.RuntimeMessages,
			pluginCtrl.DynamicList,
			publicCfgCtrl.Frontend,
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
			filectrl.NewV1(),
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
			pluginCtrl.ResourceList,
		)
		bindDynamicPluginAPIRoutes(group, runtime.pluginSvc)
	})
}

// bindUploadRoutes serves files from the configured upload directory under the
// versioned API namespace without going through the SPA static fallback.
func bindUploadRoutes(parent *ghttp.RouterGroup, uploadPath string) {
	parent.Group("/uploads", func(group *ghttp.RouterGroup) {
		group.ALL("/*any", func(r *ghttp.Request) {
			var (
				pathSuffix = r.GetRouter("any").String()
				filePath   = gfile.Join(uploadPath, pathSuffix)
			)
			if !gfile.Exists(filePath) {
				r.Response.WriteStatus(http.StatusNotFound)
				r.ExitAll()
				return
			}
			r.Response.ServeFile(filePath)
			r.ExitAll()
		})
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
func bindSourcePluginHTTPRoutes(ctx context.Context, server *ghttp.Server, runtime *httpRuntime) {
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
	runtime.pluginSvc.StartRuntimeReconciler(ctx)
}

// bindFrontendAssetRoutes registers the final catch-all static route for host
// frontend assets and SPA fallback after all API and plugin routes are bound.
func bindFrontendAssetRoutes(
	ctx context.Context,
	server *ghttp.Server,
	pluginSvc pluginsvc.Service,
) error {
	subFS, err := fs.Sub(packed.Files, "public")
	if err != nil {
		logger.Panicf(ctx, "load embedded frontend assets failed: %v", err)
		return err
	}
	fileServer := http.FileServer(http.FS(subFS))
	server.BindHandler("/*", newFrontendAssetHandler(subFS, fileServer, pluginSvc))
	return nil
}

// newFrontendAssetHandler creates the SPA/static-file handler used as the last
// route in the server so API and plugin routes get first chance to match.
func newFrontendAssetHandler(
	subFS fs.FS,
	fileServer http.Handler,
	pluginSvc pluginsvc.Service,
) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if serveRuntimePluginAsset(r, pluginSvc, path) {
			return
		}
		if serveEmbeddedFrontendAsset(r, subFS, fileServer, path) {
			return
		}
		serveSPAFallback(r, fileServer)
	}
}

// serveRuntimePluginAsset serves versioned dynamic plugin frontend assets when
// the request path belongs to the public plugin-asset namespace.
func serveRuntimePluginAsset(
	r *ghttp.Request,
	pluginSvc pluginsvc.Service,
	path string,
) bool {
	// Runtime plugin assets must be checked before the host falls back to the
	// embedded frontend bundle. They share the same public static entrypoint,
	// but plugin assets are governed by plugin ID, version, and enabled state.
	// If the host served the generic SPA assets first, a valid plugin asset URL
	// could be swallowed by the host fallback and bypass the runtime-specific
	// access rules that ResolveRuntimeFrontendAsset enforces.
	pluginID, version, assetPath, ok := parsePluginAssetRequestPath(path)
	if !ok {
		return false
	}
	out, resolveErr := pluginSvc.ResolveRuntimeFrontendAsset(
		r.Context(),
		pluginID,
		version,
		assetPath,
	)
	if resolveErr != nil {
		r.Response.WriteStatus(http.StatusNotFound)
		r.ExitAll()
		return true
	}
	r.Response.Header().Set("Content-Type", out.ContentType)
	r.Response.Write(out.Content)
	r.ExitAll()
	return true
}

// serveEmbeddedFrontendAsset serves one concrete embedded frontend file when
// it exists and lets callers fall through to the SPA fallback otherwise.
func serveEmbeddedFrontendAsset(
	r *ghttp.Request,
	subFS fs.FS,
	fileServer http.Handler,
	path string,
) bool {
	f, err := subFS.Open(path)
	if err != nil {
		return false
	}
	if closeErr := f.Close(); closeErr != nil {
		logger.Warningf(
			r.Context(),
			"close embedded frontend asset failed path=%s err=%v",
			path,
			closeErr,
		)
	}
	fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
	r.ExitAll()
	return true
}

// serveSPAFallback serves index.html for unmatched frontend routes so browser
// refreshes on client-side routes are handled by the Vue application.
func serveSPAFallback(r *ghttp.Request, fileServer http.Handler) {
	r.Request.URL.Path = "/index.html"
	fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
	r.ExitAll()
}

// dispatchSystemStartedHook notifies enabled plugins after all host routes and
// frontend asset handlers are available.
func dispatchSystemStartedHook(ctx context.Context, pluginSvc pluginsvc.Service) {
	if err := pluginSvc.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointSystemStarted,
		map[string]any{},
	); err != nil {
		logger.Warningf(
			ctx,
			"dispatch plugin backend extension point failed point=%s err=%v",
			pluginhost.ExtensionPointSystemStarted,
			err,
		)
	}
}

// bindHostedOpenAPIDocs disables the GoFrame built-in OpenAPI and Swagger
// endpoints, then binds the host-managed OpenAPI JSON handler at the configured path.
func bindHostedOpenAPIDocs(
	_ context.Context,
	server *ghttp.Server,
	apiDocSvc apidoc.Service,
	apiDocPath string,
) {
	if server == nil {
		return
	}

	server.SetOpenApiPath("")
	server.SetSwaggerPath("")

	apiDocPath = strings.TrimSpace(apiDocPath)
	if apiDocPath == "" || apiDocSvc == nil {
		return
	}

	apiDocBizCtxSvc := bizctx.New()
	apiDocI18nSvc := i18nsvc.New()
	server.BindHandler(apiDocPath, func(r *ghttp.Request) {
		apiDocBizCtxSvc.Init(r, &model.Context{})
		locale := apiDocI18nSvc.ResolveRequestLocale(r)
		r.SetCtx(gi18n.WithLanguage(r.Context(), locale))
		apiDocBizCtxSvc.SetLocale(r.Context(), locale)
		r.Response.Header().Set("Content-Language", locale)

		document, err := apiDocSvc.Build(r.Context(), server)
		if err != nil {
			logger.Warningf(r.Context(), "build hosted OpenAPI document failed: %v", err)
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.Write("build hosted OpenAPI document failed")
			r.ExitAll()
			return
		}
		r.Response.WriteJson(document)
		r.ExitAll()
	})
}

// parsePluginAssetRequestPath splits one public `/plugin-assets/...` request
// path into plugin identity, version, and relative asset path parts.
func parsePluginAssetRequestPath(path string) (
	pluginID string,
	version string,
	assetPath string,
	ok bool,
) {
	normalizedPath := strings.Trim(strings.TrimSpace(path), "/")
	if normalizedPath == "" {
		return "", "", "", false
	}

	pathParts := strings.Split(normalizedPath, "/")
	if len(pathParts) < 3 || pathParts[0] != "plugin-assets" {
		return "", "", "", false
	}
	if strings.TrimSpace(pathParts[1]) == "" || strings.TrimSpace(pathParts[2]) == "" {
		return "", "", "", false
	}

	pluginID = pathParts[1]
	version = pathParts[2]
	if len(pathParts) == 3 {
		return pluginID, version, "", true
	}
	return pluginID, version, strings.Join(pathParts[3:], "/"), true
}
