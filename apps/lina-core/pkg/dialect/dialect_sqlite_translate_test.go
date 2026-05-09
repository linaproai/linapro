// This file tests SQLite DDL translation against PostgreSQL-source SQL.

package dialect

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSQLiteTranslateDDLExecutesPostgreSQLFixture verifies the public SQLite
// dialect translates representative PG source SQL into executable SQLite SQL.
func TestSQLiteTranslateDDLExecutesPostgreSQLFixture(t *testing.T) {
	ctx := context.Background()
	input := `
CREATE TABLE IF NOT EXISTS sys_user (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    profile BYTEA,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "type" CHAR(1) NOT NULL DEFAULT 'U'
);
COMMENT ON TABLE sys_user IS 'User table';
COMMENT ON COLUMN sys_user.username IS 'Username';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user (username);
CREATE INDEX IF NOT EXISTS idx_sys_user_created_at ON sys_user (created_at);
INSERT INTO sys_user (username, nickname, "type") VALUES ('admin', 'Administrator', 'U') ON CONFLICT DO NOTHING;
INSERT INTO sys_user (username, nickname, "type") VALUES ('Admin', 'Duplicate case', 'U') ON CONFLICT DO NOTHING;
`

	translated, err := sqliteDialectForTest(t).TranslateDDL(ctx, "fixture.sql", input)
	if err != nil {
		t.Fatalf("translate PG fixture failed: %v", err)
	}
	for _, needle := range []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"username TEXT NOT NULL",
		"profile BLOB",
		`"type" TEXT NOT NULL DEFAULT 'U'`,
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user (username);",
		"CREATE INDEX IF NOT EXISTS idx_sys_user_created_at ON sys_user (created_at);",
		"ON CONFLICT DO NOTHING",
	} {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, ctx, db)
	for index, statement := range SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute translated statement %d failed: %v\nSQL:\n%s", index+1, err, statement)
		}
	}

	count, err := db.GetValue(ctx, "SELECT COUNT(*) FROM sys_user")
	if err != nil {
		t.Fatalf("read translated fixture row count failed: %v", err)
	}
	if count.String() != "2" {
		t.Fatalf("expected deterministic unique index to keep case-distinct rows, got %s", count.String())
	}
	indexCount, err := db.GetValue(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name IN ('uk_sys_user_username', 'idx_sys_user_created_at')",
	)
	if err != nil {
		t.Fatalf("read translated fixture index count failed: %v", err)
	}
	if indexCount.String() != "2" {
		t.Fatalf("expected translated fixture to create both indexes, got %s", indexCount.String())
	}
}

// TestSQLiteTranslateDDLProjectSQLAssetsSmoke executes all project SQL assets
// once the parallel SQL rewrite has fully converted them to PostgreSQL source.
func TestSQLiteTranslateDDLProjectSQLAssetsSmoke(t *testing.T) {
	ctx := context.Background()
	installAssets := collectProjectSQLAssets(t, sqlAssetGroupInstall)
	mockAssets := collectProjectSQLAssets(t, sqlAssetGroupMock)
	uninstallAssets := collectProjectSQLAssets(t, sqlAssetGroupUninstall)
	if len(installAssets) == 0 {
		t.Fatal("expected install SQL assets")
	}
	if len(mockAssets) == 0 {
		t.Fatal("expected mock SQL assets")
	}
	if len(uninstallAssets) == 0 {
		t.Fatal("expected uninstall SQL assets")
	}

	allAssets := append(append([]sqlTestAsset{}, installAssets...), mockAssets...)
	allAssets = append(allAssets, uninstallAssets...)
	skipIfProjectSQLStillUsesLegacyMySQL(t, allAssets)

	db := newSQLiteTestDB(t)
	defer closeDB(t, ctx, db)
	executeSQLiteSQLAssets(t, ctx, db, installAssets)
	executeSQLiteSQLAssets(t, ctx, db, mockAssets)
	executeSQLiteSQLAssets(t, ctx, db, uninstallAssets)
}

