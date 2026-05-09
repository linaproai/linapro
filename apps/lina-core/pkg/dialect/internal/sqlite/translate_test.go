// This file tests PostgreSQL-source SQL translation into SQLite syntax.

package sqlite

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/os/glog"

	"lina-core/pkg/logger"
)

// TestTranslateDDLPostgreSQLSubset verifies the supported PG source syntax is
// translated to SQLite-compatible SQL while preserving compatible statements.
func TestTranslateDDLPostgreSQLSubset(t *testing.T) {
	t.Parallel()

	input := `
CREATE TABLE IF NOT EXISTS sys_config (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    retries SMALLINT NOT NULL DEFAULT 0,
    username VARCHAR(64) NOT NULL,
    fixed_code CHAR(2) NOT NULL,
    description TEXT,
    payload BYTEA,
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    ratio NUMERIC(10, 2),
    score REAL,
    weighted DOUBLE PRECISION,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "key" VARCHAR(128) NOT NULL,
    "value" TEXT,
    UNIQUE ("key"),
    CHECK (tenant_id >= 0)
);
COMMENT ON TABLE sys_config IS 'Config table';
COMMENT ON COLUMN sys_config."key" IS 'Config key';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_config_key ON sys_config ("key");
INSERT INTO sys_config ("key", "value", tenant_id) VALUES ('Admin', 'x', 1) ON CONFLICT DO NOTHING;
`

	translated, err := TranslateDDL(context.Background(), "subset.sql", input)
	if err != nil {
		t.Fatalf("translate PostgreSQL subset failed: %v", err)
	}

	required := []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"tenant_id INTEGER NOT NULL DEFAULT 0",
		"retries INTEGER NOT NULL DEFAULT 0",
		"username TEXT NOT NULL",
		"fixed_code TEXT NOT NULL",
		"description TEXT",
		"payload BLOB",
		"amount NUMERIC NOT NULL DEFAULT 0",
		"ratio NUMERIC",
		"score REAL",
		"weighted REAL",
		"created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP",
		`"key" TEXT NOT NULL`,
		`"value" TEXT`,
		`UNIQUE ("key")`,
		"CHECK (tenant_id >= 0)",
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_config_key ON sys_config ("key");`,
		`INSERT INTO sys_config ("key", "value", tenant_id) VALUES ('Admin', 'x', 1) ON CONFLICT DO NOTHING;`,
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}

	forbidden := []string{
		"CREATE COLLATION",
		"COMMENT ON",
		"linapro_ci",
		"COLLATE NOCASE",
		"BIGINT",
		"SMALLINT",
		"VARCHAR",
		"CHAR(2)",
		"BYTEA",
		"TIMESTAMP NOT NULL",
		"DECIMAL",
		"DOUBLE PRECISION",
	}
	for _, needle := range forbidden {
		if strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL not to contain %q, got:\n%s", needle, translated)
		}
	}
}

// TestTranslateDDLRewritesNowFunction verifies legacy-compatible PG source that
// uses NOW() is normalized for SQLite execution.
func TestTranslateDDLRewritesNowFunction(t *testing.T) {
	t.Parallel()

	input := "INSERT INTO demo (created_at) VALUES (NOW()) ON CONFLICT DO NOTHING;"
	translated, err := TranslateDDL(context.Background(), "now.sql", input)
	if err != nil {
		t.Fatalf("translate NOW fixture failed: %v", err)
	}
	if !strings.Contains(translated, "VALUES (CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;") {
		t.Fatalf("expected NOW() to become CURRENT_TIMESTAMP, got:\n%s", translated)
	}
}

// TestTranslateDDLRewritesTimestampLiterals verifies PG timestamp literal
// syntax used by seed/mock data is normalized for SQLite execution.
func TestTranslateDDLRewritesTimestampLiterals(t *testing.T) {
	t.Parallel()

	input := `
INSERT INTO demo (read_at, sent_at)
SELECT NULL::TIMESTAMP, TIMESTAMP '2026-04-20 09:00:10'
ON CONFLICT DO NOTHING;`
	translated, err := TranslateDDL(context.Background(), "timestamp-literal.sql", input)
	if err != nil {
		t.Fatalf("translate timestamp literal fixture failed: %v", err)
	}
	required := []string{
		"SELECT NULL, '2026-04-20 09:00:10'",
		"ON CONFLICT DO NOTHING",
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}
	forbidden := []string{"::TIMESTAMP", "TIMESTAMP '"}
	for _, needle := range forbidden {
		if strings.Contains(strings.ToUpper(translated), needle) {
			t.Fatalf("expected translated SQL not to contain %q, got:\n%s", needle, translated)
		}
	}
}

// TestTranslateDDLRejectsCollationClause verifies project SQL does not rely on
// explicit collation clauses after switching to default deterministic behavior.
func TestTranslateDDLRejectsCollationClause(t *testing.T) {
	t.Parallel()

	inputs := []string{`
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, m.id
FROM sys_menu m
WHERE m.menu_key COLLATE "C" NOT LIKE 'plugin:%'
  AND note <> 'COLLATE "C" literal'
ON CONFLICT DO NOTHING;`, `
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, m.id
FROM sys_menu m
WHERE m.menu_key COLLATE C NOT LIKE 'plugin:%'
ON CONFLICT DO NOTHING;`}
	for index, input := range inputs {
		_, err := TranslateDDL(context.Background(), "explicit-collation.sql", input)
		if err == nil {
			t.Fatalf("expected explicit collation fixture %d to be rejected", index+1)
		}
		if !strings.Contains(err.Error(), "COLLATE") {
			t.Fatalf("expected COLLATE error, got: %v", err)
		}
	}
	translated, err := TranslateDDL(
		context.Background(),
		"collation-literal.sql",
		`INSERT INTO demo (note) VALUES ('COLLATE "C" literal') ON CONFLICT DO NOTHING;`,
	)
	if err != nil {
		t.Fatalf("translate collation literal fixture failed: %v", err)
	}
	if !strings.Contains(translated, `'COLLATE "C" literal'`) {
		t.Fatalf("expected string literal to remain unchanged, got:\n%s", translated)
	}
}

// TestTranslateDDLDropsIdentityTablePrimaryKey verifies a table-level primary
// key is removed when the same identity column already owns SQLite's inline key.
func TestTranslateDDLDropsIdentityTablePrimaryKey(t *testing.T) {
	t.Parallel()

	input := `
CREATE TABLE IF NOT EXISTS demo (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(64),
    CONSTRAINT pk_demo PRIMARY KEY (id)
);`
	translated, err := TranslateDDL(context.Background(), "identity-primary.sql", input)
	if err != nil {
		t.Fatalf("translate identity primary key fixture failed: %v", err)
	}
	if !strings.Contains(translated, "id INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Fatalf("expected identity column to become SQLite autoincrement, got:\n%s", translated)
	}
	if strings.Contains(translated, "PRIMARY KEY (id)") {
		t.Fatalf("expected duplicate table primary key to be dropped, got:\n%s", translated)
	}
}

// TestTranslateDDLPreservesNonIdentityTablePrimaryKey verifies natural primary
// keys remain table constraints.
func TestTranslateDDLPreservesNonIdentityTablePrimaryKey(t *testing.T) {
	t.Parallel()

	input := `
CREATE TABLE IF NOT EXISTS sys_online_session (
    token_id VARCHAR(64) NOT NULL,
    user_id INT NOT NULL DEFAULT 0,
    PRIMARY KEY (token_id)
);`
	translated, err := TranslateDDL(context.Background(), "natural-primary.sql", input)
	if err != nil {
		t.Fatalf("translate natural primary key fixture failed: %v", err)
	}
	if !strings.Contains(translated, "token_id TEXT NOT NULL") {
		t.Fatalf("expected token_id type conversion, got:\n%s", translated)
	}
	if !strings.Contains(translated, "PRIMARY KEY (token_id)") {
		t.Fatalf("expected natural primary key to be preserved, got:\n%s", translated)
	}
}

// TestTranslateDDLDropsCommentsWithDebugLogPreservingIndexes verifies SQLite
// translation skips PG comment metadata with diagnostics while leaving index
// maintenance statements on the compatible-SQL path.
func TestTranslateDDLDropsCommentsWithDebugLogPreservingIndexes(t *testing.T) {
	ctx := context.Background()
	input := `
CREATE TABLE IF NOT EXISTS sys_config (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "key" VARCHAR(64) NOT NULL,
    "value" TEXT
);
COMMENT ON TABLE sys_config IS 'Config table';
COMMENT ON COLUMN sys_config."key" IS 'Config key';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_config_key ON sys_config ("key");
CREATE INDEX IF NOT EXISTS idx_sys_config_value ON sys_config ("value");
DROP INDEX IF EXISTS idx_sys_config_old;
REINDEX uk_sys_config_key;
`

	var messages []string
	log := logger.Logger()
	previousLevel := log.GetLevel()
	log.SetLevel(previousLevel | glog.LEVEL_DEBU)
	log.SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		messages = append(messages, in.ValuesContent())
	})
	t.Cleanup(func() {
		log.SetHandlers()
		log.SetLevel(previousLevel)
	})

	translated, err := TranslateDDL(ctx, "comments-indexes.sql", input)
	if err != nil {
		t.Fatalf("translate comment and index fixture failed: %v", err)
	}
	if strings.Contains(translated, "COMMENT ON") {
		t.Fatalf("expected COMMENT ON statements to be dropped, got:\n%s", translated)
	}
	required := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_config_key ON sys_config ("key");`,
		`CREATE INDEX IF NOT EXISTS idx_sys_config_value ON sys_config ("value");`,
		`DROP INDEX IF EXISTS idx_sys_config_old;`,
		`REINDEX uk_sys_config_key;`,
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to preserve %q, got:\n%s", needle, translated)
		}
	}
	if len(messages) != 2 {
		t.Fatalf("expected two skipped comment debug logs, got %d: %#v", len(messages), messages)
	}
	for _, needle := range []string{
		"skip PostgreSQL comment statement during SQLite translation",
		"source=comments-indexes.sql",
		"line=",
	} {
		if !containsAnyLogMessage(messages, needle) {
			t.Fatalf("expected skipped comment debug logs to contain %q, got %#v", needle, messages)
		}
	}
}

