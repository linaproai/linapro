// This file covers runtime migration replay behaviors owned by migration.

package migration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
)

// TestExecuteManifestSQLFilesReplaysInstallSQL verifies that install SQL can be
// replayed after uninstall and still records a clean migration sequence.
func TestExecuteManifestSQLFilesReplaysInstallSQL(t *testing.T) {
	services := testutil.NewServices()
	service := services.Migration
	ctx := context.Background()

	pluginID := "plugin-dev-dynamic-reinstall"
	tableName := "plugin_runtime_reinstall_log"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Reinstall Plugin",
		"v0.9.1",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-reinstall-create.sql",
				Content: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
			},
			{
				Key:     "002-plugin-dev-dynamic-reinstall-seed.sql",
				Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('install-ran');", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-reinstall.sql",
				Content: fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			},
		},
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	dropTestTableIfExists(t, ctx, tableName)
	t.Cleanup(func() {
		dropTestTableIfExists(t, ctx, tableName)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected first install to succeed, got error: %v", err)
	}
	assertTestTableRowCount(t, ctx, tableName, 1)

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionUninstall); err != nil {
		t.Fatalf("expected uninstall to succeed, got error: %v", err)
	}
	assertTestTableMissing(t, ctx, tableName)

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected reinstall to succeed, got error: %v", err)
	}
	assertTestTableRowCount(t, ctx, tableName, 1)
}

// TestExecuteManifestSQLFilesRollsBackSQLAndLedgerOnSQLFailure verifies that a
// failed lifecycle SQL file leaves neither partially applied data nor migration
// ledger rows behind.
func TestExecuteManifestSQLFilesRollsBackSQLAndLedgerOnSQLFailure(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID  = "plugin-dev-dynamic-sql-failure-rollback"
		tableName = "plugin_sql_failure_rollback_log"
	)
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"SQL Failure Rollback Plugin",
		"v0.9.2",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-sql-failure-rollback-create.sql",
				Content: fmt.Sprintf("CREATE TABLE %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
			},
			{
				Key:     "002-plugin-dev-dynamic-sql-failure-rollback-broken.sql",
				Content: "INSERT INTO this_table_does_not_exist (id) VALUES (42);",
			},
		},
		nil,
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	dropTestTableIfExists(t, ctx, tableName)
	t.Cleanup(func() {
		dropTestTableIfExists(t, ctx, tableName)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	err = services.Migration.ExecuteManifestSQLFiles(ctx, manifest, plugintypes.MigrationDirectionInstall)
	if err == nil {
		t.Fatalf("expected install SQL to fail")
	}
	if !strings.Contains(err.Error(), "002-plugin-dev-dynamic-sql-failure-rollback-broken.sql") {
		t.Fatalf("expected failure to mention the broken SQL file, got: %v", err)
	}
	assertTestTableMissing(t, ctx, tableName)
	assertMigrationRowCount(t, ctx, pluginID, plugintypes.MigrationDirectionInstall, 0)
}

// TestExecuteManifestSQLFilesRollsBackSQLOnLedgerFailure verifies that a
// migration ledger write failure rolls back SQL statements that already ran in
// the same lifecycle transaction.
func TestExecuteManifestSQLFilesRollsBackSQLOnLedgerFailure(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID  = "plugin-dev-dynamic-ledger-failure-rollback"
		tableName = "plugin_ledger_failure_rollback_log"
	)
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Ledger Failure Rollback Plugin",
		"v0.9.3",
		nil,
		nil,
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}
	manifest.RuntimeArtifact.InstallSQLAssets = []*catalog.ArtifactSQLAsset{
		{
			Key:     "001-plugin-dev-dynamic-ledger-failure-rollback-create.sql",
			Content: fmt.Sprintf("CREATE TABLE %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
		},
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	dropTestTableIfExists(t, ctx, tableName)
	t.Cleanup(func() {
		dropTestTableIfExists(t, ctx, tableName)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	oversizedPhase := plugintypes.MigrationDirection(strings.Repeat("ledger-failure-", 8))
	err = services.Migration.ExecuteManifestSQLFiles(ctx, manifest, oversizedPhase)
	if err == nil {
		t.Fatalf("expected migration ledger write to fail")
	}
	assertTestTableMissing(t, ctx, tableName)
	assertMigrationRowCount(t, ctx, pluginID, oversizedPhase, 0)
}

// dropTestTableIfExists removes the temporary table used by migration replay tests.
func dropTestTableIfExists(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	if _, err := g.DB().Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)); err != nil {
		t.Fatalf("expected test table cleanup to succeed, got error: %v", err)
	}
}

// assertTestTableMissing verifies that the migration uninstall step removed the table.
func assertTestTableMissing(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	all, err := g.DB().GetAll(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1;", tableName)
	if err != nil {
		t.Fatalf("expected table existence query to succeed, got error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected table %s to be dropped, got rows: %#v", tableName, all)
	}
}

// assertMigrationRowCount verifies lifecycle migration ledger rows for one phase.
func assertMigrationRowCount(t *testing.T, ctx context.Context, pluginID string, phase plugintypes.MigrationDirection, expected int) {
	t.Helper()

	value, err := g.DB().GetValue(ctx, "SELECT COUNT(1) FROM sys_plugin_migration WHERE plugin_id = ? AND phase = ?;", pluginID, phase.String())
	if err != nil {
		t.Fatalf("expected migration ledger count query to succeed, got error: %v", err)
	}
	if value.Int() != expected {
		t.Fatalf("expected %d migration rows for plugin=%s phase=%s, got %d", expected, pluginID, phase.String(), value.Int())
	}
}

// assertTestTableRowCount verifies the row count produced by replayed migration SQL.
func assertTestTableRowCount(t *testing.T, ctx context.Context, tableName string, expected int) {
	t.Helper()

	value, err := g.DB().GetValue(ctx, fmt.Sprintf("SELECT COUNT(1) FROM %s;", tableName))
	if err != nil {
		t.Fatalf("expected row count query to succeed, got error: %v", err)
	}
	if value.Int() != expected {
		t.Fatalf("expected table %s to contain %d rows, got %d", tableName, expected, value.Int())
	}
}
