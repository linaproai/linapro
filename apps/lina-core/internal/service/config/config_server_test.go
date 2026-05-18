// This file verifies LinaPro-specific server extension configuration loading.

package config

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// TestGetServerExtensionsUsesConfiguredApiDocPath verifies the hosted OpenAPI
// route override is loaded from the server extensions section.
func TestGetServerExtensionsUsesConfiguredApiDocPath(t *testing.T) {
	setTestServerConfigAdapter(t, `
server:
  extensions:
    apiDocPath: "/custom-api.json"
`)

	cfg := New().GetServerExtensions(context.Background())

	if cfg.ApiDocPath != "/custom-api.json" {
		t.Fatalf("expected hosted api doc path to be loaded, got %q", cfg.ApiDocPath)
	}
}

// TestGetServerExtensionsUsesDefaultWhenPathMissing verifies the hosted OpenAPI
// route falls back to the project default when config omits it.
func TestGetServerExtensionsUsesDefaultWhenPathMissing(t *testing.T) {
	setTestServerConfigAdapter(t, `
server:
  address: ":8080"
`)

	cfg := New().GetServerExtensions(context.Background())

	if cfg.ApiDocPath != defaultServerApiDocPath {
		t.Fatalf("expected default hosted api doc path %q, got %q", defaultServerApiDocPath, cfg.ApiDocPath)
	}
}

// setTestServerConfigAdapter swaps the process config adapter for one test case
// and restores the original adapter afterward.
func setTestServerConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)
	resetStaticConfigCaches()

	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
		resetStaticConfigCaches()
	})
}
