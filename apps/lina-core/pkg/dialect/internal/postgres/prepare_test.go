// This file tests PostgreSQL database preparation link parsing and validation.

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"os"
	"strings"
	"testing"
	"time"
)

// TestConfigNodeFromLink verifies database links are parsed by GoFrame's
// database component instead of package-local string splitting.
func TestConfigNodeFromLink(t *testing.T) {
	t.Parallel()

	link := "pgsql:postgres:secret@tcp(127.0.0.1:5432)/linapro?sslmode=disable&connect_timeout=5"
	node, err := ConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse PostgreSQL config node failed: %v", err)
	}
	if node.Type != "pgsql" {
		t.Fatalf("expected pgsql type, got %q", node.Type)
	}
	if node.User != "postgres" {
		t.Fatalf("expected postgres user, got %q", node.User)
	}
	if node.Pass != "secret" {
		t.Fatalf("expected password to be preserved, got %q", node.Pass)
	}
	if node.Protocol != "tcp" {
		t.Fatalf("expected tcp protocol, got %q", node.Protocol)
	}
	if node.Host != "127.0.0.1" {
		t.Fatalf("expected localhost host, got %q", node.Host)
	}
	if node.Port != "5432" {
		t.Fatalf("expected 5432 port, got %q", node.Port)
	}
	if node.Name != "linapro" {
		t.Fatalf("expected linapro database name, got %q", node.Name)
	}
	if !strings.Contains(node.Extra, "sslmode=disable") || !strings.Contains(node.Extra, "connect_timeout=5") {
		t.Fatalf("expected extra parameters to be preserved, got %q", node.Extra)
	}
}

// TestConfigNodeFromLinkRejectsMissingDatabase verifies init refuses links that
// cannot identify the target database to create or rebuild.
func TestConfigNodeFromLinkRejectsMissingDatabase(t *testing.T) {
	t.Parallel()

	for _, link := range []string{
		"pgsql:postgres:secret@tcp(127.0.0.1:5432)",
		"pgsql:postgres:secret@tcp(127.0.0.1:5432)/?sslmode=disable",
	} {
		link := link
		t.Run(link, func(t *testing.T) {
			t.Parallel()

			if _, err := ConfigNodeFromLink(link); err == nil {
				t.Fatal("expected missing database name error")
			}
		})
	}
}

// TestQuoteIdentifier verifies bootstrap SQL escapes PostgreSQL identifiers
// instead of concatenating raw names.
func TestQuoteIdentifier(t *testing.T) {
	t.Parallel()

	got, err := QuoteIdentifier(`lina"pro`)
	if err != nil {
		t.Fatalf("quote identifier failed: %v", err)
	}
	if got != `"lina""pro"` {
		t.Fatalf("expected escaped identifier, got %q", got)
	}
	if _, err = QuoteIdentifier(""); err == nil {
		t.Fatal("expected empty identifier error")
	}
	if _, err = QuoteIdentifier("bad\x00name"); err == nil {
		t.Fatal("expected NUL identifier error")
	}
}

// TestPrepareDatabaseRejectsInvalidLinks verifies validation errors return
// before any PostgreSQL network connection is attempted.
func TestPrepareDatabaseRejectsInvalidLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		link string
	}{
		{name: "invalid pattern", link: "pgsql:not-a-valid-link"},
		{name: "missing database", link: "pgsql:postgres:secret@tcp(127.0.0.1:5432)/"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := PrepareDatabase(context.Background(), test.link, false); err == nil {
				t.Fatal("expected invalid PostgreSQL link to fail")
			}
		})
	}
}

