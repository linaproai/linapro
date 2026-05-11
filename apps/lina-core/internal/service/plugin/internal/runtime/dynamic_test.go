// This file covers runtime-state projection and upload recovery behaviors owned by runtime.

package runtime_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
)

// TestListRuntimeStatesProjectsMissingRuntimeArtifactWithoutMutatingRegistry verifies
// that public runtime projections do not mutate registry state on missing artifacts.
func TestListRuntimeStatesProjectsMissingRuntimeArtifactWithoutMutatingRegistry(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dynamic-runtime-state-readonly"

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime State Readonly Plugin",
		"v0.9.7",
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}

	registryBefore, err := services.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatalf("expected runtime registry row to exist before artifact removal")
	}
	if registryBefore.Installed != catalog.InstalledYes || registryBefore.Status != catalog.StatusEnabled {
		t.Fatalf("expected runtime registry row to remain installed+enabled before projection, got installed=%d enabled=%d", registryBefore.Installed, registryBefore.Status)
	}

	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	runtimeStates, err := services.Runtime.ListRuntimeStates(ctx)
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}

	runtimeState := findRuntimeStateItem(runtimeStates.List, pluginID)
	if runtimeState == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in public runtime states")
	}
	if runtimeState.Installed != catalog.InstalledNo || runtimeState.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected public runtime state projection to return uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}

	registryAfter, err := services.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after projection to succeed, got error: %v", err)
	}
	if registryAfter == nil {
		t.Fatalf("expected runtime registry row to remain after runtime-state projection")
	}
	if registryAfter.Installed != catalog.InstalledYes || registryAfter.Status != catalog.StatusEnabled {
		t.Fatalf("expected runtime-state projection to avoid mutating sys_plugin, got installed=%d enabled=%d", registryAfter.Installed, registryAfter.Status)
	}
}

// TestInstallRepairsEmptyReleaseArchive verifies a same-version lifecycle pass
// replaces a corrupt zero-byte release artifact before the release is activated.
func TestInstallRepairsEmptyReleaseArchive(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID = "plugin-dynamic-empty-release-repair"
		version  = "v0.9.8"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Empty Release Repair Plugin",
		version,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	releaseRoot := filepath.Join(testutil.TestDynamicStorageDir(), "releases", pluginID)
	if err := os.RemoveAll(releaseRoot); err != nil {
		t.Fatalf("failed to clear release archive root: %v", err)
	}
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if err := os.RemoveAll(releaseRoot); err != nil {
			t.Fatalf("failed to cleanup release archive root: %v", err)
		}
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if manifest.RuntimeArtifact == nil {
		t.Fatal("expected runtime artifact metadata")
	}

	archivePath := filepath.Join(
		releaseRoot,
		version,
		manifest.RuntimeArtifact.Checksum,
		runtime.BuildArtifactFileName(pluginID),
	)
	if err = os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		t.Fatalf("failed to create empty release archive directory: %v", err)
	}
	if err = os.WriteFile(archivePath, nil, 0o644); err != nil {
		t.Fatalf("failed to seed empty release archive: %v", err)
	}

	if err = services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected install to repair empty release archive, got error: %v", err)
	}

	sourceContent, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read source artifact: %v", err)
	}
	archiveContent, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("failed to read repaired release archive: %v", err)
	}
	if len(archiveContent) == 0 {
		t.Fatal("expected repaired release archive to be non-empty")
	}
	if !bytes.Equal(archiveContent, sourceContent) {
		t.Fatal("expected repaired release archive to match source artifact")
	}

	registry, err := services.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected runtime registry row after install")
	}
	activeManifest, err := services.Runtime.LoadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		t.Fatalf("expected active manifest to load from repaired archive, got error: %v", err)
	}
	if activeManifest == nil || activeManifest.ID != pluginID {
		t.Fatalf("expected repaired active manifest for %s, got %#v", pluginID, activeManifest)
	}
	if err = services.Lifecycle.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected uninstall to load repaired archive, got error: %v", err)
	}
}

// TestUploadDynamicPackageAllowsRecoveryWhenArtifactIsMissing verifies that a
// fresh upload can recover a dynamic plugin whose staged artifact disappeared.
func TestUploadDynamicPackageAllowsRecoveryWhenArtifactIsMissing(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dynamic-upload-recover"

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Upload Recover Plugin",
		"v0.9.6",
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read dynamic artifact content: %v", err)
	}

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := services.Runtime.StoreUploadedPackage(
		ctx,
		runtime.BuildArtifactFileName(pluginID),
		content,
		false,
	)
	if err != nil {
		t.Fatalf("expected runtime upload recovery to succeed, got error: %v", err)
	}
	if out.Installed != catalog.InstalledNo {
		t.Fatalf("expected recovery upload to keep plugin uninstalled, got %d", out.Installed)
	}
	if out.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected recovery upload to keep plugin disabled, got %d", out.Enabled)
	}

	exists, _, err := services.Runtime.HasArtifactStorageFile(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected uploaded dynamic artifact lookup to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatalf("expected recovery upload to restore dynamic artifact into storage")
	}
}

// findRuntimeStateItem returns the matching runtime-state projection for one plugin ID.
func findRuntimeStateItem(items []*runtime.PluginDynamicStateItem, pluginID string) *runtime.PluginDynamicStateItem {
	for _, item := range items {
		if item != nil && item.Id == pluginID {
			return item
		}
	}
	return nil
}
