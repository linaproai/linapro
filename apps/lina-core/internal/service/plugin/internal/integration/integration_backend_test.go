// This file covers generic plugin backend resource projection helpers.

package integration

import (
	"testing"

	"lina-core/internal/service/datascope"
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

// TestResolvePluginResourceDataScopeModeUsesHostScopeValues ensures plugin
// resource queries do not drift from the host role data-scope enum.
func TestResolvePluginResourceDataScopeModeUsesHostScopeValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		scope datascope.Scope
		want  pluginResourceDataScopeMode
	}{
		{name: "none denies rows", scope: datascope.ScopeNone, want: pluginResourceDataScopeDeny},
		{name: "all grants rows", scope: datascope.ScopeAll, want: pluginResourceDataScopeAll},
		{name: "tenant grants tenant-bound rows", scope: datascope.ScopeTenant, want: pluginResourceDataScopeAll},
		{name: "department uses dept filter", scope: datascope.ScopeDept, want: pluginResourceDataScopeDept},
		{name: "self uses user filter", scope: datascope.ScopeSelf, want: pluginResourceDataScopeSelf},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolvePluginResourceDataScopeMode(int(tc.scope))
			if got != tc.want {
				t.Fatalf("expected mode %d for scope %d, got %d", tc.want, tc.scope, got)
			}
		})
	}

	got := resolvePluginResourceDataScopeMode(99)
	if got != pluginResourceDataScopeDeny {
		t.Fatalf("expected unknown data-scope to deny rows, got %d", got)
	}
}
