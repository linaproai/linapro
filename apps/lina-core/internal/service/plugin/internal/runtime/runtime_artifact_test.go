// This file covers runtime artifact discovery, validation, and parsing behaviors.

package runtime_test

import (
	"encoding/base64"
	"encoding/json"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestScanPluginManifestsDiscoversRuntimePluginFromStorage verifies that
// dynamic plugins are discovered directly from the runtime storage directory.
func TestScanPluginManifestsDiscoversRuntimePluginFromStorage(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dev-dynamic-storage-scan"
	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Storage Scan Plugin",
		"v0.9.1",
		nil,
		nil,
	)

	manifests, err := services.Catalog.ScanManifests()
	if err != nil {
		t.Fatalf("expected scan to discover dynamic artifact from storage path, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest == nil || manifest.ID != pluginID {
			continue
		}
		if manifest.RuntimeArtifact == nil {
			t.Fatalf("expected dynamic artifact metadata to be loaded for %s", pluginID)
		}
		return
	}
	t.Fatalf("expected dynamic plugin %s to be discovered from storage path", pluginID)
}

// TestScanPluginManifestsDropsRuntimePluginAfterArtifactRemoval verifies that
// removing the staged artifact removes the plugin from fresh scans.
func TestScanPluginManifestsDropsRuntimePluginAfterArtifactRemoval(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dev-dynamic-missing-scan"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Scan Plugin",
		"v0.9.2",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated dynamic artifact: %v", err)
	}

	manifests, err := services.Catalog.ScanManifests()
	if err != nil {
		t.Fatalf("expected scan to succeed after dynamic artifact removal, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			t.Fatalf("expected removed dynamic plugin %s to disappear from scan results", pluginID)
		}
	}
}

// TestEnsureRuntimeArtifactAvailableRejectsMissingGeneratedWasm verifies that
// lifecycle preconditions fail with actionable guidance when the wasm artifact is missing.
func TestEnsureRuntimeArtifactAvailableRejectsMissingGeneratedWasm(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dev-dynamic-missing-install"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Install Plugin",
		"v0.9.3",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated dynamic artifact: %v", err)
	}

	manifest := &catalog.Manifest{
		ID:      pluginID,
		Name:    "Runtime Missing Install Plugin",
		Version: "v0.9.3",
		Type:    pluginv1.PluginTypeDynamic.String(),
		RootDir: filepath.Dir(artifactPath),
	}

	strictErr := services.Runtime.ValidateRuntimeArtifact(manifest, filepath.Dir(artifactPath))
	if strictErr == nil {
		t.Fatalf("expected strict runtime validation to report a missing artifact, got: %v", strictErr)
	}
	if expected := filepath.ToSlash(testutil.RuntimeArtifactRelativePath(pluginID)); !strings.Contains(strictErr.Error(), expected) {
		t.Fatalf("expected strict runtime validation to mention missing wasm path %q, got: %v", expected, strictErr)
	}

	err := services.Runtime.EnsureRuntimeArtifactAvailable(manifest, "install")
	if err == nil {
		t.Fatalf("expected lifecycle precondition to reject missing dynamic artifact")
	}
	if expected := "make wasm p=" + pluginID; !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle precondition error to mention %q, got: %v", expected, err)
	}
	if expected := filepath.ToSlash(testutil.RuntimeArtifactRelativePath(pluginID)); !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle precondition error to mention missing wasm path %q, got: %v", expected, err)
	}
}

// TestValidateRuntimeArtifactTreatsOmittedPublicAssetIndexAsDefault verifies
// dynamic plugin.yaml and embedded manifest comparison handles the default
// public_assets index consistently.
func TestValidateRuntimeArtifactTreatsOmittedPublicAssetIndexAsDefault(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-public-index-default",
		"Runtime Public Index Default Plugin",
		"v0.4.0",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-public-index-default"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-public-index-default",
			Name:    "Runtime Public Index Default Plugin",
			Version: "v0.4.0",
			Type:    pluginv1.PluginTypeDynamic.String(),
			PublicAssets: []*catalog.PublicAssetSpec{
				{Source: "frontend/pages", Mount: "/", Index: "index.html"},
			},
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:      "plugin-dev-dynamic-public-index-default",
		Name:    "Runtime Public Index Default Plugin",
		Version: "v0.4.0",
		Type:    pluginv1.PluginTypeDynamic.String(),
		RootDir: pluginDir,
		PublicAssets: []*catalog.PublicAssetSpec{
			{Source: "frontend/pages", Mount: "/"},
		},
	}

	if err := services.Runtime.ValidateRuntimeArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected omitted index to match explicit default index, got error: %v", err)
	}
}

