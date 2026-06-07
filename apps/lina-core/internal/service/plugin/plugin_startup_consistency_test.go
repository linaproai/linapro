// This file verifies startup consistency checks for plugin tenant governance.

package plugin

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const startupConsistencyMembershipTable = "plugin_linapro_tenant_core_user_membership"

// TestValidateStartupConsistencyRequiresInjectedTenantCapability verifies
// startup validation fails fast instead of building an implicit tenant service.
func TestValidateStartupConsistencyRequiresInjectedTenantCapability(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
	)
	service.SetTenantStartupCapability(nil)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "requires injected tenant capability service")
}

// TestValidateStartupConsistencyUsesInjectedTenantCapability verifies tenant
// membership checks run through the explicitly wired tenant capability.
func TestValidateStartupConsistencyUsesInjectedTenantCapability(t *testing.T) {
	var (
		service   = newTestService()
		ctx       = context.Background()
		tenantSvc = &startupConsistencyTenantCapability{details: []string{"injected tenant capability used"}}
	)
	service.SetTenantStartupCapability(tenantSvc)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "injected tenant capability used")
	if tenantSvc.calls != 1 {
		t.Fatalf("expected one injected tenant capability call, got %d", tenantSvc.calls)
	}
}

// TestValidateStartupConsistencyRejectsInvalidPluginGovernance verifies invalid
// scope_nature/install_mode combinations fail before serving requests.
func TestValidateStartupConsistencyRejectsInvalidPluginGovernance(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-startup-invalid-governance"
	)
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Startup Invalid Governance",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: "invalid_scope",
		InstallMode: catalog.InstallModeTenantScoped.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err,
		"plugin "+pluginID+" has invalid scope_nature invalid_scope",
	)
}

// TestValidateStartupConsistencyRejectsPlatformOnlyTenantScoped verifies
// platform-only plugins must remain globally installed.
func TestValidateStartupConsistencyRejectsPlatformOnlyTenantScoped(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-startup-platform-tenant-scoped"
	)
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Startup Platform Tenant Scoped",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeTenantScoped.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform_only plugin "+pluginID+" must use global install_mode")
}

// TestValidateStartupConsistencyRejectsPlatformUserMembership verifies
// platform users are not allowed to carry active tenant memberships.
func TestValidateStartupConsistencyRejectsPlatformUserMembership(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		username = "startup-platform-member"
		tenantID = 19001
	)
	cleanupStartupConsistencyPlugin(t, ctx, tenantcap.ProviderPluginID)
	cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID)
	service.SetTenantStartupCapability(&startupConsistencyTenantCapability{
		available:           true,
		validateMemberships: true,
	})
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, tenantcap.ProviderPluginID) })
	t.Cleanup(func() { cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    tenantcap.ProviderPluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})
	userID := insertStartupConsistencyUser(t, ctx, username, int(tenantcap.PLATFORM))
	insertStartupConsistencyTenantMembership(t, ctx, userID, tenantID, 1)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform user "+username)
}

// TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider
// verifies linapro-tenant-core enablement requires a registered tenantcap provider.
func TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = tenantcap.ProviderPluginID
	)
	service.SetTenantStartupCapability(&startupConsistencyTenantCapability{})
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "linapro-tenant-core plugin is enabled but capability tenant provider is not active")
}

// TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider verifies
// provider registration satisfies linapro-tenant-core startup consistency.
func TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = tenantcap.ProviderPluginID
	)
	service.SetTenantStartupCapability(&startupConsistencyTenantCapability{available: true})
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})

	if err := service.ValidateStartupConsistency(ctx); err != nil {
		t.Fatalf("expected registered provider to satisfy startup consistency, got %v", err)
	}
}

// startupConsistencyTenantCapability records startup membership validation calls.
type startupConsistencyTenantCapability struct {
	available           bool
	validateMemberships bool
	calls               int
	details             []string
}

// Available reports an active tenant capability for startup consistency tests.
func (s *startupConsistencyTenantCapability) Available(context.Context) bool {
	return s.available
}

// ValidateUserMembershipStartupConsistency records the injected startup check.
func (s *startupConsistencyTenantCapability) ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error) {
	s.calls++
	if s.validateMemberships {
		return validateStartupConsistencyTestMemberships(ctx)
	}
	return s.details, nil
}

// validateStartupConsistencyTestMemberships simulates the plugin-owned startup
// consistency check without making host production code know plugin tables.
func validateStartupConsistencyTestMemberships(ctx context.Context) ([]string, error) {
	rows := make([]struct {
		Id       int    `json:"id" orm:"id"`
		Username string `json:"username" orm:"username"`
	}, 0)
	err := dao.SysUser.Ctx(ctx).
		As("u").
		Fields("u.id, u.username").
		InnerJoin(
			startupConsistencyMembershipTable+" m",
			"m.user_id = u.id AND m.deleted_at IS NULL AND m.status = 1",
		).
		Where("u.tenant_id", int(tenantcap.PLATFORM)).
		Limit(10).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	details := make([]string, 0, len(rows))
	for _, row := range rows {
		details = append(details, "platform user "+row.Username+"("+gconv.String(row.Id)+") must not have active tenant membership")
	}
	return details, nil
}

