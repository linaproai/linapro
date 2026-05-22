// This file verifies linactl command parsing, plugin discovery, asset packing,
// and cross-platform path helper behavior.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"linactl/internal/devservice"
	"linactl/internal/fileutil"
	"linactl/internal/plugins"
	"linactl/internal/repository"
	"linactl/internal/runtimei18n"
	"linactl/internal/toolutil"
)

func init() {
	sql.Register("linactl_envcheck_test", envCheckSQLDriver{version: "14.13"})
	sql.Register("linactl_envcheck_error_test", envCheckSQLDriver{queryErr: errors.New("database unavailable")})
}

func TestParseCommandInputSupportsMakeStyleParams(t *testing.T) {
	input, err := parseCommandInput([]string{"confirm=init", "rebuild=true", "--platforms=linux/amd64,linux/arm64", "AGENT=ClaudeCode", "-h", "extra"})
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
	if input.Get("agent") != "ClaudeCode" {
		t.Fatalf("upper-case key did not resolve normalized parameter")
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

// TestCommandRegistryUsesEnvironmentCommands verifies environment setup moved
// out of the dev command namespace.
func TestCommandRegistryUsesEnvironmentCommands(t *testing.T) {
	registry := commandRegistry()
	for _, name := range []string{"env.check", "env.setup"} {
		if _, ok := registry[name]; !ok {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
	if _, ok := registry["dev.setup"]; ok {
		t.Fatalf("legacy command %q should not be registered", "dev.setup")
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
	for _, command := range []string{"env.check", "env.setup"} {
		if !strings.Contains(output, "\n  "+command+" ") {
			t.Fatalf("root help should include environment command %q:\n%s", command, output)
		}
	}
	if strings.Contains(output, "\n  dev.setup ") {
		t.Fatalf("root help should not include legacy dev.setup command:\n%s", output)
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
	application.env = toolutil.SetEnvValue(os.Environ(), "GITHUB_REF_NAME", "v1.2.3-rc.1")

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

// TestRunEnvCheckPrintsToolStatusTable verifies env.check reports every
// prerequisite in one stable table without depending on host-installed tools.
func TestRunEnvCheckPrintsToolStatusTable(t *testing.T) {
	root := t.TempDir()
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	tools := []envTool{
		{Name: "Go", Required: ">= 1.25.0", MinVersion: "1.25.0", SuccessRemark: "Go ok"},
		{Name: "Node.js", Required: ">= 20.19.0", MinVersion: "20.19.0", MissingRemark: "Install Node.js"},
		{Name: "pnpm", Required: ">= 10.0.0", MinVersion: "10.0.0"},
	}
	results := map[string]envProbeResult{
		"Go":      {Output: "go version go1.25.1 darwin/arm64"},
		"Node.js": {Missing: true},
		"pnpm":    {Output: "9.12.0"},
	}
	rows := collectEnvCheckRows(context.Background(), application, tools, func(_ context.Context, _ *app, tool envTool) envProbeResult {
		return results[tool.Name]
	})

	var stdout bytes.Buffer
	if err := printEnvCheckTable(&stdout, rows); err != nil {
		t.Fatalf("printEnvCheckTable returned error: %v", err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"+",
		"| Name",
		"| Remark",
		"Name",
		"Current Version",
		"Required Version",
		"Satisfied",
		"Go",
		"1.25.1",
		">= 1.25.0",
		"Yes",
		"Node.js",
		"not found",
		"Install Node.js",
		"pnpm",
		"9.12.0",
		"upgrade required",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected env.check table to contain %q:\n%s", expected, output)
		}
	}
}

// TestProbePostgreSQLServerVersionUsesCoreConfig verifies PostgreSQL checks
// query the configured server version through Go's database driver.
func TestProbePostgreSQLServerVersionUsesCoreConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), `
database:
  default:
    link: "pgsql:postgres:secret@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
`)
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("PostgreSQL probe must not execute external client %q %q", name, args)
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	connection, err := loadEnvPostgreSQLConnection(root)
	if err != nil {
		t.Fatalf("loadEnvPostgreSQLConnection returned error: %v", err)
	}
	if got := connection.dsn(); got != "postgres://postgres:secret@127.0.0.1:5432/linapro?sslmode=disable" {
		t.Fatalf("unexpected PostgreSQL DSN: %q", got)
	}

	output, err := queryPostgreSQLServerVersionWithDriver(context.Background(), "linactl_envcheck_test", connection)
	if err != nil {
		t.Fatalf("queryPostgreSQLServerVersionWithDriver returned error: %v", err)
	}
	if output != "14.13" {
		t.Fatalf("expected server version output, got %q", output)
	}
}

// TestProbePostgreSQLServerVersionFailureIncludesRemark verifies connection or
// query failures are reported in the PostgreSQL row remark.
func TestProbePostgreSQLServerVersionFailureIncludesRemark(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), `
database:
  default:
    link: "pgsql:postgres:secret@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
`)
	connection, err := loadEnvPostgreSQLConnection(root)
	if err != nil {
		t.Fatalf("loadEnvPostgreSQLConnection returned error: %v", err)
	}
	_, err = queryPostgreSQLServerVersionWithDriver(context.Background(), "linactl_envcheck_error_test", connection)
	if err == nil {
		t.Fatalf("expected failing test driver to return an error")
	}
	result := envProbeResult{
		Err:    err,
		Remark: "could not query PostgreSQL server version using apps/lina-core/manifest/config/config.yaml database.default.link: " + shortEnvOutput(err.Error()),
	}
	row := evaluateEnvTool(envTool{Name: "PostgreSQL", Required: ">= 14.0.0", MinVersion: "14.0.0"}, result)
	if row.Current != "unavailable" {
		t.Fatalf("expected unavailable current version, got %q", row.Current)
	}
	if !strings.Contains(row.Remark, "query PostgreSQL server version") {
		t.Fatalf("expected PostgreSQL query failure remark, got %q", row.Remark)
	}
}

// TestEvaluatePostgreSQLConfigFailureIncludesRemark verifies PostgreSQL probe
// failures explain why the server version could not be detected.
func TestEvaluatePostgreSQLConfigFailureIncludesRemark(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), `
database:
  default:
    link: "mysql:root:secret@tcp(127.0.0.1:3306)/linapro"
`)
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	tool := envTool{
		Name:       "PostgreSQL",
		ProbeKind:  envProbeKindPostgreSQLServer,
		Required:   ">= 14.0.0",
		MinVersion: "14.0.0",
	}
	row := evaluateEnvTool(tool, probeEnvTool(context.Background(), application, tool))
	if row.Current != "unavailable" {
		t.Fatalf("expected unavailable current version, got %q", row.Current)
	}
	for _, expected := range []string{
		"could not load PostgreSQL database link",
		"configured database type",
		"not PostgreSQL",
	} {
		if !strings.Contains(row.Remark, expected) {
			t.Fatalf("expected PostgreSQL failure remark to contain %q:\n%s", expected, row.Remark)
		}
	}
}

