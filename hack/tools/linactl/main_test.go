// This file verifies linactl command parsing, plugin discovery, asset packing,
// and cross-platform path helper behavior.

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
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

// TestCommandRegistryUsesDottedTestCommands guards the public test command names.
func TestCommandRegistryUsesDottedTestCommands(t *testing.T) {
	registry := commandRegistry()
	for _, name := range []string{"test.go", "test.host", "test.plugins", "test.scripts"} {
		if _, ok := registry[name]; !ok {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
	for _, name := range []string{"test-go", "test-host", "test-plugins", "test-scripts"} {
		if _, ok := registry[name]; ok {
			t.Fatalf("legacy command %q should not be registered", name)
		}
	}
}

// TestCommandRegistryUsesDottedImageBuildCommand guards the public image
// staging command name.
func TestCommandRegistryUsesDottedImageBuildCommand(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["image.build"]; !ok {
		t.Fatalf("expected command %q to be registered", "image.build")
	}
	if _, ok := registry["image-build"]; ok {
		t.Fatalf("legacy command %q should not be registered", "image-build")
	}
}

// TestCommandRegistryUsesDottedPackAssetsCommand guards the public manifest
// asset packing command name.
func TestCommandRegistryUsesDottedPackAssetsCommand(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["pack.assets"]; !ok {
		t.Fatalf("expected command %q to be registered", "pack.assets")
	}
	if _, ok := registry["prepare-packed-assets"]; ok {
		t.Fatalf("legacy command %q should not be registered", "prepare-packed-assets")
	}
	if normalized := normalizeCommandName("prepare-packed-assets"); normalized != "prepare-packed-assets" {
		t.Fatalf("legacy command name should not be normalized to a public alias, got %q", normalized)
	}
}

// TestCommandRegistryIncludesReleaseTagCheck verifies the public release
// governance command name.
func TestCommandRegistryIncludesReleaseTagCheck(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["release.tag.check"]; !ok {
		t.Fatalf("expected command %q to be registered", "release.tag.check")
	}
}

// TestPrintHelpHidesInternalCommands verifies root make help lists only
// repository-level commands by default.
func TestPrintHelpHidesInternalCommands(t *testing.T) {
	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))

	if err := application.printHelp(false); err != nil {
		t.Fatalf("printHelp returned error: %v", err)
	}
	output := stdout.String()
	for _, command := range []string{"cli", "cli.install", "ctrl", "dao"} {
		if strings.Contains(output, "\n  "+command+" ") {
			t.Fatalf("root help should hide internal command %q:\n%s", command, output)
		}
	}
	if !strings.Contains(output, "\n  build ") {
		t.Fatalf("root help should still list build command:\n%s", output)
	}
}

// TestPrintHelpAllIncludesInternalCommands verifies operators can still inspect
// the full linactl command list explicitly.
func TestPrintHelpAllIncludesInternalCommands(t *testing.T) {
	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))

	if err := application.printHelp(true); err != nil {
		t.Fatalf("printHelp returned error: %v", err)
	}
	output := stdout.String()
	for _, command := range []string{"cli", "cli.install", "ctrl", "dao"} {
		if !strings.Contains(output, "\n  "+command+" ") {
			t.Fatalf("full help should include internal command %q:\n%s", command, output)
		}
	}
}

// TestRunReleaseTagCheckAcceptsMatchingMetadataVersion verifies the happy path.
func TestRunReleaseTagCheckAcceptsMatchingMetadataVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: \"v1.2.3\"\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"tag": "v1.2.3"}})
	if err != nil {
		t.Fatalf("runReleaseTagCheck returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Release tag v1.2.3 matches framework.version") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

// TestRunReleaseTagCheckUsesGitHubRefNameFallback verifies tag workflow input.
func TestRunReleaseTagCheckUsesGitHubRefNameFallback(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2.3-rc.1\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = setEnvValue(os.Environ(), "GITHUB_REF_NAME", "v1.2.3-rc.1")

	err := runReleaseTagCheck(context.Background(), application, commandInput{})
	if err != nil {
		t.Fatalf("runReleaseTagCheck should use GITHUB_REF_NAME fallback: %v", err)
	}
}

// TestRunReleaseTagCheckPrintsValidatedFrameworkVersion verifies automation output.
func TestRunReleaseTagCheckPrintsValidatedFrameworkVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2.3\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"print_version": "1"}})
	if err != nil {
		t.Fatalf("runReleaseTagCheck returned error: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "v1.2.3" {
		t.Fatalf("expected printed version, got: %q", stdout.String())
	}
}

