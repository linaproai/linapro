// This file verifies the published plugin business-context bridge.

package bizctx

import (
	"context"
	"testing"

	"lina-core/internal/model"
	internalbizctx "lina-core/internal/service/bizctx"
)

// TestServiceAdapterReturnsCurrentSnapshotFromHostContext verifies source
// plugins can observe host-authenticated tenant and impersonation metadata.
func TestServiceAdapterReturnsCurrentSnapshotFromHostContext(t *testing.T) {
	service := New()
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{
		UserId:          1,
		Username:        "admin",
		TenantId:        22,
		ActingUserId:    1,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	current := service.Current(ctx)
	if current.UserID != 1 || current.Username != "admin" {
		t.Fatalf("expected authenticated user metadata")
	}
	if current.TenantID != 22 || current.ActingUserID != 1 {
		t.Fatalf("expected tenant and acting user metadata")
	}
	if !current.ActingAsTenant || !current.IsImpersonation || current.PlatformBypass {
		t.Fatalf("expected impersonation metadata")
	}
}

// TestServiceAdapterReturnsCurrentSnapshotFromPlatformContext verifies platform
// context is exposed as a bypass snapshot.
func TestServiceAdapterReturnsCurrentSnapshotFromPlatformContext(t *testing.T) {
	service := New()
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{
		UserId:          7,
		Username:        "operator",
		TenantId:        0,
		ActingUserId:    9,
		ActingAsTenant:  false,
		IsImpersonation: true,
	})

	current := service.Current(ctx)
	if current.UserID != 7 || current.Username != "operator" {
		t.Fatalf("expected authenticated user metadata")
	}
	if current.TenantID != 0 || current.ActingUserID != 9 {
		t.Fatalf("expected tenant and acting user metadata")
	}
	if current.ActingAsTenant || !current.IsImpersonation || !current.PlatformBypass {
		t.Fatalf("expected platform impersonation metadata")
	}
}

// TestServiceAdapterReadsHostContextWithoutInternalService verifies explicit
// service test doubles still read the host context model directly.
func TestServiceAdapterReadsHostContextWithoutInternalService(t *testing.T) {
	service := &serviceAdapter{}
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{
		UserId:   8,
		Username: "fallback",
		TenantId: 3,
	})

	current := service.Current(ctx)
	if current.UserID != 8 || current.Username != "fallback" || current.TenantID != 3 {
		t.Fatalf("expected host context fallback")
	}
	if current.PlatformBypass {
		t.Fatalf("expected tenant context not to bypass platform filters")
	}
}

// TestServiceAdapterIgnoresUnexpectedContextValue verifies context reads reject
// plugin-local structs with host-like fields.
func TestServiceAdapterIgnoresUnexpectedContextValue(t *testing.T) {
	service := &serviceAdapter{}
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, struct {
		UserId   int
		Username string
		TenantId int
	}{
		UserId:   8,
		Username: "unexpected",
		TenantId: 3,
	})

	current := service.Current(ctx)
	if current != (CurrentContext{}) {
		t.Fatalf("expected zero context for unexpected context value, got %+v", current)
	}
}
