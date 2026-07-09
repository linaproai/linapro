// This file verifies the plugin governance scanner with self-contained
// temporary repository fixtures. It avoids depending on the real repository
// migration state so the scanner can be evolved independently of active
// OpenSpec work.

package plugingovernance

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScanFindsPluginGovernanceViolations verifies the scanner catches the
// production governance categories that cannot be delegated to Go's internal
// package visibility rules.
func TestScanFindsPluginGovernanceViolations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), `
id: linapro-demo-dynamic
type: dynamic
hostServices:
  - service: data
    methods: [list]
    resources:
      tables:
        - plugin_linapro_demo_dynamic_record
        - sys_user
        - plugin_linapro_other_record
        - audit_log
  - service: org
    methods:
      - status
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "hack", "config.yaml"), `
gfcli:
  gen:
    dao:
      - tables: >-
          plugin_linapro_demo_dynamic_record,
          sys_user
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "internal", "dao", "internal", "sys_user.go"), "package internal\n")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "internal", "service", "demo.go"), `
package service

func direct() {
	_ = shared.TableSysUser
	_ = g.DB().Model("sys_role")
	_ = db.Raw("select * from sys_menu")
	_ = pluginbridge.Org()
	_ = HostServicesForPlugin()
	_ = HostServiceMethodTenantCurrent
}
`)

	report, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	for _, rule := range []string{
		ruleConfigCoreTable,
		ruleGeneratedCoreTableFile,
		ruleGoSharedCoreTable,
		ruleGoModelCoreTable,
		ruleGoSQLCoreTable,
		ruleLegacyPluginbridgeClient,
		ruleLegacyHostServiceHelper,
		ruleLegacyHostServiceMethod,
		ruleDataCoreTable,
		ruleDataForeignPluginTable,
		ruleDataUnownedTable,
		ruleManifestLegacyMethod,
	} {
		if report.Summary.ByRule[rule] == 0 {
			t.Fatalf("expected rule %s in findings, got %#v", rule, report.Summary.ByRule)
		}
	}
}

// TestScanRejectsLegacyBackendConfig verifies backend/hack/config.yaml is no
// longer an accepted plugin DAO config path.
func TestScanRejectsLegacyBackendConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: linapro-demo-dynamic\n")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "hack", "config.yaml"), `
gfcli:
  gen:
    dao:
      - tables: plugin_linapro_demo_dynamic_record
`)

	report, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if report.Summary.ByRule[ruleConfigLegacyBackendPath] != 1 {
		t.Fatalf("expected legacy config path finding, got %#v", report.Summary.ByRule)
	}
	if report.Summary.ConfigFiles != 0 {
		t.Fatalf("legacy backend config must not count as valid config file, got %d", report.Summary.ConfigFiles)
	}
}

// TestScanRejectsGeneratedDAOWithoutRootConfig verifies generated DAO artifacts
// remain reproducible from plugin-root hack/config.yaml.
func TestScanRejectsGeneratedDAOWithoutRootConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), "id: linapro-demo-dynamic\n")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "internal", "dao", "demo.go"), "package dao\n")

	report, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if report.Summary.ByRule[ruleConfigMissingRootConfig] != 1 {
		t.Fatalf("expected missing root config finding, got %#v", report.Summary.ByRule)
	}
}

// TestScanAllowsHostCoreInternalSyntaxAlreadyBlockedByGo verifies the scanner
// does not duplicate checks for impossible plugin imports or type references
// already rejected by Go's internal package boundary.
func TestScanAllowsHostCoreInternalSyntaxAlreadyBlockedByGo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), `
id: linapro-demo-dynamic
type: dynamic
hostServices:
  - service: data
    methods: [list]
    resources:
      tables:
        - plugin_linapro_demo_dynamic_record
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "internal", "service", "impossible.go"), `
package service

import (
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

func impossible() {
	_ = dao.SysUser
	_ = do.SysUser{}
	_ = entity.SysUser{}
}
`)

	report, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings for Go-internal compiler boundary syntax, got %#v", report.Findings)
	}
}

// TestScanAllowsControlledNonProductionExceptions verifies tests, mock-data SQL,
// install SQL, and migration paths remain outside production runtime scanning.
func TestScanAllowsControlledNonProductionExceptions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), `
id: linapro-demo-dynamic
type: dynamic
hostServices:
  - service: data
    methods: [list]
    resources:
      tables:
        - plugin_linapro_demo_dynamic_record
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "hack", "config.yaml"), `
gfcli:
  gen:
    dao:
      - tables: plugin_linapro_demo_dynamic_record
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "backend", "internal", "service", "demo_test.go"), `
package service

func testOnly() {
	_ = dao.SysUser
	_ = g.DB().Model("sys_user")
}
`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "hack", "tests", "e2e", "TC001-demo.ts"), `const table = "sys_user";`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "manifest", "sql", "001-install.sql"), `select * from sys_user;`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "manifest", "sql", "mock-data", "001-mock.sql"), `insert into sys_user(id) values (1);`)
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "migrations", "001.sql"), `update sys_user set id = id;`)

	report, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings for controlled exceptions, got %#v", report.Findings)
	}
}

// TestRunCheckReportsTextAndError verifies the linactl command wrapper can
// surface scanner findings without shell-specific behavior.
func TestRunCheckReportsTextAndError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pluginRoot := filepath.Join(root, "apps", "lina-plugins", "linapro-demo-dynamic")
	writePluginGovernanceFile(t, filepath.Join(pluginRoot, "plugin.yaml"), `
id: linapro-demo-dynamic
hostServices:
  - service: data
    resources:
      tables:
        - sys_user
`)

	var output bytes.Buffer
	err := RunCheck(root, &output, Options{})
	if err == nil {
		t.Fatal("expected RunCheck to fail when findings exist")
	}
	text := output.String()
	for _, expected := range []string{"Plugin check summary", ruleDataCoreTable, "sys_user"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, text)
		}
	}
}

// writePluginGovernanceFile writes one fixture file and creates parent
// directories for self-contained scanner tests.
func writePluginGovernanceFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimLeft(content, "\n")), 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", path, err)
	}
}
