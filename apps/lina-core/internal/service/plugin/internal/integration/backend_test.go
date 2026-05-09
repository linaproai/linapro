// This file covers generic plugin backend resource projection helpers.

package integration

import (
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
)

// TestResolvePluginResourceRecordValueFallsBackToColumnName verifies resource
// rows remain usable when a driver returns physical column keys.
func TestResolvePluginResourceRecordValueFallsBackToColumnName(t *testing.T) {
	t.Parallel()

	field := &catalog.ResourceField{Name: "userName", Column: "user_name"}
	value := resolvePluginResourceRecordValue(
		map[string]interface{}{
			"user_name": "admin",
		},
		field,
	)

	if value != "admin" {
		t.Fatalf("expected field value from physical column fallback, got %#v", value)
	}
}
