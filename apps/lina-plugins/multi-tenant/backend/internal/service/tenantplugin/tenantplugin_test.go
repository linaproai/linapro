// This file verifies tenant plugin-governance service behavior.

package tenantplugin

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/dao"
	"lina-plugin-multi-tenant/backend/internal/model/do"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// testBizContextService returns a stable plugin-visible business context.
type testBizContextService struct {
	current plugincontract.CurrentContext
}

// Current returns the configured business context snapshot.
func (s testBizContextService) Current(context.Context) plugincontract.CurrentContext {
	return s.current
}

// TestTenantPluginListFiltersTenantScopedPlugins verifies tenant admins only
// see installed tenant-aware tenant-scoped plugins.
func TestTenantPluginListFiltersTenantScopedPlugins(t *testing.T) {
	ctx := context.Background()
	configureTenantPluginTestDB(t, ctx)

	const (
		tenantID         = 424261
		tenantPluginID   = "tc-tenant-plugin-list"
		globalPluginID   = "tc-tenant-plugin-global"
		platformPluginID = "tc-tenant-plugin-platform"
	)
	insertTenantPluginRegistry(t, ctx, tenantPluginID, pluginScopeNatureTenantAware, pluginInstallModeTenantScoped)
	insertTenantPluginRegistry(t, ctx, globalPluginID, pluginScopeNatureTenantAware, pluginInstallModeGlobal)
	insertTenantPluginRegistry(t, ctx, platformPluginID, pluginScopeNaturePlatformOnly, pluginInstallModeGlobal)
	t.Cleanup(func() {
		cleanupTenantPluginRows(t, ctx, tenantPluginID, globalPluginID, platformPluginID)
	})

	service := newTenantPluginTestService(tenantID)
	if err := service.SetEnabled(ctx, tenantPluginID, true); err != nil {
		t.Fatalf("enable tenant plugin failed: %v", err)
	}
	out, err := service.List(ctx)
	if err != nil {
		t.Fatalf("list tenant plugins failed: %v", err)
	}
	var found *Entity
	for _, item := range out.List {
		if item.Id == globalPluginID || item.Id == platformPluginID {
			t.Fatalf("non tenant-scoped plugin leaked into tenant list: %#v", item)
		}
		if item.Id == tenantPluginID {
			found = item
		}
	}
	if found == nil {
		t.Fatalf("expected tenant-scoped plugin %s in list, got %#v", tenantPluginID, out.List)
	}
	if found.TenantEnabled != 1 {
		t.Fatalf("expected tenant plugin enabled projection, got %#v", found)
	}
}

// TestTenantPluginSetEnabledRequiresTenantContext verifies platform context
// cannot mutate tenant-scoped enablement rows.
func TestTenantPluginSetEnabledRequiresTenantContext(t *testing.T) {
	ctx := context.Background()
	configureTenantPluginTestDB(t, ctx)

	err := New(testBizContextService{}).SetEnabled(ctx, "missing", true)
	if err == nil {
		t.Fatal("expected tenant context error")
	}
}

