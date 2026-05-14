// This file implements manifest asset packing commands.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runPreparePackedAssets rebuilds the host manifest embed workspace.
func runPreparePackedAssets(_ context.Context, a *app, _ commandInput) error {
	sourceDir := filepath.Join(a.root, "apps", "lina-core", "manifest")
	targetDir := filepath.Join(a.root, "apps", "lina-core", "internal", "packed", "manifest")

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("clean packed manifest directory: %w", err)
	}

	dirs := []string{
		filepath.Join(targetDir, "config"),
		filepath.Join(targetDir, "sql"),
		filepath.Join(targetDir, "i18n"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(sourceDir, "config", "config.template.yaml"): filepath.Join(targetDir, "config", "config.template.yaml"),
		filepath.Join(sourceDir, "config", "metadata.yaml"):        filepath.Join(targetDir, "config", "metadata.yaml"),
	}
	for src, dst := range files {
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}

	if err := copyDirContents(filepath.Join(sourceDir, "sql"), filepath.Join(targetDir, "sql")); err != nil {
		return err
	}
	if err := copyDirContents(filepath.Join(sourceDir, "i18n"), filepath.Join(targetDir, "i18n")); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, ".gitkeep"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("write packed manifest .gitkeep: %w", err)
	}

	fmt.Fprintf(a.stdout, "packed manifest assets prepared: %s\n", relativePath(a.root, targetDir))
	return nil
}

// runWasm builds dynamic Wasm plugin artifacts or lists them in dry-run mode.
