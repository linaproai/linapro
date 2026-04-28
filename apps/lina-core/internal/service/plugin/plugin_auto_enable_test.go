// This file verifies startup bootstrap behavior driven by plugin.autoEnable.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	runtimepkg "lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

// TestBootstrapAutoEnableInstallsAndEnablesSourcePlugin verifies startup
// bootstrap promotes a discovered source plugin to the enabled state.
func TestBootstrapAutoEnableInstallsAndEnablesSourcePlugin(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-source-auto-enable"
		version  = "v0.1.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\n"+
			"name: Source Auto Enable Plugin\n"+
			"version: "+version+"\n"+
			"type: source\n",
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
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		t.Fatalf("expected source plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != catalog.HostStateEnabled.String() {
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

// TestBootstrapAutoEnableReusesDynamicAuthorizationSnapshot verifies startup
// bootstrap can reinstall and enable a dynamic plugin after one confirmed host
// service authorization snapshot already exists for the target release.
func TestBootstrapAutoEnableReusesDynamicAuthorizationSnapshot(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = newTestService()
		pluginID = "plugin-dynamic-auto-enable-auth"
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
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: pluginbridge.HostServiceStorage,
					Methods: []string{pluginbridge.HostServiceMethodStorageGet},
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
				Service: pluginbridge.HostServiceStorage,
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

	if err := service.Install(ctx, pluginID, authorization); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled, authorization); err != nil {
		t.Fatalf("expected initial dynamic plugin enable to succeed, got error: %v", err)
	}
	if err := service.Uninstall(ctx, pluginID); err != nil {
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
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		t.Fatalf("expected dynamic plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != catalog.HostStateEnabled.String() {
		t.Fatalf("expected dynamic plugin current state enabled, got %s", registry.CurrentState)
	}

	release, err := bootstrapService.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		t.Fatalf("expected dynamic plugin release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatal("expected dynamic plugin release row after startup bootstrap")
	}
	snapshot, err := bootstrapService.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
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
		pluginID = "plugin-dynamic-auto-enable-auth-missing"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtimepkg.BuildArtifactFileName(pluginID))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Auto Enable Missing Authorization Plugin",
			Version: "v0.6.1",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceNetwork,
					Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
					Resources: []*pluginbridge.HostServiceResourceSpec{
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
		pluginID = "plugin-dynamic-auto-enable-cluster"
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
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		t.Fatalf("expected cluster bootstrap plugin to be installed and enabled, got %#v", registry)
	}
	if registry.CurrentState != catalog.HostStateEnabled.String() {
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
