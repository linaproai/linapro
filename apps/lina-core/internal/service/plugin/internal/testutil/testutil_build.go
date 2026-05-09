// This file contains repository-location helpers and external runtime artifact build helpers.

package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// bundledRuntimeSampleOnce ensures the bundled dynamic sample is built once per test process.
var bundledRuntimeSampleOnce sync.Once

// bundledRuntimeSampleErr stores a build failure captured by bundledRuntimeSampleOnce.
var bundledRuntimeSampleErr error

// RuntimeBuildOutput describes one artifact produced by the hack/tools/build-wasm helper in tests.
type RuntimeBuildOutput struct {
	// ArtifactPath is the on-disk path of the produced wasm artifact.
	ArtifactPath string
	// Content is the artifact byte content.
	Content []byte
}

// FindRepoRoot walks up from startDir until it locates the repository root.
func FindRepoRoot(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.work")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if strings.Contains(abs, string(filepath.Separator)) {
		for dir = abs; ; dir = filepath.Dir(dir) {
			if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
				if _, statErr2 := os.Stat(filepath.Join(filepath.Dir(dir), "go.work")); statErr2 == nil {
					return filepath.Dir(dir), nil
				}
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return abs, nil
}

// EnsureBundledRuntimeSampleArtifactForTests builds the bundled dynamic sample once for package tests.
func EnsureBundledRuntimeSampleArtifactForTests(t *testing.T) {
	t.Helper()

	bundledRuntimeSampleOnce.Do(func() {
		repoRoot, err := FindRepoRoot(".")
		if err != nil {
			bundledRuntimeSampleErr = err
			return
		}

		pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
		if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
			if os.IsNotExist(statErr) {
				return
			}
			bundledRuntimeSampleErr = statErr
			return
		}

		builderDir := filepath.Join(repoRoot, "hack", "tools", "build-wasm")
		cmd := exec.Command(
			"go",
			"run",
			".",
			"--plugin-dir",
			pluginDir,
			"--output-dir",
			testDynamicStorageDir,
		)
		cmd.Dir = builderDir
		cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
		output, err := cmd.CombinedOutput()
		if err != nil {
			bundledRuntimeSampleErr = fmt.Errorf("run hack/tools/build-wasm failed: %w output=%s", err, string(output))
		}
	})

	if bundledRuntimeSampleErr != nil {
		t.Fatalf("failed to prepare bundled dynamic sample: %v", bundledRuntimeSampleErr)
	}
}

// BuildRuntimeArtifactWithHackTool runs hack/tools/build-wasm for one plugin source directory.
func BuildRuntimeArtifactWithHackTool(t *testing.T, pluginDir string) *RuntimeBuildOutput {
	t.Helper()

	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	builderDir := filepath.Join(repoRoot, "hack", "tools", "build-wasm")
	outputDir := filepath.Join(t.TempDir(), "output")
	cmd := exec.Command("go", "run", ".", "--plugin-dir", pluginDir, "--output-dir", outputDir)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run hack/tools/build-wasm: %v output=%s", err, string(output))
	}

	type manifestIDHolder struct {
		ID string `yaml:"id"`
	}
	manifestContent, err := os.ReadFile(filepath.Join(pluginDir, "plugin.yaml"))
	if err != nil {
		t.Fatalf("failed to read plugin.yaml: %v", err)
	}
	var holder manifestIDHolder
	if err = yaml.Unmarshal(manifestContent, &holder); err != nil {
		t.Fatalf("failed to parse plugin.yaml: %v", err)
	}
	artifactPath := filepath.Join(outputDir, holder.ID+".wasm")
	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read built artifact: %v", err)
	}
	return &RuntimeBuildOutput{
		ArtifactPath: artifactPath,
		Content:      content,
	}
}