// TestTranslateDDLIgnoresKeywordsInsideQuotes verifies unsupported-keyword
// detection and type conversion do not inspect strings or quoted identifiers.
func TestTranslateDDLIgnoresKeywordsInsideQuotes(t *testing.T) {
	t.Parallel()

	input := `
CREATE TABLE IF NOT EXISTS quoted_demo (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "serial" VARCHAR(64) NOT NULL,
    note TEXT NOT NULL
);
INSERT INTO quoted_demo ("serial", note)
VALUES ('JSONB SERIAL AUTO_INCREMENT should stay literal', 'DOUBLE PRECISION literal')
ON CONFLICT DO NOTHING;
`
	translated, err := TranslateDDL(context.Background(), "quoted.sql", input)
	if err != nil {
		t.Fatalf("translate quoted fixture failed: %v", err)
	}
	for _, needle := range []string{
		`"serial" TEXT NOT NULL`,
		`'JSONB SERIAL AUTO_INCREMENT should stay literal'`,
		`'DOUBLE PRECISION literal'`,
	} {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}
}

// containsAnyLogMessage reports whether any captured log contains a fragment.
func containsAnyLogMessage(messages []string, needle string) bool {
	for _, message := range messages {
		if strings.Contains(message, needle) {
			return true
		}
	}
	return false
}

