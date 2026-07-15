// This file verifies dependency governance at the plugin facade lifecycle
// boundaries rather than only inside the pure dependency resolver.

package plugin

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// TestInstallBlocksUninstalledSourceDependency verifies the management install
// path requires hard dependencies to be installed explicitly before the target.
func TestInstallBlocksUninstalledSourceDependency(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.Background()
		dependencyID = "plugin-dev-source-hard-dependency"
		targetID     = "plugin-dev-source-hard-target"
	)

	createTestSourceDependencyPlugin(t, dependencyID, "Source Hard Dependency", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		targetID,
		"Source Hard Target",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+dependencyID+"\n"+
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, dependencyID, targetID)

	_, err := service.Install(ctx, targetID, InstallOptions{})
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected uninstalled dependency to block target install, got error: %v", err)
	}
	assertPluginNoRegistryRow(t, ctx, service, targetID)
	if _, err = service.Install(ctx, dependencyID, InstallOptions{}); err != nil {
		t.Fatalf("expected explicit dependency install to succeed, got error: %v", err)
	}
	result, err := service.Install(ctx, targetID, InstallOptions{})
	if err != nil {
		t.Fatalf("expected target install after explicit dependency install to succeed, got error: %v", err)
	}
	if result == nil || len(result.Blockers) != 0 {
		t.Fatalf("expected successful dependency check without blockers, got %#v", result)
	}
	assertPluginInstalledState(t, ctx, service, dependencyID, statusflag.Installed.Int(), statusflag.Disabled.Int())
	assertPluginInstalledState(t, ctx, service, targetID, statusflag.Installed.Int(), statusflag.Disabled.Int())
}

// TestInstallBlocksDependencyViolationBeforeSideEffects verifies hard
// dependency failures stop the requested plugin before registry installation.
func TestInstallBlocksDependencyViolationBeforeSideEffects(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-source-manual-blocked"
	)

	createTestSourceDependencyPlugin(
		t,
		pluginID,
		"Source Manual Blocked",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: plugin-dev-source-missing-manual\n"+
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, pluginID, "plugin-dev-source-missing-manual")

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
		baseID     = "plugin-dev-source-reverse-base"
		consumerID = "plugin-dev-source-reverse-consumer"
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
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, baseID, InstallOptions{}); err != nil {
		t.Fatalf("expected base install to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install after base install, got error: %v", err)
	}

	err := service.Uninstall(ctx, baseID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginReverseDependencyBlocked) {
		t.Fatalf("expected reverse dependency blocked bizerr, got %v", err)
	}
	assertPluginInstalledState(t, ctx, service, baseID, statusflag.Installed.Int(), statusflag.Disabled.Int())
}

// TestDisableBlocksEnabledReverseOwnerHostServiceDependency verifies owner
// plugin disable is blocked before it can break enabled owner-aware dynamic
// host service consumers.
func TestDisableBlocksEnabledReverseOwnerHostServiceDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		ownerID    = "plugin-dev-source-owner-disable-base"
		consumerID = "plugin-dev-dynamic-owner-disable-consumer"
	)

	createTestSourceDependencyPlugin(t, ownerID, "Source Owner Disable Base", "v0.1.0", "")
	writeTestDynamicOwnerHostServiceArtifactWithDependencies(
		t,
		consumerID,
		"Dynamic Owner Disable Consumer",
		"v0.1.0",
		&plugintypes.DependencySpec{Plugins: []*plugintypes.PluginDependencySpec{
			testPluginDependencySpec(ownerID, ">=0.1.0"),
		}},
		[]*protocol.HostServiceSpec{{
			Owner:   ownerID,
			Service: "ai",
			Version: "v1",
			Methods: []string{
				"text.generate",
			},
		}},
		buildVersionedRuntimeFrontendAssets("owner-disable-consumer"),
	)
	cleanupTestPluginIDs(t, ctx, ownerID, consumerID)

	if _, err := service.Install(ctx, ownerID, InstallOptions{}); err != nil {
		t.Fatalf("expected owner install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, ownerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected owner enable to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install after owner install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, consumerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected consumer enable after owner install to succeed, got error: %v", err)
	}

	result, err := service.CheckPluginDependencies(ctx, ownerID)
	if err != nil {
		t.Fatalf("expected owner dependency check to succeed, got error: %v", err)
	}
	if len(result.ReverseDependents) != 1 || result.ReverseDependents[0].PluginID != consumerID {
		t.Fatalf("expected reverse dependent %s, got %#v", consumerID, result.ReverseDependents)
	}
	hostServices := result.ReverseDependents[0].OwnerHostServices
	if len(hostServices) != 1 ||
		hostServices[0].Owner != ownerID ||
		hostServices[0].Service != "ai" ||
		hostServices[0].Version != "v1" ||
		len(hostServices[0].Methods) != 1 ||
		hostServices[0].Methods[0] != "text.generate" {
		t.Fatalf("expected owner host service summary for %s, got %#v", ownerID, hostServices)
	}

	err = service.UpdateStatus(ctx, ownerID, UpdateStatusOptions{Status: statusflag.Disabled.Int()})
	if !bizerr.Is(err, CodePluginReverseEnabledDependencyBlocked) {
		t.Fatalf("expected enabled reverse dependency blocked owner disable, got %v", err)
	}
	assertPluginInstalledState(t, ctx, service, ownerID, statusflag.Installed.Int(), statusflag.EnabledValue.Int())
}

