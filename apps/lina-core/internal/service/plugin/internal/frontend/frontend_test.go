// This file verifies that the bundled dynamic demo plugin artifact stays aligned with its review sources.

package frontend_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	pluginfrontend "lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

func TestPluginDemoRuntimePluginMatchesReviewSource(t *testing.T) {
	services := testutil.NewServices()

	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	manifest := &catalog.Manifest{}
	if err = services.Catalog.LoadManifestFromYAML(manifestPath, manifest); err != nil {
		t.Fatalf("failed to load dynamic plugin manifest: %v", err)
	}

	expectedFrontendAssets := mustCollectDynamicFrontendAssets(t, pluginDir)
	expectedInstallSQLAssets := mustCollectDynamicSQLAssets(t, pluginDir, false)
	expectedUninstallSQLAssets := mustCollectDynamicSQLAssets(t, pluginDir, true)

	stagedPluginDir := stageRuntimePluginForValidation(t, pluginDir)
	stagedManifestPath := filepath.Join(stagedPluginDir, "plugin.yaml")
	stagedManifest := &catalog.Manifest{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		Description:  manifest.Description,
		Author:       manifest.Author,
		Homepage:     manifest.Homepage,
		License:      manifest.License,
		ManifestPath: stagedManifestPath,
		RootDir:      stagedPluginDir,
	}
	if err = services.Catalog.ValidateManifest(stagedManifest, stagedManifestPath); err != nil {
		t.Fatalf("expected dynamic sample plugin manifest to be valid, got error: %v", err)
	}

	if stagedManifest.RuntimeArtifact == nil || stagedManifest.RuntimeArtifact.Manifest == nil {
		t.Fatalf("expected dynamic artifact metadata to be loaded")
	}
	if stagedManifest.RuntimeArtifact.Manifest.ID != "plugin-demo-dynamic" {
		t.Fatalf("expected dynamic plugin id plugin-demo-dynamic, got %s", stagedManifest.RuntimeArtifact.Manifest.ID)
	}
	if stagedManifest.RuntimeArtifact.Manifest.Name != manifest.Name {
		t.Fatalf("expected dynamic plugin name %s, got %s", manifest.Name, stagedManifest.RuntimeArtifact.Manifest.Name)
	}
	if stagedManifest.RuntimeArtifact.Manifest.Version != manifest.Version {
		t.Fatalf("expected dynamic plugin version %s, got %s", manifest.Version, stagedManifest.RuntimeArtifact.Manifest.Version)
	}
	if stagedManifest.RuntimeArtifact.RuntimeKind != pluginbridge.RuntimeKindWasm {
		t.Fatalf("expected runtime kind %s, got %s", pluginbridge.RuntimeKindWasm, stagedManifest.RuntimeArtifact.RuntimeKind)
	}
	if stagedManifest.RuntimeArtifact.ABIVersion != pluginbridge.SupportedABIVersion {
		t.Fatalf("expected runtime ABI %s, got %s", pluginbridge.SupportedABIVersion, stagedManifest.RuntimeArtifact.ABIVersion)
	}

	assertRuntimeFrontendAssetsMatch(t, expectedFrontendAssets, stagedManifest.RuntimeArtifact.FrontendAssets)
	assertRuntimeSQLAssetsMatch(t, expectedInstallSQLAssets, stagedManifest.RuntimeArtifact.InstallSQLAssets)
	assertRuntimeSQLAssetsMatch(t, expectedUninstallSQLAssets, stagedManifest.RuntimeArtifact.UninstallSQLAssets)

	bundle, err := services.Frontend.EnsureBundleReader(context.Background(), stagedManifest)
	if err != nil {
		t.Fatalf("expected dynamic frontend bundle to load, got error: %v", err)
	}
	if bundle == nil || !bundle.HasAsset("mount.js") {
		t.Fatalf("expected dynamic frontend bundle to expose mount.js")
	}
	if _, statErr := os.Stat(filepath.Join(stagedPluginDir, "runtime", "frontend-assets")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no extracted dynamic frontend-assets directory, got err=%v", statErr)
	}

	hostedBaseURL := services.Frontend.BuildRuntimeFrontendPublicBaseURL(stagedManifest.ID, stagedManifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey:    "plugin:plugin-demo-dynamic:main-entry",
			Name:       "动态插件示例",
			Path:       hostedBaseURL + "mount.js",
			Component:  pluginfrontend.DynamicPageComponentPath,
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
			IsFrame:    0,
		},
	}
	if err = services.Frontend.ValidateHostedMenuBindings(context.Background(), stagedManifest, menus); err != nil {
		t.Fatalf("expected dynamic sample menu contract to be valid, got error: %v", err)
	}
}

