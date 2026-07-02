// This file implements Go static lint commands. It runs golangci-lint through
// the current Go workspace modules and applies a dedicated multi-target dead
// code check for packages that contain `wasip1`-gated guest code.

package main

import (
	"context"
	"errors"
	"fmt"
	"go/build/constraint"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"linactl/internal/toolutil"
)

const (
	goLintConfigFile              = ".golangci.yml"
	goLintVersionFile             = ".golangci-lint-version"
	goLintModulePath              = "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	goLintBinaryName              = "golangci-lint"
	goLintStaticcheckVersionFile  = ".staticcheck-version"
	goLintStaticcheckModulePath   = "honnef.co/go/tools/cmd/staticcheck"
	goLintStaticcheckBinaryName   = "staticcheck"
	goLintStaticcheckDeadCodeFlag = "U1000"
)

// goLintToolSpec describes a Go-installed command-line tool that repository Go
// lint commands must pin before use.
type goLintToolSpec struct {
	VersionFile string
	ModulePath  string
	BinaryName  string
	VersionArg  string
}

// goLintModulePlan captures how one module should be linted, including any
// package split needed for build-tag-sensitive dead-code analysis.
type goLintModulePlan struct {
	ModuleDir              string
	HostPackages           []string
	GuestSensitivePackages []string
}

// goLintModuleSummary records per-module lint timing for command output.
type goLintModuleSummary struct {
	ModuleDir              string
	GuestSensitivePackages int
	Elapsed                time.Duration
}

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

	plans, err := goLintModulePlans(ctx, &workspaceApp, modules)
	if err != nil {
		return err
	}

	configPath := filepath.Join(a.root, goLintConfigFile)
	golangciLintPath, golangciLintVersion, err := ensureGoLintBinary(ctx, a)
	if err != nil {
		return err
	}
	staticcheckPath, staticcheckVersion, err := ensureGoLintStaticcheckBinary(ctx, a)
	if err != nil {
		return err
	}

	var guestPackageCount int
	for _, plan := range plans {
		guestPackageCount += len(plan.GuestSensitivePackages)
	}
	fmt.Fprintf(
		a.stdout,
		"Go lint plan: modules=%d plugins=%t fix=%t config=%s golangci-lint=%s staticcheck=%s guestPackages=%d\n",
		len(plans),
		pluginsEnabled,
		fix,
		toolutil.RelativePath(a.root, configPath),
		golangciLintVersion,
		staticcheckVersion,
		guestPackageCount,
	)

	summaries := make([]goLintModuleSummary, 0, len(plans))
	for _, plan := range plans {
		startedAt := time.Now()
		if err = runGoLintModulePlan(ctx, a, env, golangciLintPath, staticcheckPath, configPath, fix, plan); err != nil {
			return err
		}
		summaries = append(summaries, goLintModuleSummary{
			ModuleDir:              plan.ModuleDir,
			GuestSensitivePackages: len(plan.GuestSensitivePackages),
			Elapsed:                time.Since(startedAt),
		})
	}

	fmt.Fprintf(
		a.stdout,
		"Go lint summary: modules=%d plugins=%t fix=%t guestPackages=%d\n",
		len(summaries),
		pluginsEnabled,
		fix,
		guestPackageCount,
	)
	for _, summary := range summaries {
		fmt.Fprintf(
			a.stdout,
			"- %s: guestPackages=%d elapsed=%s\n",
			toolutil.RelativePath(a.root, summary.ModuleDir),
			summary.GuestSensitivePackages,
			summary.Elapsed.Truncate(time.Millisecond),
		)
	}
	return nil
}

// runGoLintModulePlan executes the appropriate lint commands for one module.
func runGoLintModulePlan(
	ctx context.Context,
	a *app,
	env []string,
	golangciLintPath string,
	staticcheckPath string,
	configPath string,
	fix bool,
	plan goLintModulePlan,
) error {
	moduleLabel := toolutil.RelativePath(a.root, plan.ModuleDir)
	if err := runGoLintCommand(
		ctx,
		a,
		plan.ModuleDir,
		env,
		golangciLintPath,
		moduleLabel,
		goLintCommandArgs(configPath, fix, []string{"./..."}),
	); err != nil {
		return err
	}

	if len(plan.HostPackages) > 0 {
		if err := runGoLintStaticcheck(
			ctx,
			a,
			plan.ModuleDir,
			env,
			moduleLabel+" deadcode",
			staticcheckPath,
			plan.HostPackages,
		); err != nil {
			return err
		}
	}
	if len(plan.GuestSensitivePackages) > 0 {
		if err := runGoLintStaticcheckMatrix(
			ctx,
			a,
			plan.ModuleDir,
			env,
			moduleLabel+" guest-sensitive deadcode",
			staticcheckPath,
			plan.GuestSensitivePackages,
		); err != nil {
			return err
		}
	}
	return nil
}

