// This file verifies SQL execution helpers stop on failure and ignore empty
// files during manifest SQL processing.

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestExecuteSQLFilesWithExecutorStopsAfterFirstError verifies execution halts
// at the first failing SQL file and returns the failing file name.
func TestExecuteSQLFilesWithExecutorStopsAfterFirstError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	files := []string{
		writeTestSQLFile(t, tempDir, "001-first.sql", "FIRST"),
		writeTestSQLFile(t, tempDir, "002-second.sql", "SECOND"),
		writeTestSQLFile(t, tempDir, "003-third.sql", "THIRD"),
	}

	var executedSQL []string
	err := executeSQLFilesWithExecutor(context.Background(), files, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		if sql == "SECOND" {
			return errors.New("boom")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected execution error")
	}
	if !strings.Contains(err.Error(), "002-second.sql") {
		t.Fatalf("expected error %q to contain failing file name", err.Error())
	}
	if !reflect.DeepEqual(executedSQL, []string{"FIRST", "SECOND"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"FIRST", "SECOND"}, executedSQL)
	}
}

// TestExecuteSQLFilesWithExecutorSkipsEmptyFiles verifies blank SQL files are
// ignored while non-empty files still execute in order.
func TestExecuteSQLFilesWithExecutorSkipsEmptyFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	files := []string{
		writeTestSQLFile(t, tempDir, "001-empty.sql", ""),
		writeTestSQLFile(t, tempDir, "002-seed.sql", "SEED"),
	}

	var executedSQL []string
	err := executeSQLFilesWithExecutor(context.Background(), files, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(executedSQL, []string{"SEED"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"SEED"}, executedSQL)
	}
}

// writeTestSQLFile writes one temporary SQL file for command helper tests.
func writeTestSQLFile(t *testing.T, dir string, name string, contents string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
