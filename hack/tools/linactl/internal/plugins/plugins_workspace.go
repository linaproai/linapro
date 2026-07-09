// This file classifies official plugin workspaces and resolves plugin build
// mode. It owns the official-plugin preflight path used by dev, build, wasm,
// and test commands.

package plugins

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"linactl/internal/toolutil"
)

// InspectOfficialWorkspace classifies the submodule checkout without
// parsing every manifest, keeping preflight checks fast and side-effect free.
func InspectOfficialWorkspace(root string) (OfficialWorkspace, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	state := OfficialWorkspace{Root: pluginRoot, State: WorkspaceStateMissing}
	info, err := os.Stat(pluginRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, fmt.Errorf("stat official plugin workspace: %w", err)
	}
	if !info.IsDir() {
		state.State = WorkspaceStateInvalid
		return state, nil
	}
	matches, err := filepath.Glob(filepath.Join(pluginRoot, "*", "plugin.yaml"))
	if err != nil {
		return state, fmt.Errorf("scan official plugin manifests: %w", err)
	}
	state.ManifestCount = len(matches)
	if len(matches) == 0 {
		state.State = WorkspaceStateEmpty
		return state, nil
	}
	state.State = WorkspaceStateReady
	return state, nil
}

// RequireOfficialWorkspace fails with an actionable submodule hint when
// a command explicitly needs official plugin source files.
func RequireOfficialWorkspace(root string) error {
	workspace, err := InspectOfficialWorkspace(root)
	if err != nil {
		return err
	}
	return RequireOfficialWorkspaceState(root, workspace)
}

// RequireOfficialWorkspaceState fails with an actionable submodule hint
// when a command explicitly needs official plugin source files.
func RequireOfficialWorkspaceState(root string, workspace OfficialWorkspace) error {
	if workspace.State == WorkspaceStateReady {
		return nil
	}
	return fmt.Errorf(
		"official plugin workspace is %s at %s; initialize it with `%s`",
		workspace.State,
		toolutil.RelativePath(root, workspace.Root),
		InitCommand,
	)
}

// PrepareBuildEnv resolves plugin mode, prepares the temporary
// plugin workspace when required, and returns the selected process environment.
func PrepareBuildEnv(_ context.Context, runtime Runtime, input Input) (bool, []string, error) {
	enabled, workspace, err := ResolveBuildMode(runtime.Root, input)
	if err != nil {
		return false, nil, err
	}
	workspacePath, err := PrepareOfficialWorkspace(runtime.Root, enabled, workspace)
	if err != nil {
		return false, nil, err
	}
	if enabled {
		fmt.Fprintf(runtime.Stdout, "Official plugin mode: enabled (%d manifests)\n", workspace.ManifestCount)
	} else {
		fmt.Fprintln(runtime.Stdout, "Official plugin mode: host-only")
	}
	return enabled, BuildEnv(runtime.Root, runtime.Env, enabled, workspacePath), nil
}

// ResolveBuildMode honors an explicit plugins parameter and
// otherwise auto-enables plugins when manifests are present in apps/lina-plugins.
func ResolveBuildMode(root string, input Input) (bool, OfficialWorkspace, error) {
	workspace, err := InspectOfficialWorkspace(root)
	if err != nil {
		return false, workspace, err
	}
	if input.Has("plugins") {
		enabled, parseErr := input.Bool("plugins", false)
		if parseErr != nil {
			return false, workspace, parseErr
		}
		if enabled {
			if requireErr := RequireOfficialWorkspaceState(root, workspace); requireErr != nil {
				return false, workspace, requireErr
			}
		}
		return enabled, workspace, nil
	}
	return workspace.State == WorkspaceStateReady, workspace, nil
}

// PrepareOfficialWorkspace writes the ignored temporary Go workspace used
// by plugin-full builds. Host-only builds leave the root workspace untouched.
func PrepareOfficialWorkspace(root string, enabled bool, workspace OfficialWorkspace) (string, error) {
	if !enabled {
		return "", nil
	}
	return WriteOfficialWorkspace(root, workspace)
}
