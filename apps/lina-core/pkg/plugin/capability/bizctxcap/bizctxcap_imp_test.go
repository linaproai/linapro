// This file verifies the published plugin business-context bridge.

package bizctxcap

import (
	"context"
	"testing"
)

// TestWithCurrentContextProvidesPluginVisibleSnapshot verifies source-plugin
// tests and adapters can inject context without importing host internal types.
func TestWithCurrentContextProvidesPluginVisibleSnapshot(t *testing.T) {
	service := New(nil)
	ctx := WithCurrentContext(context.Background(), CurrentContext{
		UserID:          3,
		TenantID:        12,
		ActingUserID:    7,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	current := service.Current(ctx)
	if current.UserID != 3 || current.TenantID != 12 || current.ActingUserID != 7 {
		t.Fatalf("expected injected context snapshot, got %+v", current)
	}
	if !current.ActingAsTenant || !current.IsImpersonation || current.PlatformBypass {
		t.Fatalf("expected tenant impersonation snapshot, got %+v", current)
	}
}

// TestNewUsesProvider verifies plugin tests can inject a custom context provider.
func TestNewUsesProvider(t *testing.T) {
	service := New(staticContextProvider{current: CurrentContext{UserID: 9, TenantID: 4}})
	current := service.Current(context.Background())
	if current.UserID != 9 || current.TenantID != 4 {
		t.Fatalf("expected provider context, got %+v", current)
	}
}

// staticContextProvider provides one fixed context snapshot for tests.
type staticContextProvider struct {
	current CurrentContext
}

// Current returns the fixed test context snapshot.
func (p staticContextProvider) Current(context.Context) CurrentContext {
	return p.current
}
