// This file implements the plugins.init command for source-plugin workspace setup.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"linactl/internal/plugins"
	"linactl/internal/toolutil"
)

// runPluginsInit converts apps/lina-plugins from a submodule into an ordinary
// source-plugin directory while preserving files.
func runPluginsInit(ctx context.Context, a *app, _ commandInput) error {
	runtime := pluginRuntime(a)
	workspace, err := plugins.InspectManagedWorkspace(ctx, runtime)
	if err != nil {
		return err
	}
	switch workspace.State {
	case plugins.ManagedWorkspaceOrdinary:
		fmt.Fprintf(a.stdout, "Plugin workspace already ordinary: %s\n", toolutil.RelativePath(a.root, workspace.Root))
		return nil
	case plugins.ManagedWorkspaceMissing:
		if err = os.MkdirAll(workspace.Root, 0o755); err != nil {
			return fmt.Errorf("create plugin workspace: %w", err)
		}
		fmt.Fprintf(a.stdout, "Plugin workspace created: %s\n", toolutil.RelativePath(a.root, workspace.Root))
		return nil
	case plugins.ManagedWorkspaceInvalid:
		return fmt.Errorf("plugin workspace is invalid: %s", toolutil.RelativePath(a.root, workspace.Root))
	case plugins.ManagedWorkspaceNestedGit:
		return fmt.Errorf("plugin workspace contains nested Git metadata; remove %s manually before plugins.init", toolutil.RelativePath(a.root, filepath.Join(workspace.Root, ".git")))
	case plugins.ManagedWorkspaceSubmodule:
		return plugins.ConvertSubmoduleToDirectory(ctx, runtime, workspace)
	default:
		return fmt.Errorf("unknown plugin workspace state: %s", workspace.State)
	}
}
