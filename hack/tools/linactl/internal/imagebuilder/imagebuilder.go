// Package imagebuilder implements LinaPro Docker image staging and build
// orchestration for linactl. It owns repository image configuration parsing,
// target platform validation, host binary staging, and Docker CLI argument
// construction while command files keep the higher-level build ordering.
package imagebuilder

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
)

// Input exposes normalized linactl command parameters.
type Input interface {
	Get(string) string
	GetDefault(string, string) string
	Has(string) bool
	Bool(string, bool) (bool, error)
	ParamMap() map[string]string
}

// CommandRunner runs one child command in a working directory.
type CommandRunner func(context.Context, string, string, ...string) error

type commandRunner struct {
	root    string
	verbose bool
	run     CommandRunner
}

// Run builds image artifacts, performs preflight validation, or invokes Docker
// using the supplied output writers.
func Run(ctx context.Context, root string, input Input, run CommandRunner, extra ...string) error {
	return RunWithOutput(ctx, root, input, run, nil, nil, extra...)
}

// RunWithOutput is the testable implementation behind Run.
func RunWithOutput(ctx context.Context, root string, input Input, run CommandRunner, stdout io.Writer, stderr io.Writer, extra ...string) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	opts, specified, err := optionsFromInput(input, extra)
	if err != nil {
		return err
	}

	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = filepath.Join("hack", "config.yaml")
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
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
	if err = applyImageOverrides(&cfg.Image, opts, specified); err != nil {
		return err
	}
	if err = normalizeImageConfig(ctx, root, &cfg.Image); err != nil {
		return err
	}
	if opts.Preflight {
		return validateImageBuildRequest(cfg.Image, cfg.Build)
	}

	verbose, err := parseOptionalBool(opts.Verbose, false)
	if err != nil {
		return fmt.Errorf("parse verbose: %w", err)
	}
	runner := commandRunner{
		root:    root,
		verbose: verbose,
		run:     run,
	}

	if !opts.BuildOnly {
		err = validateImageBuildRequest(cfg.Image, cfg.Build)
	}
	if err != nil {
		return err
	}

	for _, target := range cfg.Build.Targets {
		binaryPath := buildOutputBinaryPath(root, cfg.Build, target)
		if err = validateExistingBinary(stdout, binaryPath); err != nil {
			return err
		}
		stagedBinaryPath := imageStagedBinaryPath(root, cfg.Build, target)
		if err = stageImageBinary(stdout, stderr, binaryPath, stagedBinaryPath); err != nil {
			return err
		}
	}

	if opts.BuildOnly {
		fmt.Fprintf(stdout, "✓ image build artifacts are ready: %s\n", filepath.Join(root, conventionImageBinaryRoot))
		return nil
	}

	imageRef := buildImageRef(cfg.Image)
	if err = buildDockerImage(ctx, stdout, root, cfg.Image, cfg.Build, runner, imageRef); err != nil {
		return err
	}
	if cfg.Build.MultiPlatform() {
		fmt.Fprintf(stdout, "✓ Docker image pushed: %s\n", imageRef)
		return nil
	}
	if cfg.Image.Push {
		dockerRunner := runner
		dockerRunner.verbose = true
		if err = dockerRunner.Run(ctx, ".", nil, "docker", "push", imageRef); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "✓ Docker image pushed: %s\n", imageRef)
		return nil
	}
	fmt.Fprintf(stdout, "✓ Docker image built: %s\n", imageRef)
	return nil
}

func optionsFromInput(input Input, extra []string) (cliOptions, map[string]bool, error) {
	opts := cliOptions{
		ConfigPath: input.GetDefault("config", filepath.Join("hack", "config.yaml")),
	}
	specified := map[string]bool{}
	optionKeys := map[string]*string{
		"image":       &opts.Image,
		"tag":         &opts.Tag,
		"registry":    &opts.Registry,
		"push":        &opts.Push,
		"platforms":   &opts.Platforms,
		"cgo-enabled": &opts.CGOEnabled,
		"output-dir":  &opts.OutputDir,
		"binary-name": &opts.BinaryName,
		"base-image":  &opts.BaseImage,
		"config":      &opts.ConfigPath,
		"verbose":     &opts.Verbose,
	}
	params := input.ParamMap()
	for key, target := range optionKeys {
		value, exists := params[key]
		if !exists {
			continue
		}
		*target = value
		specified[key] = true
	}
	if value, exists := params["v"]; exists {
		opts.Verbose = value
		specified["verbose"] = true
	}
	if input.Has("preflight") {
		preflight, err := input.Bool("preflight", false)
		if err != nil {
			return opts, nil, err
		}
		opts.Preflight = preflight
	}
	if input.Has("build_only") {
		buildOnly, err := input.Bool("build_only", false)
		if err != nil {
			return opts, nil, err
		}
		opts.BuildOnly = buildOnly
	}
	for _, item := range extra {
		switch item {
		case "--preflight":
			opts.Preflight = true
		case "--build-only":
			opts.BuildOnly = true
		case "":
			continue
		default:
			return opts, nil, fmt.Errorf("unsupported imagebuilder option %q", item)
		}
	}
	return opts, specified, nil
}

// Run executes one child process in a repository-relative directory.
func (r commandRunner) Run(ctx context.Context, dir string, _ []string, name string, args ...string) error {
	workingDir := filepath.Join(r.root, dir)
	return r.run(ctx, workingDir, name, args...)
}
