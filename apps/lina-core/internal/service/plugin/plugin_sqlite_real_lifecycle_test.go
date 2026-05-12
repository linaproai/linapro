// This file verifies source and dynamic plugin lifecycle contracts on SQLite.

package plugin

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/internal/packed"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/dialect"
)

const sqliteRealLifecycleChildEnv = "LINA_SQLITE_REAL_PLUGIN_LIFECYCLE_CHILD"

// TestSQLitePluginFacadeLifecycle verifies source and dynamic plugin lifecycle
// flows in a child process so SQLite-specific GoFrame globals never leak into the
// default plugin test package database.
func TestSQLitePluginFacadeLifecycle(t *testing.T) {
	if os.Getenv(sqliteRealLifecycleChildEnv) == "1" {
		t.Skip("parent test only launches the isolated SQLite lifecycle child process")
	}

	dbPath := filepath.Join(t.TempDir(), "linapro-plugin-facade-lifecycle.db")
	cmd := exec.Command(os.Args[0], "-test.run=^TestSQLitePluginFacadeLifecycleChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(),
		sqliteRealLifecycleChildEnv+"=1",
		"LINA_SQLITE_REAL_PLUGIN_LIFECYCLE_DB="+dbPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("SQLite plugin facade lifecycle child test failed: %v\n%s", err, string(output))
	}
}

// TestSQLitePluginFacadeLifecycleChild performs the plugin lifecycle checks
// against one temporary SQLite database initialized from the embedded host SQL.
func TestSQLitePluginFacadeLifecycleChild(t *testing.T) {
	if os.Getenv(sqliteRealLifecycleChildEnv) != "1" {
		t.Skip("SQLite lifecycle child test is executed by TestSQLitePluginFacadeLifecycle")
	}

	ctx := context.Background()
	dbPath := strings.TrimSpace(os.Getenv("LINA_SQLITE_REAL_PLUGIN_LIFECYCLE_DB"))
	if dbPath == "" {
		t.Fatal("LINA_SQLITE_REAL_PLUGIN_LIFECYCLE_DB must be set")
	}
	setupSQLiteRealPluginDatabase(t, ctx, "sqlite::@file("+dbPath+")")

	service := newTestService()
	sourcePluginID, sourceTableName := createSQLiteLifecycleSourcePlugin(t)
	verifySQLitePluginLifecycle(t, ctx, service, sourcePluginID, []string{sourceTableName})

	dynamicPluginID, dynamicTableName := createSQLiteLifecycleDynamicPlugin(t)
	verifySQLitePluginLifecycle(t, ctx, service, dynamicPluginID, []string{dynamicTableName})
}

// setupSQLiteRealPluginDatabase points GoFrame at one temporary SQLite file and
// initializes the embedded host DDL/seed assets through the public dialect path.
func setupSQLiteRealPluginDatabase(t *testing.T, ctx context.Context, link string) {
	t.Helper()

	dbDialect, err := dialect.From(link)
	if err != nil {
		t.Fatalf("resolve SQLite dialect: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		t.Fatalf("prepare SQLite real plugin database: %v", err)
	}
	if err = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure GoFrame SQLite database: %v", err)
	}
	adapter, err := gcfg.NewAdapterContent(sqliteRealPluginConfig(link))
	if err != nil {
		t.Fatalf("create SQLite config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	assets, err := readEmbeddedHostSQLAssets()
	if err != nil {
		t.Fatalf("read embedded host SQL assets: %v", err)
	}
	for _, asset := range assets {
		translated, translateErr := dbDialect.TranslateDDL(ctx, asset.path, asset.content)
		if translateErr != nil {
			t.Fatalf("translate embedded host SQL asset %s: %v", asset.path, translateErr)
		}
		for index, statement := range dialect.SplitSQLStatements(translated) {
			if _, err = g.DB().Exec(ctx, statement); err != nil {
				t.Fatalf("execute embedded host SQL asset %s statement %d: %v\n%s", asset.path, index+1, err, statement)
			}
		}
	}
}

// sqliteRealPluginConfig returns the minimal runtime config required by plugin
// facade lifecycle paths while keeping database.default.link as the only
// database dialect source.
func sqliteRealPluginConfig(link string) string {
	return `database:
  default:
    link: "` + link + `"
jwt:
  secret: "sqlite-real-plugin-test-secret"
  expire: 24h
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
cluster:
  enabled: false
upload:
  path: "temp/upload"
  maxSize: 20
plugin:
  dynamic:
    storagePath: "temp/output"
  autoEnable: []
`
}

// embeddedSQLAsset stores one ordered embedded SQL file.
type embeddedSQLAsset struct {
	path    string
	content string
}

// readEmbeddedHostSQLAssets loads embedded host SQL delivery files in lexical
// order, excluding mock-data assets so lifecycle checks start from seed-only state.
func readEmbeddedHostSQLAssets() ([]embeddedSQLAsset, error) {
	entries, err := fs.ReadDir(packed.Files, "manifest/sql")
	if err != nil {
		return nil, err
	}
	assets := make([]embeddedSQLAsset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		assetPath := path.Join("manifest/sql", entry.Name())
		content, readErr := fs.ReadFile(packed.Files, assetPath)
		if readErr != nil {
			return nil, readErr
		}
		assets = append(assets, embeddedSQLAsset{
			path:    assetPath,
			content: string(content),
		})
	}
	sort.SliceStable(assets, func(i int, j int) bool {
		return assets[i].path < assets[j].path
	})
	return assets, nil
}

// createSQLiteLifecycleSourcePlugin creates one source-plugin fixture with SQL
// assets that exercise install and uninstall table transitions.
func createSQLiteLifecycleSourcePlugin(t *testing.T) (string, string) {
	t.Helper()

	const (
		pluginID  = "plugin-sqlite-source-lifecycle"
		tableName = "plugin_sqlite_source_lifecycle"
	)
	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"),
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-"+pluginID+".sql"),
		fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
	)
	return pluginID, tableName
}

// createSQLiteLifecycleDynamicPlugin creates one dynamic runtime artifact with
// SQL assets that exercise install and uninstall table transitions.
func createSQLiteLifecycleDynamicPlugin(t *testing.T) (string, string) {
	t.Helper()

	const (
		pluginID  = "plugin-sqlite-dynamic-lifecycle"
		tableName = "plugin_sqlite_dynamic_lifecycle"
	)
	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"SQLite Dynamic Lifecycle Plugin",
		"v0.1.0",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-sqlite-dynamic-lifecycle.sql",
				Content: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-sqlite-dynamic-lifecycle-uninstall.sql",
				Content: fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			},
		},
	)
	return pluginID, tableName
}

