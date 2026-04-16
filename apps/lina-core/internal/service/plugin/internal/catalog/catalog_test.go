// This file contains unit tests for manifest validation, convention-based
// resource discovery, and review-oriented plugin metadata helpers.

package catalog_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

func TestValidatePluginManifestAcceptsMinimalSourcePlugin(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-manifest-valid")

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:          "plugin-manifest-valid",
		Name:        "Manifest Validation Plugin",
		Version:     "0.1.0",
		Type:        catalog.TypeSource.String(),
		Description: "A valid source plugin manifest used by unit tests.",
		Author:      "test-suite",
		License:     "Apache-2.0",
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected manifest to be valid, got error: %v", err)
	}
}

func TestValidatePluginManifestRejectsMissingBackendEntryForSourcePlugin(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-missing-backend")
	if err := os.Remove(filepath.Join(pluginDir, "backend", "plugin.go")); err != nil {
		t.Fatalf("failed to remove backend entry: %v", err)
	}

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:      "plugin-missing-backend",
		Name:    "Missing Backend Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
	}

	err := svcs.Catalog.ValidateManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "backend/plugin.go") {
		t.Fatalf("expected missing backend entry error, got: %v", err)
	}
}

func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedWasmMetadata(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-valid",
		"Runtime Validation Plugin",
		"v0.2.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-valid.sql", Content: "SELECT 1;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-valid.sql", Content: "SELECT 2;"},
		},
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:          "plugin-dynamic-valid",
		Name:        "Runtime Validation Plugin",
		Version:     "v0.2.0",
		Type:        catalog.TypeDynamic.String(),
		Description: "A valid dynamic plugin manifest used by unit tests.",
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}
	if manifest.RuntimeArtifact == nil {
		t.Fatalf("expected dynamic artifact metadata to be loaded")
	}
	if manifest.RuntimeArtifact.RuntimeKind != pluginbridge.RuntimeKindWasm {
		t.Fatalf("expected runtime kind wasm, got %s", manifest.RuntimeArtifact.RuntimeKind)
	}
	if manifest.RuntimeArtifact.ABIVersion != pluginbridge.SupportedABIVersion {
		t.Fatalf("expected ABI version %s, got %s", pluginbridge.SupportedABIVersion, manifest.RuntimeArtifact.ABIVersion)
	}
}

func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedFrontendAssets(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-frontend",
		"Runtime Frontend Plugin",
		"v0.2.1",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend')")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:      "plugin-dynamic-frontend",
		Name:    "Runtime Frontend Plugin",
		Version: "v0.2.1",
		Type:    catalog.TypeDynamic.String(),
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected dynamic frontend manifest to be valid, got error: %v", err)
	}
	if manifest.RuntimeArtifact == nil {
		t.Fatalf("expected dynamic artifact metadata to be loaded")
	}
	if len(manifest.RuntimeArtifact.FrontendAssets) != 2 {
		t.Fatalf("expected 2 frontend assets, got %d", len(manifest.RuntimeArtifact.FrontendAssets))
	}
	if manifest.RuntimeArtifact.FrontendAssets[0].Path != "index.html" {
		t.Fatalf("expected normalized frontend asset path index.html, got %s", manifest.RuntimeArtifact.FrontendAssets[0].Path)
	}
}

func TestValidatePluginManifestRejectsMismatchedRuntimeWasmManifest(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-mismatch",
		"Runtime Mismatch Plugin",
		"v0.3.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	testutil.WriteRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, runtime.BuildArtifactRelativePath("plugin-dynamic-mismatch")),
		&catalog.ArtifactManifest{
			ID:      "plugin-dynamic-other",
			Name:    "Runtime Mismatch Plugin",
			Version: "v0.3.0",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:   pluginbridge.RuntimeKindWasm,
			ABIVersion:    pluginbridge.SupportedABIVersion,
			SQLAssetCount: 1,
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
		nil,
		nil,
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:      "plugin-dynamic-mismatch",
		Name:    "Runtime Mismatch Plugin",
		Version: "v0.3.0",
		Type:    catalog.TypeDynamic.String(),
	}

	err := svcs.Catalog.ValidateManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "嵌入清单 ID") {
		t.Fatalf("expected embedded manifest mismatch error, got: %v", err)
	}
}

