// This file verifies planning helpers and SQL replay helpers used by the
// framework-upgrade service.

package frameworkupgrade

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSortSemanticVersionsDesc verifies semver values sort from newest to oldest.
func TestSortSemanticVersionsDesc(t *testing.T) {
	t.Parallel()

	sorted, err := sortSemanticVersionsDesc([]string{"v0.5.0", "v0.7.0", "v0.6.1"})
	if err != nil {
		t.Fatalf("sort versions: %v", err)
	}
	expected := []string{"v0.7.0", "v0.6.1", "v0.5.0"}
	if !reflect.DeepEqual(sorted, expected) {
		t.Fatalf("expected %v, got %v", expected, sorted)
	}
}

// TestParseRemoteTagsOutputSkipsInvalidRefs verifies only semver tags are returned.
func TestParseRemoteTagsOutputSkipsInvalidRefs(t *testing.T) {
	t.Parallel()

	output := `111 refs/tags/v0.5.0
222 refs/tags/not-a-version
333 refs/tags/v0.6.0
`
	tags := parseRemoteTagsOutput(output)
	expected := []string{"v0.5.0", "v0.6.0"}
	if !reflect.DeepEqual(tags, expected) {
		t.Fatalf("expected %v, got %v", expected, tags)
	}
}

// TestScanTargetSQLFilesSortsFiles verifies host SQL replay always starts from
// the first sorted SQL file in the target release.
func TestScanTargetSQLFilesSortsFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sqlDir := filepath.Join(tempDir, "apps", "lina-core", "manifest", "sql")
	if err := os.MkdirAll(sqlDir, 0o755); err != nil {
		t.Fatalf("mkdir sql dir: %v", err)
	}

	writeFrameworkUpgradeFile(t, filepath.Join(sqlDir, "010-third.sql"), "THIRD")
	writeFrameworkUpgradeFile(t, filepath.Join(sqlDir, "001-first.sql"), "FIRST")
	writeFrameworkUpgradeFile(t, filepath.Join(sqlDir, "002-second.sql"), "SECOND")

	files, err := scanTargetSQLFiles(tempDir)
	if err != nil {
		t.Fatalf("scan target sql files: %v", err)
	}
	expected := []string{
		filepath.Join(sqlDir, "001-first.sql"),
		filepath.Join(sqlDir, "002-second.sql"),
		filepath.Join(sqlDir, "010-third.sql"),
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %v, got %v", expected, files)
	}
}

// TestExecuteUpgradeSQLFilesWithExecutorStopsAfterFirstError verifies replay
// stops at the first failing SQL file and keeps the execution order stable.
func TestExecuteUpgradeSQLFilesWithExecutorStopsAfterFirstError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	files := []string{
		writeFrameworkUpgradeFile(t, filepath.Join(tempDir, "001-first.sql"), "FIRST"),
		writeFrameworkUpgradeFile(t, filepath.Join(tempDir, "002-second.sql"), "SECOND"),
		writeFrameworkUpgradeFile(t, filepath.Join(tempDir, "003-third.sql"), "THIRD"),
	}

	var executedSQL []string
	executedFiles, err := executeUpgradeSQLFilesWithExecutor(context.Background(), files, func(ctx context.Context, sql string) error {
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
	if !reflect.DeepEqual(executedFiles, []string{"001-first.sql"}) {
		t.Fatalf("expected successful files %v, got %v", []string{"001-first.sql"}, executedFiles)
	}
}

// TestReadTargetFrameworkMetadata verifies framework metadata can be loaded from one target checkout.
func TestReadTargetFrameworkMetadata(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	metadataPath := filepath.Join(tempDir, "apps", "lina-core", "manifest", "config", "metadata.yaml")
	writeFrameworkUpgradeFile(t, metadataPath, `framework:
  name: "LinaPro"
  version: "v1.2.3"
  description: "Framework"
  homepage: "https://linapro.ai"
  repositoryUrl: "https://github.com/example/linapro"
  license: "MIT"
`)

	info, err := readTargetFrameworkMetadata(tempDir)
	if err != nil {
		t.Fatalf("read target framework metadata: %v", err)
	}
	if info.Name != "LinaPro" {
		t.Fatalf("expected name LinaPro, got %q", info.Name)
	}
	if info.Version != "v1.2.3" {
		t.Fatalf("expected version v1.2.3, got %q", info.Version)
	}
	if info.RepositoryURL != "https://github.com/example/linapro" {
		t.Fatalf("expected repository url https://github.com/example/linapro, got %q", info.RepositoryURL)
	}
}

// writeFrameworkUpgradeFile writes one temporary file for framework-upgrade tests.
func writeFrameworkUpgradeFile(t *testing.T, path string, contents string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
