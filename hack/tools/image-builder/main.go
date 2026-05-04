// This file implements the cross-platform Docker image build orchestration used
// by the repository-level make image target.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// rootConfig stores repository-level tool configuration from hack/config.yaml.
type rootConfig struct {
	Image imageConfig `yaml:"image"`
}

// Repository convention paths used by LinaPro image builds.
const (
	conventionBackendDir        = "apps/lina-core"
	conventionFrontendDir       = "apps/lina-vben"
	conventionPluginsDir        = "apps/lina-plugins"
	conventionEmbedPublicDir    = "apps/lina-core/internal/packed/public"
	conventionManifestSourceDir = "apps/lina-core/manifest"
	conventionPackedManifestDir = "apps/lina-core/internal/packed/manifest"
	conventionWasmBuilderDir    = "hack/tools/build-wasm"
)

// imageConfig stores user-facing image build defaults.
type imageConfig struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	Registry   string `yaml:"registry"`
	Push       bool   `yaml:"push"`
	BaseImage  string `yaml:"baseImage"`
	OS         string `yaml:"os"`
	Arch       string `yaml:"arch"`
	Platform   string `yaml:"platform"`
	Dockerfile string `yaml:"dockerfile"`
	OutputDir  string `yaml:"outputDir"`
	BinaryName string `yaml:"binaryName"`
}

// cliOptions stores one invocation's command-line overrides.
type cliOptions struct {
	ConfigPath string
	BuildOnly  bool
	Image      string
	Tag        string
	Registry   string
	Push       string
	OS         string
	Arch       string
	Platform   string
	BaseImage  string
	SkipBuild  string
	Verbose    string
}

// pluginManifest stores the minimal plugin metadata needed for dynamic plugin discovery.
type pluginManifest struct {
	Type string `yaml:"type"`
}

// commandRunner executes external tools from the repository root.
type commandRunner struct {
	Root    string
	Verbose bool
}

// main runs the image builder and renders concise errors for make output.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "image-builder: %v\n", err)
		os.Exit(1)
	}
}

// run parses options, loads configuration, and executes the requested workflow.
func run() error {
	opts, specified := parseOptions()
	repoRoot, configPath, err := discoverRepoRoot(opts.ConfigPath)
	if err != nil {
		return err
	}
	cfg := defaultImageConfig()
	if err = loadConfig(configPath, &cfg); err != nil {
		return err
	}
	if err = applyOverrides(&cfg, opts, specified); err != nil {
		return err
	}
	if err = normalizeConfig(repoRoot, &cfg); err != nil {
		return err
	}

	verbose, err := parseOptionalBool(opts.Verbose, false)
	if err != nil {
		return fmt.Errorf("parse verbose: %w", err)
	}
	skipBuild, err := parseOptionalBool(opts.SkipBuild, false)
	if err != nil {
		return fmt.Errorf("parse skip-build: %w", err)
	}

	runner := commandRunner{Root: repoRoot, Verbose: verbose}
	binaryPath := filepath.Join(repoRoot, cfg.OutputDir, cfg.BinaryName)
	if !skipBuild {
		if err = prepareImageBuild(repoRoot, cfg, runner, binaryPath); err != nil {
			return err
		}
	} else if err = validateExistingBinary(binaryPath); err != nil {
		return err
	}

	if opts.BuildOnly {
		fmt.Printf("✓ image build artifact is ready: %s\n", binaryPath)
		return nil
	}

	if cfg.Push {
		if strings.TrimSpace(cfg.Registry) == "" {
			return errors.New("push=true requires image.registry in hack/config.yaml, registry=<prefix>, or LINAPRO_IMAGE_REGISTRY")
		}
	}

	imageRef := buildImageRef(cfg)
	if err = buildDockerImage(repoRoot, cfg, runner, imageRef); err != nil {
		return err
	}
	if cfg.Push {
		dockerRunner := runner
		dockerRunner.Verbose = true
		if err = dockerRunner.Run(".", nil, "docker", "push", imageRef); err != nil {
			return err
		}
		fmt.Printf("✓ Docker image pushed: %s\n", imageRef)
		return nil
	}
	fmt.Printf("✓ Docker image built: %s\n", imageRef)
	return nil
}