// TestRunReleaseTagCheckRejectsMismatchedTag verifies equality enforcement.
func TestRunReleaseTagCheckRejectsMismatchedTag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2.3\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"tag": "v1.2.4"}})
	if err == nil || !strings.Contains(err.Error(), `release tag "v1.2.4" must equal metadata framework.version "v1.2.3"`) {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
}

// TestRunReleaseTagCheckRejectsInvalidFrameworkVersion verifies format enforcement.
func TestRunReleaseTagCheckRejectsInvalidFrameworkVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"tag": "v1.2"}})
	if err == nil || !strings.Contains(err.Error(), "must match vMAJOR.MINOR.PATCH") {
		t.Fatalf("expected invalid version error, got: %v", err)
	}
}

// TestRunReleaseTagCheckRejectsMissingFrameworkVersion verifies metadata presence.
func TestRunReleaseTagCheckRejectsMissingFrameworkVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  name: LinaPro\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"tag": "v1.2.3"}})
	if err == nil || !strings.Contains(err.Error(), "metadata framework.version is empty") {
		t.Fatalf("expected missing version error, got: %v", err)
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

func TestPrintStatusTableIncludesDevelopmentServiceDetails(t *testing.T) {
	var stdout bytes.Buffer
	err := printStatusTable(&stdout, []serviceStatusRow{
		{
			Service: "Backend",
			Status:  "running",
			URL:     "http://127.0.0.1:8080/",
			PID:     "12345",
			PIDFile: "temp/pids/backend.pid",
			LogFile: "temp/lina-core.log",
		},
		{
			Service: "Frontend",
			Status:  "stopped",
			URL:     "http://127.0.0.1:5666/",
			PID:     "-",
			PIDFile: "temp/pids/frontend.pid",
			LogFile: "temp/lina-vben.log",
		},
	})
	if err != nil {
		t.Fatalf("printStatusTable returned error: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"+",
		"| Service",
		"| Backend",
		"| Frontend",
		"| running",
		"| stopped",
		"temp/pids/backend.pid",
		"temp/lina-vben.log",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected status table to contain %q, got:\n%s", expected, output)
		}
	}
}

// TestRunI18nCheckRunsBothChecksWhenScanFails verifies merged checks still
// report message coverage results when the scanner fails.
func TestRunI18nCheckRunsBothChecksWhenScanFails(t *testing.T) {
	root := t.TempDir()
	toolDir := filepath.Join(root, "hack", "tools", "runtime-i18n")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir runtime i18n tool dir: %v", err)
	}

	var calls []string
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		if len(args) == 3 && args[0] == "run" && args[1] == "." && args[2] == "scan" {
			return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
	}

	err := runI18nCheck(context.Background(), application, commandInput{})
	if err == nil {
		t.Fatalf("expected i18n check to fail when scan fails")
	}
	expected := []string{
		"go run . scan",
		"go run . messages",
	}
	if strings.Join(calls, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected i18n check calls:\n%s", strings.Join(calls, "\n"))
	}
}

// TestCommandRegistryUsesSingleI18nCheckEntry verifies the public command list
// exposes only the merged i18n check entry.
func TestCommandRegistryUsesSingleI18nCheckEntry(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["i18n.check"]; !ok {
		t.Fatalf("expected i18n.check command to be registered")
	}
	for _, removed := range []string{"check-runtime-i18n", "check-runtime-i18n-messages"} {
		if _, ok := registry[removed]; ok {
			t.Fatalf("expected old i18n command %s to be removed", removed)
		}
	}
}

func TestWaitHTTPAcceptsRedirectWithoutFollowingLoop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "./", http.StatusMovedPermanently)
	}))
	defer server.Close()

	pidFile := filepath.Join(t.TempDir(), "service.pid")
	if err := os.WriteFile(pidFile, []byte("12345"), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}
	if err := waitHTTP("Backend", server.URL+"/", pidFile, "service.log", time.Second); err != nil {
		t.Fatalf("waitHTTP should accept redirect readiness responses: %v", err)
	}
}

