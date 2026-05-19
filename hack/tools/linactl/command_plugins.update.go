// This file implements the plugins.update command for configured source plugins.

package main

import (
	"context"

	"linactl/internal/plugins"
)

// runPluginsUpdate updates configured plugins in apps/lina-plugins.
func runPluginsUpdate(ctx context.Context, a *app, input commandInput) error {
	return plugins.InstallOrUpdate(ctx, pluginRuntime(a), input, true)
}