// parseOptions reads flags and records which values were explicitly set.
func parseOptions() (cliOptions, map[string]bool) {
	opts := cliOptions{}
	flag.StringVar(&opts.ConfigPath, "config", "hack/config.yaml", "Repository tool config path")
	flag.BoolVar(&opts.BuildOnly, "build-only", false, "Prepare image build artifacts without running docker build")
	flag.StringVar(&opts.Image, "image", "", "Override image repository name")
	flag.StringVar(&opts.Tag, "tag", "", "Override image tag")
	flag.StringVar(&opts.Registry, "registry", "", "Override image registry prefix")
	flag.StringVar(&opts.Push, "push", "", "Override push behavior")
	flag.StringVar(&opts.OS, "os", "", "Override target OS")
	flag.StringVar(&opts.Arch, "arch", "", "Override target architecture")
	flag.StringVar(&opts.Platform, "platform", "", "Override Docker platform")
	flag.StringVar(&opts.BaseImage, "base-image", "", "Override Docker base image")
	flag.StringVar(&opts.SkipBuild, "skip-build", "", "Skip frontend, wasm, and backend binary build")
	flag.StringVar(&opts.Verbose, "verbose", "", "Show child command output")
	flag.Parse()

	specified := map[string]bool{}
	flag.Visit(func(item *flag.Flag) {
		specified[item.Name] = true
	})
	return opts, specified
}

// defaultImageConfig returns stable defaults used when config values are omitted.
func defaultImageConfig() imageConfig {
	return imageConfig{
		Name:       "linapro",
		BaseImage:  "alpine:3.22",
		OS:         "linux",
		Arch:       "auto",
		Platform:   "auto",
		Dockerfile: "hack/docker/Dockerfile",
		OutputDir:  "temp/image",
		BinaryName: "lina",
	}
}

// discoverRepoRoot searches upward until the configured file is found.
func discoverRepoRoot(configPath string) (string, string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	if filepath.IsAbs(configPath) {
		if _, statErr := os.Stat(configPath); statErr != nil {
			return "", "", statErr
		}
		return filepath.Dir(filepath.Dir(configPath)), configPath, nil
	}
	current := start
	for {
		candidate := filepath.Join(current, configPath)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return current, candidate, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", "", fmt.Errorf("cannot find %s from %s or its parents", configPath, start)
}

// loadConfig overlays image config from a YAML file.
func loadConfig(configPath string, cfg *imageConfig) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	parsed := rootConfig{Image: *cfg}
	if err = yaml.Unmarshal(content, &parsed); err != nil {
		return err
	}
	*cfg = parsed.Image
	return nil
}

// applyOverrides merges environment and command-line overrides into config values.
func applyOverrides(cfg *imageConfig, opts cliOptions, specified map[string]bool) error {
	if envRegistry := strings.TrimSpace(os.Getenv("LINAPRO_IMAGE_REGISTRY")); envRegistry != "" && !specified["registry"] {
		cfg.Registry = envRegistry
	}
	if specified["image"] {
		cfg.Name = opts.Image
	}
	if specified["tag"] {
		cfg.Tag = opts.Tag
	}
	if specified["registry"] {
		cfg.Registry = opts.Registry
	}
	if specified["push"] {
		value, err := parseOptionalBool(opts.Push, cfg.Push)
		if err != nil {
			return fmt.Errorf("parse push: %w", err)
		}
		cfg.Push = value
	}
	if specified["os"] {
		cfg.OS = opts.OS
	}
	if specified["arch"] {
		cfg.Arch = opts.Arch
	}
	if specified["platform"] {
		cfg.Platform = opts.Platform
	}
	if specified["base-image"] {
		cfg.BaseImage = opts.BaseImage
	}
	return nil
}

