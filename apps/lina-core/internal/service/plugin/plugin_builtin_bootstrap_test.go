// This file verifies startup reconciliation for distribution=builtin source
// plugins, including lifecycle convergence, mock-data exclusion, safe upgrade,
// abnormal downgrade failure, and cluster primary handoff behavior.

package plugin

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
)

// TestBootstrapBuiltinPluginsInstallsAndEnablesSourcePlugin verifies builtin
// source plugins converge without plugin.autoEnable configuration.
func TestBootstrapBuiltinPluginsInstallsAndEnablesSourcePlugin(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dev-builtin-bootstrap-install"
	)

	createTestSourceDependencyPlugin(
		t,
		pluginID,
		"Builtin Bootstrap Install",
		"v0.1.0",
		"distribution: builtin\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID)

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected builtin bootstrap to succeed, got error: %v", err)
	}

	assertPluginInstalledState(t, ctx, service, pluginID, plugintypes.InstalledYes, plugintypes.StatusEnabled)
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected builtin current state enabled, got %s", registry.CurrentState)
	}
}

// TestBootstrapBuiltinPluginsDoesNotLoadMockData verifies builtin startup
// install never loads manifest/sql/mock-data even when the plugin ships it.
func TestBootstrapBuiltinPluginsDoesNotLoadMockData(t *testing.T) {
	var (
		ctx       = context.Background()
		service   = newTestService()
		pluginID  = "plugin-dev-builtin-bootstrap-no-mock"
		mockTable = "plugin_builtin_bootstrap_no_mock_demo"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: Builtin Bootstrap No Mock\n"+
			"version: v0.1.0\n"+
			"type: source\n"+
			"distribution: builtin\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"),
		"CREATE TABLE IF NOT EXISTS "+mockTable+" (id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, marker VARCHAR(32) NOT NULL);",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "mock-data", "001-"+pluginID+"-mock.sql"),
		"INSERT INTO "+mockTable+" (marker) VALUES ('builtin-mock-row');",
	)

	dropMockTableIfExists(t, ctx, mockTable)
	cleanupTestPluginIDs(t, ctx, pluginID)
	t.Cleanup(func() {
		dropMockTableIfExists(t, ctx, mockTable)
	})

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected builtin bootstrap to succeed, got error: %v", err)
	}
	if rows := mockTableRowCount(t, ctx, mockTable); rows != 0 {
		t.Fatalf("expected builtin bootstrap not to load mock-data rows, got %d", rows)
	}
	if mockMigrations := mockMigrationRowCount(t, ctx, pluginID); mockMigrations != 0 {
		t.Fatalf("expected no mock migration records for builtin bootstrap, got %d", mockMigrations)
	}
}

// TestBootstrapBuiltinPluginsUpgradesPendingSourcePlugin verifies startup
// builtin reconciliation executes the unified runtime-upgrade path.
func TestBootstrapBuiltinPluginsUpgradesPendingSourcePlugin(t *testing.T) {
	var (
		ctx         = context.Background()
		service     = newTestService()
		pluginID    = "plugin-dev-builtin-bootstrap-upgrade"
		oldVersion  = "v0.1.0"
		newVersion  = "v0.2.0"
		manifestDir = testutil.CreateTestPluginDir(t, pluginID)
		manifest    = filepath.Join(manifestDir, "plugin.yaml")
	)

	writeTestSourcePluginManifestWithExtra(
		t,
		manifest,
		pluginID,
		"Builtin Bootstrap Upgrade",
		oldVersion,
		"plugin:"+pluginID+":old",
		"distribution: builtin\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID)

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected initial builtin bootstrap to succeed, got error: %v", err)
	}
	writeTestSourcePluginManifestWithExtra(
		t,
		manifest,
		pluginID,
		"Builtin Bootstrap Upgrade",
		newVersion,
		"plugin:"+pluginID+":new",
		"distribution: builtin\n",
	)

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected builtin bootstrap upgrade to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry.Version != newVersion {
		t.Fatalf("expected builtin effective version %s after startup upgrade, got %s", newVersion, registry.Version)
	}
	assertPluginInstalledState(t, ctx, service, pluginID, plugintypes.InstalledYes, plugintypes.StatusEnabled)
}

// TestBootstrapBuiltinPluginsFailsOnDiscoveredDowngrade verifies lower source
// versions fail startup and do not demote the effective release.
func TestBootstrapBuiltinPluginsFailsOnDiscoveredDowngrade(t *testing.T) {
	var (
		ctx         = context.Background()
		service     = newTestService()
		pluginID    = "plugin-dev-builtin-bootstrap-downgrade"
		oldVersion  = "v0.1.0"
		newVersion  = "v0.2.0"
		manifestDir = testutil.CreateTestPluginDir(t, pluginID)
		manifest    = filepath.Join(manifestDir, "plugin.yaml")
	)

	writeTestSourcePluginManifestWithExtra(
		t,
		manifest,
		pluginID,
		"Builtin Bootstrap Downgrade",
		oldVersion,
		"plugin:"+pluginID+":old",
		"distribution: builtin\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID)

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected initial builtin bootstrap to succeed, got error: %v", err)
	}
	writeTestSourcePluginManifestWithExtra(
		t,
		manifest,
		pluginID,
		"Builtin Bootstrap Downgrade",
		newVersion,
		"plugin:"+pluginID+":new",
		"distribution: builtin\n",
	)
	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected builtin upgrade bootstrap to succeed, got error: %v", err)
	}
	writeTestSourcePluginManifestWithExtra(
		t,
		manifest,
		pluginID,
		"Builtin Bootstrap Downgrade",
		oldVersion,
		"plugin:"+pluginID+":old",
		"distribution: builtin\n",
	)

	err := service.BootstrapBuiltinPlugins(ctx)
	if !bizerr.Is(err, CodePluginRuntimeUpgradeUnavailable) {
		t.Fatalf("expected runtime-upgrade-unavailable error on builtin downgrade, got %v", err)
	}
	registry, lookupErr := service.getPluginRegistry(ctx, pluginID)
	if lookupErr != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", lookupErr)
	}
	if registry.Version != newVersion {
		t.Fatalf("expected effective version to remain %s after downgrade failure, got %s", newVersion, registry.Version)
	}
}

// TestBootstrapBuiltinPluginsWaitsUntilCurrentNodeBecomesPrimary verifies
// cluster startup waits before executing shared builtin lifecycle writes.
func TestBootstrapBuiltinPluginsWaitsUntilCurrentNodeBecomesPrimary(t *testing.T) {
	var (
		ctx      = context.Background()
		pluginID = "plugin-dev-builtin-bootstrap-cluster"
		topology = &testTopology{
			enabled: true,
			primary: false,
			nodeID:  "builtin-bootstrap-follower",
		}
		service = newTestServiceWithTopology(topology)
	)

	createTestSourceDependencyPlugin(
		t,
		pluginID,
		"Builtin Bootstrap Cluster",
		"v0.1.0",
		"distribution: builtin\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID)
	config.SetPluginAutoEnableOverride(nil)
	t.Cleanup(func() {
		config.SetPluginAutoEnableOverride(nil)
	})

	timer := time.AfterFunc(150*time.Millisecond, func() {
		topology.SetPrimary(true)
	})
	t.Cleanup(func() {
		timer.Stop()
	})

	if err := service.BootstrapBuiltinPlugins(ctx); err != nil {
		t.Fatalf("expected clustered builtin bootstrap to succeed after primary handoff, got error: %v", err)
	}
	assertPluginInstalledState(t, ctx, service, pluginID, plugintypes.InstalledYes, plugintypes.StatusEnabled)
}