func TestRunDevStartsServicesAsAsyncProcessesAndPrintsFinalStatus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name == "go" && len(args) >= 1 && args[0] == "build" {
			return exec.Command("true")
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
	}
	application.waitHTTP = func(_ string, _ string, pidPath string, _ string, _ time.Duration) error {
		if readPID(pidPath) == 0 {
			return os.ErrNotExist
		}
		return nil
	}

	start := time.Now()
	if err := runDev(context.Background(), application, commandInput{
		Params: map[string]string{
			"skip_wasm": "true",
		},
	}); err != nil {
		t.Fatalf("runDev returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("runDev appears to have waited for service processes to exit: %s", elapsed)
	}
	for _, path := range []string{
		filepath.Join(root, "temp", "pids", "backend.pid"),
		filepath.Join(root, "temp", "pids", "frontend.pid"),
	} {
		pid := readPID(path)
		if pid == 0 {
			t.Fatalf("expected pid file %s to contain a service process id", path)
		}
		process, err := os.FindProcess(pid)
		if err == nil {
			if killErr := process.Kill(); killErr != nil {
				t.Logf("kill service process %d: %v", pid, killErr)
			}
		}
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove pid file %s: %v", path, err)
		}
	}

	output := stdout.String()
	statusTitleIndex := strings.LastIndex(output, "LinaPro Framework Status")
	if statusTitleIndex < 0 {
		t.Fatalf("expected final status title in output:\n%s", output)
	}
	finalOutput := output[statusTitleIndex:]
	for _, expected := range []string{
		"| Service",
		"| Backend",
		"| Frontend",
		"temp/pids/backend.pid",
		"temp/lina-vben.log",
	} {
		if !strings.Contains(finalOutput, expected) {
			t.Fatalf("expected final status output to contain %q, got:\n%s", expected, finalOutput)
		}
	}
}

