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

	"github.com/gogf/gf/v2/os/gfile"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	menusvc "lina-core/internal/service/menu"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// TestValidatePluginManifestAcceptsMinimalSourcePlugin verifies that the
// minimal required source-plugin structure passes validation.
func TestValidatePluginManifestAcceptsMinimalSourcePlugin(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-manifest-valid")

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:          "acme-demo-manifest-valid",
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

// TestLoadManifestFromYAMLReadsI18NPolicy verifies plugin.yaml i18n policy is
// part of the framework manifest contract rather than test-only parsing.
func TestLoadManifestFromYAMLReadsI18NPolicy(t *testing.T) {
	cases := []struct {
		name        string
		content     string
		wantPolicy  bool
		wantEnabled bool
		wantDefault string
		wantLocales []string
	}{
		{
			name:        "missing policy opts out",
			content:     "id: demo\nname: Demo\n",
			wantEnabled: false,
		},
		{
			name:        "enabled true opts in",
			content:     "id: demo\nname: Demo\ni18n:\n  enabled: true\n  default: zh-CN\n  locales:\n    - locale: zh-CN\n      nativeName: 简体中文\n    - locale: en-US\n      nativeName: English\n",
			wantPolicy:  true,
			wantEnabled: true,
			wantDefault: "zh-CN",
			wantLocales: []string{"zh-CN", "en-US"},
		},
		{
			name:        "enabled false remains opted out",
			content:     "id: demo\nname: Demo\ni18n:\n  enabled: false\n  default: zh-CN\n",
			wantPolicy:  true,
			wantEnabled: false,
			wantDefault: "zh-CN",
		},
		{
			name:        "missing enabled remains opted out",
			content:     "id: demo\nname: Demo\ni18n:\n  default: zh-CN\n",
			wantPolicy:  true,
			wantEnabled: false,
			wantDefault: "zh-CN",
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			manifestPath := filepath.Join(t.TempDir(), "plugin.yaml")
			if err := os.WriteFile(manifestPath, []byte(item.content), 0o644); err != nil {
				t.Fatalf("write manifest fixture failed: %v", err)
			}

			manifest := &catalog.Manifest{}
			if err := testutil.NewServices().Catalog.LoadManifestFromYAML(manifestPath, manifest); err != nil {
				t.Fatalf("expected manifest to load, got error: %v", err)
			}
			if !item.wantPolicy {
				if manifest.I18N != nil {
					t.Fatalf("expected missing i18n policy to stay nil, got %#v", manifest.I18N)
				}
			} else if manifest.I18N == nil {
				t.Fatal("expected i18n policy to be parsed")
			}
			if got := manifest.I18NEnabled(); got != item.wantEnabled {
				t.Fatalf("unexpected i18n enabled value: got %v want %v", got, item.wantEnabled)
			}
			if item.wantDefault != "" && manifest.I18N.Default != item.wantDefault {
				t.Fatalf("unexpected i18n.default value: got %q want %q", manifest.I18N.Default, item.wantDefault)
			}
			if len(item.wantLocales) > 0 {
				if len(manifest.I18N.Locales) != len(item.wantLocales) {
					t.Fatalf("unexpected i18n.locales length: got %d want %d", len(manifest.I18N.Locales), len(item.wantLocales))
				}
				for index, wantLocale := range item.wantLocales {
					if manifest.I18N.Locales[index].Locale != wantLocale {
						t.Fatalf("unexpected i18n.locales[%d].locale: got %q want %q", index, manifest.I18N.Locales[index].Locale, wantLocale)
					}
				}
			}
		})
	}
}

// TestParsePluginIDReturnsSuggestedIdentityParts verifies plugin IDs can expose
// suggested author, domain, and capability parts without making that structure
// a runtime acceptance requirement.
func TestParsePluginIDReturnsSuggestedIdentityParts(t *testing.T) {
	parts, err := catalog.ParsePluginID("linapro-content-notice")
	if err != nil {
		t.Fatalf("expected structured plugin ID to parse, got %v", err)
	}
	if parts.Author != "linapro" || parts.Domain != "content" || parts.Capability != "notice" {
		t.Fatalf("unexpected plugin ID parts: %#v", parts)
	}

	parts, err = catalog.ParsePluginID("linapro-ops-demo-guard")
	if err != nil {
		t.Fatalf("expected multi-word capability to parse, got %v", err)
	}
	if parts.Author != "linapro" || parts.Domain != "ops" || parts.Capability != "demo-guard" {
		t.Fatalf("unexpected multi-word capability parts: %#v", parts)
	}

	parts, err = catalog.ParsePluginID("demo-control")
	if err != nil {
		t.Fatalf("expected non-three-segment plugin ID to parse, got %v", err)
	}
	if parts.Author != "demo" || parts.Domain != "control" || parts.Capability != "" {
		t.Fatalf("unexpected non-three-segment parts: %#v", parts)
	}
}

