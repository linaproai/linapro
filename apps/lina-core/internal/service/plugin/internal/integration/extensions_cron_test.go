// This file covers managed-cron collection behavior across source and dynamic
// plugin manifests.

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

// recordingDynamicCronExecutor captures which manifests are sent to dynamic
// cron discovery without executing any runtime code.
type recordingDynamicCronExecutor struct {
	discoverPluginIDs []string
	contracts         []*pluginbridge.CronContract
}

// DiscoverCronContracts records the manifest passed to dynamic discovery.
func (e *recordingDynamicCronExecutor) DiscoverCronContracts(
	_ context.Context,
	manifest *catalog.Manifest,
) ([]*pluginbridge.CronContract, error) {
	if manifest != nil {
		e.discoverPluginIDs = append(e.discoverPluginIDs, manifest.ID)
	}
	return e.contracts, nil
}

// ExecuteDeclaredCronJob is unused by this regression test.
func (e *recordingDynamicCronExecutor) ExecuteDeclaredCronJob(
	_ context.Context,
	_ *catalog.Manifest,
	_ *pluginbridge.CronContract,
) error {
	return nil
}

// rewriteRuntimeArtifactHostServices stores host-service declarations in the
// artifact itself so ScanManifests observes the same data production code uses.
func rewriteRuntimeArtifactHostServices(
	t *testing.T,
	artifactPath string,
	manifest *catalog.Manifest,
	hostServices []*pluginbridge.HostServiceSpec,
) {
	t.Helper()
	if manifest == nil || manifest.RuntimeArtifact == nil {
		t.Fatal("expected runtime manifest with artifact metadata")
	}
	manifest.RuntimeArtifact.HostServices = hostServices
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		manifest.RuntimeArtifact.Manifest,
		manifest.RuntimeArtifact,
		manifest.RuntimeArtifact.FrontendAssets,
		manifest.RuntimeArtifact.InstallSQLAssets,
		manifest.RuntimeArtifact.UninstallSQLAssets,
		nil,
		manifest.RuntimeArtifact.RouteContracts,
		manifest.RuntimeArtifact.BridgeSpec,
	)
}

// TestListExecutableCronJobsSkipsDynamicDiscoveryForSourcePlugins verifies source
// manifests keep using callback-based cron registration and are not sent to the
// dynamic Wasm cron-discovery path.
func TestListExecutableCronJobsSkipsDynamicDiscoveryForSourcePlugins(t *testing.T) {
	services := testutil.NewServices()
	executor := &recordingDynamicCronExecutor{}
	services.Integration.SetDynamicCronExecutor(executor)

	pluginID := "plugin-source-cron-dynamic-skip"
	testutil.CreateTestPluginDir(t, pluginID)

	manifests, err := services.Catalog.ScanManifests()
	if err != nil {
		t.Fatalf("expected manifest scan to succeed, got error: %v", err)
	}

	sourcePluginIDs := make(map[string]struct{})
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
			sourcePluginIDs[manifest.ID] = struct{}{}
		}
	}
	if len(sourcePluginIDs) == 0 {
		t.Fatal("expected at least one source plugin manifest in test repository")
	}

	items, err := services.Integration.ListExecutableCronJobs(context.Background())
	if err != nil {
		t.Fatalf("expected managed cron listing to succeed, got error: %v", err)
	}
	if !managedCronListContainsPlugin(items, pluginID) {
		t.Fatalf("expected managed cron list to include source plugin %s", pluginID)
	}

	for _, pluginID := range executor.discoverPluginIDs {
		if _, exists := sourcePluginIDs[pluginID]; exists {
			t.Fatalf("expected source plugin %s to skip dynamic cron discovery", pluginID)
		}
	}
}