// TestParseRuntimeArtifactLoadsRoutesAndBridgeSpec verifies that artifact
// parsing restores routes, bridge metadata, and host-service capabilities.
func TestParseRuntimeArtifactLoadsRoutesAndBridgeSpec(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-routes",
		"Runtime Route Plugin",
		"v0.3.0",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-routes"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-routes",
			Name:    "Runtime Route Plugin",
			Version: "v0.3.0",
			Type:    pluginv1.PluginTypeDynamic.String(),
			Dependencies: &plugintypes.DependencySpec{
				Plugins: []*plugintypes.PluginDependencySpec{
					{ID: "linapro-ai-core", Version: ">=0.1.0"},
				},
			},
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
			RouteCount:         1,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceRuntime,
					Methods: []string{
						protocol.HostServiceMethodRuntimeLogWrite,
						protocol.HostServiceMethodRuntimeStateGet,
					},
				},
				{
					Owner:   "linapro-ai-core",
					Service: "ai",
					Version: "v1",
					Methods: []string{
						"text.method_status.get",
					},
				},
			},
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/review-summary",
				Method:      "GET",
				Access:      protocol.AccessLogin,
				Permission:  "plugin-dev-dynamic-routes:review:view",
				RequestType: "ReviewSummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact load to succeed, got error: %v", err)
	}
	if len(manifest.Routes) != 1 || manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		t.Fatalf("expected runtime artifact to expose routes and executable bridge, got routes=%d bridge=%#v", len(manifest.Routes), manifest.BridgeSpec)
	}
	if _, ok := manifest.HostCapabilities[protocol.CapabilityRuntime]; !ok {
		t.Fatalf("expected runtime capability to be restored, got %#v", manifest.HostCapabilities)
	}
	if len(manifest.HostServices) != 2 {
		t.Fatalf("expected runtime host service snapshot to be restored, got %#v", manifest.HostServices)
	}
	var ownerAIService *protocol.HostServiceSpec
	for _, item := range manifest.HostServices {
		if item.Owner == "linapro-ai-core" && item.Service == "ai" {
			ownerAIService = item
			break
		}
	}
	if ownerAIService == nil || ownerAIService.Version != "v1" ||
		len(ownerAIService.Methods) != 1 ||
		ownerAIService.Methods[0] != "text.method_status.get" {
		t.Fatalf("expected owner-aware AI host service snapshot to be restored, got %#v", manifest.HostServices)
	}
}

// TestParseRuntimeArtifactRejectsOwnerHostServiceWithoutDependency verifies
// dynamic artifacts cannot bypass owner dependency checks through embedded
// host-service sections.
func TestParseRuntimeArtifactRejectsOwnerHostServiceWithoutDependency(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-owner-missing-dependency",
		"Runtime Owner Missing Dependency Plugin",
		"v0.3.1",
		nil,
		nil,
	)
	artifactPath := filepath.Join(
		pluginDir,
		testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-owner-missing-dependency"),
	)
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-owner-missing-dependency",
			Name:    "Runtime Owner Missing Dependency Plugin",
			Version: "v0.3.1",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Owner:   "linapro-ai-core",
					Service: "ai",
					Version: "v1",
					Methods: []string{
						"text.generate",
					},
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

	_, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err == nil || !strings.Contains(err.Error(), "dependencies.plugins entry") {
		t.Fatalf("expected owner host service dependency error from artifact parse, got %v", err)
	}
}