// TestValidatePluginIDEnforcesOnlyRuntimeSafetyBoundary verifies runtime
// validation rejects unsafe identifiers without hard-coding plugin taxonomy.
func TestValidatePluginIDEnforcesOnlyRuntimeSafetyBoundary(t *testing.T) {
	for _, pluginID := range []string{
		"demo-control",
		"acme-random-report",
		"acme-org-core",
		"plugin-demo-source",
	} {
		if err := catalog.ValidatePluginID(pluginID); err != nil {
			t.Fatalf("expected runtime-safe plugin ID %s to validate, got %v", pluginID, err)
		}
	}

	tests := []struct {
		name     string
		pluginID string
		want     string
	}{
		{name: "uppercase", pluginID: "Acme-linapro-org-core", want: "kebab-case"},
		{name: "overlong", pluginID: "acme-demo-" + strings.Repeat("x", catalog.MaxPluginIDLength), want: "length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := catalog.ValidatePluginID(tt.pluginID)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected plugin ID error containing %q, got %v", tt.want, err)
			}
			if !bizerr.Is(err, catalog.CodePluginIDInvalid) {
				t.Fatalf("expected stable plugin ID bizerr, got %v", err)
			}
			messageErr, ok := bizerr.As(err)
			if !ok {
				t.Fatalf("expected structured plugin ID bizerr, got %v", err)
			}
			if messageErr.MessageKey() != "error.plugin.id.invalid" {
				t.Fatalf("expected plugin ID message key, got %s", messageErr.MessageKey())
			}
		})
	}
}

// TestValidatePluginManifestNormalizesDependencyDefaults verifies dependency
// declarations are accepted and receive deterministic defaults.
func TestValidatePluginManifestNormalizesDependencyDefaults(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-dependency-valid")
	manifestFile := filepath.Join(pluginDir, "plugin.yaml")

	manifest := &catalog.Manifest{
		ID:      "acme-demo-dependency-valid",
		Name:    "Plugin Dependency Valid",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		Dependencies: &catalog.DependencySpec{
			Framework: &catalog.FrameworkDependencySpec{Version: " >=0.1.0 <1.0.0 "},
			Plugins: []*catalog.PluginDependencySpec{
				{
					ID:      " linapro-tenant-core ",
					Version: " >=0.1.0 ",
				},
				{
					ID:       "linapro-org-core",
					Version:  ">=0.1.0",
					Required: boolPtr(false),
					Install:  " auto ",
				},
			},
		},
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected dependency manifest to be valid, got error: %v", err)
	}
	if manifest.Dependencies.Framework.Version != ">=0.1.0 <1.0.0" {
		t.Fatalf("expected framework range to be trimmed, got %q", manifest.Dependencies.Framework.Version)
	}
	firstDependency := manifest.Dependencies.Plugins[0]
	if firstDependency.ID != "linapro-tenant-core" {
		t.Fatalf("expected dependency ID to be trimmed, got %q", firstDependency.ID)
	}
	if firstDependency.Required == nil || !*firstDependency.Required {
		t.Fatalf("expected required default true, got %#v", firstDependency.Required)
	}
	if firstDependency.Install != catalog.DependencyInstallModeManual.String() {
		t.Fatalf("expected install default manual, got %q", firstDependency.Install)
	}
	if manifest.Dependencies.Plugins[1].Install != catalog.DependencyInstallModeAuto.String() {
		t.Fatalf("expected explicit auto install mode, got %q", manifest.Dependencies.Plugins[1].Install)
	}
}

// TestValidatePluginManifestRejectsInvalidDependencies verifies dependency
// structural errors are caught during manifest validation.
func TestValidatePluginManifestRejectsInvalidDependencies(t *testing.T) {
	tests := []struct {
		name         string
		dependencies *catalog.DependencySpec
		want         string
	}{
		{
			name: "empty dependency id",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: ""}},
			},
			want: "missing id",
		},
		{
			name: "invalid dependency id",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: "Bad_ID"}},
			},
			want: "kebab-case",
		},
		{
			name: "self dependency",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: "acme-demo-dependency-invalid"}},
			},
			want: "cannot depend on itself",
		},
		{
			name: "duplicate dependency",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{
					{ID: "linapro-tenant-core"},
					{ID: "linapro-tenant-core"},
				},
			},
			want: "duplicate",
		},
		{
			name: "legacy shaped dependency id is allowed",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: "plugin-demo-source"}},
			},
			want: "",
		},
		{
			name: "invalid dependency version range",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: "linapro-tenant-core", Version: ">= v0.1.0"}},
			},
			want: "version",
		},
		{
			name: "invalid install mode",
			dependencies: &catalog.DependencySpec{
				Plugins: []*catalog.PluginDependencySpec{{ID: "linapro-tenant-core", Install: "sometimes"}},
			},
			want: "manual/auto",
		},
		{
			name: "invalid framework version range",
			dependencies: &catalog.DependencySpec{
				Framework: &catalog.FrameworkDependencySpec{Version: "0.1"},
			},
			want: "framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcs := testutil.NewServices()
			pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-dependency-invalid")
			manifest := &catalog.Manifest{
				ID:           "acme-demo-dependency-invalid",
				Name:         "Plugin Dependency Invalid",
				Version:      "0.1.0",
				Type:         catalog.TypeSource.String(),
				Dependencies: tt.dependencies,
			}

			err := svcs.Catalog.ValidateManifest(manifest, filepath.Join(pluginDir, "plugin.yaml"))
			if tt.want == "" {
				if err != nil {
					t.Fatalf("expected dependency validation to pass, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected dependency validation error containing %q, got %v", tt.want, err)
			}
		})
	}
}

