// Package goframecli wraps embedded GoFrame CLI code generation for linactl.
// Parent linactl commands dispatch to a hidden child linactl command so
// GoFrame CLI process-style exits and global state stay isolated.
package goframecli

import (
	"context"
	"fmt"
	"path/filepath"

	"linactl/internal/toolrun"
)

const hiddenCommand = "__goframe"

// Executable resolves the current linactl executable path.
type Executable func() (string, error)

// Run executes a whitelisted embedded GoFrame CLI command inside the target
// GoFrame project directory through linactl's hidden child command.
func Run(ctx context.Context, target Target, executable Executable, run toolrun.Runner, args ...string) error {
	if err := validateArgs(args); err != nil {
		return err
	}
	binary, err := executable()
	if err != nil {
		return fmt.Errorf("resolve linactl executable: %w", err)
	}
	if target.WorkDir == "" {
		return fmt.Errorf("GoFrame code generation work directory is empty")
	}
	if target.ConfigDir == "" {
		return fmt.Errorf("GoFrame code generation config directory is empty")
	}
	childArgs := append([]string{hiddenCommand, "--config-dir=" + target.ConfigDir}, args...)
	return run(ctx, toolrun.Options{Dir: target.WorkDir}, binary, childArgs...)
}

// coreDir returns the GoFrame project directory used for code generation.
func coreDir(root string) string {
	return filepath.Join(root, "apps", "lina-core")
}

// validateArgs restricts the embedded GoFrame surface to the code generation
// paths owned by linactl.
func validateArgs(args []string) error {
	if len(args) != 2 || args[0] != "gen" || (args[1] != "ctrl" && args[1] != "dao") {
		return fmt.Errorf("embedded GoFrame only supports gen ctrl or gen dao")
	}
	return nil
}