// TestDisableAllowsInstalledButDisabledReverseDependents verifies disable uses
// the runtime axis and does not require uninstalling disabled dependents first.
func TestDisableAllowsInstalledButDisabledReverseDependents(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		ownerID    = "plugin-dev-source-owner-disable-relaxed"
		consumerID = "plugin-dev-source-consumer-disable-relaxed"
	)

	createTestSourceDependencyPlugin(t, ownerID, "Source Owner Disable Relaxed", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		consumerID,
		"Source Consumer Disable Relaxed",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+ownerID+"\n"+
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, ownerID, consumerID)

	if _, err := service.Install(ctx, ownerID, InstallOptions{}); err != nil {
		t.Fatalf("expected owner install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, ownerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected owner enable to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, consumerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected consumer enable to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, consumerID, UpdateStatusOptions{Status: statusflag.Disabled.Int()}); err != nil {
		t.Fatalf("expected consumer disable to succeed, got error: %v", err)
	}

	if err := service.UpdateStatus(ctx, ownerID, UpdateStatusOptions{Status: statusflag.Disabled.Int()}); err != nil {
		t.Fatalf("expected owner disable after disabled dependents, got error: %v", err)
	}
	assertPluginInstalledState(t, ctx, service, ownerID, statusflag.Installed.Int(), statusflag.Disabled.Int())
	assertPluginInstalledState(t, ctx, service, consumerID, statusflag.Installed.Int(), statusflag.Disabled.Int())

	err := service.Uninstall(ctx, ownerID, UninstallOptions{PurgeStorageData: true})
	if !bizerr.Is(err, CodePluginReverseDependencyBlocked) {
		t.Fatalf("expected uninstall still blocked by installed dependent, got %v", err)
	}
}

// TestEnableBlocksInstalledButDisabledHardDependency verifies enable requires
// hard dependencies to be enabled, not only installed.
func TestEnableBlocksInstalledButDisabledHardDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		baseID     = "plugin-dev-source-enable-axis-base"
		consumerID = "plugin-dev-source-enable-axis-consumer"
	)

	createTestSourceDependencyPlugin(t, baseID, "Source Enable Axis Base", "v0.1.0", "")
	createTestSourceDependencyPlugin(
		t,
		consumerID,
		"Source Enable Axis Consumer",
		"v0.1.0",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: "+baseID+"\n"+
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, baseID, InstallOptions{}); err != nil {
		t.Fatalf("expected base install to succeed, got error: %v", err)
	}
	// Base remains installed+disabled.
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install with disabled dependency, got error: %v", err)
	}
	err := service.UpdateStatus(ctx, consumerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()})
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected enable blocked by disabled dependency, got %v", err)
	}
	assertPluginInstalledState(t, ctx, service, consumerID, statusflag.Installed.Int(), statusflag.Disabled.Int())
}

