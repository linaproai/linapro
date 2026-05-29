// This file verifies active-release manifest resource projection for dynamic
// plugin host services.

package runtime

import (
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
)

func TestBuildArtifactResourceViews(t *testing.T) {
	manifest := &catalog.Manifest{
		RuntimeArtifact: &catalog.ArtifactSpec{
			ManifestResources: []*catalog.ArtifactManifestResource{
				{
					Path:    "manifest/config/config.example.yaml",
					Content: []byte("template: true\n"),
				},
				{
					Path:    "manifest/config/config.yaml",
					Content: []byte("runtime: true\n"),
				},
				{
					Path:    "manifest/metadata.yaml",
					Content: []byte("name: demo\n"),
				},
				{
					Path:    "manifest/sql/001-schema.sql",
					Content: []byte("CREATE TABLE plugin_demo(id bigint);\n"),
				},
				{
					Path:    "manifest/i18n/zh-CN/plugin.json",
					Content: []byte(`{"plugin.demo":"demo"}`),
				},
				{
					Path:    "manifest/resources/policy.yaml",
					Content: []byte("enabled: true\n"),
				},
			},
		},
	}

	defaultConfig := buildArtifactDefaultConfig(manifest)
	if string(defaultConfig) != "runtime: true\n" {
		t.Fatalf("expected runtime config only, got %q", string(defaultConfig))
	}

	resources := buildArtifactManifestResources(manifest)
	if len(resources) != 6 {
		t.Fatalf("expected all manifest resources, got %#v", resources)
	}
	if string(resources["metadata.yaml"]) != "name: demo\n" {
		t.Fatalf("expected metadata resource, got %#v", resources)
	}
	if string(resources["config/config.example.yaml"]) != "template: true\n" {
		t.Fatalf("expected config template resource, got %#v", resources)
	}
	if string(resources["config/config.yaml"]) != "runtime: true\n" {
		t.Fatalf("expected config resource, got %#v", resources)
	}
	if string(resources["sql/001-schema.sql"]) != "CREATE TABLE plugin_demo(id bigint);\n" {
		t.Fatalf("expected SQL resource, got %#v", resources)
	}
	if string(resources["i18n/zh-CN/plugin.json"]) != `{"plugin.demo":"demo"}` {
		t.Fatalf("expected i18n resource, got %#v", resources)
	}
	if string(resources["resources/policy.yaml"]) != "enabled: true\n" {
		t.Fatalf("expected policy resource, got %#v", resources)
	}
}
