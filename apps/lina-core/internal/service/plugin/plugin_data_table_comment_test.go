// This file verifies best-effort host data-table comment lookup behavior.

package plugin

import (
	"context"
	"reflect"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
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

// TestResolveDataTableCommentsReadsInformationSchema verifies comment lookup
// uses information_schema metadata successfully for a known host table.
func TestResolveDataTableCommentsReadsInformationSchema(t *testing.T) {
	comments := newTestService().ResolveDataTableComments(context.Background(), []string{" sys_plugin ", "sys_plugin"})
	if comments["sys_plugin"] == "" {
		t.Fatalf("expected sys_plugin table comment to be resolved, got %v", comments)
	}
}
