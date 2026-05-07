// This file verifies the delivery database debug defaults used for startup logs.

package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

// TestDatabaseDebugDefaultsOffInDeliveryConfig verifies startup logs do not
// emit ORM SQL details unless operators explicitly enable database debug.
func TestDatabaseDebugDefaultsOffInDeliveryConfig(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "manifest", "config", "config.template.yaml")
	assertDatabaseDebugDisabled(t, templatePath)
	packedTemplatePath := filepath.Join("..", "..", "..", "internal", "packed", "manifest", "config", "config.template.yaml")
	assertDatabaseDebugDisabledIfExists(t, packedTemplatePath)

	// Local config.yaml is intentionally git-ignored, but when present it should
	// follow the same default to keep developer startup logs quiet.
	localPath := filepath.Join("..", "..", "..", "manifest", "config", "config.yaml")
	assertDatabaseDebugDisabledIfExists(t, localPath)
}

// assertDatabaseDebugDisabled verifies one config file disables ORM SQL debug
// logging by default.
func assertDatabaseDebugDisabled(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read delivery config %s: %v", path, err)
	}
	if !strings.Contains(string(content), "debug: false") {
		t.Fatalf("expected %s to disable database debug by default", path)
	}
}

// assertDatabaseDebugDisabledIfExists verifies ignored local or packed config
// files only when they are present in the current workspace.
func assertDatabaseDebugDisabledIfExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		assertDatabaseDebugDisabled(t, path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat optional config %s: %v", path, err)
	}
}

// TestDatabaseDebugCanBeEnabledExplicitly verifies SQL diagnostics can still be
// enabled through an explicit config override.
func TestDatabaseDebugCanBeEnabledExplicitly(t *testing.T) {
	setTestServerConfigAdapter(t, `
database:
  default:
    debug: true
`)

	value, err := g.Cfg().Get(context.Background(), "database.default.debug")
	if err != nil {
		t.Fatalf("read explicit database debug config: %v", err)
	}
	if !value.Bool() {
		t.Fatal("expected explicit database.default.debug=true to be readable")
	}
}
