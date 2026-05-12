// Package main implements LinaPro's cross-platform development command entrypoint.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// defaultBackendPort is the standard backend development port.
	defaultBackendPort = 8080
	// defaultFrontendPort is the standard frontend development port.
	defaultFrontendPort = 5666
	// defaultWaitTimeout bounds development service readiness checks.
	defaultWaitTimeout = 60 * time.Second
)

// errHelpRequested marks help output as a successful early return.
var errHelpRequested = errors.New("help requested")

// commandSpec describes one supported linactl command.
type commandSpec struct {
	Name        string
	Description string
	Usage       string
	Run         func(context.Context, *app, commandInput) error
}

// commandInput stores parsed command arguments.
type commandInput struct {
	Args   []string
	Params map[string]string
}

// app stores one linactl invocation's process dependencies and repository paths.
type app struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader

	root string
	env  []string

	execCommand func(context.Context, string, ...string) *exec.Cmd
	waitHTTP    func(string, string, string, string, time.Duration) error
}

// rootConfig stores repository-level tool configuration from hack/config.yaml.
type rootConfig struct {
	Build buildConfig `yaml:"build"`
	Image imageConfig `yaml:"image"`
}

// buildConfig stores user-facing build defaults.
type buildConfig struct {
	Platforms  []string `yaml:"platforms"`
	CGOEnabled bool     `yaml:"cgoEnabled"`
	OutputDir  string   `yaml:"outputDir"`
	BinaryName string   `yaml:"binaryName"`
}

// imageConfig stores user-facing image metadata defaults.
type imageConfig struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	Registry   string `yaml:"registry"`
	Push       bool   `yaml:"push"`
	BaseImage  string `yaml:"baseImage"`
	Dockerfile string `yaml:"dockerfile"`
}

// targetPlatform stores one normalized Go target platform.
type targetPlatform struct {
	OS   string
	Arch string
}

// serviceConfig stores development service paths and ports.
type serviceConfig struct {
	Name      string
	URL       string
	Port      int
	PIDPath   string
	LogPath   string
	WorkDir   string
	StartName string
	StartArgs []string
}

// serviceStatusRow stores one printable development service status row.
type serviceStatusRow struct {
	Service string
	Status  string
	URL     string
	PID     string
	PIDFile string
	LogFile string
}

// pluginManifest stores the plugin fields needed by linactl.
type pluginManifest struct {
	Type string `yaml:"type"`
}

func main() {
	application := newApp(os.Stdout, os.Stderr, os.Stdin)
	if err := application.run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, errHelpRequested) {
			return
		}
		fmt.Fprintf(application.stderr, "linactl: %v\n", err)
		os.Exit(1)
	}
}

// newApp creates a command application with default process dependencies.
func newApp(stdout io.Writer, stderr io.Writer, stdin io.Reader) *app {
	return &app{
		stdout:      stdout,
		stderr:      stderr,
		stdin:       stdin,
		env:         os.Environ(),
		execCommand: exec.CommandContext,
		waitHTTP:    waitHTTP,
	}
}

// run parses the command and dispatches to the command handler.
func (a *app) run(ctx context.Context, args []string) error {
	repoRoot, err := discoverRepoRoot()
	if err != nil {
		return err
	}
	a.root = repoRoot

	if len(args) == 0 {
		return a.printHelp()
	}

	name := normalizeCommandName(args[0])
	if name == "help" {
		if len(args) > 1 {
			name = normalizeCommandName(args[1])
			if spec, ok := commandRegistry()[name]; ok {
				printCommandHelp(a.stdout, spec)
				return nil
			}
			return fmt.Errorf("unknown command %q", args[1])
		}
		return a.printHelp()
	}
	if name == "-h" || name == "--help" {
		return a.printHelp()
	}

	spec, ok := commandRegistry()[name]
	if !ok {
		return fmt.Errorf("unknown command %q; run linactl help", args[0])
	}

	input, err := parseCommandInput(args[1:])
	if err != nil {
		return err
	}
	if input.HasBool("help") || input.HasBool("h") {
		printCommandHelp(a.stdout, spec)
		return errHelpRequested
	}
	return spec.Run(ctx, a, input)
}

