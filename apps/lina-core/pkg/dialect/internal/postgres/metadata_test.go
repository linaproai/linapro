// This file tests PostgreSQL table metadata queries.

package postgres

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"os"
	"strings"
	"testing"
	"time"
)

// TestQueryTableMetadataEmptyInput verifies the no-table path remains a pure
// unit test that does not require a database connection.
func TestQueryTableMetadataEmptyInput(t *testing.T) {
	t.Parallel()

	metas, err := QueryTableMetadata(context.Background(), nil, "public", nil)
	if err != nil {
		t.Fatalf("query empty PostgreSQL metadata failed: %v", err)
	}
	if len(metas) != 0 {
		t.Fatalf("expected no metadata for empty table input, got %#v", metas)
	}
}

// TestQueryTableMetadataWithPostgreSQL verifies metadata queries against a real
// PostgreSQL database when LINA_TEST_PGSQL_LINK is explicitly provided.
func TestQueryTableMetadataWithPostgreSQL(t *testing.T) {
	link := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if link == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL metadata integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		t.Fatalf("open PostgreSQL database failed: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL database failed: %v", closeErr)
		}
	})

	tableName := "dialect_metadata_test"
	if _, err = db.Exec(ctx, `DROP TABLE IF EXISTS dialect_metadata_test`); err != nil {
		t.Fatalf("drop stale PostgreSQL metadata table failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := db.Exec(context.Background(), `DROP TABLE IF EXISTS dialect_metadata_test`); cleanupErr != nil {
			t.Errorf("drop PostgreSQL metadata table failed: %v", cleanupErr)
		}
	})
	if _, err = db.Exec(ctx, `CREATE TABLE dialect_metadata_test (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY)`); err != nil {
		t.Fatalf("create PostgreSQL metadata table failed: %v", err)
	}
	if _, err = db.Exec(ctx, `COMMENT ON TABLE dialect_metadata_test IS 'metadata test table'`); err != nil {
		t.Fatalf("comment PostgreSQL metadata table failed: %v", err)
	}

	metas, err := QueryTableMetadata(ctx, db, "", []string{tableName, "missing_table"})
	if err != nil {
		t.Fatalf("query PostgreSQL metadata failed: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("expected one metadata row, got %#v", metas)
	}
	if metas[0].TableName != tableName {
		t.Fatalf("expected table name %q, got %q", tableName, metas[0].TableName)
	}
	if metas[0].TableComment != "metadata test table" {
		t.Fatalf("expected table comment, got %q", metas[0].TableComment)
	}
}