// TestMatchesSemanticVersionRange verifies dependency version constraints use
// deterministic semver comparison semantics.
func TestMatchesSemanticVersionRange(t *testing.T) {
	matches, err := catalog.MatchesSemanticVersionRange("v0.6.1", ">=0.6.0 <0.7.0")
	if err != nil {
		t.Fatalf("expected range match to parse, got %v", err)
	}
	if !matches {
		t.Fatal("expected v0.6.1 to satisfy >=0.6.0 <0.7.0")
	}

	matches, err = catalog.MatchesSemanticVersionRange("v0.7.0", ">=0.6.0 <0.7.0")
	if err != nil {
		t.Fatalf("expected range mismatch to parse, got %v", err)
	}
	if matches {
		t.Fatal("expected v0.7.0 not to satisfy >=0.6.0 <0.7.0")
	}
}

// TestValidatePluginManifestRejectsMissingBackendEntryForSourcePlugin verifies
// that source plugins must still provide backend/plugin.go.
func TestValidatePluginManifestRejectsMissingBackendEntryForSourcePlugin(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-missing-backend")
	if err := os.Remove(filepath.Join(pluginDir, "backend", "plugin.go")); err != nil {
		t.Fatalf("failed to remove backend entry: %v", err)
	}

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:      "acme-demo-missing-backend",
		Name:    "Missing Backend Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
	}

	err := svcs.Catalog.ValidateManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "backend/plugin.go") {
		t.Fatalf("expected missing backend entry error, got: %v", err)
	}
}

// TestScanPluginManifestsReportsInvalidEmbeddedSourceManifest verifies an
// invalid registered source plugin remains a hard scan failure.
func TestScanPluginManifestsReportsInvalidEmbeddedSourceManifest(t *testing.T) {
	svcs := testutil.NewServices()

	const pluginID = "acme-demo-invalid-embedded"
	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte("id: acme-demo-invalid-embedded\nname: Invalid Plugin\nversion: invalid\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\n")},
	})
	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("failed to register invalid source plugin fixture: %v", err)
	}
	t.Cleanup(cleanup)

	_, scanErr := svcs.Catalog.ScanManifests()
	if scanErr == nil || !strings.Contains(scanErr.Error(), "version") {
		t.Fatalf("expected invalid embedded source manifest error, got: %v", scanErr)
	}
}

// TestValidateManifestUsesManifestRootDir verifies that source manifest
// validation resolves SQL assets from the manifest root instead of the current
// working directory.
func TestValidateManifestUsesManifestRootDir(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-manifest-rootdir")
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	manifest := &catalog.Manifest{
		ID:      "acme-demo-manifest-rootdir",
		Name:    "Manifest RootDir Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
	}
	if err := os.Remove(filepath.Join(pluginDir, "manifest", "sql", "001-acme-demo-manifest-rootdir.sql")); err != nil {
		t.Fatalf("failed to remove plugin install sql: %v", err)
	}
	if err := os.Remove(filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-acme-demo-manifest-rootdir.sql")); err != nil {
		t.Fatalf("failed to remove plugin uninstall sql: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "manifest", "sql"), 0o755); err != nil {
		t.Fatalf("failed to recreate sql dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest", "sql", "001-acme-demo-manifest-rootdir.sql"), []byte("SELECT 1;\n"), 0o644); err != nil {
		t.Fatalf("failed to write plugin install sql: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-acme-demo-manifest-rootdir.sql"), []byte("SELECT 1;\n"), 0o644); err != nil {
		t.Fatalf("failed to write plugin uninstall sql: %v", err)
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected manifest validation to use plugin root dir, got error: %v", err)
	}
}

// TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedWasmMetadata verifies
// that dynamic plugins validate from embedded runtime artifact metadata alone.
func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedWasmMetadata(t *testing.T) {
	svcs := testutil.NewServices()
	supportsMultiTenant := true
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"acme-demo-dynamic-valid",
		"Runtime Validation Plugin",
		"v0.2.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-valid.sql", Content: "SELECT 1;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-valid.sql", Content: "SELECT 2;"},
		},
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:                  "acme-demo-dynamic-valid",
		Name:                "Runtime Validation Plugin",
		Version:             "v0.2.0",
		Type:                catalog.TypeDynamic.String(),
		ScopeNature:         catalog.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		Description:         "A valid dynamic plugin manifest used by unit tests.",
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
	if !manifest.SupportsTenantGovernance() {
		t.Fatalf("expected dynamic manifest to keep supports_multi_tenant=true")
	}
	if manifest.ScopeNature != catalog.ScopeNatureTenantAware.String() || manifest.DefaultInstallMode != catalog.InstallModeTenantScoped.String() {
		t.Fatalf("unexpected dynamic tenant governance: scope=%s mode=%s", manifest.ScopeNature, manifest.DefaultInstallMode)
	}
}

// TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedFrontendAssets verifies
// that runtime artifacts carrying embedded frontend assets remain valid.
func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedFrontendAssets(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"acme-demo-dynamic-frontend",
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
		ID:      "acme-demo-dynamic-frontend",
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

// TestValidatePluginManifestRejectsMismatchedRuntimeWasmManifest verifies that
// embedded runtime manifest identity must match the outer plugin manifest.
func TestValidatePluginManifestRejectsMismatchedRuntimeWasmManifest(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"acme-demo-dynamic-mismatch",
		"Runtime Mismatch Plugin",
		"v0.3.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	testutil.WriteRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, runtime.BuildArtifactRelativePath("acme-demo-dynamic-mismatch")),
		&catalog.ArtifactManifest{
			ID:      "acme-demo-dynamic-other",
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
			{Key: "001-acme-demo-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
		nil,
		nil,
		nil,
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &catalog.Manifest{
		ID:      "acme-demo-dynamic-mismatch",
		Name:    "Runtime Mismatch Plugin",
		Version: "v0.3.0",
		Type:    catalog.TypeDynamic.String(),
	}

	err := svcs.Catalog.ValidateManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "embedded manifest ID") {
		t.Fatalf("expected embedded manifest mismatch error, got: %v", err)
	}
}

// TestScanPluginManifestsRejectsDuplicatePluginIDs verifies that discovery
// fails fast when a registered source plugin and runtime artifact share an ID.
func TestScanPluginManifestsRejectsDuplicatePluginIDs(t *testing.T) {
	svcs := testutil.NewServices()
	pluginID := "acme-demo-duplicate-id"

	testutil.CreateTestPluginDir(t, pluginID)
	testutil.CreateTestRuntimeStorageArtifact(t, pluginID, "Duplicate Runtime Plugin", "v0.1.0", nil, nil)

	_, err := svcs.Catalog.ScanManifests()
	if err == nil || !strings.Contains(err.Error(), "plugin ID is duplicated") {
		t.Fatalf("expected duplicate plugin id error, got: %v", err)
	}
}

// TestScanPluginManifestsRejectsDuplicateRuntimeArtifactPluginIDs verifies that
// runtime artifact discovery rejects duplicate dynamic plugin IDs.
func TestScanPluginManifestsRejectsDuplicateRuntimeArtifactPluginIDs(t *testing.T) {
	svcs := testutil.NewServices()

	testutil.CreateTestRuntimeStorageArtifactWithFilename(
		t,
		"acme-demo-dynamic-duplicate-a.wasm",
		"acme-demo-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)
	testutil.CreateTestRuntimeStorageArtifactWithFilename(
		t,
		"acme-demo-dynamic-duplicate-b.wasm",
		"acme-demo-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)

	_, err := svcs.Catalog.ScanManifests()
	if err == nil || !strings.Contains(err.Error(), "dynamic plugin ID is duplicated") {
		t.Fatalf("expected duplicate dynamic plugin id error, got: %v", err)
	}
}

// TestStoreUploadedRuntimePackageWritesCanonicalWasmIntoRuntimeStorage verifies
// that uploaded runtime packages are persisted at the canonical storage path.
func TestStoreUploadedRuntimePackageWritesCanonicalWasmIntoRuntimeStorage(t *testing.T) {
	svcs := testutil.NewServices()
	ctx := context.Background()

	pluginID := "acme-demo-dynamic-upload-storage"
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

// TestDiscoverPluginSQLPathsUsesDirectoryConvention verifies install and
// uninstall SQL discovery follows the source-plugin directory convention.
func TestDiscoverPluginSQLPathsUsesDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-sql-convention")

	installPaths := svcs.Catalog.DiscoverSQLPaths(pluginDir, false)
	if len(installPaths) != 1 || installPaths[0] != "manifest/sql/001-acme-demo-sql-convention.sql" {
		t.Fatalf("unexpected install sql paths: %#v", installPaths)
	}

	uninstallPaths := svcs.Catalog.DiscoverSQLPaths(pluginDir, true)
	if len(uninstallPaths) != 1 || uninstallPaths[0] != "manifest/sql/uninstall/001-acme-demo-sql-convention.sql" {
		t.Fatalf("unexpected uninstall sql paths: %#v", uninstallPaths)
	}
}

// TestDiscoverPluginVuePathsUseDirectoryConvention verifies page and slot
// discovery follows the source-plugin frontend directory convention.
func TestDiscoverPluginVuePathsUseDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-vue-convention")

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

// TestBuildPluginManifestSnapshotIncludesDirectoryDiscoveredAssets verifies
// source-plugin snapshots include discovered page, slot, and SQL counts.
func TestBuildPluginManifestSnapshotIncludesDirectoryDiscoveredAssets(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-snapshot")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	testutil.WriteTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	snapshot, err := svcs.Catalog.BuildManifestSnapshot(&catalog.Manifest{
		ID:          "acme-demo-snapshot",
		Name:        "Snapshot Plugin",
		Version:     "0.1.0",
		Type:        catalog.TypeSource.String(),
		Description: "Snapshot test plugin",
		Menus: []*catalog.MenuSpec{
			{
				Key:  "plugin:acme-demo-snapshot:sidebar-entry",
				Name: "Snapshot Plugin",
				Type: catalog.MenuTypePage.String(),
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

// TestBuildPluginManifestSnapshotIncludesRuntimeArtifactMetadata verifies that
// dynamic snapshots include runtime artifact metadata and bridge settings.
func TestBuildPluginManifestSnapshotIncludesRuntimeArtifactMetadata(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"acme-demo-dynamic-snapshot",
		"Runtime Snapshot Plugin",
		"v0.4.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-snapshot.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "acme-demo-dynamic-snapshot",
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

// TestBuildPluginResourceRefDescriptorsDoNotPersistConcreteFilePaths verifies
// that governance descriptors store abstract identities instead of file paths.
func TestBuildPluginResourceRefDescriptorsDoNotPersistConcreteFilePaths(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-resource-summary")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	testutil.WriteTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	descriptors := svcs.Integration.BuildResourceRefDescriptors(&catalog.Manifest{
		ID:      "acme-demo-resource-summary",
		Name:    "Resource Summary Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		Menus: []*catalog.MenuSpec{
			{
				Key:  "plugin:acme-demo-resource-summary:sidebar-entry",
				Name: "Resource Summary Plugin",
				Type: catalog.MenuTypePage.String(),
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

// TestBuildPluginResourceRefDescriptorsSummarizeRuntimeArtifact verifies that
// runtime governance descriptors summarize artifact capabilities and assets.
func TestBuildPluginResourceRefDescriptorsSummarizeRuntimeArtifact(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"acme-demo-dynamic-resource-summary",
		"Runtime Resource Summary Plugin",
		"v0.5.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-resource-summary.sql", Content: "SELECT 1;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-resource-summary.sql", Content: "SELECT 2;"},
		},
	)

	manifest := &catalog.Manifest{
		ID:           "acme-demo-dynamic-resource-summary",
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

// TestResolvePluginSQLAssetsPrefersEmbeddedRuntimeSQL verifies that dynamic
// plugins prefer embedded SQL assets over source-directory conventions.
func TestResolvePluginSQLAssetsPrefersEmbeddedRuntimeSQL(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"acme-demo-dynamic-sql-assets",
		"Runtime SQL Assets Plugin",
		"v0.6.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-sql-assets.sql", Content: "SELECT 1;"},
			{Key: "002-acme-demo-dynamic-sql-assets.sql", Content: "SELECT 2;"},
		},
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-acme-demo-dynamic-sql-assets.sql", Content: "SELECT 3;"},
		},
	)

	manifest := &catalog.Manifest{
		ID:           "acme-demo-dynamic-sql-assets",
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
	if len(installAssets) != 2 || installAssets[0].Key != "001-acme-demo-dynamic-sql-assets.sql" {
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

// TestResolvePluginSQLAssetsFallsBackToDirectoryConvention verifies that
// source plugins resolve SQL assets from their directory structure.
func TestResolvePluginSQLAssetsFallsBackToDirectoryConvention(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-directory-sql-assets")

	manifest := &catalog.Manifest{
		ID:           "acme-demo-directory-sql-assets",
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
	if len(installAssets) != 1 || installAssets[0].Key != "001-acme-demo-directory-sql-assets.sql" {
		t.Fatalf("unexpected directory install assets: %#v", installAssets)
	}
}

// TestScanEmbeddedSourcePluginManifestsUsesPluginEmbeddedFiles verifies that
// embedded source plugins are scanned from their packaged file sets.
func TestScanEmbeddedSourcePluginManifestsUsesPluginEmbeddedFiles(t *testing.T) {
	svcs := testutil.NewServices()

	const pluginID = "acme-demo-embedded-manifest"
	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml":                                &fstest.MapFile{Data: []byte("id: acme-demo-embedded-manifest\nname: Embedded Manifest Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n")},
		"frontend/pages/main-entry.vue":              &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"frontend/slots/layout.header.after/tip.vue": &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"manifest/sql/001-acme-demo-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 1;\n"),
		},
		"manifest/sql/uninstall/001-acme-demo-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 2;\n"),
		},
	})
	if err := pluginhost.RegisterSourcePlugin(sourcePlugin); err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}

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
	if target.ManifestPath != "embedded/source-plugins/acme-demo-embedded-manifest/plugin.yaml" {
		t.Fatalf("unexpected embedded manifest path: %s", target.ManifestPath)
	}
	if len(svcs.Catalog.ListFrontendPagePaths(target)) != 1 {
		t.Fatalf("expected embedded frontend page paths to be discovered")
	}
	if len(svcs.Catalog.ListFrontendSlotPaths(target)) != 1 {
		t.Fatalf("expected embedded frontend slot paths to be discovered")
	}
}

// TestResolvePluginSQLAssetsUsesEmbeddedSourcePluginFiles verifies that
// embedded source plugins resolve SQL assets from embedded filesystems.
func TestResolvePluginSQLAssetsUsesEmbeddedSourcePluginFiles(t *testing.T) {
	svcs := testutil.NewServices()

	manifest := &catalog.Manifest{
		ID:      "acme-demo-embedded-sql-assets",
		Name:    "Embedded SQL Assets Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		SourcePlugin: func() pluginhost.SourcePluginDefinition {
			sourcePlugin := pluginhost.NewSourcePlugin("acme-demo-embedded-sql-assets")
			sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
				"plugin.yaml": &fstest.MapFile{Data: []byte("id: acme-demo-embedded-sql-assets\nname: Embedded SQL Assets Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n")},
				"manifest/sql/001-acme-demo-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 1;\n"),
				},
				"manifest/sql/uninstall/001-acme-demo-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 2;\n"),
				},
			})
			definition, ok := sourcePlugin.(pluginhost.SourcePluginDefinition)
			if !ok {
				t.Fatalf("expected embedded source plugin to expose host definition view")
			}
			return definition
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