// commandRegistry returns the supported command list keyed by command name.
func commandRegistry() map[string]commandSpec {
	specs := []commandSpec{
		{Name: "help", Description: "Show available cross-platform commands.", Usage: "linactl help [command]", Run: runHelp},
		{Name: "dev", Description: "Restart backend and frontend development services.", Usage: "linactl dev [backend_port=8080] [frontend_port=5666] [skip_wasm=true]", Run: runDev},
		{Name: "stop", Description: "Stop backend and frontend development services started by linactl.", Usage: "linactl stop [backend_port=8080] [frontend_port=5666]", Run: runStop},
		{Name: "status", Description: "Show backend and frontend service status.", Usage: "linactl status [backend_port=8080] [frontend_port=5666]", Run: runStatus},
		{Name: "prepare-packed-assets", Description: "Prepare host manifest assets for embedding.", Usage: "linactl prepare-packed-assets", Run: runPreparePackedAssets},
		{Name: "wasm", Description: "Build dynamic Wasm plugin artifacts.", Usage: "linactl wasm [p=<plugin-id>] [out=temp/output] [dry_run=true]", Run: runWasm},
		{Name: "build", Description: "Build frontend assets, plugin artifacts, and host binaries.", Usage: "linactl build [platforms=linux/amd64] [verbose=1]", Run: runBuild},
		{Name: "image", Description: "Build the production Docker image using existing image-builder.", Usage: "linactl image [tag=v0.6.0] [push=1]", Run: runImage},
		{Name: "image-build", Description: "Stage image build artifacts without invoking Docker build.", Usage: "linactl image-build [tag=v0.6.0]", Run: runImageBuild},
		{Name: "init", Description: "Initialize the database with DDL and seed data.", Usage: "linactl init confirm=init [rebuild=true]", Run: runInit},
		{Name: "mock", Description: "Load optional mock demo data.", Usage: "linactl mock confirm=mock", Run: runMock},
		{Name: "test", Description: "Run the Playwright E2E test suite.", Usage: "linactl test", Run: runTest},
		{Name: "test-go", Description: "Run Go unit tests for workspace modules.", Usage: "linactl test-go [race=true] [verbose=true]", Run: runTestGo},
		{Name: "test-scripts", Description: "Run repository tool smoke tests.", Usage: "linactl test-scripts", Run: runTestScripts},
		{Name: "check-runtime-i18n", Description: "Scan runtime-visible code for hard-coded text.", Usage: "linactl check-runtime-i18n", Run: runCheckRuntimeI18n},
		{Name: "check-runtime-i18n-messages", Description: "Validate runtime i18n message key coverage.", Usage: "linactl check-runtime-i18n-messages", Run: runCheckRuntimeI18nMessages},
		{Name: "cli", Description: "Install or update the GoFrame CLI.", Usage: "linactl cli", Run: runCLIInstall},
		{Name: "cli.install", Description: "Install the GoFrame CLI only when missing.", Usage: "linactl cli.install", Run: runCLIInstallIfMissing},
		{Name: "ctrl", Description: "Generate GoFrame controllers.", Usage: "linactl ctrl", Run: runGF("gen", "ctrl")},
		{Name: "dao", Description: "Generate GoFrame DAO/DO/Entity files.", Usage: "linactl dao", Run: runGF("gen", "dao")},
		{Name: "enums", Description: "Generate GoFrame enum files.", Usage: "linactl enums", Run: runGF("gen", "enums")},
		{Name: "service", Description: "Generate GoFrame service files.", Usage: "linactl service", Run: runGF("gen", "service")},
		{Name: "pb", Description: "Generate protobuf files.", Usage: "linactl pb", Run: runGF("gen", "pb")},
		{Name: "pbentity", Description: "Generate protobuf entity files.", Usage: "linactl pbentity", Run: runGF("gen", "pbentity")},
		{Name: "deploy", Description: "Apply kustomize manifests to the current kubectl context.", Usage: "linactl deploy [env=<overlay>] [tag=develop]", Run: runDeploy},
	}

	registry := make(map[string]commandSpec, len(specs))
	for _, spec := range specs {
		registry[spec.Name] = spec
	}
	return registry
}

// normalizeCommandName converts historical make target aliases to linactl command names.
func normalizeCommandName(name string) string {
	switch strings.TrimSpace(name) {
	case "prepare":
		return "prepare-packed-assets"
	default:
		return strings.TrimSpace(name)
	}
}

// parseCommandInput accepts make-style key=value parameters and standard flags.
func parseCommandInput(args []string) (commandInput, error) {
	input := commandInput{Params: map[string]string{}}
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if strings.HasPrefix(arg, "--") {
			trimmed := strings.TrimPrefix(arg, "--")
			if trimmed == "" {
				return input, fmt.Errorf("invalid empty flag")
			}
			key, value, ok := strings.Cut(trimmed, "=")
			key = normalizeParamKey(key)
			if !ok {
				input.Params[key] = "true"
				continue
			}
			if key == "" {
				return input, fmt.Errorf("invalid flag %q", arg)
			}
			input.Params[key] = value
			continue
		}
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			input.Params[normalizeParamKey(strings.TrimPrefix(arg, "-"))] = "true"
			continue
		}
		if key, value, ok := strings.Cut(arg, "="); ok {
			key = normalizeParamKey(key)
			if key == "" {
				return input, fmt.Errorf("invalid parameter %q", arg)
			}
			input.Params[key] = value
			continue
		}
		input.Args = append(input.Args, arg)
	}
	return input, nil
}

// Get returns a parsed parameter value.
func (i commandInput) Get(key string) string {
	return i.Params[normalizeParamKey(key)]
}

