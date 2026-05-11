// This file audits first-party source plugin protected route middleware chains.

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSourcePluginProtectedRoutesIncludeTenancy verifies source plugin APIs use
// the same Auth -> Tenancy -> Permission chain as host protected static APIs.
func TestSourcePluginProtectedRoutesIncludeTenancy(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	pluginFiles, err := filepath.Glob(filepath.Join(repoRoot, "apps", "lina-plugins", "*", "backend", "plugin.go"))
	if err != nil {
		t.Fatalf("scan source plugin route files failed: %v", err)
	}
	if len(pluginFiles) == 0 {
		t.Fatal("expected source plugin route files")
	}

	for _, file := range pluginFiles {
		t.Run(filepath.Base(filepath.Dir(filepath.Dir(file))), func(t *testing.T) {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatalf("read source plugin route file failed: %v", readErr)
			}
			text := string(content)
			if !strings.Contains(text, "middlewares.Auth()") {
				return
			}
			authIndex := strings.Index(text, "middlewares.Auth()")
			tenancyIndex := strings.Index(text, "middlewares.Tenancy()")
			permissionIndex := strings.Index(text, "middlewares.Permission()")
			if tenancyIndex < 0 {
				t.Fatalf("protected source plugin route must include Tenancy middleware: %s", file)
			}
			if !(authIndex < tenancyIndex && tenancyIndex < permissionIndex) {
				t.Fatalf("protected source plugin route must use Auth -> Tenancy -> Permission order: %s", file)
			}
		})
	}
}