// runGoLintCommand executes one golangci-lint invocation and prints an
// auditable label before dispatch.
func runGoLintCommand(
	ctx context.Context,
	a *app,
	moduleDir string,
	env []string,
	golangciLintPath string,
	label string,
	args []string,
) error {
	fmt.Fprintf(a.stdout, "==> %s %s (%s)\n", goLintDisplayBinary(a.root, goLintBinaryName, golangciLintPath), strings.Join(args, " "), label)
	return a.runCommand(ctx, commandOptions{Dir: moduleDir, Env: env}, golangciLintPath, args...)
}

// runGoLintStaticcheck executes the dead-code check for packages whose symbols
// are evaluated under the current host build target.
func runGoLintStaticcheck(
	ctx context.Context,
	a *app,
	moduleDir string,
	env []string,
	label string,
	staticcheckPath string,
	packages []string,
) error {
	args := []string{
		"-checks=" + goLintStaticcheckDeadCodeFlag,
		"-tests=false",
	}
	args = append(args, packages...)
	fmt.Fprintf(a.stdout, "==> %s %s (%s)\n", goLintDisplayBinary(a.root, goLintStaticcheckBinaryName, staticcheckPath), strings.Join(args, " "), label)
	return a.runCommand(ctx, commandOptions{Dir: moduleDir, Env: env}, staticcheckPath, args...)
}

// runGoLintStaticcheckMatrix executes the dead-code check for build-tag-
// sensitive packages across both host and wasm guest targets.
func runGoLintStaticcheckMatrix(
	ctx context.Context,
	a *app,
	moduleDir string,
	env []string,
	label string,
	staticcheckPath string,
	packages []string,
) error {
	args := []string{
		"-checks=" + goLintStaticcheckDeadCodeFlag,
		"-tests=false",
		"-matrix",
	}
	args = append(args, packages...)
	fmt.Fprintf(a.stdout, "==> %s %s (%s)\n", goLintDisplayBinary(a.root, goLintStaticcheckBinaryName, staticcheckPath), strings.Join(args, " "), label)
	return a.runCommand(
		ctx,
		commandOptions{
			Dir:   moduleDir,
			Env:   env,
			Stdin: strings.NewReader(goLintStaticcheckMatrixInput()),
		},
		staticcheckPath,
		args...,
	)
}

// goLintCommandArgs builds one golangci-lint argument list for a module.
func goLintCommandArgs(configPath string, fix bool, packages []string) []string {
	args := []string{"run", "--config", configPath}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, packages...)
	return args
}

// ensureGoLintBinary returns a golangci-lint binary matching the repository
// pinned version, installing it with `go install` when needed.
func ensureGoLintBinary(ctx context.Context, a *app) (string, string, error) {
	return ensureGoLintToolBinary(ctx, a, goLintToolSpec{
		VersionFile: goLintVersionFile,
		ModulePath:  goLintModulePath,
		BinaryName:  goLintBinaryName,
		VersionArg:  "--version",
	})
}

// ensureGoLintStaticcheckBinary returns a staticcheck binary matching the
// repository pinned version, installing it with `go install` when needed.
func ensureGoLintStaticcheckBinary(ctx context.Context, a *app) (string, string, error) {
	return ensureGoLintToolBinary(ctx, a, goLintToolSpec{
		VersionFile: goLintStaticcheckVersionFile,
		ModulePath:  goLintStaticcheckModulePath,
		BinaryName:  goLintStaticcheckBinaryName,
		VersionArg:  "-version",
	})
}

func ensureGoLintToolBinary(ctx context.Context, a *app, spec goLintToolSpec) (string, string, error) {
	version, err := loadGoLintToolVersion(a.root, spec.VersionFile)
	if err != nil {
		return "", "", err
	}

	if path, lookErr := a.lookPath(spec.BinaryName); lookErr == nil {
		if goLintBinaryVersionMatches(ctx, a, a.env, path, spec.VersionArg, version) {
			fmt.Fprintf(a.stdout, "Using %s %s from PATH\n", spec.BinaryName, version)
			return spec.BinaryName, version, nil
		}
	}

	installEnv := goLintToolInstallEnv(a.env)
	binaryPath, err := goLintInstalledBinaryPath(ctx, a, installEnv, spec.BinaryName)
	if err != nil {
		return "", "", err
	}
	if toolutil.FileExists(binaryPath) && goLintBinaryVersionMatches(ctx, a, installEnv, binaryPath, spec.VersionArg, version) {
		fmt.Fprintf(a.stdout, "Using %s %s from %s\n", spec.BinaryName, version, binaryPath)
		return binaryPath, version, nil
	}

	moduleVersion := spec.ModulePath + "@" + version
	fmt.Fprintf(a.stdout, "Installing %s %s via go install\n", spec.BinaryName, version)
	if err = a.runCommand(ctx, commandOptions{Dir: a.root, Env: installEnv}, "go", "install", moduleVersion); err != nil {
		return "", "", err
	}
	if !goLintBinaryVersionMatches(ctx, a, installEnv, binaryPath, spec.VersionArg, version) {
		return "", "", fmt.Errorf("installed %s does not report expected version %s", binaryPath, version)
	}
	return binaryPath, version, nil
}

