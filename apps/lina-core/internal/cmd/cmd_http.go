// This file contains the HTTP command entrypoint and delegates detailed
// startup responsibilities to focused HTTP helper files.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
	"lina-core/pkg/logger"
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
	s := g.Server()
	configSvc := config.New()
	if err = configureHTTPServer(ctx, s, configSvc); err != nil {
		return nil, err
	}

	runtime, err := newHTTPRuntime(ctx, configSvc)
	if err != nil {
		return nil, err
	}
	startupCtx, startupCollector, err := newHTTPStartupContext(ctx, runtime)
	if err != nil {
		return nil, err
	}
	if err = startHTTPRuntimeBeforeSourceRoutes(startupCtx, runtime); err != nil {
		return nil, err
	}

	bindHostAPIRoutes(startupCtx, s, runtime)
	if err = registerSourcePluginHTTPRoutes(startupCtx, s, runtime); err != nil {
		logger.Panicf(startupCtx, "register plugin routes failed: %v", err)
	}
	if err = finishHTTPRuntimeAfterSourceRoutes(startupCtx, runtime); err != nil {
		return nil, err
	}
	completeSourcePluginHTTPRoutes(startupCtx, ctx, runtime)
	if err = bindFrontendAssetRoutes(startupCtx, s, runtime.pluginSvc); err != nil {
		return nil, err
	}

	bindHostedOpenAPIDocs(startupCtx, s, runtime.apiDocSvc, runtime.serverCfg.ApiDocPath, runtime.i18nSvc, runtime.bizCtxSvc)
	logHTTPStartupSummary(startupCtx, startupCollector)
	dispatchSystemStartedHook(ctx, runtime.pluginSvc)

	s.Run()
	if err = shutdownHTTPRuntime(ctx, runtime, configSvc); err != nil {
		return nil, err
	}
	return
}
