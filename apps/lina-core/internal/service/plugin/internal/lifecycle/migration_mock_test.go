// This file covers the mock-data SQL execution surface added by the
// plugin-install-with-mock-data change: the rejection guard on
// ExecuteManifestSQLFiles, the resolver branch for the mock direction, and
// the transactional rollback semantics of ExecuteManifestMockSQLFilesInTx.

package lifecycle_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/testutil"
)

// TestExecuteManifestSQLFilesRejectsMockDirection verifies that the public
// ExecuteManifestSQLFiles entry point refuses the mock direction so callers
// cannot accidentally bypass the transactional execution path.
func TestExecuteManifestSQLFilesRejectsMockDirection(t *testing.T) {
	services := testutil.NewServices()

	err := services.Lifecycle.ExecuteManifestSQLFiles(
		context.Background(),
		&catalog.Manifest{ID: "plugin-mock-rejection"},
		catalog.MigrationDirectionMock,
	)
	if err == nil {
		t.Fatalf("expected mock direction rejection, got nil error")
	}
	if !strings.Contains(err.Error(), "ExecuteManifestMockSQLFilesInTx") {
		t.Fatalf("expected error to point callers at the transactional entry, got: %v", err)
	}
}

// TestResolveSQLAssetsHandlesMockDirection verifies that the lifecycle
// resolver locates mock-data SQL when callers ask for the mock direction and
// continues to keep install/uninstall scans disjoint.
func TestResolveSQLAssetsHandlesMockDirection(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-mock-resolver")

	mockDir := filepath.Join(pluginDir, "manifest", "sql", "mock-data")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		t.Fatalf("failed to create mock-data dir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(mockDir, "001-plugin-mock-resolver.sql"),
		[]byte("INSERT INTO sys_user(username) VALUES ('demo') ON CONFLICT DO NOTHING;"),
		0o644,
	); err != nil {
		t.Fatalf("failed to write mock SQL: %v", err)
	}

	manifest := &catalog.Manifest{
		ID:           "plugin-mock-resolver",
		Name:         "Mock Resolver Plugin",
		Version:      "0.1.0",
		Type:         catalog.TypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}

	assets, err := services.Lifecycle.ResolveSQLAssets(manifest, catalog.MigrationDirectionMock)
	if err != nil {
		t.Fatalf("expected mock resolver to succeed, got error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected one mock asset, got %d: %#v", len(assets), assets)
	}
	if assets[0].Key != "001-plugin-mock-resolver.sql" {
		t.Fatalf("unexpected mock asset key: %s", assets[0].Key)
	}
	if !strings.Contains(assets[0].Content, "INSERT INTO sys_user") {
		t.Fatalf("unexpected mock asset content: %s", assets[0].Content)
	}
}

// TestExecuteManifestMockSQLFilesInTxCommitsAllSuccess verifies the happy path
// where every mock-data SQL file commits its data and writes one
// sys_plugin_migration row per file.
func TestExecuteManifestMockSQLFilesInTxCommitsAllSuccess(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID  = "plugin-mock-data-commit"
		tableName = "plugin_mock_data_commit_log"
	)
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Mock Data Commit Plugin",
		"v0.1.0",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-commit.sql",
				Content: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-commit.sql",
				Content: fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			},
		},
	)
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}
	manifest.RuntimeArtifact.MockSQLAssets = []*catalog.ArtifactSQLAsset{
		{
			Key:     "001-plugin-mock-data-commit-mock.sql",
			Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('mock-row-1');", tableName),
		},
		{
			Key:     "002-plugin-mock-data-commit-mock.sql",
			Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('mock-row-2');", tableName),
		},
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	dropMockTestTable(t, ctx, tableName)
	t.Cleanup(func() {
		dropMockTestTable(t, ctx, tableName)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}
	if err = services.Lifecycle.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected install SQL to succeed, got error: %v", err)
	}

	var result lifecycle.MockSQLExecutionResult
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result = services.Lifecycle.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		return result.Err
	})
	if txErr != nil {
		t.Fatalf("expected mock SQL to commit, got error: %v", txErr)
	}
	if len(result.ExecutedFiles) != 2 {
		t.Fatalf("expected 2 executed files, got %d: %#v", len(result.ExecutedFiles), result.ExecutedFiles)
	}
	assertMockTableRowCount(t, ctx, tableName, 2)

	rows, err := g.DB().GetAll(ctx, "SELECT migration_key, status FROM sys_plugin_migration WHERE plugin_id = ? AND phase = ? ORDER BY execution_order;", pluginID, catalog.MigrationDirectionMock.String())
	if err != nil {
		t.Fatalf("expected migration ledger query to succeed, got error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 mock migration rows, got %d: %#v", len(rows), rows)
	}
	if rows[0]["status"].String() != catalog.MigrationExecutionStatusSucceeded.String() {
		t.Fatalf("expected mock migration #1 to be succeeded, got: %s", rows[0]["status"].String())
	}
}