func TestScanPluginManifestsRejectsDuplicatePluginIDs(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-duplicate-id")

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifestContent := strings.Join([]string{
		"id: plugin-demo-source",
		"name: Duplicate Plugin",
		"version: 0.1.0",
		"type: source",
		"description: Duplicate id test plugin",
		"author: test-suite",
		"license: Apache-2.0",
		"",
	}, "\n")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("failed to write duplicate manifest: %v", err)
	}

	_, err := svcs.Catalog.ScanManifests()
	if err == nil || !strings.Contains(err.Error(), "插件ID重复") {
		t.Fatalf("expected duplicate plugin id error, got: %v", err)
	}
}

func TestScanPluginManifestsRejectsDuplicateRuntimeArtifactPluginIDs(t *testing.T) {
	svcs := testutil.NewServices()

	testutil.CreateTestRuntimeStorageArtifactWithFilename(
		t,
		"plugin-dynamic-duplicate-a.wasm",
		"plugin-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)
	testutil.CreateTestRuntimeStorageArtifactWithFilename(
		t,
		"plugin-dynamic-duplicate-b.wasm",
		"plugin-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)

	_, err := svcs.Catalog.ScanManifests()
	if err == nil || !strings.Contains(err.Error(), "动态插件ID重复") {
		t.Fatalf("expected duplicate dynamic plugin id error, got: %v", err)
	}
}

func TestStoreUploadedRuntimePackageWritesCanonicalWasmIntoRuntimeStorage(t *testing.T) {
	svcs := testutil.NewServices()
	ctx := context.Background()

	pluginID := "plugin-dynamic-upload-storage"
	content := testutil.BuildTestRuntimeWasmContent(
		t,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Runtime Upload Storage Plugin",
			Version: "v0.5.0",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
	)

	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	storageArtifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName(pluginID))
	if err = os.Remove(storageArtifactPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove stale storage artifact %s: %v", storageArtifactPath, err)
	}
	t.Cleanup(func() {
		if cleanupErr := os.Remove(storageArtifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove storage artifact %s: %v", storageArtifactPath, cleanupErr)
		}
	})
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	out, err := svcs.Runtime.StoreUploadedPackage(ctx, "blob", content, true)
	if err != nil {
		t.Fatalf("expected runtime upload to succeed, got error: %v", err)
	}
	if out.Id != pluginID {
		t.Fatalf("expected uploaded plugin id %s, got %s", pluginID, out.Id)
	}
	if !gfile.Exists(storageArtifactPath) {
		t.Fatalf("expected dynamic artifact to be written into storage path: %s", storageArtifactPath)
	}
	if sourceManifestPath := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID, "plugin.yaml"); gfile.Exists(sourceManifestPath) {
		t.Fatalf("expected upload to stop creating source-tree plugin manifests, found: %s", sourceManifestPath)
	}
}

func TestDiscoverPluginSQLPathsUsesDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-sql-convention")

	installPaths := svcs.Catalog.DiscoverSQLPaths(pluginDir, false)
	if len(installPaths) != 1 || installPaths[0] != "manifest/sql/001-plugin-sql-convention.sql" {
		t.Fatalf("unexpected install sql paths: %#v", installPaths)
	}

	uninstallPaths := svcs.Catalog.DiscoverSQLPaths(pluginDir, true)
	if len(uninstallPaths) != 1 || uninstallPaths[0] != "manifest/sql/uninstall/001-plugin-sql-convention.sql" {
		t.Fatalf("unexpected uninstall sql paths: %#v", uninstallPaths)
	}
}

func TestDiscoverPluginVuePathsUseDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-vue-convention")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	testutil.WriteTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	pagePaths := svcs.Catalog.DiscoverPagePaths(pluginDir)
	if len(pagePaths) != 1 || pagePaths[0] != "frontend/pages/main-entry.vue" {
		t.Fatalf("unexpected page paths: %#v", pagePaths)
	}

	slotPaths := svcs.Catalog.DiscoverSlotPaths(pluginDir)
	if len(slotPaths) != 1 || slotPaths[0] != "frontend/slots/dashboard.workspace.after/workspace-card.vue" {
		t.Fatalf("unexpected slot paths: %#v", slotPaths)
	}
}

