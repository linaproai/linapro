// This file verifies image build platform normalization and release safety
// checks without invoking Docker.

package imagebuilder

import (
	"reflect"
	"runtime"
	"testing"
)

// TestNormalizeBuildConfigSinglePlatform verifies one configured platform keeps
// the standard single-binary output contract.
func TestNormalizeBuildConfigSinglePlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"linux/arm64"}

	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	if cfg.Platform != "linux/arm64" {
		t.Fatalf("Platform = %q, want linux/arm64", cfg.Platform)
	}
	if cfg.MultiPlatform() {
		t.Fatalf("MultiPlatform = true, want false")
	}
	path := buildOutputBinaryRelPath(cfg, cfg.Targets[0])
	if path != "temp/output/lina" {
		t.Fatalf("single-platform binary path = %q, want temp/output/lina", path)
	}
}

// TestNormalizeBuildConfigAutoPlatform verifies auto resolves to the current
// Linux image platform for the local architecture.
func TestNormalizeBuildConfigAutoPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"auto"}

	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	want := "linux/" + runtime.GOARCH
	if cfg.Platform != want {
		t.Fatalf("Platform = %q, want %q", cfg.Platform, want)
	}
	if len(cfg.Platforms) != 1 || cfg.Platforms[0] != want {
		t.Fatalf("Platforms = %#v, want [%q]", cfg.Platforms, want)
	}
}

// TestNormalizeBuildConfigMultiPlatform verifies platform lists become stable
// per-platform binary output directories.
func TestNormalizeBuildConfigMultiPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"linux/amd64", "linux/arm64"}

	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	if !cfg.MultiPlatform() {
		t.Fatalf("MultiPlatform = false, want true")
	}
	if cfg.Platform != "linux/amd64,linux/arm64" {
		t.Fatalf("Platform = %q, want linux/amd64,linux/arm64", cfg.Platform)
	}
	if got := buildOutputBinaryRelPath(cfg, cfg.Targets[0]); got != "temp/output/linux_amd64/lina" {
		t.Fatalf("amd64 binary path = %q, want temp/output/linux_amd64/lina", got)
	}
	if got := buildOutputBinaryRelPath(cfg, cfg.Targets[1]); got != "temp/output/linux_arm64/lina" {
		t.Fatalf("arm64 binary path = %q, want temp/output/linux_arm64/lina", got)
	}
}

// TestSplitPlatformCSV verifies command-line platform overrides use a compact
// comma-separated form that becomes the config array.
func TestSplitPlatformCSV(t *testing.T) {
	got, err := splitPlatformCSV("linux/amd64,linux/arm64")
	if err != nil {
		t.Fatalf("splitPlatformCSV returned error: %v", err)
	}

	want := []string{"linux/amd64", "linux/arm64"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitPlatformCSV = %#v, want %#v", got, want)
	}
}

// TestValidateImageBuildRequestRequiresPushForMultiPlatform verifies buildx
// multi-platform publishing is rejected unless push is enabled.
func TestValidateImageBuildRequestRequiresPushForMultiPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"linux/amd64", "linux/arm64"}
	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	image := defaultImageConfig()
	image.Push = false
	if err := validateImageBuildRequest(image, cfg); err == nil {
		t.Fatalf("validateImageBuildRequest returned nil, want error")
	}

	image.Push = true
	image.Registry = "ghcr.io/linaproai"
	if err := validateImageBuildRequest(image, cfg); err != nil {
		t.Fatalf("validateImageBuildRequest returned error: %v", err)
	}
}

// TestValidateImageBuildRequestRejectsNonLinux verifies Docker image builds do
// not accidentally use host-only platforms such as darwin/arm64.
func TestValidateImageBuildRequestRejectsNonLinux(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"darwin/arm64"}
	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	image := defaultImageConfig()
	if err := validateImageBuildRequest(image, cfg); err == nil {
		t.Fatalf("validateImageBuildRequest returned nil, want non-linux platform error")
	}
}