func loadGoLintToolVersion(root string, versionFile string) (string, error) {
	versionPath := filepath.Join(root, versionFile)
	content, err := os.ReadFile(versionPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", toolutil.RelativePath(root, versionPath), err)
	}
	version := strings.TrimSpace(string(content))
	if version == "" {
		return "", fmt.Errorf("%s is empty", toolutil.RelativePath(root, versionPath))
	}
	return version, nil
}

// goLintToolInstallEnv isolates external tool installation from repository
// workspaces, build tags, and cross-compilation variables.
func goLintToolInstallEnv(env []string) []string {
	next := toolutil.SetEnvValue(env, "GOWORK", "off")
	for _, key := range []string{"GOFLAGS", "GOOS", "GOARCH", "GOARM", "GOAMD64", "GOWASM"} {
		next = toolutil.RemoveEnvValue(next, key)
	}
	return next
}

// goLintInstalledBinaryPath resolves where `go install` writes one lint tool.
func goLintInstalledBinaryPath(ctx context.Context, a *app, env []string, binaryName string) (string, error) {
	gobin, err := goLintGoEnv(ctx, a, env, "GOBIN")
	if err != nil {
		return "", err
	}
	if gobin != "" {
		return filepath.Join(gobin, toolutil.ExecutableName(binaryName)), nil
	}

	gopath, err := goLintGoEnv(ctx, a, env, "GOPATH")
	if err != nil {
		return "", err
	}
	for _, item := range filepath.SplitList(gopath) {
		item = strings.TrimSpace(item)
		if item != "" {
			return filepath.Join(item, "bin", toolutil.ExecutableName(binaryName)), nil
		}
	}
	return "", fmt.Errorf("go env GOPATH is empty; cannot locate installed %s binary", binaryName)
}

// goLintGoEnv reads one `go env` value with the tool-install environment.
func goLintGoEnv(ctx context.Context, a *app, env []string, key string) (string, error) {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Env: env}, "go", "env", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// goLintBinaryVersionMatches checks whether one binary reports the pinned
// version. Failed probes are treated as mismatch so auto-install can recover.
func goLintBinaryVersionMatches(ctx context.Context, a *app, env []string, binaryPath string, versionArg string, version string) bool {
	output, err := a.runCommandOutput(ctx, commandOptions{Dir: a.root, Env: env}, binaryPath, versionArg)
	if err != nil {
		return false
	}
	return goLintVersionOutputMatches(output, version)
}

// goLintVersionOutputMatches parses lint tool version output without depending
// on an exact upstream sentence shape.
func goLintVersionOutputMatches(output string, version string) bool {
	expected := strings.TrimPrefix(strings.TrimSpace(version), "v")
	for _, field := range strings.Fields(output) {
		field = strings.Trim(field, " \t\r\n,;()[]{}")
		field = strings.TrimPrefix(field, "v")
		if field == expected {
			return true
		}
	}
	return false
}

// goLintDisplayBinary keeps lint command logs concise while still showing
// absolute installed paths when PATH cannot provide the pinned binary.
func goLintDisplayBinary(root string, binaryName string, binaryPath string) string {
	if binaryPath == binaryName {
		return binaryName
	}
	return toolutil.RelativePath(root, binaryPath)
}

// goLintStaticcheckMatrixInput returns the build matrix consumed by
// `staticcheck -matrix`.
func goLintStaticcheckMatrixInput() string {
	return "host:\nwasm_guest: GOOS=wasip1 GOARCH=wasm\n"
}

