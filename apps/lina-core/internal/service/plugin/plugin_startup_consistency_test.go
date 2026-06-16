// This file verifies startup consistency checks for plugin tenant governance.

package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/governance"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
	capabilityhostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	capabilitymanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

const startupConsistencyMembershipTable = "plugin_linapro_tenant_core_user_membership"

// TestEnsurePlatformGovernanceAllowsSingleTenantMode verifies disabled tenancy
// keeps plugin governance available in platform-only deployments.
func TestEnsurePlatformGovernanceAllowsSingleTenantMode(t *testing.T) {
	err := governance.EnsurePlatformContext(context.Background(), pluginTenantGuard{enabled: false})
	if err != nil {
		t.Fatalf("expected disabled tenancy to allow plugin governance, got %v", err)
	}
}

// TestEnsurePlatformGovernanceRejectsTenantContext verifies active
// multi-tenancy requires platform all-data context for lifecycle writes.
func TestEnsurePlatformGovernanceRejectsTenantContext(t *testing.T) {
	err := governance.EnsurePlatformContext(context.Background(), pluginTenantGuard{enabled: true, platformBypass: false})
	if !bizerr.Is(err, tenantcap.CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestEnsurePlatformGovernanceAllowsPlatformBypass verifies platform all-data
// context can perform plugin governance actions.
func TestEnsurePlatformGovernanceAllowsPlatformBypass(t *testing.T) {
	err := governance.EnsurePlatformContext(context.Background(), pluginTenantGuard{enabled: true, platformBypass: true})
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
	svc, err := newTestServiceWithTopologyAndTenantDeps(nil, nil, nil, newPluginPlatformGuardTenantService(t))
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}
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
			if err := tc.run(); !bizerr.Is(err, tenantcap.CodePlatformPermissionRequired) {
				t.Fatalf("expected platform permission error, got %v", err)
			}
		})
	}
}

// TestWithStartupDataSnapshotReusesCatalogAndIntegrationSnapshots verifies one
// startup context does not rebuild equivalent plugin snapshots repeatedly.
func TestWithStartupDataSnapshotReusesCatalogAndIntegrationSnapshots(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	service := newTestService()

	startupCtx, err := service.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("build first startup snapshot: %v", err)
	}
	startupCtx, err = service.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("reuse second startup snapshot: %v", err)
	}
	if _, err = service.ReadOnlyList(startupCtx); err != nil {
		t.Fatalf("read plugin list with startup snapshots: %v", err)
	}

	snapshot := startupstats.FromContext(startupCtx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 1 {
		t.Fatalf("expected one integration snapshot build, got %d", got)
	}
}

// TestReadOnlyListOnlyBuildsCatalogSnapshot verifies management read paths do
// not load integration snapshots that are only needed by startup sync.
func TestReadOnlyListOnlyBuildsCatalogSnapshot(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	service := newTestService()

	if _, err := service.ReadOnlyList(ctx); err != nil {
		t.Fatalf("read plugin list with catalog startup snapshot: %v", err)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds); got != 0 {
		t.Fatalf("expected no integration snapshot build, got %d", got)
	}
}

// TestNewRequiresInjectedTenantCapability verifies startup tenant consistency
// dependencies must be explicit at construction time.
func TestNewRequiresInjectedTenantCapability(t *testing.T) {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
		i18nSvc        = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		pluginRuntime  = NewRuntimeDelegate()
		orgSvc         = orgspi.New(nil, pluginRuntime)
		roleSvc        = role.New(pluginRuntime, bizCtxProvider, configProvider, i18nSvc, orgSvc, tenantspi.New(nil, pluginRuntime, bizCtxProvider))
		capabilities   = newRootTestCapabilities(bizCtxProvider, pluginRuntime)
	)
	_, err := New(
		nil,
		configProvider,
		bizCtxProvider,
		cacheCoordSvc,
		i18nSvc,
		session.NewDBStore(),
		roleSvc,
		locker.New(),
		nil,
		capabilities,
		orgSvc,
		nil,
		tenantspi.New(nil, pluginRuntime, bizCtxProvider),
		tenantspi.New(nil, pluginRuntime, bizCtxProvider),
		capabilityconfig.NewConfigFactory("", ""),
		capabilityhostconfig.New(mustHostConfigRawReader(configProvider)),
		capabilitymanifest.NewFactory(""),
	)
	if err == nil || !strings.Contains(err.Error(), "tenant startup capability") {
		t.Fatalf("expected tenant startup dependency error, got %v", err)
	}
}

// TestValidateStartupConsistencyUsesInjectedTenantCapability verifies tenant
// membership checks run through the explicitly wired tenant capability.
func TestValidateStartupConsistencyUsesInjectedTenantCapability(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantSvc = &startupConsistencyTenantCapability{details: []string{"injected tenant capability used"}}
	)
	service, err := newTestServiceWithTopologyAndTenantDeps(nil, tenantSvc, nil, nil)
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}

	err = service.ValidateStartupConsistency(ctx)
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
		Type:        plugintypes.TypeSource.String(),
		Installed:   plugintypes.InstalledYes,
		Status:      plugintypes.StatusEnabled,
		ScopeNature: "invalid_scope",
		InstallMode: plugintypes.InstallModeTenantScoped.String(),
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
		Type:        plugintypes.TypeSource.String(),
		Installed:   plugintypes.InstalledYes,
		Status:      plugintypes.StatusEnabled,
		ScopeNature: plugintypes.ScopeNaturePlatformOnly.String(),
		InstallMode: plugintypes.InstallModeTenantScoped.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform_only plugin "+pluginID+" must use global install_mode")
}