// TestPrepareDatabaseRebuildWithPostgreSQL verifies the real PostgreSQL
// rebuild path drops the target database after terminating active connections.
func TestPrepareDatabaseRebuildWithPostgreSQL(t *testing.T) {
	link := postgresIntegrationLink(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	targetLink := postgresLinkWithDatabase(t, link, fmt.Sprintf("linapro_rebuild_%d", time.Now().UnixNano()))
	if err := PrepareDatabase(ctx, targetLink, false); err != nil {
		t.Fatalf("prepare PostgreSQL target database failed: %v", err)
	}
	t.Cleanup(func() {
		dropPostgresIntegrationDatabase(t, targetLink)
	})

	targetDB := openPostgresIntegrationDB(t, targetLink)
	t.Cleanup(func() {
		if closeErr := targetDB.Close(context.Background()); closeErr != nil && !isExpectedTerminatedConnectionError(closeErr) {
			t.Errorf("close target PostgreSQL database failed: %v", closeErr)
		}
	})
	if _, err := targetDB.Exec(ctx, `CREATE TABLE rebuild_marker (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY)`); err != nil {
		t.Fatalf("create rebuild marker table failed: %v", err)
	}

	blockingDB := openPostgresIntegrationDB(t, targetLink)
	if _, err := blockingDB.Exec(ctx, `SELECT 1`); err != nil {
		t.Fatalf("establish active PostgreSQL target connection failed: %v", err)
	}

	if err := PrepareDatabase(ctx, targetLink, true); err != nil {
		if closeErr := blockingDB.Close(context.Background()); closeErr != nil && !isExpectedTerminatedConnectionError(closeErr) {
			t.Errorf("close blocking PostgreSQL connection after rebuild failure failed: %v", closeErr)
		}
		t.Fatalf("rebuild PostgreSQL target database failed: %v", err)
	}
	if closeErr := blockingDB.Close(context.Background()); closeErr != nil && !isExpectedTerminatedConnectionError(closeErr) {
		t.Fatalf("close terminated PostgreSQL connection failed: %v", closeErr)
	}

	reopened := openPostgresIntegrationDB(t, targetLink)
	t.Cleanup(func() {
		if closeErr := reopened.Close(context.Background()); closeErr != nil {
			t.Errorf("close rebuilt PostgreSQL database failed: %v", closeErr)
		}
	})
	exists, err := postgresTableExists(ctx, reopened, "rebuild_marker")
	if err != nil {
		t.Fatalf("check rebuilt marker table failed: %v", err)
	}
	if exists {
		t.Fatal("expected rebuild to drop previous marker table")
	}
}

// postgresIntegrationLink returns the explicit PostgreSQL integration-test
// link, skipping when the caller did not opt in to database-backed tests.
func postgresIntegrationLink(t *testing.T) string {
	t.Helper()

	link := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if link == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL integration tests")
	}
	return link
}

// openPostgresIntegrationDB opens a PostgreSQL database and arranges for
// callers to own the returned connection lifetime.
func openPostgresIntegrationDB(t *testing.T, link string) gdb.DB {
	t.Helper()

	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		t.Fatalf("open PostgreSQL database failed: %v", err)
	}
	return db
}

// postgresLinkWithDatabase returns link with its database name replaced.
func postgresLinkWithDatabase(t *testing.T, link string, databaseName string) string {
	t.Helper()

	node, err := ConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse PostgreSQL link failed: %v", err)
	}
	return postgresLinkFromNode(*node, databaseName)
}

// postgresSystemLink returns link pointed at the PostgreSQL maintenance
// database.
func postgresSystemLink(t *testing.T, link string) string {
	t.Helper()

	node, err := ConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse PostgreSQL link failed: %v", err)
	}
	return postgresLinkFromNode(*node, systemDatabaseName)
}

// postgresLinkFromNode formats the GoFrame pgsql link from a parsed config.
func postgresLinkFromNode(node gdb.ConfigNode, databaseName string) string {
	extra := strings.TrimSpace(node.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/%s%s", node.User, node.Pass, node.Protocol, node.Host, node.Port, databaseName, extra)
}

// dropPostgresIntegrationDatabase removes one temporary database created by a
// PostgreSQL integration test.
func dropPostgresIntegrationDatabase(t *testing.T, targetLink string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	targetNode, err := ConfigNodeFromLink(targetLink)
	if err != nil {
		t.Errorf("parse target PostgreSQL link for cleanup failed: %v", err)
		return
	}
	systemDB := openPostgresIntegrationDB(t, postgresSystemLink(t, targetLink))
	defer func() {
		if closeErr := systemDB.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL cleanup connection failed: %v", closeErr)
		}
	}()
	if _, err = systemDB.Exec(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		targetNode.Name,
	); err != nil {
		t.Errorf("terminate PostgreSQL cleanup connections failed: %v", err)
		return
	}
	quotedName, err := QuoteIdentifier(targetNode.Name)
	if err != nil {
		t.Errorf("quote PostgreSQL cleanup database name failed: %v", err)
		return
	}
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		t.Errorf("drop PostgreSQL integration database failed: %v", err)
	}
}

// postgresTableExists reports whether a table exists in the public schema.
func postgresTableExists(ctx context.Context, db gdb.DB, tableName string) (bool, error) {
	value, err := db.GetValue(
		ctx,
		`SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1`,
		tableName,
	)
	if err != nil {
		return false, err
	}
	return !value.IsNil(), nil
}

// isExpectedTerminatedConnectionError treats server-terminated connections as
// valid after rebuild explicitly kills active backends.
func isExpectedTerminatedConnectionError(err error) bool {
	if err == nil || errors.Is(err, sql.ErrConnDone) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "closed") ||
		strings.Contains(message, "bad connection") ||
		strings.Contains(message, "terminating connection") ||
		strings.Contains(message, "conn busy")
}