// TestTenantPluginSetEnabledBumpsPluginRuntimeRevision verifies tenant
// enablement mutations publish a shared plugin-runtime cache revision.
func TestTenantPluginSetEnabledBumpsPluginRuntimeRevision(t *testing.T) {
	ctx := context.Background()
	configureTenantPluginTestDB(t, ctx)

	const (
		tenantID       = 424263
		tenantPluginID = "tc-tenant-plugin-runtime-revision"
	)
	insertTenantPluginRegistry(t, ctx, tenantPluginID, pluginScopeNatureTenantAware, pluginInstallModeTenantScoped)
	t.Cleanup(func() {
		cleanupTenantPluginRows(t, ctx, tenantPluginID)
	})

	beforeRevision := pluginRuntimeRevisionForTest(t, ctx)
	service := newTenantPluginTestService(tenantID)

	if err := service.SetEnabled(ctx, tenantPluginID, false); err != nil {
		t.Fatalf("disable tenant plugin failed: %v", err)
	}
	if err := service.SetEnabled(ctx, tenantPluginID, false); err != nil {
		t.Fatalf("repeat disable tenant plugin failed: %v", err)
	}
	if enabled := tenantPluginEnabledForTest(t, ctx, tenantPluginID, tenantID); enabled {
		t.Fatalf("expected plugin %s disabled for tenant %d", tenantPluginID, tenantID)
	}

	if err := service.SetEnabled(ctx, tenantPluginID, true); err != nil {
		t.Fatalf("enable tenant plugin failed: %v", err)
	}
	if err := service.SetEnabled(ctx, tenantPluginID, true); err != nil {
		t.Fatalf("repeat enable tenant plugin failed: %v", err)
	}
	if enabled := tenantPluginEnabledForTest(t, ctx, tenantPluginID, tenantID); !enabled {
		t.Fatalf("expected plugin %s enabled for tenant %d", tenantPluginID, tenantID)
	}

	afterRevision := pluginRuntimeRevisionForTest(t, ctx)
	if afterRevision < beforeRevision+4 {
		t.Fatalf("expected plugin-runtime revision to increase at least 4 times from %d, got %d", beforeRevision, afterRevision)
	}
}

// TestProvisionForTenantAutoEnablesPolicyEnabledTenantPlugins verifies new tenants
// receive enablement rows for tenant-scoped plugins allowed by platform policy.
func TestProvisionForTenantAutoEnablesPolicyEnabledTenantPlugins(t *testing.T) {
	ctx := context.Background()
	configureTenantPluginTestDB(t, ctx)

	const (
		tenantID        = 424262
		defaultPluginID = "tc-default-for-new-tenant"
		manualPluginID  = "tc-manual-for-new-tenant"
	)
	insertTenantPluginRegistryWithDefault(t, ctx, defaultPluginID, pluginScopeNatureTenantAware, pluginInstallModeTenantScoped, true)
	insertTenantPluginRegistryWithDefault(t, ctx, manualPluginID, pluginScopeNatureTenantAware, pluginInstallModeTenantScoped, false)
	t.Cleanup(func() {
		cleanupTenantPluginRows(t, ctx, defaultPluginID, manualPluginID)
	})

	if err := New(testBizContextService{}).ProvisionForTenant(ctx, tenantID); err != nil {
		t.Fatalf("provision default tenant plugins failed: %v", err)
	}

	if enabled := tenantPluginEnabledForTest(t, ctx, defaultPluginID, tenantID); !enabled {
		t.Fatalf("expected default plugin %s enabled for tenant %d", defaultPluginID, tenantID)
	}
	if enabled := tenantPluginEnabledForTest(t, ctx, manualPluginID, tenantID); enabled {
		t.Fatalf("expected manual plugin %s to remain disabled for tenant %d", manualPluginID, tenantID)
	}
}

// TestProvisionForTenantSkipsAutoEnablePolicyWhenPluginDisabled verifies the policy
// does not install or enable host-disabled plugins for new tenants.
func TestProvisionForTenantSkipsAutoEnablePolicyWhenPluginDisabled(t *testing.T) {
	ctx := context.Background()
	configureTenantPluginTestDB(t, ctx)

	const (
		tenantID = 424264
		pluginID = "tc-auto-policy-disabled-host-plugin"
	)
	insertTenantPluginRegistryWithDefault(t, ctx, pluginID, pluginScopeNatureTenantAware, pluginInstallModeTenantScoped, true)
	if _, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{Status: 0}).
		Update(); err != nil {
		t.Fatalf("disable plugin registry %s failed: %v", pluginID, err)
	}
	t.Cleanup(func() {
		cleanupTenantPluginRows(t, ctx, pluginID)
	})

	if err := New(testBizContextService{}).ProvisionForTenant(ctx, tenantID); err != nil {
		t.Fatalf("provision default tenant plugins failed: %v", err)
	}

	if enabled := tenantPluginEnabledForTest(t, ctx, pluginID, tenantID); enabled {
		t.Fatalf("expected disabled host plugin %s to remain disabled for tenant %d", pluginID, tenantID)
	}
}