// TestExecuteManifestMockSQLFilesInTxRollsBackOnFailure verifies that any mock
// SQL failure rolls back both the executed mock data rows and their
// corresponding sys_plugin_migration ledger entries together inside one
// transaction, leaving the install SQL phase results intact.
func TestExecuteManifestMockSQLFilesInTxRollsBackOnFailure(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID  = "plugin-mock-data-rollback"
		tableName = "plugin_mock_data_rollback_log"
	)
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Mock Data Rollback Plugin",
		"v0.1.0",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-rollback.sql",
				Content: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-rollback.sql",
				Content: fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			},
		},
	)
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}
	manifest.RuntimeArtifact.MockSQLAssets = []*catalog.ArtifactSQLAsset{
		{
			Key:     "001-plugin-mock-data-rollback-mock.sql",
			Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('mock-row-1');", tableName),
		},
		{
			Key:     "002-plugin-mock-data-rollback-broken-mock.sql",
			Content: "INSERT INTO this_table_does_not_exist (id) VALUES (42);",
		},
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	dropMockTestTable(t, ctx, tableName)
	t.Cleanup(func() {
		dropMockTestTable(t, ctx, tableName)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}
	if err = services.Lifecycle.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected install SQL to succeed, got error: %v", err)
	}

	var result lifecycle.MockSQLExecutionResult
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result = services.Lifecycle.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		return result.Err
	})
	if txErr == nil {
		t.Fatalf("expected mock SQL to fail, got nil error")
	}
	if result.FailedFile != "002-plugin-mock-data-rollback-broken-mock.sql" {
		t.Fatalf("expected failed file to be the second mock asset, got: %s", result.FailedFile)
	}
	if len(result.ExecutedFiles) != 1 || result.ExecutedFiles[0] != "001-plugin-mock-data-rollback-mock.sql" {
		t.Fatalf("expected first mock to have executed before rollback, got: %#v", result.ExecutedFiles)
	}

	// Mock data rows should be gone (rollback discarded the first INSERT).
	assertMockTableRowCount(t, ctx, tableName, 0)

	// Mock migration ledger should be empty (rolled back together).
	rows, err := g.DB().GetAll(ctx, "SELECT migration_key FROM sys_plugin_migration WHERE plugin_id = ? AND phase = ?;", pluginID, catalog.MigrationDirectionMock.String())
	if err != nil {
		t.Fatalf("expected migration ledger query to succeed, got error: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected zero mock migration rows after rollback, got %d: %#v", len(rows), rows)
	}

	// Install ledger should remain intact (mock rollback does not touch install rows).
	installRows, err := g.DB().GetAll(ctx, "SELECT migration_key FROM sys_plugin_migration WHERE plugin_id = ? AND phase = ?;", pluginID, catalog.MigrationDirectionInstall.String())
	if err != nil {
		t.Fatalf("expected install ledger query to succeed, got error: %v", err)
	}
	if len(installRows) == 0 {
		t.Fatalf("expected install ledger to remain after mock rollback, got 0 rows")
	}
}

// TestExecuteManifestMockSQLFilesInTxNoMockReturnsZeroValue verifies that
// running the mock entry point against a manifest without any mock SQL is a
// no-op: zero ledger writes, zero executed files, no error.
func TestExecuteManifestMockSQLFilesInTxNoMockReturnsZeroValue(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-mock-data-empty"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Mock Data Empty Plugin",
		"v0.1.0",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-empty.sql",
				Content: "SELECT 1;",
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-mock-data-empty.sql",
				Content: "SELECT 1;",
			},
		},
	)
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic storage artifact to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}

	var result lifecycle.MockSQLExecutionResult
	txErr := dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		result = services.Lifecycle.ExecuteManifestMockSQLFilesInTx(txCtx, manifest)
		return result.Err
	})
	if txErr != nil {
		t.Fatalf("expected empty mock to succeed, got error: %v", txErr)
	}
	if result.Err != nil {
		t.Fatalf("expected nil result.Err, got: %v", result.Err)
	}
	if len(result.ExecutedFiles) != 0 {
		t.Fatalf("expected zero executed files, got: %#v", result.ExecutedFiles)
	}
	if result.FailedFile != "" {
		t.Fatalf("expected empty FailedFile, got: %s", result.FailedFile)
	}
}

// TestMockDataLoadErrorUnwrapsCause verifies the typed *MockDataLoadError
// surfaces the underlying database error to errors.Is/errors.As callers.
func TestMockDataLoadErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("syntax error near 'this_table_does_not_exist'")
	loadErr := &lifecycle.MockDataLoadError{
		PluginID:        "plugin-x",
		FailedFile:      "001-plugin-x-mock.sql",
		RolledBackFiles: []string{"001-plugin-x-mock.sql"},
		Cause:           cause,
	}
	if !errors.Is(loadErr, cause) {
		t.Fatalf("expected errors.Is to recognize Cause via Unwrap")
	}
	var typed *lifecycle.MockDataLoadError
	if !errors.As(loadErr, &typed) {
		t.Fatalf("expected errors.As to recover the typed mock load error")
	}
	if typed.FailedFile != "001-plugin-x-mock.sql" {
		t.Fatalf("unexpected typed.FailedFile: %s", typed.FailedFile)
	}
}

// dropMockTestTable removes the temporary table used by mock SQL tests.
func dropMockTestTable(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	if _, err := g.DB().Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)); err != nil {
		t.Fatalf("expected mock test table cleanup to succeed, got error: %v", err)
	}
}

// assertMockTableRowCount verifies the row count produced by mock SQL.
func assertMockTableRowCount(t *testing.T, ctx context.Context, tableName string, expected int) {
	t.Helper()

	value, err := g.DB().GetValue(ctx, fmt.Sprintf("SELECT COUNT(1) FROM %s;", tableName))
	if err != nil {
		t.Fatalf("expected row count query to succeed, got error: %v", err)
	}
	if value.Int() != expected {
		t.Fatalf("expected table %s to contain %d rows, got %d", tableName, expected, value.Int())
	}
}
