// This file resolves filesystem paths from static config into stable runtime
// paths for host-local storage and development artifacts.

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

// repositoryRootSearchLimit bounds upward repository-root discovery so an
// unexpected working directory cannot scan the whole filesystem tree.
const repositoryRootSearchLimit = 12

// resolveRuntimePathWithDefault resolves a configured filesystem path, falling
// back to defaultPath when the configured value is blank.
func resolveRuntimePathWithDefault(configuredPath string, defaultPath string) string {
	cleanedPath := cleanConfigPath(configuredPath)
	if cleanedPath == "" {
		cleanedPath = cleanConfigPath(defaultPath)
	}
	return resolveRuntimePath(cleanedPath)
}

// resolveRuntimePath resolves one configured filesystem path against the
// repository root when possible, falling back to the process working directory.
func resolveRuntimePath(configuredPath string) string {
	cleanedPath := cleanConfigPath(configuredPath)
	if cleanedPath == "" || filepath.IsAbs(cleanedPath) {
		return cleanedPath
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return cleanedPath
	}
	return resolveRuntimePathFromWorkingDir(cleanedPath, workingDir)
}

// resolveRuntimePathFromWorkingDir resolves one relative path from a supplied
// working directory. Tests use this helper to verify root anchoring behavior.
func resolveRuntimePathFromWorkingDir(configuredPath string, workingDir string) string {
	cleanedPath := cleanConfigPath(configuredPath)
	if cleanedPath == "" || filepath.IsAbs(cleanedPath) {
		return cleanedPath
	}

	baseDir := cleanConfigPath(workingDir)
	if baseDir == "" {
		baseDir = "."
	}
	if absWorkingDir, err := filepath.Abs(baseDir); err == nil {
		baseDir = absWorkingDir
	}
	if repoRoot, err := findRepositoryRoot(baseDir); err == nil {
		baseDir = repoRoot
	}
	return filepath.Clean(filepath.Join(baseDir, cleanedPath))
}

// cleanConfigPath trims whitespace and normalizes path separators without
// changing whether the path is absolute or relative.
func cleanConfigPath(configuredPath string) string {
	trimmedPath := strings.TrimSpace(configuredPath)
	if trimmedPath == "" {
		return ""
	}
	return filepath.Clean(trimmedPath)
}

// findRepositoryRoot walks upward from startDir until it finds the LinaPro
// repository root markers used by local development and installation layouts.
func findRepositoryRoot(startDir string) (string, error) {
	current, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	current = filepath.Clean(current)

	for depth := 0; depth < repositoryRootSearchLimit; depth++ {
		if isRepositoryRoot(current) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", gerror.Newf("repository root not found from %s", startDir)
}

// isRepositoryRoot reports whether dir looks like the LinaPro repository root.
func isRepositoryRoot(dir string) bool {
	if gfile.Exists(filepath.Join(dir, "go.work")) &&
		gfile.Exists(filepath.Join(dir, "apps", "lina-core")) {
		return true
	}
	return gfile.Exists(filepath.Join(dir, "apps", "lina-core", "go.mod")) &&
		gfile.Exists(filepath.Join(dir, "apps", "lina-vben", "package.json"))
}
