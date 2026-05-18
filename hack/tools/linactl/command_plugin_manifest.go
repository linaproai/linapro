// This file validates dynamic plugin manifests.

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateDynamicPlugin verifies that a plugin exists and is dynamic.
func validateDynamicPlugin(pluginRoot string, plugin string) error {
	manifest := filepath.Join(pluginRoot, plugin, "plugin.yaml")
	if !fileExists(manifest) {
		return fmt.Errorf("plugin does not exist: %s", plugin)
	}
	dynamic, err := isDynamicPlugin(manifest)
	if err != nil {
		return err
	}
	if !dynamic {
		return fmt.Errorf("plugin is not dynamic and cannot be built as wasm: %s", plugin)
	}
	return nil
}

// isDynamicPlugin reports whether a plugin manifest declares dynamic type.
func isDynamicPlugin(manifestPath string) (bool, error) {
	manifest, err := readPluginManifest(manifestPath)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic"), nil
}
