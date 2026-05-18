// This file verifies tenant and impersonation fields in request business context.

package bizctx

import (
	"context"
	"testing"

	"lina-core/internal/model"
)

// TestSetTenantAndImpersonationMutateBusinessContext verifies tenant fields are request-scoped.
func TestSetTenantAndImpersonationMutateBusinessContext(t *testing.T) {
	service := New()
	businessCtx := &model.Context{}
	ctx := context.WithValue(context.Background(), ContextKey, businessCtx)

	service.SetTenant(ctx, 12)
	service.SetImpersonation(ctx, 1, 12, true, true)

	if businessCtx.TenantId != 12 || !businessCtx.ActingAsTenant || !businessCtx.IsImpersonation {
		t.Fatalf("expected tenant impersonation fields to be set, got %#v", businessCtx)
	}
	if businessCtx.ActingUserId != 1 {
		t.Fatalf("expected acting user field, got %#v", businessCtx)
	}
}
