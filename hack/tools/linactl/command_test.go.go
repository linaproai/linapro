// This file implements the test.go command for Go workspace test execution.

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"linactl/internal/plugins"
	"linactl/internal/toolutil"
)

// runTestGo runs Go tests for each workspace module.
func runTestGo(ctx context.Context, a *app, input commandInput) error {
	race, err := input.Bool("race", true)
	if err != nil {
		return err
	}
	verbose, err := input.Bool("verbose", true)
	if err != nil {
		return err
	}
	_, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
	if err != nil {
		return err
	}

	workspaceApp := *a
	workspaceApp.env = env
	modules, err := goWorkspaceModules(ctx, &workspaceApp)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return errors.New("no Go workspace modules discovered")
	}
	for _, moduleDir := range modules {
		// Backend packages share the CI PostgreSQL schema and plugin runtime
		// artifacts, so package-level parallelism can expose transient fixture
		// rows from another package process. Keep packages serial while still
		// allowing each package to run with its normal test behavior.
		args := []string{"test", "-p=1"}
		if race {
			args = append(args, "-race")
		}
		if verbose {
			args = append(args, "-v")
		}
		args = append(args, "./...")
		fmt.Fprintf(a.stdout, "==> go %s (%s)\n", strings.Join(args, " "), toolutil.RelativePath(a.root, moduleDir))
		if err = a.runCommand(ctx, commandOptions{Dir: moduleDir, Env: env}, "go", args...); err != nil {
			return err
		}
	}
	return nil
}

// goWorkspaceModules lists module directories from the current Go workspace.
func goWorkspaceModules(ctx context.Context, a *app) ([]string, error) {
	cmd := a.execCommand(ctx, "go", "list", "-m", "-f", "{{.Dir}}")
	cmd.Dir = a.root
	cmd.Env = a.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return nil, fmt.Errorf("list Go workspace modules: %w: %s", err, message)
		}
		return nil, fmt.Errorf("list Go workspace modules: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var modules []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !isGeneratedOfficialPluginAggregateModule(a.root, line) {
			modules = append(modules, line)
		}
	}
	return modules, nil
}

// isGeneratedOfficialPluginAggregateModule reports whether a module directory
// is the ignored aggregate bridge used only to satisfy host blank imports.
func isGeneratedOfficialPluginAggregateModule(root string, moduleDir string) bool {
	if strings.TrimSpace(moduleDir) == "" {
		return false
	}
	return filepath.Clean(moduleDir) == filepath.Clean(plugins.AggregateModuleDir(root))
}
