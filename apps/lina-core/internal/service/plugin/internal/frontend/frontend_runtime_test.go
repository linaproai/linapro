// This file covers declared public asset resolution for source and dynamic plugins.

package frontend_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	pluginfrontend "lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestResolveRuntimeFrontendAssetRequiresPublicAssetDeclaration verifies that
// dynamic artifact frontend assets are not publicly served without public_assets.
func TestResolveRuntimeFrontendAssetRequiresPublicAssetDeclaration(t *testing.T) {
	services := testutil.NewServices()
	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName("plugin-dev-public-asset-required"))
	t.Cleanup(func() {
		if err := os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to cleanup runtime artifact %s: %v", artifactPath, err)
		}
	})
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  "plugin-dev-public-asset-required",
			Name:                "Public Asset Required Plugin",
			Version:             "v0.1.0",
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &testutil.DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: 1,
		},
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html>declared only</html>")),
				ContentType:   "text/html; charset=utf-8",
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if _, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath); err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got %v", err)
	}

	_, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), "plugin-dev-public-asset-required", "v0.1.0", "index.html")
	if err == nil || !strings.Contains(err.Error(), "public assets") {
		t.Fatalf("expected undeclared public asset to be rejected, got %v", err)
	}
}

// TestResolveRuntimeFrontendAssetMapsDynamicPublicAssets verifies declared
// dynamic public_assets map browser paths to artifact frontend asset prefixes.
func TestResolveRuntimeFrontendAssetMapsDynamicPublicAssets(t *testing.T) {
	services := testutil.NewServices()
	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		"plugin-dev-dynamic-public-assets",
		"Dynamic Public Assets Plugin",
		"v0.1.0",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html>public dynamic</html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "frontend/pages/landing.htm",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html>dynamic landing</html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "frontend/pages/assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('public dynamic');")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	asset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), "plugin-dev-dynamic-public-assets", "v0.1.0", "assets/app.js")
	if err != nil {
		t.Fatalf("expected declared dynamic public asset to resolve, got %v", err)
	}
	if !strings.Contains(string(asset.Content), "public dynamic") {
		t.Fatalf("expected declared public asset content, got %s", string(asset.Content))
	}

	rootAsset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), "plugin-dev-dynamic-public-assets", "v0.1.0", "")
	if err != nil {
		t.Fatalf("expected default dynamic index asset to resolve, got %v", err)
	}
	if !strings.Contains(string(rootAsset.Content), "public dynamic") {
		t.Fatalf("expected default dynamic index content, got %s", string(rootAsset.Content))
	}

	if _, err = services.Frontend.ResolveRuntimeFrontendAsset(
		context.Background(),
		"plugin-dev-dynamic-public-assets",
		"v0.1.0",
		"../plugin.yaml",
	); err == nil {
		t.Fatal("expected traversal-style public asset request to be rejected")
	}
}

// TestResolveRuntimeFrontendAssetUsesDeclaredIndex verifies public_assets index
// selects the directory default file instead of hard-coding index.html.
func TestResolveRuntimeFrontendAssetUsesDeclaredIndex(t *testing.T) {
	services := testutil.NewServices()
	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	pluginID := "plugin-dev-dynamic-public-index"
	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if err := os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to cleanup runtime artifact %s: %v", artifactPath, err)
		}
	})
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                "Public Asset Index Plugin",
			Version:             "v0.1.0",
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &testutil.DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
			PublicAssets: []*catalog.PublicAssetSpec{
				{Source: "frontend/pages", Mount: "docs", Index: "index.htm"},
			},
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: 1,
		},
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.htm",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html>custom dynamic index</html>")),
				ContentType:   "text/html; charset=utf-8",
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if _, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath); err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got %v", err)
	}

	asset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), pluginID, "v0.1.0", "docs")
	if err != nil {
		t.Fatalf("expected declared dynamic index asset to resolve, got %v", err)
	}
	if !strings.Contains(string(asset.Content), "custom dynamic index") {
		t.Fatalf("expected declared dynamic index content, got %s", string(asset.Content))
	}
}

