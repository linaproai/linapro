// This file verifies tenant-resolution middleware behavior.

package middleware

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/bizerr"
	tenantcap "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// tenancyTestTenantService provides deterministic tenantcap outcomes for middleware tests.
type tenancyTestTenantService struct {
	tenantspi.Service

	enabled       bool
	resolveResult *tenantcap.ResolverResult
	resolveErr    error
	resolveCalls  int
}

// Enabled returns the configured linapro-tenant-core enablement state.
func (s *tenancyTestTenantService) Available(context.Context) bool {
	return s.enabled
}

// PlatformBypass returns the fixed platform-governance state for middleware tests.
func (s *tenancyTestTenantService) PlatformBypass(context.Context) bool {
	return false
}

// ResolveTenant records one middleware resolver call and returns the configured result.
func (s *tenancyTestTenantService) ResolveTenant(context.Context, *ghttp.Request) (*tenantcap.ResolverResult, error) {
	s.resolveCalls++
	return s.resolveResult, s.resolveErr
}

// TestTenancyDisabledInjectsPlatformAndContinues verifies disabled tenancy
// short-circuits to platform context and calls the next handler.
func TestTenancyDisabledInjectsPlatformAndContinues(t *testing.T) {
	tenantSvc := &tenancyTestTenantService{enabled: false}
	status, body := runTenancyMiddlewareRequest(t, tenantSvc)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", status, body)
	}
	if strings.TrimSpace(body) != "tenant=0 acting=false impersonation=false actingUser=0" {
		t.Fatalf("expected platform context body, got %q", body)
	}
	if tenantSvc.resolveCalls != 0 {
		t.Fatalf("expected disabled tenancy not to resolve provider, got %d calls", tenantSvc.resolveCalls)
	}
}

// TestTenancyEnabledInjectsResolvedTenant verifies an enabled resolver result
// is injected into business context before the next handler runs.
func TestTenancyEnabledInjectsResolvedTenant(t *testing.T) {
	tenantSvc := &tenancyTestTenantService{
		enabled: true,
		resolveResult: &tenantcap.ResolverResult{
			TenantID:        42,
			Matched:         true,
			ActingAsTenant:  true,
			ActingUserID:    7,
			IsImpersonation: true,
		},
	}
	status, body := runTenancyMiddlewareRequest(t, tenantSvc)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", status, body)
	}
	if strings.TrimSpace(body) != "tenant=42 acting=true impersonation=true actingUser=7" {
		t.Fatalf("expected resolved tenant context body, got %q", body)
	}
	if tenantSvc.resolveCalls != 1 {
		t.Fatalf("expected one provider resolve call, got %d", tenantSvc.resolveCalls)
	}
}

// TestTenancyEnabledRequiresMatchedTenant verifies an exhausted resolver chain
// fails closed with an unauthorized response.
func TestTenancyEnabledRequiresMatchedTenant(t *testing.T) {
	tenantSvc := &tenancyTestTenantService{enabled: true}
	status, body := runTenancyMiddlewareRequest(t, tenantSvc)

	if status != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for missing tenant, got %d body=%s", status, body)
	}
	if strings.Contains(body, "tenant=") {
		t.Fatalf("expected next handler not to run, got body %q", body)
	}
}

// TestTenancyResolverTenantRequiredErrorUsesUnauthorized verifies provider
// tenant-required errors map to an authentication-style failure.
func TestTenancyResolverTenantRequiredErrorUsesUnauthorized(t *testing.T) {
	tenantSvc := &tenancyTestTenantService{
		enabled:    true,
		resolveErr: bizerr.NewCode(tenantcap.CodeTenantRequired),
	}
	status, body := runTenancyMiddlewareRequest(t, tenantSvc)

	if status != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for tenant-required error, got %d body=%s", status, body)
	}
}

// TestTenancyResolverForbiddenErrorUsesForbidden verifies provider authorization
// errors map to forbidden and do not continue the request chain.
func TestTenancyResolverForbiddenErrorUsesForbidden(t *testing.T) {
	tenantSvc := &tenancyTestTenantService{
		enabled:    true,
		resolveErr: bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", 9)),
	}
	status, body := runTenancyMiddlewareRequest(t, tenantSvc)

	if status != http.StatusForbidden {
		t.Fatalf("expected status 403 for forbidden tenant, got %d body=%s", status, body)
	}
}

// runTenancyMiddlewareRequest serves one request through Ctx and Tenancy and
// returns the observed response status and body.
func runTenancyMiddlewareRequest(t *testing.T, tenantSvc tenantspi.Service) (int, string) {
	t.Helper()

	svc := &serviceImpl{
		bizCtxSvc: bizctx.New(),
		configSvc: hostconfig.New(),
		i18nSvc:   i18nsvc.New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil)),
		tenantSvc: tenantSvc,
	}
	server := g.Server("middleware-tenancy-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(svc.Ctx, svc.Tenancy)
		group.GET("/tenancy", func(r *ghttp.Request) {
			businessCtx := ctxFromRequest(t, r)
			r.Response.Writef(
				"tenant=%d acting=%t impersonation=%t actingUser=%d",
				businessCtx.TenantId,
				businessCtx.ActingAsTenant,
				businessCtx.IsImpersonation,
				businessCtx.ActingUserId,
			)
		})
	})
	if err := server.Start(); err != nil {
		t.Fatalf("start tenancy middleware server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown tenancy middleware server: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)

	response, err := http.Get("http://" + server.GetListenedAddress() + "/tenancy")
	if err != nil {
		t.Fatalf("send tenancy middleware request: %v", err)
	}
	defer func() {
		if err = response.Body.Close(); err != nil {
			t.Fatalf("close tenancy middleware response: %v", err)
		}
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read tenancy middleware response: %v", err)
	}
	return response.StatusCode, string(body)
}

// ctxFromRequest returns the business context stored by middleware Ctx.
func ctxFromRequest(t *testing.T, r *ghttp.Request) *model.Context {
	t.Helper()

	value := r.Context().Value(gctx.StrKey("BizCtx"))
	businessCtx, ok := value.(*model.Context)
	if !ok || businessCtx == nil {
		t.Fatalf("expected business context in request")
	}
	return businessCtx
}
