// This file verifies SQL statement splitting, asset loading, and per-statement
// execution for development-only init/mock command internals.

package sqlassets

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSplitSQLStatementsSkipsCommentsAndWhitespace verifies comment-only
// fragments and blank sections are ignored during statement parsing.
func TestSplitSQLStatementsSkipsCommentsAndWhitespace(t *testing.T) {
	t.Parallel()

	content := `
-- file header comment

/* block comment */
CREATE TABLE demo(id INT);

-- trailing comment
INSERT INTO demo(id) VALUES (1);
`

	got := Split(content)
	want := []string{
		"CREATE TABLE demo(id INT)",
		"INSERT INTO demo(id) VALUES (1)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected statements %v, got %v", want, got)
	}
}

// TestSplitSQLStatementsKeepsSemicolonsInsideLiterals verifies semicolons
// inside quoted literals or identifiers do not split statements.
func TestSplitSQLStatementsKeepsSemicolonsInsideLiterals(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"INSERT INTO demo(note) VALUES ('a;b');",
		"INSERT INTO demo(note) VALUES (\"c;d\");",
		"SELECT `semi;colon` FROM demo;",
	}, "\n")

	got := Split(content)
	want := []string{
		"INSERT INTO demo(note) VALUES ('a;b')",
		"INSERT INTO demo(note) VALUES (\"c;d\")",
		"SELECT `semi;colon` FROM demo",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected statements %v, got %v", want, got)
	}
}

// TestExecuteSQLAssetsWithExecutorSplitsMultiStatementFiles verifies one SQL
// file containing multiple statements executes each statement in order.
func TestExecuteSQLAssetsWithExecutorSplitsMultiStatementFiles(t *testing.T) {
	t.Parallel()

	assets := []Asset{
		{Path: "manifest/sql/001-seed.sql", Content: "FIRST;\nSECOND;\nTHIRD;"},
	}

	var executedSQL []string
	err := ExecuteWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(executedSQL, []string{"FIRST", "SECOND", "THIRD"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"FIRST", "SECOND", "THIRD"}, executedSQL)
	}
}

// TestExecuteSQLAssetsWithExecutorStopsAfterFailingStatement verifies failure
// inside one multi-statement SQL file stops the remaining statements
// immediately and reports the statement index.
func TestExecuteSQLAssetsWithExecutorStopsAfterFailingStatement(t *testing.T) {
	t.Parallel()

	assets := []Asset{
		{Path: "manifest/sql/001-seed.sql", Content: "FIRST;\nSECOND;\nTHIRD;"},
		{Path: "manifest/sql/002-after.sql", Content: "FOURTH;"},
	}

	var executedSQL []string
	err := ExecuteWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		if sql == "SECOND" {
			return errors.New("boom")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected execution error")
	}
	if !strings.Contains(err.Error(), "001-seed.sql") {
		t.Fatalf("expected error %q to contain failing file name", err.Error())
	}
	if !strings.Contains(err.Error(), "statement 2") {
		t.Fatalf("expected error %q to contain failing statement index", err.Error())
	}
	if !reflect.DeepEqual(executedSQL, []string{"FIRST", "SECOND"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"FIRST", "SECOND"}, executedSQL)
	}
}

// TestHostSQLDirsFollowConvention verifies the init and mock SQL helpers keep
// using the expected manifest directory layout.
func TestHostSQLDirsFollowConvention(t *testing.T) {
	t.Parallel()

	if got := HostInitDir(); got != "manifest/sql" {
		t.Fatalf("expected init sql dir %q, got %q", "manifest/sql", got)
	}
	if got := HostMockDir(); got != path.Join("manifest/sql", "mock-data") {
		t.Fatalf("expected mock sql dir %q, got %q", path.Join("manifest/sql", "mock-data"), got)
	}
}