// TestTranslateDDLPreservesDMLIdentifiersNamedLikeTypes verifies DML column
// names are not treated as type declarations.
func TestTranslateDDLPreservesDMLIdentifiersNamedLikeTypes(t *testing.T) {
	t.Parallel()

	input := `INSERT INTO demo (smallint, "varchar", note) VALUES (1, 'x', 'timestamp') ON CONFLICT DO NOTHING;`
	translated, err := TranslateDDL(context.Background(), "dml-identifiers.sql", input)
	if err != nil {
		t.Fatalf("translate DML identifiers fixture failed: %v", err)
	}
	required := []string{
		"(smallint, \"varchar\", note)",
		"'timestamp'",
		"ON CONFLICT DO NOTHING",
	}
	for _, needle := range required {
		if !strings.Contains(translated, needle) {
			t.Fatalf("expected translated SQL to contain %q, got:\n%s", needle, translated)
		}
	}
	if strings.Contains(translated, "INTEGER,") || strings.Contains(translated, "\"TEXT\"") {
		t.Fatalf("expected DML identifiers not to be type-rewritten, got:\n%s", translated)
	}
}

// TestTranslateDDLUnsupportedPostgreSQLSyntax verifies unsupported PG features
// fail with source, line, and keyword diagnostics.
func TestTranslateDDLUnsupportedPostgreSQLSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceName string
		input      string
		keyword    string
	}{
		{
			name:       "jsonb",
			sourceName: "bad/jsonb.sql",
			input:      "\n\nCREATE TABLE demo (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, payload JSONB);",
			keyword:    "JSONB",
		},
		{
			name:       "trigger",
			sourceName: "bad/trigger.sql",
			input:      "CREATE TRIGGER trg_demo BEFORE UPDATE ON demo EXECUTE FUNCTION touch_demo();",
			keyword:    "CREATE TRIGGER",
		},
		{
			name:       "serial",
			sourceName: "bad/serial.sql",
			input:      "CREATE TABLE demo (id BIGSERIAL PRIMARY KEY);",
			keyword:    "BIGSERIAL",
		},
		{
			name:       "merge",
			sourceName: "bad/merge.sql",
			input:      "MERGE INTO demo USING incoming ON demo.id = incoming.id WHEN MATCHED THEN UPDATE SET name = incoming.name;",
			keyword:    "MERGE",
		},
		{
			name:       "generated expression",
			sourceName: "bad/generated-expression.sql",
			input:      "CREATE TABLE demo (code TEXT, code_upper TEXT GENERATED ALWAYS AS (upper(code)) STORED);",
			keyword:    "GENERATED ALWAYS AS",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := TranslateDDL(context.Background(), test.sourceName, test.input)
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

// TestTranslateDDLRejectsLegacyMySQLSource verifies old MySQL-source syntax no
// longer receives best-effort translation.
func TestTranslateDDLRejectsLegacyMySQLSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		keyword string
	}{
		{
			name:    "auto increment",
			input:   "CREATE TABLE demo (id INT PRIMARY KEY AUTO_INCREMENT);",
			keyword: "AUTO_INCREMENT",
		},
		{
			name:    "insert ignore",
			input:   "INSERT IGNORE INTO demo (name) VALUES ('x');",
			keyword: "INSERT IGNORE",
		},
		{
			name:    "engine",
			input:   "CREATE TABLE demo (id INT) ENGINE=InnoDB;",
			keyword: "ENGINE/CHARSET",
		},
		{
			name:    "backticks",
			input:   "CREATE TABLE `demo` (`id` INT);",
			keyword: "backtick identifier",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := TranslateDDL(context.Background(), "legacy.sql", test.input)
			if err == nil {
				t.Fatal("expected legacy MySQL source SQL to fail")
			}
			if !strings.Contains(err.Error(), test.keyword) {
				t.Fatalf("expected error to contain %q, got %q", test.keyword, err.Error())
			}
		})
	}
}