// TestValidateStartupConsistencyRejectsPlatformUserMembership verifies
// platform users are not allowed to carry active tenant memberships.
func TestValidateStartupConsistencyRejectsPlatformUserMembership(t *testing.T) {
	var (
		ctx      = context.Background()
		username = "startup-platform-member"
		tenantID = 19001
	)
	cleanupStartupConsistencyPlugin(t, ctx, tenantcap.ProviderPluginID)
	cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID)
	service, err := newTestServiceWithTopologyAndTenantDeps(nil, &startupConsistencyTenantCapability{
		available:           true,
		validateMemberships: true,
	}, nil, nil)
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, tenantcap.ProviderPluginID) })
	t.Cleanup(func() { cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    tenantcap.ProviderPluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        plugintypes.TypeSource.String(),
		Installed:   plugintypes.InstalledYes,
		Status:      plugintypes.StatusEnabled,
		ScopeNature: plugintypes.ScopeNaturePlatformOnly.String(),
		InstallMode: plugintypes.InstallModeGlobal.String(),
	})
	userID := insertStartupConsistencyUser(t, ctx, username, int(tenantcap.PLATFORM))
	insertStartupConsistencyTenantMembership(t, ctx, userID, tenantID, 1)

	err = service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform user "+username)
}

// TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider
// verifies linapro-tenant-core enablement requires a registered tenantcap provider.
func TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = tenantcap.ProviderPluginID
	)
	service, err := newTestServiceWithTopologyAndTenantDeps(nil, &startupConsistencyTenantCapability{}, nil, nil)
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        plugintypes.TypeSource.String(),
		Installed:   plugintypes.InstalledYes,
		Status:      plugintypes.StatusEnabled,
		ScopeNature: plugintypes.ScopeNaturePlatformOnly.String(),
		InstallMode: plugintypes.InstallModeGlobal.String(),
	})

	err = service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "linapro-tenant-core plugin is enabled but capability tenant provider is not active")
}

// TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider verifies
// provider registration satisfies linapro-tenant-core startup consistency.
func TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = tenantcap.ProviderPluginID
	)
	service, err := newTestServiceWithTopologyAndTenantDeps(nil, &startupConsistencyTenantCapability{available: true}, nil, nil)
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        plugintypes.TypeSource.String(),
		Installed:   plugintypes.InstalledYes,
		Status:      plugintypes.StatusEnabled,
		ScopeNature: plugintypes.ScopeNaturePlatformOnly.String(),
		InstallMode: plugintypes.InstallModeGlobal.String(),
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
			Type:                plugintypes.TypeSource.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
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
			Type:                plugintypes.TypeSource.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  plugintypes.InstallModeGlobal.String(),
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
			Type:                plugintypes.TypeSource.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supports,
			DefaultInstallMode:  plugintypes.InstallModeGlobal.String(),
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
		Data(do.SysPlugin{InstallMode: plugintypes.InstallModeTenantScoped.String()}).
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

// pluginTenantGuard is the narrow tenantcap fake needed by plugin platform-guard tests.
type pluginTenantGuard struct {
	enabled        bool
	platformBypass bool
}

// Enabled returns whether multi-tenancy is active in this test.
func (g pluginTenantGuard) Available(context.Context) bool {
	return g.enabled
}

// PlatformBypass returns whether the test context is platform all-data.
func (g pluginTenantGuard) PlatformBypass(context.Context) bool {
	return g.platformBypass
}

// newPluginPlatformGuardTenantService creates a real tenantcap service with
// one enabled test provider for plugin facade entry-point tests.
func newPluginPlatformGuardTenantService(t *testing.T) tenantspi.RuntimeService {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-plugin-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return pluginPlatformGuardProvider{}, nil
	}); err != nil {
		t.Fatalf("register plugin tenant provider: %v", err)
	}
	return tenantspi.New(manager, pluginPlatformGuardProviderRuntime{pluginID: providerPluginID}, bizctx.New())
}

// pluginPlatformGuardProviderRuntime marks exactly one test provider plugin enabled.
type pluginPlatformGuardProviderRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r pluginPlatformGuardProviderRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
}

// TenantProviderEnv returns an empty typed provider environment in plugin facade tests.
func (pluginPlatformGuardProviderRuntime) TenantProviderEnv(string) tenantspi.ProviderEnv {
	return tenantspi.ProviderEnv{}
}

// pluginPlatformGuardProvider satisfies the tenantcap provider contract for
// tests that only need provider presence.
type pluginPlatformGuardProvider struct{}

// ResolveTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ValidateUserInTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// ListUserTenants is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant is unused by plugin platform-guard tests.
func (pluginPlatformGuardProvider) SwitchTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}
