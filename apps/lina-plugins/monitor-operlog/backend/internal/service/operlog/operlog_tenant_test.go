// This file verifies tenant audit metadata resolution for operation logs.

package operlog

import (
	"context"
	"testing"

	"lina-core/pkg/pluginservice/bizctx"
)

// TestResolveAuditTenantContextReadsBizContext verifies tenant metadata comes from bizctx.
func TestResolveAuditTenantContextReadsBizContext(t *testing.T) {
	ctx := bizctx.WithCurrentContext(context.Background(), bizctx.CurrentContext{
		UserID:   3,
		TenantID: 12,
	})

	actual := resolveAuditTenantContext(ctx, nil, nil, nil, nil)
	if actual.TenantID != 12 || actual.OnBehalfOfTenantID != 0 {
		t.Fatalf("expected tenant 12, got %#v", actual)
	}
	if actual.ActingUserID != 3 || actual.IsImpersonation {
		t.Fatalf("expected regular tenant audit metadata, got %#v", actual)
	}
}

// TestResolveAuditTenantContextReadsImpersonation verifies impersonation metadata comes from bizctx.
func TestResolveAuditTenantContextReadsImpersonation(t *testing.T) {
	ctx := bizctx.WithCurrentContext(context.Background(), bizctx.CurrentContext{
		UserID:          10,
		TenantID:        12,
		ActingUserID:    3,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	actual := resolveAuditTenantContext(ctx, nil, nil, nil, nil)
	if actual.TenantID != 12 || actual.OnBehalfOfTenantID != 12 {
		t.Fatalf("expected impersonation tenant metadata, got %#v", actual)
	}
	if actual.ActingUserID != 3 || !actual.IsImpersonation {
		t.Fatalf("expected impersonation audit metadata, got %#v", actual)
	}
}

// TestResolveAuditTenantContextHonorsExplicitOverrides verifies callers can override context defaults.
func TestResolveAuditTenantContextHonorsExplicitOverrides(t *testing.T) {
	tenantID := 21
	actingUserID := 4
	onBehalfOfTenantID := 22
	isImpersonation := true

	actual := resolveAuditTenantContext(
		context.Background(),
		&tenantID,
		&actingUserID,
		&onBehalfOfTenantID,
		&isImpersonation,
	)
	if actual.TenantID != 21 || actual.ActingUserID != 4 || actual.OnBehalfOfTenantID != 22 || !actual.IsImpersonation {
		t.Fatalf("expected explicit overrides, got %#v", actual)
	}
}
