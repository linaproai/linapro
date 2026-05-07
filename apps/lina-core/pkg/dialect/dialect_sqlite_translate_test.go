// This file tests SQLite DDL translation against project SQL assets.

package dialect

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/glog"

	"lina-core/pkg/logger"
)

// TestSQLiteTranslateDDLExecutesProjectInstallSQLAssets verifies current host
// and plugin install SQL assets can be translated and executed by SQLite in the
// same order used by init plus plugin installation.
func TestSQLiteTranslateDDLExecutesProjectInstallSQLAssets(t *testing.T) {
	ctx := context.Background()
	assets := collectProjectSQLAssets(t, sqlAssetGroupInstall)
	if len(assets) == 0 {
		t.Fatal("expected SQL assets")
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, ctx, db)
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

// TestSQLiteTranslateDDLExecutesProjectMockSQLAssets verifies current mock SQL
// assets can be translated and executed after host and plugin schema assets.
func TestSQLiteTranslateDDLExecutesProjectMockSQLAssets(t *testing.T) {
	ctx := context.Background()
	db := newSQLiteTestDB(t)
	defer closeDB(t, ctx, db)
	executeSQLiteSQLAssets(t, ctx, db, collectProjectSQLAssets(t, sqlAssetGroupInstall))

	mockAssets := collectProjectSQLAssets(t, sqlAssetGroupMock)
	if len(mockAssets) == 0 {
		t.Fatal("expected mock SQL assets")
	}
	executeSQLiteSQLAssets(t, ctx, db, mockAssets)
}

// TestSQLiteTranslateDDLExecutesProjectUninstallSQLAssets verifies uninstall
// assets can be translated and executed after install assets.
func TestSQLiteTranslateDDLExecutesProjectUninstallSQLAssets(t *testing.T) {
	ctx := context.Background()
	db := newSQLiteTestDB(t)
	defer closeDB(t, ctx, db)
	executeSQLiteSQLAssets(t, ctx, db, collectProjectSQLAssets(t, sqlAssetGroupInstall))

	uninstallAssets := collectProjectSQLAssets(t, sqlAssetGroupUninstall)
	if len(uninstallAssets) == 0 {
		t.Fatal("expected uninstall SQL assets")
	}
	executeSQLiteSQLAssets(t, ctx, db, uninstallAssets)
}

// TestSQLiteTranslateDDLCurrentDMLFunctions verifies MySQL functions that
// appear in current DML assets are rewritten to SQLite-compatible syntax.
func TestSQLiteTranslateDDLCurrentDMLFunctions(t *testing.T) {
	input := "INSERT INTO demo (path) SELECT CONCAT('0,', parent.id) FROM parent;"
	translated, err := sqliteDialectForTest(t).TranslateDDL(context.Background(), "functions.sql", input)
	if err != nil {
		t.Fatalf("translate function fixture failed: %v", err)
	}
	if !strings.Contains(translated, "('0,' || parent.id)") {
		t.Fatalf("expected CONCAT to be rewritten, got:\n%s", translated)
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, context.Background(), db)
	statements := []string{
		"CREATE TABLE parent (id INTEGER PRIMARY KEY);",
		"CREATE TABLE demo (path TEXT);",
		"INSERT INTO parent (id) VALUES (42);",
	}
	for _, statement := range statements {
		if _, err = db.Exec(context.Background(), statement); err != nil {
			t.Fatalf("execute setup SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	for _, statement := range SplitSQLStatements(translated) {
		if _, err = db.Exec(context.Background(), statement); err != nil {
			t.Fatalf("execute translated function SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	value, err := db.GetValue(context.Background(), "SELECT path FROM demo")
	if err != nil {
		t.Fatalf("read translated function result failed: %v", err)
	}
	if value.String() != "0,42" {
		t.Fatalf("expected concatenated value 0,42, got %s", value.String())
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

// TestSQLiteTranslateDDLRealSyntaxShapes verifies known current SQL shapes that
// are easy to miss when only examples are covered.
func TestSQLiteTranslateDDLRealSyntaxShapes(t *testing.T) {
	input := `
CREATE TABLE IF NOT EXISTS ` + "`" + `shape` + "`" + ` (
    ` + "`" + `id` + "`" + ` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Config parameter ID',
    code VARCHAR(64) NOT NULL COMMENT 'Code',
    PRIMARY KEY (` + "`" + `id` + "`" + `),
    UNIQUE INDEX uk_shape_code (code),
    UNIQUE KEY uk_shape_expr ((NULLIF(code, '')))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Shape table';
`
	translated, err := sqliteDialectForTest(t).TranslateDDL(context.Background(), "shape.sql", input)
	if err != nil {
		t.Fatalf("translate shape fixture failed: %v", err)
	}
	required := []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_shape_code ON shape (code);",
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_shape_expr ON shape ((NULLIF(code, '')));",
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, context.Background(), db)
	for _, statement := range SplitSQLStatements(translated) {
		if _, err = db.Exec(context.Background(), statement); err != nil {
			t.Fatalf("execute translated shape SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
}

// TestSQLiteTranslateDDLPreservesSingleColumnTablePrimaryKey verifies table
// primary keys are not dropped when they do not describe an auto-increment ID.
func TestSQLiteTranslateDDLPreservesSingleColumnTablePrimaryKey(t *testing.T) {
	input := `
CREATE TABLE IF NOT EXISTS sys_online_session (
    token_id VARCHAR(64) NOT NULL COMMENT 'Session token ID',
    user_id INT NOT NULL DEFAULT 0 COMMENT 'User ID',
    PRIMARY KEY (token_id)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='Online session table';
`
	translated, err := sqliteDialectForTest(t).TranslateDDL(context.Background(), "online-session.sql", input)
	if err != nil {
		t.Fatalf("translate online session fixture failed: %v", err)
	}
	if !strings.Contains(translated, "PRIMARY KEY (token_id)") {
		t.Fatalf("expected translated SQL to preserve token_id primary key, got:\n%s", translated)
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, context.Background(), db)
	for _, statement := range SplitSQLStatements(translated) {
		if _, err = db.Exec(context.Background(), statement); err != nil {
			t.Fatalf("execute translated online session SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	for _, statement := range []string{
		"INSERT INTO sys_online_session (token_id, user_id) VALUES ('token-1', 1)",
		"INSERT INTO sys_online_session (token_id, user_id) VALUES ('token-1', 2)",
	} {
		_, err = db.Exec(context.Background(), statement)
	}
	if err == nil {
		t.Fatal("expected duplicate token_id insert to fail because primary key was preserved")
	}
}

// TestSQLiteTranslateDDLIgnoresSQLKeywordCase verifies keyword and identifier
// matching does not depend on one canonical SQL keyword casing.
func TestSQLiteTranslateDDLIgnoresSQLKeywordCase(t *testing.T) {
	input := `
create table if not exists ` + "`" + `mixed_case` + "`" + ` (
    ` + "`" + `id` + "`" + ` bigint unsigned not null auto_increment comment 'Identifier',
    code varchar(64) not null comment 'Code',
    primary key (` + "`" + `ID` + "`" + `),
    unique key uk_mixed_case_code (code)
) engine=InnoDB default charset=utf8mb4 comment='Mixed case table';
insert ignore into mixed_case (code) select concat('A', 'B') from dual;
`
	translated, err := sqliteDialectForTest(t).TranslateDDL(context.Background(), "mixed-case.sql", input)
	if err != nil {
		t.Fatalf("translate mixed-case fixture failed: %v", err)
	}
	required := []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_mixed_case_code ON mixed_case (code);",
		"INSERT OR IGNORE INTO mixed_case (code) select ('A' || 'B');",
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}

	db := newSQLiteTestDB(t)
	defer closeDB(t, context.Background(), db)
	for _, statement := range SplitSQLStatements(translated) {
		if _, err = db.Exec(context.Background(), statement); err != nil {
			t.Fatalf("execute translated mixed-case SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
	value, err := db.GetValue(context.Background(), "SELECT code FROM mixed_case")
	if err != nil {
		t.Fatalf("read mixed-case translated result failed: %v", err)
	}
	if value.String() != "AB" {
		t.Fatalf("expected concatenated value AB, got %s", value.String())
	}
}

// TestSQLiteTranslateDDLUnsupportedSyntax verifies unsupported MySQL syntax
// returns source-aware errors.
func TestSQLiteTranslateDDLUnsupportedSyntax(t *testing.T) {
	tests := []struct {
		name       string
		sourceName string
		input      string
		keyword    string
	}{
		{
			name:       "fulltext",
			sourceName: "bad/fulltext.sql",
			input:      "CREATE TABLE demo (id INT PRIMARY KEY AUTO_INCREMENT, FULLTEXT INDEX ft_name (name));",
			keyword:    "FULLTEXT",
		},
		{
			name:       "generated",
			sourceName: "bad/generated.sql",
			input:      "CREATE TABLE demo (id INT PRIMARY KEY AUTO_INCREMENT, code VARCHAR(64), code_upper VARCHAR(64) GENERATED ALWAYS AS (UPPER(code)));",
			keyword:    "GENERATED ALWAYS AS",
		},
		{
			name:       "on duplicate",
			sourceName: "bad/on-duplicate.sql",
			input:      "INSERT INTO demo (code) VALUES ('x') ON DUPLICATE KEY UPDATE code = VALUES(code);",
			keyword:    "ON DUPLICATE KEY UPDATE",
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

// TestMySQLOnStartupDoesNotWarn verifies MySQL startup hooks do not produce
// SQLite warnings or override the cluster runtime.
func TestMySQLOnStartupDoesNotWarn(t *testing.T) {
	runtime := &fakeRuntimeConfig{}
	var warnings []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		warnings = append(warnings, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	dbDialect, err := From("mysql:root:pass@tcp(127.0.0.1:3306)/linapro")
	if err != nil {
		t.Fatalf("resolve MySQL dialect failed: %v", err)
	}
	if err := dbDialect.OnStartup(context.Background(), runtime); err != nil {
		t.Fatalf("run MySQL startup hook failed: %v", err)
	}
	if runtime.called {
		t.Fatal("expected MySQL startup hook not to override cluster mode")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no MySQL startup warnings, got %#v", warnings)
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
