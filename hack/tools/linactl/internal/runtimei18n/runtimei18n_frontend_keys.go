// This file validates frontend static $t key references against the effective
// host and plugin runtime i18n catalogs.

package runtimei18n

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var frontendStaticI18NCallPattern = regexp.MustCompile(`\$t\(\s*(?:"([^"\n]+)"|'([^'\n]+)')`)

// frontendI18NKeyReference stores one static frontend $t key usage.
type frontendI18NKeyReference struct {
	Path string
	Line int
	Key  string
}

// validateFrontendI18NKeyReferences validates static frontend $t calls against
// app, host runtime, and plugin runtime catalogs.
func validateFrontendI18NKeyReferences(repoRoot string) ([]string, error) {
	baseCatalogs, err := buildFrontendBaseI18NCatalogs(repoRoot)
	if err != nil {
		return nil, err
	}

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	errors := make([]string, 0)
	hostSourceRoot := filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src")
	hostFiles, err := iterFrontendKeySourceFiles(repoRoot, hostSourceRoot)
	if err != nil {
		return nil, err
	}
	hostCatalogs := copyFrontendI18NCatalogs(baseCatalogs)
	if err := mergeAllSourcePluginI18NCatalogs(hostCatalogs, pluginRoot); err != nil {
		return nil, err
	}
	hostErrors, err := validateFrontendI18NSourceFiles(repoRoot, "host:web", hostCatalogs, hostFiles)
	if err != nil {
		return nil, err
	}
	errors = append(errors, hostErrors...)

	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			sort.Strings(errors)
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
		pluginSourceRoot := filepath.Join(pluginDir, "frontend")
		pluginFiles, fileErr := iterFrontendKeySourceFiles(repoRoot, pluginSourceRoot)
		if fileErr != nil {
			return nil, fileErr
		}
		if len(pluginFiles) == 0 {
			continue
		}
		effectiveCatalogs := copyFrontendI18NCatalogs(baseCatalogs)
		pluginI18NEnabled, policyErr := sourcePluginRuntimeI18NEnabled(pluginDir)
		if policyErr != nil {
			return nil, policyErr
		}
		if pluginI18NEnabled {
			if mergeErr := mergeRuntimeI18NCatalogRoot(effectiveCatalogs, filepath.Join(pluginDir, "manifest", "i18n")); mergeErr != nil {
				return nil, mergeErr
			}
		}
		pluginErrors, validateErr := validateFrontendI18NSourceFiles(repoRoot, "plugin:"+pluginID, effectiveCatalogs, pluginFiles)
		if validateErr != nil {
			return nil, validateErr
		}
		errors = append(errors, pluginErrors...)
	}
	sort.Strings(errors)
	return errors, nil
}

// buildFrontendBaseI18NCatalogs returns host frontend catalogs shared by app and
// plugin pages.
func buildFrontendBaseI18NCatalogs(repoRoot string) (map[string]map[string]string, error) {
	catalogs := make(map[string]map[string]string)
	namespacedRoots := []string{
		filepath.Join(repoRoot, "apps", "lina-vben", "packages", "locales", "src", "langs"),
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "locales", "langs"),
	}
	for _, root := range namespacedRoots {
		rootCatalogs, err := readNamespacedFrontendI18NCatalogRoot(root)
		if err != nil {
			return nil, err
		}
		mergeFrontendI18NCatalogs(catalogs, rootCatalogs)
	}
	if err := mergeRuntimeI18NCatalogRoot(catalogs, filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n")); err != nil {
		return nil, err
	}
	return catalogs, nil
}

// mergeAllSourcePluginI18NCatalogs merges every source plugin runtime catalog
// into the host frontend effective catalog, matching the runtime bundle exposed
// by the host i18n service.
func mergeAllSourcePluginI18NCatalogs(catalogs map[string]map[string]string, pluginRoot string) error {
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read plugin root %s: %w", pluginRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginDir := filepath.Join(pluginRoot, entry.Name())
		enabled, policyErr := sourcePluginRuntimeI18NEnabled(pluginDir)
		if policyErr != nil {
			return policyErr
		}
		if !enabled {
			continue
		}
		if err := mergeRuntimeI18NCatalogRoot(catalogs, filepath.Join(pluginDir, "manifest", "i18n")); err != nil {
			return err
		}
	}
	return nil
}

