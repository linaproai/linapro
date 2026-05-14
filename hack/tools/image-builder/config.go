// This file defines image-builder configuration models and normalization.

package main

import (
	"errors"
	"fmt"
	"os"
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

// MultiPlatform reports whether the build targets more than one platform.
func (cfg buildConfig) MultiPlatform() bool {
	return len(cfg.Targets) > 1
}
