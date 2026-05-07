// This file tests SQLite database preparation and link parsing.

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestPathFromLink verifies GoFrame SQLite file links are parsed without
// accepting unsupported shorthand forms.
func TestPathFromLink(t *testing.T) {
	t.Parallel()

	path, err := PathFromLink("sqlite::@file(./temp/sqlite/linapro.db)")
	if err != nil {
		t.Fatalf("parse SQLite link failed: %v", err)
	}
	if path != "./temp/sqlite/linapro.db" {
		t.Fatalf("expected SQLite path ./temp/sqlite/linapro.db, got %s", path)
	}

	if _, err = PathFromLink("sqlite::(./temp/sqlite/linapro.db)"); err == nil {
		t.Fatal("expected unsupported SQLite shorthand link to fail")
	}
}

// TestPrepareDatabaseCreatesParentDir verifies init can create a SQLite parent
// directory before the database file exists.
func TestPrepareDatabaseCreatesParentDir(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "nested", "linapro.db")
	link := "sqlite::@file(" + dbPath + ")"
	if err := PrepareDatabase(context.Background(), link, false); err != nil {
		t.Fatalf("prepare SQLite database failed: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); err != nil {
		t.Fatalf("expected SQLite parent directory to exist: %v", err)
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatalf("expected SQLite db file to be created later by the driver, got err=%v", err)
	}
}

// TestPrepareDatabaseRebuildDeletesDatabaseFiles verifies rebuild removes the
// primary database file and common WAL sidecar files.
func TestPrepareDatabaseRebuildDeletesDatabaseFiles(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "linapro.db")
	for _, path := range DatabaseFiles(dbPath) {
		if err := os.WriteFile(path, []byte("stale"), 0o600); err != nil {
			t.Fatalf("seed SQLite file %s failed: %v", path, err)
		}
	}

	link := "sqlite::@file(" + dbPath + ")"
	if err := PrepareDatabase(context.Background(), link, true); err != nil {
		t.Fatalf("rebuild SQLite database failed: %v", err)
	}
	for _, path := range DatabaseFiles(dbPath) {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected SQLite file %s to be deleted, got err=%v", path, err)
		}
	}
}