// GetDefault returns a parameter value or the provided default.
func (i commandInput) GetDefault(key string, fallback string) string {
	if value, ok := i.Params[normalizeParamKey(key)]; ok && value != "" {
		return value
	}
	return fallback
}

// HasBool reports whether a flag-style boolean parameter is true.
func (i commandInput) HasBool(key string) bool {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok {
		return false
	}
	parsed, err := parseBool(value, false)
	if err != nil {
		return false
	}
	return parsed
}

// Bool returns a parsed boolean parameter.
func (i commandInput) Bool(key string, fallback bool) (bool, error) {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok {
		return fallback, nil
	}
	return parseBool(value, fallback)
}

// runHelp prints the top-level help output.
func runHelp(_ context.Context, a *app, _ commandInput) error {
	return a.printHelp()
}

// printHelp writes the command overview and platform-specific entry examples.
func (a *app) printHelp() error {
	specs := commandRegistry()
	names := make([]string, 0, len(specs))
	for name := range specs {
		names = append(names, name)
	}
	sort.Strings(names)

	maxName := 0
	for _, name := range names {
		if len(name) > maxName {
			maxName = len(name)
		}
	}

	fmt.Fprintln(a.stdout, "Usage: linactl <command> [key=value] [--flag=value]")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Windows:")
	fmt.Fprintln(a.stdout, "  cmd.exe:     make.cmd help")
	fmt.Fprintln(a.stdout, "  PowerShell:  .\\make.cmd help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Linux/macOS:")
	fmt.Fprintln(a.stdout, "  make help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	for _, name := range names {
		spec := specs[name]
		fmt.Fprintf(a.stdout, "  %-*s  %s\n", maxName, spec.Name, spec.Description)
	}
	return nil
}

// printCommandHelp writes usage for one command.
func printCommandHelp(out io.Writer, spec commandSpec) {
	fmt.Fprintf(out, "Usage: %s\n\n%s\n", spec.Usage, spec.Description)
}

// runPreparePackedAssets rebuilds the host manifest embed workspace.
func runPreparePackedAssets(_ context.Context, a *app, _ commandInput) error {
	sourceDir := filepath.Join(a.root, "apps", "lina-core", "manifest")
	targetDir := filepath.Join(a.root, "apps", "lina-core", "internal", "packed", "manifest")

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("clean packed manifest directory: %w", err)
	}

	dirs := []string{
		filepath.Join(targetDir, "config"),
		filepath.Join(targetDir, "sql"),
		filepath.Join(targetDir, "i18n"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(sourceDir, "config", "config.template.yaml"): filepath.Join(targetDir, "config", "config.template.yaml"),
		filepath.Join(sourceDir, "config", "metadata.yaml"):        filepath.Join(targetDir, "config", "metadata.yaml"),
	}
	for src, dst := range files {
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}

	if err := copyDirContents(filepath.Join(sourceDir, "sql"), filepath.Join(targetDir, "sql")); err != nil {
		return err
	}
	if err := copyDirContents(filepath.Join(sourceDir, "i18n"), filepath.Join(targetDir, "i18n")); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, ".gitkeep"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("write packed manifest .gitkeep: %w", err)
	}

	fmt.Fprintf(a.stdout, "packed manifest assets prepared: %s\n", relativePath(a.root, targetDir))
	return nil
}

