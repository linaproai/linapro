// This file provides the host-service tenant scoping adapter used while the
// formal tenantcap service is registered by the multi-tenant capability.

package datascope

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/model"
	internalbizctx "lina-core/internal/service/bizctx"
)

const (
	// PlatformTenantID identifies the platform tenant and the single-tenant
	// fallback scope.
	PlatformTenantID = 0
	// TenantColumn is the conventional tenant discriminator column.
	TenantColumn = "tenant_id"
)

type tenantContextKey struct{}

// WithTenantScope returns a derived context whose tenant discriminator takes
// precedence over request bizctx for narrowly scoped internal reads.
func WithTenantScope(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// WithTenantForTest injects a tenant into context for service-layer tests until
// tenancy middleware writes the field into the shared bizctx model.
func WithTenantForTest(ctx context.Context, tenantID int) context.Context {
	return WithTenantScope(ctx, tenantID)
}

// CurrentTenantID resolves the tenant id from test context or future bizctx
// fields. Missing tenancy data intentionally falls back to PLATFORM so the
// default single-tenant installation stays unchanged.
func CurrentTenantID(ctx context.Context) int {
	if ctx == nil {
		return PlatformTenantID
	}
	if tenantID, ok := ctx.Value(tenantContextKey{}).(int); ok {
		return tenantID
	}
	return tenantIDFromBizContext(ctx)
}

// PlatformBypass reports whether the current context should skip tenant
// filtering. The temporary adapter treats tenant 0 as platform/single-tenant.
func PlatformBypass(ctx context.Context) bool {
	return CurrentTenantID(ctx) == PlatformTenantID
}

// ApplyTenantScope constrains a model by tenant id when a non-platform tenant
// is active. The caller supplies a qualified column when joins are present.
func ApplyTenantScope(ctx context.Context, model *gdb.Model, tenantColumn string) *gdb.Model {
	if model == nil || PlatformBypass(ctx) {
		return model
	}
	column := strings.TrimSpace(tenantColumn)
	if column == "" {
		column = TenantColumn
	}
	return model.Where(column, CurrentTenantID(ctx))
}

// CacheKey builds tenant-partitioned cache keys for host service caches.
func CacheKey(ctx context.Context, scope string, key string) string {
	return fmt.Sprintf("%s:tenant=%d:%s", strings.TrimSpace(scope), CurrentTenantID(ctx), strings.TrimSpace(key))
}

// tenantIDFromBizContext reads the shared business context tenant field.
func tenantIDFromBizContext(ctx context.Context) int {
	if bizCtx, ok := ctx.Value(internalbizctx.ContextKey).(*model.Context); ok && bizCtx != nil {
		return bizCtx.TenantId
	}
	return PlatformTenantID
}
