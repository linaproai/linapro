// This file verifies linactl command parsing, plugin discovery, asset packing,
// version metadata, source-plugin management, and cross-platform path helper
// behavior.

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
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"linactl/internal/config"
	"linactl/internal/devservice"
	"linactl/internal/fileutil"
	"linactl/internal/plugins"
	"linactl/internal/process"
	"linactl/internal/repository"
	"linactl/internal/runtimei18n"
	"linactl/internal/toolutil"
)

func init() {
	sql.Register("linactl_envcheck_test", envCheckSQLDriver{version: "14.13"})
	sql.Register("linactl_envcheck_error_test", envCheckSQLDriver{queryErr: errors.New("database unavailable")})
}

func TestParseCommandInputSupportsMakeStyleParams(t *testing.T) {
	input, err := parseCommandInput([]string{"confirm=init", "rebuild=true", "--platforms=linux/amd64,linux/arm64", "--dash-key=literal", "UPPER=ClaudeCode", "-h", "extra"})
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
	if input.Get("dash-key") != "literal" {
		t.Fatalf("hyphenated literal key mismatch: %q", input.Get("dash-key"))
	}
	if input.Get("dash_key") != "" {
		t.Fatalf("hyphenated key should not resolve as snake_case parameter")
	}
	if input.Get("upper") != "" {
		t.Fatalf("upper-case key should not resolve as lower-case parameter")
	}
	if input.Get("UPPER") != "ClaudeCode" {
		t.Fatalf("expected upper-case key to remain case-sensitive")
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

// TestCommandRegistryUsesDottedLintGoCommand guards the public Go static lint
// command name.
func TestCommandRegistryUsesDottedLintGoCommand(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["lint.go"]; !ok {
		t.Fatalf("expected command %q to be registered", "lint.go")
	}
	if _, ok := registry["lint-go"]; ok {
		t.Fatalf("legacy command %q should not be registered", "lint-go")
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

// TestCommandRegistryUsesDottedDatabaseCommands guards the public database
// command names. Top-level upgrade is reserved for framework source upgrades
// (see command_upgrade.go), not the legacy database alias of db.upgrade.
func TestCommandRegistryUsesDottedDatabaseCommands(t *testing.T) {
	registry := commandRegistry()
	for _, name := range []string{"db.init", "db.upgrade", "db.mock"} {
		if _, ok := registry[name]; !ok {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
	for _, name := range []string{"init", "mock"} {
		if _, ok := registry[name]; ok {
			t.Fatalf("legacy database command %q should not be registered", name)
		}
	}
	if _, ok := registry["upgrade"]; !ok {
		t.Fatalf("expected framework upgrade command %q to be registered", "upgrade")
	}
}

// TestRunUpgradeDispatchesCoreUpgrade verifies db.upgrade replays host SQL via
// the Lina core upgrade command instead of reusing destructive init flags.
func TestRunUpgradeDispatchesCoreUpgrade(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir core dir: %v", err)
	}
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	var calls []capturedCommand
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	err := runUpgrade(context.Background(), application, commandInput{Params: map[string]string{"confirm": "upgrade"}})
	if err != nil {
		t.Fatalf("runUpgrade returned error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected one child command, got %d: %#v", len(calls), calls)
	}
	call := calls[0]
	if call.name != "go" {
		t.Fatalf("child command name mismatch: got %q want %q", call.name, "go")
	}
	if call.cmd.Dir != filepath.Join(root, "apps", "lina-core") {
		t.Fatalf("child command dir mismatch: got %q", call.cmd.Dir)
	}
	expectedArgs := []string{"run", "main.go", "upgrade", "--confirm=upgrade", "--sql-source=local"}
	if len(call.args) != len(expectedArgs) {
		t.Fatalf("child command args length mismatch: got %#v want %#v", call.args, expectedArgs)
	}
	for i := range expectedArgs {
		if call.args[i] != expectedArgs[i] {
			t.Fatalf("child command args mismatch: got %#v want %#v", call.args, expectedArgs)
		}
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

// TestCommandRegistryUsesPluginsCheck verifies plugin governance uses the
// generic plugins.check entrypoint.
func TestCommandRegistryUsesPluginsCheck(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["plugins.check"]; !ok {
		t.Fatalf("expected command %q to be registered", "plugins.check")
	}
	if _, ok := registry["plugins.boundary.check"]; ok {
		t.Fatalf("legacy command %q should not be registered", "plugins.boundary.check")
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
	for _, command := range []string{"ctrl", "dao", "__goframe"} {
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
	for _, command := range []string{"ctrl", "dao"} {
		if !strings.Contains(output, "\n  "+command+" ") {
			t.Fatalf("full help should include internal command %q:\n%s", command, output)
		}
	}
	if strings.Contains(output, "\n  __goframe ") {
		t.Fatalf("full help should still hide hidden child command __goframe:\n%s", output)
	}
}

// TestHiddenGoFrameCommandHelpIsNotPublic verifies direct help lookup cannot
// turn the hidden bridge into a documented command surface.
func TestHiddenGoFrameCommandHelpIsNotPublic(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	err := application.run(context.Background(), []string{"help", "__goframe"})
	if err == nil || !strings.Contains(err.Error(), `unknown command "__goframe"`) {
		t.Fatalf("expected hidden command help to be rejected, got %v", err)
	}
}

// TestCommandRegistryRemovesExternalGoFrameInstaller verifies code generation
// no longer exposes commands that download an external gf binary.
func TestCommandRegistryRemovesExternalGoFrameInstaller(t *testing.T) {
	registry := commandRegistry()
	for _, command := range []string{"cli", "cli.install"} {
		if _, ok := registry[command]; ok {
			t.Fatalf("legacy external GoFrame command %q should not be registered", command)
		}
	}
}

// TestRunCtrlDispatchesEmbeddedGoFrame verifies linactl ctrl starts the hidden
// linactl child entry instead of resolving or executing gf from PATH.
func TestRunCtrlDispatchesEmbeddedGoFrame(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "hack", "config.yaml"), "gfcli: {}\n")
	application, calls := newGoFrameDispatchTestApp(t, root, filepath.Join(root, "linactl-test"))

	if err := runCtrl(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runCtrl returned error: %v", err)
	}

	requireSingleGoFrameDispatch(t, calls, filepath.Join(root, "apps", "lina-core"), filepath.Join(root, "linactl-test"), []string{"__goframe", "--config-dir=" + filepath.Join(root, "apps", "lina-core", "hack"), "gen", "ctrl"})
}

// TestRunDaoDispatchesEmbeddedGoFrame verifies linactl dao starts the hidden
// linactl child entry instead of resolving or executing gf from PATH.
func TestRunDaoDispatchesEmbeddedGoFrame(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "hack", "config.yaml"), "gfcli: {}\n")
	application, calls := newGoFrameDispatchTestApp(t, root, filepath.Join(root, "linactl-test"))

	if err := runDao(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runDao returned error: %v", err)
	}

	requireSingleGoFrameDispatch(t, calls, filepath.Join(root, "apps", "lina-core"), filepath.Join(root, "linactl-test"), []string{"__goframe", "--config-dir=" + filepath.Join(root, "apps", "lina-core", "hack"), "gen", "dao"})
}

// TestRunCtrlDispatchesToPluginBackend verifies explicit dir targeting resolves
// to the plugin backend work directory and plugin-root config directory.
func TestRunCtrlDispatchesToPluginBackend(t *testing.T) {
	var (
		root          = t.TempDir()
		pluginRoot    = filepath.Join(root, "apps", "lina-plugins", "demo-plugin")
		pluginBackend = filepath.Join(pluginRoot, "backend")
	)
	if err := os.MkdirAll(pluginBackend, 0o755); err != nil {
		t.Fatalf("mkdir plugin backend: %v", err)
	}
	writeFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: demo-plugin\n")
	writeFile(t, filepath.Join(pluginRoot, "hack", "config.yaml"), "gfcli: {}\n")
	application, calls := newGoFrameDispatchTestApp(t, root, filepath.Join(root, "linactl-test"))

	if err := runCtrl(context.Background(), application, commandInput{Params: map[string]string{"dir": "apps/lina-plugins/demo-plugin/backend"}}); err != nil {
		t.Fatalf("runCtrl returned error: %v", err)
	}

	requireSingleGoFrameDispatch(t, calls, pluginBackend, filepath.Join(root, "linactl-test"), []string{"__goframe", "--config-dir=" + filepath.Join(pluginRoot, "hack"), "gen", "ctrl"})
}

// TestRunDaoDispatchesToExplicitBackendDir verifies Makefile wrappers can pass
// a relative backend directory directly.
func TestRunDaoDispatchesToExplicitBackendDir(t *testing.T) {
	var (
		root          = t.TempDir()
		pluginRoot    = filepath.Join(root, "apps", "lina-plugins", "demo-plugin")
		pluginBackend = filepath.Join(pluginRoot, "backend")
	)
	if err := os.MkdirAll(pluginBackend, 0o755); err != nil {
		t.Fatalf("mkdir plugin backend: %v", err)
	}
	writeFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: demo-plugin\n")
	writeFile(t, filepath.Join(pluginRoot, "hack", "config.yaml"), "gfcli: {}\n")
	application, calls := newGoFrameDispatchTestApp(t, root, filepath.Join(root, "linactl-test"))

	if err := runDao(context.Background(), application, commandInput{Params: map[string]string{"dir": "apps/lina-plugins/demo-plugin/backend"}}); err != nil {
		t.Fatalf("runDao returned error: %v", err)
	}

	requireSingleGoFrameDispatch(t, calls, pluginBackend, filepath.Join(root, "linactl-test"), []string{"__goframe", "--config-dir=" + filepath.Join(pluginRoot, "hack"), "gen", "dao"})
}

// TestRunDaoRejectsTargetWithoutGoFrameConfig keeps failures actionable before
// the hidden GoFrame subprocess is started.
func TestRunDaoRejectsTargetWithoutGoFrameConfig(t *testing.T) {
	var (
		root          = t.TempDir()
		pluginRoot    = filepath.Join(root, "apps", "lina-plugins", "demo-plugin")
		pluginBackend = filepath.Join(pluginRoot, "backend")
	)
	if err := os.MkdirAll(pluginBackend, 0o755); err != nil {
		t.Fatalf("mkdir plugin backend: %v", err)
	}
	writeFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: demo-plugin\n")
	application, calls := newGoFrameDispatchTestApp(t, root, filepath.Join(root, "linactl-test"))

	err := runDao(context.Background(), application, commandInput{Params: map[string]string{"dir": "apps/lina-plugins/demo-plugin/backend"}})
	if err == nil || !strings.Contains(err.Error(), "missing config.yaml") {
		t.Fatalf("expected missing config error, got %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("hidden GoFrame child should not run for invalid target: %#v", *calls)
	}
}

// TestRunGoFrameRejectsLegacyTargetParams verifies code generation target
// selection has one explicit surface: dir=<path>.
func TestRunGoFrameRejectsLegacyTargetParams(t *testing.T) {
	application, calls := newGoFrameDispatchTestApp(t, t.TempDir(), filepath.Join(t.TempDir(), "linactl-test"))

	for _, tc := range []struct {
		name   string
		params map[string]string
	}{
		{name: "p", params: map[string]string{"p": "demo-plugin"}},
		{name: "plugin", params: map[string]string{"plugin": "demo-plugin"}},
		{name: "target", params: map[string]string{"target": "apps/lina-core"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := runDao(context.Background(), application, commandInput{Params: tc.params})
			if err == nil || !strings.Contains(err.Error(), "is not supported; use dir=<path>") {
				t.Fatalf("expected legacy parameter error, got %v", err)
			}
		})
	}
	if len(*calls) != 0 {
		t.Fatalf("hidden GoFrame child should not run for legacy target params: %#v", *calls)
	}
}

// TestRunEmbeddedGoFrameRejectsParameters verifies the hidden entry has a
// narrow positional surface and cannot be used as a generic gf proxy.
func TestRunEmbeddedGoFrameRejectsParameters(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = t.TempDir()

	err := runEmbeddedGoFrame(context.Background(), application, commandInput{
		Args:   []string{"gen", "ctrl"},
		Params: map[string]string{"path": "api"},
	})
	if err == nil || !strings.Contains(err.Error(), "only supports config-dir plus positional commands") {
		t.Fatalf("expected hidden command parameter error, got %v", err)
	}
}

// TestEmbeddedGoFrameCtrlSmokeWithoutExternalGF runs the hidden child command
// against a minimal GoFrame project with PATH cleared. It proves controller
// generation uses the embedded GoFrame CLI module instead of a gf executable.
func TestEmbeddedGoFrameCtrlSmokeWithoutExternalGF(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	if err := os.MkdirAll(filepath.Join(coreDir, "hack"), 0o755); err != nil {
		t.Fatalf("mkdir core hack dir: %v", err)
	}
	writeFile(t, filepath.Join(coreDir, "hack", "config.yaml"), "gfcli: {}\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), `module example.com/smoke

go 1.25.0

require github.com/gogf/gf/v2 v2.10.1
`)
	writeFile(t, filepath.Join(coreDir, "api", "demo", "v1", "hello.go"), `package v1

import "github.com/gogf/gf/v2/frame/g"

type HelloReq struct {
	g.Meta `+"`"+`path:"/hello" method:"get"`+"`"+`
}

type HelloRes struct{}
`)

	helper := exec.Command(os.Args[0], "-test.run=TestHelperEmbeddedGoFrameCtrl", "--", root)
	helper.Dir = coreDir
	helper.Env = []string{
		"PATH=",
		"HOME=" + t.TempDir(),
		"LINACTL_TEST_EMBEDDED_GOFRAME=1",
	}
	output, err := helper.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded GoFrame ctrl helper failed: %v\n%s", err, string(output))
	}

	controllerFile := filepath.Join(coreDir, "internal", "controller", "demo", "demo_v1_hello.go")
	content, err := os.ReadFile(controllerFile)
	if err != nil {
		t.Fatalf("read generated controller %s: %v\nhelper output:\n%s", controllerFile, err, string(output))
	}
	if !strings.Contains(string(content), "func (c *ControllerV1) Hello(") {
		t.Fatalf("generated controller does not contain Hello method:\n%s", string(content))
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

// TestRunReleaseTagCheckRequiresExplicitTag verifies environment variables are
// not accepted as implicit release inputs.
func TestRunReleaseTagCheckRequiresExplicitTag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2.3-rc.1\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "release tag is empty; pass tag=<version>") {
		t.Fatalf("expected explicit tag error, got %v", err)
	}
}

// TestRunReleaseTagCheckPrintsValidatedFrameworkVersion verifies automation output.
func TestRunReleaseTagCheckPrintsValidatedFrameworkVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: v1.2.3\n")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runReleaseTagCheck(context.Background(), application, commandInput{Params: map[string]string{"print-version": "1"}})
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

// TestDefaultEnvToolsUseNodeCLIsForFrontendChecks verifies env.check avoids
// Windows batch shims for project-local frontend CLI version probes.
func TestDefaultEnvToolsUseNodeCLIsForFrontendChecks(t *testing.T) {
	root := t.TempDir()
	tools := defaultEnvTools(root)

	vite := requireEnvTool(t, tools, "Vite")
	viteCLI := envViteCLIPath(root)
	if vite.Command != "node" {
		t.Fatalf("expected Vite probe to use node, got %q", vite.Command)
	}
	if len(vite.Args) != 2 || vite.Args[0] != viteCLI || vite.Args[1] != "--version" {
		t.Fatalf("unexpected Vite probe args: %#v", vite.Args)
	}
	if vite.RequiredPath != viteCLI {
		t.Fatalf("expected Vite required path %q, got %q", viteCLI, vite.RequiredPath)
	}
	if strings.Contains(strings.Join(vite.Args, " "), ".cmd") {
		t.Fatalf("Vite probe must not use a Windows batch shim: %#v", vite.Args)
	}

	playwright := requireEnvTool(t, tools, "Playwright")
	playwrightCLI := envPlaywrightCLIPath(root)
	if playwright.Command != "node" {
		t.Fatalf("expected Playwright probe to use node, got %q", playwright.Command)
	}
	if len(playwright.Args) != 2 || playwright.Args[0] != playwrightCLI || playwright.Args[1] != "--version" {
		t.Fatalf("unexpected Playwright probe args: %#v", playwright.Args)
	}
	if playwright.RequiredPath != playwrightCLI {
		t.Fatalf("expected Playwright required path %q, got %q", playwrightCLI, playwright.RequiredPath)
	}
	if strings.Contains(strings.Join(playwright.Args, " "), "pnpm exec") {
		t.Fatalf("Playwright probe must not use pnpm exec: %#v", playwright.Args)
	}
}

// TestProbeEnvToolReportsMissingRequiredPathWithoutRunningCommand verifies a
// missing local CLI is reported as a dependency issue before spawning a child.
func TestProbeEnvToolReportsMissingRequiredPathWithoutRunningCommand(t *testing.T) {
	root := t.TempDir()
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("probe must not run %s %v when RequiredPath is missing", name, args)
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	result := probeEnvTool(context.Background(), application, envTool{
		Name:         "Playwright",
		Command:      "node",
		Args:         []string{envPlaywrightCLIPath(root), "--version"},
		RequiredPath: envPlaywrightCLIPath(root),
	})
	if !result.Missing {
		t.Fatalf("expected missing result for absent RequiredPath")
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
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
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
		switch {
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			return exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			return exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperRecordWorkingDirectory", "--")
	}

	if err := runEnvSetup(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runEnvSetup returned error: %v", err)
	}

	got := strings.Join(commands, "\n")
	expected := strings.Join([]string{
		"golangci-lint --version",
		"staticcheck -version",
		"pnpm install",
		"pnpm exec playwright install --with-deps --only-shell chromium",
	}, "\n")
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

// TestRunEnvSetupInstallsGoLintToolsFirst verifies env.setup prepares the
// repository-pinned Go lint tools before frontend and browser setup.
func TestRunEnvSetupInstallsGoLintToolsFirst(t *testing.T) {
	root := t.TempDir()
	gopath := filepath.Join(root, "gopath")
	golangciLintBinary := filepath.Join(gopath, "bin", toolutil.ExecutableName("golangci-lint"))
	staticcheckBinary := filepath.Join(gopath, "bin", toolutil.ExecutableName("staticcheck"))
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "hack", "tests"), 0o755); err != nil {
		t.Fatalf("mkdir test workspace: %v", err)
	}

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = []string{
		"GOWORK=" + filepath.Join(root, "temp", "go.work.plugins"),
		"GOFLAGS=-tags=official_plugins",
		"GOOS=wasip1",
		"GOARCH=wasm",
		"GOWASM=satconv",
	}
	application.lookPath = func(name string) (string, error) {
		switch name {
		case "golangci-lint", "staticcheck":
			return "", fmt.Errorf("%s missing", name)
		default:
			return name, nil
		}
	}
	var calls []capturedCommand
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "env GOBIN":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "")
		case name == "go" && strings.Join(args, " ") == "env GOPATH":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", gopath)
		case name == golangciLintBinary && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == staticcheckBinary && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	if err := runEnvSetup(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runEnvSetup returned error: %v", err)
	}

	expected := []string{
		"go env GOBIN",
		"go env GOPATH",
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2",
		golangciLintBinary + " --version",
		"go env GOBIN",
		"go env GOPATH",
		"go install honnef.co/go/tools/cmd/staticcheck@v0.7.0",
		staticcheckBinary + " -version",
		"pnpm install",
		"pnpm exec playwright install --with-deps --only-shell chromium",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected env.setup command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	for _, index := range []int{2, 6} {
		installEnv := calls[index].cmd.Env
		if got := toolutil.EnvValue(installEnv, "GOWORK"); got != "off" {
			t.Fatalf("expected install to force GOWORK=off, got %q", got)
		}
		for _, key := range []string{"GOFLAGS", "GOOS", "GOARCH", "GOWASM"} {
			if got := toolutil.EnvValue(installEnv, key); got != "" {
				t.Fatalf("expected install env to remove %s, got %q", key, got)
			}
		}
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

	plugins, err := dynamicPlugins(root)
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

// TestRunBuildRunsPluginConfigCommandsBeforeBackendCompile verifies
// plugin-owned build commands run before Go embed compilation.
func TestRunBuildRunsPluginConfigCommandsBeforeBackendCompile(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)

	calls := runBuildWithCapturedCommands(t, root, nil, commandInput{Params: map[string]string{"plugins": "1"}})

	pluginBuildIndex := -1
	backendBuildIndex := -1
	for index, call := range calls {
		if call.name == "pnpm" && len(call.args) >= 4 && call.args[0] == "--dir" && call.args[2] == "run" && call.args[3] == "build" && call.cmd.Dir == filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox") {
			pluginBuildIndex = index
		}
		if call.name == "go" && len(call.args) >= 1 && call.args[0] == "build" && call.cmd.Dir == filepath.Join(root, "apps", "lina-core") {
			backendBuildIndex = index
		}
	}
	if pluginBuildIndex < 0 {
		t.Fatalf("expected plugin build command call, got %#v", calls)
	}
	if backendBuildIndex < 0 {
		t.Fatalf("expected backend go build call, got %#v", calls)
	}
	if pluginBuildIndex > backendBuildIndex {
		t.Fatalf("plugin build command must run before backend build, calls=%#v", calls)
	}
}

func TestRunBuildDirBuildsSelectedPluginOnly(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "another-plugin", "plugin.yaml"), "id: another-plugin\ntype: source\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "another-plugin", "hack", "config.yaml"), "build:\n  commands:\n    - node build.mjs\n")
	staleWorkspace := filepath.Join(root, "stale.work")

	calls := runBuildWithCapturedCommands(t, root, []string{
		"GOWORK=" + staleWorkspace,
		"GOFLAGS=-mod=mod -tags=netgo",
	}, commandInput{Params: map[string]string{
		"dir":     "apps/lina-plugins/john-ai-agentbox",
		"plugins": "1",
	}})

	if len(calls) != 1 {
		t.Fatalf("expected one plugin build call, got %#v", calls)
	}
	call := calls[0]
	if call.name != "pnpm" || len(call.args) < 4 || call.args[0] != "--dir" || call.args[1] != filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox", "frontend") || call.args[2] != "run" || call.args[3] != "build" {
		t.Fatalf("unexpected selected plugin build command: %#v", call)
	}
	if call.cmd.Dir != filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox") {
		t.Fatalf("selected plugin build dir mismatch: %q", call.cmd.Dir)
	}
	if got := toolutil.EnvValue(call.cmd.Env, "GOWORK"); got != filepath.Join(root, "temp", "go.work.plugins") {
		t.Fatalf("expected selected plugin build to use prepared plugin workspace, got %q", got)
	}
	if got := toolutil.EnvValue(call.cmd.Env, plugins.SourcePluginsEnvKey); got != "1" {
		t.Fatalf("expected selected plugin build to enable source plugin env, got %q", got)
	}
	if got := toolutil.EnvValue(call.cmd.Env, "GOFLAGS"); !strings.Contains(got, plugins.OfficialBuildTag) {
		t.Fatalf("expected selected plugin build to set official plugin build tag, got %q", got)
	}
}

func TestRunBuildDirRunsArbitraryHackConfigCommands(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	targetDir := filepath.Join(root, "tools", "custom-builder")
	writeFile(t, filepath.Join(targetDir, "hack", "config.yaml"), "build:\n  commands:\n    - node build.mjs --target \"$(BUILD_DIR)\" --repo \"$(REPO_ROOT)\"\n")

	calls := runBuildWithCapturedCommands(t, root, nil, commandInput{Params: map[string]string{
		"dir": "tools/custom-builder",
	}})

	if len(calls) != 1 {
		t.Fatalf("expected one configured build command, got %#v", calls)
	}
	call := calls[0]
	if call.name != "node" || len(call.args) != 5 || call.args[0] != "build.mjs" || call.args[1] != "--target" || call.args[2] != targetDir || call.args[3] != "--repo" || call.args[4] != root {
		t.Fatalf("unexpected configured build command: %#v", call)
	}
	if call.cmd.Dir != targetDir {
		t.Fatalf("configured build dir mismatch: got %q want %q", call.cmd.Dir, targetDir)
	}
	if got := toolutil.EnvValue(call.cmd.Env, plugins.SourcePluginsEnvKey); got != "0" {
		t.Fatalf("expected arbitrary configured build to use host-only plugin env, got %q", got)
	}
}

func TestRunBuildDirRejectsPackageBuildWithoutHackConfig(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	writeFile(t, filepath.Join(root, "tools", "package-builder", "package.json"), `{"scripts":{"build":"vite build"}}`)

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("build without hack/config.yaml must not execute %s %#v", name, args)
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
	}

	err := runBuild(context.Background(), application, commandInput{Params: map[string]string{"dir": "tools/package-builder"}})
	if err == nil || !strings.Contains(err.Error(), "build dir has no hack/config.yaml") {
		t.Fatalf("expected missing hack/config.yaml error, got %v", err)
	}
}

func TestRunBuildWithoutDirBuildsAllPlugins(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "another-plugin", "plugin.yaml"), "id: another-plugin\ntype: source\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "another-plugin", "hack", "config.yaml"), "build:\n  commands:\n    - node build.mjs\n")

	calls := runBuildWithCapturedCommands(t, root, nil, commandInput{Params: map[string]string{"plugins": "1"}})

	built := map[string]bool{}
	for _, call := range calls {
		if call.cmd.Dir == filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox") {
			built["john-ai-agentbox"] = true
		}
		if call.cmd.Dir == filepath.Join(root, "apps", "lina-plugins", "another-plugin") {
			built["another-plugin"] = true
		}
	}
	for _, pluginID := range []string{"john-ai-agentbox", "another-plugin"} {
		if !built[pluginID] {
			t.Fatalf("expected root build to run build commands for %s, calls=%#v", pluginID, calls)
		}
	}
}

func TestRunBuildDirBuildsHostFrontendOnly(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)

	calls := runBuildWithCapturedCommands(t, root, nil, commandInput{Params: map[string]string{
		"dir": "apps/lina-vben",
	}})

	if len(calls) != 1 {
		t.Fatalf("expected one frontend build call, got %#v", calls)
	}
	call := calls[0]
	if call.name != "pnpm" || len(call.args) < 2 || call.args[0] != "run" || call.args[1] != "build" {
		t.Fatalf("unexpected frontend build command: %#v", call)
	}
	if call.cmd.Dir != filepath.Join(root, "apps", "lina-vben") {
		t.Fatalf("frontend build dir mismatch: %q", call.cmd.Dir)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-core", "internal", "packed", "public", "index.html")) {
		t.Fatalf("host frontend build did not refresh packed public assets")
	}
}

func TestRunBuildDirBuildsHostBackendWithPreparedAssets(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)

	calls := runBuildWithCapturedCommands(t, root, nil, commandInput{Params: map[string]string{
		"dir": "apps/lina-core",
	}})

	if len(calls) != 2 {
		t.Fatalf("expected frontend and backend build calls, got %#v", calls)
	}
	if calls[0].name != "pnpm" || calls[0].cmd.Dir != filepath.Join(root, "apps", "lina-vben") {
		t.Fatalf("expected frontend build first, got %#v", calls[0])
	}
	if calls[1].name != "go" || len(calls[1].args) < 1 || calls[1].args[0] != "build" || calls[1].cmd.Dir != filepath.Join(root, "apps", "lina-core") {
		t.Fatalf("expected backend go build second, got %#v", calls[1])
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-core", "internal", "packed", "manifest", "config", "config.template.yaml")) {
		t.Fatalf("host backend build did not prepare packed manifest assets")
	}
}

func TestRunDevDirUsesBuildDirLogic(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	targetDir := filepath.Join(root, "tools", "dev-target")
	writeFile(t, filepath.Join(targetDir, "hack", "config.yaml"), "build:\n  commands:\n    - node dev-build.mjs --target \"$(BUILD_DIR)\"\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	application.portInUse = func(int) bool {
		t.Fatalf("runDev dir should not enter service port checks")
		return false
	}

	if err := runDev(context.Background(), application, commandInput{Params: map[string]string{
		"dir": "tools/dev-target",
	}}); err != nil {
		t.Fatalf("runDev returned error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected one configured build command, got %#v", calls)
	}
	call := calls[0]
	if call.name != "node" || len(call.args) != 3 || call.args[0] != "dev-build.mjs" || call.args[1] != "--target" || call.args[2] != targetDir {
		t.Fatalf("unexpected dev dir build command: %#v", call)
	}
	if call.cmd.Dir != targetDir {
		t.Fatalf("dev dir build command dir mismatch: got %q want %q", call.cmd.Dir, targetDir)
	}
}

func TestRunStopDirRunsConfiguredCommands(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	targetDir := filepath.Join(root, "tools", "service-target")
	writeFile(t, filepath.Join(targetDir, "hack", "config.yaml"), "stop:\n  commands:\n    - node stop.mjs --target \"$(TARGET_DIR)\" --repo \"$(REPO_ROOT)\"\n")

	var calls []capturedCommand
	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	if err := runStop(context.Background(), application, commandInput{Params: map[string]string{
		"dir": "tools/service-target",
	}}); err != nil {
		t.Fatalf("runStop returned error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected one configured stop command, got %#v", calls)
	}
	call := calls[0]
	if call.name != "node" || len(call.args) != 5 || call.args[0] != "stop.mjs" || call.args[1] != "--target" || call.args[2] != targetDir || call.args[3] != "--repo" || call.args[4] != root {
		t.Fatalf("unexpected stop command: %#v", call)
	}
	if call.cmd.Dir != targetDir {
		t.Fatalf("stop command dir mismatch: got %q want %q", call.cmd.Dir, targetDir)
	}
	if call.cmd.Stdout != &stdout {
		t.Fatalf("configured stop command output should be forwarded")
	}
}

func TestRunStatusDirRunsConfiguredCommands(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	targetDir := filepath.Join(root, "tools", "service-target")
	writeFile(t, filepath.Join(targetDir, "hack", "config.yaml"), "status:\n  commands:\n    - node status.mjs --target \"$(BUILD_DIR)\"\n")

	var calls []capturedCommand
	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	application.portInUse = func(int) bool {
		t.Fatalf("runStatus dir should not enter default service status checks")
		return false
	}

	if err := runStatus(context.Background(), application, commandInput{Params: map[string]string{
		"dir": "tools/service-target",
	}}); err != nil {
		t.Fatalf("runStatus returned error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected one configured status command, got %#v", calls)
	}
	call := calls[0]
	if call.name != "node" || len(call.args) != 3 || call.args[0] != "status.mjs" || call.args[1] != "--target" || call.args[2] != targetDir {
		t.Fatalf("unexpected status command: %#v", call)
	}
	if call.cmd.Dir != targetDir {
		t.Fatalf("status command dir mismatch: got %q want %q", call.cmd.Dir, targetDir)
	}
	if call.cmd.Stdout != &stdout {
		t.Fatalf("configured status command output should be forwarded")
	}
}

func TestRunStopDirRequiresHackConfig(t *testing.T) {
	root := t.TempDir()
	writeBuildFixture(t, root)
	targetDir := filepath.Join(root, "tools", "missing-config")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	err := runStop(context.Background(), application, commandInput{Params: map[string]string{
		"dir": "tools/missing-config",
	}})
	if err == nil {
		t.Fatalf("expected runStop to reject dir without hack/config.yaml")
	}
	if !strings.Contains(err.Error(), "stop dir has no hack/config.yaml") {
		t.Fatalf("unexpected missing config error: %v", err)
	}
}

func TestResolveBuildConfigStepsSkipsTargetsWithoutCommands(t *testing.T) {
	var (
		root         = t.TempDir()
		withBuild    = filepath.Join(root, "apps", "lina-plugins", "with-build")
		withoutBuild = filepath.Join(root, "apps", "lina-plugins", "without-build")
	)
	writeFile(t, filepath.Join(withBuild, "plugin.yaml"), "id: with-build\n")
	writeFile(t, filepath.Join(withBuild, "hack", "config.yaml"), "build:\n  commands:\n    - node build.mjs\n    - pnpm --dir \"$(BUILD_DIR)/frontend\" run build\n")
	writeFile(t, filepath.Join(withoutBuild, "plugin.yaml"), "id: without-build\n")
	writeFile(t, filepath.Join(withoutBuild, "hack", "config.yaml"), "gfcli:\n  gen:\n    dao: []\n")

	plugins, err := discoverPluginBuildRoots(root)
	if err != nil {
		t.Fatalf("discoverPluginBuildRoots returned error: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("unexpected plugin roots: %#v", plugins)
	}

	steps, exists, err := resolveBuildConfigSteps(root, withBuild)
	if err != nil {
		t.Fatalf("resolveBuildConfigSteps returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected build config to exist")
	}
	if len(steps) != 2 {
		t.Fatalf("unexpected plugin build steps: %#v", steps)
	}
	if steps[1].Command != "pnpm" || steps[1].Args[1] != filepath.Join(withBuild, "frontend") {
		t.Fatalf("expected build dir expansion in command, got %#v", steps[1])
	}
	emptySteps, exists, err := resolveBuildConfigSteps(root, withoutBuild)
	if err != nil {
		t.Fatalf("resolveBuildConfigSteps without commands returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected empty build config to exist")
	}
	if len(emptySteps) != 0 {
		t.Fatalf("expected no build steps, got %#v", emptySteps)
	}
	missingSteps, exists, err := resolveBuildConfigSteps(root, filepath.Join(root, "missing-config"))
	if err != nil {
		t.Fatalf("resolveBuildConfigSteps missing config returned error: %v", err)
	}
	if exists {
		t.Fatalf("missing build config should not be reported as existing")
	}
	if len(missingSteps) != 0 {
		t.Fatalf("expected no steps for missing config, got %#v", missingSteps)
	}

	serviceTarget := filepath.Join(root, "tools", "service-target")
	writeFile(t, filepath.Join(serviceTarget, "hack", "config.yaml"), "stop:\n  commands:\n    - node stop.mjs\nstatus:\n  commands:\n    - node status.mjs\n")
	stopSteps, exists, err := resolveCommandConfigSteps(root, serviceTarget, "stop")
	if err != nil {
		t.Fatalf("resolveCommandConfigSteps stop returned error: %v", err)
	}
	if !exists || len(stopSteps) != 1 || stopSteps[0].Command != "node" || stopSteps[0].Args[0] != "stop.mjs" {
		t.Fatalf("unexpected stop steps: exists=%t steps=%#v", exists, stopSteps)
	}
	statusSteps, exists, err := resolveCommandConfigSteps(root, serviceTarget, "status")
	if err != nil {
		t.Fatalf("resolveCommandConfigSteps status returned error: %v", err)
	}
	if !exists || len(statusSteps) != 1 || statusSteps[0].Command != "node" || statusSteps[0].Args[0] != "status.mjs" {
		t.Fatalf("unexpected status steps: exists=%t steps=%#v", exists, statusSteps)
	}
}

func TestResolveBuildConfigStepsLeavesUnknownVariablesLiteral(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "apps", "lina-plugins", "with-build")
	writeFile(t, filepath.Join(targetDir, "hack", "config.yaml"), "build:\n  commands:\n    - pnpm --dir \"$(UNKNOWN_ROOT)/frontend\" run build\n")

	steps, exists, err := resolveBuildConfigSteps(root, targetDir)
	if err != nil {
		t.Fatalf("resolveBuildConfigSteps returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected config existence to be reported")
	}
	if len(steps) != 1 || steps[0].Args[1] != "$(UNKNOWN_ROOT)/frontend" {
		t.Fatalf("expected unknown variable to stay literal, got %#v", steps)
	}
}

func TestRunWasmResolvesExplicitRelativeOutputFromRepositoryRoot(t *testing.T) {
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
			"out": "temp/output",
			"dir": filepath.Join("apps", "lina-plugins", "linapro-demo-dynamic"),
		},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	expected := filepath.Join(root, "temp", "output")
	artifactPath := filepath.Join(expected, "linapro-demo-dynamic.wasm")
	if !fileutil.FileExists(artifactPath) {
		t.Fatalf("expected wasm artifact at %s", artifactPath)
	}
	workspaceArtifactPath := filepath.Join(workDir, "temp", "output", "linapro-demo-dynamic.wasm")
	if fileutil.FileExists(workspaceArtifactPath) {
		t.Fatalf("wasm artifact should not be written under plugin workspace: %s", workspaceArtifactPath)
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
		Params: map[string]string{"dir": filepath.Join("apps", "lina-plugins", "linapro-demo-dynamic")},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	expected := filepath.Join(root, "temp", "output")
	artifactPath := filepath.Join(expected, "linapro-demo-dynamic.wasm")
	if !fileutil.FileExists(artifactPath) {
		t.Fatalf("expected wasm artifact at %s", artifactPath)
	}
}

func TestRunWasmDoesNotUsePluginIDParameterForSelection(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeDynamicPluginManifest(t, filepath.Join(pluginRoot, "dynamic-a"), "dynamic-a")
	writeDynamicPluginManifest(t, filepath.Join(pluginRoot, "dynamic-b"), "dynamic-b")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	if err := runWasm(context.Background(), application, commandInput{
		Params: map[string]string{
			"p":       "dynamic-a",
			"dry-run": "true",
		},
	}); err != nil {
		t.Fatalf("runWasm returned error: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Building dynamic wasm plugin: dynamic-a",
		"Building dynamic wasm plugin: dynamic-b",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected dry-run output to contain %q, got:\n%s", expected, output)
		}
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
			Entry:   "Lina Core",
			Status:  "running",
			URL:     "http://127.0.0.1:9120/",
			PID:     "12345",
			PIDFile: "temp/pids/lina-core.pid",
			LogFile: "temp/lina-core.log",
		},
		{
			Entry:   "Lina Vben",
			Status:  "stopped",
			URL:     "http://127.0.0.1:5666/",
			PID:     "-",
			PIDFile: "temp/pids/lina-vben.pid",
			LogFile: "temp/lina-vben.log",
		},
	})
	if err != nil {
		t.Fatalf("devservice.PrintStatusTable returned error: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"+",
		"| Entry",
		"| Lina Core",
		"| Lina Vben",
		"| running",
		"| stopped",
		"temp/pids/lina-core.pid",
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
	// Use the current test process PID so process.Alive reports the recorded
	// process as live; we are validating redirect handling, not liveness.
	// 使用当前测试进程 PID，让 process.Alive 视为存活，专注校验重定向处理。
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}
	if err := devservice.WaitHTTP("Backend", server.URL+"/", pidFile, "service.log", time.Second, nil); err != nil {
		t.Fatalf("devservice.WaitHTTP should accept redirect readiness responses: %v", err)
	}
}

// TestWaitHTTPFailsFastWhenProcessExits验证 waitHTTP 在 PID 进程不存活时立即返回错误，
// 避免子进程已 fatal 退出却被外部端口占用方"假装"为就绪。
func TestWaitHTTPFailsFastWhenProcessExits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// External occupant returns 404 — would have looked "ready" before.
		// 模拟外部占用方返回 404，旧逻辑会误判为就绪。
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	pidFile := filepath.Join(t.TempDir(), "service.pid")
	// Spawn a tiny process and Wait for it so we get a PID that is
	// guaranteed to no longer be alive when waitHTTP inspects it.
	// 启动并等待一个一次性子进程，得到一个保证已退出的真实 PID。
	helper := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
	if err := helper.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	deadPID := helper.Process.Pid
	if err := helper.Wait(); err != nil {
		t.Fatalf("helper Wait: %v", err)
	}
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(deadPID)), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}

	var (
		start   = time.Now()
		err     = devservice.WaitHTTP("Backend", server.URL+"/api.json", pidFile, "service.log", 10*time.Second, process.Alive)
		elapsed = time.Since(start)
	)
	if err == nil {
		t.Fatalf("devservice.WaitHTTP should fail when recorded process is not alive")
	}
	if !strings.Contains(err.Error(), "process") {
		t.Fatalf("expected process-exited error, got: %v", err)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("devservice.WaitHTTP should fail fast on dead process, took %s", elapsed)
	}
}

// TestRunDevRejectsOccupiedPort验证当端口已被外部进程占用时，runDev 立即报错而非
// 让后端在静默 fatal 后被假就绪覆盖。
func TestRunDevRejectsOccupiedPort(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	// portcheck.Verify 在 runDev 入口校验后端 server.address 与前端 vite proxy
	// target 是否与 defaultBackendPort 对齐，用例需要自带最小一致夹具，否则
	// 测试会在端口校验前的 portcheck 阶段失败，无法覆盖 EnsurePortsAvailable
	// 的端口占用拒绝逻辑。
	// portcheck.Verify runs at the start of runDev and requires both the
	// backend server.address and the frontend vite proxy target to match
	// the supplied backend port; without these the test would fail in
	// portcheck before reaching the EnsurePortsAvailable check it intends
	// to validate.
	backendAddress := fmt.Sprintf("server:\n  address: \":%d\"\n", defaultBackendPort)
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), backendAddress)
	viteConfig := fmt.Sprintf("proxy: { '/api': { target: 'http://localhost:%d' } }\n", defaultBackendPort)
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "vite.config.mts"), viteConfig)
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	// Backend port pretends to be occupied by an external process.
	// 模拟后端端口被外部进程占用。
	application.portInUse = func(port int) bool {
		return port == defaultBackendPort
	}
	application.processList = func() ([]process.Info, error) {
		return nil, nil
	}

	err := runDev(context.Background(), application, commandInput{Params: map[string]string{"skip_wasm": "true"}})
	if err == nil {
		t.Fatalf("runDev should fail when backend port is occupied")
	}
	if !strings.Contains(err.Error(), "backend port") || !strings.Contains(err.Error(), "already in use") {
		t.Fatalf("expected port-in-use error, got: %v", err)
	}
}

// TestRunDevWaitsForManagedServicePortsToRelease verifies that runDev allows
// a short release window after stopping PID-file-backed services before it
// rejects occupied development ports.
func TestRunDevWaitsForManagedServicePortsToRelease(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), "server:\n  address: \":9120\"\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "vite.config.mts"), "proxy: { '/api': { target: 'http://localhost:9120' } }\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	services := devservice.Services(root, defaultBackendPort, defaultFrontendPort)
	for _, service := range services {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
		if err := cmd.Start(); err != nil {
			t.Fatalf("start existing %s helper: %v", service.Name, err)
		}
		t.Cleanup(func() {
			if killErr := cmd.Process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
				t.Logf("kill existing helper %d: %v", cmd.Process.Pid, killErr)
			}
			if waitErr := cmd.Wait(); waitErr != nil {
				t.Logf("wait existing helper %d: %v", cmd.Process.Pid, waitErr)
			}
		})
		writeFile(t, service.PIDPath, strconv.Itoa(cmd.Process.Pid))
	}

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

	probes := 0
	application.portInUse = func(int) bool {
		probes++
		return probes <= 4
	}

	if err := runDev(context.Background(), application, commandInput{Params: map[string]string{"skip_wasm": "true"}}); err != nil {
		t.Fatalf("runDev should wait for managed service ports to release, got: %v", err)
	}
	if probes < 5 {
		t.Fatalf("expected runDev to probe ports after stopping managed services, got %d probes", probes)
	}
	for _, service := range services {
		pid := devservice.ReadPID(service.PIDPath)
		if pid == 0 {
			t.Fatalf("expected restarted %s pid file to be written", service.Name)
		}
		process, err := os.FindProcess(pid)
		if err == nil {
			if killErr := process.Kill(); killErr != nil {
				t.Logf("kill restarted %s process %d: %v", service.Name, pid, killErr)
			}
		}
		if err = os.Remove(service.PIDPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove restarted %s pid file: %v", service.Name, err)
		}
	}
}

// TestRunDevOnlyWaitsForStoppedManagedPorts verifies that dev does not wait
// on unrelated occupied ports before surfacing the external occupant error.
func TestRunDevOnlyWaitsForStoppedManagedPorts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.yaml"), "server:\n  address: \":9120\"\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "vite.config.mts"), "proxy: { '/api': { target: 'http://localhost:9120' } }\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "metadata: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "select 1;\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"), "{}\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"), 0o755); err != nil {
		t.Fatalf("mkdir frontend workdir: %v", err)
	}
	writeFrontendDependencySentinel(t, root)

	services := devservice.Services(root, defaultBackendPort, defaultFrontendPort)
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start existing backend helper: %v", err)
	}
	t.Cleanup(func() {
		if killErr := cmd.Process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			t.Logf("kill existing backend helper %d: %v", cmd.Process.Pid, killErr)
		}
		if waitErr := cmd.Wait(); waitErr != nil {
			t.Logf("wait existing backend helper %d: %v", cmd.Process.Pid, waitErr)
		}
	})
	writeFile(t, services[0].PIDPath, strconv.Itoa(cmd.Process.Pid))

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}

	backendProbes := 0
	frontendProbes := 0
	application.portInUse = func(port int) bool {
		switch port {
		case defaultBackendPort:
			backendProbes++
			return false
		case defaultFrontendPort:
			frontendProbes++
			return true
		default:
			return false
		}
	}
	application.processList = func() ([]process.Info, error) {
		return nil, nil
	}

	err := runDev(context.Background(), application, commandInput{Params: map[string]string{"skip_wasm": "true"}})
	if err == nil {
		t.Fatalf("runDev should reject externally occupied frontend port")
	}
	if !strings.Contains(err.Error(), "frontend port") || !strings.Contains(err.Error(), "already in use") {
		t.Fatalf("expected frontend port-in-use error, got: %v", err)
	}
	if frontendProbes != 2 {
		t.Fatalf("expected frontend port to be checked by project-occupant cleanup and final validation, got %d probes", frontendProbes)
	}
	if backendProbes < 2 {
		t.Fatalf("expected backend port to be checked by release wait and final validation, got %d probes", backendProbes)
	}
}

func TestRunDevStopsCurrentProjectFrontendPortOccupant(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
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
	application.execCommand = func(_ context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	application.waitHTTP = func(_ string, _ string, pidPath string, _ string, _ time.Duration) error {
		if devservice.ReadPID(pidPath) == 0 {
			return os.ErrNotExist
		}
		return nil
	}

	frontendOccupied := true
	application.portInUse = func(port int) bool {
		return port == defaultFrontendPort && frontendOccupied
	}
	application.processList = func() ([]process.Info, error) {
		return []process.Info{
			{
				PID: 43210,
				Args: []string{
					"node",
					filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "..", "vite", "bin", "vite.js"),
					"--mode",
					"development",
				},
				CWD: filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"),
			},
		}, nil
	}

	killedPID := 0
	application.processKill = func(pid int) error {
		killedPID = pid
		frontendOccupied = false
		return nil
	}

	if err := runDev(context.Background(), application, commandInput{Params: map[string]string{"skip_wasm": "true"}}); err != nil {
		t.Fatalf("runDev should stop current-project frontend occupant, got: %v", err)
	}
	if killedPID != 43210 {
		t.Fatalf("expected current-project frontend process to be stopped, got pid %d", killedPID)
	}
	if !strings.Contains(stdout.String(), "Frontend port 5666 is occupied by current project process 43210; stopped it") {
		t.Fatalf("expected cleanup message, got: %s", stdout.String())
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
	application.execCommand = func(runCtx context.Context, name string, args ...string) *exec.Cmd {
		if name == "go" && len(args) >= 1 && args[0] == "build" {
			return exec.Command("true")
		}
		// Regression for https://github.com/linaproai/linapro/issues/85:
		// runDev must start Vite through node instead of Windows batch shims.
		isFrontend := name == "node" && len(args) > 0 && args[0] == toolutil.ViteCommand(root)
		if isFrontend {
			if strings.Contains(strings.Join(append([]string{name}, args...), " "), ".cmd") {
				t.Fatalf("frontend process must not use a Windows batch shim: %s %#v", name, args)
			}
			serviceEnv, _ := runCtx.Value(devservice.RunnerContextServiceEnvKey).([]string)
			if got := toolutil.EnvValue(serviceEnv, "LINAPRO_FRONTEND_DEV_SERVER_URL"); got != "" {
				t.Fatalf("frontend process must not receive backend proxy env, got %q", got)
			}
		} else if serviceEnv, _ := runCtx.Value(devservice.RunnerContextServiceEnvKey).([]string); toolutil.EnvValue(serviceEnv, "LINAPRO_FRONTEND_DEV_SERVER_URL") != "http://127.0.0.1:5666" {
			got := toolutil.EnvValue(serviceEnv, "LINAPRO_FRONTEND_DEV_SERVER_URL")
			t.Fatalf("backend process must receive frontend dev server URL, got %q", got)
		}
		return exec.Command(os.Args[0], "-test.run=TestHelperLongRunningProcess", "--")
	}
	application.waitHTTP = func(_ string, _ string, pidPath string, _ string, _ time.Duration) error {
		if devservice.ReadPID(pidPath) == 0 {
			return os.ErrNotExist
		}
		return nil
	}
	// Treat development ports as free regardless of the host machine state so
	// the test does not flap when port 8080/5666 are bound by other processes.
	// 测试中将开发端口视为空闲，避免本机其他进程占用导致用例不稳定。
	application.portInUse = func(int) bool { return false }

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
		filepath.Join(root, "temp", "pids", "lina-core.pid"),
		filepath.Join(root, "temp", "pids", "lina-vben.pid"),
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
	for _, expected := range []string{
		"Lina Core is ready: http://127.0.0.1:9120/",
		"Lina Vben is ready: http://127.0.0.1:5666/",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected readiness output to contain %q, got:\n%s", expected, output)
		}
	}
	statusTitleIndex := strings.LastIndex(output, "LinaPro Framework Status")
	if statusTitleIndex < 0 {
		t.Fatalf("expected final status title in output:\n%s", output)
	}
	finalOutput := output[statusTitleIndex:]
	for _, expected := range []string{
		"| Entry",
		"| Lina Core",
		"| Lina Vben",
		"temp/pids/lina-core.pid",
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
	// Treat development ports as free regardless of the host machine state so
	// the test does not flap when port 8080/5666 are bound by other processes.
	// 测试中将开发端口视为空闲，避免本机其他进程占用导致用例不稳定。
	application.portInUse = func(int) bool { return false }

	if err := runDev(context.Background(), application, commandInput{Params: map[string]string{"plugins": "1"}}); err != nil {
		t.Fatalf("runDev returned error: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "temp", "pids", "lina-core.pid"),
		filepath.Join(root, "temp", "pids", "lina-vben.pid"),
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
	if !strings.Contains(stdout.String(), "Source plugin pages are mounted inside Lina Vben.") {
		t.Fatalf("expected plugin entry hint, got:\n%s", stdout.String())
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
	if err == nil || !strings.Contains(err.Error(), "invalid boolean value") {
		t.Fatalf("expected non-boolean plugin mode value to be rejected, got auto=%v err=%v", auto, err)
	}
}

func TestOfficialPluginGoWorkUsesDiscoversPluginModules(t *testing.T) {
	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	writeFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\ntype: source\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\ntype: source\n")

	workspace, err := plugins.InspectOfficialWorkspace(root)
	if err != nil {
		t.Fatalf("plugins.InspectOfficialWorkspace returned error: %v", err)
	}
	uses, err := plugins.GoWorkUses(root, workspace)
	if err != nil {
		t.Fatalf("plugins.GoWorkUses returned error: %v", err)
	}
	got := strings.Join(uses, ",")
	expected := "./apps/lina-plugins/plugin-a,./apps/lina-plugins/plugin-b"
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
	var (
		root         = t.TempDir()
		coreDir      = filepath.Join(root, "apps", "lina-core")
		aggregateDir = plugins.AggregateModuleDir(root)
	)
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

// TestGoLintWorkspaceModulesKeepsGeneratedOfficialPluginAggregate verifies
// lint.go leaves the generated plugin bridge visible so .golangci.yml remains
// the single source of generated-code exclusions.
func TestGoLintWorkspaceModulesKeepsGeneratedOfficialPluginAggregate(t *testing.T) {
	var (
		root         = t.TempDir()
		coreDir      = filepath.Join(root, "apps", "lina-core")
		aggregateDir = plugins.AggregateModuleDir(root)
	)
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

	modules, err := goLintWorkspaceModules(context.Background(), application)
	if err != nil {
		t.Fatalf("goLintWorkspaceModules returned error: %v", err)
	}
	if len(modules) != 2 || !samePath(t, modules[0], coreDir) || !samePath(t, modules[1], aggregateDir) {
		t.Fatalf("unexpected lint workspace modules: %#v", modules)
	}
}

// TestRunLintGoDispatchesHostWorkspaceModules verifies host-only lint runs
// golangci-lint once per discovered workspace module with the repository config.
func TestRunLintGoDispatchesHostWorkspaceModules(t *testing.T) {
	var (
		root    = t.TempDir()
		coreDir = filepath.Join(root, "apps", "lina-core")
		toolDir = filepath.Join(root, "hack", "tools", "linactl")
	)
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(toolDir, "go.mod"), "module linactl\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = []string{
		"GOWORK=" + filepath.Join(root, "stale.work"),
		"GOFLAGS=-mod=mod -tags=official_plugins,netgo",
	}
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	packageListCount := 0
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "list -m -f {{.Dir}}":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, toolDir)
		case name == "go" && strings.Join(args, " ") == "list -f {{.Dir}} ./...":
			packageListCount++
			if packageListCount == 1 {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", coreDir)
			} else {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", toolDir)
			}
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	input := commandInput{Params: map[string]string{"plugins": "0"}}
	if err := runLintGo(context.Background(), application, input); err != nil {
		t.Fatalf("runLintGo returned error: %v", err)
	}

	expectedConfig := filepath.Join(root, ".golangci.yml")
	expected := []string{
		"go list -m -f {{.Dir}}",
		"go list -f {{.Dir}} ./...",
		"go list -f {{.Dir}} ./...",
		"golangci-lint --version",
		"staticcheck -version",
		"golangci-lint run --config " + expectedConfig + " ./...",
		"staticcheck -checks=U1000 -tests=false .",
		"golangci-lint run --config " + expectedConfig + " ./...",
		"staticcheck -checks=U1000 -tests=false .",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected lint command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	if calls[5].cmd.Dir != coreDir || calls[7].cmd.Dir != toolDir {
		t.Fatalf("unexpected lint command dirs: %q %q", calls[5].cmd.Dir, calls[7].cmd.Dir)
	}
	if got := toolutil.EnvValue(calls[5].cmd.Env, "GOWORK"); got != "" {
		t.Fatalf("expected host-only lint to clear inherited GOWORK, got %q", got)
	}
	if got := toolutil.EnvValue(calls[5].cmd.Env, plugins.SourcePluginsEnvKey); got != "0" {
		t.Fatalf("expected host-only lint to disable source plugin discovery, got %q", got)
	}
	if got := toolutil.EnvValue(calls[5].cmd.Env, "GOFLAGS"); strings.Contains(got, plugins.OfficialBuildTag) {
		t.Fatalf("expected host-only lint to remove official plugin tag, got %q", got)
	}
}

// TestRunLintGoFixAppendsFix verifies automatic formatting remains an explicit
// opt-in flag.
func TestRunLintGoFixAppendsFix(t *testing.T) {
	var (
		root    = t.TempDir()
		coreDir = filepath.Join(root, "apps", "lina-core")
		toolDir = filepath.Join(root, "hack", "tools", "linactl")
	)
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(toolDir, "go.mod"), "module linactl\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	packageListCount := 0
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "list -m -f {{.Dir}}":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, toolDir)
		case name == "go" && strings.Join(args, " ") == "list -f {{.Dir}} ./...":
			packageListCount++
			if packageListCount == 1 {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", coreDir)
			} else {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", toolDir)
			}
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	input := commandInput{Params: map[string]string{"plugins": "0", "fix": "true"}}
	if err := runLintGo(context.Background(), application, input); err != nil {
		t.Fatalf("runLintGo returned error: %v", err)
	}

	expectedArgs := []string{"run", "--config", filepath.Join(root, ".golangci.yml"), "--fix", "./..."}
	if len(calls) < 9 {
		t.Fatalf("expected lint child commands, got %#v", calls)
	}
	if strings.Join(calls[5].args, "\x00") != strings.Join(expectedArgs, "\x00") {
		t.Fatalf("unexpected lint args: got %#v want %#v", calls[5].args, expectedArgs)
	}
}

// TestEnsureGoLintBinaryUsesPathVersion verifies a matching PATH binary is
// reused without running go install.
func TestEnsureGoLintBinaryUsesPathVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		if name == "golangci-lint" && strings.Join(args, " ") == "--version" {
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, version, err := ensureGoLintBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintBinary returned error: %v", err)
	}
	if path != "golangci-lint" || version != "v2.12.2" {
		t.Fatalf("unexpected binary resolution: path=%q version=%q", path, version)
	}
	if got := commandLines(calls); strings.Join(got, "\n") != "golangci-lint --version" {
		t.Fatalf("unexpected command sequence: %#v", got)
	}
}

// TestEnsureGoLintBinaryInstallsMissingBinary verifies a missing PATH binary is
// installed with an environment isolated from repository workspaces and tags.
func TestEnsureGoLintBinaryInstallsMissingBinary(t *testing.T) {
	root := t.TempDir()
	gopath := filepath.Join(root, "gopath")
	expectedBinary := filepath.Join(gopath, "bin", toolutil.ExecutableName("golangci-lint"))
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = []string{
		"GOWORK=" + filepath.Join(root, "temp", "go.work.plugins"),
		"GOFLAGS=-tags=official_plugins",
		"GOOS=wasip1",
		"GOARCH=wasm",
	}
	application.lookPath = func(name string) (string, error) {
		if name == "golangci-lint" {
			return "", errors.New("missing golangci-lint")
		}
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "env GOBIN":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "")
		case name == "go" && strings.Join(args, " ") == "env GOPATH":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", gopath)
		case name == expectedBinary && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, version, err := ensureGoLintBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintBinary returned error: %v", err)
	}
	if path != expectedBinary || version != "v2.12.2" {
		t.Fatalf("unexpected binary resolution: path=%q version=%q", path, version)
	}
	expected := []string{
		"go env GOBIN",
		"go env GOPATH",
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2",
		expectedBinary + " --version",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected install command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	installEnv := calls[2].cmd.Env
	if got := toolutil.EnvValue(installEnv, "GOWORK"); got != "off" {
		t.Fatalf("expected install to force GOWORK=off, got %q", got)
	}
	for _, key := range []string{"GOFLAGS", "GOOS", "GOARCH"} {
		if got := toolutil.EnvValue(installEnv, key); got != "" {
			t.Fatalf("expected install env to remove %s, got %q", key, got)
		}
	}
}

// TestEnsureGoLintBinaryInstallsVersionMismatch verifies a stale PATH binary
// is replaced by the repository-pinned version.
func TestEnsureGoLintBinaryInstallsVersionMismatch(t *testing.T) {
	root := t.TempDir()
	gobin := filepath.Join(root, "bin")
	expectedBinary := filepath.Join(gobin, toolutil.ExecutableName("golangci-lint"))
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.11.0")
		case name == "go" && strings.Join(args, " ") == "env GOBIN":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", gobin)
		case name == expectedBinary && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, _, err := ensureGoLintBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintBinary returned error: %v", err)
	}
	if path != expectedBinary {
		t.Fatalf("expected installed binary path %q, got %q", expectedBinary, path)
	}
	expected := []string{
		"golangci-lint --version",
		"go env GOBIN",
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2",
		expectedBinary + " --version",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected install command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
}

// TestEnsureGoLintBinaryRejectsEmptyPinnedVersion verifies version-lock file
// errors stay actionable before any child process runs.
func TestEnsureGoLintBinaryRejectsEmptyPinnedVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("unexpected child command: %s %s", name, strings.Join(args, " "))
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	_, _, err := ensureGoLintBinary(context.Background(), application)
	if err == nil || !strings.Contains(err.Error(), ".golangci-lint-version is empty") {
		t.Fatalf("expected empty version error, got %v", err)
	}
}

// TestEnsureGoLintStaticcheckBinaryUsesPathVersion verifies a matching PATH
// staticcheck binary is reused without running go install.
func TestEnsureGoLintStaticcheckBinaryUsesPathVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		if name == "staticcheck" && strings.Join(args, " ") == "-version" {
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck 2026.1.1 (0.7.0)")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, version, err := ensureGoLintStaticcheckBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintStaticcheckBinary returned error: %v", err)
	}
	if path != "staticcheck" || version != "v0.7.0" {
		t.Fatalf("unexpected binary resolution: path=%q version=%q", path, version)
	}
	if got := commandLines(calls); strings.Join(got, "\n") != "staticcheck -version" {
		t.Fatalf("unexpected command sequence: %#v", got)
	}
}

// TestEnsureGoLintStaticcheckBinaryInstallsMissingBinary verifies a missing PATH
// staticcheck binary is installed with the same isolated tool environment.
func TestEnsureGoLintStaticcheckBinaryInstallsMissingBinary(t *testing.T) {
	root := t.TempDir()
	gopath := filepath.Join(root, "gopath")
	expectedBinary := filepath.Join(gopath, "bin", toolutil.ExecutableName("staticcheck"))
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = []string{
		"GOWORK=" + filepath.Join(root, "temp", "go.work.plugins"),
		"GOFLAGS=-tags=official_plugins",
		"GOOS=wasip1",
		"GOARCH=wasm",
		"GOWASM=satconv",
	}
	application.lookPath = func(name string) (string, error) {
		if name == "staticcheck" {
			return "", errors.New("missing staticcheck")
		}
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "env GOBIN":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "")
		case name == "go" && strings.Join(args, " ") == "env GOPATH":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", gopath)
		case name == expectedBinary && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, version, err := ensureGoLintStaticcheckBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintStaticcheckBinary returned error: %v", err)
	}
	if path != expectedBinary || version != "v0.7.0" {
		t.Fatalf("unexpected binary resolution: path=%q version=%q", path, version)
	}
	expected := []string{
		"go env GOBIN",
		"go env GOPATH",
		"go install honnef.co/go/tools/cmd/staticcheck@v0.7.0",
		expectedBinary + " -version",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected install command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	installEnv := calls[2].cmd.Env
	if got := toolutil.EnvValue(installEnv, "GOWORK"); got != "off" {
		t.Fatalf("expected install to force GOWORK=off, got %q", got)
	}
	for _, key := range []string{"GOFLAGS", "GOOS", "GOARCH", "GOWASM"} {
		if got := toolutil.EnvValue(installEnv, key); got != "" {
			t.Fatalf("expected install env to remove %s, got %q", key, got)
		}
	}
}

// TestEnsureGoLintStaticcheckBinaryInstallsVersionMismatch verifies a stale PATH
// staticcheck binary is replaced by the repository-pinned version.
func TestEnsureGoLintStaticcheckBinaryInstallsVersionMismatch(t *testing.T) {
	root := t.TempDir()
	gobin := filepath.Join(root, "bin")
	expectedBinary := filepath.Join(gobin, toolutil.ExecutableName("staticcheck"))
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.6.1")
		case name == "go" && strings.Join(args, " ") == "env GOBIN":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", gobin)
		case name == expectedBinary && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck 2026.1.1 (0.7.0)")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	path, _, err := ensureGoLintStaticcheckBinary(context.Background(), application)
	if err != nil {
		t.Fatalf("ensureGoLintStaticcheckBinary returned error: %v", err)
	}
	if path != expectedBinary {
		t.Fatalf("expected installed binary path %q, got %q", expectedBinary, path)
	}
	expected := []string{
		"staticcheck -version",
		"go env GOBIN",
		"go install honnef.co/go/tools/cmd/staticcheck@v0.7.0",
		expectedBinary + " -version",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected install command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
}

// TestEnsureGoLintStaticcheckBinaryRejectsEmptyPinnedVersion verifies the
// staticcheck version-lock file is validated before any child process runs.
func TestEnsureGoLintStaticcheckBinaryRejectsEmptyPinnedVersion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("unexpected child command: %s %s", name, strings.Join(args, " "))
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	_, _, err := ensureGoLintStaticcheckBinary(context.Background(), application)
	if err == nil || !strings.Contains(err.Error(), ".staticcheck-version is empty") {
		t.Fatalf("expected empty version error, got %v", err)
	}
}

func TestGoLintVersionOutputMatches(t *testing.T) {
	if !goLintVersionOutputMatches("golangci-lint has version 2.12.2 built with go1.25.4", "v2.12.2") {
		t.Fatalf("expected version output to match pinned version")
	}
	if goLintVersionOutputMatches("golangci-lint has version 2.12.20 built with go1.25.4", "v2.12.2") {
		t.Fatalf("expected distinct patch version not to match")
	}
	if !goLintVersionOutputMatches("staticcheck 2026.1.1 (0.7.0)", "v0.7.0") {
		t.Fatalf("expected staticcheck release output to match pinned module version")
	}
	if goLintVersionOutputMatches("staticcheck 2026.1.1 (0.7.10)", "v0.7.0") {
		t.Fatalf("expected distinct staticcheck patch version not to match")
	}
}

// TestRunLintGoPluginFullUsesPreparedWorkspace verifies plugin-full lint reuses
// the temporary official plugin Go workspace and official build tag.
func TestRunLintGoPluginFullUsesPreparedWorkspace(t *testing.T) {
	var (
		root      = t.TempDir()
		coreDir   = filepath.Join(root, "apps", "lina-core")
		pluginDir = filepath.Join(root, "apps", "lina-plugins", "plugin-a")
	)
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n\nuse ./apps/lina-core\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(pluginDir, "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginDir, "plugin.yaml"), "id: plugin-a\ntype: source\n")
	writeFile(t, filepath.Join(pluginDir, "backend", "plugin.go"), "package backend\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	packageListCount := 0
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "list -m -f {{.Dir}}":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, pluginDir)
		case name == "go" && strings.Join(args, " ") == "list -f {{.Dir}} ./...":
			packageListCount++
			if packageListCount == 1 {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", coreDir)
			} else {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", pluginDir)
			}
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	input := commandInput{Params: map[string]string{"plugins": "1"}}
	if err := runLintGo(context.Background(), application, input); err != nil {
		t.Fatalf("runLintGo returned error: %v", err)
	}

	workspacePath := filepath.Join(root, "temp", "go.work.plugins")
	if !fileutil.FileExists(workspacePath) {
		t.Fatalf("expected plugin-full lint to prepare %s", workspacePath)
	}
	if got := toolutil.EnvValue(calls[0].cmd.Env, "GOWORK"); got != workspacePath {
		t.Fatalf("expected go list to use plugin workspace, got %q", got)
	}
	if got := toolutil.EnvValue(calls[5].cmd.Env, plugins.SourcePluginsEnvKey); got != "1" {
		t.Fatalf("expected plugin-full lint to enable source plugin discovery, got %q", got)
	}
	if got := toolutil.EnvValue(calls[5].cmd.Env, "GOFLAGS"); !strings.Contains(got, plugins.OfficialBuildTag) {
		t.Fatalf("expected plugin-full lint to add official plugin tag, got %q", got)
	}
}

// TestGoLintModulePlanForDirSeparatesGuestSensitivePackages verifies package
// planning keeps `wasip1`-gated packages out of the host-only dead-code pass.
func TestGoLintModulePlanForDirSeparatesGuestSensitivePackages(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "apps", "lina-core")
	writeFile(t, filepath.Join(moduleDir, "go.mod"), "module lina-core\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(moduleDir, "hostpkg", "host.go"), "package hostpkg\n")
	writeFile(t, filepath.Join(moduleDir, "pkg", "plugin", "pluginbridge", "recordstore", "recordstore.go"), "package recordstore\n\ntype Tx struct{}\n")
	writeFile(t, filepath.Join(moduleDir, "pkg", "plugin", "pluginbridge", "recordstore", "recordstore_exec_stub.go"), "//go:build !wasip1\n\npackage recordstore\n")
	writeFile(t, filepath.Join(moduleDir, "pkg", "plugin", "pluginbridge", "recordstore", "recordstore_exec_wasip1.go"), "//go:build wasip1\n\npackage recordstore\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.env = toolutil.RemoveEnvValue(application.env, "GOWORK")

	plan, err := goLintModulePlanForDir(context.Background(), application, moduleDir)
	if err != nil {
		t.Fatalf("goLintModulePlanForDir returned error: %v", err)
	}
	if !sameStringSet(plan.HostPackages, []string{"./hostpkg"}) {
		t.Fatalf("unexpected host packages: %#v", plan.HostPackages)
	}
	if !sameStringSet(plan.GuestSensitivePackages, []string{"./pkg/plugin/pluginbridge/recordstore"}) {
		t.Fatalf("unexpected guest-sensitive packages: %#v", plan.GuestSensitivePackages)
	}
}

// TestRunLintGoDispatchesGuestSensitiveDeadCodeChecks verifies lint.go checks
// normal packages with U1000 and guest-sensitive packages with the staticcheck
// dead-code matrix.
func TestRunLintGoDispatchesGuestSensitiveDeadCodeChecks(t *testing.T) {
	var (
		root         = t.TempDir()
		coreDir      = filepath.Join(root, "apps", "lina-core")
		toolDir      = filepath.Join(root, "hack", "tools", "linactl")
		guestPkgDir  = filepath.Join(coreDir, "pkg", "plugin", "pluginbridge", "recordstore")
		guestPkgPath = "./pkg/plugin/pluginbridge/recordstore"
	)
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(toolDir, "go.mod"), "module linactl\n")
	writeFile(t, filepath.Join(guestPkgDir, "recordstore_exec_stub.go"), "//go:build !wasip1\n\npackage recordstore\n")
	writeFile(t, filepath.Join(guestPkgDir, "recordstore_exec_wasip1.go"), "//go:build wasip1\n\npackage recordstore\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	packageListCount := 0
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "list -m -f {{.Dir}}":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, toolDir)
		case name == "go" && strings.Join(args, " ") == "list -f {{.Dir}} ./...":
			packageListCount++
			if packageListCount == 1 {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", coreDir, guestPkgDir)
			} else {
				cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", toolDir)
			}
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	if err := runLintGo(context.Background(), application, commandInput{Params: map[string]string{"plugins": "0"}}); err != nil {
		t.Fatalf("runLintGo returned error: %v", err)
	}

	expectedConfig := filepath.Join(root, ".golangci.yml")
	expected := []string{
		"go list -m -f {{.Dir}}",
		"go list -f {{.Dir}} ./...",
		"go list -f {{.Dir}} ./...",
		"golangci-lint --version",
		"staticcheck -version",
		"golangci-lint run --config " + expectedConfig + " ./...",
		"staticcheck -checks=U1000 -tests=false .",
		"staticcheck -checks=U1000 -tests=false -matrix " + guestPkgPath,
		"golangci-lint run --config " + expectedConfig + " ./...",
		"staticcheck -checks=U1000 -tests=false .",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected lint command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	if calls[7].cmd.Stdin == nil {
		t.Fatalf("expected staticcheck matrix command to receive stdin")
	}
}

// TestRunGoLintModulePlanSkipsHostDeadCodeWithoutHostPackages verifies a module
// with only guest-sensitive packages avoids an empty host U1000 invocation.
func TestRunGoLintModulePlanSkipsHostDeadCodeWithoutHostPackages(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "apps", "lina-core")
	writeFile(t, filepath.Join(moduleDir, "go.mod"), "module lina-core\n")
	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	plan := goLintModulePlan{
		ModuleDir:              moduleDir,
		GuestSensitivePackages: []string{"./guest"},
	}
	if err := runGoLintModulePlan(
		context.Background(),
		application,
		nil,
		"goLintBinary",
		"staticcheckBinary",
		filepath.Join(root, ".golangci.yml"),
		false,
		plan,
	); err != nil {
		t.Fatalf("runGoLintModulePlan returned error: %v", err)
	}

	expected := []string{
		"goLintBinary run --config " + filepath.Join(root, ".golangci.yml") + " ./...",
		"staticcheckBinary -checks=U1000 -tests=false -matrix ./guest",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected lint command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	if calls[1].cmd.Stdin == nil {
		t.Fatalf("expected guest-sensitive staticcheck matrix command to receive stdin")
	}
}

// TestGoLintStaticcheckMatrixInput verifies the staticcheck dead-code matrix is
// pinned to the host and wasm guest targets.
func TestGoLintStaticcheckMatrixInput(t *testing.T) {
	const expected = "host:\nwasm_guest: GOOS=wasip1 GOARCH=wasm\n"
	if got := goLintStaticcheckMatrixInput(); got != expected {
		t.Fatalf("unexpected staticcheck matrix input:\ngot:\n%s\nexpected:\n%s", got, expected)
	}
}

// TestGoLintResolveModuleDirMapsPluginRootAndSubdir verifies dir resolution for
// host modules, plugin roots, and nested package directories.
func TestGoLintResolveModuleDirMapsPluginRootAndSubdir(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	coreSubDir := filepath.Join(coreDir, "internal", "service", "auth")
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "demo-plugin")
	pluginBackend := filepath.Join(pluginRoot, "backend")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(coreSubDir, "auth.go"), "package auth\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: demo-plugin\n")
	writeFile(t, filepath.Join(pluginBackend, "go.mod"), "module demo-plugin\n")

	gotCore, err := goLintResolveModuleDir(root, "apps/lina-core")
	if err != nil {
		t.Fatalf("resolve apps/lina-core: %v", err)
	}
	if !samePath(t, gotCore, coreDir) {
		t.Fatalf("apps/lina-core resolved to %s, want %s", gotCore, coreDir)
	}

	gotSub, err := goLintResolveModuleDir(root, filepath.Join("apps", "lina-core", "internal", "service", "auth"))
	if err != nil {
		t.Fatalf("resolve core subdir: %v", err)
	}
	if !samePath(t, gotSub, coreDir) {
		t.Fatalf("core subdir resolved to %s, want %s", gotSub, coreDir)
	}

	gotPlugin, err := goLintResolveModuleDir(root, "apps/lina-plugins/demo-plugin")
	if err != nil {
		t.Fatalf("resolve plugin root: %v", err)
	}
	if !samePath(t, gotPlugin, pluginBackend) {
		t.Fatalf("plugin root resolved to %s, want %s", gotPlugin, pluginBackend)
	}

	if _, err := goLintResolveModuleDir(root, "apps/missing"); err == nil {
		t.Fatalf("expected missing dir to fail")
	}
	if _, err := goLintResolveModuleDir(root, "apps"); err == nil {
		t.Fatalf("expected non-module dir to fail")
	}
	if _, err := goLintResolveModuleDir(root, ""); err == nil {
		t.Fatalf("expected empty dir to fail")
	}
}

// TestGoLintFilterModulesByDirKeepsMatchingModule verifies workspace filtering
// for dir-scoped lint and rejects modules outside the current workspace.
func TestGoLintFilterModulesByDirKeepsMatchingModule(t *testing.T) {
	root := t.TempDir()
	coreDir := filepath.Join(root, "apps", "lina-core")
	toolDir := filepath.Join(root, "hack", "tools", "linactl")
	pluginDir := filepath.Join(root, "apps", "lina-plugins", "demo", "backend")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(toolDir, "go.mod"), "module linactl\n")
	writeFile(t, filepath.Join(pluginDir, "go.mod"), "module demo\n")

	filtered, err := goLintFilterModulesByDir([]string{coreDir, toolDir}, coreDir)
	if err != nil {
		t.Fatalf("filter modules: %v", err)
	}
	if len(filtered) != 1 || !samePath(t, filtered[0], coreDir) {
		t.Fatalf("unexpected filtered modules: %#v", filtered)
	}

	if _, err := goLintFilterModulesByDir([]string{coreDir, toolDir}, pluginDir); err == nil {
		t.Fatalf("expected out-of-workspace module to fail")
	}
}

// TestRunLintGoDirScopesToSingleModule verifies dir= only dispatches lint for
// the resolved module instead of the full workspace list.
func TestRunLintGoDirScopesToSingleModule(t *testing.T) {
	var (
		root    = t.TempDir()
		coreDir = filepath.Join(root, "apps", "lina-core")
		toolDir = filepath.Join(root, "hack", "tools", "linactl")
	)
	writeFile(t, filepath.Join(root, ".golangci-lint-version"), "v2.12.2\n")
	writeFile(t, filepath.Join(root, ".staticcheck-version"), "v0.7.0\n")
	writeFile(t, filepath.Join(coreDir, "go.mod"), "module lina-core\n")
	writeFile(t, filepath.Join(toolDir, "go.mod"), "module linactl\n")

	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		switch {
		case name == "go" && strings.Join(args, " ") == "list -m -f {{.Dir}}":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", coreDir, toolDir)
		case name == "go" && strings.Join(args, " ") == "list -f {{.Dir}} ./...":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", coreDir)
		case name == "golangci-lint" && strings.Join(args, " ") == "--version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "golangci-lint has version 2.12.2")
		case name == "staticcheck" && strings.Join(args, " ") == "-version":
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintLines", "--", "staticcheck v0.7.0")
		}
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}

	if err := runLintGo(context.Background(), application, commandInput{Params: map[string]string{
		"plugins": "0",
		"dir":     "apps/lina-core",
	}}); err != nil {
		t.Fatalf("runLintGo returned error: %v", err)
	}

	expectedConfig := filepath.Join(root, ".golangci.yml")
	expected := []string{
		"go list -m -f {{.Dir}}",
		"go list -f {{.Dir}} ./...",
		"golangci-lint --version",
		"staticcheck -version",
		"golangci-lint run --config " + expectedConfig + " ./...",
		"staticcheck -checks=U1000 -tests=false .",
	}
	if got := commandLines(calls); strings.Join(got, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected lint command sequence:\ngot:\n%s\nexpected:\n%s", strings.Join(got, "\n"), strings.Join(expected, "\n"))
	}
	for _, call := range calls {
		if call.name == "golangci-lint" && strings.Join(call.args, " ") != "--version" {
			if !samePath(t, call.cmd.Dir, coreDir) {
				t.Fatalf("expected golangci-lint to run in %s, got %s", coreDir, call.cmd.Dir)
			}
		}
	}
}

// TestGoLintFileHasWasip1BuildParsesBuildExpressions verifies build-tag
// detection handles equivalent `go:build` expressions instead of relying on a
// single string shape.
func TestGoLintFileHasWasip1BuildParsesBuildExpressions(t *testing.T) {
	cases := map[string]struct {
		content string
		want    bool
	}{
		"wasip1 and wasm": {
			content: "//go:build wasm && wasip1\n\npackage recordstore\n",
			want:    true,
		},
		"not wasip1": {
			content: "//go:build !wasip1\n\npackage recordstore\n",
			want:    true,
		},
		"host only": {
			content: "//go:build linux || darwin\n\npackage recordstore\n",
			want:    false,
		},
		"no build constraint": {
			content: "package recordstore\n",
			want:    false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := goLintFileHasWasip1Build([]byte(tc.content)); got != tc.want {
				t.Fatalf("goLintFileHasWasip1Build() = %t, want %t", got, tc.want)
			}
		})
	}
}

// TestRunLintGoPluginFullRequiresWorkspace verifies explicit plugin-full lint
// fails with the same actionable workspace hint as build and test commands.
func TestRunLintGoPluginFullRequiresWorkspace(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = t.TempDir()
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("unexpected child command: %s %s", name, strings.Join(args, " "))
		return exec.Command(os.Args[0], "-test.run=TestHelperCommandFailure", "--")
	}

	input := commandInput{Params: map[string]string{"plugins": "1"}}
	err := runLintGo(context.Background(), application, input)
	if err == nil {
		t.Fatalf("expected plugin-full lint to fail without official plugin workspace")
	}
	if !strings.Contains(err.Error(), "official plugin workspace") || !strings.Contains(err.Error(), plugins.InitCommand) {
		t.Fatalf("expected actionable plugin workspace error, got %v", err)
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
	var (
		root         = t.TempDir()
		coreDir      = filepath.Join(root, "apps", "lina-core")
		aggregateDir = plugins.AggregateModuleDir(root)
	)
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
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "go.mod"), "module plugin-b\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "plugin.yaml"), "id: plugin-b\ntype: source\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-b", "backend", "plugin.go"), "package backend\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "go.mod"), "module plugin-a\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "plugin.yaml"), "id: plugin-a\ntype: source\n")
	writeFile(t, filepath.Join(pluginRoot, "plugin-a", "backend", "plugin.go"), "package backend\n")

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
	./official-plugins
	../apps/lina-plugins/plugin-a
	../apps/lina-plugins/plugin-b
)
`
	if string(pluginContent) != expected {
		t.Fatalf("unexpected temporary plugin go.work:\n%s", string(pluginContent))
	}
	aggregateGo, err := os.ReadFile(filepath.Join(root, "temp", "official-plugins", "plugins.go"))
	if err != nil {
		t.Fatalf("read aggregate plugins.go: %v", err)
	}
	aggregateText := string(aggregateGo)
	for _, expectedImport := range []string{`_ "plugin-a/backend"`, `_ "plugin-b/backend"`} {
		if !strings.Contains(aggregateText, expectedImport) {
			t.Fatalf("expected aggregate imports to contain %s, got:\n%s", expectedImport, aggregateText)
		}
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
	writeWindowsWrapperAttributes(t, root)
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
	writeWindowsWrapperAttributes(t, root)
	writeFile(t, filepath.Join(root, "hack", "scripts", "legacy.sh"), "#!/usr/bin/env bash\n")

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), "hack/scripts contains legacy script") {
		t.Fatalf("expected legacy script validation error, got %v", err)
	}
}

func TestValidateRepositoryToolingRejectsStaleMakeCmdWorkspaceOverride(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\nset GOWORK=off\r\ngo run . %*\r\n")
	writeWindowsWrapperAttributes(t, root)

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), "must not force GOWORK=off") {
		t.Fatalf("expected stale GOWORK validation error, got %v", err)
	}
}

func TestValidateRepositoryToolingRejectsNonASCIIMakeCmd(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\nREM 用途：Windows wrapper\r\ngo run . %*\r\n")
	writeWindowsWrapperAttributes(t, root)

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), "ASCII-only") {
		t.Fatalf("expected ASCII-only validation error, got %v", err)
	}
}

func TestValidateRepositoryToolingRejectsMissingWindowsWrapperAttributes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "make.cmd"), "@echo off\r\ngo run . %*\r\n")
	writeFile(t, filepath.Join(root, ".gitattributes"), "*.cmd text eol=crlf\n")

	err := repository.ValidateTooling(root, commandNames())
	if err == nil || !strings.Contains(err.Error(), ".gitattributes must set *.bat text eol=crlf") {
		t.Fatalf("expected missing .bat eol validation error, got %v", err)
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

func writeWindowsWrapperAttributes(t *testing.T, root string) {
	t.Helper()
	writeFile(t, filepath.Join(root, ".gitattributes"), "*.cmd text eol=crlf\n*.bat text eol=crlf\n")
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
		`cp -R apps/lina-core/api "$smoke_root/apps/lina-core/api"`,
		`cp -R apps/lina-core/pkg "$smoke_root/apps/lina-core/pkg"`,
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
		`cp -R apps/lina-core/api "$smoke_root/apps/lina-core/api"`,
		`cp -R apps/lina-core/pkg "$smoke_root/apps/lina-core/pkg"`,
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

// TestReusableSmokeWorkflowsDoNotCopyRemovedPluginCapabilityData guards the
// isolated CI smoke fixtures after plugin capability packages are split by
// domain instead of living under the removed capability/data path.
func TestReusableSmokeWorkflowsDoNotCopyRemovedPluginCapabilityData(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	for _, workflow := range []string{
		filepath.Join(root, ".github", "workflows", "reusable-plugin-command-smoke.yml"),
		filepath.Join(root, ".github", "workflows", "reusable-make-command-smoke.yml"),
	} {
		content, readErr := os.ReadFile(workflow)
		if readErr != nil {
			t.Fatalf("read workflow %s: %v", workflow, readErr)
		}
		if strings.Contains(string(content), "apps/lina-core/pkg/plugin/capability/data") {
			t.Fatalf("workflow %s must not copy removed plugin capability/data package", workflow)
		}
	}
}

// TestHostOnlyArtifactSmokeUsesSystemInfoReadiness guards the host-only
// artifact smoke action after the anonymous health endpoint was removed.
func TestHostOnlyArtifactSmokeUsesSystemInfoReadiness(t *testing.T) {
	root, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		t.Fatalf("discover repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".github", "actions", "host-only-artifact-smoke", "action.yml"))
	if err != nil {
		t.Fatalf("read host-only artifact smoke action: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		`/api/v1/auth/login`,
		`/api/v1/system/info`,
		`clusterEnabled`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("host-only artifact smoke action missing %q", expected)
		}
	}
	if strings.Contains(text, "/api/v1/health") {
		t.Fatal("host-only artifact smoke action must not poll the removed health endpoint")
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

func TestCommandRegistryIncludesVersion(t *testing.T) {
	registry := commandRegistry()
	if _, ok := registry["version"]; !ok {
		t.Fatalf("expected command %q to be registered", "version")
	}
}

func TestRunVersionUpdatesMetadataAndReadmeImages(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), `framework:
  name: LinaPro
  version: "v0.3.0"
openapi:
  version: "v0.1.0"
`)
	writeFile(t, filepath.Join(root, "README.md"), `<img src="https://linapro.ai/img/logo.png" width="300" />
[![CI](https://example.com/badge.svg?style=flat)](https://example.com)
![Preview](https://linapro.ai/img/preview.webp?v=0.3.0)
`)
	writeFile(t, filepath.Join(root, "README.zh-CN.md"), `<img src='https://linapro.ai/img/zh-logo.png?old=1&v=0.3.0' />
![Preview](https://linapro.ai/img/zh-preview.webp)
`)

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root

	err := runVersion(context.Background(), application, commandInput{Params: map[string]string{"to": "v1.2.3"}})
	if err != nil {
		t.Fatalf("runVersion returned error: %v", err)
	}

	metadata := readFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"))
	if !strings.Contains(metadata, `  version: "v1.2.3"`) {
		t.Fatalf("framework.version was not updated:\n%s", metadata)
	}
	if !strings.Contains(metadata, `openapi:
  version: "v0.1.0"`) {
		t.Fatalf("non-framework version should not be changed:\n%s", metadata)
	}

	readme := readFile(t, filepath.Join(root, "README.md"))
	for _, fragment := range []string{
		`src="https://linapro.ai/img/logo.png?v=1.2.3"`,
		`https://example.com/badge.svg?style=flat&v=1.2.3`,
		`https://linapro.ai/img/preview.webp?v=1.2.3`,
	} {
		if !strings.Contains(readme, fragment) {
			t.Fatalf("README.md missing %q:\n%s", fragment, readme)
		}
	}

	chineseReadme := readFile(t, filepath.Join(root, "README.zh-CN.md"))
	for _, fragment := range []string{
		`src='https://linapro.ai/img/zh-logo.png?old=1&v=1.2.3'`,
		`https://linapro.ai/img/zh-preview.webp?v=1.2.3`,
	} {
		if !strings.Contains(chineseReadme, fragment) {
			t.Fatalf("README.zh-CN.md missing %q:\n%s", fragment, chineseReadme)
		}
	}
	if !strings.Contains(stdout.String(), "Updated framework.version to v1.2.3") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestRunVersionRejectsInvalidTarget(t *testing.T) {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = t.TempDir()

	err := runVersion(context.Background(), application, commandInput{Params: map[string]string{"to": "1.2.3"}})
	if err == nil || !strings.Contains(err.Error(), "must match vMAJOR.MINOR.PATCH") {
		t.Fatalf("expected invalid target error, got: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

// TestValidatePluginConfigAcceptsStringItemsAndFilters verifies configured
// source and plugin filters keep items as plain string plugin IDs.
func TestValidatePluginConfigAcceptsStringItemsAndFilters(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"custom": {
				Repo:  "https://example.com/custom.git",
				Root:  "apps/lina-plugins",
				Ref:   "main",
				Items: []string{"linapro-content-notice"},
			},
			"official": {
				Repo:  "https://example.com/official.git",
				Root:  ".",
				Ref:   "main",
				Items: []string{"linapro-tenant-core", "linapro-org-core"},
			},
		},
	}

	plan, err := plugins.ValidateConfig(cfg, commandInput{Params: map[string]string{"p": "linapro-org-core", "source": "official"}})
	if err != nil {
		t.Fatalf("plugins.ValidateConfig returned error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Fatalf("expected one filtered item, got %#v", plan.Items)
	}
	item := plan.Items[0]
	if item.ID != "linapro-org-core" || item.Source != "official" || item.Root != "." {
		t.Fatalf("unexpected filtered item: %#v", item)
	}
}

// TestValidatePluginConfigRejectsDuplicatePluginIDs verifies duplicate plugin
// IDs across sources are rejected before any workspace write.
func TestValidatePluginConfigRejectsDuplicatePluginIDs(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"a": {Repo: "repo-a", Root: ".", Ref: "main", Items: []string{"linapro-tenant-core"}},
			"b": {Repo: "repo-b", Root: ".", Ref: "main", Items: []string{"linapro-tenant-core"}},
		},
	}

	_, err := plugins.ValidateConfig(cfg, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "multiple sources") {
		t.Fatalf("expected duplicate plugin validation error, got %v", err)
	}
}

// TestValidatePluginConfigRejectsWildcardMixedWithExplicitIDs verifies a
// source cannot mix "*" with individual plugin IDs.
func TestValidatePluginConfigRejectsWildcardMixedWithExplicitIDs(t *testing.T) {
	cfg := config.Plugins{
		Sources: map[string]config.PluginSource{
			"official": {Repo: "repo", Root: ".", Ref: "main", Items: []string{"*", "linapro-tenant-core"}},
		},
	}

	_, err := plugins.ValidateConfig(cfg, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "cannot mix wildcard") {
		t.Fatalf("expected wildcard mix validation error, got %v", err)
	}
}

// TestValidatePluginSourceRootRejectsUnsafePaths verifies source roots cannot
// escape the remote repository or use platform-specific drive paths.
func TestValidatePluginSourceRootRejectsUnsafePaths(t *testing.T) {
	invalid := []string{"", "..", "../plugins", "/tmp/plugins", `C:\plugins`, "C:/plugins", "apps/../secret", "apps\\plugins"}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			if _, err := plugins.ValidateSourceRoot(value); err == nil {
				t.Fatalf("expected invalid root %q to fail", value)
			}
		})
	}
}

// TestLoadPluginPlanRejectsNonStringItems verifies YAML objects in items fail
// because plugin items must remain a string array.
func TestLoadPluginPlanRejectsNonStringItems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir lina-core: %v", err)
	}
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), `plugins:
  sources:
    official:
      repo: "https://example.com/plugins.git"
      root: "."
      ref: "main"
      items:
        - id: linapro-tenant-core
`)

	_, err := plugins.LoadPlan(root, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected non-string item YAML error, got %v", err)
	}
}