// TestGetRegistryReleaseFallsBackWhenReleasePointerIsDangling verifies that
// catalog reads tolerate registry rows whose release_id no longer points to an
// existing release row.
func TestGetRegistryReleaseFallsBackWhenReleasePointerIsDangling(t *testing.T) {
	var (
		ctx      = context.Background()
		svcs     = testutil.NewServices()
		pluginID = "acme-demo-dangling-release-pointer"
		version  = "9.9.9"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:     pluginID,
		Name:         "Dangling Release Pointer Plugin",
		Version:      version,
		Type:         catalog.TypeDynamic.String(),
		Installed:    catalog.InstalledYes,
		Status:       catalog.StatusEnabled,
		DesiredState: catalog.LifecycleStateRuntimeEnabled.String(),
		CurrentState: catalog.LifecycleStateRuntimeEnabled.String(),
		Generation:   int64(1),
		ReleaseId:    987654321,
		ScopeNature:  catalog.ScopeNatureTenantAware.String(),
		InstallMode:  catalog.InstallModeTenantScoped.String(),
		ManifestPath: "runtime/acme-demo-dangling-release-pointer/plugin.yaml",
		Checksum:     "dangling-release-pointer",
		Remark:       "Dangling release pointer test plugin",
	}).InsertAndGetId(); err != nil {
		t.Fatalf("failed to insert plugin registry row: %v", err)
	}
	insertID, err := dao.SysPluginRelease.Ctx(ctx).Data(do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: version,
		Type:           catalog.TypeDynamic.String(),
		RuntimeKind:    pluginbridge.RuntimeKindWasm,
		Status:         catalog.ReleaseStatusActive.String(),
		ManifestPath:   "runtime/acme-demo-dangling-release-pointer/plugin.yaml",
		PackagePath:    "runtime/acme-demo-dangling-release-pointer.wasm",
		Checksum:       "dangling-release-pointer",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("failed to insert fallback plugin release row: %v", err)
	}

	registry, err := svcs.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	release, err := svcs.Catalog.GetRegistryRelease(ctx, registry)
	if err != nil {
		t.Fatalf("expected dangling release pointer to fall back to plugin version, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected fallback release to be returned")
	}
	if release.Id != int(insertID) {
		t.Fatalf("expected fallback release id %d, got %d", insertID, release.Id)
	}
}