// TestListExecutableCronJobsSkipsPendingUpgradeDynamicPlugin verifies dynamic cron
// declarations are not discovered while the plugin waits for runtime upgrade.
func TestListExecutableCronJobsSkipsPendingUpgradeDynamicPlugin(t *testing.T) {
	services := testutil.NewServices()
	executor := &recordingDynamicCronExecutor{}
	services.Integration.SetDynamicCronExecutor(executor)

	ctx := context.Background()
	const (
		pluginID   = "plugin-dynamic-cron-pending-upgrade"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Cron Pending Upgrade Plugin",
		oldVersion,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic cron manifest to load, got error: %v", err)
	}
	cronHostServices := []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceCron,
		Methods: []string{
			pluginbridge.HostServiceMethodCronRegister,
		},
	}}
	rewriteRuntimeArtifactHostServices(t, artifactPath, manifest, cronHostServices)
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	manifest.HostServices = cronHostServices
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic cron manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic cron plugin install state to be set, got error: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic cron plugin enable state to be set, got error: %v", err)
	}

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Cron Pending Upgrade Plugin",
		newVersion,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	newManifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic cron manifest to load, got error: %v", err)
	}
	rewriteRuntimeArtifactHostServices(t, artifactPath, newManifest, cronHostServices)
	newManifest.HostServices = cronHostServices
	if _, err = services.Catalog.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic cron manifest sync to succeed, got error: %v", err)
	}

	items, err := services.Integration.ListExecutableCronJobs(ctx)
	if err != nil {
		t.Fatalf("expected managed cron list to succeed, got error: %v", err)
	}
	if managedCronListContainsPlugin(items, pluginID) {
		t.Fatalf("expected pending-upgrade dynamic plugin %s to contribute no cron jobs", pluginID)
	}
	if len(executor.discoverPluginIDs) != 0 {
		t.Fatalf("expected pending-upgrade dynamic plugin to skip cron discovery, got %#v", executor.discoverPluginIDs)
	}
}

// TestListCronDeclarationsDiscoversDisabledDynamicPlugin verifies management
// review can display dynamic cron declarations before the plugin is enabled.
func TestListCronDeclarationsDiscoversDisabledDynamicPlugin(t *testing.T) {
	services := testutil.NewServices()
	executor := &recordingDynamicCronExecutor{
		contracts: []*pluginbridge.CronContract{
			{
				Name:           "heartbeat",
				DisplayName:    "Dynamic Plugin Heartbeat",
				Description:    "Runs a dynamic heartbeat.",
				Pattern:        "# */10 * * * *",
				Timezone:       pluginbridge.DefaultCronContractTimezone,
				Scope:          pluginbridge.CronScopeAllNode,
				Concurrency:    pluginbridge.CronConcurrencySingleton,
				MaxConcurrency: 1,
				TimeoutSeconds: 30,
				InternalPath:   "/cron-heartbeat",
			},
		},
	}
	services.Integration.SetDynamicCronExecutor(executor)

	ctx := context.Background()
	const pluginID = "plugin-dynamic-cron-review"
	const sourcePluginID = "plugin-source-cron-uninstalled-review"
	testutil.CreateTestPluginDir(t, sourcePluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, sourcePluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, sourcePluginID)
	})
	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Cron Review Plugin",
		"v0.1.0",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic cron manifest to load, got error: %v", err)
	}
	cronHostServices := []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceCron,
		Methods: []string{
			pluginbridge.HostServiceMethodCronRegister,
		},
	}}
	rewriteRuntimeArtifactHostServices(t, artifactPath, manifest, cronHostServices)
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	manifest.HostServices = cronHostServices
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic cron manifest sync to succeed, got error: %v", err)
	}

	managedItems, err := services.Integration.ListExecutableCronJobsByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected executable managed cron list to succeed, got error: %v", err)
	}
	if len(managedItems) != 0 {
		t.Fatalf("expected disabled dynamic plugin to expose no executable cron jobs, got %#v", managedItems)
	}
	installedItems, err := services.Integration.ListInstalledCronDeclarations(ctx)
	if err != nil {
		t.Fatalf("expected installed declaration cron list to succeed, got error: %v", err)
	}
	if managedCronListContainsPlugin(installedItems, pluginID) {
		t.Fatalf("expected uninstalled dynamic plugin %s to expose no scheduled-job declarations, got %#v", pluginID, installedItems)
	}
	if managedCronListContainsPlugin(installedItems, sourcePluginID) {
		t.Fatalf("expected uninstalled source plugin %s to expose no scheduled-job declarations, got %#v", sourcePluginID, installedItems)
	}

	declaredItems, err := services.Integration.ListCronDeclarationsByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected declared cron list to succeed, got error: %v", err)
	}
	if len(declaredItems) != 1 {
		t.Fatalf("expected one declared cron item, got %#v", declaredItems)
	}
	if declaredItems[0].Name != "heartbeat" || declaredItems[0].Pattern != "# */10 * * * *" {
		t.Fatalf("unexpected declared cron item: %#v", declaredItems[0])
	}
	if len(executor.discoverPluginIDs) != 1 || executor.discoverPluginIDs[0] != pluginID {
		t.Fatalf("expected disabled dynamic plugin to be discovered for review only, got %#v", executor.discoverPluginIDs)
	}
}