// TestSourceOwnerUpgradeBlocksInstalledReverseOwnerHostServiceDependency
// verifies source owner upgrades are blocked when they would break an installed
// dynamic consumer's owner-aware host service dependency range.
func TestSourceOwnerUpgradeBlocksInstalledReverseOwnerHostServiceDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		ownerID    = "plugin-dev-source-owner-upgrade-base"
		consumerID = "plugin-dev-dynamic-owner-upgrade-consumer"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, ownerID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifestWithExtra(
		t,
		manifestPath,
		ownerID,
		"Source Owner Upgrade Base",
		oldVersion,
		"plugin:plugin-dev-source-owner-upgrade-base:old",
		"",
	)
	writeTestDynamicOwnerHostServiceArtifactWithDependencies(
		t,
		consumerID,
		"Dynamic Owner Upgrade Consumer",
		oldVersion,
		&plugintypes.DependencySpec{Plugins: []*plugintypes.PluginDependencySpec{
			testPluginDependencySpec(ownerID, ">=0.1.0 <0.2.0"),
		}},
		[]*protocol.HostServiceSpec{{
			Owner:   ownerID,
			Service: "ai",
			Version: "v1",
			Methods: []string{
				"text.generate",
			},
		}},
		buildVersionedRuntimeFrontendAssets("owner-upgrade-consumer"),
	)
	cleanupTestPluginIDs(t, ctx, ownerID, consumerID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, ownerID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, ownerID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected initial source sync to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, ownerID, InstallOptions{}); err != nil {
		t.Fatalf("expected owner install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, ownerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected owner enable to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install after owner install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, consumerID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected consumer enable after owner install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifestWithExtra(
		t,
		manifestPath,
		ownerID,
		"Source Owner Upgrade Base",
		newVersion,
		"plugin:plugin-dev-source-owner-upgrade-base:new",
		"",
	)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source rescan to prepare owner upgrade candidate, got error: %v", err)
	}

	_, err := service.UpgradeSourcePlugin(ctx, ownerID)
	if !bizerr.Is(err, CodePluginReverseDependencyBlocked) {
		t.Fatalf("expected reverse dependency blocked owner upgrade, got %v", err)
	}
	assertPluginInstalledState(t, ctx, service, ownerID, statusflag.Installed.Int(), statusflag.EnabledValue.Int())
	registry, lookupErr := service.getPluginRegistry(ctx, ownerID)
	if lookupErr != nil {
		t.Fatalf("expected owner registry lookup after blocked upgrade to succeed, got error: %v", lookupErr)
	}
	if registry == nil || registry.Version != oldVersion {
		t.Fatalf("expected owner effective version to stay %s after blocked upgrade, got %#v", oldVersion, registry)
	}
}

