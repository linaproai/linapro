// This file tests plugin-list controller helpers that project startup
// auto-enable configuration into management-list view fields.

package plugin

import "testing"

// TestBuildAutoEnableManagedSetNormalizesPluginIDs verifies startup-managed
// plugin IDs are trimmed and de-duplicated before list projection uses them.
func TestBuildAutoEnableManagedSetNormalizesPluginIDs(t *testing.T) {
	managedSet := buildAutoEnableManagedSet([]string{
		" plugin-demo-source ",
		"plugin-demo-source",
		"",
		"monitor-server",
	})

	if len(managedSet) != 2 {
		t.Fatalf("expected 2 unique managed plugin IDs, got %d", len(managedSet))
	}
	if !managedSet["plugin-demo-source"] {
		t.Fatal("expected plugin-demo-source to be marked as managed")
	}
	if !managedSet["monitor-server"] {
		t.Fatal("expected monitor-server to be marked as managed")
	}
	if managedSet[""] {
		t.Fatal("expected blank plugin ID to be ignored")
	}
}
