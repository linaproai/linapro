// This file implements Go module dependency cleanup commands.

package main

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// runTidy executes go mod tidy in each maintained Go module directory.
func runTidy(ctx context.Context, a *app, _ commandInput) error {
	modules, err := discoverGoModuleDirs(a.root)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return fmt.Errorf("no go.mod files discovered under %s", a.root)
	}
	for _, moduleDir := range modules {
		fmt.Fprintf(a.stdout, "==> go mod tidy (%s)\n", relativePath(a.root, moduleDir))
		if err = a.runCommand(ctx, commandOptions{Dir: moduleDir}, "go", "mod", "tidy"); err != nil {
			return err
		}
	}
	fmt.Fprintf(a.stdout, "tidied %d Go module(s)\n", len(modules))
	return nil
}

// discoverGoModuleDirs returns sorted module directories under the repository.
func discoverGoModuleDirs(root string) ([]string, error) {
	var modules []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if shouldSkipTidyDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "go.mod" {
			modules = append(modules, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan Go modules: %w", err)
	}
	sort.Strings(modules)
	return modules, nil
}

// shouldSkipTidyDir reports whether a directory is outside maintained source.
func shouldSkipTidyDir(root string, path string) bool {
	if path == root {
		return false
	}
	rel := relativePath(root, path)
	if rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		switch part {
		case ".git", ".tmp", "dist", "node_modules", "temp":
			return true
		}
	}
	return false
}