// TestCheckPluginDependenciesKeepsReverseBlockersOutOfInstallBlockers verifies
// management preflight responses keep install blockers and uninstall blockers
// in separate sections so install dialogs are not disabled by downstream users.
func TestCheckPluginDependenciesKeepsReverseBlockersOutOfInstallBlockers(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		baseID     = "plugin-dev-source-check-reverse-base"
		consumerID = "plugin-dev-source-check-reverse-consumer"
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
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, baseID, InstallOptions{}); err != nil {
		t.Fatalf("expected base install to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install after base install, got error: %v", err)
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
		baseID     = "plugin-dev-source-check-unknown-base"
		consumerID = "plugin-dev-source-check-unknown-consumer"
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
			"      version: \">=0.1.0\"\n",
	)
	cleanupTestPluginIDs(t, ctx, baseID, consumerID)

	if _, err := service.Install(ctx, baseID, InstallOptions{}); err != nil {
		t.Fatalf("expected base install to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, consumerID, InstallOptions{}); err != nil {
		t.Fatalf("expected consumer install after base install, got error: %v", err)
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

// TestCheckPluginDependenciesIgnoresStaleRegistryRowsWithoutRelease verifies
// registry rows that have no discovered manifest and no release snapshot do not
// block unrelated plugin lifecycle checks.
func TestCheckPluginDependenciesIgnoresStaleRegistryRowsWithoutRelease(t *testing.T) {
	var (
		service       = newTestService()
		ctx           = context.Background()
		pluginID      = "plugin-dev-source-stale-row-target"
		stalePluginID = "platform-only-plugin-stale-row"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Legacy Row Target", "v0.1.0", "")
	cleanupTestPluginIDs(t, ctx, pluginID, stalePluginID)
	if _, err := dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:     stalePluginID,
		Name:         "Stale Plugin Row",
		Version:      "v0.1.0",
		Type:         pluginv1.PluginTypeSource.String(),
		Installed:    statusflag.Installed.Int(),
		Status:       statusflag.EnabledValue.Int(),
		DesiredState: plugintypes.HostStateEnabled.String(),
		CurrentState: plugintypes.HostStateEnabled.String(),
		Generation:   int64(1),
	}).Insert(); err != nil {
		t.Fatalf("expected stale registry row insert to succeed, got error: %v", err)
	}

	result, err := service.CheckPluginDependencies(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected dependency check to ignore stale registry row, got error: %v", err)
	}
	if len(result.ReverseBlockers) != 0 {
		t.Fatalf("expected no reverse blockers from stale registry row, got %#v", result.ReverseBlockers)
	}
}

// TestBootstrapAutoEnableBlocksUninstalledDependency verifies startup
// auto-enable does not install dependencies implicitly from plugin manifests.
func TestBootstrapAutoEnableBlocksUninstalledDependency(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.Background()
		dependencyID = "plugin-dev-source-bootstrap-dependency"
		targetID     = "plugin-dev-source-bootstrap-target"
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
			"      version: \">=0.1.0\"\n",
	)
	setTestPluginAutoEnableOverride([]string{targetID})
	cleanupTestPluginIDs(t, ctx, dependencyID, targetID)
	t.Cleanup(func() {
		setTestPluginAutoEnableOverride(nil)
	})

	err := service.BootstrapAutoEnable(ctx)
	if !bizerr.Is(err, CodePluginDependencyBlocked) {
		t.Fatalf("expected startup auto-enable to block on uninstalled dependency, got error: %v", err)
	}
	assertPluginInstalledState(t, ctx, service, targetID, statusflag.Uninstalled.Int(), statusflag.Disabled.Int())
}

// TestSourcePluginUpgradeBlocksUnsatisfiedDependency verifies explicit source
// upgrades validate the candidate manifest before switching the effective
// release version.
func TestSourcePluginUpgradeBlocksUnsatisfiedDependency(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-dependency-block"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Dependency Block", oldVersion, "plugin:plugin-dev-source-upgrade-dependency-block:old")
	cleanupTestPluginIDs(t, ctx, pluginID, "plugin-dev-source-upgrade-missing")
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
		"plugin:plugin-dev-source-upgrade-dependency-block:new",
		"dependencies:\n"+
			"  plugins:\n"+
			"    - id: plugin-dev-source-upgrade-missing\n"+
			"      version: \">=0.1.0\"\n",
	)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
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
		pluginID = "plugin-dev-dynamic-refresh-dependency-block"
		version  = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})
	writeTestDynamicStorageArtifactWithDependencies(t, pluginID, "Dynamic Refresh Dependency Block", version, nil, buildVersionedRuntimeFrontendAssets("stable"))
	service.catalogSvc.InvalidateManifestCache(pluginID)
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
		&plugintypes.DependencySpec{Plugins: []*plugintypes.PluginDependencySpec{
			testPluginDependencySpec("plugin-dev-dynamic-refresh-missing", ">=0.1.0"),
		}},
		buildVersionedRuntimeFrontendAssets("blocked"),
	)
	service.catalogSvc.InvalidateManifestCache(pluginID)

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

