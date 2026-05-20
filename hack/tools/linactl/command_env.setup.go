// This file implements the env.setup command for frontend and browser setup.

package main

import (
	"context"
	"fmt"
	"path/filepath"
)

// runEnvSetup installs frontend dependencies, Playwright browsers, and OS dependencies.
func runEnvSetup(ctx context.Context, a *app, _ commandInput) error {
	if err := ensureFrontendDeps(ctx, a); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(a.stdout, "Installing Playwright Chromium browser and OS dependencies..."); err != nil {
		return fmt.Errorf("write setup output: %w", err)
	}
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "exec", "playwright", "install", "--with-deps", "chromium")
}