// TestStartupDataSnapshotReusesReleaseByIDAndVersion verifies one catalog
// snapshot backs both release lookup shapes and can be explicitly refreshed
// after the authority database row changes.
func TestStartupDataSnapshotReusesReleaseByIDAndVersion(t *testing.T) {
	var (
		ctx      = context.Background()
		svcs     = testutil.NewServices()
		pluginID = "acme-demo-release-snapshot-reuse"
		version  = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Release Snapshot Reuse",
		Version:            version,
		Type:               catalog.TypeDynamic.String(),
		ScopeNature:        catalog.ScopeNatureTenantAware.String(),
		DefaultInstallMode: catalog.InstallModeTenantScoped.String(),
	}
	registry, err := svcs.Catalog.SyncManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected manifest sync to create registry and release, got error: %v", err)
	}
	if err = svcs.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected installed marker update to succeed, got error: %v", err)
	}
	registry, err = svcs.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	registry, err = svcs.Catalog.SyncRegistryReleaseReference(ctx, registry, manifest)
	if err != nil {
		t.Fatalf("expected release reference sync to succeed, got error: %v", err)
	}
	if registry == nil || registry.ReleaseId <= 0 {
		t.Fatalf("expected registry to point at release, got %#v", registry)
	}

	snapshotCtx, err := svcs.Catalog.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("expected catalog snapshot to build, got error: %v", err)
	}
	byID, err := svcs.Catalog.GetReleaseByID(snapshotCtx, registry.ReleaseId)
	if err != nil {
		t.Fatalf("expected release lookup by id to succeed, got error: %v", err)
	}
	byVersion, err := svcs.Catalog.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected release lookup by plugin version to succeed, got error: %v", err)
	}
	if byID == nil || byVersion == nil || byID.Id != byVersion.Id {
		t.Fatalf("expected both lookup shapes to return the same release, got byID=%#v byVersion=%#v", byID, byVersion)
	}

	const refreshedChecksum = "release-snapshot-refreshed"
	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: registry.ReleaseId}).
		Data(do.SysPluginRelease{Checksum: refreshedChecksum}).
		Update(); err != nil {
		t.Fatalf("expected release checksum update to succeed, got error: %v", err)
	}
	staleByVersion, err := svcs.Catalog.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected stale snapshot lookup to succeed, got error: %v", err)
	}
	if staleByVersion == nil || staleByVersion.Checksum == refreshedChecksum {
		t.Fatalf("expected request snapshot to remain stable before explicit refresh, got %#v", staleByVersion)
	}

	refreshed, err := svcs.Catalog.RefreshStartupReleaseByID(snapshotCtx, registry.ReleaseId)
	if err != nil {
		t.Fatalf("expected snapshot refresh to succeed, got error: %v", err)
	}
	if refreshed == nil || refreshed.Checksum != refreshedChecksum {
		t.Fatalf("expected refreshed release checksum %s, got %#v", refreshedChecksum, refreshed)
	}
	refreshedByVersion, err := svcs.Catalog.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected refreshed version lookup to succeed, got error: %v", err)
	}
	if refreshedByVersion == nil || refreshedByVersion.Checksum != refreshedChecksum {
		t.Fatalf("expected version index to use refreshed release checksum %s, got %#v", refreshedChecksum, refreshedByVersion)
	}
}

