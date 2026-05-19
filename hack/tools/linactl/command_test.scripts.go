// This file implements the test.scripts command for repository tooling smoke checks.

package main

import (
	"context"
	"fmt"
	"path/filepath"

	"linactl/internal/repository"
)

// runTestScripts runs cross-platform repository tooling smoke checks.
func runTestScripts(ctx context.Context, a *app, _ commandInput) error {
	if err := repository.ValidateTooling(a.root, commandNames()); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "==> go test . (hack/tools/linactl)")
	if err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tools", "linactl")}, "go", "test", ".", "-count=1"); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "repository tool smoke checks passed")
	return nil
}
