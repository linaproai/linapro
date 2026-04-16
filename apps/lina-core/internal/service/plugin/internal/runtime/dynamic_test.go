// This file covers runtime-state projection and upload recovery behaviors owned by runtime.

package runtime_test

import (
	"context"
	"os"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
)

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

func findRuntimeStateItem(items []*runtime.PluginDynamicStateItem, pluginID string) *runtime.PluginDynamicStateItem {
	for _, item := range items {
		if item != nil && item.Id == pluginID {
			return item
		}
	}
	return nil
}
