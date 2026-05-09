// This file verifies PostgreSQL dialect behavior against a real database when
// integration tests are explicitly enabled.

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
)

// TestPostgreSQLDriverAndORMReservedColumns verifies the configured GoFrame
// PostgreSQL driver can connect, insert identity rows, and quote reserved
// columns in ORM operations.
func TestPostgreSQLDriverAndORMReservedColumns(t *testing.T) {
	link := postgresIntegrationLink(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := openPostgresIntegrationDB(t, link)
	t.Cleanup(func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL database failed: %v", closeErr)
		}
	})

	version, err := DatabaseVersion(ctx, db)
	if err != nil {
		t.Fatalf("query PostgreSQL version failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(version), "postgresql") {
		t.Fatalf("expected PostgreSQL version string, got %q", version)
	}

	tableName := "dialect_kv_identity_test"
	if _, err = db.Exec(ctx, `DROP TABLE IF EXISTS dialect_kv_identity_test`); err != nil {
		t.Fatalf("drop stale reserved-column table failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := db.Exec(context.Background(), `DROP TABLE IF EXISTS dialect_kv_identity_test`); cleanupErr != nil {
			t.Errorf("drop reserved-column table failed: %v", cleanupErr)
		}
	})
	if _, err = db.Exec(ctx, `CREATE TABLE dialect_kv_identity_test (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		"key" VARCHAR(64) NOT NULL,
		"value" TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create reserved-column table failed: %v", err)
	}

	result, err := db.Model(tableName).Ctx(ctx).Data(gdb.Map{
		"key":        "alpha",
		"value":      "first",
		"created_at": gtime.Now(),
	}).Insert()
	if err != nil {
		t.Fatalf("insert reserved-column row through ORM failed: %v", err)
	}
	insertID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read PostgreSQL identity insert id failed: %v", err)
	}
	if insertID <= 0 {
		t.Fatalf("expected positive identity id, got %d", insertID)
	}

	var row struct {
		Key   string `orm:"key"`
		Value string `orm:"value"`
	}
	if err = db.Model(tableName).Ctx(ctx).Where("key", "alpha").Scan(&row); err != nil {
		t.Fatalf("query reserved-column row through ORM failed: %v", err)
	}
	if row.Key != "alpha" || row.Value != "first" {
		t.Fatalf("unexpected reserved-column row: %#v", row)
	}
}

// TestPostgreSQLDefaultCollationKeepsCaseDistinctKeys verifies the delivered
// PostgreSQL schema no longer depends on custom case-insensitive collations.
func TestPostgreSQLDefaultCollationKeepsCaseDistinctKeys(t *testing.T) {
	link := postgresIntegrationLink(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := openPostgresIntegrationDB(t, link)
	t.Cleanup(func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL database failed: %v", closeErr)
		}
	})

	if _, err := db.Exec(ctx, `DROP TABLE IF EXISTS dialect_default_collation_test`); err != nil {
		t.Fatalf("drop stale default-collation table failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := db.Exec(context.Background(), `DROP TABLE IF EXISTS dialect_default_collation_test`); cleanupErr != nil {
			t.Errorf("drop default-collation table failed: %v", cleanupErr)
		}
	})
	if _, err := db.Exec(ctx, `CREATE TABLE dialect_default_collation_test (
		username VARCHAR(64) NOT NULL
	)`); err != nil {
		t.Fatalf("create default-collation table failed: %v", err)
	}
	if _, err := db.Exec(ctx, `CREATE UNIQUE INDEX uk_dialect_default_collation_username ON dialect_default_collation_test (username)`); err != nil {
		t.Fatalf("create default-collation unique index failed: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO dialect_default_collation_test (username) VALUES ($1)`, "admin"); err != nil {
		t.Fatalf("insert first default-collation row failed: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO dialect_default_collation_test (username) VALUES ($1)`, "Admin"); err != nil {
		t.Fatalf("expected case-distinct username to be accepted: %v", err)
	}
	count, err := db.GetValue(ctx, `SELECT COUNT(*) FROM dialect_default_collation_test`)
	if err != nil {
		t.Fatalf("read default-collation row count failed: %v", err)
	}
	if count.String() != "2" {
		t.Fatalf("expected two case-distinct rows, got %s", count.String())
	}
}