// TestResolveRuntimeFrontendAssetMapsSourcePublicAssets verifies source plugins
// can serve only files under their declared public_assets source directories.
func TestResolveRuntimeFrontendAssetMapsSourcePublicAssets(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-dev-source-public-assets")
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "frontend", "public", "logo.txt"), "source-public-logo")
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "frontend", "public", "empty.txt"), "")
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "frontend", "public", "index.htm"), "source-public-index")
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-source-public-assets\nname: Source Public Assets Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\npublic_assets:\n  - source: frontend/public\n    mount: /\n    index: index.htm\n",
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-source-public-assets",
		Name:         "Source Public Assets Plugin",
		Version:      "0.1.0",
		Type:         plugintypes.TypeSource.String(),
		RootDir:      pluginDir,
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		PublicAssets: []*catalog.PublicAssetSpec{{Source: "frontend/public", Mount: "/", Index: "index.htm"}},
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected source manifest to validate, got %v", err)
	}

	asset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "logo.txt")
	if err != nil {
		t.Fatalf("expected source public asset to resolve, got %v", err)
	}
	if string(asset.Content) != "source-public-logo" {
		t.Fatalf("expected source public asset content, got %s", string(asset.Content))
	}

	emptyAsset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "empty.txt")
	if err != nil {
		t.Fatalf("expected empty source public asset to resolve, got %v", err)
	}
	if len(emptyAsset.Content) != 0 {
		t.Fatalf("expected empty source public asset content, got %q", string(emptyAsset.Content))
	}

	indexAsset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "")
	if err != nil {
		t.Fatalf("expected source public asset index to resolve, got %v", err)
	}
	if string(indexAsset.Content) != "source-public-index" {
		t.Fatalf("expected source public asset index content, got %s", string(indexAsset.Content))
	}

	if _, err = services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "plugin.yaml"); err == nil {
		t.Fatal("expected undeclared source asset to be rejected")
	}
}

// TestResolveRuntimeFrontendAssetServesPluginOwnedNonFrontendDirectory verifies
// that public_assets is an explicit publication boundary rather than a
// frontend-directory whitelist.
func TestResolveRuntimeFrontendAssetServesPluginOwnedNonFrontendDirectory(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-dev-source-public-backend")
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "backend", "public.txt"), "published backend-owned file")
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-source-public-backend\nname: Source Public Backend Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\npublic_assets:\n  - source: backend\n    mount: backend\n",
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-source-public-backend",
		Name:         "Source Public Backend Plugin",
		Version:      "0.1.0",
		Type:         plugintypes.TypeSource.String(),
		RootDir:      pluginDir,
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		PublicAssets: []*catalog.PublicAssetSpec{{Source: "backend", Mount: "backend"}},
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected source manifest to validate plugin-owned public asset source, got %v", err)
	}

	asset, err := services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "backend/public.txt")
	if err != nil {
		t.Fatalf("expected declared plugin-owned public asset to resolve, got %v", err)
	}
	if string(asset.Content) != "published backend-owned file" {
		t.Fatalf("expected declared plugin-owned asset content, got %s", string(asset.Content))
	}
}

// TestResolveRuntimeFrontendAssetRejectsSymlinkEscape verifies source asset
// serving rejects symlink paths that would leave the plugin resource boundary.
func TestResolveRuntimeFrontendAssetRejectsSymlinkEscape(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-dev-source-public-symlink")
	outsideDir := t.TempDir()
	testutil.WriteTestFile(t, filepath.Join(outsideDir, "secret.txt"), "outside")
	linkPath := filepath.Join(pluginDir, "frontend", "linked-public")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Fatalf("failed to create source public asset symlink: %v", err)
	}
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-source-public-symlink\nname: Source Public Symlink Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\npublic_assets:\n  - source: frontend/linked-public\n    mount: /\n",
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-source-public-symlink",
		Name:         "Source Public Symlink Plugin",
		Version:      "0.1.0",
		Type:         plugintypes.TypeSource.String(),
		RootDir:      pluginDir,
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		PublicAssets: []*catalog.PublicAssetSpec{{Source: "frontend/linked-public", Mount: "/"}},
	}
	err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath)
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected symlinked public asset source validation to fail, got %v", err)
	}

	if _, err = services.Frontend.ResolveRuntimeFrontendAsset(context.Background(), manifest.ID, manifest.Version, "secret.txt"); err == nil {
		t.Fatal("expected symlinked public asset request to be rejected")
	}
}