// sourcePluginRuntimeI18NEnabled reads the minimal plugin.yaml i18n.enabled
// policy needed by this static checker.
func sourcePluginRuntimeI18NEnabled(pluginDir string) (bool, error) {
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read plugin manifest %s: %w", manifestPath, err)
	}

	lines := strings.Split(string(content), "\n")
	inI18N := false
	i18nIndent := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := leadingSpaceCount(line)
		if !inI18N {
			if trimmed == "i18n:" {
				inI18N = true
				i18nIndent = indent
			}
			continue
		}
		if indent <= i18nIndent {
			return false, nil
		}
		if !strings.HasPrefix(trimmed, "enabled:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:"))
		return strings.Trim(value, `"'`) == "true", nil
	}
	return false, nil
}

// leadingSpaceCount counts the indentation prefix for minimal YAML block
// traversal.
func leadingSpaceCount(line string) int {
	for index, char := range line {
		if char != ' ' {
			return index
		}
	}
	return len(line)
}

// readNamespacedFrontendI18NCatalogRoot reads locale JSON files whose filename
// is used as the top-level frontend message namespace.
func readNamespacedFrontendI18NCatalogRoot(root string) (map[string]map[string]string, error) {
	catalogs := make(map[string]map[string]string)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return catalogs, nil
		}
		return nil, fmt.Errorf("read frontend i18n root %s: %w", root, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		localeDir := filepath.Join(root, entry.Name())
		messages, readErr := readNamespacedFrontendLocaleDirectory(localeDir)
		if readErr != nil {
			return nil, readErr
		}
		if len(messages) > 0 {
			catalogs[entry.Name()] = messages
		}
	}
	return catalogs, nil
}

// readNamespacedFrontendLocaleDirectory reads one frontend locale directory.
func readNamespacedFrontendLocaleDirectory(localeDir string) (map[string]string, error) {
	messages := make(map[string]string)
	walkErr := filepath.WalkDir(localeDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			return nil
		}
		relPath, relErr := filepath.Rel(localeDir, path)
		if relErr != nil {
			return relErr
		}
		namespace := filepath.ToSlash(strings.TrimSuffix(relPath, filepath.Ext(relPath)))
		fileMessages, readErr := readNamespacedFrontendLocaleFile(path, namespace)
		if readErr != nil {
			return readErr
		}
		for key, value := range fileMessages {
			messages[key] = value
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("read frontend locale dir %s: %w", localeDir, walkErr)
	}
	return messages, nil
}

// readNamespacedFrontendLocaleFile flattens one JSON file under its namespace.
func readNamespacedFrontendLocaleFile(path string, namespace string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read frontend locale JSON %s: %w", path, err)
	}
	var payload interface{}
	if err = json.Unmarshal(content, &payload); err != nil {
		return nil, fmt.Errorf("invalid frontend locale JSON %s: %w", path, err)
	}
	return flattenRuntimeJSON(payload, namespace), nil
}

// mergeRuntimeI18NCatalogRoot merges direct host or plugin runtime i18n JSON
// files into existing frontend catalogs.
func mergeRuntimeI18NCatalogRoot(catalogs map[string]map[string]string, i18nDir string) error {
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read runtime i18n root %s: %w", i18nDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		localeDir := filepath.Join(i18nDir, entry.Name())
		messages, _, readErr := readRuntimeLocaleDirectory(localeDir)
		if readErr != nil {
			return readErr
		}
		if len(messages) == 0 {
			continue
		}
		mergeFrontendI18NCatalogs(catalogs, map[string]map[string]string{
			entry.Name(): messages,
		})
	}
	return nil
}

// iterFrontendKeySourceFiles returns frontend source files that can contain
// static $t key references.
func iterFrontendKeySourceFiles(repoRoot string, sourceRoot string) ([]string, error) {
	files := make([]string, 0)
	if _, err := os.Stat(sourceRoot); err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, fmt.Errorf("stat frontend source root %s: %w", sourceRoot, err)
	}
	walkErr := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if _, ok := excludedPathParts[entry.Name()]; ok {
				return filepath.SkipDir
			}
			return nil
		}
		if !isFrontendI18NKeySourceSuffix(path) {
			return nil
		}
		relPath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		relPath = filepath.ToSlash(relPath)
		if pathMatchesAny(relPath, testFixturePatterns) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("scan frontend source root %s: %w", sourceRoot, walkErr)
	}
	sort.Strings(files)
	return files, nil
}

