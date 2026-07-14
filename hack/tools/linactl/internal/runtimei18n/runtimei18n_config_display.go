// This file validates config-management display metadata keys
// (config.<sys_config.key>.name / .remark) for the host and for plugins that
// declare hostconfigcap.SysConfigKey constants with i18n.enabled=true.
//
// Runtime list projection uses these keys in sysconfig.localizeConfigEntity;
// locale parity alone cannot catch a key that is missing from every locale.

package runtimei18n

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// sysConfigKeyAssignmentPattern matches typed SysConfigKey string constants.
	// Example: ConfigKeyRegion hostconfigcap.SysConfigKey = "plugin.x.region"
	sysConfigKeyAssignmentPattern = regexp.MustCompile(
		`(?m)SysConfigKey\s*=\s*"((?:sys|demo|plugin)\.[^"]+)"`,
	)
	// hostProtectedConfigKeyConstPattern matches host-owned protected key constants.
	hostProtectedConfigKeyConstPattern = regexp.MustCompile(
		`(?m)(?:RuntimeParamKey|PublicFrontendSettingKey)[A-Za-z0-9_]*\s*=\s*"((?:sys|demo)\.[^"]+)"`,
	)
	// sysConfigSQLKeyPattern extracts config keys from sys_config seed SQL rows.
	sysConfigSQLKeyPattern = regexp.MustCompile(
		`'((?:sys|demo)\.[a-zA-Z0-9_.]+)'`,
	)
	// insertSysConfigPattern detects SQL that seeds sys_config rows.
	insertSysConfigPattern = regexp.MustCompile(`(?i)INSERT\s+INTO\s+sys_config\b`)
)

// validateConfigDisplayMetadataKeys ensures host and i18n-enabled plugin config
// keys have config.<key>.name and config.<key>.remark in every runtime locale.
func validateConfigDisplayMetadataKeys(repoRoot string) ([]string, error) {
	errors := make([]string, 0)

	hostKeys, err := collectHostConfigDisplayKeys(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(hostKeys) > 0 {
		hostErrors, hostErr := validateConfigKeysAgainstI18N(
			"host:core",
			filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n"),
			hostKeys,
		)
		if hostErr != nil {
			return nil, hostErr
		}
		errors = append(errors, hostErrors...)
	}

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return errors, nil
		}
		return nil, fmt.Errorf("read plugin root %s: %w", pluginRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginID := entry.Name()
		pluginDir := filepath.Join(pluginRoot, pluginID)
		pluginYAML := filepath.Join(pluginDir, "plugin.yaml")
		if _, statErr := os.Stat(pluginYAML); statErr != nil {
			continue
		}
		enabled, enabledErr := pluginI18nEnabled(pluginYAML)
		if enabledErr != nil {
			return nil, fmt.Errorf("plugin %s: %w", pluginID, enabledErr)
		}
		if !enabled {
			continue
		}
		keys, collectErr := collectModuleSysConfigKeys(pluginDir)
		if collectErr != nil {
			return nil, collectErr
		}
		if len(keys) == 0 {
			continue
		}
		scope := "plugin:" + pluginID
		pluginErrors, pluginErr := validateConfigKeysAgainstI18N(
			scope,
			filepath.Join(pluginDir, "manifest", "i18n"),
			keys,
		)
		if pluginErr != nil {
			return nil, pluginErr
		}
		errors = append(errors, pluginErrors...)
	}
	return errors, nil
}

// validateConfigKeysAgainstI18N checks that every config key has name+remark
// translation entries in each runtime locale under i18nDir.
func validateConfigKeysAgainstI18N(scope string, i18nDir string, configKeys map[string]struct{}) ([]string, error) {
	localeMaps, err := loadRuntimeLocaleMaps(i18nDir)
	if err != nil {
		if os.IsNotExist(err) {
			return missingConfigDisplayKeysForLocales(scope, configKeys, []string{"(missing i18n dir)"}), nil
		}
		return nil, err
	}
	if len(localeMaps) == 0 {
		return missingConfigDisplayKeysForLocales(scope, configKeys, []string{"(no locales)"}), nil
	}

	errors := make([]string, 0)
	for _, locale := range sortedMapKeys(localeMaps) {
		messages := localeMaps[locale]
		for _, configKey := range sortedMapKeys(configKeys) {
			for _, field := range []string{"name", "remark"} {
				messageKey := "config." + configKey + "." + field
				if _, ok := messages[messageKey]; !ok {
					errors = append(errors, fmt.Sprintf(
						"%s: locale %s missing config display key for %s: %s",
						scope,
						locale,
						configKey,
						messageKey,
					))
				}
			}
		}
	}
	return errors, nil
}

func missingConfigDisplayKeysForLocales(scope string, configKeys map[string]struct{}, locales []string) []string {
	errors := make([]string, 0, len(configKeys)*2*len(locales))
	for _, locale := range locales {
		for _, configKey := range sortedMapKeys(configKeys) {
			for _, field := range []string{"name", "remark"} {
				errors = append(errors, fmt.Sprintf(
					"%s: locale %s missing config display key for %s: config.%s.%s",
					scope,
					locale,
					configKey,
					configKey,
					field,
				))
			}
		}
	}
	return errors
}

// collectHostConfigDisplayKeys unions host SQL seed keys and host Go constants.
func collectHostConfigDisplayKeys(repoRoot string) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	sqlRoot := filepath.Join(repoRoot, "apps", "lina-core", "manifest", "sql")
	if _, err := os.Stat(sqlRoot); err == nil {
		if err := filepath.WalkDir(sqlRoot, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			text := string(content)
			if !insertSysConfigPattern.MatchString(text) {
				return nil
			}
			for _, match := range sysConfigSQLKeyPattern.FindAllStringSubmatch(text, -1) {
				key := strings.TrimSpace(match[1])
				if key != "" {
					result[key] = struct{}{}
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("walk host sys_config SQL under %s: %w", sqlRoot, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat host sys_config SQL root %s: %w", sqlRoot, err)
	}

	hostRoot := filepath.Join(repoRoot, "apps", "lina-core")
	if _, err := os.Stat(hostRoot); err == nil {
		goKeys, err := collectHostGoConfigKeys(hostRoot)
		if err != nil {
			return nil, err
		}
		for key := range goKeys {
			result[key] = struct{}{}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat host root %s: %w", hostRoot, err)
	}
	return result, nil
}

// collectHostGoConfigKeys scans host production Go for protected config key constants.
func collectHostGoConfigKeys(moduleRoot string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", ".git", "hack", "manifest", "frontend", "dist", "temp":
				if path != moduleRoot {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(content)
		for _, match := range sysConfigKeyAssignmentPattern.FindAllStringSubmatch(text, -1) {
			key := strings.TrimSpace(match[1])
			if key != "" {
				result[key] = struct{}{}
			}
		}
		for _, match := range hostProtectedConfigKeyConstPattern.FindAllStringSubmatch(text, -1) {
			key := strings.TrimSpace(match[1])
			if key != "" {
				result[key] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk host config key constants under %s: %w", moduleRoot, err)
	}
	return result, nil
}

// collectModuleSysConfigKeys collects SysConfigKey string constants under one module.
func collectModuleSysConfigKeys(moduleRoot string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", ".git", "hack", "manifest", "frontend", "dist", "temp":
				if path != moduleRoot {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, match := range sysConfigKeyAssignmentPattern.FindAllStringSubmatch(string(content), -1) {
			key := strings.TrimSpace(match[1])
			if key != "" {
				result[key] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk SysConfigKey constants under %s: %w", moduleRoot, err)
	}
	return result, nil
}
