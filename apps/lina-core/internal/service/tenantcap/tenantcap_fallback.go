// This file implements platform fallback helpers for tenant-overridable reads.

package tenantcap

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	pkgtenantcap "lina-core/pkg/tenantcap"
)

// FallbackScanner loads records for one tenant-specific or platform-default model.
type FallbackScanner[T any] func(ctx context.Context, tenantID TenantID) ([]T, error)

// ReadWithPlatformFallback reads current-tenant records first and falls back to
// platform records when the tenant has no override rows.
func (s *serviceImpl) ReadWithPlatformFallback(ctx context.Context, scanner FallbackScanner[any]) ([]any, error) {
	if scanner == nil {
		return nil, nil
	}
	currentTenant := s.Current(ctx)
	items, err := scanner(ctx, currentTenant)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 || currentTenant == pkgtenantcap.PLATFORM {
		return items, nil
	}
	return scanner(ctx, pkgtenantcap.PLATFORM)
}

// ApplyPlatformFallbackWhere injects one platform-fallback tenant predicate
// for callers that need to compose fallback reads directly in SQL.
func ApplyPlatformFallbackWhere(model *gdb.Model, tenantColumn string, tenantID TenantID) *gdb.Model {
	if model == nil {
		return nil
	}
	return model.WhereIn(tenantColumn, []int{int(tenantID), int(pkgtenantcap.PLATFORM)})
}