func TestBuildPluginManifestSnapshotIncludesDirectoryDiscoveredAssets(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-snapshot")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	testutil.WriteTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	snapshot, err := svcs.Catalog.BuildManifestSnapshot(&catalog.Manifest{
		ID:          "plugin-snapshot",
		Name:        "Snapshot Plugin",
		Version:     "0.1.0",
		Type:        catalog.TypeSource.String(),
		Description: "Snapshot test plugin",
		Menus: []*catalog.MenuSpec{
			{
				Key:  "plugin:plugin-snapshot:sidebar-entry",
				Name: "Snapshot Plugin",
				Type: "M",
			},
		},
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	})
	if err != nil {
		t.Fatalf("expected snapshot to build, got error: %v", err)
	}

	for _, expected := range []string{
		"frontendPageCount: 1",
		"frontendSlotCount: 1",
		"installSqlCount: 1",
		"menuCount: 1",
	} {
		if !strings.Contains(snapshot, expected) {
			t.Fatalf("expected snapshot to contain %s, got: %s", expected, snapshot)
		}
	}
}

func TestBuildPluginManifestSnapshotIncludesRuntimeArtifactMetadata(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-snapshot",
		"Runtime Snapshot Plugin",
		"v0.4.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-snapshot.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-snapshot",
		Name:         "Runtime Snapshot Plugin",
		Version:      "v0.4.0",
		Type:         catalog.TypeDynamic.String(),
		Description:  "Runtime snapshot test plugin",
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := svcs.Runtime.ValidateRuntimeArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	snapshot, err := svcs.Catalog.BuildManifestSnapshot(manifest)
	if err != nil {
		t.Fatalf("expected snapshot to build, got error: %v", err)
	}

	for _, expected := range []string{
		"runtimeKind: wasm",
		"runtimeAbiVersion: v1",
		"runtimeFrontendAssetCount: 2",
		"runtimeSqlAssetCount: 1",
	} {
		if !strings.Contains(snapshot, expected) {
			t.Fatalf("expected snapshot to contain %s, got: %s", expected, snapshot)
		}
	}
}

func TestBuildPluginResourceRefDescriptorsDoNotPersistConcreteFilePaths(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-resource-summary")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	testutil.WriteTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	descriptors := svcs.Integration.BuildResourceRefDescriptors(&catalog.Manifest{
		ID:      "plugin-resource-summary",
		Name:    "Resource Summary Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		Menus: []*catalog.MenuSpec{
			{
				Key:  "plugin:plugin-resource-summary:sidebar-entry",
				Name: "Resource Summary Plugin",
				Type: "M",
			},
		},
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	})
	if len(descriptors) == 0 {
		t.Fatalf("expected resource descriptors to be generated")
	}

	foundMenuDescriptor := false
	for _, descriptor := range descriptors {
		if descriptor == nil {
			continue
		}
		if descriptor.Kind == catalog.ResourceKindMenu {
			foundMenuDescriptor = true
		}
		if strings.Contains(descriptor.Key, "/") || strings.Contains(descriptor.OwnerKey, "/") {
			t.Fatalf("expected abstract resource identifiers without concrete file paths, got %#v", descriptor)
		}
		if strings.Contains(descriptor.Remark, ".vue") || strings.Contains(descriptor.Remark, ".sql") {
			t.Fatalf("expected remark to summarize resources without concrete file paths, got %#v", descriptor)
		}
	}
	if !foundMenuDescriptor {
		t.Fatalf("expected manifest menu descriptor to be generated")
	}
}

