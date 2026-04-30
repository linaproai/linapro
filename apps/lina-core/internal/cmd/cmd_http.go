// This file contains the HTTP command entrypoint and delegates detailed
// startup responsibilities to focused HTTP helper files.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
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
	if err = shutdownHTTPRuntime(ctx, runtime, configSvc); err != nil {
		return nil, err
	}
	return
}
