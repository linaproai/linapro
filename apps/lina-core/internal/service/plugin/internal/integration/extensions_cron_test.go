// This file covers managed-cron collection behavior across source and dynamic
// plugin manifests.

package integration_test

import (
	"context"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
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

	if _, err = services.Integration.ListManagedCronJobs(context.Background()); err != nil {
		t.Fatalf("expected managed cron listing to succeed, got error: %v", err)
	}

	for _, pluginID := range executor.discoverPluginIDs {
		if _, exists := sourcePluginIDs[pluginID]; exists {
			t.Fatalf("expected source plugin %s to skip dynamic cron discovery", pluginID)
		}
	}
}
