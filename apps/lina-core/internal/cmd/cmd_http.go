// This file wires the host HTTP server, static API route groups, dynamic plugin
// routes, and frontend asset serving.

package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/controller/auth"
	configctrl "lina-core/internal/controller/config"
	"lina-core/internal/controller/dept"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
	jobctrl "lina-core/internal/controller/job"
	jobgroupctrl "lina-core/internal/controller/jobgroup"
	jobhandlerctrl "lina-core/internal/controller/jobhandler"
	joblogctrl "lina-core/internal/controller/joblog"
	"lina-core/internal/controller/loginlog"
	"lina-core/internal/controller/menu"
	monitorctrl "lina-core/internal/controller/monitor"
	"lina-core/internal/controller/notice"
	"lina-core/internal/controller/operlog"
	pluginctrl "lina-core/internal/controller/plugin"
	"lina-core/internal/controller/post"
	publicconfigctrl "lina-core/internal/controller/publicconfig"
	"lina-core/internal/controller/role"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/packed"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/cron"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	schedulerpkg "lina-core/internal/service/jobmgmt/scheduler"
	"lina-core/internal/service/jobmgmt/shellexec"
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

// Http bootstraps the host HTTP server, static API routes, plugin routes, and
// embedded frontend asset serving.
func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	var (
		s         = g.Server()
		configSvc = config.New()
		loggerCfg = configSvc.GetLogger(ctx)
	)

	if err = logger.BindServer(s, logger.ServerOutputConfig{
		Path:   loggerCfg.Path,
		File:   loggerCfg.File,
		Stdout: loggerCfg.Stdout,
	}); err != nil {
		return nil, err
	}
	// The host applies request-size limits in middleware so multipart uploads
	// can follow the runtime-effective sys.upload.maxSize value per request
	// instead of being clipped by GoFrame's static 8MB default at server entry.
	s.SetClientMaxBodySize(0)

	var (
		clusterCfg = configSvc.GetCluster(ctx)
		clusterSvc = cluster.New(clusterCfg)
		pluginSvc  = pluginsvc.New(clusterSvc)
		apiDocSvc  = apidoc.New(configSvc, pluginSvc)
	)

	var (
		jobRegistry    = jobhandlersvc.New()
		shellExecSvc   = shellexec.New(configSvc)
		jobScheduler   = schedulerpkg.New(clusterSvc, jobRegistry, shellExecSvc)
		jobMgmtSvc     = jobmgmtsvc.New(configSvc, jobRegistry, jobScheduler)
		middlewareSvc  = middleware.New()
		authCtrl       = auth.NewV1()
		pluginCtrl     = pluginctrl.NewV1(clusterSvc)
		publicCfgCtrl  = publicconfigctrl.NewV1()
		jobCtrl        = jobctrl.NewV1(jobMgmtSvc)
		jobGroupCtrl   = jobgroupctrl.NewV1(jobMgmtSvc)
		jobLogCtrl     = joblogctrl.NewV1(jobMgmtSvc)
		jobHandlerCtrl = jobhandlerctrl.NewV1(jobRegistry)
	)
	if err = jobhandlersvc.RegisterHostHandlers(jobRegistry, jobMgmtSvc); err != nil {
		return nil, err
	}
	if err = pluginSvc.SyncSourcePlugins(ctx); err != nil {
		return nil, err
	}
	if _, err = jobhandlersvc.AttachPluginLifecycle(ctx, jobRegistry, pluginSvc); err != nil {
		return nil, err
	}

	var (
		sessionCfg = configSvc.GetSession(ctx)
		monCfg     = configSvc.GetMonitor(ctx)
		serverCfg  = configSvc.GetServerExtensions(ctx)
		cronSvc    = cron.New(sessionCfg, monCfg, middlewareSvc.SessionStore(), clusterSvc, jobScheduler)
		uploadPath = configSvc.GetUploadPath(ctx)
	)
	clusterSvc.Start(ctx)

	// Start all cron jobs (session cleanup, server monitor, etc.)
	cronSvc.Start(ctx)

	m.bindHostedOpenAPIDocs(ctx, s, apiDocSvc, serverCfg.ApiDocPath)

	// =============================================================================================
	// Dynamic routes registering.
	// =============================================================================================

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			ghttp.MiddlewareHandlerResponse,
			middlewareSvc.CORS,
			middlewareSvc.RequestBodyLimit,
			middlewareSvc.Ctx,
		)

		// Static file serving for uploads.
		group.Group("/uploads", func(group *ghttp.RouterGroup) {
			group.ALL("/*any", func(r *ghttp.Request) {
				var (
					pathSuffix = r.GetRouter("any").String()
					filePath   = gfile.Join(uploadPath, pathSuffix)
				)
				if !gfile.Exists(filePath) {
					r.Response.WriteStatus(404)
					r.ExitAll()
					return
				}
				r.Response.ServeFile(filePath)
				r.ExitAll()
			})
		})

		// Public routes (no auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(
				authCtrl.Login,
				pluginCtrl.DynamicList,
				publicCfgCtrl.Frontend,
			)
		})

		// Protected routes (auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middlewareSvc.Auth,
				middlewareSvc.OperLog,
				middlewareSvc.Permission,
			)
			group.Bind(
				authCtrl.Logout,
				user.NewV1(),
				dict.NewV1(),
				dept.NewV1(),
				post.NewV1(),
				menu.NewV1(),
				role.NewV1(),
				notice.NewV1(),
				usermsg.NewV1(),
				loginlog.NewV1(),
				operlog.NewV1(),
				sysinfo.NewV1(),
				filectrl.NewV1(),
				monitorctrl.NewV1(),
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
		})

		// Dynamic plugin routes reuse the standard RouterGroup + Middleware flow,
		// while their route matching and governance remain host-owned.
		group.Group("/extensions", func(group *ghttp.RouterGroup) {
			group.Middleware(
				pluginSvc.PrepareDynamicRouteMiddleware,
				pluginSvc.AuthenticateDynamicRouteMiddleware,
				middlewareSvc.OperLog,
			)
			pluginSvc.RegisterDynamicRouteDispatcher(group)
		})
	})

	var pluginGroup *ghttp.RouterGroup
	s.Group("/", func(group *ghttp.RouterGroup) {
		pluginGroup = group
	})
	if err = pluginSvc.RegisterHTTPRoutes(ctx, pluginGroup, middlewareSvc.PublishedRouteMiddlewares()); err != nil {
		logger.Panicf(ctx, "register plugin routes failed: %v", err)
	}
	if err = pluginSvc.PrewarmRuntimeFrontendBundles(ctx); err != nil {
		logger.Warningf(ctx, "prewarm runtime frontend bundles failed: %v", err)
	}
	pluginSvc.StartRuntimeReconciler(ctx)

	// =============================================================================================
	// Static service for frontend assets.
	// =============================================================================================

	// Serve embedded frontend static files
	subFS, err := fs.Sub(packed.Files, "public")
	if err != nil {
		logger.Panicf(ctx, "load embedded frontend assets failed: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))
	s.BindHandler("/*", func(r *ghttp.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		// Runtime plugin assets must be checked before the host falls back to the
		// embedded frontend bundle. They share the same public static entrypoint,
		// but plugin assets are governed by plugin ID, version, and enabled state.
		// If the host served the generic SPA assets first, a valid plugin asset URL
		// could be swallowed by the host fallback and bypass the runtime-specific
		// access rules that ResolveRuntimeFrontendAsset enforces.
		if pluginID, version, assetPath, ok := parsePluginAssetRequestPath(path); ok {
			out, resolveErr := pluginSvc.ResolveRuntimeFrontendAsset(
				r.Context(),
				pluginID,
				version,
				assetPath,
			)
			if resolveErr != nil {
				r.Response.WriteStatus(http.StatusNotFound)
				r.ExitAll()
				return
			}
			r.Response.Header().Set("Content-Type", out.ContentType)
			r.Response.Write(out.Content)
			r.ExitAll()
			return
		}
		f, err := subFS.Open(path)
		if err == nil {
			if closeErr := f.Close(); closeErr != nil {
				logger.Warningf(r.Context(), "close embedded frontend asset failed path=%s err=%v", path, closeErr)
			}
			fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
			r.ExitAll()
			return
		}
		// SPA fallback: serve index.html for unmatched paths
		r.Request.URL.Path = "/index.html"
		fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
		r.ExitAll()
	})

	if err = pluginSvc.DispatchHookEvent(
		ctx, pluginhost.ExtensionPointSystemStarted, map[string]any{},
	); err != nil {
		logger.Warningf(
			ctx,
			"dispatch plugin backend extension point failed point=%s err=%v",
			pluginhost.ExtensionPointSystemStarted,
			err,
		)
	}

	s.Run()
	return
}

// bindHostedOpenAPIDocs disables the GoFrame built-in OpenAPI and Swagger
// endpoints, then binds the host-managed OpenAPI JSON handler at the configured path.
func (m *Main) bindHostedOpenAPIDocs(
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

	server.BindHandler(apiDocPath, func(r *ghttp.Request) {
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
