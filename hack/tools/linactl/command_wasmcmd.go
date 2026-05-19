// This file implements the wasm command. It uses the wasmcmd suffix because
// Go treats files ending in _wasm.go as GOARCH=wasm-only build files.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"linactl/internal/fileutil"
	"linactl/internal/plugins"
	"linactl/internal/wasmbuilder"
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

	if pluginDir := input.Get("plugin_dir"); pluginDir != "" {
		if !filepath.IsAbs(pluginDir) {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve current directory for wasm plugin dir: %w", err)
			}
			pluginDir = filepath.Join(cwd, pluginDir)
		}
		if err := preparePluginDirWorkspace(a.root, pluginDir); err != nil {
			return err
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
		fmt.Fprintf(a.stdout, "Building dynamic wasm plugin from: %s\n", pluginDir)
		if dryRun {
			return nil
		}
		out, err := wasmbuilder.WriteRuntimeWasmArtifactFromSource(pluginDir, outDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "built runtime artifact: %s\n", out.ArtifactPath)
		return nil
	}

	workspace, err := plugins.InspectOfficialWorkspace(a.root)
	if err != nil {
		return err
	}
	if err = plugins.RequireOfficialWorkspaceState(a.root, workspace); err != nil {
		return err
	}
	if _, err = plugins.PrepareOfficialWorkspace(a.root, true, workspace); err != nil {
		return err
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
		out, err := wasmbuilder.WriteRuntimeWasmArtifactFromSource(filepath.Join(a.root, "apps", "lina-plugins", plugin), outDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "built runtime artifact: %s\n", out.ArtifactPath)
	}
	return nil
}

// preparePluginDirWorkspace prepares the temporary official plugin workspace
// when an explicit plugin_dir points inside apps/lina-plugins.
func preparePluginDirWorkspace(root string, pluginDir string) error {
	officialRoot := filepath.Join(root, "apps", "lina-plugins")
	relativePath, err := filepath.Rel(officialRoot, pluginDir)
	if err != nil || relativePath == "." || relativePath == "" || strings.HasPrefix(relativePath, "..") {
		return nil
	}
	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		return err
	}
	if err = plugins.RequireOfficialWorkspaceState(root, workspace); err != nil {
		return err
	}
	_, err = plugins.PrepareOfficialWorkspace(root, true, workspace)
	return err
}

// dynamicPlugins returns dynamic plugin IDs, optionally validating one plugin.
func dynamicPlugins(root string, plugin string) ([]string, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	if err := plugins.RequireOfficialWorkspace(root); err != nil {
		return nil, err
	}
	if plugin != "" {
		if err := plugins.ValidateDynamic(pluginRoot, plugin); err != nil {
			return nil, err
		}
		return []string{plugin}, nil
	}

	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		return nil, fmt.Errorf("read plugin directory: %w", err)
	}

	var pluginIDs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest := filepath.Join(pluginRoot, entry.Name(), "plugin.yaml")
		if !fileutil.FileExists(manifest) {
			continue
		}
		isDynamic, err := plugins.IsDynamic(manifest)
		if err != nil {
			return nil, err
		}
		if isDynamic {
			pluginIDs = append(pluginIDs, entry.Name())
		}
	}
	sort.Strings(pluginIDs)
	return pluginIDs, nil
}
