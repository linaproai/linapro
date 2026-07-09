// This file implements the build command for frontend assets and host binaries.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"linactl/internal/config"
	"linactl/internal/fileutil"
	"linactl/internal/plugins"
	"linactl/internal/toolutil"
)

// packedPublicPlaceholderName is tracked so Go embed patterns keep compiling
// before generated frontend build assets exist in a clean checkout.
const packedPublicPlaceholderName = ".gitkeep"

type buildOptions struct {
	Targets    []targetPlatform
	OutputDir  string
	BinaryName string
	CGOEnabled string
	Verbose    bool
}

// pluginBuildStep is one command line declared by a plugin hack/config.yaml.
type pluginBuildStep struct {
	Command string
	Args    []string
}

type pluginBuildConfig struct {
	Build struct {
		Commands []string `yaml:"commands"`
	} `yaml:"build"`
}

// runBuild builds frontend assets, plugin artifacts, and host binaries.
func runBuild(ctx context.Context, a *app, input commandInput) error {
	options, err := resolveBuildOptions(a.root, input)
	if err != nil {
		return err
	}
	if dir := strings.TrimSpace(input.Get("dir")); dir != "" {
		return runBuildDir(ctx, a, input, options, dir)
	}

	pluginsEnabled, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(options.OutputDir); err != nil {
		return fmt.Errorf("clean build output directory: %w", err)
	}
	if err = os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create build output directory: %w", err)
	}
	if err = runHostFrontendBuild(ctx, a, env, options.Verbose); err != nil {
		return err
	}
	if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
		return err
	}
	if pluginsEnabled {
		if err = runPluginBuilds(ctx, a, env, options); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(a.stdout, "Skipping official plugin wasm build in host-only mode")
	}

	return runHostBackendBuild(ctx, a, env, options)
}

func resolveBuildOptions(root string, input commandInput) (buildOptions, error) {
	cfg, err := loadRootConfig(root, input)
	if err != nil {
		return buildOptions{}, err
	}
	targets, err := normalizePlatforms(cfg.Build.Platforms, input.Get("platforms"))
	if err != nil {
		return buildOptions{}, err
	}
	verbose, err := input.Bool("verbose", false)
	if err != nil {
		return buildOptions{}, err
	}
	if !verbose {
		verbose, err = input.Bool("v", false)
		if err != nil {
			return buildOptions{}, err
		}
	}

	outputDir := input.GetDefault("output_dir", cfg.Build.OutputDir)
	if outputDir == "" {
		outputDir = "temp/output"
	}
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(root, outputDir)
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
		enabled, parseErr := toolutil.ParseBool(raw, false)
		if parseErr != nil {
			return buildOptions{}, parseErr
		}
		if enabled {
			cgoEnabled = "1"
		} else {
			cgoEnabled = "0"
		}
	}
	return buildOptions{
		Targets:    targets,
		OutputDir:  outputDir,
		BinaryName: binaryName,
		CGOEnabled: cgoEnabled,
		Verbose:    verbose,
	}, nil
}

func runBuildDir(ctx context.Context, a *app, input commandInput, options buildOptions, rawDir string) error {
	dir, err := resolveBuildDir(a.root, rawDir)
	if err != nil {
		return err
	}
	if sameBuildPath(dir, a.root) {
		delete(input.Params, "dir")
		return runBuild(ctx, a, input)
	}
	switch {
	case sameBuildPath(dir, filepath.Join(a.root, "apps", "lina-vben")):
		env := plugins.BuildEnv(a.root, a.env, false, "")
		return runHostFrontendBuild(ctx, a, env, options.Verbose)
	case sameBuildPath(dir, filepath.Join(a.root, "apps", "lina-core")):
		env := plugins.BuildEnv(a.root, a.env, false, "")
		if err = os.MkdirAll(options.OutputDir, 0o755); err != nil {
			return fmt.Errorf("create build output directory: %w", err)
		}
		if err = runHostFrontendBuild(ctx, a, env, options.Verbose); err != nil {
			return err
		}
		if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
			return err
		}
		return runHostBackendBuild(ctx, a, env, options)
	case isOfficialPluginDir(a.root, dir):
		_, env, err := prepareOfficialPluginBuildEnv(ctx, a, commandInput{Params: map[string]string{"plugins": "1"}})
		if err != nil {
			return err
		}
		return runOnePluginBuild(ctx, a, dir, env, options)
	default:
		env := plugins.BuildEnv(a.root, a.env, false, "")
		return runPackageBuildDir(ctx, a, dir, env, options.Verbose)
	}
}

func sameBuildPath(left string, right string) bool {
	leftClean, leftErr := filepath.Abs(filepath.Clean(left))
	rightClean, rightErr := filepath.Abs(filepath.Clean(right))
	if leftErr != nil || rightErr != nil {
		return filepath.Clean(left) == filepath.Clean(right)
	}
	return leftClean == rightClean
}

