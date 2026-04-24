// Package backend wires the demo-control source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	democontrolplugin "lina-plugin-demo-control"
	middlewaresvc "lina-plugin-demo-control/backend/internal/service/middleware"
)

// demo-control plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "demo-control"
)

// init registers the embedded demo-control source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(democontrolplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerGlobalMiddleware,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerGlobalMiddleware binds the demo read-only guard into the host-wide
// system request chain published to source plugins.
func registerGlobalMiddleware(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	guardSvc := middlewaresvc.New()
	registrar.GlobalMiddlewares().Bind("/*", guardSvc.Guard)
	return nil
}
