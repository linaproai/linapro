// This file verifies dependency governance at the plugin facade lifecycle
// boundaries rather than only inside the pure dependency resolver.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
)

// TestInstallAutoInstallsSourceDependencyBeforeTarget verifies the management
// install path installs automatic dependencies before installing the requested
// plugin and returns the auto-installed dependency list.
func TestInstallAutoInstallsSourceDependencyBeforeTarget(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.Background()
		dependencyID = "plugin-source-auto-dependency"
		targetID     = "plugin-source-auto-target"
	)

	createTestSourceDependencyPlugin(t, dependencyID, "Source Auto Dependency", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		targetID,
		"Source Auto Target",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+dependencyID+"\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: auto\n",
	)
	cleanupTestPluginIDs(t, ctx, dependencyID, targetID)

	result, err := service.Install(ctx, targetID, InstallOptions{})
	if err != nil {
		t.Fatalf("expected target install with automatic dependency to succeed, got error: %v", err)
	}
	if result == nil || len(result.AutoInstalled) != 1 || result.AutoInstalled[0].PluginID != dependencyID {
		t.Fatalf("expected auto-installed dependency %s, got %#v", dependencyID, result)
	}
	assertPluginInstalledState(t, ctx, service, dependencyID, catalog.InstalledYes, catalog.StatusDisabled)
	assertPluginInstalledState(t, ctx, service, targetID, catalog.InstalledYes, catalog.StatusDisabled)
}

// TestInstallBlocksDependencyViolationBeforeSideEffects verifies hard
// dependency failures stop the requested plugin before registry installation.
func TestInstallBlocksDependencyViolationBeforeSideEffects(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-source-manual-blocked"
	)

	createTestSourceDependencyPlugin(
		t,
		pluginID,
		"Source Manual Blocked",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: plugin-source-missing-manual\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: manual\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID, "plugin-source-missing-manual")

	_, err := service.Install(ctx, pluginID, InstallOptions{})
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected dependency blocked bizerr, got %v", err)
	}
	registry, lookupErr := service.getPluginRegistry(ctx, pluginID)
	if lookupErr != nil {
		t.Fatalf("expected blocked plugin registry lookup to succeed, got error: %v", lookupErr)
	}
	if registry != nil {
		t.Fatalf("expected dependency-blocked install to leave no registry side effect, got %#v", registry)
	}
}

// TestUninstallBlocksInstalledReverseHardDependency verifies destructive
// uninstall checks release snapshots for installed downstream dependencies.
func TestUninstallBlocksInstalledReverseHardDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		baseID     = "plugin-source-reverse-base"
		consumerID = "plugin-source-reverse-consumer"
	)

	createTestSourceDependencyPlugin(t, baseID, "Source Reverse Base", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		consumerID,
		"Source Reverse Consumer",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+baseID+"\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: auto\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install to auto-install base, got error: %v", err)
	}

	err := service.Uninstall(ctx, baseID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginReverseDependencyBlocked) {
		t.Fatalf("expected reverse dependency blocked bizerr, got %v", err)
	}
	assertPluginInstalledState(t, ctx, service, baseID, catalog.InstalledYes, catalog.StatusDisabled)
}

// TestCheckPluginDependenciesKeepsReverseBlockersOutOfInstallBlockers verifies
// management preflight responses keep install blockers and uninstall blockers
// in separate sections so install dialogs are not disabled by downstream users.
func TestCheckPluginDependenciesKeepsReverseBlockersOutOfInstallBlockers(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		baseID     = "plugin-source-check-reverse-base"
		consumerID = "plugin-source-check-reverse-consumer"
	)

	createTestSourceDependencyPlugin(t, baseID, "Source Check Reverse Base", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		consumerID,
		"Source Check Reverse Consumer",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+baseID+"\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: auto\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install to auto-install base, got error: %v", err)
	}

	result, err := service.CheckPluginDependencies(ctx, baseID)
	if err != nil {
		t.Fatalf("expected dependency check to succeed, got error: %v", err)
	}
	if len(result.Blockers) != 0 {
		t.Fatalf("expected install blockers to stay empty, got %#v", result.Blockers)
	}
	if len(result.ReverseDependents) != 1 || result.ReverseDependents[0].PluginID != consumerID {
		t.Fatalf("expected reverse dependent %s, got %#v", consumerID, result.ReverseDependents)
	}
	if len(result.ReverseBlockers) != 1 || result.ReverseBlockers[0].Code != "reverse_dependency" {
		t.Fatalf("expected reverse dependency blocker, got %#v", result.ReverseBlockers)
	}
}