func isOfficialPluginDir(root string, dir string) bool {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	relative, err := filepath.Rel(pluginRoot, dir)
	if err != nil || relative == "." || relative == "" || strings.HasPrefix(relative, "..") {
		return false
	}
	if strings.Contains(relative, string(filepath.Separator)) {
		return false
	}
	return fileutil.FileExists(filepath.Join(dir, "plugin.yaml"))
}

func runOnePluginBuild(ctx context.Context, a *app, pluginDir string, env []string, options buildOptions) error {
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifest, err := plugins.ReadManifest(manifestPath)
	if err != nil {
		return err
	}
	steps, err := resolvePluginConfigBuildSteps(a.root, pluginDir)
	if err != nil {
		return err
	}
	if len(steps) > 0 {
		fmt.Fprintf(a.stdout, "Building plugin: %s\n", toolutil.RelativePath(a.root, pluginDir))
		if err = runPluginBuildSteps(ctx, a, pluginDir, env, options.Verbose, steps); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(a.stdout, "No plugin build commands configured: %s\n", toolutil.RelativePath(a.root, pluginDir))
	}
	if strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic") {
		return runWasm(ctx, a, commandInput{Params: map[string]string{"plugin_dir": pluginDir, "out": options.OutputDir}})
	}
	return nil
}

func runPackageBuildDir(ctx context.Context, a *app, dir string, env []string, verbose bool) error {
	packagePath := filepath.Join(dir, "package.json")
	if !fileutil.FileExists(packagePath) {
		return fmt.Errorf("build dir is not a supported target and has no package.json build script: %s", dir)
	}
	hasBuild, err := hasPackageBuildScript(packagePath)
	if err != nil {
		return err
	}
	if !hasBuild {
		return fmt.Errorf("build dir package.json has no non-empty build script: %s", dir)
	}
	fmt.Fprintf(a.stdout, "Building package: %s\n", toolutil.RelativePath(a.root, dir))
	return a.runCommand(ctx, commandOptions{Dir: dir, Env: env, Quiet: !verbose}, "pnpm", "run", "build")
}

func resolveBuildDir(root string, rawDir string) (string, error) {
	if strings.TrimSpace(rawDir) == "" {
		return "", errors.New("build dir cannot be empty")
	}
	dir := rawDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	clean, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", fmt.Errorf("resolve build dir %q: %w", rawDir, err)
	}
	info, err := os.Stat(clean)
	if err != nil {
		return "", fmt.Errorf("stat build dir %s: %w", clean, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("build dir is not a directory: %s", clean)
	}
	return clean, nil
}

func runHostFrontendBuild(ctx context.Context, a *app, env []string, verbose bool) error {
	fmt.Fprintln(a.stdout, "Building frontend...")
	if err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-vben"), Env: env, Quiet: !verbose}, "pnpm", "run", "build"); err != nil {
		return err
	}

	embedDir := filepath.Join(a.root, "apps", "lina-core", "internal", "packed", "public")
	if err := os.RemoveAll(embedDir); err != nil {
		return fmt.Errorf("clean frontend embed directory: %w", err)
	}
	if err := os.MkdirAll(embedDir, 0o755); err != nil {
		return fmt.Errorf("create frontend embed directory: %w", err)
	}
	if err := fileutil.CopyDirContents(filepath.Join(a.root, "apps", "lina-vben", "apps", "web-antd", "dist"), embedDir); err != nil {
		return err
	}
	if err := ensurePackedPublicPlaceholder(embedDir); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Host frontend embedded assets generated")
	return nil
}

func runHostBackendBuild(ctx context.Context, a *app, env []string, options buildOptions) error {
	multiPlatform := len(options.Targets) > 1
	for _, target := range options.Targets {
		targetBinary := filepath.Join(options.OutputDir, toolutil.ExecutableName(options.BinaryName))
		if multiPlatform {
			targetBinary = filepath.Join(options.OutputDir, target.OS+"_"+target.Arch, toolutil.ExecutableName(options.BinaryName))
		}
		if err := os.MkdirAll(filepath.Dir(targetBinary), 0o755); err != nil {
			return fmt.Errorf("create backend output directory: %w", err)
		}
		fmt.Fprintf(a.stdout, "Building backend for %s/%s...\n", target.OS, target.Arch)
		targetEnv := toolutil.SetEnvValue(toolutil.SetEnvValue(toolutil.SetEnvValue(env, "CGO_ENABLED", options.CGOEnabled), "GOOS", target.OS), "GOARCH", target.Arch)
		err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core"), Env: targetEnv, Quiet: !options.Verbose}, "go", "build", "-o", targetBinary, ".")
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Build complete: %s\n", toolutil.RelativePath(a.root, targetBinary))
	}
	return nil
}