// insertStartupConsistencyPlugin inserts one isolated plugin row for validation tests.
func insertStartupConsistencyPlugin(t *testing.T, ctx context.Context, data do.SysPlugin) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).Data(data).Insert(); err != nil {
		t.Fatalf("insert startup consistency plugin: %v", err)
	}
}

// cleanupStartupConsistencyPlugin removes one isolated plugin row.
func cleanupStartupConsistencyPlugin(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{PluginId: pluginID}).Delete(); err != nil {
		t.Fatalf("cleanup startup consistency plugin: %v", err)
	}
}

// insertStartupConsistencyUser inserts one user row for startup validation tests.
func insertStartupConsistencyUser(t *testing.T, ctx context.Context, username string, tenantID int) int64 {
	t.Helper()

	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "startup-consistency-test",
		Nickname: username,
		Status:   1,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert startup consistency user: %v", err)
	}
	return id
}

// insertStartupConsistencyTenantMembership inserts one membership row for startup validation tests.
func insertStartupConsistencyTenantMembership(t *testing.T, ctx context.Context, userID int64, tenantID int, status int) {
	t.Helper()

	ensureStartupConsistencyTenantMembershipTable(t, ctx)
	_, err := dao.SysUser.DB().Model(startupConsistencyMembershipTable).Data(startupConsistencyMembershipRow{
		UserID:    userID,
		TenantID:  tenantID,
		Status:    status,
		CreatedBy: 0,
		UpdatedBy: 0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert startup consistency membership: %v", err)
	}
}

// ensureStartupConsistencyTenantMembershipTable creates the plugin-owned
// membership table required by startup validation tests when the plugin schema
// has not been installed in the local test database.
func ensureStartupConsistencyTenantMembershipTable(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS plugin_linapro_tenant_core_user_membership (
			"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			"user_id" BIGINT NOT NULL,
			"tenant_id" BIGINT NOT NULL,
			"status" SMALLINT NOT NULL DEFAULT 1,
			"joined_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"created_by" BIGINT NOT NULL DEFAULT 0,
			"updated_by" BIGINT NOT NULL DEFAULT 0,
			"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"deleted_at" TIMESTAMP,
			CONSTRAINT uk_plugin_linapro_tenant_core_membership_user_tenant UNIQUE ("user_id", "tenant_id")
		)`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "joined_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "created_by" BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "updated_by" BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE plugin_linapro_tenant_core_user_membership ADD COLUMN IF NOT EXISTS "deleted_at" TIMESTAMP`,
		`CREATE INDEX IF NOT EXISTS idx_plugin_linapro_tenant_core_membership_tenant
			ON plugin_linapro_tenant_core_user_membership ("tenant_id", "status")`,
		`CREATE INDEX IF NOT EXISTS idx_plugin_linapro_tenant_core_membership_user
			ON plugin_linapro_tenant_core_user_membership ("user_id", "status")`,
	}
	for _, statement := range statements {
		if _, err := dao.SysUser.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("ensure startup consistency membership table: %v", err)
		}
	}
}

// startupConsistencyMembershipRow models the plugin membership columns touched
// by startup consistency tests without importing plugin-internal generated DOs.
type startupConsistencyMembershipRow struct {
	UserID    int64 `orm:"user_id"`
	TenantID  int   `orm:"tenant_id"`
	Status    int   `orm:"status"`
	CreatedBy int   `orm:"created_by"`
	UpdatedBy int   `orm:"updated_by"`
}

// cleanupStartupConsistencyUserMembership removes startup validation user fixtures.
func cleanupStartupConsistencyUserMembership(t *testing.T, ctx context.Context, username string, tenantID int) {
	t.Helper()

	ensureStartupConsistencyTenantMembershipTable(t, ctx)
	var user *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Username: username}).Scan(&user); err != nil {
		t.Fatalf("query startup consistency user cleanup: %v", err)
	}
	if user != nil {
		if _, err := dao.SysUser.DB().Model(startupConsistencyMembershipTable).
			Unscoped().
			Where("user_id", user.Id).
			Delete(); err != nil {
			t.Fatalf("cleanup startup consistency membership by user: %v", err)
		}
	}
	if _, err := dao.SysUser.DB().Model(startupConsistencyMembershipTable).
		Unscoped().
		Where("tenant_id", tenantID).
		Delete(); err != nil {
		t.Fatalf("cleanup startup consistency membership by tenant: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Username: username}).Delete(); err != nil {
		t.Fatalf("cleanup startup consistency user: %v", err)
	}
}

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

// assertStartupConsistencyErrorContains verifies startup errors use the stable
// bizerr code and include actionable details.
func assertStartupConsistencyErrorContains(t *testing.T, err error, expectedDetails ...string) {
	t.Helper()

	if !bizerr.Is(err, CodePluginStartupConsistencyFailed) {
		t.Fatalf("expected startup consistency bizerr, got %v", err)
	}
	message := err.Error()
	for _, detail := range expectedDetails {
		if !strings.Contains(message, detail) {
			t.Fatalf("expected startup consistency error to include %q, got %q", detail, message)
		}
	}
}