// TestParseRuntimeArtifactLoadsManifestResources verifies dynamic artifact
// manifest resources are decoded and exposed on the active artifact view.
func TestParseRuntimeArtifactLoadsManifestResources(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-manifest-resources",
		"Runtime Manifest Resource Plugin",
		"v0.3.9",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-manifest-resources"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-manifest-resources",
			Name:    "Runtime Manifest Resource Plugin",
			Version: "v0.3.9",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:           protocol.RuntimeKindWasm,
			ABIVersion:            protocol.SupportedABIVersion,
			ManifestResourceCount: 5,
			ManifestResources: []*catalog.ArtifactManifestResource{
				{
					Path:          "manifest/config/config.yaml",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("feature:\n  enabled: true\n")),
				},
				{
					Path:          "manifest/config/config.example.yaml",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("feature:\n  enabled: false\n")),
				},
				{
					Path:          "manifest/metadata.yaml",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("title: demo\n")),
				},
				{
					Path:          "manifest/sql/001-schema.sql",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("CREATE TABLE plugin_demo(id bigint);\n")),
				},
				{
					Path:          "manifest/i18n/zh-CN/plugin.json",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte(`{"plugin.demo":"demo"}`)),
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

	artifact, err := services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact with manifest resources to parse, got error: %v", err)
	}
	if artifact.ManifestResourceCount != 5 || len(artifact.ManifestResources) != 5 {
		t.Fatalf("expected five manifest resources, got count=%d resources=%#v", artifact.ManifestResourceCount, artifact.ManifestResources)
	}
	if string(artifact.ManifestResources[0].Content) != "feature:\n  enabled: true\n" {
		t.Fatalf("expected decoded config content, got %q", string(artifact.ManifestResources[0].Content))
	}
}

// TestParseRuntimeArtifactRejectsInvalidManifestResourcePath verifies artifact
// resources cannot smuggle absolute or traversal paths.
func TestParseRuntimeArtifactRejectsInvalidManifestResourcePath(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-manifest-invalid",
		"Runtime Manifest Invalid Plugin",
		"v0.3.10",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-manifest-invalid"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-manifest-invalid",
			Name:    "Runtime Manifest Invalid Plugin",
			Version: "v0.3.10",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:           protocol.RuntimeKindWasm,
			ABIVersion:            protocol.SupportedABIVersion,
			ManifestResourceCount: 1,
			ManifestResources: []*catalog.ArtifactManifestResource{
				{
					Path:          "../manifest/metadata.yaml",
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("SELECT 1;\n")),
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

	_, err := services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err == nil {
		t.Fatal("expected invalid manifest resource path to be rejected")
	}
	if !strings.Contains(err.Error(), "escapes manifest root") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

// TestParseRuntimeArtifactLoadsLifecycleContracts verifies dynamic Before*
// lifecycle declarations are restored from the dedicated artifact section.
func TestParseRuntimeArtifactLoadsLifecycleContracts(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-lifecycle-contracts",
		"Runtime Lifecycle Plugin",
		"v0.3.8",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-lifecycle-contracts"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-lifecycle-contracts",
			Name:    "Runtime Lifecycle Plugin",
			Version: "v0.3.8",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			LifecycleContracts: []*protocol.LifecycleContract{
				{
					Operation:    protocol.LifecycleOperationBeforeInstall,
					RequestType:  "DynamicBeforeInstallReq",
					InternalPath: "__lifecycle/before-install/",
					TimeoutMs:    1500,
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

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact load to succeed, got error: %v", err)
	}
	if len(manifest.LifecycleHandlers) != 1 {
		t.Fatalf("expected one lifecycle handler, got %#v", manifest.LifecycleHandlers)
	}
	handler := manifest.LifecycleHandlers[0]
	if handler.Operation != protocol.LifecycleOperationBeforeInstall ||
		handler.RequestType != "DynamicBeforeInstallReq" ||
		handler.InternalPath != "/__lifecycle/before-install" {
		t.Fatalf("unexpected lifecycle handler: %#v", handler)
	}
}

// TestParseRuntimeArtifactPreservesDependencies verifies dynamic artifacts
// expose embedded dependency declarations through manifest loading.
func TestParseRuntimeArtifactPreservesDependencies(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-dependencies",
		"Runtime Dependency Plugin",
		"v0.3.7",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-dependencies"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-dependencies",
			Name:    "Runtime Dependency Plugin",
			Version: "v0.3.7",
			Type:    pluginv1.PluginTypeDynamic.String(),
			Dependencies: &plugintypes.DependencySpec{
				Framework: &plugintypes.FrameworkDependencySpec{Version: ">=0.1.0 <1.0.0"},
				Plugins: []*plugintypes.PluginDependencySpec{
					{
						ID:      "linapro-tenant-core",
						Version: ">=0.1.0",
					},
				},
			},
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

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact load to succeed, got error: %v", err)
	}
	if manifest.Dependencies == nil || manifest.Dependencies.Framework == nil {
		t.Fatalf("expected dependencies to be restored, got %#v", manifest.Dependencies)
	}
	if manifest.Dependencies.Framework.Version != ">=0.1.0 <1.0.0" {
		t.Fatalf("unexpected framework dependency: %#v", manifest.Dependencies.Framework)
	}
	if len(manifest.Dependencies.Plugins) != 1 {
		t.Fatalf("expected one plugin dependency, got %#v", manifest.Dependencies.Plugins)
	}
	dependency := manifest.Dependencies.Plugins[0]
	if dependency.ID != "linapro-tenant-core" || dependency.Version != ">=0.1.0" {
		t.Fatalf("unexpected plugin dependency: %#v", dependency)
	}
}

// TestParseRuntimeArtifactRejectsUnsupportedDependencyPolicyFields verifies
// dynamic artifact manifests cannot carry plugin dependency policy fields.
func TestParseRuntimeArtifactRejectsUnsupportedDependencyPolicyFields(t *testing.T) {
	tests := []struct {
		name         string
		manifestBody map[string]any
		want         string
	}{
		{
			name: "plugin dependency required",
			manifestBody: map[string]any{
				"dependencies": map[string]any{
					"plugins": []map[string]any{
						{"id": "linapro-tenant-core", "required": true},
					},
				},
			},
			want: "dependencies.plugins[0].required",
		},
		{
			name: "plugin dependency install",
			manifestBody: map[string]any{
				"dependencies": map[string]any{
					"plugins": []map[string]any{
						{"id": "linapro-tenant-core", "install": "auto"},
					},
				},
			},
			want: "dependencies.plugins[0].install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := testutil.NewServices()
			manifestBody := map[string]any{
				"id":                  "plugin-dev-dynamic-policy",
				"name":                "Runtime Policy Plugin",
				"version":             "v0.3.8",
				"type":                pluginv1.PluginTypeDynamic.String(),
				"scopeNature":         pluginv1.ScopeNatureTenantAware.String(),
				"supportsMultiTenant": true,
				"defaultInstallMode":  pluginv1.InstallModeTenantScoped.String(),
			}
			for key, value := range tt.manifestBody {
				manifestBody[key] = value
			}

			content := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
			content = appendTestRuntimeCustomSection(t, content, protocol.WasmSectionManifest, manifestBody)
			content = appendTestRuntimeCustomSection(t, content, protocol.WasmSectionRuntime, map[string]any{
				"runtimeKind": protocol.RuntimeKindWasm,
				"abiVersion":  protocol.SupportedABIVersion,
			})

			_, err := services.Runtime.ParseRuntimeWasmArtifactContent("policy.wasm", content)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected unsupported field error containing %q, got %v", tt.want, err)
			}
		})
	}
}