// configureTenantPluginTestDB points the package test at local PostgreSQL.
func configureTenantPluginTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure tenant plugin test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close tenant plugin test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore tenant plugin test database config failed: %v", err)
		}
	})
}

// insertTenantPluginRegistry creates one host plugin registry row for tests.
func insertTenantPluginRegistry(t *testing.T, ctx context.Context, pluginID string, scopeNature string, installMode string) {
	t.Helper()
	insertTenantPluginRegistryWithDefault(t, ctx, pluginID, scopeNature, installMode, false)
}

// newTenantPluginTestService creates a service with a fixed tenant context.
func newTenantPluginTestService(tenantID int) Service {
	return &serviceImpl{bizCtxSvc: testBizContextService{current: plugincontract.CurrentContext{TenantID: tenantID}}}
}

// insertTenantPluginRegistryWithDefault creates one host plugin registry row for tests.
func insertTenantPluginRegistryWithDefault(t *testing.T, ctx context.Context, pluginID string, scopeNature string, installMode string, autoEnableForNewTenants bool) {
	t.Helper()
	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{PluginId: pluginID}).Delete(); err != nil {
		t.Fatalf("cleanup stale plugin registry %s failed: %v", pluginID, err)
	}
	_, err := dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:                pluginID,
		Name:                    pluginID,
		Version:                 "v0.1.0",
		Type:                    pluginTypeSource,
		Installed:               pluginInstalledYes,
		Status:                  pluginStatusEnabled,
		DesiredState:            pluginHostStateEnabled,
		CurrentState:            pluginHostStateEnabled,
		ScopeNature:             scopeNature,
		InstallMode:             installMode,
		AutoEnableForNewTenants: autoEnableForNewTenants,
	}).InsertIgnore()
	if err != nil {
		t.Fatalf("insert plugin registry %s failed: %v", pluginID, err)
	}
}

// tenantPluginEnabledForTest reads one tenant enablement row.
func tenantPluginEnabledForTest(t *testing.T, ctx context.Context, pluginID string, tenantID int64) bool {
	t.Helper()
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: pluginID,
			TenantId: tenantID,
			StateKey: tenantEnablementStateKey,
		}).
		Value(dao.SysPluginState.Columns().Enabled)
	if err != nil {
		t.Fatalf("read tenant plugin state failed: %v", err)
	}
	return value != nil && !value.IsNil() && value.Bool()
}

// pluginRuntimeRevisionForTest reads the shared plugin-runtime revision row.
func pluginRuntimeRevisionForTest(t *testing.T, ctx context.Context) int64 {
	t.Helper()
	value, err := g.DB().Model(tableSysCacheRevision).Safe().Ctx(ctx).
		Where(pluginRuntimeCacheRevisionDO{
			TenantId: shared.PlatformTenantID,
			Domain:   pluginRuntimeCacheDomain,
			Scope:    pluginRuntimeCacheScopeGlobal,
		}).
		Value("revision")
	if err != nil {
		t.Fatalf("read plugin-runtime cache revision failed: %v", err)
	}
	if value == nil || value.IsNil() {
		return 0
	}
	return value.Int64()
}

// cleanupTenantPluginRows removes host plugin registry and enablement rows.
func cleanupTenantPluginRows(t *testing.T, ctx context.Context, pluginIDs ...string) {
	t.Helper()
	if _, err := dao.SysPluginState.Ctx(ctx).Unscoped().WhereIn(dao.SysPluginState.Columns().PluginId, pluginIDs).Delete(); err != nil {
		t.Errorf("cleanup tenant plugin state failed: %v", err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().WhereIn(dao.SysPlugin.Columns().PluginId, pluginIDs).Delete(); err != nil {
		t.Errorf("cleanup tenant plugin registry failed: %v", err)
	}
}
