// This file verifies startup bootstrap behavior driven by plugin.autoEnable.

package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	runtimepkg "lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBootstrapAutoEnableInstallsAndEnablesSourcePlugin verifies startup
// bootstrap promotes a discovered source plugin to the enabled state.
func TestBootstrapAutoEnableInstallsAndEnablesSourcePlugin(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-source-auto-enable"
		version  = "v0.1.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: Source Auto Enable Plugin\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n",
	)

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if err := service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected source plugin startup bootstrap to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected source plugin registry row after startup bootstrap")
	}
	if registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
		t.Fatalf("expected source plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected source plugin current state enabled, got %s", registry.CurrentState)
	}

	release, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected source plugin release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatal("expected source plugin release row after startup bootstrap")
	}
}

// TestBootstrapAutoEnableSourcePluginUpdatesStartupSnapshot verifies startup
// bootstrap keeps the shared startup snapshot fresh after installing a source
// plugin so the immediate enable step can observe the installed state.
func TestBootstrapAutoEnableSourcePluginUpdatesStartupSnapshot(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-source-auto-enable-snapshot"
		version  = "v0.1.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: Source Auto Enable Snapshot Plugin\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n",
	)

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	startupCtx, err := service.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("expected startup snapshot context build to succeed, got error: %v", err)
	}
	if err = service.BootstrapAutoEnable(startupCtx); err != nil {
		t.Fatalf("expected source plugin startup bootstrap with snapshot to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected source plugin registry row after startup bootstrap")
	}
	if registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
		t.Fatalf("expected source plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected source plugin current state enabled, got %s", registry.CurrentState)
	}
}

// TestBootstrapAutoEnableReusesDynamicAuthorizationSnapshot verifies startup
// bootstrap can reinstall and enable a dynamic plugin after one confirmed host
// service authorization snapshot already exists for the target release.
func TestBootstrapAutoEnableReusesDynamicAuthorizationSnapshot(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-dynamic-auto-enable-auth"
		version  = "v0.6.0"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtimepkg.BuildArtifactFileName(pluginID))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Auto Enable Authorization Plugin",
			Version: version,
			Type:    plugintypes.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceRuntime,
					Methods: []string{protocol.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{protocol.HostServiceMethodStorageGet},
					Paths:   []string{"private-files/"},
				},
			},
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	authorization := &HostServiceAuthorizationInput{
		Services: []*HostServiceAuthorizationDecision{
			{
				Service: protocol.HostServiceStorage,
				Paths:   []string{"private-files/"},
			},
		},
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	if _, err := service.Install(ctx, pluginID, InstallOptions{Authorization: authorization}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, plugintypes.StatusEnabled, authorization); err != nil {
		t.Fatalf("expected initial dynamic plugin enable to succeed, got error: %v", err)
	}
	if err := service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err != nil {
		t.Fatalf("expected dynamic plugin uninstall to succeed, got error: %v", err)
	}

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})

	bootstrapService := newTestService()
	if err := bootstrapService.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected dynamic plugin startup bootstrap to reuse authorization snapshot, got error: %v", err)
	}

	registry, err := bootstrapService.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected dynamic plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected dynamic plugin registry row after startup bootstrap")
	}
	if registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
		t.Fatalf("expected dynamic plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected dynamic plugin current state enabled, got %s", registry.CurrentState)
	}

	release, err := bootstrapService.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic plugin release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatal("expected dynamic plugin release row after startup bootstrap")
	}
	snapshot, err := bootstrapService.storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		t.Fatalf("expected manifest snapshot parse to succeed, got error: %v", err)
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		t.Fatalf("expected confirmed authorization snapshot after startup bootstrap, got %#v", snapshot)
	}
}

