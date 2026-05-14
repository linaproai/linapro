// This file validates dynamic plugin manifests.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return false, fmt.Errorf("read plugin manifest %s: %w", manifestPath, err)
	}
	var manifest pluginManifest
	if err = yaml.Unmarshal(content, &manifest); err != nil {
		return false, fmt.Errorf("parse plugin manifest %s: %w", manifestPath, err)
	}
	return strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic"), nil
}
