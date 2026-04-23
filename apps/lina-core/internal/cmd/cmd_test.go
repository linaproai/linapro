// This file verifies shared command helpers for explicit confirmations, SQL
// directory conventions, and SQL file execution behavior.

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/os/gfile"
)

// TestRequireCommandConfirmation verifies sensitive command confirmation tokens
// are enforced for init, mock, and upgrade operations.
func TestRequireCommandConfirmation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		commandName    string
		confirmValue   string
		wantErr        bool
		wantSubstrings []string
	}{
		{
			name:         "init accepts matching confirmation",
			commandName:  initCommandName,
			confirmValue: initCommandName,
		},
		{
			name:         "mock accepts matching confirmation",
			commandName:  mockCommandName,
			confirmValue: mockCommandName,
		},
		{
			name:         "upgrade accepts matching confirmation",
			commandName:  upgradeCommandName,
			confirmValue: upgradeCommandName,
		},
		{
			name:         "init rejects missing confirmation",
			commandName:  initCommandName,
			confirmValue: "",
			wantErr:      true,
			wantSubstrings: []string{
				"命令 init 涉及敏感升级或数据库操作",
				makeConfirmationExample(initCommandName),
				goRunConfirmationExample(initCommandName),
			},
		},
		{
			name:         "mock rejects wrong confirmation",
			commandName:  mockCommandName,
			confirmValue: initCommandName,
			wantErr:      true,
			wantSubstrings: []string{
				"命令 mock 涉及敏感升级或数据库操作",
				makeConfirmationExample(mockCommandName),
				goRunConfirmationExample(mockCommandName),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requireCommandConfirmation(tt.commandName, tt.confirmValue)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for command %q", tt.commandName)
				}
				for _, substring := range tt.wantSubstrings {
					if !strings.Contains(err.Error(), substring) {
						t.Fatalf("expected error %q to contain %q", err.Error(), substring)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestHostSqlDirsFollowConvention verifies the init and mock SQL helpers keep
// using the expected manifest directory layout.
func TestHostSqlDirsFollowConvention(t *testing.T) {
	t.Parallel()

	if got := hostInitSqlDir(); got != "manifest/sql" {
		t.Fatalf("expected init sql dir %q, got %q", "manifest/sql", got)
	}
	if got := hostMockSqlDir(); got != gfile.Join("manifest/sql", "mock-data") {
		t.Fatalf("expected mock sql dir %q, got %q", gfile.Join("manifest/sql", "mock-data"), got)
	}
}

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

// writeTestSQLFile writes one temporary SQL file for shared command helper tests.
func writeTestSQLFile(t *testing.T, dir string, name string, contents string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
