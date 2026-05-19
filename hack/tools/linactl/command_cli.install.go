// This file implements the cli.install command for conditional GoFrame CLI setup.

package main

import (
	"context"
	"fmt"
)

// runCLIInstallIfMissing installs the GoFrame CLI only when gf is absent.
func runCLIInstallIfMissing(ctx context.Context, a *app, input commandInput) error {
	if err := a.runCommand(ctx, commandOptions{Quiet: true}, "gf", "-v"); err == nil {
		return nil
	}
	fmt.Fprintln(a.stdout, "GoFrame CLI is not installed; starting automatic installation...")
	return runCLIInstall(ctx, a, input)
}
