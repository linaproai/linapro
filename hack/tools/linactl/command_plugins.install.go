// This file implements the plugins.install command for configured source plugins.

package main

import (
	"context"

	"linactl/internal/plugins"
)

// runPluginsInstall installs configured plugins into apps/lina-plugins.
func runPluginsInstall(ctx context.Context, a *app, input commandInput) error {
	return plugins.InstallOrUpdate(ctx, pluginRuntime(a), input, false)
}
