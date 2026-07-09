// This file verifies plugin capability lifecycle adapter behavior remains
// stable when the contract and adapter live in the package entry file.

package plugincap

import (
	"context"
	"errors"
	"testing"
)

// TestLifecycleAdapterDelegatesToService verifies lifecycle calls are trimmed and delegated.
func TestLifecycleAdapterDelegatesToService(t *testing.T) {
	var (
		ctx       = context.Background()
		lifecycle = &recordingLifecycleService{}
		svc       = NewLifecycle(lifecycle)
	)

	if err := svc.EnsureTenantPluginDisableAllowed(ctx, " demo-plugin ", 11); err != nil {
		t.Fatalf("ensure tenant plugin disable: %v", err)
	}
	if lifecycle.disablePluginID != "demo-plugin" || lifecycle.disableTenantID != 11 {
		t.Fatalf("unexpected disable call plugin=%q tenant=%d", lifecycle.disablePluginID, lifecycle.disableTenantID)
	}

	svc.NotifyTenantPluginDisabled(ctx, " demo-plugin ", 12)
	if lifecycle.disabledPluginID != "demo-plugin" || lifecycle.disabledTenantID != 12 {
		t.Fatalf("unexpected disabled notification plugin=%q tenant=%d", lifecycle.disabledPluginID, lifecycle.disabledTenantID)
	}

	if err := svc.EnsureTenantDeleteAllowed(ctx, 13); err != nil {
		t.Fatalf("ensure tenant delete: %v", err)
	}
	if lifecycle.deleteTenantID != 13 {
		t.Fatalf("unexpected tenant delete ensure tenant=%d", lifecycle.deleteTenantID)
	}

	svc.NotifyTenantDeleted(ctx, 14)
	if lifecycle.deletedTenantID != 14 {
		t.Fatalf("unexpected tenant deleted notification tenant=%d", lifecycle.deletedTenantID)
	}
}

// TestLifecycleAdapterPropagatesServiceError verifies precondition errors are returned.
func TestLifecycleAdapterPropagatesServiceError(t *testing.T) {
	expectedErr := errors.New("blocked")
	svc := NewLifecycle(&recordingLifecycleService{disableErr: expectedErr})

	err := svc.EnsureTenantPluginDisableAllowed(context.Background(), "demo-plugin", 11)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected propagated error %v, got %v", expectedErr, err)
	}
}

// TestLifecycleAdapterAllowsMissingService verifies nil services keep no-op semantics.
func TestLifecycleAdapterAllowsMissingService(t *testing.T) {
	svc := NewLifecycle(nil)

	if err := svc.EnsureTenantPluginDisableAllowed(context.Background(), "demo-plugin", 11); err != nil {
		t.Fatalf("expected nil service ensure to succeed, got %v", err)
	}
	svc.NotifyTenantPluginDisabled(context.Background(), "demo-plugin", 11)
	if err := svc.EnsureTenantDeleteAllowed(context.Background(), 11); err != nil {
		t.Fatalf("expected nil service tenant delete ensure to succeed, got %v", err)
	}
	svc.NotifyTenantDeleted(context.Background(), 11)
}

// recordingLifecycleService records lifecycle calls for adapter tests.
type recordingLifecycleService struct {
	disablePluginID  string
	disableTenantID  int
	disabledPluginID string
	disabledTenantID int
	deleteTenantID   int
	deletedTenantID  int
	disableErr       error
}

// EnsureTenantPluginDisableAllowed records tenant plugin disable precondition calls.
func (r *recordingLifecycleService) EnsureTenantPluginDisableAllowed(_ context.Context, pluginID string, tenantID int) error {
	r.disablePluginID = pluginID
	r.disableTenantID = tenantID
	return r.disableErr
}

// NotifyTenantPluginDisabled records tenant plugin disable notifications.
func (r *recordingLifecycleService) NotifyTenantPluginDisabled(_ context.Context, pluginID string, tenantID int) {
	r.disabledPluginID = pluginID
	r.disabledTenantID = tenantID
}

// EnsureTenantDeleteAllowed records tenant delete precondition calls.
func (r *recordingLifecycleService) EnsureTenantDeleteAllowed(_ context.Context, tenantID int) error {
	r.deleteTenantID = tenantID
	return nil
}

// NotifyTenantDeleted records tenant delete notifications.
func (r *recordingLifecycleService) NotifyTenantDeleted(_ context.Context, tenantID int) {
	r.deletedTenantID = tenantID
}
