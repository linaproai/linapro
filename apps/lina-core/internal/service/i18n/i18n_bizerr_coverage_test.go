// This file verifies every host bizerr.Code messageKey exists in host runtime
// error.json for all supported locales. Message keys are always derived from
// error codes via bizerr.MessageKey (convention over configuration).
package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestHostBizerrMessageKeysCoveredByErrorJSON ensures caller-visible host error
// codes resolve to a real runtime i18n key under manifest/i18n/<locale>/error.json.
func TestHostBizerrMessageKeysCoveredByErrorJSON(t *testing.T) {
	moduleRoot := resolveLinaCoreModuleRoot(t)
	codes := collectHostBizerrDefinitions(t, moduleRoot)
	if len(codes) == 0 {
		t.Fatal("expected host bizerr definitions under apps/lina-core")
	}

	locales := []string{DefaultLocale, EnglishLocale}
	for _, locale := range locales {
		keys := loadFlattenedErrorJSON(t, filepath.Join(moduleRoot, "manifest", "i18n", locale, "error.json"))
		var missing []string
		for _, item := range codes {
			if _, ok := keys[item.messageKey]; !ok {
				missing = append(missing, item.errorCode+" -> "+item.messageKey)
			}
		}
		if len(missing) > 0 {
			t.Fatalf(
				"locale %s missing %d bizerr messageKey(s) in error.json:\n  %s",
				locale,
				len(missing),
				strings.Join(missing, "\n  "),
			)
		}
	}
}

type hostBizerrDef struct {
	errorCode  string
	messageKey string
}

var (
	bizerrDefineCallPattern = regexp.MustCompile(
		`(?s)bizerr\.MustDefine\((.*?)\)`,
	)
	bizerrStringLiteralPattern = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

// collectHostBizerrDefinitions walks host Go sources (excluding tests) and
// derives messageKey for each MustDefine call via bizerr.MessageKey.
func collectHostBizerrDefinitions(t *testing.T, moduleRoot string) []hostBizerrDef {
	t.Helper()

	byCode := make(map[string]hostBizerrDef)
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" {
				return filepath.SkipDir
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
		if !strings.Contains(string(content), "bizerr.MustDefine") {
			return nil
		}
		for _, match := range bizerrDefineCallPattern.FindAllStringSubmatch(string(content), -1) {
			body := match[1]
			literals := bizerrStringLiteralPattern.FindAllStringSubmatch(body, -1)
			if len(literals) < 2 {
				continue
			}
			errorCode := literals[0][1]
			byCode[errorCode] = hostBizerrDef{
				errorCode:  errorCode,
				messageKey: bizerr.MessageKey(errorCode),
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk host module for bizerr definitions: %v", err)
	}

	result := make([]hostBizerrDef, 0, len(byCode))
	for _, item := range byCode {
		result = append(result, item)
	}
	return result
}

// loadFlattenedErrorJSON loads and flattens one host error.json catalog.
func loadFlattenedErrorJSON(t *testing.T, path string) map[string]string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var payload any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return flattenJSONStrings(payload, "")
}

// flattenJSONStrings flattens nested JSON objects into dotted string keys.
func flattenJSONStrings(value any, prefix string) map[string]string {
	result := make(map[string]string)
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			keyText := strings.TrimSpace(key)
			if keyText == "" {
				continue
			}
			next := keyText
			if prefix != "" {
				next = prefix + "." + keyText
			}
			for nestedKey, nestedValue := range flattenJSONStrings(nested, next) {
				result[nestedKey] = nestedValue
			}
		}
	case string:
		if prefix != "" {
			result[prefix] = typed
		}
	}
	return result
}

// resolveLinaCoreModuleRoot returns the apps/lina-core directory containing go.mod.
func resolveLinaCoreModuleRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path for lina-core module root")
	}
	dir := filepath.Dir(currentFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate apps/lina-core go.mod")
		}
		dir = parent
	}
}
