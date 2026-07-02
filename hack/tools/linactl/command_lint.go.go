// This file implements Go static lint commands. It runs golangci-lint through
// the current Go workspace modules so host-only and plugin-full modes share the
// same plugin workspace preparation logic as build and test commands.

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"linactl/internal/toolutil"
)

// runLintGo runs golangci-lint for each Go module in the selected workspace.
func runLintGo(ctx context.Context, a *app, input commandInput) error {
	fix, err := input.Bool("fix", false)
	if err != nil {
		return err
	}
	pluginsEnabled, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
	if err != nil {
		return err
	}

	workspaceApp := *a
	workspaceApp.env = env
	modules, err := goLintWorkspaceModules(ctx, &workspaceApp)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return errors.New("no Go workspace modules discovered")
	}

	configPath := filepath.Join(a.root, ".golangci.yml")
	args := []string{"run", "--config", configPath}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, "./...")

	fmt.Fprintf(
		a.stdout,
		"Go lint plan: modules=%d plugins=%t fix=%t config=%s\n",
		len(modules),
		pluginsEnabled,
		fix,
		toolutil.RelativePath(a.root, configPath),
	)

	summaries := make([]goLintModuleSummary, 0, len(modules))
	for _, moduleDir := range modules {
		startedAt := time.Now()
		moduleLabel := toolutil.RelativePath(a.root, moduleDir)
		fmt.Fprintf(a.stdout, "==> golangci-lint %s (%s)\n", strings.Join(args, " "), moduleLabel)
		if err = a.runCommand(ctx, commandOptions{Dir: moduleDir, Env: env}, "golangci-lint", args...); err != nil {
			return err
		}
		summaries = append(summaries, goLintModuleSummary{
			ModuleDir: moduleDir,
			Elapsed:   time.Since(startedAt),
		})
	}

	fmt.Fprintf(a.stdout, "Go lint summary: modules=%d plugins=%t fix=%t\n", len(modules), pluginsEnabled, fix)
	for _, summary := range summaries {
		fmt.Fprintf(
			a.stdout,
			"- %s: elapsed=%s\n",
			toolutil.RelativePath(a.root, summary.ModuleDir),
			summary.Elapsed.Truncate(time.Millisecond),
		)
	}
	return nil
}

// goLintModuleSummary records per-module lint timing for command output.
type goLintModuleSummary struct {
	ModuleDir string
	Elapsed   time.Duration
}

// goLintWorkspaceModules lists all modules in the selected workspace. Unlike
// test.go planning, lint keeps generated workspace modules visible so exclusions
// in .golangci.yml remain the single source for generated-code handling.
func goLintWorkspaceModules(ctx context.Context, a *app) ([]string, error) {
	output, stderr, err := runGoDiscoveryCommand(ctx, a, a.root, "list", "-m", "-f", "{{.Dir}}")
	if err != nil {
		message := goDiscoveryErrorOutput(output, stderr)
		if message != "" {
			return nil, fmt.Errorf("list Go workspace modules for lint: %w: %s", err, message)
		}
		return nil, fmt.Errorf("list Go workspace modules for lint: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	modules := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			modules = append(modules, line)
		}
	}
	return modules, nil
}
