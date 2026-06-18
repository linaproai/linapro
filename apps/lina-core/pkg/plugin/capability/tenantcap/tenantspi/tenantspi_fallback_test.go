// This file verifies tenant capability fallback behavior when no provider is
// active. These checks keep optional multi-tenancy from causing host runtime
// 500s when tenant capability is disabled.

package tenantspi

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const tenantFallbackProviderID = "tenantcap-fallback-test-provider"

func TestDisabledTenantCapabilityReturnsPlatformFallbacks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil, tenantFallbackBizCtx{current: bizctxcap.CurrentContext{
		UserID:   10,
		TenantID: 22,
	}})

	if svc.Available(ctx) {
		t.Fatal("expected disabled tenant capability")
	}
	if svc.Available(ctx) {
		t.Fatal("expected unavailable tenant provider")
	}
	if status := svc.Status(ctx); status.Available || status.ActiveProvider != "" {
		t.Fatalf("expected unavailable status without active provider, got %#v", status)
	}
	if current := svc.Current(ctx); current != 22 {
		t.Fatalf("expected current tenant from business context, got %d", current)
	}

	model, err := svc.Apply(ctx, nil, "tenant_id")
	if err != nil {
		t.Fatalf("apply tenant scope returned error: %v", err)
	}
	if model != nil {
		t.Fatalf("expected nil model to remain unchanged, got %#v", model)
	}

	userModel, applied, err := svc.ApplyUserTenantScope(ctx, nil, "user_id")
	if err != nil {
		t.Fatalf("apply user tenant scope returned error: %v", err)
	}
	if userModel != nil || applied {
		t.Fatalf("expected no user tenant scope when disabled, got model=%#v applied=%v", userModel, applied)
	}

	filterModel, applied, err := svc.ApplyUserTenantFilter(ctx, nil, "user_id", 22)
	if err != nil {
		t.Fatalf("apply user tenant filter returned error: %v", err)
	}
	if filterModel != nil || applied {
		t.Fatalf("expected no user tenant filter when disabled, got model=%#v applied=%v", filterModel, applied)
	}

	resolved, err := svc.ResolveTenant(ctx, nil)
	if err != nil {
		t.Fatalf("resolve tenant returned error: %v", err)
	}
	if resolved == nil || !resolved.Matched || resolved.TenantID != tenantcap.PLATFORM {
		t.Fatalf("expected platform resolver fallback, got %#v", resolved)
	}

	if err = svc.EnsureTenantVisible(ctx, 99); err != nil {
		t.Fatalf("disabled tenant capability should not reject visibility checks: %v", err)
	}
	if err = svc.ValidateUserInTenant(ctx, 10, 99); err != nil {
		t.Fatalf("disabled tenant capability should not reject membership checks: %v", err)
	}
	if err = svc.SwitchTenant(ctx, 10, 99); err != nil {
		t.Fatalf("disabled tenant capability should not reject tenant switches: %v", err)
	}
	if err = svc.EnsureUsersInTenant(ctx, []int{10, 11}, 99); err != nil {
		t.Fatalf("disabled tenant capability should not reject bulk membership checks: %v", err)
	}
}

func TestDisabledTenantCapabilityReturnsNeutralMembershipAndStartupValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil, tenantFallbackBizCtx{current: bizctxcap.CurrentContext{TenantID: 22}})

	tenants, err := svc.ListUserTenants(ctx, 10)
	if err != nil {
		t.Fatalf("list user tenants returned error: %v", err)
	}
	if len(tenants) != 0 {
		t.Fatalf("expected empty tenant list, got %#v", tenants)
	}

	projections, err := svc.ListUserTenantProjections(ctx, []int{10, 11})
	if err != nil {
		t.Fatalf("list user tenant projections returned error: %v", err)
	}
	if len(projections) != 0 {
		t.Fatalf("expected empty tenant projections, got %#v", projections)
	}

	plan, err := svc.ResolveUserTenantAssignment(ctx, []tenantcap.TenantID{33}, tenantcap.UserTenantAssignmentCreate)
	if err != nil {
		t.Fatalf("resolve user tenant assignment returned error: %v", err)
	}
	if plan == nil || plan.PrimaryTenant != 22 || plan.ShouldReplace {
		t.Fatalf("expected current-tenant no-op assignment plan, got %#v", plan)
	}
	if err = svc.ReplaceUserTenantAssignments(ctx, 10, plan); err != nil {
		t.Fatalf("replace user tenant assignments should be a no-op when disabled: %v", err)
	}
	if failures, err := svc.ValidateUserMembershipStartupConsistency(ctx); err != nil || len(failures) != 0 {
		t.Fatalf("expected no startup consistency failures, got failures=%#v err=%v", failures, err)
	}
	if err = svc.ProvisionAutoEnabledTenantPlugins(ctx); err != nil {
		t.Fatalf("tenant plugin provisioning should be a no-op when disabled: %v", err)
	}
}