func TestRunDevPassesRepositoryWasmOutputWhenPluginsEnabled(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n\nuse ./apps/lina-core\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "go.mod"), "module plugin-demo-dynamic\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "plugin.yaml"), "type: dynamic\n")
	if err := os.MkdirAll(filepath.Join(root, "hack", "tools", "build-wasm"), 0o755); err != nil {
		t.Fatalf("mkdir build-wasm workdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	var wasmOutputDir string
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name == "go" && len(args) >= 3 && args[0] == "run" && args[1] == "." {
			for index := 0; index < len(args)-1; index++ {
				if args[index] == "--output-dir" {
					wasmOutputDir = args[index+1]
				}
			}
			return exec.Command("true")
		}
		if name == "go" && len(args) >= 1 && args[0] == "build" {
			return exec.Command("true")
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
	}
	application.waitHTTP = func(_ string, _ string, pidPath string, _ string, _ time.Duration) error {
		if readPID(pidPath) == 0 {
			return os.ErrNotExist
		}
		return nil
	}

	if err := runDev(context.Background(), application, commandInput{Params: map[string]string{"plugins": "1"}}); err != nil {
		t.Fatalf("runDev returned error: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "temp", "pids", "backend.pid"),
		filepath.Join(root, "temp", "pids", "frontend.pid"),
	} {
		pid := readPID(path)
		if pid > 0 {
			if process, err := os.FindProcess(pid); err == nil {
				if killErr := process.Kill(); killErr != nil {
					t.Logf("kill service process %d: %v", pid, killErr)
				}
			}
		}
	}
	expected := filepath.Join(root, "temp", "output")
	if !samePath(t, wasmOutputDir, expected) {
		t.Fatalf("expected dev wasm output %s, got %s", expected, wasmOutputDir)
	}
}

func TestOfficialPluginBuildEnvSeparatesHostOnlyAndPluginFullModes(t *testing.T) {
	root := t.TempDir()
	input := []string{
		"GOWORK=/tmp/stale.work",
		"GOFLAGS=-mod=mod -tags=official_plugins,netgo -count=1",
		"LINAPRO_SOURCE_PLUGINS=1",
	}

	hostOnly := officialPluginBuildEnv(root, input, false, "")
	if got := envValue(hostOnly, "GOWORK"); got != "" {
		t.Fatalf("expected host-only GOWORK to be unset, got %q", got)
	}
	if got := envValue(hostOnly, "LINAPRO_SOURCE_PLUGINS"); got != "0" {
		t.Fatalf("expected host-only plugin frontend discovery to be disabled, got %q", got)
	}
	if got := envValue(hostOnly, "GOFLAGS"); strings.Contains(got, officialPluginsBuildTag) {
		t.Fatalf("expected host-only GOFLAGS to remove official plugin tag, got %q", got)
	}

	pluginWorkspace := filepath.Join(root, "temp", "go.work.plugins")
	pluginFull := officialPluginBuildEnv(root, hostOnly, true, pluginWorkspace)
	if got := envValue(pluginFull, "GOWORK"); got != pluginWorkspace {
		t.Fatalf("expected plugin-full GOWORK to use temporary plugin workspace, got %q", got)
	}
	if got := envValue(pluginFull, "LINAPRO_SOURCE_PLUGINS"); got != "1" {
		t.Fatalf("expected plugin-full frontend discovery to be enabled, got %q", got)
	}
	if got := envValue(pluginFull, "GOFLAGS"); !strings.Contains(got, "-tags=netgo,"+officialPluginsBuildTag) {
		t.Fatalf("expected plugin-full GOFLAGS to merge official plugin tag with existing tags, got %q", got)
	}
}

func TestResolveOfficialPluginBuildModeAutoDetectsWorkspace(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	enabled, workspace, err := resolveOfficialPluginBuildMode(root, commandInput{Params: map[string]string{}})
	if err != nil {
		t.Fatalf("resolveOfficialPluginBuildMode returned error: %v", err)
	}
	if !enabled {
		t.Fatalf("expected plugin mode to be auto-enabled when manifests exist")
	}
	if workspace.State != pluginWorkspaceStateReady {
		t.Fatalf("expected ready plugin workspace, got %s", workspace.State)
	}

	disabled, _, err := resolveOfficialPluginBuildMode(root, commandInput{Params: map[string]string{"plugins": "0"}})
	if err != nil {
		t.Fatalf("explicit host-only mode returned error: %v", err)
	}
	if disabled {
		t.Fatalf("expected explicit plugins=0 to disable plugin mode")
	}

	auto, _, err := resolveOfficialPluginBuildMode(root, commandInput{Params: map[string]string{"plugins": "auto"}})
	if err != nil {
		t.Fatalf("explicit plugins=auto returned error: %v", err)
	}
	if !auto {
		t.Fatalf("expected plugins=auto to use workspace detection")
	}
}

func TestOfficialPluginGoWorkUsesDiscoversPluginModules(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		t.Fatalf("inspectOfficialPluginWorkspace returned error: %v", err)
	}
	uses, err := officialPluginGoWorkUses(root, workspace)
	if err != nil {
		t.Fatalf("officialPluginGoWorkUses returned error: %v", err)
	}
	got := strings.Join(uses, ",")
	expected := "./apps/lina-plugins,./apps/lina-plugins/plugin-a,./apps/lina-plugins/plugin-b"
	if got != expected {
		t.Fatalf("unexpected plugin go.work uses: got %s expected %s", got, expected)
	}
}

// TestOfficialPluginBackendImportsDiscoversSourcePlugins verifies the generated
// aggregate module imports only source plugin backend registrations.
func TestOfficialPluginBackendImportsDiscoversSourcePlugins(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "source-b", "go.mod"), "module source-b\n")
	writeFile(t, filepath.Join(pluginRoot, "source-b", "plugin.yaml"), "id: source-b\ntype: source\n")
	writeFile(t, filepath.Join(pluginRoot, "source-b", "backend", "plugin.go"), "package backend\n")
	writeFile(t, filepath.Join(pluginRoot, "dynamic-a", "go.mod"), "module dynamic-a\n")
	writeFile(t, filepath.Join(pluginRoot, "dynamic-a", "plugin.yaml"), "id: dynamic-a\ntype: dynamic\n")
	writeFile(t, filepath.Join(pluginRoot, "dynamic-a", "backend", "plugin.go"), "package backend\n")
	writeFile(t, filepath.Join(pluginRoot, "source-a", "go.mod"), "module source-a\n")
	writeFile(t, filepath.Join(pluginRoot, "source-a", "plugin.yaml"), "id: source-a\ntype: source\n")
	writeFile(t, filepath.Join(pluginRoot, "source-a", "backend", "plugin.go"), "package backend\n")

	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		t.Fatalf("inspectOfficialPluginWorkspace returned error: %v", err)
	}
	imports, err := officialPluginBackendImports(workspace)
	if err != nil {
		t.Fatalf("officialPluginBackendImports returned error: %v", err)
	}

	var got []string
	for _, item := range imports {
		got = append(got, item.Import)
	}
	expected := "source-a/backend,source-b/backend"
	if strings.Join(got, ",") != expected {
		t.Fatalf("unexpected source plugin imports: got %s expected %s", strings.Join(got, ","), expected)
	}
}

