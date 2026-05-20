// This file implements the test command. It uses the testcmd suffix because
// Go excludes files named command_test.go from normal builds.

package main

import (
	"context"
	"path/filepath"
	"strings"

	"linactl/internal/playwright"
	"linactl/internal/plugins"
)

// runTest starts the requested Playwright E2E test suite scope.
func runTest(ctx context.Context, a *app, input commandInput) error {
	if err := playwright.EnsureBrowsers(ctx); err != nil {
		return err
	}
	scope := strings.TrimSpace(input.GetDefault("scope", "full"))
	switch {
	case scope == "host":
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:host")
	case scope == "full":
		if err := plugins.RequireOfficialWorkspace(a.root); err != nil {
			return err
		}
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test")
	case scope == "plugins" || strings.HasPrefix(scope, "plugin:"):
		if err := plugins.RequireOfficialWorkspace(a.root); err != nil {
			return err
		}
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:module", "--", scope)
	default:
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:module", "--", scope)
	}
}