// TestRemoveGitSubmoduleSectionPreservesOtherSections verifies only the plugin
// submodule section is removed from a Git config-style file.
func TestRemoveGitSubmoduleSectionPreservesOtherSections(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".gitmodules")
	writeFile(t, configPath, `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
[submodule "docs"]
	path = docs
	url = https://example.com/docs.git
`)

	if err := plugins.RemoveGitSubmoduleSection(configPath, plugins.ManagedRootRelativePath); err != nil {
		t.Fatalf("plugins.RemoveGitSubmoduleSection returned error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "apps/lina-plugins") {
		t.Fatalf("target submodule section was not removed:\n%s", text)
	}
	if !strings.Contains(text, `[submodule "docs"]`) {
		t.Fatalf("unrelated submodule section was not preserved:\n%s", text)
	}
}

// TestRemoveGitSubmoduleSectionStopsAtAnyNextSection verifies submodule
// removal does not delete following non-submodule Git config sections.
func TestRemoveGitSubmoduleSectionStopsAtAnyNextSection(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config")
	writeFile(t, configPath, `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
[remote "origin"]
	url = https://example.com/project.git
`)

	if err := plugins.RemoveGitSubmoduleSection(configPath, plugins.ManagedRootRelativePath); err != nil {
		t.Fatalf("plugins.RemoveGitSubmoduleSection returned error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "apps/lina-plugins") {
		t.Fatalf("target submodule section was not removed:\n%s", text)
	}
	if !strings.Contains(text, `[core]`) || !strings.Contains(text, `[remote "origin"]`) {
		t.Fatalf("non-submodule sections were not preserved:\n%s", text)
	}
}

// TestRunPluginsInitConvertsGitlinkAndPreservesFiles verifies plugins.init
// removes submodule metadata without deleting plugin files.
func TestRunPluginsInitConvertsGitlinkAndPreservesFiles(t *testing.T) {
	root := newGitRepo(t)
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
[submodule "docs"]
	path = docs
	url = https://example.com/docs.git
`)
	writeFile(t, filepath.Join(root, ".git", "config"), `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "demo", "plugin.yaml"), "id: demo\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	writeFile(t, filepath.Join(root, ".git", "modules", "apps", "lina-plugins", "config"), "[core]\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInit(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInit returned error: %v", err)
	}

	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "demo", "plugin.yaml")) {
		t.Fatalf("plugin file was not preserved")
	}
	gitmodules, err := os.ReadFile(filepath.Join(root, ".gitmodules"))
	if err != nil {
		t.Fatalf("read .gitmodules: %v", err)
	}
	if strings.Contains(string(gitmodules), "apps/lina-plugins") || !strings.Contains(string(gitmodules), `"docs"`) {
		t.Fatalf("unexpected .gitmodules content:\n%s", string(gitmodules))
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.init: %s", stage)
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", ".git")) || fileutil.DirExists(filepath.Join(root, ".git", "modules", "apps", "lina-plugins")) {
		t.Fatalf("submodule metadata was not cleaned")
	}
}

