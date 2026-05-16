// This file implements database, test, GoFrame CLI, and deployment commands.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// runInit initializes the configured database after explicit confirmation.
func runInit(ctx context.Context, a *app, input commandInput) error {
	if input.Get("confirm") != "init" {
		return errors.New("init requires explicit confirmation: linactl init confirm=init")
	}

	args := []string{"run", "main.go", "init", "--confirm=init", "--sql-source=local"}
	if rebuild := input.Get("rebuild"); rebuild != "" {
		args = append(args, "--rebuild="+rebuild)
	}

	var output bytes.Buffer
	err := a.runCommand(ctx, commandOptions{
		Dir:    filepath.Join(a.root, "apps", "lina-core"),
		Stdout: io.MultiWriter(a.stdout, &output),
		Stderr: io.MultiWriter(a.stderr, &output),
	}, "go", args...)
	if err != nil {
		text := strings.ToLower(output.String())
		if isConnectionFailure(text) {
			fmt.Fprintln(a.stderr, "PostgreSQL is not ready. Start PostgreSQL first.")
			fmt.Fprintln(a.stderr, "Local example: docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine")
		}
		return err
	}
	fmt.Fprintln(a.stdout, "Database initialization complete")
	return nil
}

// runMock loads optional mock data after explicit confirmation.
func runMock(ctx context.Context, a *app, input commandInput) error {
	if input.Get("confirm") != "mock" {
		return errors.New("mock requires explicit confirmation: linactl mock confirm=mock")
	}
	err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core")}, "go", "run", "main.go", "mock", "--confirm=mock", "--sql-source=local")
	if err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Mock data load complete")
	return nil
}

// runTest starts the requested Playwright E2E test suite scope.
func runTest(ctx context.Context, a *app, input commandInput) error {
	if err := ensurePlaywrightBrowsers(ctx, a); err != nil {
		return err
	}
	scope := strings.TrimSpace(input.GetDefault("scope", "full"))
	switch {
	case scope == "host":
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:host")
	case scope == "full":
		if err := requireOfficialPluginWorkspace(a.root); err != nil {
			return err
		}
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test")
	case scope == "plugins" || strings.HasPrefix(scope, "plugin:"):
		if err := requireOfficialPluginWorkspace(a.root); err != nil {
			return err
		}
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:module", "--", scope)
	default:
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test:module", "--", scope)
	}
}

// runDevSetup installs frontend dependencies, Playwright browsers, and OS dependencies.
func runDevSetup(ctx context.Context, a *app, _ commandInput) error {
	if err := ensureFrontendDeps(ctx, a); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Installing Playwright Chromium browser and OS dependencies...")
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "exec", "playwright", "install", "--with-deps", "chromium")
}

// ensurePlaywrightBrowsers checks that Playwright's Chromium browser is installed.
// If the browser cache directory is missing, it prints a clear error with the fix command.
func ensurePlaywrightBrowsers(ctx context.Context, a *app) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	var cacheDir string
	switch runtime.GOOS {
	case "linux":
		cacheDir = filepath.Join(home, ".cache", "ms-playwright")
	case "darwin":
		cacheDir = filepath.Join(home, "Library", "Caches", "ms-playwright")
	default:
		// Windows and others: Playwright uses a self-contained bundle; skip detection.
		return nil
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Playwright browsers not installed. Run: make dev.setup")
		}
		return fmt.Errorf("check Playwright browser cache: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "chromium") {
			return nil
		}
	}

	return fmt.Errorf("Playwright Chromium browser not found in %s. Run: make dev.setup", cacheDir)
}

// ensureFrontendDeps checks that the frontend node_modules are installed.
// If the vite binary is missing, it runs pnpm install automatically.
func ensureFrontendDeps(ctx context.Context, a *app) error {
	vite := viteCommand(a.root)
	if _, err := os.Stat(vite); err == nil {
		return nil
	}
	fmt.Fprintln(a.stdout, "Frontend dependencies not installed; running pnpm install...")
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-vben")}, "pnpm", "install")
}

// runTestHost starts the host-owned Playwright E2E test suite.
func runTestHost(ctx context.Context, a *app, _ commandInput) error {
	return runTest(ctx, a, commandInput{Params: map[string]string{"scope": "host"}})
}

