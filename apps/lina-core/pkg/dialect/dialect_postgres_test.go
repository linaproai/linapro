// This file tests PostgreSQL dialect factory behavior.

package dialect

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestFromResolvesPostgreSQL verifies pgsql: links dispatch to the PostgreSQL
// dialect and preserve PostgreSQL-source DDL.
func TestFromResolvesPostgreSQL(t *testing.T) {
	t.Parallel()

	dbDialect, err := From("pgsql:postgres:secret@tcp(127.0.0.1:5432)/linapro?sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if dbDialect.Name() != "postgres" {
		t.Fatalf("expected postgres dialect, got %s", dbDialect.Name())
	}
	if !dbDialect.SupportsCluster() {
		t.Fatal("expected PostgreSQL dialect to support cluster mode")
	}

	ddl := "CREATE TABLE demo (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"
	translated, err := dbDialect.TranslateDDL(context.Background(), "demo.sql", ddl)
	if err != nil {
		t.Fatalf("translate PostgreSQL DDL failed: %v", err)
	}
	if translated != ddl {
		t.Fatalf("expected PostgreSQL DDL to be unchanged, got %q", translated)
	}
}

// TestFromRejectsMySQL verifies MySQL links fail with the explicit removal
// error instead of falling through to the generic unsupported-dialect path.
func TestFromRejectsMySQL(t *testing.T) {
	t.Parallel()

	_, err := From("mysql:root:secret@tcp(127.0.0.1:3306)/linapro")
	if err == nil {
		t.Fatal("expected MySQL dialect to be rejected")
	}
	if !bizerr.Is(err, CodeDialectMySQLUnsupported) {
		t.Fatalf("expected MySQL unsupported business error, got %v", err)
	}
}

// TestFromRejectsPostgresAlias verifies only GoFrame's pgsql: prefix is
// accepted for PostgreSQL links.
func TestFromRejectsPostgresAlias(t *testing.T) {
	t.Parallel()

	if _, err := From("postgres:postgres://localhost/linapro"); err == nil {
		t.Fatal("expected postgres: alias to be unsupported")
	}
}