// TestListInstalledCronDeclarationsDiscoversInstalledDisabledDynamicPlugin
// verifies scheduled-job projection can show installed dynamic cron jobs before
// the plugin business entries are enabled.
func TestListInstalledCronDeclarationsDiscoversInstalledDisabledDynamicPlugin(t *testing.T) {
	services := testutil.NewServices()
	executor := &recordingDynamicCronExecutor{
		contracts: []*pluginbridge.CronContract{
			{
				Name:           "heartbeat",
				DisplayName:    "Dynamic Plugin Heartbeat",
				Description:    "Runs a dynamic heartbeat.",
				Pattern:        "# */10 * * * *",
				Timezone:       pluginbridge.DefaultCronContractTimezone,
				Scope:          pluginbridge.CronScopeAllNode,
				Concurrency:    pluginbridge.CronConcurrencySingleton,
				MaxConcurrency: 1,
				TimeoutSeconds: 30,
				InternalPath:   "/cron-heartbeat",
			},
		},
	}
	services.Integration.SetDynamicCronExecutor(executor)

	ctx := context.Background()
	const pluginID = "plugin-dynamic-cron-installed-review"
	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Cron Installed Review Plugin",
		"v0.1.0",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic cron manifest to load, got error: %v", err)
	}
	cronHostServices := []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceCron,
		Methods: []string{
			pluginbridge.HostServiceMethodCronRegister,
		},
	}}
	rewriteRuntimeArtifactHostServices(t, artifactPath, manifest, cronHostServices)
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	manifest.HostServices = cronHostServices
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic cron manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic cron plugin install state to be set, got error: %v", err)
	}

	installedItems, err := services.Integration.ListInstalledCronDeclarations(ctx)
	if err != nil {
		t.Fatalf("expected installed declaration cron list to succeed, got error: %v", err)
	}
	if !managedCronListContainsPlugin(installedItems, pluginID) {
		t.Fatalf("expected installed disabled dynamic plugin %s to expose scheduled-job declarations, got %#v", pluginID, installedItems)
	}
	if len(executor.discoverPluginIDs) != 1 || executor.discoverPluginIDs[0] != pluginID {
		t.Fatalf("expected installed disabled dynamic plugin to be discovered once for projection, got %#v", executor.discoverPluginIDs)
	}
}

// TestListCronDeclarationsDiscoversBundledDynamicSample verifies the plugin
// list authorization preview path can execute the real bundled Wasm sample's
// cron registration route before the plugin is installed.
func TestListCronDeclarationsDiscoversBundledDynamicSample(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	ctx := context.Background()
	const pluginID = "plugin-demo-dynamic"
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	storageArtifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName(pluginID))
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("expected repo root to resolve, got error: %v", err)
	}
	uploadArtifactPath := filepath.Join(repoRoot, "temp", "output", runtime.BuildArtifactFileName(pluginID))
	if _, err = os.Stat(uploadArtifactPath); err == nil {
		content, readErr := os.ReadFile(uploadArtifactPath)
		if readErr != nil {
			t.Fatalf("expected E2E-built runtime artifact to read, got error: %v", readErr)
		}
		if writeErr := os.WriteFile(storageArtifactPath, content, 0o644); writeErr != nil {
			t.Fatalf("expected E2E-built runtime artifact to copy into test storage, got error: %v", writeErr)
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("expected E2E-built runtime artifact stat to succeed or miss, got error: %v", err)
	}

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(storageArtifactPath)
	if err != nil {
		t.Fatalf("expected bundled dynamic manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected bundled dynamic manifest sync to succeed, got error: %v", err)
	}

	declaredItems, err := services.Integration.ListCronDeclarationsByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected bundled dynamic cron declaration list to succeed, got error: %v", err)
	}
	if len(declaredItems) != 1 {
		t.Fatalf("expected one bundled dynamic cron declaration, got %#v", declaredItems)
	}
	if declaredItems[0].Name != "heartbeat" || declaredItems[0].Pattern != "# */10 * * * *" {
		t.Fatalf("unexpected bundled dynamic cron declaration: %#v", declaredItems[0])
	}
}

// managedCronListContainsPlugin reports whether a managed cron list includes
// at least one definition owned by pluginID.
func managedCronListContainsPlugin(items []integration.ManagedCronJob, pluginID string) bool {
	for _, item := range items {
		if item.PluginID == pluginID {
			return true
		}
	}
	return false
}
