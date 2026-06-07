// This file verifies source-plugin visible apidoc route-key helpers remain
// stable for dynamic-plugin route text resolution.

package apidoccap

import "testing"

// TestBuildDynamicOperationKeyUsesDottedPath verifies dynamic route keys
// preserve path segment boundaries without hard-coded API-version rules.
func TestBuildDynamicOperationKeyUsesDottedPath(t *testing.T) {
	key := BuildDynamicOperationKey("/x/linapro-demo-dynamic/interface/m1/backend-summary", "GET")
	if key != "plugins.linapro_demo_dynamic.paths.get.interface.m1.backend_summary" {
		t.Fatalf("expected dotted path key, got %s", key)
	}
}
