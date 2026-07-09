// This file verifies the published plugin business-context context helpers.

package bizctxcap

import (
	"context"
	"testing"
)

// TestWithCurrentContextProvidesPluginVisibleSnapshot verifies source-plugin
// tests and adapters can inject context without importing host internal types.
func TestWithCurrentContextProvidesPluginVisibleSnapshot(t *testing.T) {
	ctx := WithCurrentContext(context.Background(), CurrentContext{
		UserID:          3,
		TenantID:        12,
		ActingUserID:    7,
		ActingAsTenant:  true,
		IsImpersonation: true,
		Permissions:     []string{"system:user:list"},
	})

	current := CurrentFromContext(ctx)
	if current.UserID != 3 || current.TenantID != 12 || current.ActingUserID != 7 {
		t.Fatalf("expected injected context snapshot, got %+v", current)
	}
	if !current.ActingAsTenant || !current.IsImpersonation || current.PlatformBypass {
		t.Fatalf("expected tenant impersonation snapshot, got %+v", current)
	}
	if len(current.Permissions) != 1 || current.Permissions[0] != "system:user:list" {
		t.Fatalf("expected cloned permissions, got %+v", current.Permissions)
	}
}

// TestWithCurrentContextMarksPlatformBypass verifies platform-scope helper
// semantics remain available without a public service implementation.
func TestWithCurrentContextMarksPlatformBypass(t *testing.T) {
	current := CurrentFromContext(WithCurrentContext(context.Background(), CurrentContext{TenantID: 0}))
	if !current.PlatformBypass {
		t.Fatalf("expected platform bypass for platform tenant, got %+v", current)
	}
}

// TestCurrentFromContextClonesPermissions verifies callers cannot mutate the
// stored context snapshot through a returned permission slice.
func TestCurrentFromContextClonesPermissions(t *testing.T) {
	ctx := WithCurrentContext(context.Background(), CurrentContext{Permissions: []string{"one"}})
	first := CurrentFromContext(ctx)
	first.Permissions[0] = "mutated"

	second := CurrentFromContext(ctx)
	if second.Permissions[0] != "one" {
		t.Fatalf("expected context permissions to remain immutable, got %+v", second.Permissions)
	}
}
