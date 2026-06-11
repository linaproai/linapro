// This file verifies package-level pluginhost dependency boundaries.

package pluginhost

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestPluginHostDoesNotImportPluginBridge verifies source-plugin host contracts
// stay independent from dynamic-plugin bridge runtime packages.
func TestPluginHostDoesNotImportPluginBridge(t *testing.T) {
	t.Parallel()

	violations := collectPluginhostImportViolations(t, ".", func(importPath string) bool {
		return importPath == "lina-core/pkg/plugin/pluginbridge" ||
			strings.HasPrefix(importPath, "lina-core/pkg/plugin/pluginbridge/")
	})
	for _, violation := range violations {
		t.Errorf("pluginhost dependency direction violation: %s", violation)
	}
}

// collectPluginhostImportViolations walks production Go files under scanRoot
// and reports imports rejected by match.
func collectPluginhostImportViolations(t *testing.T, scanRoot string, match func(importPath string) bool) []string {
	t.Helper()

	var violations []string
	walkErr := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		parsed, parseErr := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, importSpec := range parsed.Imports {
			importPath, unquoteErr := strconv.Unquote(importSpec.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			if match(importPath) {
				violations = append(violations, path+" imports "+importPath)
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("scan %s for forbidden imports: %v", scanRoot, walkErr)
	}
	return violations
}
