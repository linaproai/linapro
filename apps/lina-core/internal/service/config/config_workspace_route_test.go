// This file verifies the composeWorkspaceRouteURL helper that backs the
// BuildWorkspaceRouteURL service method so source-plugin OAuth callbacks
// reach the right /oauth-handoff URL across hash and history routers.

package config

import "testing"

// TestComposeWorkspaceRouteURL verifies the four production combinations of
// base path and router backend used by the OAuth handoff redirect, plus
// input-normalization edge cases (missing leading slash, accidental
// "?" prefix on the raw query).
func TestComposeWorkspaceRouteURL(t *testing.T) {
	cases := []struct {
		name       string
		basePath   string
		routerMode string
		routePath  string
		rawQuery   string
		expected   string
	}{
		{
			name:       "hash base admin with query",
			basePath:   "/admin",
			routerMode: WorkspaceRouterModeHash,
			routePath:  "/oauth-handoff",
			rawQuery:   "k=v",
			expected:   "/admin/#/oauth-handoff?k=v",
		},
		{
			name:       "hash base admin without query",
			basePath:   "/admin",
			routerMode: WorkspaceRouterModeHash,
			routePath:  "/oauth-handoff",
			rawQuery:   "",
			expected:   "/admin/#/oauth-handoff",
		},
		{
			name:       "hash root base",
			basePath:   "/",
			routerMode: WorkspaceRouterModeHash,
			routePath:  "/oauth-handoff",
			rawQuery:   "k=v",
			expected:   "/#/oauth-handoff?k=v",
		},
		{
			name:       "history base admin",
			basePath:   "/admin",
			routerMode: WorkspaceRouterModeHistory,
			routePath:  "/oauth-handoff",
			rawQuery:   "k=v",
			expected:   "/admin/oauth-handoff?k=v",
		},
		{
			name:       "history root base",
			basePath:   "/",
			routerMode: WorkspaceRouterModeHistory,
			routePath:  "/oauth-handoff",
			rawQuery:   "k=v",
			expected:   "/oauth-handoff?k=v",
		},
		{
			name:       "missing leading slash on route",
			basePath:   "/admin",
			routerMode: WorkspaceRouterModeHash,
			routePath:  "oauth-handoff",
			rawQuery:   "",
			expected:   "/admin/#/oauth-handoff",
		},
		{
			name:       "strips leading question mark on query",
			basePath:   "/admin",
			routerMode: WorkspaceRouterModeHash,
			routePath:  "/oauth-handoff",
			rawQuery:   "?k=v",
			expected:   "/admin/#/oauth-handoff?k=v",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := composeWorkspaceRouteURL(tc.basePath, tc.routerMode, tc.routePath, tc.rawQuery); got != tc.expected {
				t.Fatalf("composeWorkspaceRouteURL = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestMustNormalizeWorkspaceRouterModeDefaultsToHash verifies the helper
// falls back to hash mode for empty inputs.
func TestMustNormalizeWorkspaceRouterModeDefaultsToHash(t *testing.T) {
	if got := mustNormalizeWorkspaceRouterMode(""); got != WorkspaceRouterModeHash {
		t.Fatalf("default router mode = %q, want %q", got, WorkspaceRouterModeHash)
	}
}

// TestMustNormalizeWorkspaceRouterModeAcceptsHistory verifies the helper
// honors history backend when explicitly configured.
func TestMustNormalizeWorkspaceRouterModeAcceptsHistory(t *testing.T) {
	if got := mustNormalizeWorkspaceRouterMode("history"); got != WorkspaceRouterModeHistory {
		t.Fatalf("router mode = %q, want %q", got, WorkspaceRouterModeHistory)
	}
}

// TestMustNormalizeWorkspaceRouterModeRejectsInvalidValues verifies invalid
// router-mode values fail fast.
func TestMustNormalizeWorkspaceRouterModeRejectsInvalidValues(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected invalid router mode to panic")
		}
	}()
	_ = mustNormalizeWorkspaceRouterMode("memory")
}
