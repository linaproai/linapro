// This file implements the plugins.init command for source-plugin workspace setup.

package main

import (
	"context"

	"linactl/internal/plugins"
)

// runPluginsInit converts apps/lina-plugins from a submodule into an ordinary
// source-plugin directory while preserving files.
func runPluginsInit(ctx context.Context, a *app, _ commandInput) error {
	_, err := plugins.EnsureManagedWorkspaceReady(ctx, pluginRuntime(a), true)
	return err
}
