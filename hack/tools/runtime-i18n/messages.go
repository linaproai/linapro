// This file validates runtime i18n JSON key coverage for host and plugin
// scopes.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// defaultRuntimeLocale is the baseline locale preferred for key coverage checks.
	defaultRuntimeLocale = "zh-CN"
)

// validateRuntimeI18NMessages validates all host and plugin runtime i18n scopes.
func validateRuntimeI18NMessages(repoRoot string) ([]string, error) {
	scopes, err := iterRuntimeI18NScopes(repoRoot)
	if err != nil {
		return nil, err
	}

	errors := make([]string, 0)
	for _, scope := range scopes {
		scopeErrors, scopeErr := validateRuntimeI18NScope(scope.Name, scope.Path)
		if scopeErr != nil {
			return nil, scopeErr
		}
		errors = append(errors, scopeErrors...)
	}
	return errors, nil
}

// runtimeI18NScope stores one host or plugin i18n resource root.
type runtimeI18NScope struct {
	Name string
	Path string
}

// iterRuntimeI18NScopes returns host and plugin runtime i18n directories.
func iterRuntimeI18NScopes(repoRoot string) ([]runtimeI18NScope, error) {
	scopes := []runtimeI18NScope{{
		Name: "host:core",
		Path: filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n"),
	}}

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return scopes, nil
		}
		return nil, fmt.Errorf("read plugin root %s: %w", pluginRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		i18nDir := filepath.Join(pluginRoot, entry.Name(), "manifest", "i18n")
		info, statErr := os.Stat(i18nDir)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return nil, fmt.Errorf("stat plugin i18n dir %s: %w", i18nDir, statErr)
		}
		if info.IsDir() {
			scopes = append(scopes, runtimeI18NScope{
				Name: "plugin:" + entry.Name(),
				Path: i18nDir,
			})
		}
	}
	return scopes, nil
}

// validateRuntimeI18NScope validates key coverage for one i18n resource root.
func validateRuntimeI18NScope(scope string, i18nDir string) ([]string, error) {
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		return nil, fmt.Errorf("read i18n dir %s: %w", i18nDir, err)
	}

	errors := make([]string, 0)
	localeMaps := make(map[string]map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		localeDir := filepath.Join(i18nDir, entry.Name())
		messages, duplicateErrors, readErr := readRuntimeLocaleDirectory(localeDir)
		if readErr != nil {
			return nil, readErr
		}
		for _, item := range duplicateErrors {
			errors = append(errors, scope+": "+item)
		}
		if len(messages) > 0 {
			localeMaps[entry.Name()] = messages
		}
	}
	if len(localeMaps) < 2 {
		return errors, nil
	}

	baselineLocale := chooseRuntimeBaselineLocale(localeMaps)
	baselineKeys := mapKeys(localeMaps[baselineLocale])
	for _, locale := range sortedMapKeys(localeMaps) {
		if locale == baselineLocale {
			continue
		}
		messages := localeMaps[locale]
		for _, key := range sortedDifference(baselineKeys, mapKeys(messages)) {
			errors = append(errors, fmt.Sprintf("%s: %s missing key from %s: %s", scope, locale, baselineLocale, key))
		}
		for _, key := range sortedDifference(mapKeys(messages), baselineKeys) {
			errors = append(errors, fmt.Sprintf("%s: %s has key not present in %s: %s", scope, locale, baselineLocale, key))
		}
	}
	return errors, nil
}

// readRuntimeLocaleDirectory reads direct JSON runtime files from one locale directory.
func readRuntimeLocaleDirectory(localeDir string) (map[string]string, []string, error) {
	entries, err := os.ReadDir(localeDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read locale dir %s: %w", localeDir, err)
	}

	messages := make(map[string]string)
	errors := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		fileMessages, readErr := readRuntimeLocaleFile(filepath.Join(localeDir, entry.Name()))
		if readErr != nil {
			return nil, nil, readErr
		}
		for key, value := range fileMessages {
			if _, exists := messages[key]; exists {
				errors = append(errors, fmt.Sprintf("%s: duplicate runtime i18n key: %s", localeDir, key))
				continue
			}
			messages[key] = value
		}
	}
	return messages, errors, nil
}

// readRuntimeLocaleFile reads and flattens one runtime locale JSON file.
func readRuntimeLocaleFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read locale JSON %s: %w", path, err)
	}
	var payload interface{}
	if err = json.Unmarshal(content, &payload); err != nil {
		return nil, fmt.Errorf("invalid locale JSON %s: %w", path, err)
	}
	return flattenRuntimeJSON(payload, ""), nil
}

// flattenRuntimeJSON flattens nested JSON values into dotted string keys.
func flattenRuntimeJSON(value interface{}, prefix string) map[string]string {
	result := make(map[string]string)
	switch typed := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			keyText := strings.TrimSpace(key)
			if keyText == "" {
				continue
			}
			nestedPrefix := keyText
			if prefix != "" {
				nestedPrefix = prefix + "." + keyText
			}
			for nestedKey, nestedValue := range flattenRuntimeJSON(typed[key], nestedPrefix) {
				result[nestedKey] = nestedValue
			}
		}
	case string:
		result[prefix] = typed
	case float64:
		result[prefix] = fmt.Sprint(typed)
	case bool:
		result[prefix] = fmt.Sprint(typed)
	case nil:
		result[prefix] = ""
	}
	return result
}

// chooseRuntimeBaselineLocale chooses the baseline locale for one scope.
func chooseRuntimeBaselineLocale(localeMaps map[string]map[string]string) string {
	if _, ok := localeMaps[defaultRuntimeLocale]; ok {
		return defaultRuntimeLocale
	}
	if _, ok := localeMaps["en-US"]; ok {
		return "en-US"
	}
	locales := sortedMapKeys(localeMaps)
	return locales[0]
}

// emitMessageCoverage writes human-readable coverage results.
func emitMessageCoverage(out io.Writer, errors []string) error {
	if len(errors) == 0 {
		return writeLine(out, "Runtime i18n message coverage passed for host and plugin scopes.")
	}
	if err := writeLine(out, fmt.Sprintf("Runtime i18n message coverage found %d issue(s):", len(errors))); err != nil {
		return err
	}
	for _, item := range errors {
		if err := writeLine(out, item); err != nil {
			return err
		}
	}
	return nil
}