// isFrontendI18NKeySourceSuffix reports whether one frontend file can contain
// static $t key references.
func isFrontendI18NKeySourceSuffix(path string) bool {
	switch filepath.Ext(path) {
	case ".js", ".ts", ".tsx", ".vue":
		return true
	default:
		return false
	}
}

// validateFrontendI18NSourceFiles validates all static $t references in files.
func validateFrontendI18NSourceFiles(repoRoot string, scope string, catalogs map[string]map[string]string, files []string) ([]string, error) {
	errors := make([]string, 0)
	for _, path := range files {
		references, err := extractFrontendI18NKeyReferences(path)
		if err != nil {
			return nil, err
		}
		for _, reference := range references {
			displayPath, relErr := filepath.Rel(repoRoot, reference.Path)
			if relErr != nil {
				return nil, relErr
			}
			displayPath = filepath.ToSlash(displayPath)
			for _, locale := range sortedMapKeys(catalogs) {
				if _, ok := catalogs[locale][reference.Key]; ok {
					continue
				}
				errors = append(errors, fmt.Sprintf(
					"%s: %s missing frontend key referenced by %s:%d: %s",
					scope,
					locale,
					displayPath,
					reference.Line,
					reference.Key,
				))
			}
		}
	}
	return errors, nil
}

// extractFrontendI18NKeyReferences extracts static $t("key") and $t('key')
// references from one frontend source file.
func extractFrontendI18NKeyReferences(path string) ([]frontendI18NKeyReference, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read frontend source file %s: %w", path, err)
	}
	references := make([]frontendI18NKeyReference, 0)
	for index, line := range strings.Split(string(content), "\n") {
		matches := frontendStaticI18NCallPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			key := strings.TrimSpace(match[1])
			if key == "" {
				key = strings.TrimSpace(match[2])
			}
			if key == "" {
				continue
			}
			references = append(references, frontendI18NKeyReference{
				Path: path,
				Line: index + 1,
				Key:  key,
			})
		}
	}
	return references, nil
}

// mergeFrontendI18NCatalogs merges source catalogs into target catalogs.
func mergeFrontendI18NCatalogs(target map[string]map[string]string, source map[string]map[string]string) {
	for locale, messages := range source {
		if _, ok := target[locale]; !ok {
			target[locale] = make(map[string]string)
		}
		for key, value := range messages {
			target[locale][key] = value
		}
	}
}

// copyFrontendI18NCatalogs returns a deep copy of frontend catalogs.
func copyFrontendI18NCatalogs(source map[string]map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(source))
	for locale, messages := range source {
		result[locale] = make(map[string]string, len(messages))
		for key, value := range messages {
			result[locale][key] = value
		}
	}
	return result
}

// emitFrontendKeyCoverage writes human-readable frontend key coverage results.
func emitFrontendKeyCoverage(out io.Writer, errors []string) error {
	if len(errors) == 0 {
		return writeLine(out, "Runtime i18n frontend key coverage passed for host and plugin pages.")
	}
	if err := writeLine(out, fmt.Sprintf("Runtime i18n frontend key coverage found %d issue(s):", len(errors))); err != nil {
		return err
	}
	for _, item := range errors {
		if err := writeLine(out, item); err != nil {
			return err
		}
	}
	return nil
}

