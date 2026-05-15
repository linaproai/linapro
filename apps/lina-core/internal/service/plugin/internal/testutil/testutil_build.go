// This file contains repository-location helpers and external runtime artifact build helpers.

package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// bundledRuntimeSampleOnce ensures the bundled dynamic sample is built once per test process.
var bundledRuntimeSampleOnce sync.Once

// bundledRuntimeSampleErr stores a build failure captured by bundledRuntimeSampleOnce.
var bundledRuntimeSampleErr error

// bundledRuntimeSampleMissing reports whether the bundled sample plugin is absent.
var bundledRuntimeSampleMissing bool

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
				bundledRuntimeSampleMissing = true
				return
			}
			bundledRuntimeSampleErr = statErr
			return
		}

		builderDir := filepath.Join(repoRoot, "hack", "tools", "build-wasm")
		if err = ensureBuildWasmPluginWorkspace(repoRoot, pluginDir); err != nil {
			bundledRuntimeSampleErr = err
			return
		}
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
		cmd.Env = append(os.Environ(), "GOWORK="+selectBuildWasmGoWork(repoRoot, pluginDir))
		output, err := cmd.CombinedOutput()
		if err != nil {
			bundledRuntimeSampleErr = fmt.Errorf("run hack/tools/build-wasm failed: %w output=%s", err, string(output))
		}
	})

	if bundledRuntimeSampleMissing {
		t.Skip("official plugin workspace is not initialized")
	}
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
	if err = ensureBuildWasmPluginWorkspace(repoRoot, pluginDir); err != nil {
		t.Fatalf("failed to prepare temporary plugin workspace: %v", err)
	}
	cmd := exec.Command("go", "run", ".", "--plugin-dir", pluginDir, "--output-dir", outputDir)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+selectBuildWasmGoWork(repoRoot, pluginDir))
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

// selectBuildWasmGoWork chooses the Go workspace used to run the build-wasm
// tool itself. Guest runtime builds use temp/go.work.plugins when the plugin is
// inside the official plugin workspace.
func selectBuildWasmGoWork(repoRoot string, pluginDir string) string {
	return filepath.Join(repoRoot, "go.work")
}

// ensureBuildWasmPluginWorkspace mirrors linactl's temporary workspace setup
// for tests that invoke the build-wasm helper directly.
func ensureBuildWasmPluginWorkspace(repoRoot string, pluginDir string) error {
	officialRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	absolutePluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return fmt.Errorf("resolve plugin directory: %w", err)
	}
	relativePath, err := filepath.Rel(officialRoot, absolutePluginDir)
	if err != nil || relativePath == "." || strings.HasPrefix(relativePath, "..") {
		return nil
	}

	rootContent, err := os.ReadFile(filepath.Join(repoRoot, "go.work"))
	if err != nil {
		return fmt.Errorf("read root go.work: %w", err)
	}
	version := parseBuildWasmGoWorkVersion(string(rootContent))
	if version == "" {
		return fmt.Errorf("root go.work is missing a go version directive")
	}

	workspacePath := filepath.Join(repoRoot, "temp", "go.work.plugins")
	uses := make([]string, 0)
	seen := make(map[string]struct{})
	addUse := func(use string) {
		normalized := normalizeBuildWasmGoWorkUse(use)
		if normalized == "" || normalized == "apps/lina-plugins" || strings.HasPrefix(normalized, "apps/lina-plugins/") {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		uses = append(uses, normalized)
	}
	for _, use := range parseBuildWasmGoWorkUses(string(rootContent)) {
		addUse(use)
	}
	if aggregateUse, aggregateErr := ensureBuildWasmAggregateModule(repoRoot, officialRoot); aggregateErr != nil {
		return aggregateErr
	} else if aggregateUse != "" {
		normalized := normalizeBuildWasmGoWorkUse(aggregateUse)
		if normalized != "" {
			if _, ok := seen[normalized]; !ok {
				seen[normalized] = struct{}{}
				uses = append(uses, normalized)
			}
		}
	}
	pluginUses, err := buildWasmPluginGoWorkUses(repoRoot, officialRoot)
	if err != nil {
		return err
	}
	for _, use := range pluginUses {
		normalized := normalizeBuildWasmGoWorkUse(use)
		if normalized == "" || normalized == "apps/lina-plugins" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		uses = append(uses, normalized)
	}

	content, err := renderBuildWasmPluginGoWork(repoRoot, workspacePath, version, uses)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(workspacePath), 0o755); err != nil {
		return fmt.Errorf("create temporary plugin workspace directory: %w", err)
	}
	if err = os.WriteFile(workspacePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write temporary plugin workspace: %w", err)
	}
	return nil
}

