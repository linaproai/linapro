// This file covers runtime migration replay behaviors owned by lifecycle.

package lifecycle_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestExecuteManifestSQLFilesReplaysInstallSQL(t *testing.T) {
	services := testutil.NewServices()
	service := services.Lifecycle
	ctx := context.Background()

	pluginID := "plugin-dynamic-reinstall"
	tableName := "plugin_runtime_reinstall_log"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Reinstall Plugin",
		"v0.9.1",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-reinstall-create.sql",
				Content: fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY AUTO_INCREMENT, marker VARCHAR(32) NOT NULL);", tableName),
			},
			{
				Key:     "002-plugin-dynamic-reinstall-seed.sql",
				Content: fmt.Sprintf("INSERT INTO %s (marker) VALUES ('install-ran');", tableName),
			},
		},
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-reinstall.sql",
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

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected first install to succeed, got error: %v", err)
	}
	assertTestTableRowCount(t, ctx, tableName, 1)

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionUninstall); err != nil {
		t.Fatalf("expected uninstall to succeed, got error: %v", err)
	}
	assertTestTableMissing(t, ctx, tableName)

	if err = service.ExecuteManifestSQLFiles(ctx, manifest, catalog.MigrationDirectionInstall); err != nil {
		t.Fatalf("expected reinstall to succeed, got error: %v", err)
	}
	assertTestTableRowCount(t, ctx, tableName, 1)
}

func dropTestTableIfExists(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	if _, err := g.DB().Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)); err != nil {
		t.Fatalf("expected test table cleanup to succeed, got error: %v", err)
	}
}

func assertTestTableMissing(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()

	all, err := g.DB().GetAll(ctx, fmt.Sprintf("SHOW TABLES LIKE '%s';", tableName))
	if err != nil {
		t.Fatalf("expected table existence query to succeed, got error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected table %s to be dropped, got rows: %#v", tableName, all)
	}
}

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
