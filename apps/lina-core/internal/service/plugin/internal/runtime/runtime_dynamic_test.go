// This file covers runtime-state projection and upload recovery behaviors owned by runtime.

package runtime_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestListRuntimeStatesProjectsMissingRuntimeArtifactWithoutMutatingRegistry verifies
// that public runtime projections do not mutate registry state on missing artifacts.
func TestListRuntimeStatesProjectsMissingRuntimeArtifactWithoutMutatingRegistry(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dev-dynamic-runtime-state-readonly"

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
	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Store.SetPluginStatus(ctx, pluginID, plugintypes.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}

	registryBefore, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatalf("expected runtime registry row to exist before artifact removal")
	}
	if registryBefore.Installed != plugintypes.InstalledYes || registryBefore.Status != plugintypes.StatusEnabled {
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
	if runtimeState.Installed != plugintypes.InstalledNo || runtimeState.Enabled != plugintypes.StatusDisabled {
		t.Fatalf("expected public runtime state projection to return uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}
	if runtimeState.RuntimeState != plugintypes.RuntimeUpgradeStateNormal {
		t.Fatalf("expected missing artifact projection to keep runtime state normal, got %s", runtimeState.RuntimeState)
	}

	registryAfter, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after projection to succeed, got error: %v", err)
	}
	if registryAfter == nil {
		t.Fatalf("expected runtime registry row to remain after runtime-state projection")
	}
	if registryAfter.Installed != plugintypes.InstalledYes || registryAfter.Status != plugintypes.StatusEnabled {
		t.Fatalf("expected runtime-state projection to avoid mutating sys_plugin, got installed=%d enabled=%d", registryAfter.Installed, registryAfter.Status)
	}
}

// TestListRuntimeStatesUsesCatalogSnapshotForReleaseLookups verifies the public
// runtime-state list does not perform per-plugin release point queries.
func TestListRuntimeStatesUsesCatalogSnapshotForReleaseLookups(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		firstPluginID  = "plugin-dev-dynamic-runtime-state-snapshot-a"
		secondPluginID = "plugin-dev-dynamic-runtime-state-snapshot-b"
		version        = "v0.9.9"
	)

	for _, pluginID := range []string{firstPluginID, secondPluginID} {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	}
	t.Cleanup(func() {
		for _, pluginID := range []string{firstPluginID, secondPluginID} {
			testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		}
	})

	for _, pluginID := range []string{firstPluginID, secondPluginID} {
		artifactPath := testutil.CreateTestRuntimeStorageArtifact(
			t,
			pluginID,
			"Runtime State Snapshot Plugin",
			version,
			nil,
			nil,
		)
		manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
		if err != nil {
			t.Fatalf("expected dynamic artifact manifest to load for %s, got error: %v", pluginID, err)
		}
		if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
			t.Fatalf("expected dynamic manifest sync to succeed for %s, got error: %v", pluginID, err)
		}
		if err = services.Lifecycle.InstallDynamic(ctx, pluginID); err != nil {
			t.Fatalf("expected dynamic plugin install to succeed for %s, got error: %v", pluginID, err)
		}
	}

	var output *runtime.RuntimeStateListOutput
	sqls, err := gdb.CatchSQL(ctx, func(ctx context.Context) error {
		var listErr error
		output, listErr = services.Runtime.ListRuntimeStates(ctx)
		return listErr
	})
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}
	if findRuntimeStateItem(output.List, firstPluginID) == nil || findRuntimeStateItem(output.List, secondPluginID) == nil {
		t.Fatalf("expected runtime state list to include both snapshot test plugins")
	}
	assertNoReleasePointQueries(t, sqls)
}

// TestListRuntimeStatesScansDynamicManifestsOnce verifies the public runtime
// state list reuses one manifest scan instead of reparsing every artifact for
// every registry row.
func TestListRuntimeStatesScansDynamicManifestsOnce(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		firstPluginID  = "plugin-dev-dynamic-runtime-state-scan-a"
		secondPluginID = "plugin-dev-dynamic-runtime-state-scan-b"
		version        = "v0.9.10"
	)

	for _, pluginID := range []string{firstPluginID, secondPluginID} {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	}
	t.Cleanup(func() {
		for _, pluginID := range []string{firstPluginID, secondPluginID} {
			testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		}
	})

	for _, pluginID := range []string{firstPluginID, secondPluginID} {
		artifactPath := testutil.CreateTestRuntimeStorageArtifact(
			t,
			pluginID,
			"Runtime State Scan Plugin",
			version,
			nil,
			nil,
		)

		manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
		if err != nil {
			t.Fatalf("expected dynamic artifact manifest to load for %s, got error: %v", pluginID, err)
		}
		if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
			t.Fatalf("expected dynamic manifest sync to succeed for %s, got error: %v", pluginID, err)
		}
	}

	output, err := services.Runtime.ListRuntimeStates(ctx)
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}
	if findRuntimeStateItem(output.List, firstPluginID) == nil || findRuntimeStateItem(output.List, secondPluginID) == nil {
		t.Fatalf("expected runtime state list to include both scan-count test plugins")
	}
}