// goLintModulePlans lists how each workspace module should be linted.
func goLintModulePlans(ctx context.Context, a *app, moduleDirs []string) ([]goLintModulePlan, error) {
	plans := make([]goLintModulePlan, 0, len(moduleDirs))
	for _, moduleDir := range moduleDirs {
		plan, err := goLintModulePlanForDir(ctx, a, moduleDir)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// goLintModulePlanForDir discovers packages in one module and isolates package
// directories that contain `wasip1` build-tag splits.
func goLintModulePlanForDir(ctx context.Context, a *app, moduleDir string) (goLintModulePlan, error) {
	packages, err := goLintPackagesForDir(ctx, a, moduleDir)
	if err != nil {
		return goLintModulePlan{}, err
	}
	guestSensitivePackages, err := goLintGuestSensitivePackages(moduleDir, packages)
	if err != nil {
		return goLintModulePlan{}, err
	}
	guestSensitiveSet := make(map[string]struct{}, len(guestSensitivePackages))
	for _, pkg := range guestSensitivePackages {
		guestSensitiveSet[pkg] = struct{}{}
	}

	hostPackages := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if _, ok := guestSensitiveSet[pkg]; ok {
			continue
		}
		hostPackages = append(hostPackages, pkg)
	}
	return goLintModulePlan{
		ModuleDir:              moduleDir,
		HostPackages:           hostPackages,
		GuestSensitivePackages: guestSensitivePackages,
	}, nil
}

// goLintPackagesForDir lists package directory patterns for one module.
func goLintPackagesForDir(ctx context.Context, a *app, moduleDir string) ([]string, error) {
	output, stderr, err := runGoDiscoveryCommand(ctx, a, moduleDir, "list", "-f", "{{.Dir}}", "./...")
	if err != nil {
		message := goDiscoveryErrorOutput(output, stderr)
		if message != "" {
			return nil, fmt.Errorf("list Go packages for %s: %w: %s", toolutil.RelativePath(a.root, moduleDir), err, message)
		}
		return nil, fmt.Errorf("list Go packages for %s: %w", toolutil.RelativePath(a.root, moduleDir), err)
	}

	packageSet := make(map[string]struct{})
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		dir := strings.TrimSpace(line)
		if dir == "" {
			continue
		}
		pkg, err := goLintPackagePatternForDir(moduleDir, dir)
		if err != nil {
			return nil, fmt.Errorf("normalize Go package dir %s: %w", dir, err)
		}
		packageSet[pkg] = struct{}{}
	}

	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages, nil
}

// goLintPackagePatternForDir converts one absolute package directory into the
// relative package argument expected by `golangci-lint` and `staticcheck`.
func goLintPackagePatternForDir(moduleDir string, packageDir string) (string, error) {
	moduleDir, err := goLintCanonicalDir(moduleDir)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(packageDir) {
		packageDir = filepath.Join(moduleDir, packageDir)
	}
	packageDir, err = goLintCanonicalDir(packageDir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(moduleDir, packageDir)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	return "./" + filepath.ToSlash(rel), nil
}

// goLintCanonicalDir normalizes a directory path so package discovery stays
// stable when `go list` resolves workspace paths through platform symlinks
// such as `/var` -> `/private/var` on macOS.
func goLintCanonicalDir(dir string) (string, error) {
	dir = filepath.Clean(dir)
	if !filepath.IsAbs(dir) {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		dir = absDir
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return dir, nil
		}
		return "", err
	}
	return resolvedDir, nil
}

// goLintGuestSensitivePackages returns package patterns whose non-test Go files
// contain `wasip1` build constraints.
func goLintGuestSensitivePackages(moduleDir string, packages []string) ([]string, error) {
	guestSensitive := make([]string, 0)
	for _, pkg := range packages {
		dir := moduleDir
		if pkg != "." {
			dir = filepath.Join(moduleDir, filepath.FromSlash(strings.TrimPrefix(pkg, "./")))
		}
		hasWasip1Build, err := goLintPackageDirHasWasip1Build(dir)
		if err != nil {
			return nil, err
		}
		if hasWasip1Build {
			guestSensitive = append(guestSensitive, pkg)
		}
	}
	sort.Strings(guestSensitive)
	return guestSensitive, nil
}

// goLintPackageDirHasWasip1Build reports whether one package directory contains
// non-test Go files gated by `wasip1` build constraints.
func goLintPackageDirHasWasip1Build(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read package dir %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return false, fmt.Errorf("read %s: %w", filepath.Join(dir, name), err)
		}
		if goLintFileHasWasip1Build(content) {
			return true, nil
		}
	}
	return false, nil
}

func goLintFileHasWasip1Build(content []byte) bool {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "//") {
			return false
		}
		if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			continue
		}
		if goLintBuildExprMentionsTag(expr, "wasip1") {
			return true
		}
	}
	return false
}

func goLintBuildExprMentionsTag(expr constraint.Expr, tag string) bool {
	switch expr := expr.(type) {
	case *constraint.TagExpr:
		return expr.Tag == tag
	case *constraint.NotExpr:
		return goLintBuildExprMentionsTag(expr.X, tag)
	case *constraint.AndExpr:
		return goLintBuildExprMentionsTag(expr.X, tag) || goLintBuildExprMentionsTag(expr.Y, tag)
	case *constraint.OrExpr:
		return goLintBuildExprMentionsTag(expr.X, tag) || goLintBuildExprMentionsTag(expr.Y, tag)
	default:
		return false
	}
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