func TestBuildPluginResourceRefDescriptorsSummarizeRuntimeArtifact(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-resource-summary",
		"Runtime Resource Summary Plugin",
		"v0.5.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-resource-summary.sql", Content: "SELECT 1;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-resource-summary.sql", Content: "SELECT 2;"},
		},
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-resource-summary",
		Name:         "Runtime Resource Summary Plugin",
		Version:      "v0.5.0",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := svcs.Runtime.ValidateRuntimeArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	descriptors := svcs.Integration.BuildResourceRefDescriptors(manifest)
	foundRuntimeArtifact := false
	for _, descriptor := range descriptors {
		if descriptor == nil {
			continue
		}
		if descriptor.Kind == catalog.ResourceKindRuntimeWasm {
			foundRuntimeArtifact = true
			if !strings.Contains(descriptor.Remark, "ABI v1") {
				t.Fatalf("expected dynamic artifact remark to mention ABI version, got %#v", descriptor)
			}
		}
	}
	if !foundRuntimeArtifact {
		t.Fatalf("expected runtime wasm descriptor to be generated")
	}
}

func TestResolvePluginSQLAssetsPrefersEmbeddedRuntimeSQL(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-sql-assets",
		"Runtime SQL Assets Plugin",
		"v0.6.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-sql-assets.sql", Content: "SELECT 1;"},
			{Key: "002-plugin-dynamic-sql-assets.sql", Content: "SELECT 2;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-sql-assets.sql", Content: "SELECT 3;"},
		},
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-sql-assets",
		Name:         "Runtime SQL Assets Plugin",
		Version:      "v0.6.0",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := svcs.Runtime.ValidateRuntimeArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	installAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected install sql assets, got error: %v", err)
	}
	if len(installAssets) != 2 || installAssets[0].Key != "001-plugin-dynamic-sql-assets.sql" {
		t.Fatalf("unexpected install assets: %#v", installAssets)
	}

	uninstallAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionUninstall)
	if err != nil {
		t.Fatalf("expected uninstall sql assets, got error: %v", err)
	}
	if len(uninstallAssets) != 1 || uninstallAssets[0].Content != "SELECT 3;" {
		t.Fatalf("unexpected uninstall assets: %#v", uninstallAssets)
	}
}

func TestResolvePluginSQLAssetsFallsBackToDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-directory-sql-assets")

	manifest := &catalog.Manifest{
		ID:           "plugin-directory-sql-assets",
		Name:         "Directory SQL Assets Plugin",
		Version:      "0.1.0",
		Type:         catalog.TypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}

	installAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected directory install sql assets, got error: %v", err)
	}
	if len(installAssets) != 1 || installAssets[0].Key != "001-plugin-directory-sql-assets.sql" {
		t.Fatalf("unexpected directory install assets: %#v", installAssets)
	}
}

func TestScanEmbeddedSourcePluginManifestsUsesPluginEmbeddedFiles(t *testing.T) {
	svcs := testutil.NewServices()

	const pluginID = "plugin-embedded-manifest"
	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml":                                &fstest.MapFile{Data: []byte("id: plugin-embedded-manifest\nname: Embedded Manifest Plugin\nversion: 0.1.0\ntype: source\n")},
		"frontend/pages/main-entry.vue":              &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"frontend/slots/layout.header.after/tip.vue": &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"manifest/sql/001-plugin-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 1;\n"),
		},
		"manifest/sql/uninstall/001-plugin-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 2;\n"),
		},
	})
	pluginhost.RegisterSourcePlugin(sourcePlugin)

	manifests, err := svcs.Catalog.ScanEmbeddedSourceManifests()
	if err != nil {
		t.Fatalf("expected embedded source manifests to load, got error: %v", err)
	}

	var target *catalog.Manifest
	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			target = manifest
			break
		}
	}
	if target == nil {
		t.Fatalf("expected embedded source plugin %s to be discovered", pluginID)
	}
	if target.ManifestPath != "embedded/source-plugins/plugin-embedded-manifest/plugin.yaml" {
		t.Fatalf("unexpected embedded manifest path: %s", target.ManifestPath)
	}
	if len(svcs.Catalog.ListFrontendPagePaths(target)) != 1 {
		t.Fatalf("expected embedded frontend page paths to be discovered")
	}
	if len(svcs.Catalog.ListFrontendSlotPaths(target)) != 1 {
		t.Fatalf("expected embedded frontend slot paths to be discovered")
	}
}

