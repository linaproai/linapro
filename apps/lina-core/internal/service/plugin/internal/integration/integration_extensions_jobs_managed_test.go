// This file covers managed scheduled-job collection for source plugins and
// Jobs-domain dynamic plugin declarations.

package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestListExecutableJobsIncludesSourcePluginDefinitions verifies source
// plugins still contribute plugin-owned scheduled jobs through the host-managed
// Jobs projection path.
func TestListExecutableJobsIncludesSourcePluginDefinitions(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dev-source-jobs-managed"
	testutil.CreateTestPluginDir(t, pluginID)

	items, err := services.Integration.ListExecutableJobs(context.Background())
	if err != nil {
		t.Fatalf("expected managed job listing to succeed, got error: %v", err)
	}
	if !managedJobListContainsPlugin(items, pluginID) {
		t.Fatalf("expected managed job list to include source plugin %s", pluginID)
	}
}

// TestListJobDeclarationsIncludesDynamicJobsRegisterDefinitions verifies
// dynamic plugins can declare built-in jobs through the Jobs-domain discovery
// method and publish executable handlers through the same management projection.
func TestListJobDeclarationsIncludesDynamicJobsRegisterDefinitions(t *testing.T) {
	ctx := context.Background()
	const pluginID = "plugin-dev-dynamic-jobs-register"
	executor := &fakeDynamicJobExecutor{
		contracts: []*protocol.JobContract{{
			Name:           "heartbeat",
			DisplayName:    "Dynamic Plugin Heartbeat",
			Description:    "Verifies dynamic Jobs declaration execution.",
			Pattern:        "# */10 * * * *",
			Timezone:       protocol.DefaultJobContractTimezone,
			Scope:          protocol.JobScopeAllNode,
			Concurrency:    protocol.JobConcurrencySingleton,
			MaxConcurrency: 1,
			TimeoutSeconds: 30,
			RequestType:    "JobHeartbeatReq",
			InternalPath:   "/job-heartbeat",
		}},
	}
	services := testutil.NewServicesWithDynamicJobExecutor(executor)

	hostServices := []*protocol.HostServiceSpec{{
		Service: protocol.HostServiceJobs,
		Methods: []string{
			protocol.HostServiceMethodJobsRegister,
		},
	}}
	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithHostServices(
		t,
		pluginID,
		"Dynamic Jobs Register Plugin",
		"v0.1.0",
		hostServices,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic manifest to load, got error: %v", err)
	}
	manifest.ScopeNature = plugintypes.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = plugintypes.InstallModeGlobal.String()
	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Store.SetPluginStatus(ctx, pluginID, plugintypes.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enabled state to be set, got error: %v", err)
	}
	services.Integration.SetPluginEnabledState(pluginID, true)

	executableItems, err := services.Integration.ListExecutableJobsByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected executable managed job list to succeed, got error: %v", err)
	}
	if len(executableItems) != 1 {
		t.Fatalf("expected dynamic plugin to expose one executable scheduled job, got %#v", executableItems)
	}
	assertDynamicJobProjection(t, executableItems[0], pluginID)
	if executableItems[0].Handler == nil {
		t.Fatalf("expected dynamic job projection to publish an execution handler")
	}
	if err = executableItems[0].Handler(ctx); err != nil {
		t.Fatalf("expected dynamic job handler to execute through runtime executor, got error: %v", err)
	}
	if executor.executeCalls != 1 || executor.executedName != "heartbeat" {
		t.Fatalf("expected one dynamic job execution for heartbeat, got calls=%d name=%q", executor.executeCalls, executor.executedName)
	}

	declaredItems, err := services.Integration.ListJobDeclarationsByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected declaration job list to succeed, got error: %v", err)
	}
	if len(declaredItems) != 1 {
		t.Fatalf("expected dynamic plugin to expose one scheduled-job declaration, got %#v", declaredItems)
	}
	assertDynamicJobProjection(t, declaredItems[0], pluginID)

	installedItems, err := services.Integration.ListInstalledJobDeclarations(ctx)
	if err != nil {
		t.Fatalf("expected installed declaration job list to succeed, got error: %v", err)
	}
	installedItem, ok := managedJobListFindPlugin(installedItems, pluginID)
	if !ok {
		t.Fatalf("expected installed dynamic plugin %s to expose scheduled-job declarations, got %#v", pluginID, installedItems)
	}
	assertDynamicJobProjection(t, installedItem, pluginID)

	if executor.discoverCalls < 3 {
		t.Fatalf("expected discovery to run for executable, declared, and installed projections, got %d", executor.discoverCalls)
	}
}

// managedJobListContainsPlugin reports whether a managed job list includes
// at least one definition owned by pluginID.
func managedJobListContainsPlugin(items []integration.ManagedJob, pluginID string) bool {
	_, ok := managedJobListFindPlugin(items, pluginID)
	return ok
}

// managedJobListFindPlugin returns the first managed job owned by pluginID.
func managedJobListFindPlugin(items []integration.ManagedJob, pluginID string) (integration.ManagedJob, bool) {
	for _, item := range items {
		if item.PluginID == pluginID {
			return item, true
		}
	}
	return integration.ManagedJob{}, false
}

// assertDynamicJobProjection verifies the host projection created from one
// dynamic Jobs declaration.
func assertDynamicJobProjection(t *testing.T, item integration.ManagedJob, pluginID string) {
	t.Helper()

	if item.PluginID != pluginID {
		t.Fatalf("expected pluginID %q, got %q", pluginID, item.PluginID)
	}
	if item.Name != "heartbeat" {
		t.Fatalf("expected job name heartbeat, got %q", item.Name)
	}
	if item.DisplayName != "Dynamic Plugin Heartbeat" {
		t.Fatalf("expected display name to be preserved, got %q", item.DisplayName)
	}
	if item.Pattern != "# */10 * * * *" {
		t.Fatalf("expected pattern to be preserved, got %q", item.Pattern)
	}
	if item.Timezone != protocol.DefaultJobContractTimezone {
		t.Fatalf("expected timezone %q, got %q", protocol.DefaultJobContractTimezone, item.Timezone)
	}
	if item.Timeout != 30*time.Second {
		t.Fatalf("expected timeout 30s, got %s", item.Timeout)
	}
	if !strings.Contains(item.Description, "dynamic Jobs declaration") {
		t.Fatalf("expected description to be preserved, got %q", item.Description)
	}
}

// fakeDynamicJobExecutor supplies deterministic dynamic Jobs declarations for
// integration tests without executing a real Wasm artifact.
type fakeDynamicJobExecutor struct {
	contracts     []*protocol.JobContract
	discoverCalls int
	executeCalls  int
	executedName  string
}

// DiscoverJobContracts returns detached declared job contracts for one plugin.
func (f *fakeDynamicJobExecutor) DiscoverJobContracts(
	_ context.Context,
	manifest *catalog.Manifest,
) ([]*protocol.JobContract, error) {
	f.discoverCalls++
	out := make([]*protocol.JobContract, 0, len(f.contracts))
	for _, contract := range f.contracts {
		if contract == nil {
			continue
		}
		snapshot := *contract
		out = append(out, &snapshot)
	}
	return out, nil
}

// ExecuteDeclaredJob records one dynamic job execution request.
func (f *fakeDynamicJobExecutor) ExecuteDeclaredJob(
	_ context.Context,
	_ *catalog.Manifest,
	contract *protocol.JobContract,
) error {
	f.executeCalls++
	if contract != nil {
		f.executedName = contract.Name
	}
	return nil
}