// normalizeConfig validates and completes derived config values.
func normalizeConfig(repoRoot string, cfg *imageConfig) error {
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.Tag = strings.TrimSpace(cfg.Tag)
	cfg.Registry = strings.Trim(strings.TrimSpace(cfg.Registry), "/")
	cfg.BaseImage = strings.TrimSpace(cfg.BaseImage)
	cfg.OS = normalizeAuto(cfg.OS, "linux")
	cfg.Arch = normalizeAuto(cfg.Arch, runtime.GOARCH)
	if cfg.Platform = normalizeAuto(cfg.Platform, cfg.OS+"/"+cfg.Arch); cfg.Platform == "" {
		cfg.Platform = cfg.OS + "/" + cfg.Arch
	}
	if cfg.Tag == "" {
		tag, err := deriveGitTag(repoRoot)
		if err != nil {
			return err
		}
		cfg.Tag = tag
	}
	if cfg.Name == "" {
		return errors.New("image.name cannot be empty")
	}
	if cfg.Tag == "" {
		return errors.New("image tag cannot be empty")
	}
	if cfg.BaseImage == "" {
		return errors.New("image.baseImage cannot be empty")
	}
	requiredPaths := map[string]string{
		"dockerfile": cfg.Dockerfile,
		"outputDir":  cfg.OutputDir,
		"binaryName": cfg.BinaryName,
	}
	for name, value := range requiredPaths {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("image.%s cannot be empty", name)
		}
	}
	return nil
}

// normalizeAuto resolves blank or auto values to a default.
func normalizeAuto(value string, fallback string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" || strings.EqualFold(normalized, "auto") {
		return strings.TrimSpace(fallback)
	}
	return normalized
}

// parseOptionalBool parses optional bool-ish values used by make variables.
func parseOptionalBool(value string, fallback bool) (bool, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(normalized)
	if err != nil {
		return false, err
	}
	return parsed, nil
}

// deriveGitTag returns a git-derived image tag with latest as the final fallback.
func deriveGitTag(repoRoot string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty", "--match", "v*")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "latest", nil
	}
	tag := strings.TrimSpace(string(output))
	if tag == "" {
		return "latest", nil
	}
	return tag, nil
}

// prepareImageBuild builds frontend assets, dynamic plugins, and the host binary.
func prepareImageBuild(repoRoot string, cfg imageConfig, runner commandRunner, binaryPath string) error {
	outputDir := filepath.Join(repoRoot, cfg.OutputDir)
	fmt.Println("Preparing image build output directory...")
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	fmt.Println("Building frontend assets...")
	if err := runner.Run(conventionFrontendDir, nil, "pnpm", "run", "build"); err != nil {
		return err
	}
	if err := syncFrontendAssets(repoRoot, cfg); err != nil {
		return err
	}
	fmt.Println("✓ host frontend assets prepared")

	if err := preparePackedManifest(repoRoot, cfg); err != nil {
		return err
	}
	fmt.Println("✓ host manifest assets prepared")

	fmt.Println("Building dynamic plugin artifacts...")
	if err := buildDynamicPlugins(repoRoot, cfg, runner, outputDir); err != nil {
		return err
	}
	fmt.Println("✓ dynamic plugin artifacts prepared")

	fmt.Printf("Building host binary for image (%s/%s)...\n", cfg.OS, cfg.Arch)
	env := []string{
		"CGO_ENABLED=0",
		"GOOS=" + cfg.OS,
		"GOARCH=" + cfg.Arch,
	}
	if err := runner.Run(conventionBackendDir, env, "go", "build", "-o", binaryPath, "."); err != nil {
		return err
	}
	return validateExistingBinary(binaryPath)
}

// syncFrontendAssets copies the Vben production build into the host embed directory.
func syncFrontendAssets(repoRoot string, cfg imageConfig) error {
	source := filepath.Join(repoRoot, conventionFrontendDir, "apps", "web-antd", "dist")
	target := filepath.Join(repoRoot, conventionEmbedPublicDir)
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	if err := copyDirContents(source, target); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(target, ".gitkeep"), []byte{}, 0o644)
}

// preparePackedManifest mirrors distributable manifest assets into the embed workspace.
func preparePackedManifest(repoRoot string, cfg imageConfig) error {
	source := filepath.Join(repoRoot, conventionManifestSourceDir)
	target := filepath.Join(repoRoot, conventionPackedManifestDir)
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(target, "config"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(target, "sql"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(target, "i18n"), 0o755); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(source, "config", "config.template.yaml"), filepath.Join(target, "config", "config.template.yaml")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(source, "config", "metadata.yaml"), filepath.Join(target, "config", "metadata.yaml")); err != nil {
		return err
	}
	if err := copyDirContents(filepath.Join(source, "sql"), filepath.Join(target, "sql")); err != nil {
		return err
	}
	if err := copyDirContents(filepath.Join(source, "i18n"), filepath.Join(target, "i18n")); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(target, ".gitkeep"), []byte{}, 0o644)
}