// TestPluginsInstallAutoInitializesSubmoduleWorkspace verifies install runs
// the same workspace initialization as plugins.init before copying plugins.
func TestPluginsInstallAutoInitializesSubmoduleWorkspace(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, ".git", "config"), `[core]
	repositoryformatversion = 0
[submodule "apps/lina-plugins"]
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	writeFile(t, filepath.Join(root, ".git", "modules", "apps", "lina-plugins", "config"), "[core]\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Plugin workspace converted to ordinary directory") || !strings.Contains(output, "Installed plugin linapro-tenant-core") {
		t.Fatalf("expected install to auto-initialize workspace and continue, got:\n%s", output)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml")) {
		t.Fatalf("plugin was not installed after auto initialization")
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.install auto initialization: %s", stage)
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", ".git")) || fileutil.DirExists(filepath.Join(root, ".git", "modules", "apps", "lina-plugins")) {
		t.Fatalf("submodule metadata was not cleaned")
	}
}

// TestPluginsInstallUpdateAndStatusUseConfiguredSources verifies install,
// update, lock writing, and status output against a local source repository.
func TestPluginsInstallUpdateAndStatusUseConfiguredSources(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "README.md"), "v1\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	var installOut bytes.Buffer
	application := newApp(&installOut, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml")) {
		t.Fatalf("plugin was not installed")
	}
	for _, expected := range []string{
		"Preparing plugin installation for 1 configured item(s)...",
		"Installing 1 plugin(s)...",
		"Synchronizing plugin source official",
		"[1/1] installing plugin linapro-tenant-core from official...",
		"Installed plugin linapro-tenant-core",
	} {
		if !strings.Contains(installOut.String(), expected) {
			t.Fatalf("expected install output to contain %q, got:\n%s", expected, installOut.String())
		}
	}
	if fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", ".git")) || !fileutil.FileExists(plugins.LockPath(root)) {
		t.Fatalf("plugin metadata or lock state is incorrect")
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")

	if err := runPluginsInstall(context.Background(), application, commandInput{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected install to reject existing plugin, got %v", err)
	}

	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")
	if err := runPluginsUpdate(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsUpdate returned error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml"))
	if err != nil {
		t.Fatalf("read updated plugin manifest: %v", err)
	}
	if !strings.Contains(string(content), "0.2.0") {
		t.Fatalf("plugin was not updated:\n%s", string(content))
	}

	var statusOut bytes.Buffer
	application.stdout = &statusOut
	if err = runPluginsStatus(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsStatus returned error: %v", err)
	}
	output := statusOut.String()
	for _, expected := range []string{
		"Plugin workspace:",
		"Querying configured plugin sources...",
		"Rendering status for 1 configured plugin(s)...",
		"| Plugin",
		"| Source",
		"| Version",
		"| Installed",
		"| Dirty",
		"| Remote",
		"| linapro-tenant-core",
		"| official",
		"| 0.2.0",
		"| true",
		"| current",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected status output to contain %q, got:\n%s", expected, output)
		}
	}

	var filteredStatusOut bytes.Buffer
	application.stdout = &filteredStatusOut
	if err = runPluginsStatus(context.Background(), application, commandInput{Params: map[string]string{"p": "linapro-tenant-core"}}); err != nil {
		t.Fatalf("filtered runPluginsStatus returned error: %v", err)
	}
	filteredOutput := filteredStatusOut.String()
	if !strings.Contains(filteredOutput, "| linapro-tenant-core") || !strings.Contains(filteredOutput, "| current") {
		t.Fatalf("expected filtered status table to include current plugin row, got:\n%s", filteredOutput)
	}
	if strings.Contains(filteredOutput, "remote=current") {
		t.Fatalf("filtered status output must use table columns, got legacy key-value output:\n%s", filteredOutput)
	}
}

// TestPluginsSourceCacheReusesCheckoutWithFetch verifies plugin source sync
// keeps one reusable checkout and refreshes it through later Git fetches.
func TestPluginsSourceCacheReusesCheckoutWithFetch(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	var firstOut bytes.Buffer
	application := newApp(&firstOut, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	cachePath := plugins.SourceCachePath(root, "official")
	if !fileutil.DirExists(filepath.Join(cachePath, ".git")) {
		t.Fatalf("expected reusable source cache at %s", cachePath)
	}
	assertNoLegacyPluginSourceTemps(t, root)

	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	var updateOut bytes.Buffer
	application.stdout = &updateOut
	if err := runPluginsUpdate(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsUpdate returned error: %v", err)
	}
	output := updateOut.String()
	if strings.Contains(output, "Cloning into") {
		t.Fatalf("expected update to reuse source cache instead of cloning again, got:\n%s", output)
	}
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "plugin.yaml"))
	if err != nil {
		t.Fatalf("read updated plugin manifest: %v", err)
	}
	if !strings.Contains(string(content), "0.2.0") {
		t.Fatalf("plugin update did not fetch latest source content:\n%s", string(content))
	}
	assertNoLegacyPluginSourceTemps(t, root)
}

// TestPluginsInstallExpandsWildcardItems verifies items ["*"] installs every
// plugin directory under the configured source root.
func TestPluginsInstallExpandsWildcardItems(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "plugins", "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "plugins", "linapro-org-core", "plugin.yaml"), "id: linapro-org-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "plugins", "not-plugin", "README.md"), "ignored\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "source plugins")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \"plugins\"\n      ref: \"master\"\n      items:\n        - \"*\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	for _, pluginID := range []string{"linapro-tenant-core", "linapro-org-core"} {
		if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", pluginID, "plugin.yaml")) {
			t.Fatalf("expected wildcard plugin %s to be installed", pluginID)
		}
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "not-plugin")) {
		t.Fatalf("directory without plugin.yaml should not be installed")
	}
}

// TestPluginsInstallWildcardHonorsPluginFilter verifies p=<plugin-id> filters
// the plugins discovered from a wildcard source.
func TestPluginsInstallWildcardHonorsPluginFilter(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	writeFile(t, filepath.Join(source, "linapro-org-core", "plugin.yaml"), "id: linapro-org-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "source plugins")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"*\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{Params: map[string]string{"p": "linapro-org-core"}}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	if !fileutil.FileExists(filepath.Join(root, "apps", "lina-plugins", "linapro-org-core", "plugin.yaml")) {
		t.Fatalf("expected filtered wildcard plugin to be installed")
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core")) {
		t.Fatalf("unexpected unfiltered wildcard plugin installed")
	}
}

// TestRunPluginsUpdateRejectsLocalChangesUnlessForced verifies update protects
// local plugin edits unless the user explicitly passes force=1.
func TestRunPluginsUpdateRejectsLocalChangesUnlessForced(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "local.txt"), "local change\n")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	err := runPluginsUpdate(context.Background(), application, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "local changes") {
		t.Fatalf("expected dirty update rejection, got %v", err)
	}
	if err = runPluginsUpdate(context.Background(), application, commandInput{Params: map[string]string{"force": "1"}}); err != nil {
		t.Fatalf("forced update returned error: %v", err)
	}
}

// TestRunPluginsUpdateRejectsCommittedLockDrift verifies update protects
// committed local plugin edits when they differ from the tool lock hash.
func TestRunPluginsUpdateRejectsCommittedLockDrift(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")

	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsInstall(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsInstall returned error: %v", err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "install plugin")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core", "local.txt"), "committed local change\n")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "local plugin customization")
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.2.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "update plugin")

	err := runPluginsUpdate(context.Background(), application, commandInput{})
	if err == nil || !strings.Contains(err.Error(), "local changes") {
		t.Fatalf("expected committed lock drift rejection, got %v", err)
	}
}

// TestPluginsStatusAutoInitializesSubmoduleWithoutPluginWrites verifies status
// initializes the workspace but still avoids plugin directory and lock writes.
func TestPluginsStatusAutoInitializesSubmoduleWithoutPluginWrites(t *testing.T) {
	root := newGitRepo(t)
	source := newGitRepo(t)
	writeFile(t, filepath.Join(source, "linapro-tenant-core", "plugin.yaml"), "id: linapro-tenant-core\nversion: 0.1.0\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial plugin")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "plugins:\n  sources:\n    official:\n      repo: \""+filepath.ToSlash(source)+"\"\n      root: \".\"\n      ref: \"master\"\n      items:\n        - \"linapro-tenant-core\"\n")
	writeFile(t, filepath.Join(root, ".gitmodules"), `[submodule "apps/lina-plugins"]
	path = apps/lina-plugins
	url = https://example.com/plugins.git