// validateModuleLevelFrontendI18NCalls detects $t() calls at module top level
// in frontend source files. Module-level $t() calls resolve at import time,
// before plugin i18n resources may be loaded, causing untranslated keys at
// runtime.
//
// Detection rules:
//   - .vue files: warns when $t() appears in <script>/<script setup> block at
//     brace depth <= 0 (top-level statements outside any function/object/class).
//   - .ts/.js files: warns when $t() appears at brace depth 0 (module top level).
//   - <template> blocks, function bodies, and object literals are not flagged.
func validateModuleLevelFrontendI18NCalls(repoRoot string) ([]string, error) {
	baseCatalogs, err := buildFrontendBaseI18NCatalogs(repoRoot)
	if err != nil {
		return nil, err
	}

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	warnings := make([]string, 0)

	hostSourceRoot := filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src")
	hostFiles, err := iterFrontendKeySourceFiles(repoRoot, hostSourceRoot)
	if err != nil {
		return nil, err
	}
	hostCatalogs := copyFrontendI18NCatalogs(baseCatalogs)
	if err := mergeAllSourcePluginI18NCatalogs(hostCatalogs, pluginRoot); err != nil {
		return nil, err
	}
	hostWarnings, err := extractModuleLevelCalls(repoRoot, "host:web", hostCatalogs, hostFiles)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, hostWarnings...)

	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if os.IsNotExist(err) {
			sort.Strings(warnings)
			return warnings, nil
		}
		return nil, fmt.Errorf("read plugin root %s: %w", pluginRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginID := entry.Name()
		pluginDir := filepath.Join(pluginRoot, pluginID)
		pluginSourceRoot := filepath.Join(pluginDir, "frontend")
		pluginFiles, fileErr := iterFrontendKeySourceFiles(repoRoot, pluginSourceRoot)
		if fileErr != nil {
			return nil, fileErr
		}
		if len(pluginFiles) == 0 {
			continue
		}
		effectiveCatalogs := copyFrontendI18NCatalogs(baseCatalogs)
		pluginI18NEnabled, policyErr := sourcePluginRuntimeI18NEnabled(pluginDir)
		if policyErr != nil {
			return nil, policyErr
		}
		if pluginI18NEnabled {
			if mergeErr := mergeRuntimeI18NCatalogRoot(effectiveCatalogs, filepath.Join(pluginDir, "manifest", "i18n")); mergeErr != nil {
				return nil, mergeErr
			}
		}
		pluginWarnings, validateErr := extractModuleLevelCalls(repoRoot, "plugin:"+pluginID, effectiveCatalogs, pluginFiles)
		if validateErr != nil {
			return nil, validateErr
		}
		warnings = append(warnings, pluginWarnings...)
	}
	sort.Strings(warnings)
	return warnings, nil
}

// extractModuleLevelCalls scans frontend source files for $t() calls at module
// top level and returns warning strings for each occurrence.
func extractModuleLevelCalls(repoRoot string, scope string, catalogs map[string]map[string]string, files []string) ([]string, error) {
	warnings := make([]string, 0)
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read frontend source file %s: %w", path, err)
		}

		displayPath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return nil, relErr
		}
		displayPath = filepath.ToSlash(displayPath)

		isVue := isVueScriptFile(path)
		inScriptBlock := !isVue // non-Vue files are always "in script"
		braceDepth := 0

		for index, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)

			if isVue {
				if strings.HasPrefix(trimmed, "<script") {
					inScriptBlock = true
					continue
				}
				if strings.HasPrefix(trimmed, "</script>") {
					inScriptBlock = false
					continue
				}
			}

			if !inScriptBlock {
				continue
			}

			for _, ch := range line {
				switch ch {
				case '{':
					braceDepth++
				case '}':
					if braceDepth > 0 {
						braceDepth--
					}
				}
			}

			// Only warn at module top level (braceDepth <= 0).
			if braceDepth > 0 {
				continue
			}

			matches := frontendStaticI18NCallPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				key := strings.TrimSpace(match[1])
				if key == "" {
					key = strings.TrimSpace(match[2])
				}
				if key == "" {
					continue
				}
				warnings = append(warnings, fmt.Sprintf(
					"%s: module-level $t() may resolve before plugin i18n loads: %s:%d: %s",
					scope,
					displayPath,
					index+1,
					key,
				))
			}
		}
	}
	return warnings, nil
}

// isVueScriptFile reports whether the file is a Vue single-file component.
func isVueScriptFile(path string) bool {
	return filepath.Ext(path) == ".vue"
}

// emitModuleLevelWarnings writes module-level $t() call warnings.
func emitModuleLevelWarnings(out io.Writer, warnings []string) error {
	if len(warnings) == 0 {
		return writeLine(out, "No module-level $t() calls detected.")
	}
	if err := writeLine(out, fmt.Sprintf("Module-level $t() calls found (%d warning(s)):", len(warnings))); err != nil {
		return err
	}
	for _, item := range warnings {
		if err := writeLine(out, "  warning: "+item); err != nil {
			return err
		}
	}
	return nil
}
