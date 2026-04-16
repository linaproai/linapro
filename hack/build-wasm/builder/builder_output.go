// This file resolves artifact output locations, runtime workspace paths, and
// repository-root detection used by the standalone wasm builder.

package builder

import (
	"os"
	"path/filepath"
	"strings"
)

func buildRuntimeArtifactFileName(pluginID string) string {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return "plugin.wasm"
	}
	return normalizedID + ".wasm"
}

func buildRuntimeArtifactRelativePath(pluginID string) string {
	return filepath.Join("runtime", buildRuntimeArtifactFileName(pluginID))
}

func buildRuntimeBuildOutputRelativePath(pluginID string) string {
	return filepath.Join("temp", buildRuntimeArtifactFileName(pluginID))
}

func resolveRuntimeArtifactOutputDir(pluginDir string, outputDir string) (string, error) {
	normalizedOutputDir := strings.TrimSpace(outputDir)
	if normalizedOutputDir != "" {
		return resolveCustomOutputDir(normalizedOutputDir)
	}

	if repoRoot, ok := findRuntimeBuildRepoRoot(pluginDir); ok {
		return filepath.Join(repoRoot, defaultRuntimeOutputDir), nil
	}
	if workingDir, err := os.Getwd(); err == nil {
		if repoRoot, ok := findRuntimeBuildRepoRoot(workingDir); ok {
			return filepath.Join(repoRoot, defaultRuntimeOutputDir), nil
		}
	}
	return filepath.Join(pluginDir, "temp"), nil
}

func findRuntimeBuildRepoRoot(startDir string) (string, bool) {
	normalizedStartDir := strings.TrimSpace(startDir)
	if normalizedStartDir == "" {
		return "", false
	}

	current, err := filepath.Abs(normalizedStartDir)
	if err != nil {
		return "", false
	}
	current = filepath.Clean(current)

	for depth := 0; depth < 8; depth++ {
		if runtimeBuildRepoRootMatches(current) {
			return current, true
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", false
}

func runtimeBuildRepoRootMatches(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "apps", "lina-core", "go.mod")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "apps", "lina-vben", "package.json")); err != nil {
		return false
	}
	return true
}

func resolveGuestRuntimeOutputPath(pluginDir string, pluginID string, outputDir string) (string, error) {
	normalizedOutputDir := strings.TrimSpace(outputDir)
	if normalizedOutputDir != "" {
		resolvedOutputDir, err := resolveCustomOutputDir(normalizedOutputDir)
		if err != nil {
			return "", err
		}
		return filepath.Join(
			resolvedOutputDir,
			runtimeWorkspaceDirName,
			buildRuntimeWorkspaceKey(pluginID),
			"runtime-plugin.wasm",
		), nil
	}

	if repoRoot, ok := findRuntimeBuildRepoRoot(pluginDir); ok {
		return filepath.Join(
			repoRoot,
			defaultRuntimeOutputDir,
			runtimeWorkspaceDirName,
			buildRuntimeWorkspaceKey(pluginID),
			"runtime-plugin.wasm",
		), nil
	}
	if workingDir, err := os.Getwd(); err == nil {
		if repoRoot, ok := findRuntimeBuildRepoRoot(workingDir); ok {
			return filepath.Join(
				repoRoot,
				defaultRuntimeOutputDir,
				runtimeWorkspaceDirName,
				buildRuntimeWorkspaceKey(pluginID),
				"runtime-plugin.wasm",
			), nil
		}
	}

	tempDir, err := os.MkdirTemp("", "lina-build-wasm-")
	if err != nil {
		return "", err
	}
	return filepath.Join(tempDir, "runtime-plugin.wasm"), nil
}

func buildRuntimeWorkspaceKey(pluginID string) string {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return "plugin"
	}
	return normalizedID
}

func resolveCustomOutputDir(outputDir string) (string, error) {
	resolvedOutputDir, err := filepath.Abs(filepath.Clean(outputDir))
	if err != nil {
		return "", err
	}
	return resolvedOutputDir, nil
}