// TestPostgreSQLProjectSQLAssetsSmoke executes all PostgreSQL-source SQL
// assets against a real PostgreSQL database when explicitly enabled.
func TestPostgreSQLProjectSQLAssetsSmoke(t *testing.T) {
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL SQL asset smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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

// TestSQLiteTranslateDDLUnsupportedPostgreSQLSyntax verifies unsupported PG
// features return source-aware diagnostics through the public dialect.
func TestSQLiteTranslateDDLUnsupportedPostgreSQLSyntax(t *testing.T) {
	tests := []struct {
		name       string
		sourceName string
		input      string
		keyword    string
	}{
		{
			name:       "jsonb",
			sourceName: "bad/jsonb.sql",
			input:      "CREATE TABLE demo (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, payload JSONB);",
			keyword:    "JSONB",
		},
		{
			name:       "trigger",
			sourceName: "bad/trigger.sql",
			input:      "CREATE TRIGGER trg_demo BEFORE UPDATE ON demo EXECUTE FUNCTION touch_demo();",
			keyword:    "CREATE TRIGGER",
		},
		{
			name:       "legacy mysql",
			sourceName: "bad/mysql.sql",
			input:      "CREATE TABLE `demo` (`id` INT);",
			keyword:    "backtick identifier",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := sqliteDialectForTest(t).TranslateDDL(context.Background(), test.sourceName, test.input)
			if err == nil {
				t.Fatal("expected unsupported syntax error")
			}
			message := err.Error()
			for _, needle := range []string{test.sourceName, "line", test.keyword} {
				if !strings.Contains(message, needle) {
					t.Fatalf("expected error to contain %q, got %q", needle, message)
				}
			}
		})
	}
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

// executeSQLiteSQLAssets translates and executes assets in order on one SQLite DB.
func executeSQLiteSQLAssets(t *testing.T, ctx context.Context, db gdb.DB, assets []sqlTestAsset) {
	t.Helper()
	for _, asset := range assets {
		asset := asset
		t.Run(asset.sourceName, func(t *testing.T) {
			translated, err := sqliteDialectForTest(t).TranslateDDL(ctx, asset.sourceName, asset.content)
			if err != nil {
				t.Fatalf("translate SQLite DDL failed: %v", err)
			}
			for index, statement := range SplitSQLStatements(translated) {
				if _, err = db.Exec(ctx, statement); err != nil {
					t.Fatalf("execute translated statement %d failed: %v\nSQL:\n%s", index+1, err, statement)
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

// sqliteDialectForTest resolves the public SQLite dialect contract for tests.
func sqliteDialectForTest(t *testing.T) Dialect {
	t.Helper()
	dbDialect, err := From("sqlite::@file(./temp/sqlite/linapro.db)")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	return dbDialect
}

// sqlTestAsset stores one SQL asset fixture.
type sqlTestAsset struct {
	sourceName string
	content    string
}

// collectProjectSQLAssets finds host and plugin SQL assets relative to the repo root.
func collectProjectSQLAssets(t *testing.T, group sqlAssetGroup) []sqlTestAsset {
	t.Helper()
	root := findRepoRoot(t)
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

// skipIfProjectSQLStillUsesLegacyMySQL skips the project smoke test until the
// SQL rewrite workstream has converted every project asset to PG source syntax.
func skipIfProjectSQLStillUsesLegacyMySQL(t *testing.T, assets []sqlTestAsset) {
	t.Helper()
	var findings []string
	for _, asset := range assets {
		for _, statement := range SplitSQLStatements(asset.content) {
			if reason := legacyMySQLSourceReason(statement); reason != "" {
				findings = append(findings, asset.sourceName+": "+reason)
				break
			}
		}
		if len(findings) >= 6 {
			break
		}
	}
	if len(findings) > 0 {
		t.Skipf("project SQL assets are not fully PostgreSQL-source yet: %s", strings.Join(findings, "; "))
	}
}

// legacyMySQLSourceReason returns a diagnostic label for source syntax that the
// PG-to-SQLite translator must no longer accept.
func legacyMySQLSourceReason(statement string) string {
	upper := strings.ToUpper(statement)
	switch {
	case strings.Contains(statement, "`"):
		return "backtick identifier"
	case strings.Contains(upper, "AUTO_INCREMENT"):
		return "AUTO_INCREMENT"
	case strings.Contains(upper, "INSERT IGNORE"):
		return "INSERT IGNORE"
	case strings.Contains(upper, "ON DUPLICATE KEY UPDATE"):
		return "ON DUPLICATE KEY UPDATE"
	case strings.Contains(upper, "ENGINE="):
		return "ENGINE="
	case strings.Contains(upper, "DEFAULT CHARSET"):
		return "DEFAULT CHARSET"
	case strings.Contains(upper, "ON UPDATE CURRENT_TIMESTAMP"):
		return "ON UPDATE CURRENT_TIMESTAMP"
	case strings.Contains(upper, "UNSIGNED"):
		return "UNSIGNED"
	case strings.Contains(upper, "TINYINT"):
		return "TINYINT"
	case strings.Contains(upper, "DATETIME"):
		return "DATETIME"
	case strings.Contains(upper, "LONGTEXT"):
		return "LONGTEXT"
	case strings.Contains(upper, "MEDIUMTEXT"):
		return "MEDIUMTEXT"
	case strings.Contains(upper, "VARBINARY"):
		return "VARBINARY"
	case strings.Contains(upper, "CREATE DATABASE"):
		return "CREATE DATABASE"
	case strings.HasPrefix(strings.TrimSpace(upper), "USE "):
		return "USE"
	default:
		return ""
	}
}

// findRepoRoot walks up from the package directory until it finds go.work.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.work")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}

// newSQLiteTestDB creates one isolated in-memory SQLite DB.
func newSQLiteTestDB(t *testing.T) gdb.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "fixture.db")
	db, err := gdb.New(gdb.ConfigNode{Link: "sqlite::@file(" + dbPath + ")"})
	if err != nil {
		t.Fatalf("create SQLite test DB failed: %v", err)
	}
	return db
}

// closeDB closes a test database and fails the test on error.
func closeDB(t *testing.T, ctx context.Context, db gdb.DB) {
	t.Helper()
	if err := db.Close(ctx); err != nil {
		t.Fatalf("close test DB failed: %v", err)
	}
}
