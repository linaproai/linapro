// This file covers managed-cron collection behavior across source and dynamic
// plugin manifests.

package integration_test

import (
	"context"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

// recordingDynamicCronExecutor captures which manifests are sent to dynamic
// cron discovery without executing any runtime code.
type recordingDynamicCronExecutor struct {
	discoverPluginIDs []string
}

// DiscoverCronContracts records the manifest passed to dynamic discovery.
func (e *recordingDynamicCronExecutor) DiscoverCronContracts(
	_ context.Context,
	manifest *catalog.Manifest,
) ([]*pluginbridge.CronContract, error) {
	if manifest != nil {
		e.discoverPluginIDs = append(e.discoverPluginIDs, manifest.ID)
	}
	return []*pluginbridge.CronContract{}, nil
}

// ExecuteDeclaredCronJob is unused by this regression test.
func (e *recordingDynamicCronExecutor) ExecuteDeclaredCronJob(
	_ context.Context,
	_ *catalog.Manifest,
	_ *pluginbridge.CronContract,
) error {
	return nil
}

// TestListManagedCronJobsSkipsDynamicDiscoveryForSourcePlugins verifies source
// manifests keep using callback-based cron registration and are not sent to the
// dynamic Wasm cron-discovery path.
func TestListManagedCronJobsSkipsDynamicDiscoveryForSourcePlugins(t *testing.T) {
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

	items, err := services.Integration.ListManagedCronJobs(context.Background())
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

// TestListManagedCronJobsSkipsPendingUpgradeDynamicPlugin verifies dynamic cron
// declarations are not discovered while the plugin waits for runtime upgrade.
func TestListManagedCronJobsSkipsPendingUpgradeDynamicPlugin(t *testing.T) {
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
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	manifest.HostServices = []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceCron,
		Methods: []string{
			pluginbridge.HostServiceMethodCronRegister,
		},
	}}
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
	newManifest.HostServices = []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceCron,
		Methods: []string{
			pluginbridge.HostServiceMethodCronRegister,
		},
	}}
	if _, err = services.Catalog.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic cron manifest sync to succeed, got error: %v", err)
	}

	items, err := services.Integration.ListManagedCronJobs(ctx)
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