func stageRuntimePluginForValidation(t *testing.T, sourcePluginDir string) string {
	t.Helper()

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, sourcePluginDir)
	targetPluginDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(targetPluginDir, "runtime"), 0o755); err != nil {
		t.Fatalf("failed to create staged runtime dir: %v", err)
	}

	manifestContent, err := os.ReadFile(filepath.Join(sourcePluginDir, "plugin.yaml"))
	if err != nil {
		t.Fatalf("failed to read source plugin manifest: %v", err)
	}
	if err = os.WriteFile(filepath.Join(targetPluginDir, "plugin.yaml"), manifestContent, 0o644); err != nil {
		t.Fatalf("failed to write staged plugin manifest: %v", err)
	}

	artifactPath := filepath.Join(targetPluginDir, runtime.BuildArtifactRelativePath("plugin-demo-dynamic"))
	if err = os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("failed to create staged artifact dir: %v", err)
	}
	if err = os.WriteFile(artifactPath, buildOut.Content, 0o644); err != nil {
		t.Fatalf("failed to write staged dynamic artifact: %v", err)
	}

	return targetPluginDir
}

func mustCollectDynamicFrontendAssets(t *testing.T, pluginDir string) []*catalog.ArtifactFrontendAsset {
	t.Helper()

	frontendDir := filepath.Join(pluginDir, "frontend", "pages")
	files, err := os.ReadDir(frontendDir)
	if err != nil {
		t.Fatalf("failed to collect dynamic frontend assets: %v", err)
	}
	assets := make([]*catalog.ArtifactFrontendAsset, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(frontendDir, file.Name()))
		if err != nil {
			t.Fatalf("failed to read dynamic frontend asset: %v", err)
		}
		assets = append(assets, &catalog.ArtifactFrontendAsset{
			Path:    file.Name(),
			Content: content,
		})
	}
	return assets
}

func mustCollectDynamicSQLAssets(t *testing.T, pluginDir string, uninstall bool) []*catalog.ArtifactSQLAsset {
	t.Helper()

	searchDir := filepath.Join(pluginDir, "manifest", "sql")
	if uninstall {
		searchDir = filepath.Join(pluginDir, "manifest", "sql", "uninstall")
	}
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*catalog.ArtifactSQLAsset{}
		}
		t.Fatalf("failed to collect dynamic SQL assets: %v", err)
	}
	assets := make([]*catalog.ArtifactSQLAsset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(searchDir, entry.Name()))
		if err != nil {
			t.Fatalf("failed to read dynamic SQL asset: %v", err)
		}
		assets = append(assets, &catalog.ArtifactSQLAsset{
			Key:     entry.Name(),
			Content: strings.TrimSpace(string(content)),
		})
	}
	return assets
}

func assertRuntimeFrontendAssetsMatch(
	t *testing.T,
	expected []*catalog.ArtifactFrontendAsset,
	actual []*catalog.ArtifactFrontendAsset,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d frontend assets, got %d", len(expected), len(actual))
	}

	expectedByPath := make(map[string]*catalog.ArtifactFrontendAsset, len(expected))
	for _, asset := range expected {
		expectedByPath[asset.Path] = asset
	}

	for _, asset := range actual {
		expectedAsset, ok := expectedByPath[asset.Path]
		if !ok {
			t.Fatalf("unexpected frontend asset path: %s", asset.Path)
		}
		if string(asset.Content) != string(expectedAsset.Content) {
			t.Fatalf("unexpected content for frontend asset %s", asset.Path)
		}
	}
}

func assertRuntimeSQLAssetsMatch(
	t *testing.T,
	expected []*catalog.ArtifactSQLAsset,
	actual []*catalog.ArtifactSQLAsset,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d SQL assets, got %d", len(expected), len(actual))
	}

	for index := range expected {
		if actual[index].Key != expected[index].Key {
			t.Fatalf("expected SQL asset key %s, got %s", expected[index].Key, actual[index].Key)
		}
		if strings.TrimSpace(actual[index].Content) != strings.TrimSpace(expected[index].Content) {
			t.Fatalf("unexpected SQL content for asset %s", expected[index].Key)
		}
	}
}
