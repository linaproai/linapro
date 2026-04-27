// This file verifies ResourceLoader behavior across host, source-plugin, and
// already-extracted dynamic-plugin i18n resources.

package i18nresource

import (
	"context"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

// fakeSourcePlugin provides a test-only source-plugin resource container.
type fakeSourcePlugin struct {
	id         string
	filesystem fs.FS
}

// ID returns the test plugin identifier.
func (p fakeSourcePlugin) ID() string {
	return p.id
}

// GetEmbeddedFiles returns the plugin-owned test filesystem.
func (p fakeSourcePlugin) GetEmbeddedFiles() fs.FS {
	return p.filesystem
}

// TestParseCatalogSupportsNestedFlatAndScalarModes verifies flat key migration
// behavior and the two supported JSON value conversion modes.
func TestParseCatalogSupportsNestedFlatAndScalarModes(t *testing.T) {
	t.Parallel()

	catalog, err := ParseCatalog([]byte(`{
  "menu": {
    "dashboard": {
      "title": "Nested Workbench"
    }
  },
  "menu.dashboard.title": "Flat Workbench",
  "feature": {
    "enabled": true
  }
}`), ValueModeStringifyScalars)
	if err != nil {
		t.Fatalf("expected stringify catalog parse to succeed: %v", err)
	}
	if actual := catalog["menu.dashboard.title"]; actual != "Flat Workbench" {
		t.Fatalf("expected flat key to override nested value, got %q", actual)
	}
	if actual := catalog["feature.enabled"]; actual != "true" {
		t.Fatalf("expected scalar value to be stringified, got %q", actual)
	}

	_, err = ParseCatalog([]byte(`{"feature":{"enabled":true}}`), ValueModeStringOnly)
	if err == nil {
		t.Fatal("expected string-only catalog parse to reject non-string leaf values")
	}
}

// TestLoadHostBundleMergesLocaleFileAndDirectory verifies apidoc-style
// resources merge the root locale file before sorted per-locale directory files.
func TestLoadHostBundleMergesLocaleFileAndDirectory(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		HostFS: fstest.MapFS{
			"manifest/i18n/apidoc/zh-CN.json": &fstest.MapFile{Data: []byte(`{
  "core": {
    "title": "Root Title"
  },
  "core.summary": "Root Summary"
}`)},
			"manifest/i18n/apidoc/zh-CN/00-base.json": &fstest.MapFile{Data: []byte(`{
  "core.title": "Directory Title",
  "core.description": "Base Description"
}`)},
			"manifest/i18n/apidoc/zh-CN/10-override.json": &fstest.MapFile{Data: []byte(`{
  "core": {
    "description": "Override Description"
  }
}`)},
			"manifest/i18n/apidoc/zh-CN/readme.txt": &fstest.MapFile{Data: []byte("ignored")},
		},
		Subdir:     "manifest/i18n/apidoc",
		LayoutMode: LayoutModeLocaleFileAndDirectory,
		ValueMode:  ValueModeStringOnly,
	}

	expected := map[string]string{
		"core.title":       "Directory Title",
		"core.summary":     "Root Summary",
		"core.description": "Override Description",
	}
	if actual := loader.LoadHostBundle(context.Background(), "zh-CN"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected host bundle: expected=%v actual=%v", expected, actual)
	}
}

// TestLoadSourcePluginBundlesLoadsEachPlugin verifies source-plugin resources
// are returned per plugin so callers can keep source attribution.
func TestLoadSourcePluginBundlesLoadsEachPlugin(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		SourcePlugins: func() []SourcePlugin {
			return []SourcePlugin{
				fakeSourcePlugin{id: "z-plugin", filesystem: fstest.MapFS{
					"manifest/i18n/en-US.json": &fstest.MapFile{Data: []byte(`{"plugin.z.name":"Z Plugin"}`)},
				}},
				fakeSourcePlugin{id: "a-plugin", filesystem: fstest.MapFS{
					"manifest/i18n/en-US.json": &fstest.MapFile{Data: []byte(`{"plugin.a.name":"A Plugin"}`)},
				}},
			}
		},
		Subdir:    "manifest/i18n",
		ValueMode: ValueModeStringifyScalars,
	}

	actual := loader.LoadSourcePluginBundles(context.Background(), "en-US")
	if actual["a-plugin"]["plugin.a.name"] != "A Plugin" {
		t.Fatalf("expected a-plugin bundle to load, got %v", actual["a-plugin"])
	}
	if actual["z-plugin"]["plugin.z.name"] != "Z Plugin" {
		t.Fatalf("expected z-plugin bundle to load, got %v", actual["z-plugin"])
	}
}

// TestRestrictedPluginScopeDropsForeignKeys verifies plugin-owned apidoc
// resources cannot override host or sibling-plugin keys.
func TestRestrictedPluginScopeDropsForeignKeys(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		SourcePlugins: func() []SourcePlugin {
			return []SourcePlugin{
				fakeSourcePlugin{id: "plugin-demo-dynamic", filesystem: fstest.MapFS{
					"manifest/i18n/apidoc/zh-CN.json": &fstest.MapFile{Data: []byte(`{
  "plugins": {
    "plugin_demo_dynamic": {
      "name": "Allowed Plugin Name",
      "internal": {
        "model": {
          "entity": {
            "User": {
              "name": "Rejected Entity Metadata"
            }
          }
        }
      }
    },
    "other_plugin": {
      "name": "Rejected Sibling Plugin Name"
    }
  },
  "core.title": "Rejected Host Title"
}`)},
				}},
			}
		},
		Subdir:      "manifest/i18n/apidoc",
		PluginScope: PluginScopeRestrictedToPluginNamespace,
		ValueMode:   ValueModeStringOnly,
		KeyFilter: func(key string) bool {
			return key != "plugins.plugin_demo_dynamic.internal.model.entity.User.name"
		},
	}

	actual := loader.LoadSourcePluginBundles(context.Background(), "zh-CN")
	expected := map[string]string{
		"plugins.plugin_demo_dynamic.name": "Allowed Plugin Name",
	}
	if !reflect.DeepEqual(actual["plugin-demo-dynamic"], expected) {
		t.Fatalf("unexpected restricted plugin bundle: expected=%v actual=%v", expected, actual["plugin-demo-dynamic"])
	}
}

// TestLoadDynamicPluginBundlesConsumesExtractedAssets verifies dynamic plugin
// release assets can be merged after WASM extraction has happened elsewhere.
func TestLoadDynamicPluginBundlesConsumesExtractedAssets(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		PluginScope: PluginScopeRestrictedToPluginNamespace,
		ValueMode:   ValueModeStringOnly,
	}
	releases := []ReleaseRef{
		{
			PluginID: "plugin-demo-dynamic",
			Assets: []LocaleAsset{
				{
					Locale: "zh-CN",
					Content: `{
  "plugins": {
    "plugin_demo_dynamic": {
      "name": "动态插件"
    }
  },
  "core.title": "Rejected Host Title"
}`,
				},
				{
					Locale:  "en-US",
					Content: `{"plugins.plugin_demo_dynamic.name":"Dynamic Plugin"}`,
				},
			},
		},
	}

	actual := loader.LoadDynamicPluginBundles(context.Background(), "zh-CN", releases)
	expected := map[string]string{
		"plugins.plugin_demo_dynamic.name": "动态插件",
	}
	if !reflect.DeepEqual(actual["plugin-demo-dynamic"], expected) {
		t.Fatalf("unexpected dynamic plugin bundle: expected=%v actual=%v", expected, actual["plugin-demo-dynamic"])
	}
}
