// Package main implements the Docker image build orchestration used by the
// repository-level make image target.
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
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// rootConfig stores repository-level tool configuration from hack/config.yaml.
type rootConfig struct {
	Build buildConfig `yaml:"build"`
	Image imageConfig `yaml:"image"`
}

// Repository convention paths used by LinaPro image builds.
const (
	conventionImageBinaryRoot = "temp/image"
	conventionImageBinaryName = "lina"
)

// buildConfig stores user-facing build defaults.
type buildConfig struct {
	Platforms  []string         `yaml:"platforms"`
	CGOEnabled bool             `yaml:"cgoEnabled"`
	OutputDir  string           `yaml:"outputDir"`
	BinaryName string           `yaml:"binaryName"`
	Platform   string           `yaml:"-"`
	Targets    []targetPlatform `yaml:"-"`
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

// cliOptions stores one invocation's command-line overrides.
type cliOptions struct {
	ConfigPath    string
	BuildOnly     bool
	Preflight     bool
	PrintBuildEnv bool
	Image         string
	Tag           string
	Registry      string
	Push          string
	Platforms     string
	CGOEnabled    string
	OutputDir     string
	BinaryName    string
	BaseImage     string
	Verbose       string
}

// targetPlatform stores one normalized Go and Docker target platform.
type targetPlatform struct {
	OS   string
	Arch string
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

	cfg := defaultRootConfig()
	if err = loadConfig(configPath, &cfg); err != nil {
		return err
	}
	if err = applyBuildOverrides(&cfg.Build, opts, specified); err != nil {
		return err
	}
	if err = normalizeBuildConfig(&cfg.Build); err != nil {
		return err
	}
	if opts.PrintBuildEnv {
		printBuildEnv(cfg.Build)
		return nil
	}

	if err = applyImageOverrides(&cfg.Image, opts, specified); err != nil {
		return err
	}
	if err = normalizeImageConfig(repoRoot, &cfg.Image); err != nil {
		return err
	}
	if opts.Preflight {
		return validateImageBuildRequest(cfg.Image, cfg.Build)
	}

	verbose, err := parseOptionalBool(opts.Verbose, false)
	if err != nil {
		return fmt.Errorf("parse verbose: %w", err)
	}

	runner := commandRunner{Root: repoRoot, Verbose: verbose}

	if !opts.BuildOnly {
		err = validateImageBuildRequest(cfg.Image, cfg.Build)
	}
	if err != nil {
		return err
	}

	for _, target := range cfg.Build.Targets {
		binaryPath := buildOutputBinaryPath(repoRoot, cfg.Build, target)
		if err = validateExistingBinary(binaryPath); err != nil {
			return err
		}
		stagedBinaryPath := imageStagedBinaryPath(repoRoot, cfg.Build, target)
		if err = stageImageBinary(binaryPath, stagedBinaryPath); err != nil {
			return err
		}
	}

	if opts.BuildOnly {
		fmt.Printf("✓ image build artifacts are ready: %s\n", filepath.Join(repoRoot, conventionImageBinaryRoot))
		return nil
	}

	imageRef := buildImageRef(cfg.Image)
	if err = buildDockerImage(repoRoot, cfg.Image, cfg.Build, runner, imageRef); err != nil {
		return err
	}
	if cfg.Build.MultiPlatform() {
		fmt.Printf("✓ Docker image pushed: %s\n", imageRef)
		return nil
	}
	if cfg.Image.Push {
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
	flag.BoolVar(&opts.Preflight, "preflight", false, "Validate image build request without checking artifacts or running docker build")
	flag.BoolVar(&opts.PrintBuildEnv, "print-build-env", false, "Print normalized build config as shell assignments")
	flag.StringVar(&opts.Image, "image", "", "Override image repository name")
	flag.StringVar(&opts.Tag, "tag", "", "Override image tag")
	flag.StringVar(&opts.Registry, "registry", "", "Override image registry prefix")
	flag.StringVar(&opts.Push, "push", "", "Override push behavior")
	flag.StringVar(&opts.Platforms, "platforms", "", "Override build target platforms")
	flag.StringVar(&opts.CGOEnabled, "cgo-enabled", "", "Override CGO build behavior")
	flag.StringVar(&opts.OutputDir, "output-dir", "", "Override build output directory")
	flag.StringVar(&opts.BinaryName, "binary-name", "", "Override build binary filename")
	flag.StringVar(&opts.BaseImage, "base-image", "", "Override Docker base image")
	flag.StringVar(&opts.Verbose, "verbose", "", "Show child command output")
	flag.Parse()

	specified := map[string]bool{}
	flag.Visit(func(item *flag.Flag) {
		specified[item.Name] = true
	})
	return opts, specified
}

// defaultRootConfig returns stable defaults used when config values are omitted.
func defaultRootConfig() rootConfig {
	return rootConfig{
		Build: defaultBuildConfig(),
		Image: defaultImageConfig(),
	}
}

// defaultBuildConfig returns stable build defaults.
func defaultBuildConfig() buildConfig {
	return buildConfig{
		Platforms:  []string{"linux/" + runtime.GOARCH},
		CGOEnabled: false,
		OutputDir:  "temp/output",
		BinaryName: "lina",
	}
}

// defaultImageConfig returns stable image metadata defaults.
func defaultImageConfig() imageConfig {
	return imageConfig{
		Name:       "linapro",
		BaseImage:  "alpine:3.22",
		Dockerfile: "hack/docker/Dockerfile",
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

// loadConfig overlays root config from a YAML file.
func loadConfig(configPath string, cfg *rootConfig) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	parsed := *cfg
	if err = yaml.Unmarshal(content, &parsed); err != nil {
		return err
	}
	*cfg = parsed
	return nil
}

// applyBuildOverrides merges command-line overrides into build config values.
func applyBuildOverrides(cfg *buildConfig, opts cliOptions, specified map[string]bool) error {
	if specified["platforms"] {
		platforms, err := splitPlatformCSV(opts.Platforms)
		if err != nil {
			return err
		}
		cfg.Platforms = platforms
	}
	if specified["cgo-enabled"] {
		value, err := parseOptionalBool(opts.CGOEnabled, cfg.CGOEnabled)
		if err != nil {
			return fmt.Errorf("parse cgo-enabled: %w", err)
		}
		cfg.CGOEnabled = value
	}
	if specified["output-dir"] {
		cfg.OutputDir = opts.OutputDir
	}
	if specified["binary-name"] {
		cfg.BinaryName = opts.BinaryName
	}
	return nil
}

// applyImageOverrides merges environment and command-line overrides into image metadata values.
func applyImageOverrides(cfg *imageConfig, opts cliOptions, specified map[string]bool) error {
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
	if specified["base-image"] {
		cfg.BaseImage = opts.BaseImage
	}
	return nil
}

// normalizeBuildConfig validates and completes derived build config values.
func normalizeBuildConfig(cfg *buildConfig) error {
	targets, err := parsePlatformList(cfg.Platforms)
	if err != nil {
		return err
	}
	cfg.Platform = joinPlatformCSV(targets)
	cfg.Platforms = platformValues(targets)
	cfg.Targets = targets
	cfg.OutputDir = filepath.Clean(strings.TrimSpace(cfg.OutputDir))
	cfg.BinaryName = strings.TrimSpace(cfg.BinaryName)

	if cfg.Platform == "" {
		return errors.New("build.platforms cannot be empty")
	}
	if cfg.OutputDir == "." || cfg.OutputDir == "" {
		return errors.New("build.outputDir cannot be empty")
	}
	if filepath.IsAbs(cfg.OutputDir) {
		return errors.New("build.outputDir must be relative to the repository root")
	}
	if cfg.BinaryName == "" {
		return errors.New("build.binaryName cannot be empty")
	}
	if strings.ContainsAny(cfg.BinaryName, `/\`) {
		return errors.New("build.binaryName must be a file name, not a path")
	}
	return nil
}

// splitPlatformCSV parses a command-line comma-separated platform override.
func splitPlatformCSV(value string) ([]string, error) {
	items := strings.Split(value, ",")
	platforms := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			return nil, errors.New("build.platforms contains an empty platform entry")
		}
		platforms = append(platforms, normalized)
	}
	if len(platforms) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	return platforms, nil
}

// parsePlatformList parses one Docker/Go platform list from configuration.
func parsePlatformList(values []string) ([]targetPlatform, error) {
	if len(values) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	targets := make([]targetPlatform, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		target, err := parseTargetPlatform(value)
		if err != nil {
			return nil, err
		}
		key := target.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	return targets, nil
}

// parseTargetPlatform parses one target platform in goos/goarch form.
func parseTargetPlatform(value string) (targetPlatform, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return targetPlatform{}, errors.New("build.platforms contains an empty platform entry")
	}
	if normalized == "auto" {
		return targetPlatform{OS: runtime.GOOS, Arch: runtime.GOARCH}, nil
	}
	parts := strings.Split(normalized, "/")
	if len(parts) != 2 {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must use goos/goarch form or auto", value)
	}
	target := targetPlatform{
		OS:   strings.TrimSpace(parts[0]),
		Arch: strings.TrimSpace(parts[1]),
	}
	if target.OS == "" || target.Arch == "" {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must include both goos and goarch", value)
	}
	if strings.ContainsAny(target.OS+target.Arch, " \t\r\n") {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must not contain whitespace", value)
	}
	return target, nil
}

// normalizeImageConfig validates and completes derived image metadata values.
func normalizeImageConfig(repoRoot string, cfg *imageConfig) error {
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.Tag = strings.TrimSpace(cfg.Tag)
	cfg.Registry = strings.Trim(strings.TrimSpace(cfg.Registry), "/")
	cfg.BaseImage = strings.TrimSpace(cfg.BaseImage)
	cfg.Dockerfile = filepath.Clean(strings.TrimSpace(cfg.Dockerfile))
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
	if cfg.Dockerfile == "." || cfg.Dockerfile == "" {
		return errors.New("image.dockerfile cannot be empty")
	}
	if filepath.IsAbs(cfg.Dockerfile) {
		return errors.New("image.dockerfile must be relative to the repository root")
	}
	return nil
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

// printBuildEnv emits normalized build config as shell-safe assignments for make recipes.
func printBuildEnv(cfg buildConfig) {
	multiPlatform := "0"
	if cfg.MultiPlatform() {
		multiPlatform = "1"
	}
	fmt.Printf("BUILD_PLATFORM=%s\n", shellQuote(cfg.Platform))
	fmt.Printf("BUILD_PLATFORMS=%s\n", shellQuote(joinPlatformSpace(cfg.Targets)))
	fmt.Printf("BUILD_PLATFORM_COUNT=%d\n", len(cfg.Targets))
	fmt.Printf("BUILD_MULTI_PLATFORM=%s\n", shellQuote(multiPlatform))
	fmt.Printf("BUILD_CGO_ENABLED=%s\n", shellQuote(cgoEnabledValue(cfg.CGOEnabled)))
	fmt.Printf("BUILD_OUTPUT_DIR=%s\n", shellQuote(filepath.ToSlash(cfg.OutputDir)))
	fmt.Printf("BUILD_BINARY_NAME=%s\n", shellQuote(cfg.BinaryName))
	fmt.Printf("BUILD_BINARY_PATH=%s\n", shellQuote(filepath.ToSlash(defaultBuildBinaryPath(cfg))))
}

// shellQuote returns a POSIX shell single-quoted literal.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// cgoEnabledValue converts a boolean into the value expected by CGO_ENABLED.
func cgoEnabledValue(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}

// validateImageBuildRequest verifies Docker build settings before invoking Docker.
func validateImageBuildRequest(image imageConfig, build buildConfig) error {
	if image.Push {
		if strings.TrimSpace(image.Registry) == "" {
			return errors.New("push=true requires image.registry in hack/config.yaml, registry=<prefix>, or LINAPRO_IMAGE_REGISTRY")
		}
	}
	if build.MultiPlatform() && !image.Push {
		return errors.New("multi-platform Docker image builds require push=1 so Docker buildx can publish a usable manifest")
	}
	return nil
}

// validateExistingBinary checks that the image input binary exists.
func validateExistingBinary(binaryPath string) error {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("build binary is missing: %s; run make build first", binaryPath)
	}
	if info.IsDir() {
		return fmt.Errorf("build binary path is a directory: %s", binaryPath)
	}
	fmt.Printf("✓ build binary is ready: %s\n", binaryPath)
	return nil
}

// stageImageBinary copies the standard build output into the Dockerfile input path.
func stageImageBinary(source string, target string) error {
	if sameFile(source, target) {
		fmt.Printf("✓ image binary staged: %s\n", target)
		return nil
	}
	if err := copyFile(source, target); err != nil {
		return err
	}
	fmt.Printf("✓ image binary staged: %s\n", target)
	return nil
}

// buildOutputBinaryPath returns the expected make build binary path for one target.
func buildOutputBinaryPath(repoRoot string, build buildConfig, target targetPlatform) string {
	return filepath.Join(repoRoot, buildOutputBinaryRelPath(build, target))
}

// buildOutputBinaryRelPath returns the repository-relative binary path for one target.
func buildOutputBinaryRelPath(build buildConfig, target targetPlatform) string {
	if build.MultiPlatform() {
		return filepath.Join(build.OutputDir, target.DirName(), build.BinaryName)
	}
	return filepath.Join(build.OutputDir, build.BinaryName)
}

// defaultBuildBinaryPath returns the printable primary build artifact path.
func defaultBuildBinaryPath(build buildConfig) string {
	if build.MultiPlatform() && len(build.Targets) > 0 {
		return buildOutputBinaryRelPath(build, build.Targets[0])
	}
	return filepath.Join(build.OutputDir, build.BinaryName)
}

// imageStagedBinaryPath returns the Dockerfile input binary path for one target.
func imageStagedBinaryPath(repoRoot string, build buildConfig, target targetPlatform) string {
	return filepath.Join(repoRoot, conventionImageBinaryRoot, target.OS, target.Arch, conventionImageBinaryName)
}

// sameFile reports whether two paths point to the same filesystem object.
func sameFile(source string, target string) bool {
	sourceAbs, sourceErr := filepath.Abs(source)
	targetAbs, targetErr := filepath.Abs(target)
	if sourceErr == nil && targetErr == nil && sourceAbs == targetAbs {
		return true
	}
	sourceInfo, sourceErr := os.Stat(source)
	if sourceErr != nil {
		return false
	}
	targetInfo, targetErr := os.Stat(target)
	if targetErr != nil {
		return false
	}
	return os.SameFile(sourceInfo, targetInfo)
}

// buildDockerImage runs docker build or buildx with the configured image platforms.
func buildDockerImage(repoRoot string, image imageConfig, build buildConfig, runner commandRunner, imageRef string) error {
	if build.MultiPlatform() {
		args := buildxDockerArgs(repoRoot, image, build, imageRef)
		fmt.Printf("Building multi-platform Docker image: %s (%s)\n", imageRef, build.Platform)
		dockerRunner := runner
		dockerRunner.Verbose = true
		return dockerRunner.Run(".", nil, "docker", args...)
	}
	target := build.Targets[0]
	args := dockerBuildArgs(repoRoot, image, target, imageRef)
	fmt.Printf("Building Docker image: %s\n", imageRef)
	dockerRunner := runner
	dockerRunner.Verbose = true
	return dockerRunner.Run(".", nil, "docker", args...)
}

// buildxDockerArgs returns Docker buildx arguments for multi-platform publishing.
func buildxDockerArgs(repoRoot string, image imageConfig, build buildConfig, imageRef string) []string {
	return []string{
		"buildx",
		"build",
		"--platform", build.Platform,
		"--build-arg", "BASE_IMAGE=" + image.BaseImage,
		"-f", filepath.Join(repoRoot, image.Dockerfile),
		"-t", imageRef,
		"--push",
		".",
	}
}

// dockerBuildArgs returns Docker build arguments for one local image platform.
func dockerBuildArgs(repoRoot string, image imageConfig, target targetPlatform, imageRef string) []string {
	return []string{
		"build",
		"--platform", target.String(),
		"--build-arg", "BASE_IMAGE=" + image.BaseImage,
		"--build-arg", "TARGETOS=" + target.OS,
		"--build-arg", "TARGETARCH=" + target.Arch,
		"-f", filepath.Join(repoRoot, image.Dockerfile),
		"-t", imageRef,
		".",
	}
}

// buildImageRef composes the final Docker image reference.
func buildImageRef(cfg imageConfig) string {
	name := cfg.Name + ":" + cfg.Tag
	if strings.TrimSpace(cfg.Registry) == "" {
		return name
	}
	return strings.Trim(cfg.Registry, "/") + "/" + name
}

// MultiPlatform reports whether the build targets more than one platform.
func (cfg buildConfig) MultiPlatform() bool {
	return len(cfg.Targets) > 1
}

// String returns the canonical goos/goarch representation.
func (p targetPlatform) String() string {
	return p.OS + "/" + p.Arch
}

// DirName returns the filesystem-safe directory segment for build outputs.
func (p targetPlatform) DirName() string {
	return p.OS + "_" + p.Arch
}

// joinPlatformCSV joins targets in Docker platform-list format.
func joinPlatformCSV(targets []targetPlatform) string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return strings.Join(values, ",")
}

// platformValues returns normalized platform strings for the in-memory config.
func platformValues(targets []targetPlatform) []string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return values
}

// joinPlatformSpace joins targets for shell for-loops in make recipes.
func joinPlatformSpace(targets []targetPlatform) string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return strings.Join(values, " ")
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

// copyFile copies one regular file preserving permission bits.
func copyFile(source string, target string) error {
	if sameFile(source, target) {
		return nil
	}
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
