// This file implements the plugins.status command for configured source plugins.

package main

import (
	"context"

	"linactl/internal/plugins"
)

// runPluginsStatus prints read-only plugin workspace and source status.
func runPluginsStatus(ctx context.Context, a *app, input commandInput) error {
	return plugins.Status(ctx, pluginRuntime(a), input)
}
