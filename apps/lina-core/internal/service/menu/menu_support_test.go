// This file verifies menu support rules for localization and platform-context
// checks around global menu-governance mutations.

package menu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// menuTestTranslator stubs the menu translation dependency.
type menuTestTranslator struct {
	i18nsvc.Service

	values map[string]string
}

// Translate returns a configured translation or the caller fallback.
func (t menuTestTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := t.values[key]; ok {
		return value
	}
	return fallback
}

// TestLocalizeMenuEntityUsesMenuKey verifies menu_key remains the preferred translation anchor.
func TestLocalizeMenuEntityUsesMenuKey(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: menuTestTranslator{
			values: map[string]string{
				"menu.dashboard.title": "Dashboard",
			},
		},
	}
	menu := &entity.SysMenu{
		MenuKey: "dashboard",
		Name:    "仪表盘",
	}

	svc.localizeMenuEntity(context.Background(), menu)

	if menu.Name != "Dashboard" {
		t.Fatalf("expected menu name to be localized by menu key, got %q", menu.Name)
	}
}

// TestLocalizeMenuEntityKeepsLiteralName verifies literal non-key names stay unchanged.
func TestLocalizeMenuEntityKeepsLiteralName(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: menuTestTranslator{
			values: map[string]string{
				"menu.dashboard.title": "Dashboard",
			},
		},
	}
	menu := &entity.SysMenu{
		Name: "Custom Menu",
	}

	svc.localizeMenuEntity(context.Background(), menu)

	if menu.Name != "Custom Menu" {
		t.Fatalf("expected literal menu name to remain unchanged, got %q", menu.Name)
	}
}

// TestEnsurePlatformMenuGovernanceAllowsSingleTenantMode verifies disabled
// tenancy keeps the host menu service usable as a platform-only deployment.
func TestEnsurePlatformMenuGovernanceAllowsSingleTenantMode(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuard{enabled: false})
	if err != nil {
		t.Fatalf("expected disabled tenancy to allow menu governance, got %v", err)
	}
}

// TestEnsurePlatformMenuGovernanceRejectsTenantContext verifies active
// multi-tenancy requires a platform all-data context for sys_menu writes.
func TestEnsurePlatformMenuGovernanceRejectsTenantContext(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuard{enabled: true, platformBypass: false})
	if !bizerr.Is(err, tenantcap.CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestEnsurePlatformMenuGovernanceAllowsPlatformBypass verifies platform
// all-data context can mutate the global menu topology.
func TestEnsurePlatformMenuGovernanceAllowsPlatformBypass(t *testing.T) {
	err := ensurePlatformMenuGovernanceContext(context.Background(), menuTenantGuard{enabled: true, platformBypass: true})
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

// menuTenantGuard is the tenantspi fake needed by menu platform-guard tests.
type menuTenantGuard struct {
	tenantspi.Service

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

// newMenuPlatformGuardTenantService creates a real tenant capability with one
// enabled test provider so menu mutation tests cover service entry points.
func newMenuPlatformGuardTenantService(t *testing.T) tenantspi.Service {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-menu-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return menuPlatformGuardProvider{}, nil
	}); err != nil {
		t.Fatalf("register menu tenant provider: %v", err)
	}
	return tenantspi.New(manager, menuPlatformGuardProviderRuntime{pluginID: providerPluginID}, nil, bizctx.New())
}

// menuPlatformGuardProviderRuntime marks exactly one test provider plugin enabled.
type menuPlatformGuardProviderRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r menuPlatformGuardProviderRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
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