// runWasm builds dynamic Wasm plugin artifacts or lists them in dry-run mode.
func runWasm(ctx context.Context, a *app, input commandInput) error {
	outDir := input.Get("out")
	if outDir == "" {
		outDir = filepath.Join(a.root, "temp", "output")
	}
	if !filepath.IsAbs(outDir) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve current directory for wasm output: %w", err)
		}
		outDir = filepath.Join(cwd, outDir)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create wasm output directory: %w", err)
	}

	plugins, err := dynamicPlugins(a.root, input.Get("p"))
	if err != nil {
		return err
	}
	if len(plugins) == 0 {
		fmt.Fprintln(a.stdout, "No buildable dynamic wasm plugins found")
		return nil
	}

	dryRun, err := input.Bool("dry_run", false)
	if err != nil {
		return err
	}
	if !dryRun {
		dryRun, err = input.Bool("dry-run", false)
		if err != nil {
			return err
		}
	}
	for _, plugin := range plugins {
		fmt.Fprintf(a.stdout, "Building dynamic wasm plugin: %s\n", plugin)
		if dryRun {
			continue
		}
		err = a.runCommand(ctx, commandOptions{
			Dir: filepath.Join(a.root, "hack", "tools", "build-wasm"),
		}, "go", "run", ".", "--plugin-dir", filepath.Join(a.root, "apps", "lina-plugins", plugin), "--output-dir", outDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// runStatus prints development service status using cross-platform checks.
func runStatus(_ context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}
	services := a.services(backendPort, frontendPort)

	if _, err = fmt.Fprintln(a.stdout, ""); err != nil {
		return fmt.Errorf("write status output: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "LinaPro Framework Status"); err != nil {
		return fmt.Errorf("write status title: %w", err)
	}

	rows := make([]serviceStatusRow, 0, len(services))
	for _, service := range services {
		status := "stopped"
		if isTCPListening(service.Port) || serviceReady(service.URL, 2*time.Second) {
			status = "running"
		}
		pid := readPID(service.PIDPath)
		pidText := "-"
		if pid > 0 {
			pidText = strconv.Itoa(pid)
		}
		rows = append(rows, serviceStatusRow{
			Service: service.Name,
			Status:  status,
			URL:     service.URL,
			PID:     pidText,
			PIDFile: relativePath(a.root, service.PIDPath),
			LogFile: relativePath(a.root, service.LogPath),
		})
	}
	if err = printStatusTable(a.stdout, rows); err != nil {
		return err
	}
	return nil
}

// runStop stops services that were started by linactl.
func runStop(_ context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintln(a.stdout, "Stopping services..."); err != nil {
		return fmt.Errorf("write stop output: %w", err)
	}
	for _, service := range a.services(backendPort, frontendPort) {
		if err = stopService(a.stdout, service); err != nil {
			return err
		}
	}
	return nil
}

// runDev builds and starts backend and frontend development services.
func runDev(ctx context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}
	skipWasm, err := input.Bool("skip_wasm", false)
	if err != nil {
		return err
	}

	stopInput := commandInput{Params: map[string]string{
		"backend_port":  strconv.Itoa(backendPort),
		"frontend_port": strconv.Itoa(frontendPort),
	}}
	if err = runStop(ctx, a, stopInput); err != nil {
		return err
	}

	tempDir := filepath.Join(a.root, "temp")
	binDir := filepath.Join(tempDir, "bin")
	if err = os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create temp bin directory: %w", err)
	}

	if !skipWasm {
		if err = runWasm(ctx, a, commandInput{Params: map[string]string{"out": filepath.Join("temp", "output")}}); err != nil {
			return err
		}
	}
	if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
		return err
	}

	backendBinary := filepath.Join(binDir, executableName("lina"))
	if err = os.Remove(backendBinary); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove existing backend binary: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "Building backend..."); err != nil {
		return fmt.Errorf("write build output: %w", err)
	}
	if err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core")}, "go", "build", "-o", backendBinary, "."); err != nil {
		return err
	}

	services := a.services(backendPort, frontendPort)
	services[0].StartName = backendBinary
	services[1].StartName = viteCommand(a.root)
	for _, service := range services {
		if err = startService(a, service); err != nil {
			return err
		}
	}

	for _, service := range services {
		if err = a.waitHTTP(service.Name, service.URL, service.PIDPath, service.LogPath, defaultWaitTimeout); err != nil {
			return err
		}
		if _, err = fmt.Fprintf(a.stdout, "%s is ready: %s\n", service.Name, service.URL); err != nil {
			return fmt.Errorf("write readiness output: %w", err)
		}
	}

	return runStatus(ctx, a, stopInput)
}

// runBuild builds frontend assets, plugin artifacts, and host binaries.
func runBuild(ctx context.Context, a *app, input commandInput) error {
	cfg, err := loadRootConfig(a.root, input)
	if err != nil {
		return err
	}
	targets, err := normalizePlatforms(cfg.Build.Platforms, input.Get("platforms"))
	if err != nil {
		return err
	}
	verbose, err := input.Bool("verbose", false)
	if err != nil {
		return err
	}
	if !verbose {
		verbose, err = input.Bool("v", false)
		if err != nil {
			return err
		}
	}

	outputDir := input.GetDefault("output_dir", cfg.Build.OutputDir)
	if outputDir == "" {
		outputDir = "temp/output"
	}
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(a.root, outputDir)
	}
	binaryName := input.GetDefault("binary_name", cfg.Build.BinaryName)
	if binaryName == "" {
		binaryName = "lina"
	}
	cgoEnabled := "0"
	if cfg.Build.CGOEnabled {
		cgoEnabled = "1"
	}
	if raw := input.Get("cgo_enabled"); raw != "" {
		enabled, parseErr := parseBool(raw, false)
		if parseErr != nil {
			return parseErr
		}
		if enabled {
			cgoEnabled = "1"
		} else {
			cgoEnabled = "0"
		}
	}

	if err = os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("clean build output directory: %w", err)
	}
	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create build output directory: %w", err)
	}

	fmt.Fprintln(a.stdout, "Building frontend...")
	if err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-vben"), Quiet: !verbose}, "pnpm", "run", "build"); err != nil {
		return err
	}

	embedDir := filepath.Join(a.root, "apps", "lina-core", "internal", "packed", "public")
	if err = os.RemoveAll(embedDir); err != nil {
		return fmt.Errorf("clean frontend embed directory: %w", err)
	}
	if err = os.MkdirAll(embedDir, 0o755); err != nil {
		return fmt.Errorf("create frontend embed directory: %w", err)
	}
	if err = copyDirContents(filepath.Join(a.root, "apps", "lina-vben", "apps", "web-antd", "dist"), embedDir); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Host frontend embedded assets generated")

	if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
		return err
	}
	if err = runWasm(ctx, a, commandInput{Params: map[string]string{"out": outputDir}}); err != nil {
		return err
	}

	multiPlatform := len(targets) > 1
	for _, target := range targets {
		targetBinary := filepath.Join(outputDir, executableName(binaryName))
		if multiPlatform {
			targetBinary = filepath.Join(outputDir, target.OS+"_"+target.Arch, executableName(binaryName))
		}
		if err = os.MkdirAll(filepath.Dir(targetBinary), 0o755); err != nil {
			return fmt.Errorf("create backend output directory: %w", err)
		}
		fmt.Fprintf(a.stdout, "Building backend for %s/%s...\n", target.OS, target.Arch)
		env := append(a.env, "CGO_ENABLED="+cgoEnabled, "GOOS="+target.OS, "GOARCH="+target.Arch)
		err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core"), Env: env, Quiet: !verbose}, "go", "build", "-o", targetBinary, ".")
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Build complete: %s\n", relativePath(a.root, targetBinary))
	}
	return nil
}

