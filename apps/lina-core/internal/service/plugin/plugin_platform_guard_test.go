// This file verifies platform-context enforcement for plugin governance entry
// points without depending on the full tenantcap service graph.

package plugin

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

// TestEnsurePlatformGovernanceAllowsSingleTenantMode verifies disabled tenancy
// keeps plugin governance available in platform-only deployments.
func TestEnsurePlatformGovernanceAllowsSingleTenantMode(t *testing.T) {
	err := ensurePlatformGovernanceContext(context.Background(), pluginTenantGuardHolder{tenantSvc: pluginTenantGuard{enabled: false}})
	if err != nil {
		t.Fatalf("expected disabled tenancy to allow plugin governance, got %v", err)
	}
}

// TestEnsurePlatformGovernanceRejectsTenantContext verifies active
// multi-tenancy requires platform all-data context for lifecycle writes.
func TestEnsurePlatformGovernanceRejectsTenantContext(t *testing.T) {
	err := ensurePlatformGovernanceContext(context.Background(), pluginTenantGuardHolder{tenantSvc: pluginTenantGuard{enabled: true, platformBypass: false}})
	if !bizerr.Is(err, pkgtenantcap.CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestEnsurePlatformGovernanceAllowsPlatformBypass verifies platform all-data
// context can perform plugin governance actions.
func TestEnsurePlatformGovernanceAllowsPlatformBypass(t *testing.T) {
	err := ensurePlatformGovernanceContext(context.Background(), pluginTenantGuardHolder{tenantSvc: pluginTenantGuard{enabled: true, platformBypass: true}})
	if err != nil {
		t.Fatalf("expected platform bypass to allow plugin governance, got %v", err)
	}
}

// TestPluginGovernanceMethodsRejectTenantContext verifies public platform
// plugin-governance methods fail before lifecycle, registry, package, or
// policy side effects when the caller is in tenant context.
func TestPluginGovernanceMethodsRejectTenantContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), bizctx.ContextKey, &model.Context{
		TenantId:  65001,
		DataScope: 1,
	})
	svc := newTestService()
	svc.SetTenantCapability(newPluginPlatformGuardTenantService(t))
	cases := []struct {
		name string
		run  func() error
	}{
		{name: "sync source plugins", run: func() error {
			return svc.SyncSourcePlugins(ctx)
		}},
		{name: "sync source plugins strict", run: func() error {
			_, err := svc.SyncSourcePluginsStrict(ctx)
			return err
		}},
		{name: "sync and list", run: func() error {
			_, err := svc.SyncAndList(ctx)
			return err
		}},
		{name: "install", run: func() error {
			_, err := svc.Install(ctx, "blocked-plugin", InstallOptions{})
			return err
		}},
		{name: "uninstall", run: func() error {
			return svc.Uninstall(ctx, "blocked-plugin", UninstallOptions{})
		}},
		{name: "update status", run: func() error {
			return svc.UpdateStatus(ctx, "blocked-plugin", 1, nil)
		}},
		{name: "enable", run: func() error {
			return svc.Enable(ctx, "blocked-plugin")
		}},
		{name: "disable", run: func() error {
			return svc.Disable(ctx, "blocked-plugin")
		}},
		{name: "upload dynamic package", run: func() error {
			_, err := svc.UploadDynamicPackage(ctx, nil)
			return err
		}},
		{name: "upgrade source plugin", run: func() error {
			_, err := svc.UpgradeSourcePlugin(ctx, "blocked-plugin")
			return err
		}},
		{name: "execute runtime upgrade", run: func() error {
			_, err := svc.ExecuteRuntimeUpgrade(ctx, "blocked-plugin", RuntimeUpgradeOptions{Confirmed: true})
			return err
		}},
		{name: "tenant provisioning policy", run: func() error {
			return svc.UpdateTenantProvisioningPolicy(ctx, "blocked-plugin", true)
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

// pluginTenantGuardHolder adapts a narrow tenant fake to the plugin guard helper.
type pluginTenantGuardHolder struct {
	tenantSvc pluginTenantGuard
}

// platformGovernanceTenantCapability returns the narrow test tenant capability.
func (h pluginTenantGuardHolder) platformGovernanceTenantCapability() platformGovernanceTenantCapability {
	return h.tenantSvc
}

// pluginTenantGuard is the narrow tenantcap fake needed by plugin platform-guard tests.
type pluginTenantGuard struct {
	enabled        bool
	platformBypass bool
}

// Enabled returns whether multi-tenancy is active in this test.
func (g pluginTenantGuard) Enabled(context.Context) bool {
	return g.enabled
}

// PlatformBypass returns whether the test context is platform all-data.
func (g pluginTenantGuard) PlatformBypass(context.Context) bool {
	return g.platformBypass
}

// newPluginPlatformGuardTenantService creates a real tenantcap service in
// active linapro-tenant-core mode for plugin facade entry-point tests.
func newPluginPlatformGuardTenantService(t *testing.T) tenantcapsvc.Service {
	t.Helper()
	previousProvider := pkgtenantcap.CurrentProvider()
	pkgtenantcap.RegisterProvider(pluginPlatformGuardProvider{})
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(previousProvider)
	})
	return tenantcapsvc.New(pluginPlatformGuardPluginState{}, bizctx.New())
}

// pluginPlatformGuardPluginState marks the linapro-tenant-core provider plugin enabled.
type pluginPlatformGuardPluginState struct{}

// IsEnabled reports the linapro-tenant-core provider plugin as enabled.
func (pluginPlatformGuardPluginState) IsEnabled(_ context.Context, pluginID string) bool {
	return pluginID == pkgtenantcap.ProviderPluginID
}

// pluginPlatformGuardProvider satisfies the tenantcap provider contract for
// tests that only need provider presence.
type pluginPlatformGuardProvider struct{}

// ResolveTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}
