// This file verifies linactl command parsing, plugin discovery, asset packing,
// and cross-platform path helper behavior.

package main

import (
	"bytes"
	"context"
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

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name == "go" && len(args) >= 1 && args[0] == "build" {
			return exec.Command("true")
		}
		if name == "pnpm" {
			return exec.Command(os.Args[0], "-test.run=TestHelperCreateFrontendDist", "--", root)
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
		filepath.Join(root, "apps", "lina-core", "internal", "packed", "public", "index.html"),
		filepath.Join(root, "apps", "lina-core", "internal", "packed", "public", ".gitkeep"),
	} {
		if !fileExists(path) {
			t.Fatalf("expected runDev to prepare frontend embed asset %s", path)
		}
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

func TestHelperLongRunningProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	time.Sleep(5 * time.Second)
}

func TestHelperCreateFrontendDist(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "--" {
		return
	}
	root := os.Args[len(os.Args)-1]
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "dist", "index.html"), "<div>dist</div>\n")
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