// TestRuntimeUpgradeStateReportsExplicitRunningMarker verifies management
// projections expose upgrade_running while an explicit runtime upgrade is in progress.
func TestRuntimeUpgradeStateReportsExplicitRunningMarker(t *testing.T) {
	var (
		ctx        = context.Background()
		svcs       = testutil.NewServices()
		pluginID   = "acme-demo-runtime-upgrade-running-marker"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	oldManifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Runtime Upgrade Running Marker",
		Version:            oldVersion,
		Type:               catalog.TypeDynamic.String(),
		ScopeNature:        catalog.ScopeNatureTenantAware.String(),
		DefaultInstallMode: catalog.InstallModeTenantScoped.String(),
	}
	registry, err := svcs.Catalog.SyncManifest(ctx, oldManifest)
	if err != nil {
		t.Fatalf("expected old manifest sync to succeed, got error: %v", err)
	}
	oldRelease, err := svcs.Catalog.GetRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old release row")
	}
	if err = svcs.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected installed state update to succeed, got error: %v", err)
	}
	registry, err = svcs.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after install marker to succeed, got error: %v", err)
	}
	registry, err = svcs.Catalog.SyncRegistryReleaseReference(ctx, registry, oldManifest)
	if err != nil {
		t.Fatalf("expected registry release reference sync to succeed, got error: %v", err)
	}
	if err = svcs.Catalog.UpdateReleaseState(ctx, oldRelease.Id, catalog.ReleaseStatusInstalled, ""); err != nil {
		t.Fatalf("expected old release state update to succeed, got error: %v", err)
	}

	newManifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Runtime Upgrade Running Marker",
		Version:            newVersion,
		Type:               catalog.TypeDynamic.String(),
		ScopeNature:        catalog.ScopeNatureTenantAware.String(),
		DefaultInstallMode: catalog.InstallModeTenantScoped.String(),
	}
	if _, err = svcs.Catalog.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new manifest sync to succeed, got error: %v", err)
	}
	if err = svcs.Catalog.SetRegistryRuntimeState(ctx, pluginID, do.SysPlugin{
		CurrentState: catalog.RuntimeUpgradeStateUpgradeRunning.String(),
	}); err != nil {
		t.Fatalf("expected running marker update to succeed, got error: %v", err)
	}

	registry, err = svcs.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	projection, err := svcs.Catalog.BuildRuntimeUpgradeState(ctx, registry, newManifest)
	if err != nil {
		t.Fatalf("expected runtime state projection to succeed, got error: %v", err)
	}
	if projection.State != catalog.RuntimeUpgradeStateUpgradeRunning {
		t.Fatalf("expected upgrade_running projection, got %#v", projection)
	}
}

// TestNormalizePluginStatusEnums verifies raw database flags are normalized
// into the new strongly typed plugin status enums before state derivation runs.
func TestNormalizePluginStatusEnums(t *testing.T) {
	if status := catalog.NormalizeStatus(1); status != catalog.PluginStatusEnabled {
		t.Fatalf("expected enabled status enum, got %#v", status)
	}
	if status := catalog.NormalizeStatus(99); status != catalog.PluginStatusDisabled {
		t.Fatalf("expected unknown status to normalize to disabled, got %#v", status)
	}
	if installed := catalog.NormalizeInstalledStatus(1); installed != catalog.PluginInstalledYes {
		t.Fatalf("expected installed status enum, got %#v", installed)
	}
	if installed := catalog.NormalizeInstalledStatus(-1); installed != catalog.PluginInstalledNo {
		t.Fatalf("expected unknown install flag to normalize to uninstalled, got %#v", installed)
	}
}

// TestDerivePluginLifecycleState verifies lifecycle-state derivation from
// installed, enabled, and failure flags.
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

// TestDerivePluginNodeState verifies node-state derivation from install and
// enablement signals exposed by governance projections.
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

// TestValidateManifestMenusAcceptsExternalParentKey verifies manifest
// structure validation does not impose host-owned menu placement policy.
func TestValidateManifestMenusAcceptsExternalParentKey(t *testing.T) {
	manifest := &catalog.Manifest{
		ID: "custom-parent-validation",
		Menus: []*catalog.MenuSpec{
			{
				Key:       "plugin:custom-parent-validation:main",
				Name:      "Custom Parent Validation",
				ParentKey: "system",
				Path:      "/custom-parent-validation",
				Type:      catalog.MenuTypePage.String(),
			},
		},
	}

	if err := catalog.ValidateManifestMenus(manifest); err != nil {
		t.Fatalf("expected plugin manifest with external parent key to be valid, got: %v", err)
	}
}

// TestValidateManifestMenusAcceptsCrossPluginParentKey verifies plugins may
// declare an external plugin menu as parent and let runtime sync resolve it.
func TestValidateManifestMenusAcceptsCrossPluginParentKey(t *testing.T) {
	manifest := &catalog.Manifest{
		ID: "linapro-org-core",
		Menus: []*catalog.MenuSpec{
			{
				Key:       "plugin:linapro-org-core:catalog",
				Name:      "组织管理",
				ParentKey: "plugin:linapro-content-notice:notice",
				Path:      "linapro-org-core-catalog",
				Type:      catalog.MenuTypeDirectory.String(),
			},
		},
	}

	if err := catalog.ValidateManifestMenus(manifest); err != nil {
		t.Fatalf("expected plugin manifest with cross-plugin parent key to be valid, got: %v", err)
	}
}