// buildDynamicPlugins discovers dynamic source plugins and invokes build-wasm for each.
func buildDynamicPlugins(repoRoot string, cfg imageConfig, runner commandRunner, outputDir string) error {
	pluginIDs, err := discoverDynamicPlugins(filepath.Join(repoRoot, conventionPluginsDir))
	if err != nil {
		return err
	}
	if len(pluginIDs) == 0 {
		fmt.Println("No dynamic wasm plugins found")
		return nil
	}
	for _, pluginID := range pluginIDs {
		fmt.Printf("Building dynamic wasm plugin: %s\n", pluginID)
		pluginDir := filepath.Join(repoRoot, conventionPluginsDir, pluginID)
		if err = runner.Run(conventionWasmBuilderDir, nil, "go", "run", ".", "--plugin-dir", pluginDir, "--output-dir", outputDir); err != nil {
			return err
		}
	}
	return nil
}

// discoverDynamicPlugins returns source plugin IDs whose manifest type is dynamic.
func discoverDynamicPlugins(pluginsRoot string) ([]string, error) {
	entries, err := os.ReadDir(pluginsRoot)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(pluginsRoot, entry.Name(), "plugin.yaml")
		content, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return nil, readErr
		}
		manifest := pluginManifest{}
		if err = yaml.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}
		if strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic") {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// validateExistingBinary checks that the image input binary exists.
func validateExistingBinary(binaryPath string) error {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("image binary is missing: %s", binaryPath)
	}
	if info.IsDir() {
		return fmt.Errorf("image binary path is a directory: %s", binaryPath)
	}
	fmt.Printf("✓ image binary is ready: %s\n", binaryPath)
	return nil
}

// buildDockerImage runs docker build with the configured image platform and base image.
func buildDockerImage(repoRoot string, cfg imageConfig, runner commandRunner, imageRef string) error {
	args := []string{
		"build",
		"--platform", cfg.Platform,
		"--build-arg", "BASE_IMAGE=" + cfg.BaseImage,
		"-f", filepath.Join(repoRoot, cfg.Dockerfile),
		"-t", imageRef,
		".",
	}
	fmt.Printf("Building Docker image: %s\n", imageRef)
	dockerRunner := runner
	dockerRunner.Verbose = true
	return dockerRunner.Run(".", nil, "docker", args...)
}

// buildImageRef composes the final Docker image reference.
func buildImageRef(cfg imageConfig) string {
	name := cfg.Name + ":" + cfg.Tag
	if strings.TrimSpace(cfg.Registry) == "" {
		return name
	}
	return strings.Trim(cfg.Registry, "/") + "/" + name
}

// Run executes one external command in a repository-relative directory.
func (r commandRunner) Run(dir string, env []string, name string, args ...string) error {
	workingDir := filepath.Join(r.Root, dir)
	cmd := exec.Command(name, args...)
	cmd.Dir = workingDir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	display := name + " " + strings.Join(args, " ")
	if r.Verbose {
		fmt.Printf("+ %s\n", display)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s in %s failed: %w\n%s", display, workingDir, err, string(output))
	}
	return nil
}

// copyDirContents copies all entries from source into target.
func copyDirContents(source string, target string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = copyPath(filepath.Join(source, entry.Name()), filepath.Join(target, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// copyPath recursively copies a file or directory.
func copyPath(source string, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err = os.MkdirAll(target, info.Mode().Perm()); err != nil {
			return err
		}
		return copyDirContents(source, target)
	}
	return copyFile(source, target)
}

// copyFile copies one regular file preserving permission bits.
func copyFile(source string, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close source file %s: %v\n", source, closeErr)
		}
	}()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close target file %s after copy failure: %v\n", target, closeErr)
		}
		return err
	}
	return out.Close()
}
