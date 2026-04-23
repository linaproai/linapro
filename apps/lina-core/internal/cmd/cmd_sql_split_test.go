// This file verifies SQL statement splitting and per-statement execution for
// development-only init/mock command helpers.

package cmd

import (
	"context"
	"errors"
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

	got := splitSQLStatements(content)
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

	got := splitSQLStatements(content)
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

	assets := []sqlAsset{
		{Path: "manifest/sql/001-seed.sql", Content: "FIRST;\nSECOND;\nTHIRD;"},
	}

	var executedSQL []string
	err := executeSQLAssetsWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
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

	assets := []sqlAsset{
		{Path: "manifest/sql/001-seed.sql", Content: "FIRST;\nSECOND;\nTHIRD;"},
		{Path: "manifest/sql/002-after.sql", Content: "FOURTH;"},
	}

	var executedSQL []string
	err := executeSQLAssetsWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
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
	if !strings.Contains(err.Error(), "第 2 条语句") {
		t.Fatalf("expected error %q to contain failing statement index", err.Error())
	}
	if !reflect.DeepEqual(executedSQL, []string{"FIRST", "SECOND"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"FIRST", "SECOND"}, executedSQL)
	}
}
