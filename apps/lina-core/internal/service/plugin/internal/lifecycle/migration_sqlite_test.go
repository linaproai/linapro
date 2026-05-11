// This file verifies plugin lifecycle SQL execution against SQLite without
// mutating the package's default PostgreSQL-oriented test database.

package lifecycle_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/dialect"
)

const sqliteLifecycleChildEnv = "LINA_SQLITE_PLUGIN_LIFECYCLE_CHILD"

// TestSQLitePluginLifecycleSQL executes the SQLite lifecycle integration test in
// a child process so global GoFrame database/config state cannot leak into the
// rest of the lifecycle test package.
func TestSQLitePluginLifecycleSQL(t *testing.T) {
	if os.Getenv(sqliteLifecycleChildEnv) == "1" {
		t.Skip("parent test only launches the isolated SQLite child process")
	}

	dbPath := filepath.Join(t.TempDir(), "linapro-plugin-lifecycle.db")
	cmd := exec.Command(os.Args[0], "-test.run=^TestSQLitePluginLifecycleSQLChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(),
		sqliteLifecycleChildEnv+"=1",
		"LINA_SQLITE_PLUGIN_LIFECYCLE_DB="+dbPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("SQLite plugin lifecycle child test failed: %v\n%s", err, string(output))
	}
}

// TestSQLitePluginLifecycleSQLChild performs the actual SQLite lifecycle test.
// It is intentionally skipped unless launched by TestSQLitePluginLifecycleSQL.
func TestSQLitePluginLifecycleSQLChild(t *testing.T) {
	if os.Getenv(sqliteLifecycleChildEnv) != "1" {
		t.Skip("SQLite lifecycle child test is executed by TestSQLitePluginLifecycleSQL")
	}

	ctx := context.Background()
	dbPath := strings.TrimSpace(os.Getenv("LINA_SQLITE_PLUGIN_LIFECYCLE_DB"))
	if dbPath == "" {
		t.Fatal("LINA_SQLITE_PLUGIN_LIFECYCLE_DB must be set")
	}
	link := "sqlite::@file(" + dbPath + ")"

	setupSQLitePluginLifecycleDatabase(t, ctx, link)
	services := testutil.NewServices()

	const (
		pluginID  = "plugin-sqlite-lifecycle"
		tableName = "plugin_sqlite_lifecycle_log"
	)
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"SQLite Lifecycle Plugin",
		"v0.1.0",
		[]*catalog.ArtifactSQLAsset{
			{
				Key: "001-plugin-sqlite-lifecycle-create.sql",
				Content: fmt.Sprintf(
					"CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);",
					tableName,
				),
			},
			{
				Key:     "002-plugin-sqlite-lifecycle-install.sql",
				Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('install');", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-sqlite-lifecycle-uninstall.sql",
				Content: fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			},
		},
	)
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("load dynamic plugin artifact: %v", err)
	}
	manifest.RuntimeArtifact.MockSQLAssets = []*catalog.ArtifactSQLAsset{
		{
			Key:     "001-plugin-sqlite-lifecycle-mock.sql",
			Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('mock');", tableName),
		},
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("sync SQLite plugin manifest: %v", err)
	}

	if err = services.Lifecycle.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		t.Fatalf("execute SQLite plugin install SQL: %v", err)
	}
	assertSQLiteLifecycleMarkers(t, ctx, tableName, []string{"install"})

	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("mark SQLite plugin installed: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("enable SQLite plugin: %v", err)
	}

	var mockResultErr error
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result := services.Lifecycle.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		mockResultErr = result.Err
		return result.Err
	})
	if txErr != nil {
		t.Fatalf("execute SQLite plugin mock SQL transaction: %v", txErr)
	}
	if mockResultErr != nil {
		t.Fatalf("execute SQLite plugin mock SQL result: %v", mockResultErr)
	}
	assertSQLiteLifecycleMarkers(t, ctx, tableName, []string{"install", "mock"})

	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusDisabled); err != nil {
		t.Fatalf("disable SQLite plugin: %v", err)
	}
	if err = services.Lifecycle.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
		t.Fatalf("execute SQLite plugin uninstall SQL: %v", err)
	}
	assertSQLiteLifecycleTableMissing(t, ctx, tableName)
	assertSQLiteLifecycleMigrationRows(t, ctx, pluginID)
}