func TestParseOptionalBoolRejectsNonStandardValue(t *testing.T) {
	if _, err := parseOptionalBool("maybe", false); err == nil {
		t.Fatalf("expected non-standard boolean value to be rejected")
	}
}

func TestOptionsFromInputUsesKebabCaseOnly(t *testing.T) {
	baseImageSnakeKey := "base" + "_" + "image"
	outputDirSnakeKey := "output" + "_" + "dir"
	binaryNameSnakeKey := "binary" + "_" + "name"

	opts, specified, err := optionsFromInput(imageBuildTestInput{params: map[string]string{
		"base-image":       "alpine:3.22",
		baseImageSnakeKey:  "ubuntu:24.04",
		"output-dir":       "temp/image",
		outputDirSnakeKey:  "temp/ignored",
		"binary-name":      "lina-custom",
		binaryNameSnakeKey: "ignored",
	}}, nil)
	if err != nil {
		t.Fatalf("optionsFromInput returned error: %v", err)
	}
	if opts.BaseImage != "alpine:3.22" || !specified["base-image"] {
		t.Fatalf("expected kebab-case base-image to be selected, opts=%#v specified=%#v", opts, specified)
	}
	if opts.OutputDir != "temp/image" || !specified["output-dir"] {
		t.Fatalf("expected kebab-case output-dir to be selected, opts=%#v specified=%#v", opts, specified)
	}
	if opts.BinaryName != "lina-custom" || !specified["binary-name"] {
		t.Fatalf("expected kebab-case binary-name to be selected, opts=%#v specified=%#v", opts, specified)
	}
	if specified[baseImageSnakeKey] || specified[outputDirSnakeKey] || specified[binaryNameSnakeKey] {
		t.Fatalf("snake_case keys must not be marked as specified: %#v", specified)
	}
}

// TestBuildxDockerArgs verifies multi-platform image publishing uses buildx
// with a platform matrix and push.
func TestBuildxDockerArgs(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platforms = []string{"linux/amd64", "linux/arm64"}
	if err := normalizeBuildConfig(&cfg); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}
	image := defaultImageConfig()
	args := buildxDockerArgs("/repo", image, cfg, "ghcr.io/linaproai/linapro:v1.0.0")

	want := []string{
		"buildx",
		"build",
		"--platform", "linux/amd64,linux/arm64",
		"--build-arg", "BASE_IMAGE=alpine:3.22",
		"-f", "/repo/hack/docker/Dockerfile",
		"-t", "ghcr.io/linaproai/linapro:v1.0.0",
		"--push",
		".",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("buildx args = %#v, want %#v", args, want)
	}
}

// TestDockerBuildArgs verifies single-platform local builds pass the target
// platform to Dockerfile ARGs.
func TestDockerBuildArgs(t *testing.T) {
	var (
		image  = defaultImageConfig()
		target = targetPlatform{OS: "linux", Arch: "arm64"}
		args   = dockerBuildArgs("/repo", image, target, "linapro:test")
	)

	want := []string{
		"build",
		"--platform", "linux/arm64",
		"--build-arg", "BASE_IMAGE=alpine:3.22",
		"--build-arg", "TARGETOS=linux",
		"--build-arg", "TARGETARCH=arm64",
		"-f", "/repo/hack/docker/Dockerfile",
		"-t", "linapro:test",
		".",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("docker build args = %#v, want %#v", args, want)
	}
}

type imageBuildTestInput struct {
	params map[string]string
}

func (input imageBuildTestInput) Get(key string) string {
	return input.params[key]
}

func (input imageBuildTestInput) GetDefault(key string, fallback string) string {
	if value := input.Get(key); value != "" {
		return value
	}
	return fallback
}

func (input imageBuildTestInput) Has(key string) bool {
	_, ok := input.params[key]
	return ok
}

func (input imageBuildTestInput) Bool(key string, fallback bool) (bool, error) {
	return parseOptionalBool(input.Get(key), fallback)
}

func (input imageBuildTestInput) ParamMap() map[string]string {
	return input.params
}