// TestBootstrapAutoEnableRejectsDynamicPluginWithoutAuthorizationSnapshot verifies
// startup bootstrap fails fast when a governed dynamic plugin has not gone
// through the regular authorization review flow yet.
func TestBootstrapAutoEnableRejectsDynamicPluginWithoutAuthorizationSnapshot(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-dynamic-auto-enable-auth-missing"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtimepkg.BuildArtifactFileName(pluginID))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Auto Enable Missing Authorization Plugin",
			Version: "v0.6.1",
			Type:    plugintypes.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceNetwork,
					Methods: []string{protocol.HostServiceMethodNetworkRequest},
					Resources: []*protocol.HostServiceResourceSpec{
						{Ref: "https://example.com/api"},
					},
				},
			},
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	err := service.BootstrapAutoEnable(ctx)
	if err == nil {
		t.Fatal("expected startup bootstrap to reject missing authorization snapshot")
	}
	if got := err.Error(); got == "" || !containsAll(got, pluginID, "authorization snapshot") {
		t.Fatalf("expected bootstrap error to mention plugin ID and authorization snapshot, got %q", got)
	}
}

// TestBootstrapAutoEnableWaitsUntilCurrentNodeBecomesPrimary verifies cluster
// startup bootstrap can wait for leader election and then perform the shared
// dynamic lifecycle actions once this node becomes primary.
func TestBootstrapAutoEnableWaitsUntilCurrentNodeBecomesPrimary(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-auto-enable-cluster"
		topology = &testTopology{
			enabled: true,
			primary: false,
			nodeID:  "bootstrap-follower",
		}
		service = newTestServiceWithTopology(topology)
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Dynamic Cluster Bootstrap Plugin",
		"v0.7.0",
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
	)

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	timer := time.AfterFunc(150*time.Millisecond, func() {
		topology.SetPrimary(true)
	})
	t.Cleanup(func() {
		timer.Stop()
	})

	if err := service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected cluster startup bootstrap to succeed after leadership handoff, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected cluster bootstrap registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected cluster bootstrap registry row after startup bootstrap")
	}
	if registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
		t.Fatalf("expected cluster bootstrap plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected cluster bootstrap plugin current state enabled, got %s", registry.CurrentState)
	}
}

// containsAll reports whether one error string contains all expected fragments.
func containsAll(message string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(message, fragment) {
			return false
		}
	}
	return true
}

// dropMockTableIfExists removes the temporary table used by the mock data
// auto-enable test so re-runs start from a clean state.
func dropMockTableIfExists(t *testing.T, ctx context.Context, tableName string) {
	t.Helper()
	if _, err := g.DB().Exec(ctx, "DROP TABLE IF EXISTS "+tableName+";"); err != nil {
		t.Fatalf("expected mock test table cleanup to succeed, got error: %v", err)
	}
}

// mockTableRowCount returns the row count of the mock data demo table created
// by the auto-enable opt-in test, or 0 if the table does not exist yet.
func mockTableRowCount(t *testing.T, ctx context.Context, tableName string) int {
	t.Helper()
	value, err := g.DB().GetValue(ctx, "SELECT COUNT(1) FROM "+tableName+";")
	if err != nil {
		t.Fatalf("expected mock table row count query to succeed, got error: %v", err)
	}
	return value.Int()
}

// mockMigrationRowCount returns the count of sys_plugin_migration rows with
// phase='mock' for the given plugin ID, used to assert the auto-enable
// bootstrap honored or skipped the mock-data load opt-in.
func mockMigrationRowCount(t *testing.T, ctx context.Context, pluginID string) int {
	t.Helper()
	value, err := g.DB().GetValue(
		ctx,
		"SELECT COUNT(1) FROM sys_plugin_migration WHERE plugin_id = ? AND phase = ?;",
		pluginID,
		plugintypes.MigrationDirectionMock.String(),
	)
	if err != nil {
		t.Fatalf("expected mock migration row count query to succeed, got error: %v", err)
	}
	return value.Int()
}

// TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn verifies the union-schema
// entry form: a bare-string entry installs without mock data, an object entry
// with WithMockData=true loads the mock-data SQL during the startup auto-install.
func TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn(t *testing.T) {
	var (
		ctx              = context.Background()
		service          = newTestService()
		pluginIDNoMock   = "plugin-dev-source-auto-enable-no-mock"
		pluginIDWithMock = "plugin-dev-source-auto-enable-with-mock"
		mockTable        = "plugin_source_auto_enable_with_mock_demo"
		version          = "v0.1.0"
	)

	// Plugin without any mock-data: bare-string entry, expected to install cleanly.
	pluginDirNoMock := testutil.CreateTestPluginDir(t, pluginIDNoMock)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDirNoMock, "plugin.yaml"),
		"id: "+pluginIDNoMock+"\nname: Source Auto Enable No Mock\nversion: "+version+"\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	// Plugin with mock-data: object entry with WithMockData=true.
	pluginDirWithMock := testutil.CreateTestPluginDir(t, pluginIDWithMock)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDirWithMock, "plugin.yaml"),
		"id: "+pluginIDWithMock+"\nname: Source Auto Enable With Mock\nversion: "+version+"\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDirWithMock, "manifest", "sql", "001-"+pluginIDWithMock+".sql"),
		"CREATE TABLE IF NOT EXISTS "+mockTable+" (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDirWithMock, "manifest", "sql", "uninstall", "001-"+pluginIDWithMock+".sql"),
		"DROP TABLE IF EXISTS "+mockTable+";",
	)
	if err := os.MkdirAll(filepath.Join(pluginDirWithMock, "manifest", "sql", "mock-data"), 0o755); err != nil {
		t.Fatalf("failed to create mock-data dir: %v", err)
	}
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDirWithMock, "manifest", "sql", "mock-data", "001-"+pluginIDWithMock+"-mock.sql"),
		"INSERT INTO "+mockTable+" (marker) VALUES ('startup-mock-row');",
	)

	configsvc.SetPluginAutoEnableEntriesOverride([]configsvc.PluginAutoEnableEntry{
		{ID: pluginIDNoMock, WithMockData: false},
		{ID: pluginIDWithMock, WithMockData: true},
	})
	dropMockTableIfExists(t, ctx, mockTable)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginIDNoMock)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginIDWithMock)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableEntriesOverride(nil)
		dropMockTableIfExists(t, ctx, mockTable)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginIDNoMock)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginIDWithMock)
	})

	if err := service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected startup bootstrap with mock opt-in to succeed, got error: %v", err)
	}

	// Plugin with WithMockData=true should have its mock data committed.
	rows := mockTableRowCount(t, ctx, mockTable)
	if rows != 1 {
		t.Fatalf("expected 1 mock row for opt-in plugin, got %d", rows)
	}
	mockMigrations := mockMigrationRowCount(t, ctx, pluginIDWithMock)
	if mockMigrations != 1 {
		t.Fatalf("expected 1 sys_plugin_migration row with phase=mock for opt-in plugin, got %d", mockMigrations)
	}

	// Plugin without WithMockData should not have any mock-phase migration rows
	// even if a mock-data directory exists in the future.
	noMockMigrations := mockMigrationRowCount(t, ctx, pluginIDNoMock)
	if noMockMigrations != 0 {
		t.Fatalf("expected 0 sys_plugin_migration rows with phase=mock for non-opt-in plugin, got %d", noMockMigrations)
	}
}

// TestReconcileAutoEnabledTenantPluginsProvisionsTenantScopedEntries verifies
// startup auto-enable bridges host-level plugin enablement into tenant-scoped
// provisioning after the linapro-tenant-core provider has registered.
func TestReconcileAutoEnabledTenantPluginsProvisionsTenantScopedEntries(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = autoEnableTenantProvisioningPluginID
		version  = "v0.1.0"
		tenantID = tenantcap.TenantID(50101)
	)
	tenantSvc := &autoEnableTenantProvisioningService{enabled: true, tenantID: tenantID}
	service, err := newTestServiceWithTopologyAndTenantDeps(nil, nil, tenantSvc, nil)
	if err != nil {
		t.Fatalf("construct plugin service: %v", err)
	}
	tenantSvc.pluginSvc = service

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: Tenant Auto Provisioning Plugin\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: true\n"+
			"default_install_mode: tenant_scoped\n",
	)

	configsvc.SetPluginAutoEnableOverride([]string{pluginID})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if err := service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected source plugin startup bootstrap to succeed, got error: %v", err)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.AutoEnableForNewTenants {
		t.Fatalf("expected bootstrap to leave tenant provisioning policy unset before provider reconciliation, got %#v", registry)
	}

	if err = service.ReconcileAutoEnabledTenantPlugins(ctx); err != nil {
		t.Fatalf("expected tenant provisioning reconciliation to succeed, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after reconciliation to succeed, got error: %v", err)
	}
	if registry == nil || !registry.AutoEnableForNewTenants {
		t.Fatalf("expected tenant provisioning policy enabled after reconciliation, got %#v", registry)
	}
	if tenantSvc.provisionCalls != 1 {
		t.Fatalf("expected tenant provisioning to run once, got %d", tenantSvc.provisionCalls)
	}
	if !service.IsEnabled(datascope.WithTenantForTest(ctx, int(tenantID)), pluginID) {
		t.Fatalf("expected plugin %s to be visible for tenant %d after provisioning", pluginID, tenantID)
	}
}