// setupSQLitePluginLifecycleDatabase points GoFrame at one temporary SQLite
// database and initializes only the plugin governance tables used by this test.
func setupSQLitePluginLifecycleDatabase(t *testing.T, ctx context.Context, link string) {
	t.Helper()

	dbDialect, err := dialect.From(link)
	if err != nil {
		t.Fatalf("resolve SQLite dialect: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		t.Fatalf("prepare SQLite lifecycle database: %v", err)
	}
	if err = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure GoFrame SQLite database: %v", err)
	}
	adapter, err := gcfg.NewAdapterContent("database:\n  default:\n    link: \"" + link + "\"\n")
	if err != nil {
		t.Fatalf("create SQLite config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, "apps", "lina-core", "manifest", "sql", "008-plugin-framework.sql"))
	if err != nil {
		t.Fatalf("read plugin governance SQL: %v", err)
	}
	translated, err := dbDialect.TranslateDDL(ctx, "apps/lina-core/manifest/sql/008-plugin-framework.sql", string(content))
	if err != nil {
		t.Fatalf("translate plugin governance SQL: %v", err)
	}
	for index, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("execute plugin governance SQL statement %d: %v\n%s", index+1, err, statement)
		}
	}
}

// assertSQLiteLifecycleMarkers verifies the plugin SQL assets inserted the
// expected marker sequence into the SQLite business table.
func assertSQLiteLifecycleMarkers(t *testing.T, ctx context.Context, tableName string, expected []string) {
	t.Helper()

	rows, err := g.DB().GetAll(ctx, fmt.Sprintf("SELECT marker FROM %s ORDER BY id;", tableName))
	if err != nil {
		t.Fatalf("query SQLite lifecycle markers: %v", err)
	}
	if len(rows) != len(expected) {
		t.Fatalf("expected %d SQLite lifecycle markers, got %d: %#v", len(expected), len(rows), rows)
	}
	for index, marker := range expected {
		if rows[index]["marker"].String() != marker {
			t.Fatalf("expected SQLite lifecycle marker %d to be %s, got %s", index, marker, rows[index]["marker"].String())
		}
	}
}

// assertSQLiteLifecycleTableMissing verifies uninstall SQL removed the plugin
// business table from SQLite.
func assertSQLiteLifecycleTableMissing(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	count, err := g.DB().GetValue(ctx, "SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?", tableName)
	if err != nil {
		t.Fatalf("query SQLite lifecycle table existence: %v", err)
	}
	if count.Int() != 0 {
		t.Fatalf("expected SQLite lifecycle table %s to be dropped, got count=%d", tableName, count.Int())
	}
}

// assertSQLiteLifecycleMigrationRows verifies install, mock, and uninstall
// phases were all recorded in the plugin migration ledger.
func assertSQLiteLifecycleMigrationRows(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	rows, err := g.DB().GetAll(
		ctx,
		"SELECT phase, status FROM sys_plugin_migration WHERE plugin_id = ? ORDER BY phase, execution_order;",
		pluginID,
	)
	if err != nil {
		t.Fatalf("query SQLite lifecycle migration ledger: %v", err)
	}

	phaseCounts := map[string]int{}
	for _, row := range rows {
		if row["status"].String() != catalog.MigrationExecutionStatusSucceeded.String() {
			t.Fatalf("expected SQLite lifecycle migration row to succeed, got: %#v", row)
		}
		phaseCounts[row["phase"].String()]++
	}
	expectedCounts := map[string]int{
		catalog.MigrationDirectionInstall.String():   2,
		catalog.MigrationDirectionMock.String():      1,
		catalog.MigrationDirectionUninstall.String(): 1,
	}
	for phase, expected := range expectedCounts {
		if phaseCounts[phase] != expected {
			t.Fatalf("expected SQLite lifecycle phase %s to have %d rows, got %d: %#v", phase, expected, phaseCounts[phase], rows)
		}
	}
}
