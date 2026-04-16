// This file loads the dynamic plugin manifest and validates manifest-level
// metadata shared by the standalone wasm builder flow.

package builder

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginbridge"
)

func validateRuntimeBuildManifest(manifest *pluginManifest, manifestPath string) error {
	if manifest == nil {
		return fmt.Errorf("dynamic plugin manifest cannot be nil")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return fmt.Errorf("dynamic plugin manifest missing id: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("dynamic plugin manifest missing name: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("dynamic plugin manifest missing version: %s", manifestPath)
	}
	manifest.Type = strings.ToLower(strings.TrimSpace(manifest.Type))
	if manifest.Type != pluginTypeDynamic {
		return fmt.Errorf("dynamic sample manifest type must be dynamic: %s", manifestPath)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return fmt.Errorf("dynamic plugin id must use kebab-case: %s", manifest.ID)
	}
	if err := validateSemanticVersion(manifest.Version); err != nil {
		return fmt.Errorf("dynamic plugin version is invalid: %w", err)
	}
	manifest.Capabilities = pluginbridge.NormalizeCapabilities(manifest.Capabilities)
	if len(manifest.Capabilities) > 0 {
		return fmt.Errorf(
			"dynamic plugin manifest no longer supports top-level capabilities; please keep only hostServices declarations (found: %s)",
			strings.Join(manifest.Capabilities, ", "),
		)
	}
	if err := pluginbridge.ValidateHostServiceSpecs(manifest.HostServices); err != nil {
		return fmt.Errorf("dynamic plugin hostServices invalid: %w", err)
	}
	manifest.HostServices = pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices)
	return nil
}

func loadYAMLFile(filePath string, target interface{}) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return fmt.Errorf("yaml file is empty: %s", filePath)
	}
	if err = yaml.Unmarshal(content, target); err != nil {
		return fmt.Errorf("failed to parse yaml file %s: %w", filePath, err)
	}
	return nil
}

func validateSemanticVersion(value string) error {
	match := pluginManifestSemverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) < 4 {
		return fmt.Errorf("version must use semver format: %s", value)
	}

	for _, raw := range match[1:4] {
		if _, err := strconv.Atoi(raw); err != nil {
			return err
		}
	}
	return nil
}