// TestResolveSource verifies the command source selection is explicit and
// defaults to embedded assets for runtime execution.
func TestResolveSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Source
		wantErr bool
	}{
		{name: "default embedded", input: "", want: SourceEmbedded},
		{name: "explicit embedded", input: "embedded", want: SourceEmbedded},
		{name: "explicit local", input: "local", want: SourceLocal},
		{name: "reject unknown", input: "filesystem", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveSource(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve source: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// TestExecuteWithExecutorStopsAfterFirstError verifies execution halts at the
// first failing SQL asset and returns the failing file name.
func TestExecuteWithExecutorStopsAfterFirstError(t *testing.T) {
	t.Parallel()

	assets := []Asset{
		{Path: "manifest/sql/001-first.sql", Content: "FIRST"},
		{Path: "manifest/sql/002-second.sql", Content: "SECOND"},
		{Path: "manifest/sql/003-third.sql", Content: "THIRD"},
	}

	var executedSQL []string
	err := ExecuteWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
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

// TestExecuteWithExecutorSkipsEmptyFiles verifies blank SQL assets are ignored
// while non-empty assets still execute in order.
func TestExecuteWithExecutorSkipsEmptyFiles(t *testing.T) {
	t.Parallel()

	assets := []Asset{
		{Path: "manifest/sql/001-empty.sql", Content: ""},
		{Path: "manifest/sql/002-seed.sql", Content: "SEED"},
	}

	var executedSQL []string
	err := ExecuteWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
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

// TestScanLocalSortsFiles verifies development-mode local SQL loading keeps
// lexical order.
func TestScanLocalSortsFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sqlDir := filepath.Join(tempDir, "manifest", "sql")
	writeTestSQLFile(t, filepath.Join(sqlDir, "010-third.sql"), "THIRD")
	writeTestSQLFile(t, filepath.Join(sqlDir, "001-first.sql"), "FIRST")
	writeTestSQLFile(t, filepath.Join(sqlDir, "002-second.sql"), "SECOND")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err = os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()

	assets, err := scanLocal(context.Background(), HostInitDir())
	if err != nil {
		t.Fatalf("scan local sql assets: %v", err)
	}
	got := []string{assets[0].Content, assets[1].Content, assets[2].Content}
	if !reflect.DeepEqual(got, []string{"FIRST", "SECOND", "THIRD"}) {
		t.Fatalf("expected ordered contents %v, got %v", []string{"FIRST", "SECOND", "THIRD"}, got)
	}
}

// TestScanEmbeddedReadsPreparedFiles verifies runtime-mode SQL loading reads
// packaged manifest assets from the embedded filesystem.
func TestScanEmbeddedReadsPreparedFiles(t *testing.T) {
	t.Parallel()

	assets, err := scanEmbedded(context.Background(), HostInitDir())
	if err != nil {
		t.Fatalf("scan embedded sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded init sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "001-user-auth-bootstrap.sql") {
		t.Fatalf("expected first embedded sql asset %q, got %q", path.Join("manifest/sql", "001-user-auth-bootstrap.sql"), assets[0].Path)
	}
}

// TestInitRuntimeDefaultUsesEmbeddedAssets verifies runtime `lina init`
// behavior defaults to the embedded manifest SQL assets.
func TestInitRuntimeDefaultUsesEmbeddedAssets(t *testing.T) {
	t.Parallel()

	source, err := ResolveSource("")
	if err != nil {
		t.Fatalf("resolve default init source: %v", err)
	}

	assets, err := ScanInit(context.Background(), source)
	if err != nil {
		t.Fatalf("scan init sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded init sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "001-user-auth-bootstrap.sql") {
		t.Fatalf("expected first embedded init sql asset %q, got %q", path.Join("manifest/sql", "001-user-auth-bootstrap.sql"), assets[0].Path)
	}
}

// TestMockRuntimeDefaultUsesEmbeddedAssets verifies runtime `lina mock`
// behavior defaults to the embedded mock-data SQL assets.
func TestMockRuntimeDefaultUsesEmbeddedAssets(t *testing.T) {
	t.Parallel()

	source, err := ResolveSource("")
	if err != nil {
		t.Fatalf("resolve default mock source: %v", err)
	}

	assets, err := ScanMock(context.Background(), source)
	if err != nil {
		t.Fatalf("scan mock sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded mock sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "mock-data", "001-users.sql") {
		t.Fatalf(
			"expected first embedded mock sql asset %q, got %q",
			path.Join("manifest/sql", "mock-data", "001-users.sql"),
			assets[0].Path,
		)
	}
}

// writeTestSQLFile writes one temporary SQL file for shared command helper tests.
func writeTestSQLFile(t *testing.T, path string, contents string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
