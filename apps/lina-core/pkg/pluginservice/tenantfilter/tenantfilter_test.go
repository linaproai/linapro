// This file verifies shared tenant filter behavior for source plugins.

package tenantfilter

import (
	"context"
	"testing"

	"lina-core/internal/model"
	internalbizctx "lina-core/internal/service/bizctx"
)

// TestCurrentReadsTenantIDFromBizContext verifies tenant ID resolution from bizctx.
func TestCurrentReadsTenantIDFromBizContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{TenantId: 42})
	if got := Current(ctx); got != 42 {
		t.Fatalf("expected tenant 42, got %d", got)
	}
}

// TestCurrentDefaultsToPlatform verifies missing bizctx remains the platform tenant.
func TestCurrentDefaultsToPlatform(t *testing.T) {
	if got := Current(context.Background()); got != 0 {
		t.Fatalf("expected platform tenant, got %d", got)
	}
}

// TestCurrentContextLeavesOnBehalfEmptyForRegularTenant verifies regular tenant
// requests do not persist impersonation-only audit fields.
func TestCurrentContextLeavesOnBehalfEmptyForRegularTenant(t *testing.T) {
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{
		UserId:   88,
		TenantId: 7,
	})

	current := CurrentContext(ctx)
	if current.TenantID != 7 || current.OnBehalfOfTenantID != 0 {
		t.Fatalf("expected regular tenant context, got %#v", current)
	}
	if current.ActingUserID != 88 || current.IsImpersonation || current.ActingAsTenant {
		t.Fatalf("expected regular actor metadata, got %#v", current)
	}
}

// TestCurrentContextReadsImpersonationFields verifies platform impersonation
// records the target tenant as the on-behalf-of tenant.
func TestCurrentContextReadsImpersonationFields(t *testing.T) {
	ctx := context.WithValue(context.Background(), internalbizctx.ContextKey, &model.Context{
		UserId:          88,
		TenantId:        7,
		ActingUserId:    99,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	current := CurrentContext(ctx)
	if current.TenantID != 7 || current.OnBehalfOfTenantID != 7 {
		t.Fatalf("expected tenant fields from context, got %#v", current)
	}
	if current.ActingUserID != 99 || !current.IsImpersonation || !current.ActingAsTenant {
		t.Fatalf("expected impersonation fields from context, got %#v", current)
	}
}
