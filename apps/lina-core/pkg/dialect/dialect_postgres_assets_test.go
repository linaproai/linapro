// This file verifies PostgreSQL SQL assets can be executed directly when an
// explicit PostgreSQL smoke database is available.

package dialect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/testsupport"
)

// TestPostgreSQLProjectSQLAssetsSmoke executes all PostgreSQL-source SQL
// assets against a real PostgreSQL database when explicitly enabled.
func TestPostgreSQLProjectSQLAssetsSmoke(t *testing.T) {
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL SQL asset smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root failed: %v", err)
	}
	if !testsupport.OfficialPluginsWorkspaceReady(repoRoot) {
		t.Skip("official plugin workspace is not initialized")
	}

	installAssets := collectProjectSQLAssets(t, sqlAssetGroupInstall)
	mockAssets := collectProjectSQLAssets(t, sqlAssetGroupMock)
	uninstallAssets := collectProjectSQLAssets(t, sqlAssetGroupUninstall)
	if len(installAssets) == 0 || len(mockAssets) == 0 || len(uninstallAssets) == 0 {
		t.Fatal("expected install, mock, and uninstall SQL assets")
	}

	dbLink := postgresSmokeDatabaseLink(t, baseLink)
	dbDialect, err := From(dbLink)
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, dbLink, true); err != nil {
		t.Fatalf("prepare PostgreSQL smoke database failed: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if cleanupErr := dropPostgreSQLSmokeDatabase(cleanupCtx, dbLink); cleanupErr != nil {
			t.Errorf("cleanup PostgreSQL smoke database failed: %v", cleanupErr)
		}
	})

	db, err := gdb.New(gdb.ConfigNode{Link: dbLink})
	if err != nil {
		t.Fatalf("open PostgreSQL smoke database failed: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL smoke database failed: %v", closeErr)
		}
	})

	executePostgreSQLSQLAssets(t, ctx, db, installAssets)
	executePostgreSQLSQLAssets(t, ctx, db, mockAssets)
	executePostgreSQLSQLAssets(t, ctx, db, uninstallAssets)
}

// executePostgreSQLSQLAssets executes assets in order on one PostgreSQL DB.
func executePostgreSQLSQLAssets(t *testing.T, ctx context.Context, db gdb.DB, assets []sqlTestAsset) {
	t.Helper()
	for _, asset := range assets {
		asset := asset
		t.Run(asset.sourceName, func(t *testing.T) {
			for index, statement := range SplitSQLStatements(asset.content) {
				if _, err := db.Exec(ctx, statement); err != nil {
					t.Fatalf("execute PostgreSQL statement %d failed: %v\nSQL:\n%s", index+1, err, statement)
				}
			}
		})
	}
}

// sqlAssetGroup identifies the lifecycle slice of project SQL assets.
type sqlAssetGroup string

const (
	sqlAssetGroupInstall   sqlAssetGroup = "install"
	sqlAssetGroupMock      sqlAssetGroup = "mock"
	sqlAssetGroupUninstall sqlAssetGroup = "uninstall"
)

// sqlAssetPatterns returns the project SQL glob patterns for one asset group.
func sqlAssetPatterns(root string, group sqlAssetGroup) []string {
	switch group {
	case sqlAssetGroupInstall:
		return []string{
			filepath.Join(root, "apps/lina-core/manifest/sql/*.sql"),
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/*.sql"),
		}
	case sqlAssetGroupMock:
		return []string{
			filepath.Join(root, "apps/lina-core/manifest/sql/mock-data/*.sql"),
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/mock-data/*.sql"),
		}
	case sqlAssetGroupUninstall:
		return []string{
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/uninstall/*.sql"),
		}
	default:
		return nil
	}
}

// sqlTestAsset stores one SQL asset fixture.
type sqlTestAsset struct {
	sourceName string
	content    string
}

// collectProjectSQLAssets finds host and plugin SQL assets relative to the repo root.
func collectProjectSQLAssets(t *testing.T, group sqlAssetGroup) []sqlTestAsset {
	t.Helper()
	root := repositoryRootForSQLAudit(t)
	var assets []sqlTestAsset
	for _, pattern := range sqlAssetPatterns(root, group) {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob SQL assets failed: %v", err)
		}
		for _, file := range files {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatalf("read SQL asset %s failed: %v", file, readErr)
			}
			rel, relErr := filepath.Rel(root, file)
			if relErr != nil {
				t.Fatalf("rel SQL asset %s failed: %v", file, relErr)
			}
			assets = append(assets, sqlTestAsset{
				sourceName: filepath.ToSlash(rel),
				content:    string(content),
			})
		}
	}
	return assets
}

// postgresSmokeDatabaseLink returns a unique database link for one smoke test.
func postgresSmokeDatabaseLink(t *testing.T, baseLink string) string {
	t.Helper()

	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		t.Fatalf("parse PostgreSQL smoke base link failed: %v", err)
	}
	if db.GetConfig() == nil {
		t.Fatal("PostgreSQL smoke base link configuration is empty")
	}
	config := db.GetConfig()
	if closeErr := db.Close(context.Background()); closeErr != nil {
		t.Fatalf("close PostgreSQL smoke base link parser failed: %v", closeErr)
	}

	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sql_smoke_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		extra,
	)
}

// dropPostgreSQLSmokeDatabase removes the temporary database created by the
// PostgreSQL asset smoke test.
func dropPostgreSQLSmokeDatabase(ctx context.Context, targetLink string) (err error) {
	targetDB, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	targetConfig := targetDB.GetConfig()
	if targetConfig == nil {
		if closeErr := targetDB.Close(ctx); closeErr != nil {
			return closeErr
		}
		return nil
	}
	targetName := strings.TrimSpace(targetConfig.Name)
	if closeErr := targetDB.Close(ctx); closeErr != nil {
		return closeErr
	}
	if targetName == "" {
		return nil
	}

	systemLink := postgresSmokeSystemDatabaseLink(*targetConfig)
	systemDB, err := gdb.New(gdb.ConfigNode{Link: systemLink})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = systemDB.Exec(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		targetName,
	); err != nil {
		return err
	}
	quotedName := `"` + strings.ReplaceAll(targetName, `"`, `""`) + `"`
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// postgresSmokeSystemDatabaseLink returns a link to PostgreSQL's maintenance
// database using the same host, credentials, and extra parameters.
func postgresSmokeSystemDatabaseLink(config gdb.ConfigNode) string {
	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		extra,
	)
}
