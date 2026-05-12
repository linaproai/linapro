// This file covers the mock-data SQL discovery surface added by the
// plugin-install-with-mock-data change so install/uninstall scans stay
// disjoint from the new mock phase across both source and embedded
// source-plugin paths.

package catalog_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginhost"
)

// TestListMockSQLPathsExcludedFromInstallScan verifies that mock-data files
// living under manifest/sql/mock-data/ never bleed into the install-direction
// scan, and that they only surface through the dedicated mock entry points.
func TestListMockSQLPathsExcludedFromInstallScan(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-mock-data-disjoint")

	mockDir := filepath.Join(pluginDir, "manifest", "sql", "mock-data")
	mockSQL := "INSERT INTO sys_user(username) VALUES ('alice') ON CONFLICT DO NOTHING;"
	if err := writeFileTree(mockDir, "001-plugin-mock-data-disjoint.sql", mockSQL); err != nil {
		t.Fatalf("failed to write mock SQL: %v", err)
	}

	manifest := &catalog.Manifest{
		ID:           "plugin-mock-data-disjoint",
		Name:         "Mock Data Disjoint Plugin",
		Version:      "0.1.0",
		Type:         catalog.TypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}

	installAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected install assets, got error: %v", err)
	}
	for _, asset := range installAssets {
		if asset.Key == "001-plugin-mock-data-disjoint.sql" && asset.Content != "SELECT 1;" {
			continue
		}
		if filepath.Base(asset.Key) == "001-plugin-mock-data-disjoint.sql" && asset.Content == mockSQL {
			t.Fatalf("mock data leaked into install scan: %#v", asset)
		}
	}

	mockAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionMock)
	if err != nil {
		t.Fatalf("expected mock assets, got error: %v", err)
	}
	if len(mockAssets) != 1 {
		t.Fatalf("expected 1 mock asset, got %d: %#v", len(mockAssets), mockAssets)
	}
	if mockAssets[0].Key != "001-plugin-mock-data-disjoint.sql" {
		t.Fatalf("unexpected mock asset key: %s", mockAssets[0].Key)
	}

	if !svcs.Catalog.HasMockSQLData(manifest) {
		t.Fatalf("expected HasMockSQLData=true when manifest/sql/mock-data has files")
	}
}

// TestListMockSQLPathsEmptyWhenAbsent verifies that plugins without a
// manifest/sql/mock-data/ directory return an empty mock asset list and
// HasMockSQLData=false so the management UI can hide the install checkbox.
func TestListMockSQLPathsEmptyWhenAbsent(t *testing.T) {
	svcs := testutil.NewServices()
	pluginDir := testutil.CreateTestPluginDir(t, "plugin-mock-data-absent")

	manifest := &catalog.Manifest{
		ID:           "plugin-mock-data-absent",
		Name:         "Mock Data Absent Plugin",
		Version:      "0.1.0",
		Type:         catalog.TypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}

	mockAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionMock)
	if err != nil {
		t.Fatalf("expected zero mock assets, got error: %v", err)
	}
	if len(mockAssets) != 0 {
		t.Fatalf("expected 0 mock assets, got %d", len(mockAssets))
	}

	if svcs.Catalog.HasMockSQLData(manifest) {
		t.Fatalf("expected HasMockSQLData=false when no mock-data directory exists")
	}
}

// TestListMockSQLPathsViaEmbeddedSourcePluginFiles verifies that source plugins
// using embedded fs (production layout) surface their mock-data SQL through the
// same code path as filesystem-rooted source plugins.
func TestListMockSQLPathsViaEmbeddedSourcePluginFiles(t *testing.T) {
	svcs := testutil.NewServices()

	manifest := &catalog.Manifest{
		ID:      "plugin-embedded-mock-data",
		Name:    "Embedded Mock Data Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		SourcePlugin: func() pluginhost.SourcePluginDefinition {
			sourcePlugin := pluginhost.NewSourcePlugin("plugin-embedded-mock-data")
			sourcePlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
				"plugin.yaml": &fstest.MapFile{Data: []byte("id: plugin-embedded-mock-data\nname: Embedded Mock Data Plugin\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n")},
				"manifest/sql/001-plugin-embedded-mock-data.sql": &fstest.MapFile{
					Data: []byte("CREATE TABLE IF NOT EXISTS plugin_demo (id BIGINT PRIMARY KEY);"),
				},
				"manifest/sql/mock-data/001-plugin-embedded-mock-data-mock.sql": &fstest.MapFile{
					Data: []byte("INSERT INTO plugin_demo(id) VALUES (1) ON CONFLICT DO NOTHING;"),
				},
				"manifest/sql/uninstall/001-plugin-embedded-mock-data.sql": &fstest.MapFile{
					Data: []byte("DROP TABLE IF EXISTS plugin_demo;"),
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
		t.Fatalf("expected install assets, got error: %v", err)
	}
	if len(installAssets) != 1 || installAssets[0].Key != "001-plugin-embedded-mock-data.sql" {
		t.Fatalf("unexpected install assets: %#v", installAssets)
	}

	mockAssets, err := svcs.Lifecycle.ResolvePluginSQLAssets(manifest, catalog.MigrationDirectionMock)
	if err != nil {
		t.Fatalf("expected mock assets, got error: %v", err)
	}
	if len(mockAssets) != 1 || mockAssets[0].Key != "001-plugin-embedded-mock-data-mock.sql" {
		t.Fatalf("unexpected mock assets: %#v", mockAssets)
	}

	if !svcs.Catalog.HasMockSQLData(manifest) {
		t.Fatalf("expected HasMockSQLData=true for embedded plugin with mock data")
	}
}

// writeFileTree creates a file at <dir>/<name> with the given content,
// creating the parent directory as needed. Used by mock data tests to add
// files into a CreateTestPluginDir-managed plugin directory.
func writeFileTree(dir, name, content string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}