// TestCheckPluginDependenciesExposesUnknownReverseSnapshotBlocker verifies
// management preflight responses expose conservative uninstall blockers even
// when no downstream dependent can be safely projected from the release snapshot.
func TestCheckPluginDependenciesExposesUnknownReverseSnapshotBlocker(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		baseID     = "plugin-source-check-unknown-base"
		consumerID = "plugin-source-check-unknown-consumer"
	)

	createTestSourceDependencyPlugin(t, baseID, "Source Check Unknown Base", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		consumerID,
		"Source Check Unknown Consumer",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+baseID+"\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: auto\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install to auto-install base, got error: %v", err)
	}
	consumerRegistry, err := service.getPluginRegistry(ctx, consumerID)
	if err != nil {
		t.Fatalf("expected consumer registry lookup to succeed, got error: %v", err)
	}
	if consumerRegistry == nil || consumerRegistry.ReleaseId <= 0 {
		t.Fatalf("expected installed consumer registry with release id, got %#v", consumerRegistry)
	}
	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: consumerRegistry.ReleaseId}).
		Data(do.SysPluginRelease{ManifestSnapshot: "not: [valid"}).
		Update(); err != nil {
		t.Fatalf("failed to corrupt consumer release snapshot for test: %v", err)
	}

	result, err := service.CheckPluginDependencies(ctx, baseID)
	if err != nil {
		t.Fatalf("expected dependency check to succeed, got error: %v", err)
	}
	if len(result.Blockers) != 0 {
		t.Fatalf("expected install blockers to stay empty, got %#v", result.Blockers)
	}
	if len(result.ReverseDependents) != 0 {
		t.Fatalf("expected unknown snapshot to hide unsafe dependent projection, got %#v", result.ReverseDependents)
	}
	if len(result.ReverseBlockers) != 1 || result.ReverseBlockers[0].Code != "dependency_snapshot_unknown" {
		t.Fatalf("expected unknown snapshot reverse blocker, got %#v", result.ReverseBlockers)
	}
}

// TestBootstrapAutoEnableInstallsDependencyWithoutImplicitEnable verifies
// startup dependency pre-install keeps dependency enablement separate from the
// explicit plugin.autoEnable target list.
func TestBootstrapAutoEnableInstallsDependencyWithoutImplicitEnable(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.Background()
		dependencyID = "plugin-source-bootstrap-dependency"
		targetID     = "plugin-source-bootstrap-target"
	)

	createTestSourceDependencyPlugin(t, dependencyID, "Source Bootstrap Dependency", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		targetID,
		"Source Bootstrap Target",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+dependencyID+"\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: auto\n",
	)
	configsvc.SetPluginAutoEnableOverride([]string{targetID})
	cleanupTestPluginIDs(t, ctx, dependencyID, targetID)
	t.Cleanup(func() {
		configsvc.SetPluginAutoEnableOverride(nil)
	})

	if err := service.BootstrapAutoEnable(ctx); err != nil {
		t.Fatalf("expected startup auto-enable with dependency to succeed, got error: %v", err)
	}
	assertPluginInstalledState(t, ctx, service, dependencyID, catalog.InstalledYes, catalog.StatusDisabled)
	assertPluginInstalledState(t, ctx, service, targetID, catalog.InstalledYes, catalog.StatusEnabled)
}

// TestSourcePluginUpgradeBlocksUnsatisfiedDependency verifies explicit source
// upgrades validate the candidate manifest before switching the effective
// release version.
func TestSourcePluginUpgradeBlocksUnsatisfiedDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-upgrade-dependency-block"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Dependency Block", oldVersion, "plugin:plugin-source-upgrade-dependency-block:old")
	cleanupTestPluginIDs(t, ctx, pluginID, "plugin-source-upgrade-missing")
	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected initial source sync to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial source install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifestWithExtra(
		t,
		manifestPath,
		pluginID,
		"Source Upgrade Dependency Block",
		newVersion,
		"plugin:plugin-source-upgrade-dependency-block:new",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: plugin-source-upgrade-missing\n"+
			"      version: \">=0.1.0\"\n"+
			"      required: true\n"+
			"      install: manual\n",
	)
	if err := service.SyncSourcePlugins(ctx); err != nil {
		t.Fatalf("expected source rescan to prepare candidate release, got error: %v", err)
	}

	_, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected dependency blocked upgrade error, got %v", err)
	}
	registry, lookupErr := service.getPluginRegistry(ctx, pluginID)
	if lookupErr != nil {
		t.Fatalf("expected registry lookup after blocked upgrade to succeed, got error: %v", lookupErr)
	}
	if registry == nil || registry.Version != oldVersion {
		t.Fatalf("expected effective version to stay %s after blocked upgrade, got %#v", oldVersion, registry)
	}
}