func TestResolvePluginSQLAssetsUsesEmbeddedSourcePluginFiles(t *testing.T) {
	svcs := testutil.NewServices()

	manifest := &catalog.Manifest{
		ID:      "plugin-embedded-sql-assets",
		Name:    "Embedded SQL Assets Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		SourcePlugin: func() *pluginhost.SourcePlugin {
			sourcePlugin := pluginhost.NewSourcePlugin("plugin-embedded-sql-assets")
			sourcePlugin.UseEmbeddedFiles(fstest.MapFS{
				"plugin.yaml": &fstest.MapFile{Data: []byte("id: plugin-embedded-sql-assets\nname: Embedded SQL Assets Plugin\nversion: 0.1.0\ntype: source\n")},
				"manifest/sql/001-plugin-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 1;\n"),
				},
				"manifest/sql/uninstall/001-plugin-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 2;\n"),
				},
			})
			return sourcePlugin
		}(),
	}

	installAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected embedded install sql assets, got error: %v", err)
	}
	if len(installAssets) != 1 || installAssets[0].Content != "SELECT 1;" {
		t.Fatalf("unexpected embedded install assets: %#v", installAssets)
	}

	uninstallAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionUninstall)
	if err != nil {
		t.Fatalf("expected embedded uninstall sql assets, got error: %v", err)
	}
	if len(uninstallAssets) != 1 || uninstallAssets[0].Content != "SELECT 2;" {
		t.Fatalf("unexpected embedded uninstall assets: %#v", uninstallAssets)
	}
}

func TestDerivePluginLifecycleState(t *testing.T) {
	testCases := []struct {
		name       string
		pluginType string
		installed  int
		enabled    int
		expected   string
	}{
		{
			name:       "source enabled",
			pluginType: catalog.TypeSource.String(),
			installed:  catalog.InstalledYes,
			enabled:    catalog.StatusEnabled,
			expected:   catalog.LifecycleStateSourceEnabled.String(),
		},
		{
			name:       "source disabled",
			pluginType: catalog.TypeSource.String(),
			installed:  catalog.InstalledYes,
			enabled:    catalog.StatusDisabled,
			expected:   catalog.LifecycleStateSourceDisabled.String(),
		},
		{
			name:       "runtime uninstalled",
			pluginType: catalog.TypeDynamic.String(),
			installed:  catalog.InstalledNo,
			enabled:    catalog.StatusDisabled,
			expected:   catalog.LifecycleStateRuntimeUninstalled.String(),
		},
		{
			name:       "runtime installed disabled",
			pluginType: catalog.TypeDynamic.String(),
			installed:  catalog.InstalledYes,
			enabled:    catalog.StatusDisabled,
			expected:   catalog.LifecycleStateRuntimeInstalled.String(),
		},
		{
			name:       "runtime enabled",
			pluginType: catalog.TypeDynamic.String(),
			installed:  catalog.InstalledYes,
			enabled:    catalog.StatusEnabled,
			expected:   catalog.LifecycleStateRuntimeEnabled.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := catalog.DeriveLifecycleState(testCase.pluginType, testCase.installed, testCase.enabled)
			if actual != testCase.expected {
				t.Fatalf("expected lifecycle state %s, got %s", testCase.expected, actual)
			}
		})
	}
}

func TestDerivePluginNodeState(t *testing.T) {
	testCases := []struct {
		name      string
		installed int
		enabled   int
		expected  string
	}{
		{
			name:      "node uninstalled",
			installed: catalog.InstalledNo,
			enabled:   catalog.StatusDisabled,
			expected:  catalog.NodeStateUninstalled.String(),
		},
		{
			name:      "node installed",
			installed: catalog.InstalledYes,
			enabled:   catalog.StatusDisabled,
			expected:  catalog.NodeStateInstalled.String(),
		},
		{
			name:      "node enabled",
			installed: catalog.InstalledYes,
			enabled:   catalog.StatusEnabled,
			expected:  catalog.NodeStateEnabled.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := catalog.DeriveNodeState(testCase.installed, testCase.enabled)
			if actual != testCase.expected {
				t.Fatalf("expected node state %s, got %s", testCase.expected, actual)
			}
		})
	}
}