// TestGoWorkspaceModulesSkipsGeneratedOfficialPluginAggregate verifies test.go
// does not run package tests from the generated bridge module itself.
func TestGoWorkspaceModulesSkipsGeneratedOfficialPluginAggregate(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	aggregateDir := officialPluginAggregateModuleDir(root)
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(aggregateDir, "go.mod"), "module lina-plugins\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "list -m -f {{.Dir}}" {
			t.Fatalf("unexpected module list command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, aggregateDir)
	}

	modules, err := goWorkspaceModules(context.Background(), application)
	if err != nil {
		t.Fatalf("goWorkspaceModules returned error: %v", err)
	}
	if len(modules) != 1 || !samePath(t, modules[0], coreDir) {
		t.Fatalf("unexpected workspace modules: %#v", modules)
	}
}

// TestGoWorkspaceModulesIncludesGoListOutputInErrors verifies CI failures keep
// the Go command's actionable workspace diagnostic instead of only exit status.
func TestGoWorkspaceModulesIncludesGoListOutputInErrors(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = t.TempDir()
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "list -m -f {{.Dir}}" {
			t.Fatalf("unexpected module list command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperPrintAndFail", "--")
	}

	_, err := goWorkspaceModules(context.Background(), application)
	if err == nil {
		t.Fatalf("expected goWorkspaceModules to return an error")
	}
	if !strings.Contains(err.Error(), "workspace diagnostic from go list") {
		t.Fatalf("expected go list output in error, got %v", err)
	}
}

// TestRunTestGoSerializesPackageExecution verifies CI uses one package process
// at a time while retaining the requested race and verbose flags.
func TestRunTestGoSerializesPackageExecution(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	aggregateDir := officialPluginAggregateModuleDir(root)
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(aggregateDir, "go.mod"), "module lina-plugins\n")

	var commands []string
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		command := name + " " + strings.Join(args, " ")
		commands = append(commands, command)
		switch command {
		case "go list -m -f {{.Dir}}":
			return exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, aggregateDir)
		case "go test -p=1 -race -v ./...":
			return exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		default:
			t.Fatalf("unexpected go command: %s", command)
			return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
		}
	}

	input := commandInput{Params: map[string]string{"plugins": "0", "race": "true", "verbose": "true"}}
	if err := runTestGo(context.Background(), application, input); err != nil {
		t.Fatalf("runTestGo returned error: %v", err)
	}

	got := strings.Join(commands, "\n")
	expected := "go list -m -f {{.Dir}}\ngo test -p=1 -race -v ./..."
	if got != expected {
		t.Fatalf("unexpected command sequence:\ngot:\n%s\nexpected:\n%s", got, expected)
	}
}

// TestDiscoverGoModuleDirsSkipsGeneratedAndDependencyDirs verifies tidy scans
// maintained source modules without entering generated or dependency trees.
func TestDiscoverGoModuleDirsSkipsGeneratedAndDependencyDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(root, "hack", "tools", "linactl", "go.mod"), "module linactl\n")
	writeFile(t, filepath.Join(root, "temp", "clone", "go.mod"), "module temp-clone\n")
	writeFile(t, filepath.Join(root, ".tmp", "spike", "go.mod"), "module spike\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "node_modules", "dep", "go.mod"), "module dep\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "dist", "go.mod"), "module dist\n")

	modules, err := discoverGoModuleDirs(root)
	if err != nil {
		t.Fatalf("discoverGoModuleDirs returned error: %v", err)
	}

	var rel []string
	for _, module := range modules {
		rel = append(rel, relativePath(root, module))
	}
	got := strings.Join(rel, ",")
	expected := "apps/lina-core,apps/lina-plugins/plugin-a,hack/tools/linactl"
	if got != expected {
		t.Fatalf("unexpected module directories: got %s expected %s", got, expected)
	}
}

