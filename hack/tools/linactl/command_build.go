// This file implements the build command for frontend assets and host binaries.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"linactl/internal/config"
	"linactl/internal/fileutil"
	"linactl/internal/toolutil"
)

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
		enabled, parseErr := toolutil.ParseBool(raw, false)
		if parseErr != nil {
			return parseErr
		}
		if enabled {
			cgoEnabled = "1"
		} else {
			cgoEnabled = "0"
		}
	}
	pluginsEnabled, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("clean build output directory: %w", err)
	}
	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create build output directory: %w", err)
	}

	fmt.Fprintln(a.stdout, "Building frontend...")
	if err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-vben"), Env: env, Quiet: !verbose}, "pnpm", "run", "build"); err != nil {
		return err
	}

	embedDir := filepath.Join(a.root, "apps", "lina-core", "internal", "packed", "public")
	if err = os.RemoveAll(embedDir); err != nil {
		return fmt.Errorf("clean frontend embed directory: %w", err)
	}
	if err = os.MkdirAll(embedDir, 0o755); err != nil {
		return fmt.Errorf("create frontend embed directory: %w", err)
	}
	if err = fileutil.CopyDirContents(filepath.Join(a.root, "apps", "lina-vben", "apps", "web-antd", "dist"), embedDir); err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Host frontend embedded assets generated")

	if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
		return err
	}
	if pluginsEnabled {
		if err = runWasm(ctx, a, commandInput{Params: map[string]string{"out": outputDir}}); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(a.stdout, "Skipping official plugin wasm build in host-only mode")
	}

	multiPlatform := len(targets) > 1
	for _, target := range targets {
		targetBinary := filepath.Join(outputDir, toolutil.ExecutableName(binaryName))
		if multiPlatform {
			targetBinary = filepath.Join(outputDir, target.OS+"_"+target.Arch, toolutil.ExecutableName(binaryName))
		}
		if err = os.MkdirAll(filepath.Dir(targetBinary), 0o755); err != nil {
			return fmt.Errorf("create backend output directory: %w", err)
		}
		fmt.Fprintf(a.stdout, "Building backend for %s/%s...\n", target.OS, target.Arch)
		targetEnv := toolutil.SetEnvValue(toolutil.SetEnvValue(toolutil.SetEnvValue(env, "CGO_ENABLED", cgoEnabled), "GOOS", target.OS), "GOARCH", target.Arch)
		err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core"), Env: targetEnv, Quiet: !verbose}, "go", "build", "-o", targetBinary, ".")
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Build complete: %s\n", toolutil.RelativePath(a.root, targetBinary))
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