// TestParseRuntimeArtifactAcceptsNestedRuntimeI18NAssets verifies runtime UI
// i18n assets can use the same nested JSON authoring format as source plugins.
func TestParseRuntimeArtifactAcceptsNestedRuntimeI18NAssets(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-runtime-i18n",
		"Runtime I18N Plugin",
		"v0.3.6",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-runtime-i18n"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-runtime-i18n",
			Name:    "Runtime I18N Plugin",
			Version: "v0.3.6",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:    protocol.RuntimeKindWasm,
			ABIVersion:     protocol.SupportedABIVersion,
			I18NAssetCount: 1,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	content = appendTestRuntimeCustomSection(
		t,
		content,
		protocol.WasmSectionI18NAssets,
		[]map[string]string{
			{
				"locale":  "zh-CN",
				"content": `{"plugin":{"plugin-dev-dynamic-runtime-i18n":{"name":"运行时国际化插件"}}}`,
			},
		},
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}

	artifact, err := services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact with nested runtime i18n assets to parse, got error: %v", err)
	}
	if artifact.I18NAssetCount != 1 {
		t.Fatalf("expected runtime i18n asset count 1, got %d", artifact.I18NAssetCount)
	}
}

// TestParseRuntimeArtifactValidatesAPIDocI18NAssets verifies that runtime
// artifact parsing validates the API-documentation i18n custom section before a
// dynamic plugin release is accepted.
func TestParseRuntimeArtifactValidatesAPIDocI18NAssets(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-apidoc-i18n",
		"Runtime APIDoc I18N Plugin",
		"v0.3.3",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-apidoc-i18n"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-apidoc-i18n",
			Name:    "Runtime APIDoc I18N Plugin",
			Version: "v0.3.3",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:          protocol.RuntimeKindWasm,
			ABIVersion:           protocol.SupportedABIVersion,
			APIDocI18NAssetCount: 1,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	content = appendTestRuntimeCustomSection(
		t,
		content,
		protocol.WasmSectionAPIDocI18NAssets,
		[]map[string]string{
			{
				"locale":  "zh-CN",
				"content": `{"plugins":{"plugin_dev_dynamic_apidoc_i18n":{"paths":{"get":{"summary":"运行时接口文档翻译"}}}}}`,
			},
		},
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}

	artifact, err := services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact with apidoc i18n assets to parse, got error: %v", err)
	}
	if artifact.APIDocI18NAssetCount != 1 {
		t.Fatalf("expected apidoc i18n asset count 1, got %d", artifact.APIDocI18NAssetCount)
	}
}

