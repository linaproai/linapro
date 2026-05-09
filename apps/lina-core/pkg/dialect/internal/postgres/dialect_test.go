// This file tests PostgreSQL dialect identity and DDL behavior.

package postgres

import (
	"context"
	"testing"
)

// TestDialectBasics verifies stable PostgreSQL dialect identity and DDL
// passthrough behavior.
func TestDialectBasics(t *testing.T) {
	t.Parallel()

	ddl := "CREATE TABLE demo (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"
	translated, err := TranslateDDL(context.Background(), "demo.sql", ddl)
	if err != nil {
		t.Fatalf("translate PostgreSQL DDL failed: %v", err)
	}
	if translated != ddl {
		t.Fatalf("expected PostgreSQL DDL to be unchanged, got %q", translated)
	}
	if Name != "postgres" {
		t.Fatalf("expected PostgreSQL dialect name, got %q", Name)
	}
	if !SupportsCluster() {
		t.Fatal("expected PostgreSQL dialect to support cluster mode")
	}
}

// TestDatabaseVersionRejectsNilDB verifies the version helper reports a clear
// error when called without a database handle.
func TestDatabaseVersionRejectsNilDB(t *testing.T) {
	t.Parallel()

	if _, err := DatabaseVersion(context.Background(), nil); err == nil {
		t.Fatal("expected nil PostgreSQL database version query to fail")
	}
}