// TestBootstrapAutoEnableRefreshesLegacyHostRuntimeArchive verifies startup
// runs a single-node dynamic reconciliation pass after manifest sync so stale
// archived artifacts with obsolete hostruntime declarations are repaired from
// the current staged artifact.
func TestBootstrapAutoEnableRefreshesLegacyHostRuntimeArchive(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-dynamic-legacy-host-config-startup"
		name     = "Legacy HostRuntime Startup Plugin"
		version  = "v0.1.0"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtimepkg.BuildArtifactFileName(pluginID))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    name,
			Version: version,
			Type:    plugintypes.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{{
				Service: protocol.HostServiceHostConfig,
				Methods: []string{protocol.HostServiceMethodHostConfigGet},
				Keys:    []string{"workspace.basePath"},
			}},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	authorization := &HostServiceAuthorizationInput{
		Services: []*HostServiceAuthorizationDecision{{
			Service: protocol.HostServiceHostConfig,
			Methods: []string{protocol.HostServiceMethodHostConfigGet},
			Keys:    []string{"workspace.basePath"},
		}},
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{Authorization: authorization}); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, plugintypes.StatusEnabled, authorization); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	release, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatal("expected release row after enable")
	}
	legacyArchivePath := filepath.Join(
		testutil.TestDynamicStorageDir(),
		"releases",
		pluginID,
		version,
		"legacy-hostruntime",
		runtimepkg.BuildArtifactFileName(pluginID),
	)
	writeLegacyHostRuntimeArtifact(t, legacyArchivePath, pluginID, name, version)
	legacySnapshot := strings.ReplaceAll(release.ManifestSnapshot, protocol.HostServiceHostConfig, "hostruntime")
	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: release.Id}).
		Data(do.SysPluginRelease{
			PackagePath:      filepath.ToSlash(strings.TrimPrefix(legacyArchivePath, testutil.TestDynamicStorageDir()+string(os.PathSeparator))),
			ManifestSnapshot: legacySnapshot,
		}).
		Update(); err != nil {
		t.Fatalf("expected release row to accept legacy archive fixture, got error: %v", err)
	}

	staleRegistry, err := service.storeSvc.RefreshStartupRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry refresh to succeed, got error: %v", err)
	}
	if staleRegistry == nil {
		t.Fatal("expected registry row before startup bootstrap")
	}
	if _, err = service.storeSvc.RefreshStartupReleaseByID(ctx, release.Id); err != nil {
		t.Fatalf("expected release refresh to succeed, got error: %v", err)
	}
	if _, err = service.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, staleRegistry); err == nil {
		t.Fatal("expected legacy active archive to fail strict artifact parsing before startup repair")
	}

	configsvc.SetPluginAutoEnableOverride(nil)
	if err = service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected startup bootstrap to repair legacy dynamic archive, got error: %v", err)
	}

	repairedRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected repaired registry lookup to succeed, got error: %v", err)
	}
	if repairedRegistry == nil || repairedRegistry.ReleaseId <= 0 {
		t.Fatalf("expected repaired registry with release id, got %#v", repairedRegistry)
	}
	repairedManifest, err := service.runtimeSvc.LoadActiveDynamicPluginManifest(ctx, repairedRegistry)
	if err != nil {
		t.Fatalf("expected repaired active manifest to load, got error: %v", err)
	}
	if len(repairedManifest.HostServices) != 1 ||
		repairedManifest.HostServices[0].Service != protocol.HostServiceHostConfig {
		t.Fatalf("expected repaired active manifest to use hostconfig, got %#v", repairedManifest.HostServices)
	}
	repairedRelease, err := service.storeSvc.GetRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected repaired release lookup to succeed, got error: %v", err)
	}
	if repairedRelease == nil || strings.Contains(repairedRelease.ManifestSnapshot, "hostruntime") {
		t.Fatalf("expected repaired release snapshot without hostruntime, got %#v", repairedRelease)
	}
	if strings.Contains(filepath.ToSlash(repairedRelease.PackagePath), "legacy-hostruntime") {
		t.Fatalf("expected repaired release package path to point at checksum archive, got %s", repairedRelease.PackagePath)
	}
}

