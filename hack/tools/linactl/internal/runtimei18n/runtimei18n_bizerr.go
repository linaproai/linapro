// This file validates that host and i18n-enabled plugin bizerr.Code messageKeys
// are covered by the corresponding runtime i18n catalogs (error.json and peers).
// Plugins with i18n.enabled != true are skipped per project i18n governance.
package runtimei18n

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	bizerrDefineCallPattern = regexp.MustCompile(
		`(?s)bizerr\.MustDefine\((.*?)\)`,
	)
	bizerrStringLiteralPattern = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

// validateBizerrMessageKeys ensures every host and i18n-enabled plugin bizerr
// messageKey exists in that scope's runtime i18n resources for each locale.
func validateBizerrMessageKeys(repoRoot string) ([]string, error) {
	errors := make([]string, 0)

	hostRoot := filepath.Join(repoRoot, "apps", "lina-core")
	if info, err := os.Stat(hostRoot); err == nil && info.IsDir() {
		hostErrors, hostErr := validateModuleBizerrMessageKeys("host:core", hostRoot, filepath.Join(hostRoot, "manifest", "i18n"))
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
		scope := "plugin:" + pluginID
		pluginErrors, pluginErr := validateModuleBizerrMessageKeys(
			scope,
			pluginDir,
			filepath.Join(pluginDir, "manifest", "i18n"),
		)
		if pluginErr != nil {
			return nil, pluginErr
		}
		errors = append(errors, pluginErrors...)
	}
	return errors, nil
}

// validateModuleBizerrMessageKeys checks one module root's bizerr definitions
// against its runtime i18n locale catalogs. Modules with no bizerr definitions
// are skipped (empty is not a failure).
func validateModuleBizerrMessageKeys(scope string, moduleRoot string, i18nDir string) ([]string, error) {
	defs, err := collectBizerrMessageKeys(moduleRoot)
	if err != nil {
		return nil, err
	}
	if len(defs) == 0 {
		return nil, nil
	}

	localeMaps, err := loadRuntimeLocaleMaps(i18nDir)
	if err != nil {
		if os.IsNotExist(err) {
			return missingBizerrKeysForLocales(scope, defs, []string{"(missing i18n dir)"}), nil
		}
		return nil, err
	}
	if len(localeMaps) == 0 {
		return missingBizerrKeysForLocales(scope, defs, []string{"(no locales)"}), nil
	}

	errors := make([]string, 0)
	errorCodes := sortedMapKeys(defs)
	for _, locale := range sortedMapKeys(localeMaps) {
		messages := localeMaps[locale]
		for _, errorCode := range errorCodes {
			messageKey := defs[errorCode]
			if messageKey == "" {
				errors = append(errors, fmt.Sprintf(
					"%s: bizerr %s has empty messageKey",
					scope,
					errorCode,
				))
				continue
			}
			if _, ok := messages[messageKey]; !ok {
				errors = append(errors, fmt.Sprintf(
					"%s: locale %s missing bizerr messageKey for %s: %s",
					scope,
					locale,
					errorCode,
					messageKey,
				))
			}
		}
	}
	return errors, nil
}

func missingBizerrKeysForLocales(scope string, defs map[string]string, locales []string) []string {
	errors := make([]string, 0, len(defs)*len(locales))
	for _, locale := range locales {
		for _, errorCode := range sortedMapKeys(defs) {
			errors = append(errors, fmt.Sprintf(
				"%s: locale %s missing bizerr messageKey for %s: %s",
				scope,
				locale,
				errorCode,
				defs[errorCode],
			))
		}
	}
	return errors
}

// collectBizerrMessageKeys walks a module tree for bizerr.MustDefine calls and
// returns errorCode -> derived messageKey. Test files and common vendor trees
// are skipped.
func collectBizerrMessageKeys(moduleRoot string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", ".git", "hack", "manifest", "frontend", "dist", "temp":
				// Skip non-production backend trees. Plugin frontend and
				// packaging dirs never define bizerr codes.
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
		if !strings.Contains(text, "bizerr.MustDefine") {
			return nil
		}
		for _, match := range bizerrDefineCallPattern.FindAllStringSubmatch(text, -1) {
			body := match[1]
			literals := bizerrStringLiteralPattern.FindAllStringSubmatch(body, -1)
			if len(literals) < 1 {
				continue
			}
			errorCode := strings.TrimSpace(literals[0][1])
			if errorCode == "" {
				continue
			}
			result[errorCode] = deriveBizerrMessageKey(errorCode)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk bizerr definitions under %s: %w", moduleRoot, err)
	}
	return result, nil
}

// deriveBizerrMessageKey mirrors lina-core/pkg/bizerr.MessageKey.
func deriveBizerrMessageKey(errorCode string) string {
	normalized := strings.ToLower(strings.TrimSpace(errorCode))
	segments := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	if len(segments) == 0 {
		return ""
	}
	return "error." + strings.Join(segments, ".")
}

// loadRuntimeLocaleMaps loads flattened runtime JSON keys (excluding nested
// apidoc directories) for each locale under an i18n root.
func loadRuntimeLocaleMaps(i18nDir string) (map[string]map[string]string, error) {
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		return nil, err
	}
	localeMaps := make(map[string]map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		localeDir := filepath.Join(i18nDir, entry.Name())
		messages, _, readErr := readRuntimeLocaleDirectory(localeDir)
		if readErr != nil {
			return nil, readErr
		}
		if len(messages) > 0 {
			localeMaps[entry.Name()] = messages
		}
	}
	return localeMaps, nil
}

// pluginI18nEnabled reports whether a plugin explicitly enables runtime i18n.
// Missing i18n config or enabled=false means single-language plugin (skip).
func pluginI18nEnabled(pluginYAMLPath string) (bool, error) {
	enabled, _, err := pluginI18nIdentity(pluginYAMLPath, "")
	return enabled, err
}