// TestRunEnvSetupInstallsFrontendAndPlaywright verifies env.setup keeps the
// former setup command's dependency installation behavior.
func TestRunEnvSetupInstallsFrontendAndPlaywright(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "hack", "tests"), 0o755); err != nil {
		t.Fatalf("mkdir test workspace: %v", err)
	}
	capturePath := filepath.Join(root, "env-setup-dirs.txt")
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = append(os.Environ(), "LINACTL_TEST_CAPTURE_DIRS="+capturePath)
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	var commands []string
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		commands = append(commands, name+" "+strings.Join(args, " "))
		return exec.Command(os.Args[0], "-test.run=TestHelperRecordWorkingDirectory", "--")
	}

	if err := runEnvSetup(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runEnvSetup returned error: %v", err)
	}

	got := strings.Join(commands, "\n")
	expected := "pnpm install\npnpm exec playwright install --with-deps chromium"
	if got != expected {
		t.Fatalf("unexpected env.setup commands:\ngot:\n%s\nexpected:\n%s", got, expected)
	}
	content, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read captured setup dirs: %v", err)
	}
	if !strings.Contains(string(content), filepath.Join(root, "apps", "lina-vben")) {
		t.Fatalf("env.setup should install frontend deps in apps/lina-vben:\n%s", string(content))
	}
	if !strings.Contains(string(content), filepath.Join(root, "hack", "tests")) {
		t.Fatalf("env.setup should install Playwright in hack/tests:\n%s", string(content))
	}
}

