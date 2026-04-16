// This file covers in-memory runtime frontend bundle loading behaviors.

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
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestEnsureBundleReaderReadsEmbeddedAssetsWithoutExtraction(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-bundle",
		"Runtime Bundle Plugin",
		"v0.4.0",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>bundle asset</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('bundle asset');")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-bundle",
		Name:         "Runtime Bundle Plugin",
		Version:      "v0.4.0",
		Type:         catalog.TypeDynamic.String(),
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

	indexContent, contentType, err := bundle.ReadAsset("")
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