// runImage builds and optionally pushes a production Docker image.
func runImage(ctx context.Context, a *app, input commandInput) error {
	if err := runImageBuilder(ctx, a, input, "--preflight"); err != nil {
		return err
	}
	if err := runBuild(ctx, a, input); err != nil {
		return err
	}
	return runImageBuilder(ctx, a, input)
}

// runImageBuild stages image build artifacts without running docker build.
func runImageBuild(ctx context.Context, a *app, input commandInput) error {
	if err := runBuild(ctx, a, input); err != nil {
		return err
	}
	return runImageBuilder(ctx, a, input, "--build-only")
}

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

// runTest starts the Playwright E2E test suite.
func runTest(ctx context.Context, a *app, _ commandInput) error {
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tests")}, "pnpm", "test")
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

	modules, err := goWorkspaceModules(ctx, a)
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
		if err = a.runCommand(ctx, commandOptions{Dir: moduleDir}, "go", args...); err != nil {
			return err
		}
	}
	return nil
}

// runTestScripts runs legacy shell smoke checks on POSIX systems.
func runTestScripts(ctx context.Context, a *app, _ commandInput) error {
	if runtime.GOOS == "windows" {
		return errors.New("test-scripts requires POSIX shell scripts; use Go tests on Windows")
	}
	scripts, err := filepath.Glob(filepath.Join(a.root, "hack", "tests", "scripts", "*.sh"))
	if err != nil {
		return err
	}
	sort.Strings(scripts)
	for _, script := range scripts {
		fmt.Fprintf(a.stdout, "==> %s\n", relativePath(a.root, script))
		if err = a.runCommand(ctx, commandOptions{Dir: a.root}, "bash", script); err != nil {
			return err
		}
	}
	return nil
}

// runCheckRuntimeI18n invokes the runtime i18n hard-coded text scanner.
func runCheckRuntimeI18n(ctx context.Context, a *app, _ commandInput) error {
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tools", "runtime-i18n")}, "go", "run", ".", "scan")
}

// runCheckRuntimeI18nMessages invokes runtime i18n message coverage validation.
func runCheckRuntimeI18nMessages(ctx context.Context, a *app, _ commandInput) error {
	return a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "hack", "tools", "runtime-i18n")}, "go", "run", ".", "messages")
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