// TestRunTidyExecutesGoModTidyForEachModule verifies tidy runs in each module
// directory so the adjacent go.sum file is the dependency checksum target.
func TestRunTidyExecutesGoModTidyForEachModule(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(root, "hack", "tools", "linactl", "go.mod"), "module linactl\n")

	capturePath := filepath.Join(root, "tidy-dirs.txt")
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = append(os.Environ(), "LINACTL_TEST_CAPTURE_DIRS="+capturePath)
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "mod tidy" {
			t.Fatalf("unexpected tidy command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperRecordWorkingDirectory", "--")
	}

	if err := runTidy(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runTidy returned error: %v", err)
	}

	content, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read captured tidy dirs: %v", err)
	}
	realRoot := root
	if evaluatedRoot, evalErr := filepath.EvalSymlinks(root); evalErr == nil {
		realRoot = evaluatedRoot
	}
	var dirs []string
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		if line != "" {
			realLine := line
			if evaluatedLine, evalErr := filepath.EvalSymlinks(line); evalErr == nil {
				realLine = evaluatedLine
			}
			dirs = append(dirs, relativePath(realRoot, realLine))
		}
	}
	got := strings.Join(dirs, ",")
	expected := "apps/lina-core,hack/tools/linactl"
	if got != expected {
		t.Fatalf("unexpected tidy directories: got %s expected %s", got, expected)
	}
}

func TestPrepareOfficialPluginWorkspaceWritesTemporaryWorkspace(t *testing.T) {
	root := t.TempDir()
	content := `go 1.25.0

use (
	./apps/lina-core
	./hack/tools/build-wasm
)
`
	writeFile(t, filepath.Join(root, "go.work"), content)
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		t.Fatalf("inspectOfficialPluginWorkspace returned error: %v", err)
	}
	workspacePath, err := prepareOfficialPluginWorkspace(root, true, workspace)
	if err != nil {
		t.Fatalf("prepareOfficialPluginWorkspace returned error: %v", err)
	}
	if workspacePath != filepath.Join(root, "temp", "go.work.plugins") {
		t.Fatalf("unexpected temporary workspace path: %s", workspacePath)
	}
	rootContent, err := os.ReadFile(filepath.Join(root, "go.work"))
	if err != nil {
		t.Fatalf("read root go.work: %v", err)
	}
	if string(rootContent) != content {
		t.Fatalf("root go.work changed unexpectedly:\n%s", string(rootContent))
	}
	pluginContent, err := os.ReadFile(workspacePath)
	if err != nil {
		t.Fatalf("read temporary plugin go.work: %v", err)
	}
	expected := `go 1.25.0

use (
	../apps/lina-core
	../hack/tools/build-wasm
	../apps/lina-plugins
	../apps/lina-plugins/plugin-a
	../apps/lina-plugins/plugin-b
)
`
	if string(pluginContent) != expected {
		t.Fatalf("unexpected temporary plugin go.work:\n%s", string(pluginContent))
	}
	if dirExists(filepath.Join(root, "temp", "official-plugins")) {
		t.Fatalf("expected existing official plugin root module to be reused without generated fallback")
	}
}