// ensureBuildWasmAggregateModule resolves the module that satisfies the
// host's official source-plugin import bridge for plugin test builds.
func ensureBuildWasmAggregateModule(repoRoot string, officialRoot string) (string, error) {
	if moduleName, err := readBuildWasmGoModuleName(filepath.Join(officialRoot, "go.mod")); err == nil && moduleName == "lina-plugins" {
		if err = os.RemoveAll(filepath.Join(repoRoot, "temp", "official-plugins")); err != nil {
			return "", fmt.Errorf("clean stale test aggregate module: %w", err)
		}
		return filepath.ToSlash(filepath.Join("apps", "lina-plugins")), nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	moduleDir := filepath.Join(repoRoot, "temp", "official-plugins")
	if err := os.RemoveAll(moduleDir); err != nil {
		return "", fmt.Errorf("clean test aggregate module: %w", err)
	}
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return "", fmt.Errorf("create test aggregate module: %w", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte("module lina-plugins\n\ngo 1.25.0\n"), 0o644); err != nil {
		return "", fmt.Errorf("write test aggregate go.mod: %w", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "plugins.go"), []byte("package linaplugins\n"), 0o644); err != nil {
		return "", fmt.Errorf("write test aggregate package: %w", err)
	}
	return filepath.ToSlash(filepath.Join("temp", "official-plugins")), nil
}

// readBuildWasmGoModuleName reads the module directive from a go.mod file.
func readBuildWasmGoModuleName(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(stripBuildWasmGoWorkComment(line))
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("%s is missing a module directive", path)
}

// buildWasmPluginGoWorkUses discovers Go modules under the official plugin workspace.
func buildWasmPluginGoWorkUses(repoRoot string, officialRoot string) ([]string, error) {
	var uses []string
	err := filepath.WalkDir(officialRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "go.mod" {
			return nil
		}
		relativePath, relErr := filepath.Rel(repoRoot, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		uses = append(uses, filepath.ToSlash(relativePath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan official plugin Go modules: %w", err)
	}
	sort.Slice(uses, func(left int, right int) bool {
		leftDepth := strings.Count(uses[left], "/")
		rightDepth := strings.Count(uses[right], "/")
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return uses[left] < uses[right]
	})
	return uses, nil
}

// renderBuildWasmPluginGoWork writes a workspace relative to temp/go.work.plugins.
func renderBuildWasmPluginGoWork(repoRoot string, workspacePath string, version string, uses []string) (string, error) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "go %s\n", version)
	if len(uses) == 0 {
		builder.WriteString("\n")
		return builder.String(), nil
	}
	builder.WriteString("\nuse (\n")
	for _, use := range uses {
		usePath, err := renderBuildWasmGoWorkUsePath(repoRoot, workspacePath, use)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&builder, "\t%s\n", usePath)
	}
	builder.WriteString(")\n")
	return builder.String(), nil
}

// renderBuildWasmGoWorkUsePath renders one normalized module path for go.work.
func renderBuildWasmGoWorkUsePath(repoRoot string, workspacePath string, use string) (string, error) {
	modulePath := filepath.FromSlash(use)
	if !filepath.IsAbs(modulePath) {
		modulePath = filepath.Join(repoRoot, modulePath)
	}
	relativePath, err := filepath.Rel(filepath.Dir(workspacePath), modulePath)
	if err != nil {
		return "", fmt.Errorf("render temporary workspace path for %s: %w", use, err)
	}
	relativePath = filepath.ToSlash(relativePath)
	if !strings.HasPrefix(relativePath, ".") {
		relativePath = "./" + relativePath
	}
	if strings.ContainsAny(relativePath, " \t\"") {
		return strconv.Quote(relativePath), nil
	}
	return relativePath, nil
}

// parseBuildWasmGoWorkVersion extracts the go directive from workspace content.
func parseBuildWasmGoWorkVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(stripBuildWasmGoWorkComment(line))
		if len(fields) >= 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// parseBuildWasmGoWorkUses extracts use entries from root go.work content.
func parseBuildWasmGoWorkUses(content string) []string {
	var (
		uses       []string
		inUseBlock bool
	)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(stripBuildWasmGoWorkComment(line))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "use (") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if trimmed == ")" {
				inUseBlock = false
				continue
			}
			if use := firstBuildWasmGoWorkField(trimmed); use != "" {
				uses = append(uses, use)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "use ") {
			if use := firstBuildWasmGoWorkField(strings.TrimSpace(strings.TrimPrefix(trimmed, "use"))); use != "" && use != "(" {
				uses = append(uses, use)
			}
		}
	}
	return uses
}

// stripBuildWasmGoWorkComment removes simple line comments from go.work syntax.
func stripBuildWasmGoWorkComment(line string) string {
	if index := strings.Index(line, "//"); index >= 0 {
		return line[:index]
	}
	return line
}

// firstBuildWasmGoWorkField returns the first path-like token from one use line.
func firstBuildWasmGoWorkField(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	if unquoted, err := strconv.Unquote(fields[0]); err == nil {
		return unquoted
	}
	return fields[0]
}

// normalizeBuildWasmGoWorkUse maps a use path to repository-relative slash form.
func normalizeBuildWasmGoWorkUse(use string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(use)), "./")
}