func TestTenantVisibilityUsesControlledRejectionWhenEnabled(t *testing.T) {
	ctx := context.Background()
	manager := newTenantFallbackManager(t)
	svc := &serviceImpl{
		manager: manager,
		runtime: tenantFallbackRuntime{enabledPluginID: tenantFallbackProviderID},
		bizCtxSvc: tenantFallbackBizCtx{current: bizctxcap.CurrentContext{
			UserID:   10,
			TenantID: 22,
		}},
	}

	err := svc.EnsureTenantVisible(ctx, 99)
	if !bizerr.Is(err, tenantcap.CodeTenantForbidden) {
		t.Fatalf("expected tenant forbidden bizerr, got %v", err)
	}
}

func TestTenantPlatformBypassAllowsVisibilityWhenEnabled(t *testing.T) {
	ctx := context.Background()
	manager := newTenantFallbackManager(t)
	svc := &serviceImpl{
		manager: manager,
		runtime: tenantFallbackRuntime{enabledPluginID: tenantFallbackProviderID},
		bizCtxSvc: tenantFallbackBizCtx{current: bizctxcap.CurrentContext{
			UserID:         10,
			TenantID:       0,
			PlatformBypass: true,
		}},
	}

	if err := svc.EnsureTenantVisible(ctx, 99); err != nil {
		t.Fatalf("platform bypass should allow tenant visibility checks: %v", err)
	}
}

func TestTenantCapabilityRejectsMultipleEnabledProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager()
	enabled := map[string]bool{
		"tenant-provider-conflict-a": true,
		"tenant-provider-conflict-b": true,
	}
	for pluginID := range enabled {
		pluginID := pluginID
		if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
			return nil, nil
		}); err != nil {
			t.Fatalf("register tenant provider %s: %v", pluginID, err)
		}
	}

	svc := New(manager, tenantConflictRuntime{enabled: enabled}, nil)
	_, err := svc.BatchGetTenants(ctx, []tenantcap.TenantID{tenantcap.PLATFORM})
	if !bizerr.Is(err, capmodel.CodeCapabilityProviderConflict) {
		t.Fatalf("expected provider conflict error, got %v", err)
	}
	status := svc.Status(ctx)
	if status.Available || status.Reason != "provider_conflict" {
		t.Fatalf("expected unavailable provider conflict status, got %#v", status)
	}
}

type tenantFallbackBizCtx struct {
	current bizctxcap.CurrentContext
}

func (s tenantFallbackBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

type tenantFallbackRuntime struct {
	enabledPluginID string
}

func (s tenantFallbackRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return s.enabledPluginID == pluginID
}

func (s tenantFallbackRuntime) TenantProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}

type tenantConflictRuntime struct {
	enabled map[string]bool
}

func (s tenantConflictRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return s.enabled[pluginID]
}

func (s tenantConflictRuntime) TenantProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}

func newTenantFallbackManager(t *testing.T) *Manager {
	t.Helper()

	manager := NewManager()
	if err := manager.RegisterFactory(
		tenantFallbackProviderID,
		func(context.Context, ProviderEnv) (Provider, error) {
			return tenantFallbackProvider{}, nil
		},
	); err != nil {
		t.Fatalf("register tenant fallback provider: %v", err)
	}
	return manager
}

type tenantFallbackProvider struct{}

func (tenantFallbackProvider) ResolveTenant(context.Context, *ghttp.Request) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

func (tenantFallbackProvider) ValidateUserInTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

func (tenantFallbackProvider) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{}, nil
}

func (tenantFallbackProvider) SwitchTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}