// runPluginBuilds runs plugin-owned builds before Go compilation so
// plugins can generate embeddable files without storing build artifacts in Git.
func runPluginBuilds(ctx context.Context, a *app, env []string, options buildOptions) error {
	pluginDirs, err := discoverPluginBuildRoots(a.root)
	if err != nil {
		return err
	}
	if len(pluginDirs) == 0 {
		fmt.Fprintln(a.stdout, "No plugins found for build")
		return nil
	}
	fmt.Fprintf(a.stdout, "Building %d plugin(s)...\n", len(pluginDirs))
	for _, pluginDir := range pluginDirs {
		if err = runOnePluginBuild(ctx, a, pluginDir, env, options); err != nil {
			return err
		}
	}
	return nil
}

// discoverPluginBuildRoots returns direct plugin roots below apps/lina-plugins.
func discoverPluginBuildRoots(root string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "apps", "lina-plugins", "*", "plugin.yaml"))
	if err != nil {
		return nil, fmt.Errorf("scan plugin build roots: %w", err)
	}
	pluginDirs := make([]string, 0, len(matches))
	for _, manifestPath := range matches {
		pluginDirs = append(pluginDirs, filepath.Dir(manifestPath))
	}
	sort.Strings(pluginDirs)
	return pluginDirs, nil
}

// runPluginBuildSteps executes the plugin-owned build steps from its root.
func runPluginBuildSteps(ctx context.Context, a *app, pluginDir string, env []string, verbose bool, steps []pluginBuildStep) error {
	for _, step := range steps {
		if err := a.runCommand(ctx, commandOptions{Dir: pluginDir, Env: env, Quiet: !verbose}, step.Command, step.Args...); err != nil {
			return err
		}
	}
	return nil
}

// resolvePluginConfigBuildSteps reads build.commands from plugin hack/config.yaml.
func resolvePluginConfigBuildSteps(root string, pluginDir string) ([]pluginBuildStep, error) {
	configPath := filepath.Join(pluginDir, "hack", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plugin build config %s: %w", configPath, err)
	}
	var cfg pluginBuildConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse plugin build config %s: %w", configPath, err)
	}
	variables := map[string]string{
		"PLUGIN_ROOT": pluginDir,
		"REPO_ROOT":   root,
	}
	steps := []pluginBuildStep{}
	for index, command := range cfg.Build.Commands {
		raw := strings.TrimSpace(expandBuildVariables(command, variables))
		if raw == "" {
			continue
		}
		fields, splitErr := splitBuildCommandLine(raw)
		if splitErr != nil {
			return nil, fmt.Errorf("parse plugin build command %d in %s: %w", index+1, configPath, splitErr)
		}
		if len(fields) == 0 {
			continue
		}
		steps = append(steps, pluginBuildStep{Command: fields[0], Args: fields[1:]})
	}
	return steps, nil
}

// expandBuildVariables expands the small $(NAME) and ${NAME} subset needed by
// plugin build command declarations.
func expandBuildVariables(value string, variables map[string]string) string {
	expanded := value
	for i := 0; i < 8; i++ {
		next := expanded
		for key, replacement := range variables {
			next = strings.ReplaceAll(next, "$("+key+")", replacement)
			next = strings.ReplaceAll(next, "${"+key+"}", replacement)
		}
		if next == expanded {
			break
		}
		expanded = next
	}
	return expanded
}

// splitBuildCommandLine splits command arguments using whitespace and basic
// quote grouping without treating Windows path backslashes as escapes.
func splitBuildCommandLine(line string) ([]string, error) {
	fields := []string{}
	var current strings.Builder
	var quote rune
	for _, char := range line {
		if quote != 0 {
			if char == quote {
				quote = 0
				continue
			}
			current.WriteRune(char)
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		if char == ' ' || char == '\t' {
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(char)
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote")
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields, nil
}

func hasPackageBuildScript(manifestPath string) (bool, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return false, fmt.Errorf("read package build manifest %s: %w", manifestPath, err)
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return false, fmt.Errorf("parse package build manifest %s: %w", manifestPath, err)
	}
	build, ok := manifest.Scripts["build"]
	return ok && strings.TrimSpace(build) != "", nil
}

// ensurePackedPublicPlaceholder recreates the tracked placeholder after the
// build command refreshes ignored frontend assets in internal/packed/public.
func ensurePackedPublicPlaceholder(embedDir string) error {
	if err := os.WriteFile(filepath.Join(embedDir, packedPublicPlaceholderName), []byte{}, 0o644); err != nil {
		return fmt.Errorf("write frontend embed placeholder: %w", err)
	}
	return nil
}

// loadRootConfig loads repository tool defaults from hack/config.yaml.
func loadRootConfig(root string, input commandInput) (config.Root, error) {
	return config.Load(root, input.GetDefault("config", filepath.Join("hack", "config.yaml")))
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