// TestRunCommandReportsMissingToolBeforeExecution verifies command execution
// keeps actionable PATH diagnostics without invoking the child process.
func TestRunCommandReportsMissingToolBeforeExecution(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("%s not found", name)
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("missing tool should not execute child command: %s %s", name, strings.Join(args, " "))
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	err := application.runCommand(context.Background(), commandOptions{}, "pnpm", "install")
	if err == nil {
		t.Fatalf("expected missing tool error")
	}
	expected := `required tool "pnpm" is not available in PATH while running pnpm install`
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected missing tool diagnostic %q, got %v", expected, err)
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
	if !fileutil.FileExists(filepath.Join(target, "config", "config.template.yaml")) {
		t.Fatalf("missing config.template.yaml")
	}
	if fileutil.FileExists(filepath.Join(target, "config", "config.yaml")) {
		t.Fatalf("config.yaml should not be embedded")
	}
	if !fileutil.FileExists(filepath.Join(target, "sql", "001.sql")) {
		t.Fatalf("missing sql file")
	}
	if !fileutil.FileExists(filepath.Join(target, "i18n", "en", "messages.json")) {
		t.Fatalf("missing i18n file")
	}
	if !fileutil.FileExists(filepath.Join(target, ".gitkeep")) {
		t.Fatalf("missing .gitkeep")
	}
}

// TestEnsurePackedPublicPlaceholderCreatesGitkeep verifies build refreshes can
// recreate the tracked frontend embed placeholder after cleaning generated files.
func TestEnsurePackedPublicPlaceholderCreatesGitkeep(t *testing.T) {
	root := t.TempDir()
	embedDir := filepath.Join(root, "apps", "lina-core", "internal", "packed", "public")
	if err := os.MkdirAll(embedDir, 0o755); err != nil {
		t.Fatalf("mkdir packed public dir: %v", err)
	}

	if err := ensurePackedPublicPlaceholder(embedDir); err != nil {
		t.Fatalf("ensurePackedPublicPlaceholder returned error: %v", err)
	}

	if !fileutil.FileExists(filepath.Join(embedDir, packedPublicPlaceholderName)) {
		t.Fatalf("missing packed public placeholder")
	}
}