// runDeploy applies a kustomize overlay through kubectl.
func runDeploy(ctx context.Context, a *app, input commandInput) error {
	envName := firstNonEmpty(input.Get("env"), input.Get("_ENV"))
	if envName == "" {
		return errors.New("deploy requires env=<overlay>")
	}
	tag := firstNonEmpty(input.Get("tag"), input.Get("TAG"), "develop")
	namespace := input.GetDefault("namespace", "default")
	deployName := input.GetDefault("deploy_name", "template-single")

	tempDir := filepath.Join(a.root, "apps", "lina-core", "temp", "kustomize")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return fmt.Errorf("create kustomize temp directory: %w", err)
	}
	overlayDir := filepath.Join(a.root, "apps", "lina-core", "manifest", "deploy", "kustomize", "overlays", envName)
	outputPath := filepath.Join(a.root, "apps", "lina-core", "temp", "kustomize.yaml")

	var manifest bytes.Buffer
	cmd := a.execCommand(ctx, "kustomize", "build")
	cmd.Dir = overlayDir
	cmd.Env = a.env
	cmd.Stdout = &manifest
	cmd.Stderr = a.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run kustomize build: %w", err)
	}
	if err := os.WriteFile(outputPath, manifest.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write kustomize output: %w", err)
	}
	if err := a.runCommand(ctx, commandOptions{Dir: a.root}, "kubectl", "apply", "-f", outputPath); err != nil {
		return err
	}
	if deployName == "" {
		return nil
	}
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"labels":{"date":"%d","tag":"%s"}}}}}`, time.Now().Unix(), tag)
	return a.runCommand(ctx, commandOptions{Dir: a.root}, "kubectl", "patch", "-n", namespace, "deployment/"+deployName, "-p", patch)
}

// Int returns a parsed integer parameter.
func (i commandInput) Int(key string, fallback int) (int, error) {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

type commandOptions struct {
	// Dir sets the child process working directory.
	Dir string
	// Env overrides the child process environment.
	Env []string
	// Quiet buffers child output unless the command fails.
	Quiet bool
	// Stdout overrides stdout forwarding.
	Stdout io.Writer
	// Stderr overrides stderr forwarding.
	Stderr io.Writer
}

// runCommand executes a child command with consistent error messages.
func (a *app) runCommand(ctx context.Context, options commandOptions, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil && !filepath.IsAbs(name) {
		return fmt.Errorf("required tool %q is not available in PATH while running %s: %w", name, strings.Join(append([]string{name}, args...), " "), err)
	}

	cmd := a.execCommand(ctx, name, args...)
	if options.Dir != "" {
		cmd.Dir = options.Dir
	}
	if len(options.Env) > 0 {
		cmd.Env = options.Env
	} else {
		cmd.Env = a.env
	}
	cmd.Stdin = a.stdin

	stdout := options.Stdout
	stderr := options.Stderr
	if stdout == nil {
		stdout = a.stdout
	}
	if stderr == nil {
		stderr = a.stderr
	}
	if options.Quiet {
		var buffer bytes.Buffer
		cmd.Stdout = &buffer
		cmd.Stderr = &buffer
		err := cmd.Run()
		if err != nil {
			fmt.Fprint(stderr, buffer.String())
			return fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, args...), " "), err)
		}
		return nil
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}
	return nil
}

// discoverRepoRoot searches upward for the LinaPro repository root.
func discoverRepoRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(current, "go.work")) && dirExists(filepath.Join(current, "apps", "lina-core")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", errors.New("cannot find LinaPro repository root")
}

// copyFile copies one regular file and creates the destination parent directory.
func copyFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", src, closeErr)
		}
	}()

	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", dst, err)
	}
	output, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if closeErr := output.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", dst, closeErr)
		}
	}()

	if _, err = io.Copy(output, input); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	return nil
}

// copyDirContents recursively copies the contents of one directory.
func copyDirContents(src string, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err = copyDirContents(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, readErr := os.Readlink(srcPath)
			if readErr != nil {
				return fmt.Errorf("read symlink %s: %w", srcPath, readErr)
			}
			if err = os.Symlink(target, dstPath); err != nil {
				return fmt.Errorf("create symlink %s: %w", dstPath, err)
			}
			continue
		}
		if err = copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// dynamicPlugins returns dynamic plugin IDs, optionally validating one plugin.
func dynamicPlugins(root string, plugin string) ([]string, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	if plugin != "" {
		if err := validateDynamicPlugin(pluginRoot, plugin); err != nil {
			return nil, err
		}
		return []string{plugin}, nil
	}

	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		return nil, fmt.Errorf("read plugin directory: %w", err)
	}

	var plugins []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest := filepath.Join(pluginRoot, entry.Name(), "plugin.yaml")
		if !fileExists(manifest) {
			continue
		}
		isDynamic, err := isDynamicPlugin(manifest)
		if err != nil {
			return nil, err
		}
		if isDynamic {
			plugins = append(plugins, entry.Name())
		}
	}
	sort.Strings(plugins)
	return plugins, nil
}

// validateDynamicPlugin verifies that a plugin exists and is dynamic.
func validateDynamicPlugin(pluginRoot string, plugin string) error {
	manifest := filepath.Join(pluginRoot, plugin, "plugin.yaml")
	if !fileExists(manifest) {
		return fmt.Errorf("plugin does not exist: %s", plugin)
	}
	dynamic, err := isDynamicPlugin(manifest)
	if err != nil {
		return err
	}
	if !dynamic {
		return fmt.Errorf("plugin is not dynamic and cannot be built as wasm: %s", plugin)
	}
	return nil
}

// isDynamicPlugin reports whether a plugin manifest declares dynamic type.
func isDynamicPlugin(manifestPath string) (bool, error) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return false, fmt.Errorf("read plugin manifest %s: %w", manifestPath, err)
	}
	var manifest pluginManifest
	if err = yaml.Unmarshal(content, &manifest); err != nil {
		return false, fmt.Errorf("parse plugin manifest %s: %w", manifestPath, err)
	}
	return strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic"), nil
}

// loadRootConfig loads repository tool defaults from hack/config.yaml.
func loadRootConfig(root string, input commandInput) (rootConfig, error) {
	cfg := rootConfig{
		Build: buildConfig{
			Platforms:  []string{"auto"},
			CGOEnabled: false,
			OutputDir:  filepath.Join("temp", "output"),
			BinaryName: "lina",
		},
		Image: imageConfig{
			Name:       "linapro",
			BaseImage:  "alpine:3.22",
			Dockerfile: filepath.Join("hack", "docker", "Dockerfile"),
		},
	}

	configPath := input.GetDefault("config", filepath.Join("hack", "config.yaml"))
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", configPath, err)
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", configPath, err)
	}
	return cfg, nil
}

// normalizePlatforms parses build platform strings into Go target tuples.
func normalizePlatforms(defaults []string, override string) ([]targetPlatform, error) {
	raw := defaults
	if override != "" {
		raw = strings.Split(override, ",")
	}
	if len(raw) == 0 {
		raw = []string{"auto"}
	}

	targets := make([]targetPlatform, 0, len(raw))
	for _, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value == "auto" {
			value = "linux/" + runtime.GOARCH
		}
		goos, goarch, ok := strings.Cut(value, "/")
		if !ok || goos == "" || goarch == "" {
			return nil, fmt.Errorf("invalid platform %q; expected goos/goarch", value)
		}
		targets = append(targets, targetPlatform{OS: goos, Arch: goarch})
	}
	if len(targets) == 0 {
		return nil, errors.New("no build platforms configured")
	}
	return targets, nil
}

// runImageBuilder invokes the existing image-builder tool with shared arguments.
func runImageBuilder(ctx context.Context, a *app, input commandInput, extra ...string) error {
	args := []string{"run", "./hack/tools/image-builder"}
	args = append(args, extra...)
	for _, item := range imageBuilderArgs(input) {
		args = append(args, item)
	}
	return a.runCommand(ctx, commandOptions{Dir: a.root}, "go", args...)
}

// imageBuilderArgs maps linactl parameters to image-builder flags.
func imageBuilderArgs(input commandInput) []string {
	known := []string{"image", "tag", "registry", "push", "platforms", "cgo_enabled", "output_dir", "binary_name", "base_image", "config", "verbose", "v"}
	var args []string
	for _, key := range known {
		value, exists := input.Params[key]
		if !exists || value == "" {
			continue
		}
		flagName := strings.ReplaceAll(key, "_", "-")
		if key == "v" {
			flagName = "verbose"
		}
		args = append(args, "--"+flagName+"="+value)
	}
	return args
}

// goWorkspaceModules lists module directories from the current Go workspace.
func goWorkspaceModules(ctx context.Context, a *app) ([]string, error) {
	cmd := a.execCommand(ctx, "go", "list", "-m", "-f", "{{.Dir}}")
	cmd.Dir = a.root
	cmd.Env = a.env
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list Go workspace modules: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var modules []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			modules = append(modules, line)
		}
	}
	return modules, nil
}

// services returns backend and frontend development service definitions.
func (a *app) services(backendPort int, frontendPort int) []serviceConfig {
	tempDir := filepath.Join(a.root, "temp")
	pidDir := filepath.Join(tempDir, "pids")
	return []serviceConfig{
		{
			Name:    "Backend",
			URL:     fmt.Sprintf("http://127.0.0.1:%d/", backendPort),
			Port:    backendPort,
			PIDPath: filepath.Join(pidDir, "backend.pid"),
			LogPath: filepath.Join(tempDir, "lina-core.log"),
			WorkDir: filepath.Join(a.root, "apps", "lina-core"),
		},
		{
			Name:      "Frontend",
			URL:       fmt.Sprintf("http://127.0.0.1:%d/", frontendPort),
			Port:      frontendPort,
			PIDPath:   filepath.Join(pidDir, "frontend.pid"),
			LogPath:   filepath.Join(tempDir, "lina-vben.log"),
			WorkDir:   filepath.Join(a.root, "apps", "lina-vben", "apps", "web-antd"),
			StartArgs: []string{"--mode", "development", "--host", "127.0.0.1", "--port", strconv.Itoa(frontendPort), "--strictPort"},
		},
	}
}

// printStatusTable renders development service status without terminal-specific dependencies.
func printStatusTable(out io.Writer, rows []serviceStatusRow) error {
	headers := []string{"Service", "Status", "URL", "PID", "PID File", "Log File"}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		values := row.values()
		for i, value := range values {
			if len(value) > widths[i] {
				widths[i] = len(value)
			}
		}
	}

	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	if err := printTableRow(out, widths, headers); err != nil {
		return err
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := printTableRow(out, widths, row.values()); err != nil {
			return err
		}
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	return nil
}

// values returns the printable table cells for one service status row.
func (r serviceStatusRow) values() []string {
	return []string{r.Service, r.Status, r.URL, r.PID, r.PIDFile, r.LogFile}
}

// printTableBorder prints one ASCII border line for a table.
func printTableBorder(out io.Writer, widths []int) error {
	if _, err := fmt.Fprint(out, "+"); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	for _, width := range widths {
		if _, err := fmt.Fprint(out, strings.Repeat("-", width+2)); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
		if _, err := fmt.Fprint(out, "+"); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	return nil
}

// printTableRow prints one padded ASCII table row.
func printTableRow(out io.Writer, widths []int, values []string) error {
	if _, err := fmt.Fprint(out, "|"); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	for i, value := range values {
		if _, err := fmt.Fprintf(out, " %-*s |", widths[i], value); err != nil {
			return fmt.Errorf("write table row: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	return nil
}

// startService starts a development service and records its PID file.
func startService(a *app, service serviceConfig) error {
	if err := os.MkdirAll(filepath.Dir(service.PIDPath), 0o755); err != nil {
		return fmt.Errorf("create PID directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(service.LogPath), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}
	logFile, err := os.OpenFile(service.LogPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", service.LogPath, err)
	}

	cmd := a.execCommand(context.Background(), service.StartName, service.StartArgs...)
	cmd.Dir = service.WorkDir
	cmd.Env = a.env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	configureDetachedProcess(cmd)
	if err = cmd.Start(); err != nil {
		if closeErr := logFile.Close(); closeErr != nil {
			return fmt.Errorf("start %s failed and close log failed: %v: %w", service.Name, closeErr, err)
		}
		return fmt.Errorf("start %s: %w", service.Name, err)
	}
	pid := cmd.Process.Pid
	if err = os.WriteFile(service.PIDPath, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return fmt.Errorf("write %s PID file: %w", service.Name, err)
	}
	if err = logFile.Close(); err != nil {
		return fmt.Errorf("close %s log file: %w", service.Name, err)
	}
	if err = cmd.Process.Release(); err != nil {
		return fmt.Errorf("release %s process: %w", service.Name, err)
	}
	fmt.Fprintf(a.stdout, "%s started: pid=%d log=%s\n", service.Name, pid, relativePath(a.root, service.LogPath))
	return nil
}

// stopService stops a PID-file-backed service when possible.
func stopService(out io.Writer, service serviceConfig) error {
	pid := readPID(service.PIDPath)
	stopped := false
	if pid > 0 {
		process, err := os.FindProcess(pid)
		if err == nil {
			if killErr := process.Kill(); killErr == nil {
				stopped = true
			}
		}
	}
	if err := os.Remove(service.PIDPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s PID file: %w", service.Name, err)
	}
	if stopped {
		fmt.Fprintf(out, "%s stopped\n", service.Name)
		return nil
	}
	fmt.Fprintf(out, "%s is not running\n", service.Name)
	return nil
}

// waitHTTP waits for one service URL to become ready.
func waitHTTP(name string, url string, pidPath string, logPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := newReadinessHTTPClient(2 * time.Second)
	for time.Now().Before(deadline) {
		if readPID(pidPath) == 0 {
			return fmt.Errorf("%s startup failed: PID file does not exist; check log: %s", name, logPath)
		}
		resp, err := client.Get(url)
		if err == nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return fmt.Errorf("close %s readiness response: %w", name, closeErr)
			}
			if resp.StatusCode < http.StatusInternalServerError {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("%s startup timed out (%s): %s; check log: %s", name, timeout, url, logPath)
}

// serviceReady reports whether an HTTP endpoint responds without server error.
func serviceReady(url string, timeout time.Duration) bool {
	client := newReadinessHTTPClient(timeout)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		return false
	}
	return resp.StatusCode < http.StatusInternalServerError
}

// newReadinessHTTPClient matches curl-style readiness by accepting redirects as responses.
func newReadinessHTTPClient(timeout time.Duration) http.Client {
	return http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// isTCPListening reports whether localhost accepts TCP connections on a port.
func isTCPListening(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		return false
	}
	return true
}

// readPID reads and validates a PID file.
func readPID(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	text := strings.TrimSpace(string(content))
	pid, err := strconv.Atoi(text)
	if err != nil || pid <= 1 {
		return 0
	}
	return pid
}

// parseBool parses command-line boolean forms accepted by linactl.
func parseBool(value string, _ bool) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
}

// isConnectionFailure detects common database connection failure messages.
func isConnectionFailure(text string) bool {
	patterns := []string{"dial tcp", "connection refused", "connect: connection", "failed to connect", "i/o timeout", "no such host"}
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// downloadFile downloads a URL to a local file.
func downloadFile(ctx context.Context, url string, dst string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close download response: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s returned %s", url, resp.Status)
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", dst, closeErr)
		}
	}()
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

// executableName returns the platform-specific executable filename.
func executableName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		return name + ".exe"
	}
	return name
}

// viteCommand returns the platform-specific Vite binary path.
func viteCommand(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite.cmd")
	}
	return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite")
}

// relativePath renders a path relative to the repository root when possible.
func relativePath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// firstNonEmpty returns the first non-empty value.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// normalizeParamKey keeps make-style and CLI-style option keys equivalent.
func normalizeParamKey(key string) string {
	return strings.ReplaceAll(strings.TrimSpace(key), "-", "_")
}

// fileExists reports whether a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists reports whether a path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// init silences the default flag package output for this custom parser.
func init() {
	flag.CommandLine.SetOutput(io.Discard)
}
