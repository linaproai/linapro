// This file verifies git status parsing keeps repository-relative paths intact.

package frameworkupgrade

import (
	"reflect"
	"testing"
)

// TestParseGitStatusPorcelain verifies porcelain output parsing keeps full paths.
func TestParseGitStatusPorcelain(t *testing.T) {
	t.Parallel()

	output := " M apps/lina-core/Makefile\n?? apps/lina-core/internal/service/frameworkupgrade/\nM  hack/tests/e2e/iam/role/TC0061-role-crud.ts\n"
	items := parseGitStatusPorcelain(output)
	expected := []string{
		"apps/lina-core/Makefile",
		"apps/lina-core/internal/service/frameworkupgrade/",
		"hack/tests/e2e/iam/role/TC0061-role-crud.ts",
	}
	if !reflect.DeepEqual(items, expected) {
		t.Fatalf("expected %v, got %v", expected, items)
	}
}
