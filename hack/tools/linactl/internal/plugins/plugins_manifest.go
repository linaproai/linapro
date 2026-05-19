// This file validates plugin manifests for dynamic Wasm builds and configured
// source-plugin management. It keeps manifest parsing reusable across official
// workspace generation, status checks, and command-specific validation.

package plugins

import (
	"fmt"
	"path/filepath"
	"strings"

	"linactl/internal/fileutil"
)

// ValidateDynamic verifies that a plugin exists and is dynamic.
func ValidateDynamic(pluginRoot string, plugin string) error {
	manifest := filepath.Join(pluginRoot, plugin, "plugin.yaml")
	if !fileutil.FileExists(manifest) {
		return fmt.Errorf("plugin does not exist: %s", plugin)
	}
	dynamic, err := IsDynamic(manifest)
	if err != nil {
		return err
	}
	if !dynamic {
		return fmt.Errorf("plugin is not dynamic and cannot be built as wasm: %s", plugin)
	}
	return nil
}

// IsDynamic reports whether a plugin manifest declares dynamic type.
func IsDynamic(manifestPath string) (bool, error) {
	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic"), nil
}
