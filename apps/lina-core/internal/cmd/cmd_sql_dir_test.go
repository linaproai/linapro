// This file verifies the fixed host SQL directory conventions used by the init
// and mock commands.

package cmd

import (
	"testing"

	"github.com/gogf/gf/v2/os/gfile"
)

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
