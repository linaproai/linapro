// This file verifies tenantcap service defaults and provider-backed behavior.

package tenantcap

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	_ "lina-core/pkg/dbdriver"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// tenantcapTestEnablement reports the multi-tenant plugin as enabled or disabled.
type tenantcapTestEnablement struct {
	enabled bool
}

// IsEnabled returns the configured enablement state.
func (e tenantcapTestEnablement) IsEnabled(_ context.Context, pluginID string) bool {
	return e.enabled && pluginID == pkgtenantcap.ProviderPluginID
}

// tenantcapTestProvider records provider calls for service delegation tests.
type tenantcapTestProvider struct {
	resolveResult *pkgtenantcap.ResolverResult
	resolveCalls  int
}

// ResolveTenant returns the configured resolver result.
func (p *tenantcapTestProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	p.resolveCalls++
	return p.resolveResult, nil
}

// ValidateUserInTenant accepts all users for provider interface completeness.
func (p *tenantcapTestProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns no tenant rows for provider interface completeness.
func (p *tenantcapTestProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant accepts all switches for provider interface completeness.
func (p *tenantcapTestProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// TestDefaultServiceNoopsWithoutProvider verifies single-tenant fallback behavior.
func TestDefaultServiceNoopsWithoutProvider(t *testing.T) {
	pkgtenantcap.RegisterProvider(nil)
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	service := New(tenantcapTestEnablement{enabled: false})
	ctx := context.Background()

	if service.Enabled(ctx) {
		t.Fatal("expected tenantcap to be disabled without provider")
	}
	if current := service.Current(ctx); current != pkgtenantcap.PLATFORM {
		t.Fatalf("expected platform tenant, got %d", current)
	}
	if service.PlatformBypass(ctx) {
		t.Fatal("expected no platform bypass without business context")
	}
}

// TestPlatformBypassRequiresPlatformContext verifies impersonation does not bypass tenant filters.
func TestPlatformBypassRequiresPlatformContext(t *testing.T) {
	service := New(tenantcapTestEnablement{enabled: true})

	platformCtx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{
		TenantId:  0,
		DataScope: 1,
	})
	if !service.PlatformBypass(platformCtx) {
		t.Fatal("expected platform role in platform context to bypass")
	}

	impersonationCtx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{
		TenantId:        10,
		DataScope:       1,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})
	if service.PlatformBypass(impersonationCtx) {
		t.Fatal("expected impersonation context to use tenant filtering")
	}
}

// TestApplyInjectsTenantPredicateWhenEnabled verifies tenant filtering is added
// only when provider-backed multi-tenancy is active and not bypassed.
func TestApplyInjectsTenantPredicateWhenEnabled(t *testing.T) {
	provider := &tenantcapTestProvider{}
	pkgtenantcap.RegisterProvider(provider)
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	service := New(tenantcapTestEnablement{enabled: true})
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{TenantId: 12})

	applied, err := service.Apply(ctx, g.DB().Model("sys_user"), "tenant_id")
	if err != nil {
		t.Fatalf("apply tenant filter failed: %v", err)
	}
	sql, err := gdb.ToSQL(ctx, func(sqlCtx context.Context) error {
		_, queryErr := applied.Ctx(sqlCtx).Count()
		return queryErr
	})
	if err != nil {
		t.Fatalf("build apply SQL failed: %v", err)
	}
	if !strings.Contains(sql, "tenant_id") || !strings.Contains(sql, "=12") {
		t.Fatalf("expected tenant predicate in SQL, got %q", sql)
	}
}

// TestApplySkipsTenantPredicateForPlatformBypass verifies platform operators
// keep unrestricted platform reads when not impersonating a tenant.
func TestApplySkipsTenantPredicateForPlatformBypass(t *testing.T) {
	provider := &tenantcapTestProvider{}
	pkgtenantcap.RegisterProvider(provider)
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	service := New(tenantcapTestEnablement{enabled: true})
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{
		TenantId:  int(pkgtenantcap.PLATFORM),
		DataScope: 1,
	})

	applied, err := service.Apply(ctx, g.DB().Model("sys_user"), "tenant_id")
	if err != nil {
		t.Fatalf("apply tenant filter failed: %v", err)
	}
	sql, err := gdb.ToSQL(ctx, func(sqlCtx context.Context) error {
		_, queryErr := applied.Ctx(sqlCtx).Count()
		return queryErr
	})
	if err != nil {
		t.Fatalf("build platform bypass SQL failed: %v", err)
	}
	if strings.Contains(sql, "tenant_id") {
		t.Fatalf("expected no tenant predicate for platform bypass, got SQL %q", sql)
	}
}

// TestResolveTenantDelegatesToProviderWhenEnabled verifies the service follows
// the provider resolution path after plugin enablement and provider registration.
func TestResolveTenantDelegatesToProviderWhenEnabled(t *testing.T) {
	provider := &tenantcapTestProvider{
		resolveResult: &pkgtenantcap.ResolverResult{TenantID: 16, Matched: true},
	}
	pkgtenantcap.RegisterProvider(provider)
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	service := New(tenantcapTestEnablement{enabled: true})
	result, err := service.ResolveTenant(context.Background(), &ghttp.Request{})
	if err != nil {
		t.Fatalf("resolve tenant failed: %v", err)
	}
	if result == nil || result.TenantID != 16 || !result.Matched {
		t.Fatalf("expected provider result, got %#v", result)
	}
	if provider.resolveCalls != 1 {
		t.Fatalf("expected one provider call, got %d", provider.resolveCalls)
	}
}

// TestReadWithPlatformFallbackUsesTenantRowsFirst verifies tenant rows suppress platform fallback.
func TestReadWithPlatformFallbackUsesTenantRowsFirst(t *testing.T) {
	service := New(tenantcapTestEnablement{enabled: true})
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{TenantId: 7})

	var seen []TenantID
	items, err := service.ReadWithPlatformFallback(ctx, func(_ context.Context, tenantID TenantID) ([]any, error) {
		seen = append(seen, tenantID)
		if tenantID == 7 {
			return []any{"tenant"}, nil
		}
		return []any{"platform"}, nil
	})
	if err != nil {
		t.Fatalf("fallback read failed: %v", err)
	}
	if !reflect.DeepEqual(seen, []TenantID{7}) {
		t.Fatalf("expected tenant-only read, got %#v", seen)
	}
	if !reflect.DeepEqual(items, []any{"tenant"}) {
		t.Fatalf("expected tenant rows, got %#v", items)
	}
}

// TestReadWithPlatformFallbackUsesPlatformWhenTenantEmpty verifies empty tenant rows fall back to platform.
func TestReadWithPlatformFallbackUsesPlatformWhenTenantEmpty(t *testing.T) {
	service := New(tenantcapTestEnablement{enabled: true})
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{TenantId: 8})

	var seen []TenantID
	items, err := service.ReadWithPlatformFallback(ctx, func(_ context.Context, tenantID TenantID) ([]any, error) {
		seen = append(seen, tenantID)
		if tenantID == pkgtenantcap.PLATFORM {
			return []any{"platform"}, nil
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("fallback read failed: %v", err)
	}
	if !reflect.DeepEqual(seen, []TenantID{8, pkgtenantcap.PLATFORM}) {
		t.Fatalf("expected tenant then platform reads, got %#v", seen)
	}
	if !reflect.DeepEqual(items, []any{"platform"}) {
		t.Fatalf("expected platform rows, got %#v", items)
	}
}