// runTestPlugins starts the official plugin Playwright E2E test suite.
func runTestPlugins(ctx context.Context, a *app, _ commandInput) error {
	return runTest(ctx, a, commandInput{Params: map[string]string{"scope": "plugins"}})
}

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
		args := []string{"test"}
		if race {
			args = append(args, "-race")
		}
		if verbose {
			args = append(args, "-v")
		}
		args = append(args, "./...")
		fmt.Fprintf(a.stdout, "==> go %s (%s)\n", strings.Join(args, " "), relativePath(a.root, moduleDir))
		if err = a.runCommand(ctx, commandOptions{Dir: moduleDir, Env: env}, "go", args...); err != nil {
			return err
		}
	}
	return nil
}

// runTestScripts runs cross-platform repository tooling smoke checks.
func runTestScripts(ctx context.Context, a *app, _ commandInput) error {
	if err := validateRepositoryTooling(a.root); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "==> go test . (hack/tools/linactl)")
	if err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tools", "linactl")}, "go", "test", ".", "-count=1"); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "repository tool smoke checks passed")
	return nil
}

// validateRepositoryTooling checks that local tooling entrypoints stay portable.
func validateRepositoryTooling(root string) error {
	makeCmd := filepath.Join(root, "make.cmd")
	content, err := os.ReadFile(makeCmd)
	if err != nil {
		return fmt.Errorf("read make.cmd wrapper: %w", err)
	}
	text := string(content)
	if !strings.Contains(text, "go run . %*") {
		return errors.New("make.cmd must delegate to linactl through go run . %*")
	}
	if strings.Contains(text, "GOWORK=off") {
		return errors.New("make.cmd must not force GOWORK=off")
	}

	legacyDir := filepath.Join(root, "hack", "scripts")
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy hack/scripts directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("hack/scripts contains legacy script %q; move maintained tooling into hack/tools/linactl or another Go tool", entries[0].Name())
	}
	return nil
}

// runI18nCheck invokes all runtime i18n governance checks.
func runI18nCheck(ctx context.Context, a *app, _ commandInput) error {
	toolDir := filepath.Join(a.root, "hack", "tools", "runtime-i18n")
	scanErr := a.runCommand(ctx, commandOptions{Dir: toolDir}, "go", "run", ".", "scan")
	messageErr := a.runCommand(ctx, commandOptions{Dir: toolDir}, "go", "run", ".", "messages")
	return errors.Join(scanErr, messageErr)
}

// runCLIInstall downloads and installs the GoFrame CLI for this platform.
func runCLIInstall(ctx context.Context, a *app, _ commandInput) error {
	tmpDir, err := os.MkdirTemp("", "linapro-gf-*")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			fmt.Fprintf(a.stderr, "warning: remove temporary gf directory: %v\n", removeErr)
		}
	}()

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	binary := executableName("gf")
	url := fmt.Sprintf("https://github.com/gogf/gf/releases/latest/download/gf_%s_%s", goos, goarch)
	archive := filepath.Join(tmpDir, binary)

	fmt.Fprintf(a.stdout, "Downloading GoFrame CLI: %s\n", url)
	if err = downloadFile(ctx, url, archive); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err = os.Chmod(archive, 0o755); err != nil {
			return fmt.Errorf("chmod gf binary: %w", err)
		}
	}
	return a.runCommand(ctx, commandOptions{Dir: tmpDir}, archive, "install", "-y")
}

// runCLIInstallIfMissing installs the GoFrame CLI only when gf is absent.
func runCLIInstallIfMissing(ctx context.Context, a *app, input commandInput) error {
	if err := a.runCommand(ctx, commandOptions{Quiet: true}, "gf", "-v"); err == nil {
		return nil
	}
	fmt.Fprintln(a.stdout, "GoFrame CLI is not installed; starting automatic installation...")
	return runCLIInstall(ctx, a, input)
}

// runGF wraps a GoFrame CLI command inside the core application directory.
func runGF(args ...string) func(context.Context, *app, commandInput) error {
	return func(ctx context.Context, a *app, input commandInput) error {
		if err := runCLIInstallIfMissing(ctx, a, input); err != nil {
			return err
		}
		return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core")}, "gf", args...)
	}
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
	return filepath.Clean(moduleDir) == filepath.Clean(officialPluginAggregateModuleDir(root))
}
