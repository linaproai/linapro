// This file covers runtime hosted-menu validation against embedded frontend assets.

package frontend_test

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	pluginfrontend "lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestValidateHostedMenuBindingsAcceptsHostedRuntimeModes(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-bindings",
		"Runtime Binding Plugin",
		"v0.3.0",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>hosted entry</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "mount.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("export function mount() {}")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-bindings",
		Name:         "Runtime Binding Plugin",
		Version:      "v0.3.0",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	hostedBaseURL := service.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey: "plugin:plugin-dynamic-bindings:iframe-entry",
			Name:    "Hosted iframe entry",
			Path:    hostedBaseURL + "index.html",
			IsFrame: 0,
		},
		{
			MenuKey: "plugin:plugin-dynamic-bindings:new-window-entry",
			Name:    "Hosted new window entry",
			Path:    hostedBaseURL + "index.html",
			IsFrame: 1,
		},
		{
			MenuKey:    "plugin:plugin-dynamic-bindings:embedded-entry",
			Name:       "Hosted embedded entry",
			Path:       hostedBaseURL + "mount.js",
			Component:  pluginfrontend.DynamicPageComponentPath,
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
			IsFrame:    0,
		},
	}

	if err := service.ValidateHostedMenuBindings(context.Background(), manifest, menus); err != nil {
		t.Fatalf("expected runtime hosted menu bindings to be valid, got error: %v", err)
	}
}

func TestValidateHostedMenuBindingsRejectsBrokenEmbeddedMountContract(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	pluginfrontend.ResetBundleCache()
	t.Cleanup(pluginfrontend.ResetBundleCache)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-broken-bindings",
		"Broken Runtime Binding Plugin",
		"v0.3.1",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>hosted entry</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dynamic-broken-bindings",
		Name:         "Broken Runtime Binding Plugin",
		Version:      "v0.3.1",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	hostedBaseURL := service.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey:    "plugin:plugin-dynamic-broken-bindings:embedded-entry",
			Name:       "Broken embedded entry",
			Path:       hostedBaseURL + "index.html",
			Component:  pluginfrontend.DynamicPageComponentPath,
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
			IsFrame:    0,
		},
	}

	err := service.ValidateHostedMenuBindings(context.Background(), manifest, menus)
	if err == nil {
		t.Fatalf("expected broken embedded mount contract to be rejected")
	}
	if expected := ".js 或 .mjs"; !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to mention %q, got: %v", expected, err)
	}
}
