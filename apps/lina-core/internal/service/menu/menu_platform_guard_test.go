// This file verifies menu governance mutations require platform context when
// multi-tenancy is active.

package menu

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
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
	if !bizerr.Is(err, pkgtenantcap.CodePlatformPermissionRequired) {
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
			if err := tc.run(); !bizerr.Is(err, pkgtenantcap.CodePlatformPermissionRequired) {
				t.Fatalf("expected platform permission error, got %v", err)
			}
		})
	}
}

// menuTenantGuardHolder adapts a narrow tenant fake to the menu guard helper.
type menuTenantGuardHolder struct {
	tenantSvc menuTenantGuard
}

// platformMenuTenantCapability returns the narrow test tenant capability.
func (h menuTenantGuardHolder) platformMenuTenantCapability() platformMenuTenantCapability {
	return h.tenantSvc
}

// menuTenantGuard is the narrow tenantcap fake needed by menu platform-guard tests.
type menuTenantGuard struct {
	enabled        bool
	platformBypass bool
}

// Enabled returns whether multi-tenancy is active in this test.
func (g menuTenantGuard) Enabled(context.Context) bool {
	return g.enabled
}

// PlatformBypass returns whether the test context is platform all-data.
func (g menuTenantGuard) PlatformBypass(context.Context) bool {
	return g.platformBypass
}

// newMenuPlatformGuardTenantService creates a real tenantcap service in active
// linapro-tenant-core mode so menu mutation tests cover service entry points.
func newMenuPlatformGuardTenantService(t *testing.T) tenantcapsvc.Service {
	t.Helper()
	previousProvider := pkgtenantcap.CurrentProvider()
	pkgtenantcap.RegisterProvider(menuPlatformGuardProvider{})
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(previousProvider)
	})
	return tenantcapsvc.New(menuPlatformGuardPluginState{}, bizctx.New())
}

// menuPlatformGuardPluginState marks the linapro-tenant-core provider plugin enabled.
type menuPlatformGuardPluginState struct{}

// IsEnabled reports the linapro-tenant-core provider plugin as enabled.
func (menuPlatformGuardPluginState) IsEnabled(_ context.Context, pluginID string) bool {
	return pluginID == pkgtenantcap.ProviderPluginID
}

// menuPlatformGuardProvider satisfies the tenantcap provider contract for
// tests that only need provider presence.
type menuPlatformGuardProvider struct{}

// ResolveTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant is unused by menu platform-guard tests.
func (menuPlatformGuardProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}
