// This file covers side-effect-free runtime upgrade previews defined in
// plugin_runtime_upgrade_preview.go.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// containsString reports whether values contains target.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan verifies that preview is
// read-only and exposes manifest snapshots, dependency checks, SQL summary,
// hostServices drift, and stable risk hints for a pending dynamic upgrade.
func TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-preview"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime upgrade preview artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Runtime Upgrade Preview Plugin",
			Version: oldVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{protocol.HostServiceMethodStorageGet},
					Paths:   []string{"reports/"},
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}

	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old release row")
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Runtime Upgrade Preview Plugin",
			Version: newVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{
						protocol.HostServiceMethodStorageGet,
						protocol.HostServiceMethodStoragePut,
					},
					Paths: []string{"reports/", "exports/"},
				},
			},
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-upgrade-preview.sql",
				Content: "CREATE TABLE IF NOT EXISTS plugin_dynamic_runtime_upgrade_preview(id INTEGER);",
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected target dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	preview, err := service.PreviewRuntimeUpgrade(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime upgrade preview to succeed, got error: %v", err)
	}
	if preview.PluginID != pluginID || preview.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected pending preview for %s, got %#v", pluginID, preview)
	}
	if preview.EffectiveVersion != oldVersion || preview.DiscoveredVersion != newVersion {
		t.Fatalf("expected versions %s/%s, got %#v", oldVersion, newVersion, preview)
	}
	if preview.FromManifest == nil || preview.FromManifest.Version != oldVersion {
		t.Fatalf("expected from manifest version %s, got %#v", oldVersion, preview.FromManifest)
	}
	if preview.ToManifest == nil || preview.ToManifest.Version != newVersion {
		t.Fatalf("expected to manifest version %s, got %#v", newVersion, preview.ToManifest)
	}
	if preview.SQLSummary.InstallSQLCount != 1 || preview.SQLSummary.RuntimeSQLAssetCount != 1 {
		t.Fatalf("expected target SQL summary to include one SQL asset, got %#v", preview.SQLSummary)
	}
	if !preview.HostServicesDiff.AuthorizationRequired || !preview.HostServicesDiff.AuthorizationChanged {
		t.Fatalf("expected host service authorization to be required and changed, got %#v", preview.HostServicesDiff)
	}
	if len(preview.HostServicesDiff.Changed) != 1 {
		t.Fatalf("expected one changed host service, got %#v", preview.HostServicesDiff)
	}
	change := preview.HostServicesDiff.Changed[0]
	if change.Service != protocol.HostServiceStorage {
		t.Fatalf("expected storage host service change, got %#v", change)
	}
	if len(change.FromPaths) != 1 || change.FromPaths[0] != "reports/" {
		t.Fatalf("expected from paths to contain reports/, got %#v", change.FromPaths)
	}
	if len(change.ToPaths) != 2 || change.ToPaths[0] != "exports/" || change.ToPaths[1] != "reports/" {
		t.Fatalf("expected target paths to contain exports/ and reports/, got %#v", change.ToPaths)
	}
	if preview.DependencyCheck == nil || preview.DependencyCheck.TargetID != pluginID {
		t.Fatalf("expected dependency check for target plugin, got %#v", preview.DependencyCheck)
	}
	if !containsString(preview.RiskHints, RuntimeUpgradeRiskHintUpgradeSQLRequiresReview) {
		t.Fatalf("expected SQL review risk hint, got %#v", preview.RiskHints)
	}
	if !containsString(preview.RiskHints, RuntimeUpgradeRiskHintHostServiceAuthorizationChanged) {
		t.Fatalf("expected host service authorization risk hint, got %#v", preview.RiskHints)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after preview to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != oldVersion || registry.ReleaseId != oldRelease.Id {
		t.Fatalf("expected preview not to switch effective release, got %#v", registry)
	}
}

// TestPreviewRuntimeUpgradeRejectsNormalPlugin verifies preview does not turn a
// non-pending plugin into an upgrade action.
func TestPreviewRuntimeUpgradeRejectsNormalPlugin(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-runtime-upgrade-preview-normal"
		version  = "v0.1.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Preview Normal Plugin",
		version,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}

	_, err = service.PreviewRuntimeUpgrade(ctx, pluginID)
	if !bizerr.Is(err, CodePluginRuntimeUpgradePreviewUnavailable) {
		t.Fatalf("expected preview unavailable bizerr, got %v", err)
	}
}
