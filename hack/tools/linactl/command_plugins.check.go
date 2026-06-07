// This file implements the plugins.check command for plugin governance scans.
// The command is a thin linactl wrapper around the cross-platform Go scanner
// in internal/plugingovernance.

package main

import (
	"context"

	"linactl/internal/plugingovernance"
)

// runPluginsCheck invokes the plugin governance scanner.
func runPluginsCheck(_ context.Context, a *app, input commandInput) error {
	return plugingovernance.RunCheck(a.root, a.stdout, plugingovernance.Options{
		Format: input.GetDefault("format", "text"),
	})
}