func TestPrepareOfficialPluginWorkspaceGeneratesFallbackAggregateModule(t *testing.T) {
	root := t.TempDir()
	content := `go 1.25.0

use (
	./apps/lina-core
	./hack/tools/build-wasm
)
`
	writeFile(t, filepath.Join(root, "go.work"), content)
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	workspace, err := inspectOfficialPluginWorkspace(root)
	if err != nil {
		t.Fatalf("inspectOfficialPluginWorkspace returned error: %v", err)
	}
	workspacePath, err := prepareOfficialPluginWorkspace(root, true, workspace)
	if err != nil {
		t.Fatalf("prepareOfficialPluginWorkspace returned error: %v", err)
	}
	pluginContent, err := os.ReadFile(workspacePath)
	if err != nil {
		t.Fatalf("read temporary plugin go.work: %v", err)
	}
	expected := `go 1.25.0

use (
	../apps/lina-core
	../hack/tools/build-wasm
	./official-plugins
	../apps/lina-plugins/plugin-a
	../apps/lina-plugins/plugin-b
)
`
	if string(pluginContent) != expected {
		t.Fatalf("unexpected fallback temporary plugin go.work:\n%s", string(pluginContent))
	}
	aggregateGoMod, err := os.ReadFile(filepath.Join(root, "temp", "official-plugins", "go.mod"))
	if err != nil {
		t.Fatalf("read aggregate go.mod: %v", err)
	}
	if string(aggregateGoMod) != "module lina-plugins\n\ngo 1.25.0\n" {
		t.Fatalf("unexpected aggregate go.mod:\n%s", string(aggregateGoMod))
	}
}

func TestValidateRepositoryToolingAllowsEmptyLegacyScriptDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\npushd \"%~dp0hack\\tools\\linactl\" || exit /b 1\r\ngo run . %*\r\n")
	if err := os.MkdirAll(filepath.Join(root, "hack", "scripts"), 0o755); err != nil {
		t.Fatalf("mkdir hack/scripts: %v", err)
	}

	if err := validateRepositoryTooling(root); err != nil {
		t.Fatalf("validateRepositoryTooling returned error: %v", err)
	}
}

func TestValidateRepositoryToolingRejectsLegacyScripts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\ngo run . %*\r\n")
	writeFile(t, filepath.Join(root, "hack", "scripts", "legacy.sh"), "#!/usr/bin/env bash\n")

	err := validateRepositoryTooling(root)
	if err == nil || !strings.Contains(err.Error(), "hack/scripts contains legacy script") {
		t.Fatalf("expected legacy script validation error, got %v", err)
	}
}

func TestValidateRepositoryToolingRejectsStaleMakeCmdWorkspaceOverride(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\nset GOWORK=off\r\ngo run . %*\r\n")

	err := validateRepositoryTooling(root)
	if err == nil || !strings.Contains(err.Error(), "must not force GOWORK=off") {
		t.Fatalf("expected stale GOWORK validation error, got %v", err)
	}
}

func TestHelperLongRunningProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	time.Sleep(5 * time.Second)
}

// TestHelperCommandSuccess exits successfully when invoked as a child command.
func TestHelperCommandSuccess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
}

// TestHelperCommandFailure exits with failure when invoked as a child command.
func TestHelperCommandFailure(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	os.Exit(1)
}

// TestHelperPrintAndFail prints a deterministic diagnostic and exits with
// failure for command-output error tests.
func TestHelperPrintAndFail(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	fmt.Fprintln(os.Stderr, "workspace diagnostic from go list")
	os.Exit(1)
}

// TestHelperRecordWorkingDirectory records the child process working directory
// for command execution tests.
func TestHelperRecordWorkingDirectory(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	capturePath := os.Getenv("LINACTL_TEST_CAPTURE_DIRS")
	if capturePath == "" {
		t.Fatalf("LINACTL_TEST_CAPTURE_DIRS is empty")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	file, err := os.OpenFile(capturePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open capture file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("close capture file: %v", closeErr)
		}
	}()
	if _, err = fmt.Fprintln(file, wd); err != nil {
		t.Fatalf("write capture file: %v", err)
	}
}

// TestHelperPrintWorkspaceModules prints supplied module directories for
// goWorkspaceModules command-output tests.
func TestHelperPrintWorkspaceModules(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-3] != "--" {
		return
	}
	for _, moduleDir := range os.Args[len(os.Args)-2:] {
		if _, err := fmt.Fprintln(os.Stdout, moduleDir); err != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
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

// writeFrontendDependencySentinel creates the Vite binary expected by
// ensureFrontendDeps so runDev unit tests do not require pnpm on PATH.
func writeFrontendDependencySentinel(t *testing.T, root string) {
	t.Helper()
	writeFile(t, viteCommand(root), "")
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
