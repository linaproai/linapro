// This file dispatches host lifecycle hooks around HTTP startup.

package cmd

import (
	"context"

	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

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
