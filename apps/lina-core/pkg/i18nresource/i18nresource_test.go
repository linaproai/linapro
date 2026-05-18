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

// TestLoadHostBundleDefaultsToLocaleDirectory verifies the loader's default
// target directory follows the current locale root convention.
func TestLoadHostBundleDefaultsToLocaleDirectory(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		HostFS: fstest.MapFS{
			"manifest/i18n/en-US/framework.json": &fstest.MapFile{Data: []byte(`{
  "framework": {
    "name": "LinaPro"
  }
}`)},
			"manifest/i18n/en-US/menu.json": &fstest.MapFile{Data: []byte(`{
  "menu.dashboard.title": "Dashboard"
}`)},
			"manifest/i18n/en-US/apidoc/common.json": &fstest.MapFile{Data: []byte(`{
  "core.common.pageNum.dc": "Page number"
}`)},
		},
		Subdir:    "manifest/i18n",
		ValueMode: ValueModeStringOnly,
	}

	expected := map[string]string{
		"framework.name":       "LinaPro",
		"menu.dashboard.title": "Dashboard",
	}
	if actual := loader.LoadHostBundle(context.Background(), "en-US"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected host bundle: expected=%v actual=%v", expected, actual)
	}
}

// TestLoadHostBundleMergesLocaleDirectoryOnly verifies runtime-style resources
// merge direct files under one locale while ignoring nested apidoc files.
func TestLoadHostBundleMergesLocaleDirectoryOnly(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		HostFS: fstest.MapFS{
			"manifest/i18n/zh-CN/menu.json": &fstest.MapFile{Data: []byte(`{
  "menu": {
    "dashboard": {
      "title": "工作台"
    }
  }
}`)},
			"manifest/i18n/zh-CN/error.json": &fstest.MapFile{Data: []byte(`{
  "error.auth.login": "登录失败"
}`)},
			"manifest/i18n/zh-CN/apidoc/core-api-auth.json": &fstest.MapFile{Data: []byte(`{
  "core.api.auth.v1.LoginReq.meta.summary": "用户登录"
}`)},
		},
		Subdir:    "manifest/i18n",
		ValueMode: ValueModeStringifyScalars,
	}

	expected := map[string]string{
		"menu.dashboard.title": "工作台",
		"error.auth.login":     "登录失败",
	}
	if actual := loader.LoadHostBundle(context.Background(), "zh-CN"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected locale-directory bundle: expected=%v actual=%v", expected, actual)
	}
}

// TestLoadHostBundleMergesRecursiveLocaleSubdirectory verifies apidoc resources
// can live under the locale directory without being mixed into runtime bundles.
func TestLoadHostBundleMergesRecursiveLocaleSubdirectory(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		HostFS: fstest.MapFS{
			"manifest/i18n/zh-CN/menu.json": &fstest.MapFile{Data: []byte(`{
  "menu": {
    "dashboard": {
      "title": "工作台"
    }
  }
}`)},
			"manifest/i18n/zh-CN/apidoc/common.json": &fstest.MapFile{Data: []byte(`{
  "core.common.pageNum.dc": "页码"
}`)},
			"manifest/i18n/zh-CN/apidoc/core-api-auth.json": &fstest.MapFile{Data: []byte(`{
  "core.api.auth.v1.LoginReq.meta.summary": "用户登录"
}`)},
		},
		Subdir:       "manifest/i18n",
		LocaleSubdir: "apidoc",
		Recursive:    true,
		ValueMode:    ValueModeStringOnly,
	}

	expected := map[string]string{
		"core.common.pageNum.dc":                 "页码",
		"core.api.auth.v1.LoginReq.meta.summary": "用户登录",
	}
	if actual := loader.LoadHostBundle(context.Background(), "zh-CN"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected locale-subdirectory bundle: expected=%v actual=%v", expected, actual)
	}
}

// TestLoadHostBundleRecursiveRootIncludesNestedResources verifies recursive
// root-directory scans are explicit and include nested locale resources.
func TestLoadHostBundleRecursiveRootIncludesNestedResources(t *testing.T) {
	t.Parallel()

	loader := ResourceLoader{
		HostFS: fstest.MapFS{
			"manifest/i18n/zh-CN/menu.json": &fstest.MapFile{Data: []byte(`{
  "menu.dashboard.title": "工作台"
}`)},
			"manifest/i18n/zh-CN/apidoc/common.json": &fstest.MapFile{Data: []byte(`{
  "core.common.pageNum.dc": "页码"
}`)},
		},
		Subdir:    "manifest/i18n",
		Recursive: true,
		ValueMode: ValueModeStringOnly,
	}

	expected := map[string]string{
		"menu.dashboard.title":   "工作台",
		"core.common.pageNum.dc": "页码",
	}
	if actual := loader.LoadHostBundle(context.Background(), "zh-CN"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected recursive root bundle: expected=%v actual=%v", expected, actual)
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
					"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(`{"plugin.z.name":"Z Plugin"}`)},
				}},
				fakeSourcePlugin{id: "a-plugin", filesystem: fstest.MapFS{
					"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(`{"plugin.a.name":"A Plugin"}`)},
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
					"manifest/i18n/zh-CN/apidoc/plugin-api-main.json": &fstest.MapFile{Data: []byte(`{
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
		Subdir:       "manifest/i18n",
		LocaleSubdir: "apidoc",
		PluginScope:  PluginScopeRestrictedToPluginNamespace,
		Recursive:    true,
		ValueMode:    ValueModeStringOnly,
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
