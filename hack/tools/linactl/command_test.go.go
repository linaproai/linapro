// This file implements the test.go command for Go workspace test execution.
// It discovers workspace modules, separates real test packages from compile
// smoke packages, and keeps package processes serial for shared backend
// fixtures such as PostgreSQL schemas and plugin runtime artifacts.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

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
	pluginsEnabled, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
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

	plans := make([]goTestModulePlan, 0, len(modules))
	totalTestPackages := 0
	totalCompilePackages := 0
	for _, moduleDir := range modules {
		plan, planErr := goTestModulePlanForDir(ctx, &workspaceApp, moduleDir)
		if planErr != nil {
			return planErr
		}
		plans = append(plans, plan)
		totalTestPackages += len(plan.TestPackages)
		totalCompilePackages += len(plan.CompilePackages)
	}
	fmt.Fprintf(
		a.stdout,
		"Go unit test plan: modules=%d test_packages=%d compile_smoke_packages=%d race=%t verbose=%t plugins=%t\n",
		len(plans),
		totalTestPackages,
		totalCompilePackages,
		race,
		verbose,
		pluginsEnabled,
	)

	summaries := make([]goTestModuleSummary, 0, len(plans))
	for _, plan := range plans {
		startedAt := time.Now()
		moduleLabel := toolutil.RelativePath(a.root, plan.ModuleDir)
		if len(plan.TestPackages) > 0 {
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
			args = append(args, plan.TestPackages...)
			fmt.Fprintf(a.stdout, "==> go %s (%s)\n", strings.Join(args, " "), moduleLabel)
			if err = a.runCommand(ctx, commandOptions{Dir: plan.ModuleDir, Env: env}, "go", args...); err != nil {
				return err
			}
		} else {
			fmt.Fprintf(a.stdout, "==> no Go test packages discovered (%s)\n", moduleLabel)
		}
		if len(plan.CompilePackages) > 0 {
			args := []string{"test", "-p=1"}
			if race {
				args = append(args, "-race")
			}
			args = append(args, "-run", "^$")
			args = append(args, plan.CompilePackages...)
			fmt.Fprintf(a.stdout, "==> go %s (%s compile smoke)\n", strings.Join(args, " "), moduleLabel)
			if err = a.runCommand(ctx, commandOptions{Dir: plan.ModuleDir, Env: env}, "go", args...); err != nil {
				return err
			}
		}
		summaries = append(summaries, goTestModuleSummary{
			ModuleDir:        plan.ModuleDir,
			TestPackages:     len(plan.TestPackages),
			CompilePackages:  len(plan.CompilePackages),
			Elapsed:          time.Since(startedAt),
			HasTestExecution: len(plan.TestPackages) > 0,
		})
	}
	fmt.Fprintf(a.stdout, "Go unit test summary: modules=%d test_packages=%d compile_smoke_packages=%d\n", len(plans), totalTestPackages, totalCompilePackages)
	for _, summary := range summaries {
		fmt.Fprintf(
			a.stdout,
			"- %s: test_packages=%d compile_smoke_packages=%d ran_tests=%t elapsed=%s\n",
			toolutil.RelativePath(a.root, summary.ModuleDir),
			summary.TestPackages,
			summary.CompilePackages,
			summary.HasTestExecution,
			summary.Elapsed.Truncate(time.Millisecond),
		)
	}
	return nil
}

// goListPackage captures the go list metadata needed to classify packages as
// test-bearing packages or compile-smoke-only packages.
type goListPackage struct {
	ImportPath   string
	TestGoFiles  []string
	XTestGoFiles []string
}

// goTestModulePlan describes which packages in one module should execute
// tests and which packages only need a compile smoke command.
type goTestModulePlan struct {
	ModuleDir       string
	TestPackages    []string
	CompilePackages []string
}

// goTestModuleSummary records the package counts and elapsed time reported
// after one module's Go test plan finishes.
type goTestModuleSummary struct {
	ModuleDir        string
	TestPackages     int
	CompilePackages  int
	HasTestExecution bool
	Elapsed          time.Duration
}

// goTestModulePlanForDir lists packages in one module and separates packages
// with test files from packages that only need a lightweight compile smoke.
func goTestModulePlanForDir(ctx context.Context, a *app, moduleDir string) (goTestModulePlan, error) {
	cmd := a.execCommand(ctx, "go", "list", "-json", "./...")
	cmd.Dir = moduleDir
	cmd.Env = a.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return goTestModulePlan{}, fmt.Errorf("list Go packages for %s: %w: %s", toolutil.RelativePath(a.root, moduleDir), err, message)
		}
		return goTestModulePlan{}, fmt.Errorf("list Go packages for %s: %w", toolutil.RelativePath(a.root, moduleDir), err)
	}

	plan := goTestModulePlan{ModuleDir: moduleDir}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg goListPackage
		decodeErr := decoder.Decode(&pkg)
		if errors.Is(decodeErr, io.EOF) {
			break
		}
		if decodeErr != nil {
			return goTestModulePlan{}, fmt.Errorf("decode Go package list for %s: %w", toolutil.RelativePath(a.root, moduleDir), decodeErr)
		}
		if strings.TrimSpace(pkg.ImportPath) == "" {
			continue
		}
		if len(pkg.TestGoFiles) > 0 || len(pkg.XTestGoFiles) > 0 {
			plan.TestPackages = append(plan.TestPackages, pkg.ImportPath)
			continue
		}
		plan.CompilePackages = append(plan.CompilePackages, pkg.ImportPath)
	}
	return plan, nil
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
