// This file verifies best-effort host data-table comment lookup behavior.

package plugin

import (
	"reflect"
	"testing"

	"lina-core/pkg/dialect"
)

// TestNormalizeDataTableNamesTrimsAndDeduplicates verifies metadata lookups
// query each non-empty table name at most once.
func TestNormalizeDataTableNamesTrimsAndDeduplicates(t *testing.T) {
	got := normalizeDataTableNames([]string{" sys_plugin ", "", "sys_user", "sys_plugin", "  "})
	want := []string{"sys_plugin", "sys_user"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected normalized table names %v, got %v", want, got)
	}
}

// TestDataTableCommentsFromMetadataMapsNonBlankComments verifies dialect
// metadata results are converted into the governance display map.
func TestDataTableCommentsFromMetadataMapsNonBlankComments(t *testing.T) {
	got := dataTableCommentsFromMetadata([]dialect.TableMeta{
		{TableName: " sys_plugin ", TableComment: " Plugin registry "},
		{TableName: "sys_user", TableComment: ""},
		{TableName: "", TableComment: "ignored"},
	})
	want := map[string]string{"sys_plugin": "Plugin registry"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected metadata comments %v, got %v", want, got)
	}
}