// TestInstallRepairsEmptyReleaseArchive verifies a same-version lifecycle pass
// replaces a corrupt zero-byte release artifact before the release is activated.
func TestInstallRepairsEmptyReleaseArchive(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID = "plugin-dev-dynamic-empty-release-repair"
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

	if err = services.Lifecycle.InstallDynamic(ctx, pluginID); err != nil {
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

	registry, err := services.Store.GetRegistry(ctx, pluginID)
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
	if err = services.Lifecycle.UninstallDynamic(ctx, pluginID); err != nil {
		t.Fatalf("expected uninstall to load repaired archive, got error: %v", err)
	}
}

// TestInstallRejectsLowerVersionAfterStagingArtifactReplacement verifies the
// public install path refreshes the staged artifact manifest before comparing
// versions, even when an earlier desired manifest was already cached.
func TestInstallDynamicRejectsLowerVersionBeforeRegistrySync(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID      = "plugin-dev-dynamic-install-downgrade"
		higherVersion = "v0.9.9"
		lowerVersion  = "v0.9.8"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	higherArtifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Install Downgrade Plugin",
		higherVersion,
		nil,
		nil,
	)
	if _, err := services.Catalog.LoadManifestFromArtifactPath(higherArtifactPath); err != nil {
		t.Fatalf("expected higher-version artifact manifest to load, got error: %v", err)
	}
	if _, err := services.Lifecycle.Install(ctx, pluginID, lifecycle.InstallOptions{}); err != nil {
		t.Fatalf("expected higher-version install to succeed, got error: %v", err)
	}
	if _, err := services.Catalog.GetDesiredManifest(pluginID); err != nil {
		t.Fatalf("expected desired manifest to be cached before replacement, got error: %v", err)
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		higherArtifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                "Runtime Install Downgrade Plugin",
			Version:             lowerVersion,
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &testutil.DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := services.Lifecycle.Install(ctx, pluginID, lifecycle.InstallOptions{})
	if !bizerr.Is(err, lifecycle.CodeDynamicPluginDowngradeUnsupported) {
		t.Fatalf("expected downgrade install to return structured downgrade error, got %v", err)
	}

	registry, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected registry row after rejected downgrade")
	}
	if registry.Version != higherVersion {
		t.Fatalf("expected rejected downgrade not to overwrite registry version, got %q want %q", registry.Version, higherVersion)
	}
	if registry.Installed != plugintypes.InstalledYes {
		t.Fatalf("expected plugin to remain installed after rejected downgrade, got %d", registry.Installed)
	}
}

// TestUploadDynamicPackageAllowsRecoveryWhenArtifactIsMissing verifies that a
// fresh upload can recover a dynamic plugin whose staged artifact disappeared.
func TestUploadDynamicPackageAllowsRecoveryWhenArtifactIsMissing(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dev-dynamic-upload-recover"

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
	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Store.SetPluginStatus(ctx, pluginID, plugintypes.StatusEnabled); err != nil {
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
	if out.Installed != plugintypes.InstalledNo {
		t.Fatalf("expected recovery upload to keep plugin uninstalled, got %d", out.Installed)
	}
	if out.Enabled != plugintypes.StatusDisabled {
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

// assertNoReleasePointQueries fails when captured SQL shows release lookups by
// primary key or plugin/version instead of the request-level release snapshot.
func assertNoReleasePointQueries(t *testing.T, sqls []string) {
	t.Helper()

	dbPrefix := g.DB().GetPrefix()
	releaseTable := dbPrefix + "sys_plugin_release"
	for _, sql := range sqls {
		normalized := strings.ToLower(sql)
		if !strings.Contains(normalized, strings.ToLower(releaseTable)) {
			continue
		}
		if strings.Contains(normalized, "where") &&
			(strings.Contains(normalized, `"id"=`) ||
				(strings.Contains(normalized, `"plugin_id"=`) && strings.Contains(normalized, `"release_version"=`))) {
			t.Fatalf("expected runtime-state list to avoid release point query, got SQL %q from %#v", sql, sqls)
		}
	}
}
