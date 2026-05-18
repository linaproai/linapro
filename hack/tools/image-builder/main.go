// Package main implements the Docker image build orchestration used by the
// repository-level make image target.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

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
