// This file verifies platform-owned tenant provisioning policy behavior.

package plugin

import (
	"context"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
)

// TestUpdateTenantProvisioningPolicySurvivesManifestSync verifies plugin.yaml
// synchronization does not overwrite the platform-owned provisioning policy.
func TestUpdateTenantProvisioningPolicySurvivesManifestSync(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-tenant-provisioning-policy"
		supports = true
		manifest = &catalog.Manifest{
			ID:                  pluginID,
			Name:                "Tenant Provisioning Policy",
			Version:             "v0.1.0",
			Type:                catalog.TypeSource.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		}
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("sync plugin manifest failed: %v", err)
	}
	if err := service.UpdateTenantProvisioningPolicy(ctx, pluginID, true); err != nil {
		t.Fatalf("enable tenant provisioning policy failed: %v", err)
	}
	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("resync plugin manifest failed: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("load plugin registry failed: %v", err)
	}
	if registry == nil || !registry.AutoEnableForNewTenants {
		t.Fatalf("expected policy to survive manifest sync, got %#v", registry)
	}
}

// TestUpdateTenantProvisioningPolicyRejectsGlobalPlugin verifies the policy only
// applies to tenant-aware tenant-scoped plugins.
func TestUpdateTenantProvisioningPolicyRejectsGlobalPlugin(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-tenant-provisioning-global"
		supports = true
		manifest = &catalog.Manifest{
			ID:                  pluginID,
			Name:                "Tenant Provisioning Global",
			Version:             "v0.1.0",
			Type:                catalog.TypeSource.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  catalog.InstallModeGlobal.String(),
		}
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("sync plugin manifest failed: %v", err)
	}

	err := service.UpdateTenantProvisioningPolicy(ctx, pluginID, true)
	if !bizerr.Is(err, CodePluginTenantProvisioningPolicyInvalid) {
		t.Fatalf("expected tenant provisioning policy validation error, got %v", err)
	}
}

// TestUpdateTenantProvisioningPolicyRejectsUnsupportedTenantGovernance verifies
// the policy stays disabled when a registry cannot use tenant-level install mode.
func TestUpdateTenantProvisioningPolicyRejectsUnsupportedTenantGovernance(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-tenant-provisioning-unsupported"
		supports = false
		manifest = &catalog.Manifest{
			ID:                  pluginID,
			Name:                "Tenant Provisioning Unsupported",
			Version:             "v0.1.0",
			Type:                catalog.TypeSource.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  catalog.InstallModeGlobal.String(),
		}
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifest.ManifestPath = filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifest.ManifestPath,
		"id: "+pluginID+"\nname: Tenant Provisioning Unsupported\nversion: v0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("sync plugin manifest failed: %v", err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{InstallMode: catalog.InstallModeTenantScoped.String()}).
		Update(); err != nil {
		t.Fatalf("prepare unsupported tenant-scoped registry failed: %v", err)
	}

	err := service.UpdateTenantProvisioningPolicy(ctx, pluginID, true)
	if !bizerr.Is(err, CodePluginTenantProvisioningPolicyInvalid) {
		t.Fatalf("expected tenant provisioning policy validation error, got %v", err)
	}
}