// writeLegacyHostRuntimeArtifact writes a strict-invalid runtime artifact that
// mirrors the historical hostruntime service spelling found in old archives.
func writeLegacyHostRuntimeArtifact(t *testing.T, filePath string, pluginID string, name string, version string) {
	t.Helper()

	manifest := &catalog.ArtifactManifest{
		ID:                  pluginID,
		Name:                name,
		Version:             version,
		Type:                plugintypes.TypeDynamic.String(),
		ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &testutil.DefaultTestSupportsMultiTenant,
		DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
	}
	runtimeMetadata := &catalog.ArtifactSpec{
		RuntimeKind: protocol.RuntimeKindWasm,
		ABIVersion:  protocol.SupportedABIVersion,
	}
	manifestContent, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal legacy artifact manifest: %v", err)
	}
	runtimeContent, err := json.Marshal(runtimeMetadata)
	if err != nil {
		t.Fatalf("failed to marshal legacy artifact runtime metadata: %v", err)
	}
	hostServicesContent := []byte(`[{"service":"hostruntime","methods":["get"],"resources":{"keys":["workspace.basePath"]}}]`)

	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendTestWasmCustomSection(wasm, protocol.WasmSectionManifest, manifestContent)
	wasm = appendTestWasmCustomSection(wasm, protocol.WasmSectionRuntime, runtimeContent)
	wasm = appendTestWasmCustomSection(wasm, protocol.WasmSectionBackendHostServices, hostServicesContent)

	if err = os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create legacy artifact dir: %v", err)
	}
	if err = os.WriteFile(filePath, wasm, 0o644); err != nil {
		t.Fatalf("failed to write legacy artifact: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := os.Remove(filePath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove legacy artifact %s: %v", filePath, cleanupErr)
		}
	})
}

// appendTestWasmCustomSection appends a WASM custom section for local fixtures.
func appendTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeTestWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	out := append([]byte{}, content...)
	out = append(out, 0x00)
	out = append(out, encodeTestWasmULEB128(uint32(len(sectionPayload)))...)
	out = append(out, sectionPayload...)
	return out
}

// encodeTestWasmULEB128 encodes one unsigned WASM section length.
func encodeTestWasmULEB128(value uint32) []byte {
	out := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		out = append(out, current)
		if value == 0 {
			return out
		}
	}
}

// autoEnableTenantProvisioningPluginID is the fixture plugin owned by the
// startup tenant provisioning test.
const autoEnableTenantProvisioningPluginID = "plugin-dev-auto-enable-tenant-provisioning"

// autoEnableTenantProvisioningService is a narrow tenantcap fake for startup
// auto-enable provisioning tests.
type autoEnableTenantProvisioningService struct {
	enabled        bool
	pluginSvc      *serviceImpl
	tenantID       tenantcap.TenantID
	provisionCalls int
}

// ProvisionAutoEnabledTenantPlugins records startup provisioning and writes the
// tenant-scoped plugin state row through the same integration path used by menus.
func (s *autoEnableTenantProvisioningService) ProvisionAutoEnabledTenantPlugins(ctx context.Context) error {
	s.provisionCalls++
	if s.pluginSvc == nil {
		return nil
	}
	return s.pluginSvc.integrationSvc.SetTenantPluginEnabledState(ctx, autoEnableTenantProvisioningPluginID, int(s.tenantID), true)
}