// verifySQLitePluginLifecycle exercises install, enable, disable, and uninstall
// for one plugin and asserts its SQL-owned tables follow the lifecycle.
func verifySQLitePluginLifecycle(
	t *testing.T,
	ctx context.Context,
	service *serviceImpl,
	pluginID string,
	tableNames []string,
) {
	t.Helper()

	if err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("install SQLite plugin %s: %v", pluginID, err)
	}
	assertSQLitePluginRegistryState(t, ctx, service, pluginID, catalog.InstalledYes, catalog.StatusDisabled)
	assertSQLitePluginTablesExist(t, ctx, tableNames)

	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("enable SQLite plugin %s: %v", pluginID, err)
	}
	assertSQLitePluginRegistryState(t, ctx, service, pluginID, catalog.InstalledYes, catalog.StatusEnabled)

	if err := service.Disable(ctx, pluginID); err != nil {
		t.Fatalf("disable SQLite plugin %s: %v", pluginID, err)
	}
	assertSQLitePluginRegistryState(t, ctx, service, pluginID, catalog.InstalledYes, catalog.StatusDisabled)

	if err := service.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("uninstall SQLite plugin %s: %v", pluginID, err)
	}
	assertSQLitePluginRegistryState(t, ctx, service, pluginID, catalog.InstalledNo, catalog.StatusDisabled)
	assertSQLitePluginTablesMissing(t, ctx, tableNames)
}

// assertSQLitePluginRegistryState verifies the registry row for one plugin.
func assertSQLitePluginRegistryState(
	t *testing.T,
	ctx context.Context,
	service *serviceImpl,
	pluginID string,
	installed int,
	status int,
) {
	t.Helper()

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("query SQLite plugin registry %s: %v", pluginID, err)
	}
	if registry == nil {
		t.Fatalf("expected SQLite plugin registry row for %s", pluginID)
	}
	if registry.Installed != installed || registry.Status != status {
		t.Fatalf("expected SQLite plugin %s installed=%d status=%d, got installed=%d status=%d", pluginID, installed, status, registry.Installed, registry.Status)
	}
}

// assertSQLitePluginTablesExist verifies every table name exists in sqlite_master.
func assertSQLitePluginTablesExist(t *testing.T, ctx context.Context, tableNames []string) {
	t.Helper()

	for _, tableName := range tableNames {
		if !sqlitePluginTableExists(t, ctx, tableName) {
			t.Fatalf("expected SQLite plugin table %s to exist", tableName)
		}
	}
}

// assertSQLitePluginTablesMissing verifies every table name has been dropped.
func assertSQLitePluginTablesMissing(t *testing.T, ctx context.Context, tableNames []string) {
	t.Helper()

	for _, tableName := range tableNames {
		if sqlitePluginTableExists(t, ctx, tableName) {
			t.Fatalf("expected SQLite plugin table %s to be dropped", tableName)
		}
	}
}

// sqlitePluginTableExists reports whether one SQLite table exists.
func sqlitePluginTableExists(t *testing.T, ctx context.Context, tableName string) bool {
	t.Helper()

	count, err := g.DB().GetValue(ctx, "SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?", tableName)
	if err != nil {
		t.Fatalf("query SQLite plugin table %s existence: %v", tableName, err)
	}
	return count.Int() > 0
}