// TestDependencySnapshotCacheReusesCatalogSnapshot verifies repeated dependency
// checks in one list projection reuse request-local dependency snapshots.
func TestDependencySnapshotCacheReusesCatalogSnapshot(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-dependency-cache"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Dependency Cache", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	readCtx := service.WithDependencySnapshotCache(ctx)
	first, err := service.buildDependencySnapshots(readCtx, nil)
	if err != nil {
		t.Fatalf("build first dependency snapshots: %v", err)
	}
	firstSnapshot := findDependencySnapshotForTest(first, pluginID)
	if firstSnapshot == nil {
		t.Fatalf("expected first dependency snapshot for %s", pluginID)
	}
	firstSnapshot.Name = "mutated by caller"

	second, err := service.buildDependencySnapshots(readCtx, nil)
	if err != nil {
		t.Fatalf("build second dependency snapshots: %v", err)
	}
	secondSnapshot := findDependencySnapshotForTest(second, pluginID)
	if secondSnapshot == nil {
		t.Fatalf("expected second dependency snapshot for %s", pluginID)
	}
	if secondSnapshot.Name == firstSnapshot.Name {
		t.Fatalf("expected cached dependency snapshots to be cloned, got mutated name %q", secondSnapshot.Name)
	}

	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected one catalog snapshot build with dependency cache, got %d", got)
	}
}

// TestCheckPluginDependenciesUsesOneRequestSnapshot verifies the public
// dependency preflight reuses one request-local snapshot for install and reverse checks.
func TestCheckPluginDependenciesUsesOneRequestSnapshot(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = startupstats.WithCollector(context.Background(), startupstats.New())
		pluginID = "plugin-dev-source-dependency-check-cache"
	)

	createTestSourceDependencyPlugin(t, pluginID, "Source Dependency Check Cache", "v0.1.0", "")
	cleanupTestPluginIDs(t, context.Background(), pluginID)

	if _, err := service.CheckPluginDependencies(ctx, pluginID); err != nil {
		t.Fatalf("expected dependency check to succeed, got error: %v", err)
	}
	snapshot := startupstats.FromContext(ctx).Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds); got != 1 {
		t.Fatalf("expected dependency check to build one shared snapshot, got %d", got)
	}
}

// findDependencySnapshotForTest returns one dependency snapshot by plugin ID.
func findDependencySnapshotForTest(items []*plugindep.PluginSnapshot, pluginID string) *plugindep.PluginSnapshot {
	for _, item := range items {
		if item != nil && item.ID == pluginID {
			return item
		}
	}
	return nil
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

// assertPluginNoRegistryRow checks that a blocked lifecycle path left no plugin
// registry side effect.
func assertPluginNoRegistryRow(
	t *testing.T,
	ctx context.Context,
	service *serviceImpl,
	pluginID string,
) {
	t.Helper()

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup for %s to succeed, got error: %v", pluginID, err)
	}
	if registry != nil {
		t.Fatalf("expected no registry row for %s, got %#v", pluginID, registry)
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

// testPluginDependencySpec creates one plugin dependency spec pointer.
func testPluginDependencySpec(
	pluginID string,
	version string,
) *plugintypes.PluginDependencySpec {
	return &plugintypes.PluginDependencySpec{
		ID:      pluginID,
		Version: version,
	}
}

// writeTestDynamicStorageArtifactWithDependencies writes a dynamic plugin
// artifact whose manifest can carry dependency declarations.
func writeTestDynamicStorageArtifactWithDependencies(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	dependencies *plugintypes.DependencySpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
) string {
	t.Helper()

	return writeTestDynamicOwnerHostServiceArtifactWithDependencies(
		t,
		pluginID,
		pluginName,
		version,
		dependencies,
		nil,
		frontendAssets,
	)
}

// writeTestDynamicOwnerHostServiceArtifactWithDependencies writes a dynamic
// plugin artifact whose effective release carries dependencies and owner
// host service declarations.
func writeTestDynamicOwnerHostServiceArtifactWithDependencies(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	dependencies *plugintypes.DependencySpec,
	hostServices []*protocol.HostServiceSpec,
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
			Type:                pluginv1.PluginTypeDynamic.String(),
			ScopeNature:         pluginv1.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supportsMultiTenant,
			DefaultInstallMode:  pluginv1.InstallModeTenantScoped.String(),
			Dependencies:        dependencies,
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        "wasm",
			ABIVersion:         "v1",
			FrontendAssetCount: len(frontendAssets),
			HostServices:       hostServices,
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
