// This file verifies embedded metadata parsing for OpenAPI and component cards.

package config

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gcfg"
)

// TestGetMetadataMergesOpenApiAndComponentSections verifies metadata parsing can
// load OpenAPI and component sections from one embedded payload.
func TestGetMetadataMergesOpenApiAndComponentSections(t *testing.T) {
	adapter, err := gcfg.NewAdapterContent(`
framework:
  name: "LinaPro"
  version: "v9.9.9"
  description: "Framework description"
  homepage: "https://linapro.ai"
  repositoryUrl: "https://github.com/example/linapro"
  license: "MIT"
openapi:
  title: "Embedded API"
  description: "Embedded description"
  version: "v9.9.9"
  serverUrl: "https://api.example.com"
  serverDescription: "ExampleEndpoint"
backend:
  - name: "GoFrame"
    version: "auto"
    url: "https://goframe.org"
    description: "GF"
frontend:
  - name: "Vue"
    version: "3.x"
    url: "https://vuejs.org"
    description: "Vue runtime"
`)
	if err != nil {
		t.Fatalf("create metadata adapter: %v", err)
	}

	cfg := &MetadataConfig{
		OpenApi: defaultOpenApiConfig(),
	}
	mustScanMetadataConfig(context.Background(), adapter, "framework", &cfg.Framework)
	mustScanMetadataConfig(context.Background(), adapter, "openapi", &cfg.OpenApi)
	mustScanMetadataConfig(context.Background(), adapter, "backend", &cfg.Backend)
	mustScanMetadataConfig(context.Background(), adapter, "frontend", &cfg.Frontend)

	if cfg.Framework.Name != "LinaPro" {
		t.Fatalf("expected framework name LinaPro, got %q", cfg.Framework.Name)
	}
	if cfg.Framework.Version != "v9.9.9" {
		t.Fatalf("expected framework version v9.9.9, got %q", cfg.Framework.Version)
	}
	if cfg.Framework.Homepage != "https://linapro.ai" {
		t.Fatalf("expected framework homepage https://linapro.ai, got %q", cfg.Framework.Homepage)
	}
	if cfg.Framework.RepositoryURL != "https://github.com/example/linapro" {
		t.Fatalf("expected framework repository url https://github.com/example/linapro, got %q", cfg.Framework.RepositoryURL)
	}
	if cfg.OpenApi.Title != "Embedded API" {
		t.Fatalf("expected embedded title, got %q", cfg.OpenApi.Title)
	}
	if cfg.OpenApi.ServerUrl != "https://api.example.com" {
		t.Fatalf("expected embedded server url, got %q", cfg.OpenApi.ServerUrl)
	}
	if len(cfg.Backend) != 1 || cfg.Backend[0].Name != "GoFrame" {
		t.Fatalf("expected one backend component, got %#v", cfg.Backend)
	}
	if len(cfg.Frontend) != 1 || cfg.Frontend[0].Name != "Vue" {
		t.Fatalf("expected one frontend component, got %#v", cfg.Frontend)
	}
}

// TestGetOpenApiUsesEmbeddedMetadataAsset verifies the embedded metadata asset
// provides the public OpenAPI document metadata.
func TestGetOpenApiUsesEmbeddedMetadataAsset(t *testing.T) {
	cfg := New().GetOpenApi(context.Background())

	if cfg.Title != "LinaPro Framework API" {
		t.Fatalf("expected embedded metadata title, got %q", cfg.Title)
	}
	if cfg.Version != "v0.5.0" {
		t.Fatalf("expected embedded metadata version v0.5.0, got %q", cfg.Version)
	}
	if cfg.ServerUrl != "http://localhost:8080" {
		t.Fatalf("expected embedded metadata server url, got %q", cfg.ServerUrl)
	}
}
