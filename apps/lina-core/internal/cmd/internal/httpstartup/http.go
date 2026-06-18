// This file exposes the narrow HTTP command runtime entrypoint used by the
// command root while keeping startup wiring details inside this package.

package httpstartup

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

// Run bootstraps the host HTTP server, static API routes, plugin routes, and
// embedded frontend asset serving.
func Run(ctx context.Context) error {
	server := g.Server()
	configSvc := config.New()
	if err := configureHTTPServer(ctx, server, configSvc); err != nil {
		return err
	}

	runtime, err := newHTTPRuntime(ctx, configSvc)
	if err != nil {
		return err
	}
	startupCtx, startupCollector, err := newHTTPStartupContext(ctx, runtime)
	if err != nil {
		return err
	}
	if err = startHTTPRuntimeBeforeSourceRoutes(startupCtx, runtime); err != nil {
		return err
	}

	bindHostAPIRoutes(startupCtx, server, runtime)
	if err = registerSourcePluginHTTPRoutes(startupCtx, server, runtime); err != nil {
		logger.Errorf(startupCtx, "register plugin routes failed: %v", err)
		return err
	}
	if err = finishHTTPRuntimeAfterSourceRoutes(startupCtx, runtime); err != nil {
		return err
	}
	completeSourcePluginHTTPRoutes(startupCtx, ctx, runtime)
	if err = bindFrontendAssetRoutes(startupCtx, server, runtime.pluginSvc, configSvc.GetWorkspaceBasePath(startupCtx)); err != nil {
		return err
	}

	bindHostedOpenAPIDocs(startupCtx, server, runtime.apiDocSvc, runtime.serverCfg.ApiDocPath, runtime.i18nSvc, runtime.bizCtxSvc)
	logHTTPStartupSummary(startupCtx, startupCollector)
	dispatchSystemStartedHook(ctx, runtime.pluginSvc)

	server.Run()
	return shutdownHTTPRuntime(ctx, runtime, server)
}
