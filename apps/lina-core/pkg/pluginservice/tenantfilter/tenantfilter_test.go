// This file verifies shared tenant filter behavior for source plugins.

package tenantfilter

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	pluginbizctx "lina-core/pkg/pluginservice/bizctx"
	"lina-core/pkg/pluginservice/contract"
)

// TestCurrentReadsTenantIDFromBizContext verifies tenant ID resolution from bizctx.
func TestCurrentReadsTenantIDFromBizContext(t *testing.T) {
	service := newTenantFilterForTest(nil)
	ctx := contract.WithCurrentContext(context.Background(), contract.CurrentContext{TenantID: 42})
	if got := service.Current(ctx); got != 42 {
		t.Fatalf("expected tenant 42, got %d", got)
	}
}

// TestCurrentDefaultsToPlatform verifies missing bizctx remains the platform tenant.
func TestCurrentDefaultsToPlatform(t *testing.T) {
	service := newTenantFilterForTest(nil)
	if got := service.Current(context.Background()); got != 0 {
		t.Fatalf("expected platform tenant, got %d", got)
	}
}

// TestCurrentContextLeavesOnBehalfEmptyForRegularTenant verifies regular tenant
// requests do not persist impersonation-only audit fields.
func TestCurrentContextLeavesOnBehalfEmptyForRegularTenant(t *testing.T) {
	service := newTenantFilterForTest(nil)
	ctx := contract.WithCurrentContext(context.Background(), contract.CurrentContext{
		UserID:   88,
		TenantID: 7,
	})

	current := service.CurrentContext(ctx)
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
	service := newTenantFilterForTest(nil)
	ctx := contract.WithCurrentContext(context.Background(), contract.CurrentContext{
		UserID:          88,
		TenantID:        7,
		ActingUserID:    99,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	current := service.CurrentContext(ctx)
	if current.TenantID != 7 || current.OnBehalfOfTenantID != 7 {
		t.Fatalf("expected tenant fields from context, got %#v", current)
	}
	if current.ActingUserID != 99 || !current.IsImpersonation || !current.ActingAsTenant {
		t.Fatalf("expected impersonation fields from context, got %#v", current)
	}
}

// TestApplyColumnUsesHostPlatformBypass verifies platform bypass policy can
// keep plugin-owned table queries unrestricted for platform operators.
func TestApplyColumnUsesHostPlatformBypass(t *testing.T) {
	service := newTenantFilterForTest(testPlatformBypassEvaluator{bypass: true})
	ctx := contract.WithCurrentContext(context.Background(), contract.CurrentContext{TenantID: 7})

	sql := buildTenantFilterSQL(t, ctx, service.Apply(ctx, g.DB().Model("plugin_record")))
	if strings.Contains(sql, contract.TenantFilterColumn) {
		t.Fatalf("expected platform bypass to skip tenant predicate, got SQL %q", sql)
	}
}

// TestApplyColumnUsesCurrentTenant verifies regular tenant requests constrain
// plugin-owned table queries by the current tenant discriminator.
func TestApplyColumnUsesCurrentTenant(t *testing.T) {
	service := newTenantFilterForTest(testPlatformBypassEvaluator{bypass: false})
	ctx := contract.WithCurrentContext(context.Background(), contract.CurrentContext{TenantID: 7})

	sql := buildTenantFilterSQL(t, ctx, service.Apply(ctx, g.DB().Model("plugin_record")))
	if !strings.Contains(sql, contract.TenantFilterColumn) || !strings.Contains(sql, "=7") {
		t.Fatalf("expected tenant predicate in SQL, got %q", sql)
	}
}

// newTenantFilterForTest creates an explicitly injected tenant filter service.
func newTenantFilterForTest(bypassEvaluator contract.PlatformBypassEvaluator) contract.TenantFilterService {
	service, err := New(pluginbizctx.New(nil), bypassEvaluator)
	if err != nil {
		panic(err)
	}
	return service
}

// buildTenantFilterSQL renders one model query into SQL for predicate assertions.
func buildTenantFilterSQL(t *testing.T, ctx context.Context, model *gdb.Model) string {
	t.Helper()

	sql, err := gdb.ToSQL(ctx, func(sqlCtx context.Context) error {
		_, queryErr := model.Ctx(sqlCtx).Count()
		return queryErr
	})
	if err != nil {
		t.Fatalf("build tenant filter SQL failed: %v", err)
	}
	return sql
}

// testPlatformBypassEvaluator returns a fixed host platform-bypass decision.
type testPlatformBypassEvaluator struct {
	bypass bool
}

// PlatformBypass returns the configured test bypass decision.
func (e testPlatformBypassEvaluator) PlatformBypass(context.Context) bool {
	return e.bypass
}