// TestValidateManifestMenusAcceptsInternalPluginParent verifies plugin menus
// can keep children inside their own manifest-declared tree.
func TestValidateManifestMenusAcceptsInternalPluginParent(t *testing.T) {
	manifest := &catalog.Manifest{
		ID: "linapro-org-core",
		Menus: []*catalog.MenuSpec{
			{
				Key:       "plugin:linapro-org-core:catalog",
				Name:      "组织管理",
				ParentKey: menusvc.Org,
				Path:      "linapro-org-core-catalog",
				Type:      catalog.MenuTypeDirectory.String(),
			},
			{
				Key:       "plugin:linapro-org-core:dept",
				Name:      "部门管理",
				ParentKey: "plugin:linapro-org-core:catalog",
				Path:      "/system/dept",
				Component: "system/plugin/dynamic-page",
				Type:      catalog.MenuTypePage.String(),
			},
		},
	}

	if err := catalog.ValidateManifestMenus(manifest); err != nil {
		t.Fatalf("expected plugin manifest menus with internal parent to be valid, got: %v", err)
	}
}

// TestValidateManifestMenusAcceptsCustomTenantParent verifies a custom parent
// key is accepted during manifest validation and resolved during menu sync.
func TestValidateManifestMenusAcceptsCustomTenantParent(t *testing.T) {
	manifest := &catalog.Manifest{
		ID: "linapro-tenant-core",
		Menus: []*catalog.MenuSpec{
			{
				Key:       "plugin:linapro-tenant-core:tenant:members",
				Name:      "成员管理",
				ParentKey: "tenant",
				Path:      "/tenant/members",
				Type:      catalog.MenuTypePage.String(),
			},
		},
	}

	if err := catalog.ValidateManifestMenus(manifest); err != nil {
		t.Fatalf("expected plugin manifest with custom parent key to be valid, got: %v", err)
	}
}

// TestValidateManifestNormalizesTenantGovernance verifies tenant governance
// manifest fields have deterministic normalization and platform-only constraints.
func TestValidateManifestNormalizesTenantGovernance(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-tenant-governance")
	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	supportsMultiTenant := false

	manifest := &catalog.Manifest{
		ID:                  "acme-demo-tenant-governance",
		Name:                "Tenant Governance Plugin",
		Version:             "0.1.0",
		Type:                catalog.TypeSource.String(),
		ScopeNature:         catalog.ScopeNaturePlatformOnly.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected manifest to validate, got %v", err)
	}
	if manifest.ScopeNature != catalog.ScopeNaturePlatformOnly.String() {
		t.Fatalf("expected platform-only scope, got %s", manifest.ScopeNature)
	}
	if manifest.DefaultInstallMode != catalog.InstallModeGlobal.String() {
		t.Fatalf("expected platform-only plugin to force global install mode, got %s", manifest.DefaultInstallMode)
	}
	if manifest.SupportsTenantGovernance() {
		t.Fatalf("expected platform-only plugin to disable tenant governance support")
	}
}

// TestValidateManifestRequiresMultiTenantSupportDeclaration verifies plugin
// manifests must explicitly declare whether tenant governance is supported.
func TestValidateManifestRequiresMultiTenantSupportDeclaration(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-tenant-governance-missing-support")
	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifestFile,
		"id: acme-demo-tenant-governance-missing-support\nname: Tenant Governance Missing Support Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\ndefault_install_mode: tenant_scoped\n",
	)

	manifest := &catalog.Manifest{
		ID:      "acme-demo-tenant-governance-missing-support",
		Name:    "Tenant Governance Missing Support Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
	}

	err := svcs.Catalog.ValidateManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "supports_multi_tenant is required") {
		t.Fatalf("expected missing supports_multi_tenant validation error, got %v", err)
	}
}

// TestValidateManifestForcesGlobalWhenTenantGovernanceUnsupported verifies
// tenant-aware plugins can explicitly opt out of tenant-level governance.
func TestValidateManifestForcesGlobalWhenTenantGovernanceUnsupported(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "acme-demo-tenant-governance-unsupported")
	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	supportsMultiTenant := false

	manifest := &catalog.Manifest{
		ID:                  "acme-demo-tenant-governance-unsupported",
		Name:                "Tenant Governance Unsupported Plugin",
		Version:             "0.1.0",
		Type:                catalog.TypeSource.String(),
		ScopeNature:         catalog.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
	}

	if err := svcs.Catalog.ValidateManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected manifest to validate, got %v", err)
	}
	if manifest.DefaultInstallMode != catalog.InstallModeGlobal.String() {
		t.Fatalf("expected unsupported tenant governance to force global install mode, got %s", manifest.DefaultInstallMode)
	}
	if manifest.SupportsTenantGovernance() {
		t.Fatalf("expected explicit supports_multi_tenant=false to disable tenant governance")
	}
}

// boolPtr returns a pointer to value for concise manifest fixtures.
func boolPtr(value bool) *bool {
	return &value
}