// TestParseRuntimeArtifactRejectsInvalidAPIDocI18NAssets verifies malformed
// apidoc i18n sections fail during artifact parsing rather than at document
// rendering time.
func TestParseRuntimeArtifactRejectsInvalidAPIDocI18NAssets(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-apidoc-i18n-invalid",
		"Runtime APIDoc I18N Invalid Plugin",
		"v0.3.4",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-apidoc-i18n-invalid"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-apidoc-i18n-invalid",
			Name:    "Runtime APIDoc I18N Invalid Plugin",
			Version: "v0.3.4",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:          protocol.RuntimeKindWasm,
			ABIVersion:           protocol.SupportedABIVersion,
			APIDocI18NAssetCount: 1,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	content = appendTestRuntimeCustomSection(
		t,
		content,
		protocol.WasmSectionAPIDocI18NAssets,
		[]map[string]string{
			{
				"locale":  "zh-CN",
				"content": `{`,
			},
		},
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}

	_, err = services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err == nil {
		t.Fatal("expected invalid apidoc i18n asset content to be rejected")
	}
	if !strings.Contains(err.Error(), protocol.WasmSectionAPIDocI18NAssets) || !strings.Contains(err.Error(), "zh-CN") {
		t.Fatalf("expected apidoc i18n error to mention section and locale, got %v", err)
	}
}

// TestParseRuntimeArtifactRejectsAPIDocI18NCountMismatch verifies artifact
// metadata cannot claim a different apidoc i18n asset count from the actual
// custom section payload.
func TestParseRuntimeArtifactRejectsAPIDocI18NCountMismatch(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dev-dynamic-apidoc-i18n-count",
		"Runtime APIDoc I18N Count Plugin",
		"v0.3.5",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, testutil.RuntimeArtifactRelativePath("plugin-dev-dynamic-apidoc-i18n-count"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dev-dynamic-apidoc-i18n-count",
			Name:    "Runtime APIDoc I18N Count Plugin",
			Version: "v0.3.5",
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:          protocol.RuntimeKindWasm,
			ABIVersion:           protocol.SupportedABIVersion,
			APIDocI18NAssetCount: 2,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	content = appendTestRuntimeCustomSection(
		t,
		content,
		protocol.WasmSectionAPIDocI18NAssets,
		[]map[string]string{
			{
				"locale":  "zh-CN",
				"content": `{"plugins.plugin_dev_dynamic_apidoc_i18n_count.paths.get.summary":"运行时接口文档翻译"}`,
			},
		},
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}

	_, err = services.Runtime.ParseRuntimeWasmArtifact(artifactPath)
	if err == nil {
		t.Fatal("expected apidoc i18n count mismatch to be rejected")
	}
	if !strings.Contains(err.Error(), "apidoc i18n") || !strings.Contains(err.Error(), "metadata=2 actual=1") {
		t.Fatalf("expected apidoc i18n count mismatch error, got %v", err)
	}
}

// appendTestRuntimeCustomSection appends one synthetic custom section to a wasm payload.
func appendTestRuntimeCustomSection(t *testing.T, content []byte, name string, payload any) []byte {
	t.Helper()

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("expected runtime custom payload marshal to succeed, got error: %v", err)
	}

	sectionPayload := append([]byte{}, encodeTestRuntimeULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, encodedPayload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodeTestRuntimeULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodeTestRuntimeULEB128 encodes one uint32 using the wasm unsigned LEB128 format.
func encodeTestRuntimeULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
