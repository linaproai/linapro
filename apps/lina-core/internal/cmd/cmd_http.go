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
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/controller/auth"
	configctrl "lina-core/internal/controller/config"
	"lina-core/internal/controller/dept"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
	"lina-core/internal/controller/loginlog"
	"lina-core/internal/controller/menu"
	monitorctrl "lina-core/internal/controller/monitor"
	"lina-core/internal/controller/notice"
	"lina-core/internal/controller/operlog"
	pluginctrl "lina-core/internal/controller/plugin"
	"lina-core/internal/controller/post"
	"lina-core/internal/controller/role"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/packed"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/cron"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

type HttpInput struct {
	g.Meta `name:"http" brief:"start http server"`
}
type HttpOutput struct{}

func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	var (
		s         = g.Server()
		configSvc = config.New()
	)

	var (
		clusterCfg = configSvc.GetCluster(ctx)
		clusterSvc = cluster.New(clusterCfg)
		pluginSvc  = pluginsvc.New(clusterSvc)
	)

	var (
		middlewareSvc = middleware.New()
		authCtrl      = auth.NewV1()
		pluginCtrl    = pluginctrl.NewV1(clusterSvc)
	)

	var (
		sessionCfg = configSvc.GetSession(ctx)
		monCfg     = configSvc.GetMonitor(ctx)
		cronSvc    = cron.New(sessionCfg, monCfg, middlewareSvc.SessionStore(), clusterSvc)
	)

	clusterSvc.Start(ctx)
	// Start all cron jobs (session cleanup, server monitor, etc.)
	cronSvc.Start(ctx)

	// Enhance OpenAPI documentation with config values and JWT security scheme.
	m.enhanceOpenAPIDocs(ctx, s, configSvc, pluginSvc)

	// =============================================================================================
	// Dynamic routes registering.
	// =============================================================================================

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			ghttp.MiddlewareHandlerResponse,
			middlewareSvc.CORS,
			middlewareSvc.Ctx,
		)

		// Static file serving for uploads.
		group.Group("/uploads", func(group *ghttp.RouterGroup) {
			group.ALL("/*any", func(r *ghttp.Request) {
				var (
					uploadCfg  = configSvc.GetUpload(r.Context())
					pathSuffix = r.GetRouter("any").String()
					filePath   = gfile.Join(uploadCfg.Path, pathSuffix)
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
	if err = pluginSvc.SyncSourcePlugins(ctx); err != nil {
		logger.Panicf(ctx, "sync plugin manifests failed: %v", err)
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

func (m *Main) enhanceOpenAPIDocs(
	ctx context.Context,
	server *ghttp.Server,
	configSvc config.Service,
	pluginSvc pluginsvc.Service,
) {
	// Set OpenAPI info from configuration
	oaiCfg := configSvc.GetOpenApi(ctx)
	oai := server.GetOpenApi()
	oai.Info.Title = oaiCfg.Title
	oai.Info.Description = oaiCfg.Description
	oai.Info.Version = oaiCfg.Version
	oai.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	oai.Config.CommonResponseDataField = "Data"

	// Set API server URL so documentation shows the correct backend address
	if oaiCfg.ServerUrl != "" {
		oai.Servers = &goai.Servers{
			{
				URL:         oaiCfg.ServerUrl,
				Description: oaiCfg.ServerDescription,
			},
		}
	}

	// Add JWT Bearer security scheme for API documentation
	oai.Components.SecuritySchemes = goai.SecuritySchemes{
		"BearerAuth": goai.SecuritySchemeRef{
			Value: &goai.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT Bearer Token Authentication",
				In:           "header",
				Name:         "Authorization",
			},
		},
	}
	oai.Security = &goai.SecurityRequirements{
		{"BearerAuth": {}},
	}
	if err := pluginSvc.ProjectDynamicRoutesToOpenAPI(ctx, oai.Paths); err != nil {
		logger.Warningf(ctx, "project dynamic plugin routes to OpenAPI failed: %v", err)
	}
}

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