func TestRunWasmResolvesExplicitRelativeOutputFromCurrentDirectory(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeDynamicPluginManifest(t, filepath.Join(pluginRoot, "linapro-demo-dynamic"), "linapro-demo-dynamic")

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

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	if err = runWasm(context.Background(), application, commandInput{
		Params: map[string]string{
			"out": "../../temp/output",
			"p":   "linapro-demo-dynamic",
		},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	expected := filepath.Clean(filepath.Join(workDir, "../../temp/output"))
	artifactPath := filepath.Join(expected, "linapro-demo-dynamic.wasm")
	if !fileutil.FileExists(artifactPath) {
		t.Fatalf("expected wasm artifact at %s", artifactPath)
	}
}

func TestRunWasmUsesRepositoryTempOutputByDefault(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeDynamicPluginManifest(t, filepath.Join(pluginRoot, "linapro-demo-dynamic"), "linapro-demo-dynamic")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	if err := runWasm(context.Background(), application, commandInput{
		Params: map[string]string{"p": "linapro-demo-dynamic"},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	expected := filepath.Join(root, "temp", "output")
	artifactPath := filepath.Join(expected, "linapro-demo-dynamic.wasm")
	if !fileutil.FileExists(artifactPath) {
		t.Fatalf("expected wasm artifact at %s", artifactPath)
	}
}

func TestExecutableNameAddsWindowsExtensionOnlyOnWindows(t *testing.T) {
	name := toolutil.ExecutableName("lina")
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
	err := devservice.PrintStatusTable(&stdout, []devservice.StatusRow{
		{
			Service: "Backend",
			Status:  "running",
			URL:     "http://127.0.0.1:9120/",
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
		t.Fatalf("devservice.PrintStatusTable returned error: %v", err)
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
	writeFile(t, filepath.Join(root, "apps", "lina-core", "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "package.json"), "{}\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "internal", "service", "demo", "demo.go"), "package demo\n\nfunc f() error { return errors.New(\"中文错误\") }\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "zh-CN", "framework.json"), "{\"framework\":{\"name\":\"LinaPro\"}}\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{\"framework\":{\"name\":\"LinaPro\"}}\n")

	var stdout bytes.Buffer
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.stdout = &stdout

	err := runI18nCheck(context.Background(), application, commandInput{})
	if err == nil {
		t.Fatalf("expected i18n check to fail when scan fails")
	}
	output := stdout.String()
	for _, expected := range []string{
		"Runtime i18n scan found",
		"Runtime i18n message coverage passed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected i18n check output to contain %q, got:\n%s", expected, output)
		}
	}
}

// TestRunI18nCheckUsesConsolidatedAllowlist verifies i18n.check reads the
// allowlist from the linactl internal runtime i18n component.
func TestRunI18nCheckUsesConsolidatedAllowlist(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "package.json"), "{}\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "internal", "service", "demo", "demo.go"), "package demo\n\nfunc f() error { return errors.New(\"中文错误\") }\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "zh-CN", "framework.json"), "{\"framework\":{\"name\":\"LinaPro\"}}\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{\"framework\":{\"name\":\"LinaPro\"}}\n")
	writeFile(t, filepath.Join(root, "hack", "tools", "linactl", "internal", "runtimei18n", "allowlist.json"), "{\"entries\":[{\"path\":\"apps/lina-core/internal/service/demo/demo.go\",\"rule\":\"go-caller-error-han\",\"category\":\"UserMessage\",\"reason\":\"test allowlist\",\"scope\":\"unit test\"}]}\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	if err := runI18nCheck(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("expected allowlisted i18n check to pass, got error: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "allowlist hits: 1") {
		t.Fatalf("expected consolidated allowlist to be used, got:\n%s", stdout.String())
	}
}

// TestRuntimeI18nSubcommandRejectsMissingRepoRoot verifies the internal
// component validates direct invocations from command wrappers.
func TestRuntimeI18nSubcommandRejectsMissingRepoRoot(t *testing.T) {
	var stdout bytes.Buffer
	exitCode, err := runtimei18n.Run("", []string{"messages"}, &stdout)
	if err == nil {
		t.Fatal("expected missing repository root to fail")
	}
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
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
	if err := devservice.WaitHTTP("Backend", server.URL+"/", pidFile, "service.log", time.Second); err != nil {
		t.Fatalf("devservice.WaitHTTP should accept redirect readiness responses: %v", err)
	}
}

func TestRunDevStartsServicesAsAsyncProcessesAndPrintsFinalStatus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	// portcheck.Verify 在 runDev 入口校验后端 server.address 与前端 vite proxy
	// target 是否与 defaultBackendPort 对齐，因此用例需要自带一份对齐到 9120
	// 的最小夹具，保持单测自包含、顺序无关。
	// portcheck.Verify runs at the start of runDev and requires the backend
	// server.address and the frontend vite proxy target to align with the
	// supplied backend port. The test owns minimal fixtures aligned to 9120
	// so the test stays self-contained and order independent.
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), "server:\n  address: \":9120\"\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "vite.config.mts"), "proxy: { '/api': { target: 'http://localhost:9120' } }\n")
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
		if devservice.ReadPID(pidPath) == 0 {
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
		pid := devservice.ReadPID(path)
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
	// 与上方 runDev 用例同样需要自带对齐到默认 backend 端口的最小夹具，使
	// portcheck.Verify 在测试沙盒中通过。
	// Self-contained fixtures aligned to defaultBackendPort so portcheck.Verify
	// passes inside the test sandbox.
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), "server:\n  address: \":9120\"\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "vite.config.mts"), "proxy: { '/api': { target: 'http://localhost:9120' } }\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "linapro-demo-dynamic", "go.mod"), "module linapro-demo-dynamic\n")
	writeDynamicPluginManifest(t, filepath.Join(pluginRoot, "linapro-demo-dynamic"), "linapro-demo-dynamic")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name == "go" && len(args) >= 1 && args[0] == "build" {
			return exec.Command("true")
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
	}
	application.waitHTTP = func(_ string, _ string, pidPath string, _ string, _ time.Duration) error {
		if devservice.ReadPID(pidPath) == 0 {
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
		pid := devservice.ReadPID(path)
		if pid > 0 {
			if process, err := os.FindProcess(pid); err == nil {
				if killErr := process.Kill(); killErr != nil {
					t.Logf("kill service process %d: %v", pid, killErr)
				}
			}
		}
	}
	expected := filepath.Join(root, "temp", "output")
	if !fileutil.FileExists(filepath.Join(expected, "linapro-demo-dynamic.wasm")) {
		t.Fatalf("expected dev wasm artifact under %s", expected)
	}
}

func TestOfficialPluginBuildEnvSeparatesHostOnlyAndPluginFullModes(t *testing.T) {
	root := t.TempDir()
	input := []string{
		"GOWORK=/tmp/stale.work",
		"GOFLAGS=-mod=mod -tags=official_plugins,netgo -count=1",
		"LINAPRO_SOURCE_PLUGINS=1",
	}

	hostOnly := plugins.BuildEnv(root, input, false, "")
	if got := toolutil.EnvValue(hostOnly, "GOWORK"); got != "" {
		t.Fatalf("expected host-only GOWORK to be unset, got %q", got)
	}
	if got := toolutil.EnvValue(hostOnly, "LINAPRO_SOURCE_PLUGINS"); got != "0" {
		t.Fatalf("expected host-only plugin frontend discovery to be disabled, got %q", got)
	}
	if got := toolutil.EnvValue(hostOnly, "GOFLAGS"); strings.Contains(got, plugins.OfficialBuildTag) {
		t.Fatalf("expected host-only GOFLAGS to remove official plugin tag, got %q", got)
	}

	pluginWorkspace := filepath.Join(root, "temp", "go.work.plugins")
	pluginFull := plugins.BuildEnv(root, hostOnly, true, pluginWorkspace)
	if got := toolutil.EnvValue(pluginFull, "GOWORK"); got != pluginWorkspace {
		t.Fatalf("expected plugin-full GOWORK to use temporary plugin workspace, got %q", got)
	}
	if got := toolutil.EnvValue(pluginFull, "LINAPRO_SOURCE_PLUGINS"); got != "1" {
		t.Fatalf("expected plugin-full frontend discovery to be enabled, got %q", got)
	}
	if got := toolutil.EnvValue(pluginFull, "GOFLAGS"); !strings.Contains(got, "-tags=netgo,"+plugins.OfficialBuildTag) {
		t.Fatalf("expected plugin-full GOFLAGS to merge official plugin tag with existing tags, got %q", got)
	}
}

func TestResolveOfficialPluginBuildModeAutoDetectsWorkspace(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	enabled, workspace, err := plugins.ResolveBuildMode(root, commandInput{Params: map[string]string{}})
	if err != nil {
		t.Fatalf("plugins.ResolveBuildMode returned error: %v", err)
	}
	if !enabled {
		t.Fatalf("expected plugin mode to be auto-enabled when manifests exist")
	}
	if workspace.State != plugins.WorkspaceStateReady {
		t.Fatalf("expected ready plugin workspace, got %s", workspace.State)
	}

	disabled, _, err := plugins.ResolveBuildMode(root, commandInput{Params: map[string]string{"plugins": "0"}})
	if err != nil {
		t.Fatalf("explicit host-only mode returned error: %v", err)
	}
	if disabled {
		t.Fatalf("expected explicit plugins=0 to disable plugin mode")
	}

	auto, _, err := plugins.ResolveBuildMode(root, commandInput{Params: map[string]string{"plugins": "auto"}})
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

	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		t.Fatalf("plugins.InspectOfficialWorkspace returned error: %v", err)
	}
	uses, err := plugins.GoWorkUses(root, workspace)
	if err != nil {
		t.Fatalf("plugins.GoWorkUses returned error: %v", err)
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

	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		t.Fatalf("plugins.InspectOfficialWorkspace returned error: %v", err)
	}
	imports, err := plugins.BackendImports(workspace)
	if err != nil {
		t.Fatalf("plugins.BackendImports returned error: %v", err)
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
	aggregateDir := plugins.AggregateModuleDir(root)
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

// TestGoWorkspaceModulesIncludesStdoutDiagnosticInErrors verifies failure
// diagnostics are preserved for tools that write errors to stdout.
func TestGoWorkspaceModulesIncludesStdoutDiagnosticInErrors(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = t.TempDir()
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "list -m -f {{.Dir}}" {
			t.Fatalf("unexpected module list command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperPrintStdoutAndFail", "--")
	}

	_, err := goWorkspaceModules(context.Background(), application)
	if err == nil {
		t.Fatalf("expected goWorkspaceModules to return an error")
	}
	if !strings.Contains(err.Error(), "stdout diagnostic from go list") {
		t.Fatalf("expected stdout diagnostic in error, got %v", err)
	}
}

// TestRunTestGoSerializesPackageExecution verifies CI uses one package process
// at a time while retaining the requested race and verbose flags for packages
// that actually contain Go tests.
func TestRunTestGoSerializesPackageExecution(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	aggregateDir := plugins.AggregateModuleDir(root)
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
		case "go list -json ./...":
			return exec.Command(os.Args[0], "-test.run=TestHelperPrintGoListPackages", "--")
		case "go test -p=1 -race -v lina-core/internal/service/plugin":
			return exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		case "go test -p=1 -race -run ^$ lina-core/internal/model":
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
	expected := "go list -m -f {{.Dir}}\ngo list -json ./...\ngo test -p=1 -race -v lina-core/internal/service/plugin\ngo test -p=1 -race -run ^$ lina-core/internal/model"
	if got != expected {
		t.Fatalf("unexpected command sequence:\ngot:\n%s\nexpected:\n%s", got, expected)
	}
}

// TestGoTestModulePlanForDirSeparatesTestAndCompilePackages verifies package
// planning only sends packages with test files through the unit-test command.
func TestGoTestModulePlanForDirSeparatesTestAndCompilePackages(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "apps", "lina-core")
	writeFile(t, filepath.Join(moduleDir, "go.mod"), "module lina-core\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "list -json ./..." {
			t.Fatalf("unexpected package list command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperPrintGoListPackages", "--")
	}

	plan, err := goTestModulePlanForDir(context.Background(), application, moduleDir)
	if err != nil {
		t.Fatalf("goTestModulePlanForDir returned error: %v", err)
	}
	if !samePath(t, plan.ModuleDir, moduleDir) {
		t.Fatalf("unexpected module dir: %s", plan.ModuleDir)
	}
	if got := strings.Join(plan.TestPackages, ","); got != "lina-core/internal/service/plugin" {
		t.Fatalf("unexpected test packages: %s", got)
	}
	if got := strings.Join(plan.CompilePackages, ","); got != "lina-core/internal/model" {
		t.Fatalf("unexpected compile packages: %s", got)
	}
}

// TestGoTestModulePlanForDirIgnoresStderrDiagnostics verifies Go discovery
// parses only stdout JSON so harmless go: diagnostics on stderr do not corrupt
// the package stream.
func TestGoTestModulePlanForDirIgnoresStderrDiagnostics(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "apps", "lina-core")
	writeFile(t, filepath.Join(moduleDir, "go.mod"), "module lina-core\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name != "go" || strings.Join(args, " ") != "list -json ./..." {
			t.Fatalf("unexpected package list command: %s %s", name, strings.Join(args, " "))
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperPrintGoListPackagesWithStderr", "--")
	}

	plan, err := goTestModulePlanForDir(context.Background(), application, moduleDir)
	if err != nil {
		t.Fatalf("goTestModulePlanForDir returned error: %v", err)
	}
	if got := strings.Join(plan.TestPackages, ","); got != "lina-core/internal/service/plugin" {
		t.Fatalf("unexpected test packages: %s", got)
	}
	if got := strings.Join(plan.CompilePackages, ","); got != "lina-core/internal/model" {
		t.Fatalf("unexpected compile packages: %s", got)
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
		rel = append(rel, toolutil.RelativePath(root, module))
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
			dirs = append(dirs, toolutil.RelativePath(realRoot, realLine))
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
	./hack/tools/linactl
)
`
	writeFile(t, filepath.Join(root, "go.work"), content)
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		t.Fatalf("plugins.InspectOfficialWorkspace returned error: %v", err)
	}
	workspacePath, err := plugins.PrepareOfficialWorkspace(root, true, workspace)
	if err != nil {
		t.Fatalf("plugins.PrepareOfficialWorkspace returned error: %v", err)
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
	../hack/tools/linactl
	../apps/lina-plugins
	../apps/lina-plugins/plugin-a
	../apps/lina-plugins/plugin-b
)
`
	if string(pluginContent) != expected {
		t.Fatalf("unexpected temporary plugin go.work:\n%s", string(pluginContent))
	}
	if fileutil.DirExists(filepath.Join(root, "temp", "official-plugins")) {
		t.Fatalf("expected existing official plugin root module to be reused without generated fallback")
	}
}

func TestPrepareOfficialPluginWorkspaceGeneratesFallbackAggregateModule(t *testing.T) {
	root := t.TempDir()
	content := `go 1.25.0

use (
	./apps/lina-core
	./hack/tools/linactl
)
`
	writeFile(t, filepath.Join(root, "go.work"), content)
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\n")

	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		t.Fatalf("plugins.InspectOfficialWorkspace returned error: %v", err)
	}
	workspacePath, err := plugins.PrepareOfficialWorkspace(root, true, workspace)
	if err != nil {
		t.Fatalf("plugins.PrepareOfficialWorkspace returned error: %v", err)
	}
	pluginContent, err := os.ReadFile(workspacePath)
	if err != nil {
		t.Fatalf("read temporary plugin go.work: %v", err)
	}
	expected := `go 1.25.0

use (
	../apps/lina-core
	../hack/tools/linactl
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

	if err := repository.ValidateTooling(root, commandNames()); err != nil {
		t.Fatalf("repository.ValidateTooling returned error: %v", err)
	}
}

func TestValidateRepositoryToolingRejectsLegacyScripts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\ngo run . %*\r\n")
	writeFile(t, filepath.Join(root, "hack", "scripts", "legacy.sh"), "#!/usr/bin/env bash\n")

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), "hack/scripts contains legacy script") {
		t.Fatalf("expected legacy script validation error, got %v", err)
	}
}

func TestValidateRepositoryToolingRejectsStaleMakeCmdWorkspaceOverride(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\nset GOWORK=off\r\ngo run . %*\r\n")

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), "must not force GOWORK=off") {
		t.Fatalf("expected stale GOWORK validation error, got %v", err)
	}
}

func TestValidateLinactlCommandFilesAcceptsRepositoryCommands(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	if err = repository.ValidateLinactlCommandFiles(root, commandNames()); err != nil {
		t.Fatalf("repository.ValidateLinactlCommandFiles returned error: %v", err)
	}
}

// TestPluginCommandSmokeFixtureIncludesLinactlLocalReplaceDeps verifies the
// isolated plugin command smoke keeps the local lina-core replacement module
// available when it copies linactl into a temporary repository.
func TestPluginCommandSmokeFixtureIncludesLinactlLocalReplaceDeps(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "reusable-plugin-command-smoke.yml"))
	if err != nil {
		t.Fatalf("read plugin command smoke workflow: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		`cp apps/lina-core/go.mod "$smoke_root/apps/lina-core/go.mod"`,
		`cp apps/lina-core/go.sum "$smoke_root/apps/lina-core/go.sum"`,
		`cp -R apps/lina-core/pkg/pluginbridge "$smoke_root/apps/lina-core/pkg/pluginbridge"`,
		`cp -R apps/lina-core/pkg/plugindb "$smoke_root/apps/lina-core/pkg/plugindb"`,
		`./apps/lina-core`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("plugin command smoke workflow missing %q", expected)
		}
	}
}

// TestMakeCommandSmokeDevFixtureIncludesLinactlLocalReplaceDeps verifies the
// isolated dev command smoke keeps linactl's local lina-core replacement module
// available even when the fixture backend is intentionally lightweight.
func TestMakeCommandSmokeDevFixtureIncludesLinactlLocalReplaceDeps(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "reusable-make-command-smoke.yml"))
	if err != nil {
		t.Fatalf("read make command smoke workflow: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		`cp apps/lina-core/go.mod "$smoke_root/apps/lina-core/go.mod"`,
		`cp apps/lina-core/go.sum "$smoke_root/apps/lina-core/go.sum"`,
		`cp -R apps/lina-core/pkg/pluginbridge "$smoke_root/apps/lina-core/pkg/pluginbridge"`,
		`cp -R apps/lina-core/pkg/plugindb "$smoke_root/apps/lina-core/pkg/plugindb"`,
		`./apps/lina-core`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("make command smoke workflow missing %q", expected)
		}
	}
	if strings.Contains(text, "module smoke-core") {
		t.Fatalf("make command smoke workflow must preserve the lina-core module path for linactl local replace")
	}
}

// TestFrontendTurboAllowsSourcePluginBuildEnv guards plugin-full frontend page discovery.
func TestFrontendTurboAllowsSourcePluginBuildEnv(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-vben", "turbo.json"))
	if err != nil {
		t.Fatalf("read frontend turbo config: %v", err)
	}

	var cfg struct {
		GlobalEnv []string `json:"globalEnv"`
	}
	if err = json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("parse frontend turbo config: %v", err)
	}

	if !containsString(cfg.GlobalEnv, plugins.SourcePluginsEnvKey) {
		t.Fatalf("frontend turbo globalEnv must include %s for plugin-full page discovery, got %#v", plugins.SourcePluginsEnvKey, cfg.GlobalEnv)
	}
}

// TestFrontendTailwindScansSourcePluginPages guards plugin-full frontend page styling.
func TestFrontendTailwindScansSourcePluginPages(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	globalCSSPath := filepath.Join(root, "apps", "lina-vben", "packages", "@core", "base", "design", "src", "css", "global.css")
	content, err := os.ReadFile(globalCSSPath)
	if err != nil {
		t.Fatalf("read frontend global CSS: %v", err)
	}

	const sourcePluginFrontendSource = "@source '../../../../../../../lina-plugins/';"
	if !strings.Contains(string(content), sourcePluginFrontendSource) {
		t.Fatalf("frontend Tailwind sources must include %s for plugin-full page styles", sourcePluginFrontendSource)
	}
}

// containsString reports whether a string slice contains the expected value.
func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
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

// TestHelperPrintStdoutAndFail prints a deterministic stdout diagnostic and
// exits with failure for command-output error tests.
func TestHelperPrintStdoutAndFail(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	fmt.Fprintln(os.Stdout, "stdout diagnostic from go list")
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

// TestHelperPrintGoListPackages prints deterministic go list -json records for
// linactl test.go planning tests.
func TestHelperPrintGoListPackages(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	fmt.Fprintln(os.Stdout, `{"ImportPath":"lina-core/internal/service/plugin","TestGoFiles":["plugin_test.go"]}`)
	fmt.Fprintln(os.Stdout, `{"ImportPath":"lina-core/internal/model"}`)
	os.Exit(0)
}

// TestHelperPrintGoListPackagesWithStderr prints go list JSON on stdout while
// emitting a diagnostic to stderr, matching Go tool output seen in CI.
func TestHelperPrintGoListPackagesWithStderr(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "--" {
		return
	}
	fmt.Fprintln(os.Stderr, "go: downloading example.com/transitive v0.0.1")
	fmt.Fprintln(os.Stdout, `{"ImportPath":"lina-core/internal/service/plugin","TestGoFiles":["plugin_test.go"]}`)
	fmt.Fprintln(os.Stdout, `{"ImportPath":"lina-core/internal/model"}`)
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

func writeDynamicPluginManifest(t *testing.T, pluginDir string, pluginID string) {
	t.Helper()
	writeFile(t, filepath.Join(pluginDir, "plugin.yaml"), fmt.Sprintf(`id: %s
name: %s
version: v0.1.0
type: dynamic
scope_nature: tenant_aware
supports_multi_tenant: false
default_install_mode: global
`, pluginID, pluginID))
}

type envCheckSQLDriver struct {
	version  string
	queryErr error
}

func (driver envCheckSQLDriver) Open(_ string) (driver.Conn, error) {
	return envCheckSQLConn{version: driver.version, queryErr: driver.queryErr}, nil
}

type envCheckSQLConn struct {
	version  string
	queryErr error
}

func (conn envCheckSQLConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not implemented by env check test driver")
}

func (conn envCheckSQLConn) Close() error {
	return nil
}

func (conn envCheckSQLConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not implemented by env check test driver")
}

func (conn envCheckSQLConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if conn.queryErr != nil {
		return nil, conn.queryErr
	}
	if query != "SHOW server_version" {
		return nil, fmt.Errorf("unexpected query %q", query)
	}
	return &envCheckSQLRows{version: conn.version}, nil
}

type envCheckSQLRows struct {
	version string
	read    bool
}

func (rows *envCheckSQLRows) Columns() []string {
	return []string{"server_version"}
}

func (rows *envCheckSQLRows) Close() error {
	return nil
}

func (rows *envCheckSQLRows) Next(dest []driver.Value) error {
	if rows.read {
		return io.EOF
	}
	rows.read = true
	dest[0] = rows.version
	return nil
}

// writeFrontendDependencySentinel creates the Vite binary expected by
// ensureFrontendDeps so runDev unit tests do not require pnpm on PATH.
func writeFrontendDependencySentinel(t *testing.T, root string) {
	t.Helper()
	writeFile(t, toolutil.ViteCommand(root), "")
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
