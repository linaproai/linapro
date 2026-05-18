// This file classifies official plugin workspaces and resolves plugin build mode.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// inspectOfficialPluginWorkspace classifies the submodule checkout without
// parsing every manifest, keeping preflight checks fast and side-effect free.
func inspectOfficialPluginWorkspace(root string) (officialPluginWorkspace, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	state := officialPluginWorkspace{Root: pluginRoot, State: pluginWorkspaceStateMissing}
	info, err := os.Stat(pluginRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, fmt.Errorf("stat official plugin workspace: %w", err)
	}
	if !info.IsDir() {
		state.State = pluginWorkspaceStateInvalid
		return state, nil
	}
	matches, err := filepath.Glob(filepath.Join(pluginRoot, "*", "plugin.yaml"))
	if err != nil {
		return state, fmt.Errorf("scan official plugin manifests: %w", err)
	}
	state.ManifestCount = len(matches)
	if len(matches) == 0 {
		state.State = pluginWorkspaceStateEmpty
		return state, nil
	}
	state.State = pluginWorkspaceStateReady
	return state, nil
}

// requireOfficialPluginWorkspace fails with an actionable submodule hint when
// a command explicitly needs official plugin source files.
func requireOfficialPluginWorkspace(root string) error {
	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		return err
	}
	return requireOfficialPluginWorkspaceState(root, workspace)
}

// requireOfficialPluginWorkspaceState fails with an actionable submodule hint
// when a command explicitly needs official plugin source files.
func requireOfficialPluginWorkspaceState(root string, workspace officialPluginWorkspace) error {
	if workspace.State == pluginWorkspaceStateReady {
		return nil
	}
	return fmt.Errorf(
		"official plugin workspace is %s at %s; initialize it with `%s`",
		workspace.State,
		relativePath(root, workspace.Root),
		officialPluginInitCommand,
	)
}

// prepareOfficialPluginBuildEnv resolves plugin mode, prepares the temporary
// plugin workspace when required, and returns the selected process environment.
func prepareOfficialPluginBuildEnv(_ context.Context, a *app, input commandInput) (bool, []string, error) {
	enabled, workspace, err := resolveOfficialPluginBuildMode(a.root, input)
	if err != nil {
		return false, nil, err
	}
	workspacePath, err := prepareOfficialPluginWorkspace(a.root, enabled, workspace)
	if err != nil {
		return false, nil, err
	}
	if enabled {
		fmt.Fprintf(a.stdout, "Official plugin mode: enabled (%d manifests)\n", workspace.ManifestCount)
	} else {
		fmt.Fprintln(a.stdout, "Official plugin mode: host-only")
	}
	return enabled, officialPluginBuildEnv(a.root, a.env, enabled, workspacePath), nil
}

// resolveOfficialPluginBuildMode honors an explicit plugins parameter and
// otherwise auto-enables plugins when manifests are present in apps/lina-plugins.
func resolveOfficialPluginBuildMode(root string, input commandInput) (bool, officialPluginWorkspace, error) {
	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		return false, workspace, err
	}
	if input.Has("plugins") {
		if strings.EqualFold(strings.TrimSpace(input.Get("plugins")), "auto") {
			return workspace.State == pluginWorkspaceStateReady, workspace, nil
		}
		enabled, parseErr := input.Bool("plugins", false)
		if parseErr != nil {
			return false, workspace, parseErr
		}
		if enabled {
			if requireErr := requireOfficialPluginWorkspaceState(root, workspace); requireErr != nil {
				return false, workspace, requireErr
			}
		}
		return enabled, workspace, nil
	}
	return workspace.State == pluginWorkspaceStateReady, workspace, nil
}

// prepareOfficialPluginWorkspace writes the ignored temporary Go workspace used
// by plugin-full builds. Host-only builds leave the root workspace untouched.
func prepareOfficialPluginWorkspace(root string, enabled bool, workspace officialPluginWorkspace) (string, error) {
	if !enabled {
		return "", nil
	}
	return writeOfficialPluginWorkspace(root, workspace)
}
