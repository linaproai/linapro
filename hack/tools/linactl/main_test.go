// This file verifies linactl command parsing, plugin discovery, asset packing,
// and cross-platform path helper behavior.

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseCommandInputSupportsMakeStyleParams(t *testing.T) {
	input, err := parseCommandInput([]string{"confirm=init", "rebuild=true", "--platforms=linux/amd64,linux/arm64", "-h", "extra"})
	if err != nil {
		t.Fatalf("parseCommandInput returned error: %v", err)
	}

	if input.Get("confirm") != "init" {
		t.Fatalf("confirm mismatch: %q", input.Get("confirm"))
	}
	if input.Get("rebuild") != "true" {
		t.Fatalf("rebuild mismatch: %q", input.Get("rebuild"))
	}
	if input.Get("platforms") != "linux/amd64,linux/arm64" {
		t.Fatalf("platforms mismatch: %q", input.Get("platforms"))
	}
	input.Params["base_image"] = "alpine"
	if input.Get("base-image") != "alpine" {
		t.Fatalf("hyphenated key did not resolve normalized parameter")
	}
	if !input.HasBool("h") {
		t.Fatalf("expected -h to be parsed as true")
	}
	if len(input.Args) != 1 || input.Args[0] != "extra" {
		t.Fatalf("unexpected positional args: %#v", input.Args)
	}
}

func TestDynamicPluginsScansYAMLManifests(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "source-plugin", "plugin.yaml"), "type: source\n")
	writeFile(t, filepath.Join(pluginRoot, "dynamic-b", "plugin.yaml"), "type: dynamic\n")
	writeFile(t, filepath.Join(pluginRoot, "dynamic-a", "plugin.yaml"), "type: dynamic\n")

	plugins, err := dynamicPlugins(root, "")
	if err != nil {
		t.Fatalf("dynamicPlugins returned error: %v", err)
	}
	got := strings.Join(plugins, ",")
	if got != "dynamic-a,dynamic-b" {
		t.Fatalf("unexpected dynamic plugin list: %s", got)
	}
}

func TestPreparePackedAssetsCopiesExpectedFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir core: %v", err)
	}
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), "local: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en", "messages.json"), "{}\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	if err := runPreparePackedAssets(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPreparePackedAssets returned error: %v", err)
	}

	target := filepath.Join(root, "apps", "lina-core", "internal", "packed", "manifest")
	if !fileExists(filepath.Join(target, "config", "config.template.yaml")) {
		t.Fatalf("missing config.template.yaml")
	}
	if fileExists(filepath.Join(target, "config", "config.yaml")) {
		t.Fatalf("config.yaml should not be embedded")
	}
	if !fileExists(filepath.Join(target, "sql", "001.sql")) {
		t.Fatalf("missing sql file")
	}
	if !fileExists(filepath.Join(target, "i18n", "en", "messages.json")) {
		t.Fatalf("missing i18n file")
	}
	if !fileExists(filepath.Join(target, ".gitkeep")) {
		t.Fatalf("missing .gitkeep")
	}
}

func TestRunWasmResolvesExplicitRelativeOutputFromCurrentDirectory(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "hack", "tools", "build-wasm", "go.mod"), "module build-wasm\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "plugin.yaml"), "type: dynamic\n")

	workDir := filepath.Join(pluginRoot)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir work dir: %v", err)
	}
	previousWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := os.Chdir(previousWorkDir); cleanupErr != nil {
			t.Fatalf("restore work dir: %v", cleanupErr)
		}
	})
	if err = os.Chdir(workDir); err != nil {
		t.Fatalf("chdir work dir: %v", err)
	}

	var capturedArgs []string
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		return exec.Command("true")
	}

	if err = runWasm(context.Background(), application, commandInput{
		Params: map[string]string{
			"out": "../../temp/output",
			"p":   "plugin-demo-dynamic",
		},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	var outputDir string
	for i := 0; i < len(capturedArgs)-1; i++ {
		if capturedArgs[i] == "--output-dir" {
			outputDir = capturedArgs[i+1]
		}
	}
	expected := filepath.Clean(filepath.Join(workDir, "../../temp/output"))
	if !samePath(t, outputDir, expected) {
		t.Fatalf("expected output dir %s, got %s from args %q", expected, outputDir, strings.Join(capturedArgs, " "))
	}
}

func TestRunWasmUsesRepositoryTempOutputByDefault(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "hack", "tools", "build-wasm", "go.mod"), "module build-wasm\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "plugin.yaml"), "type: dynamic\n")

	var capturedArgs []string
	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		return exec.Command("true")
	}

	if err := runWasm(context.Background(), application, commandInput{
		Params: map[string]string{"p": "plugin-demo-dynamic"},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	var outputDir string
	for i := 0; i < len(capturedArgs)-1; i++ {
		if capturedArgs[i] == "--output-dir" {
			outputDir = capturedArgs[i+1]
		}
	}
	expected := filepath.Join(root, "temp", "output")
	if !samePath(t, outputDir, expected) {
		t.Fatalf("expected output dir %s, got %s from args %q", expected, outputDir, strings.Join(capturedArgs, " "))
	}
}

func TestExecutableNameAddsWindowsExtensionOnlyOnWindows(t *testing.T) {
	name := executableName("lina")
	if runtime.GOOS == "windows" {
		if name != "lina.exe" {
			t.Fatalf("expected windows executable name, got %s", name)
		}
		return
	}
	if name != "lina" {
		t.Fatalf("expected non-windows executable name, got %s", name)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func samePath(t *testing.T, left string, right string) bool {
	t.Helper()
	normalizedLeft, err := filepath.EvalSymlinks(left)
	if err != nil {
		normalizedLeft = filepath.Clean(left)
	}
	normalizedRight, err := filepath.EvalSymlinks(right)
	if err != nil {
		normalizedRight = filepath.Clean(right)
	}
	return normalizedLeft == normalizedRight
}