`)
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", ".git"), "gitdir: ../../.git/modules/apps/lina-plugins\n")
	runGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,apps/lina-plugins")

	var stdout bytes.Buffer
	application := newApp(&stdout, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if err := runPluginsStatus(context.Background(), application, commandInput{}); err != nil {
		t.Fatalf("runPluginsStatus returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Plugin workspace converted to ordinary directory") || !strings.Contains(output, "Plugin workspace: apps/lina-plugins (ordinary)") {
		t.Fatalf("expected status to auto-initialize workspace and continue, got:\n%s", output)
	}
	stage := runGitOutput(t, root, "ls-files", "--stage", "--", plugins.ManagedRootRelativePath)
	if strings.Contains(stage, "160000") {
		t.Fatalf("gitlink still exists after plugins.status auto initialization: %s", stage)
	}
	if fileutil.DirExists(filepath.Join(root, "apps", "lina-plugins", "linapro-tenant-core")) {
		t.Fatalf("status must not install plugin directories")
	}
	if fileutil.FileExists(plugins.LockPath(root)) {
		t.Fatalf("status must not write plugin lock state")
	}
}

// newGitRepo creates a minimal repository shaped like a LinaPro checkout.
func newGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	runGit(t, root, "symbolic-ref", "HEAD", "refs/heads/master")
	runGit(t, root, "config", "user.email", "linactl@example.com")
	runGit(t, root, "config", "user.name", "linactl")
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-core"), 0o755); err != nil {
		t.Fatalf("mkdir lina-core: %v", err)
	}
	return root
}

// runGit executes a Git command in a test repository.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

// runGitOutput executes a Git command and returns its combined output.
func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

// assertNoLegacyPluginSourceTemps verifies source sync no longer creates
// one-shot plugin-source-* directories under temp.
func assertNoLegacyPluginSourceTemps(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "temp"))
	if err != nil {
		t.Fatalf("read temp directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "plugin-source-") {
			t.Fatalf("unexpected legacy plugin source temp directory: %s", entry.Name())
		}
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

func requireEnvTool(t *testing.T, tools []envTool, name string) envTool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("env tool %q not found", name)
	return envTool{}
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

// TestHelperEmbeddedGoFrameCtrl invokes the hidden GoFrame bridge in a separate
// process so any GoFrame CLI fatal exit cannot terminate the parent test.
func TestHelperEmbeddedGoFrameCtrl(t *testing.T) {
	if len(os.Args) < 4 || os.Args[len(os.Args)-2] != "--" || os.Getenv("LINACTL_TEST_EMBEDDED_GOFRAME") != "1" {
		return
	}
	application := newApp(os.Stdout, os.Stderr, strings.NewReader(""))
	application.root = os.Args[len(os.Args)-1]
	if err := runEmbeddedGoFrame(context.Background(), application, commandInput{
		Args:   []string{"gen", "ctrl"},
		Params: map[string]string{"config-dir": filepath.Join(application.root, "apps", "lina-core", "hack")},
	}); err != nil {
		t.Fatalf("run embedded GoFrame ctrl: %v", err)
	}
	os.Exit(0)
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

// TestHelperPrintLines prints each supplied argument on its own line for
// deterministic `go list -f` discovery tests.
func TestHelperPrintLines(t *testing.T) {
	index := -1
	for i, arg := range os.Args {
		if arg == "--" {
			index = i
			break
		}
	}
	if index == -1 || index == len(os.Args)-1 {
		return
	}
	for _, line := range os.Args[index+1:] {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
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

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	return strings.Join(leftCopy, "\x00") == strings.Join(rightCopy, "\x00")
}

type capturedCommand struct {
	name string
	args []string
	cmd  *exec.Cmd
}

func commandLines(calls []capturedCommand) []string {
	lines := make([]string, 0, len(calls))
	for _, call := range calls {
		lines = append(lines, call.name+" "+strings.Join(call.args, " "))
	}
	return lines
}

func newLintCommandTestApp(root string, firstModule string, secondModule string, calls *[]capturedCommand) *app {
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		if name == "go" {
			cmd = exec.Command(os.Args[0], "-test.run=TestHelperPrintWorkspaceModules", "--", firstModule, secondModule)
		}
		*calls = append(*calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	return application
}

func writeBuildFixture(t *testing.T, root string) {
	t.Helper()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.25.0\n\nuse ./apps/lina-core\n")
	writeFile(t, filepath.Join(root, "hack", "config.yaml"), "build:\n  platforms:\n    - auto\n")
	writeFile(t, filepath.Join(root, "apps", "lina-vben", "apps", "web-antd", "dist", "index.html"), "host")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "config.template.yaml"), "template: true\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "config", "metadata.yaml"), "framework:\n  version: test\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "sql", "001.sql"), "-- sql\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "manifest", "i18n", "en", "messages.json"), "{}\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox", "plugin.yaml"), "id: john-ai-agentbox\ntype: source\n")
	writeFile(t, filepath.Join(root, "apps", "lina-plugins", "john-ai-agentbox", "hack", "config.yaml"), "build:\n  commands:\n    - pnpm --dir \"$(BUILD_DIR)/frontend\" run build\n")
	writeFile(t, filepath.Join(root, "apps", "lina-core", "go.mod"), "module lina-core\n")
}

func runBuildWithCapturedCommands(t *testing.T, root string, env []string, input commandInput) []capturedCommand {
	t.Helper()
	var calls []capturedCommand
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	if env != nil {
		application.env = env
	}
	application.lookPath = func(name string) (string, error) {
		return name, nil
	}
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	if err := runBuild(context.Background(), application, input); err != nil {
		t.Fatalf("runBuild returned error: %v", err)
	}
	return calls
}

func newGoFrameDispatchTestApp(t *testing.T, root string, executable string) (*app, *[]capturedCommand) {
	t.Helper()
	application := newApp(ioDiscard{}, ioDiscard{}, strings.NewReader(""))
	application.root = root
	application.executable = func() (string, error) {
		return executable, nil
	}
	application.lookPath = func(name string) (string, error) {
		if name == "gf" {
			t.Fatalf("GoFrame generation must not resolve external gf from PATH")
		}
		return name, nil
	}
	var calls []capturedCommand
	application.execCommand = func(_ context.Context, name string, args ...string) *exec.Cmd {
		if name == "gf" {
			t.Fatalf("GoFrame generation must not execute external gf")
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperCommandSuccess", "--")
		calls = append(calls, capturedCommand{
			name: name,
			args: append([]string(nil), args...),
			cmd:  cmd,
		})
		return cmd
	}
	return application, &calls
}

func requireSingleGoFrameDispatch(t *testing.T, calls *[]capturedCommand, expectedDir string, executable string, expectedArgs []string) {
	t.Helper()
	if len(*calls) != 1 {
		t.Fatalf("expected one child command, got %d: %#v", len(*calls), *calls)
	}
	call := (*calls)[0]
	if call.name != executable {
		t.Fatalf("child command name mismatch: got %q want %q", call.name, executable)
	}
	if call.cmd == nil {
		t.Fatalf("child command was not captured")
	}
	if call.cmd.Dir != expectedDir {
		t.Fatalf("child command dir mismatch: got %q want %q", call.cmd.Dir, expectedDir)
	}
	if len(call.args) != len(expectedArgs) {
		t.Fatalf("child command args length mismatch: got %#v want %#v", call.args, expectedArgs)
	}
	for i := range expectedArgs {
		if call.args[i] != expectedArgs[i] {
			t.Fatalf("child command args mismatch: got %#v want %#v", call.args, expectedArgs)
		}
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

// writeFrontendDependencySentinel creates the Vite CLI expected by
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
