// This file implements dynamic plugin Wasm build commands.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

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

	workspace, err := inspectOfficialPluginWorkspace(a.root)
	if err != nil {
		return err
	}
	if err = requireOfficialPluginWorkspaceState(a.root, workspace); err != nil {
		return err
	}
	workspacePath, err := prepareOfficialPluginWorkspace(a.root, true, workspace)
	if err != nil {
		return err
	}
	env := officialPluginBuildEnv(a.root, a.env, true, workspacePath)
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
			Env: env,
		}, "go", "run", ".", "--plugin-dir", filepath.Join(a.root, "apps", "lina-plugins", plugin), "--output-dir", outDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// dynamicPlugins returns dynamic plugin IDs, optionally validating one plugin.
func dynamicPlugins(root string, plugin string) ([]string, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	if err := requireOfficialPluginWorkspace(root); err != nil {
		return nil, err
	}
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
