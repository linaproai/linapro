// This file verifies menu governance mutations require platform context when
// multi-tenancy is active.

package menu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// TestEnsurePlatformMenuGovernanceAllowsSingleTenantMode verifies disabled
// tenancy keeps the host menu service usable as a platform-only deployment.
func TestEnsurePlatformMenuGovernanceAllowsSingleTenantMode(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuardHolder{tenantSvc: menuTenantGuard{enabled: false}})
	if err != nil {
		t.Fatalf("expected disabled tenancy to allow menu governance, got %v", err)
	}
}

// TestEnsurePlatformMenuGovernanceRejectsTenantContext verifies active
// multi-tenancy requires a platform all-data context for sys_menu writes.
func TestEnsurePlatformMenuGovernanceRejectsTenantContext(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuardHolder{tenantSvc: menuTenantGuard{enabled: true, platformBypass: false}})
	if !bizerr.Is(err, tenantcap.CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestEnsurePlatformMenuGovernanceAllowsPlatformBypass verifies platform
// all-data context can mutate the global menu topology.
func TestEnsurePlatformMenuGovernanceAllowsPlatformBypass(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuardHolder{tenantSvc: menuTenantGuard{enabled: true, platformBypass: true}})
	if err != nil {
		t.Fatalf("expected platform bypass to allow menu governance, got %v", err)
	}
}

// TestMenuMutationMethodsRejectTenantContext verifies public sys_menu mutation
// methods fail before writing global menu topology in active tenant contexts.
func TestMenuMutationMethodsRejectTenantContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), bizctx.ContextKey, &model.Context{
		TenantId:  64001,
		DataScope: 1,
	})
	svc := &serviceImpl{tenantSvc: newMenuPlatformGuardTenantService(t)}
	cases := []struct {
		name string
		run  func() error
	}{
		{name: "create", run: func() error {
			_, err := svc.Create(ctx, CreateInput{Name: "blocked menu", Type: "M", Visible: 1, Status: 1})
			return err
		}},
		{name: "update", run: func() error {
			return svc.Update(ctx, UpdateInput{Id: 1, Name: "blocked menu update"})
		}},
		{name: "delete", run: func() error {
			return svc.Delete(ctx, DeleteInput{Id: 1})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !bizerr.Is(err, tenantcap.CodePlatformPermissionRequired) {
				t.Fatalf("expected platform permission error, got %v", err)
			}
		})
	}
}

// menuTenantGuardHolder adapts a narrow tenant fake to the menu guard helper.
type menuTenantGuardHolder struct {
	tenantSvc menuTenantGuard
}

// platformMenuTenantService returns the narrow test tenant capability.
func (h menuTenantGuardHolder) platformMenuTenantService() platformMenuTenantService {
	return h.tenantSvc
}

// menuTenantGuard is the narrow tenantcap fake needed by menu platform-guard tests.
type menuTenantGuard struct {
	enabled        bool
	platformBypass bool
}

// Enabled returns whether multi-tenancy is active in this test.
func (g menuTenantGuard) Available(context.Context) bool {
	return g.enabled
}

// PlatformBypass returns whether the test context is platform all-data.
func (g menuTenantGuard) PlatformBypass(context.Context) bool {
	return g.platformBypass
}

// newMenuPlatformGuardTenantService creates a real tenantcap service with one
// enabled test provider so menu mutation tests cover service entry points.
func newMenuPlatformGuardTenantService(t *testing.T) tenantcap.Service {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-menu-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return menuPlatformGuardProvider{}, nil
	}); err != nil {
		t.Fatalf("register menu tenant provider: %v", err)
	}
	return tenantspi.New(manager, menuPlatformGuardProviderRuntime{pluginID: providerPluginID}, bizctx.New())
}

// menuPlatformGuardProviderRuntime marks exactly one test provider plugin enabled.
type menuPlatformGuardProviderRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r menuPlatformGuardProviderRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
}

// TenantProviderEnv returns an empty typed provider environment in menu tests.
func (menuPlatformGuardProviderRuntime) TenantProviderEnv(string) tenantspi.ProviderEnv {
	return tenantspi.ProviderEnv{}
}

// menuPlatformGuardProvider satisfies the tenantcap provider contract for
// tests that only need provider presence.
type menuPlatformGuardProvider struct{}

// ResolveTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ValidateUserInTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// ListUserTenants is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) SwitchTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}