// TestDynamicPluginRefreshBlocksUnsatisfiedDependency verifies same-version
// dynamic refresh validates the staged manifest before changing the active
// release checksum or runtime assets.
func TestDynamicPluginRefreshBlocksUnsatisfiedDependency(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-refresh-dependency-block"
		version  = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})
	writeTestDynamicStorageArtifactWithDependencies(t, pluginID, "Dynamic Refresh Dependency Block", version, nil, buildVersionedRuntimeFrontendAssets("stable"))
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic install to succeed, got error: %v", err)
	}
	releaseBefore, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic release lookup before refresh to succeed, got error: %v", err)
	}
	if releaseBefore == nil {
		t.Fatal("expected dynamic release before refresh")
	}

	writeTestDynamicStorageArtifactWithDependencies(
		t,
		pluginID,
		"Dynamic Refresh Dependency Block",
		version,
		&catalog.DependencySpec{Plugins: []*catalog.PluginDependencySpec{
			testPluginDependencySpec("plugin-dynamic-refresh-missing", ">=0.1.0", true, catalog.DependencyInstallModeManual.String()),
		}},
		buildVersionedRuntimeFrontendAssets("blocked"),
	)

	_, err = service.Install(ctx, pluginID, InstallOptions{})
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected dependency blocked dynamic refresh error, got %v", err)
	}
	releaseAfter, err := service.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic release lookup after blocked refresh to succeed, got error: %v", err)
	}
	if releaseAfter == nil || releaseAfter.Checksum != releaseBefore.Checksum {
		t.Fatalf("expected release checksum to stay %s after blocked refresh, got %#v", releaseBefore.Checksum, releaseAfter)
	}
}

// createTestSourceDependencyPlugin writes a source plugin fixture with optional
// raw manifest suffix content appended after the standard governance fields.
func createTestSourceDependencyPlugin(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	extraManifest string,
) {
	t.Helper()

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: "+pluginName+"\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n"+
			extraManifest,
	)
}

// cleanupTestPluginIDs removes governance rows for all supplied plugin IDs and
// registers the same cleanup for the end of the current test.
func cleanupTestPluginIDs(t *testing.T, ctx context.Context, pluginIDs ...string) {
	t.Helper()

	for _, pluginID := range pluginIDs {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	}
	t.Cleanup(func() {
		for _, pluginID := range pluginIDs {
			testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		}
	})
}

// assertPluginInstalledState checks persisted plugin install and enable flags.
func assertPluginInstalledState(
	t *testing.T,
	ctx context.Context,
	service *serviceImpl,
	pluginID string,
	installed int,
	enabled int,
) {
	t.Helper()

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup for %s to succeed, got error: %v", pluginID, err)
	}
	if registry == nil {
		t.Fatalf("expected registry row for %s", pluginID)
	}
	if registry.Installed != installed || registry.Status != enabled {
		t.Fatalf("expected %s installed=%d enabled=%d, got installed=%d enabled=%d", pluginID, installed, enabled, registry.Installed, registry.Status)
	}
}

// writeTestSourcePluginManifestWithExtra writes a source-plugin manifest with
// the same menu convention as upgrade tests plus extra raw YAML content.
func writeTestSourcePluginManifestWithExtra(
	t *testing.T,
	manifestPath string,
	pluginID string,
	pluginName string,
	version string,
	menuKey string,
	extraManifest string,
) {
	t.Helper()

	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: "+pluginName+"\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n"+
			extraManifest+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: "+pluginName+"\n"+
			"    path: "+pluginID+"\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: "+pluginID+":view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)
}

// testPluginDependencySpec creates one dependency spec pointer with an explicit
// required flag so tests do not rely on catalog normalization side effects.
func testPluginDependencySpec(
	pluginID string,
	version string,
	required bool,
	install string,
) *catalog.PluginDependencySpec {
	return &catalog.PluginDependencySpec{
		ID:       pluginID,
		Version:  version,
		Required: &required,
		Install:  install,
	}
}

// writeTestDynamicStorageArtifactWithDependencies writes a dynamic plugin
// artifact whose manifest can carry dependency declarations.
func writeTestDynamicStorageArtifactWithDependencies(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	dependencies *catalog.DependencySpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
) string {
	t.Helper()

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	supportsMultiTenant := true
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                catalog.TypeDynamic.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
			Dependencies:        dependencies,
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        "wasm",
			ABIVersion:         "v1",
			FrontendAssetCount: len(frontendAssets),
		},
		frontendAssets,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove dynamic dependency artifact %s: %v", artifactPath, cleanupErr)
		}
	})
	return artifactPath
}
