// This file implements the wasm command. It uses the wasmcmd suffix because
// Go treats files ending in _wasm.go as GOARCH=wasm-only build files.

package main

import (
	"context"
	"errors"
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
		outDir = filepath.Join(a.root, outDir)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create wasm output directory: %w", err)
	}

	if rawDir := input.Get("dir"); rawDir != "" {
		pluginDir, err := resolveWasmPluginDir(a.root, rawDir)
		if err != nil {
			return err
		}
		if err := preparePluginDirWorkspace(a.root, pluginDir); err != nil {
			return err
		}
		dryRun, err := input.Bool("dry-run", false)
		if err != nil {
			return err
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
	plugins, err := dynamicPlugins(a.root)
	if err != nil {
		return err
	}
	if len(plugins) == 0 {
		fmt.Fprintln(a.stdout, "No buildable dynamic wasm plugins found")
		return nil
	}

	dryRun, err := input.Bool("dry-run", false)
	if err != nil {
		return err
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

func resolveWasmPluginDir(root string, rawDir string) (string, error) {
	if strings.TrimSpace(rawDir) == "" {
		return "", errors.New("wasm dir cannot be empty")
	}
	dir := rawDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	clean, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", fmt.Errorf("resolve wasm dir %q: %w", rawDir, err)
	}
	info, err := os.Stat(clean)
	if err != nil {
		return "", fmt.Errorf("stat wasm dir %s: %w", clean, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("wasm dir is not a directory: %s", clean)
	}
	return clean, nil
}

// preparePluginDirWorkspace prepares the temporary official plugin workspace
// when an explicit dir points inside apps/lina-plugins.
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

// dynamicPlugins returns buildable official dynamic plugin IDs.
func dynamicPlugins(root string) ([]string, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	if err := plugins.RequireOfficialWorkspace(root); err != nil {
		return nil, err
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
