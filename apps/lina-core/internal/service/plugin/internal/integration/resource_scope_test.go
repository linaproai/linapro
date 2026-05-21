// This file verifies generic plugin resource data-scope mapping against the
// host role data-scope contract used by sys_role.data_scope.

package integration

import (
	"testing"

	"lina-core/internal/service/datascope"
)

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
