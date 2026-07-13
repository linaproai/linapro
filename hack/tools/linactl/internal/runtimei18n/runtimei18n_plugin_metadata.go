// This file validates plugin management display metadata i18n keys for
// plugins with i18n.enabled=true. Runtime localization looks up
// plugin.<id>.name and plugin.<id>.description; bare name/description keys
// are not used by the plugin management list projection.

package runtimei18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// validatePluginDisplayMetadataKeys ensures every i18n-enabled plugin ships
// the management-list display keys expected by host localization:
// plugin.<plugin-id>.name and plugin.<plugin-id>.description for each locale
// that contributes runtime i18n resources.
func validatePluginDisplayMetadataKeys(repoRoot string) ([]string, error) {
	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plugin root %s: %w", pluginRoot, err)
	}

	errors := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		pluginDir := filepath.Join(pluginRoot, dirName)
		pluginYAML := filepath.Join(pluginDir, "plugin.yaml")
		if _, statErr := os.Stat(pluginYAML); statErr != nil {
			continue
		}
		enabled, pluginID, enabledErr := pluginI18nIdentity(pluginYAML, dirName)
		if enabledErr != nil {
			return nil, fmt.Errorf("plugin %s: %w", dirName, enabledErr)
		}
		if !enabled {
			continue
		}

		scope := "plugin:" + pluginID
		i18nDir := filepath.Join(pluginDir, "manifest", "i18n")
		localeMaps, loadErr := loadRuntimeLocaleMaps(i18nDir)
		if loadErr != nil {
			if os.IsNotExist(loadErr) {
				errors = append(errors, missingPluginDisplayMetadataKeys(scope, pluginID, []string{"(missing i18n dir)"})...)
				continue
			}
			return nil, loadErr
		}
		if len(localeMaps) == 0 {
			errors = append(errors, missingPluginDisplayMetadataKeys(scope, pluginID, []string{"(no locales)"})...)
			continue
		}

		nameKey := "plugin." + pluginID + ".name"
		descriptionKey := "plugin." + pluginID + ".description"
		for _, locale := range sortedMapKeys(localeMaps) {
			messages := localeMaps[locale]
			if _, ok := messages[nameKey]; !ok {
				errors = append(errors, fmt.Sprintf(
					"%s: %s missing plugin display key %s (management list localizes via plugin.<id>.name)",
					scope, locale, nameKey,
				))
			}
			if _, ok := messages[descriptionKey]; !ok {
				errors = append(errors, fmt.Sprintf(
					"%s: %s missing plugin display key %s (management list localizes via plugin.<id>.description)",
					scope, locale, descriptionKey,
				))
			}
			// Surface the common mistake of writing bare name/description at
			// the JSON root without the plugin.<id> namespace.
			if hasBarePluginMetadataKeys(messages) {
				if _, ok := messages[nameKey]; !ok {
					errors = append(errors, fmt.Sprintf(
						"%s: %s has bare name/description keys; nest them under plugin.%s in plugin.json",
						scope, locale, pluginID,
					))
				}
			}
		}
	}
	return errors, nil
}

// missingPluginDisplayMetadataKeys reports both required display keys for the
// given pseudo-locale labels when no real locale catalogs exist.
func missingPluginDisplayMetadataKeys(scope string, pluginID string, locales []string) []string {
	nameKey := "plugin." + pluginID + ".name"
	descriptionKey := "plugin." + pluginID + ".description"
	errors := make([]string, 0, len(locales)*2)
	for _, locale := range locales {
		errors = append(errors,
			fmt.Sprintf("%s: %s missing plugin display key %s", scope, locale, nameKey),
			fmt.Sprintf("%s: %s missing plugin display key %s", scope, locale, descriptionKey),
		)
	}
	return errors
}

// hasBarePluginMetadataKeys reports whether a flattened locale catalog contains
// root-level name/description keys that cannot be used by management-list
// localization (which requires the plugin.<id>.* namespace).
func hasBarePluginMetadataKeys(messages map[string]string) bool {
	if messages == nil {
		return false
	}
	_, hasName := messages["name"]
	_, hasDescription := messages["description"]
	return hasName || hasDescription
}

// pluginI18nIdentity returns whether i18n is enabled and the plugin id from
// plugin.yaml. When the yaml id is empty, directoryName is used as fallback.
func pluginI18nIdentity(pluginYAMLPath string, directoryName string) (bool, string, error) {
	content, err := os.ReadFile(pluginYAMLPath)
	if err != nil {
		return false, "", fmt.Errorf("read plugin.yaml: %w", err)
	}
	var payload pluginIdentityYAML
	if err = yaml.Unmarshal(content, &payload); err != nil {
		return false, "", fmt.Errorf("parse plugin.yaml: %w", err)
	}
	pluginID := strings.TrimSpace(payload.ID)
	if pluginID == "" {
		pluginID = strings.TrimSpace(directoryName)
	}
	enabled := payload.I18n != nil && payload.I18n.Enabled
	return enabled, pluginID, nil
}

// pluginIdentityYAML is the minimal plugin.yaml shape for i18n gating plus id.
type pluginIdentityYAML struct {
	ID   string `yaml:"id"`
	I18n *struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"i18n"`
}
