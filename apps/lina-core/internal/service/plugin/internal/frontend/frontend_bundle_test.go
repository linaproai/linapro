// This file covers in-memory runtime frontend bundle loading behaviors.

package frontend_test

import (
	"context"
	"encoding/base64"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	pluginfrontend "lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/testutil"
)

// resetBundleCache clears the package-level runtime bundle cache through the
// public frontend service contract used by production invalidation paths.
func resetBundleCache(t *testing.T, service pluginfrontend.Service) {
	t.Helper()
	service.InvalidateAllBundles(context.Background(), "test_reset")
	t.Cleanup(func() {
		service.InvalidateAllBundles(context.Background(), "test_cleanup")
	})
}

// TestEnsureBundleReaderReadsEmbeddedAssetsWithoutExtraction verifies that
// runtime assets stay in memory and are not extracted to the plugin workspace.
func TestEnsureBundleReaderReadsEmbeddedAssetsWithoutExtraction(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	resetBundleCache(t, service)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dev-dynamic-bundle",
		"Runtime Bundle Plugin",
		"v0.4.0",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>bundle asset</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "frontend/pages/assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('bundle asset');")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-dynamic-bundle",
		Name:         "Runtime Bundle Plugin",
		Version:      "v0.4.0",
		Type:         pluginv1.PluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	bundle, err := service.EnsureBundleReader(context.Background(), manifest)
	if err != nil {
		t.Fatalf("expected dynamic frontend bundle to load, got error: %v", err)
	}

	indexContent, contentType, err := bundle.ReadAsset("frontend/pages/index.html")
	if err != nil {
		t.Fatalf("expected bundle root asset to resolve, got error: %v", err)
	}
	if expected := "<html><body>bundle asset</body></html>"; !strings.Contains(string(indexContent), expected) {
		t.Fatalf("expected bundle index content to contain %q, got %q", expected, string(indexContent))
	}
	if contentType != "text/html; charset=utf-8" {
		t.Fatalf("expected html content type, got %s", contentType)
	}

	assetDir := filepath.Join(pluginDir, "runtime", "frontend-assets")
	if _, statErr := os.Stat(assetDir); !os.IsNotExist(statErr) {
		t.Fatalf("expected no extracted frontend-assets directory, got err=%v", statErr)
	}
}
