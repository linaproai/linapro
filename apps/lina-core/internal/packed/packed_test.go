package packed

import (
	"errors"
	"io/fs"
	"os"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// packedConfigTemplate stores the config fields needed by packed manifest tests.
type packedConfigTemplate struct {
	I18n packedI18nConfig `yaml:"i18n"`
}

// packedI18nConfig stores the packed i18n config section.
type packedI18nConfig struct {
	Enabled bool               `yaml:"enabled"`
	Locales []packedI18nLocale `yaml:"locales"`
}

// packedI18nLocale stores one packed i18n locale descriptor.
type packedI18nLocale struct {
	Locale     string `yaml:"locale"`
	NativeName string `yaml:"nativeName"`
}

// TestFilesEmbedPreparedManifestAssets verifies the packed embed FS contains
// the prepared manifest assets expected by runtime startup.
func TestFilesEmbedPreparedManifestAssets(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("manifest/config/config.template.yaml"); errors.Is(err, os.ErrNotExist) {
		t.Skip("packed manifest assets have not been prepared")
	}

	expectedPaths := []string{
		"manifest/sql/001-project-init.sql",
		"manifest/sql/mock-data/003-mock-users.sql",
		"manifest/config/metadata.yaml",
		"manifest/config/config.template.yaml",
	}
	for _, locale := range readPackedConfigTemplate(t).I18n.Locales {
		expectedPaths = append(expectedPaths,
			"manifest/i18n/"+locale.Locale+"/framework.json",
			"manifest/i18n/"+locale.Locale+"/menu.json",
			"manifest/i18n/"+locale.Locale+"/apidoc/common.json",
		)
	}
	sort.Strings(expectedPaths)

	for _, path := range expectedPaths {
		if _, err := fs.ReadFile(Files, path); err != nil {
			t.Fatalf("read embedded manifest asset %q: %v", path, err)
		}
	}
}

// TestFilesExcludeLocalConfig verifies developer-local config.yaml is not
// embedded into the packed manifest asset set.
func TestFilesExcludeLocalConfig(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("manifest"); errors.Is(err, os.ErrNotExist) {
		t.Skip("packed manifest assets have not been prepared")
	}

	_, err := fs.ReadFile(Files, "manifest/config/config.yaml")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected local config to be excluded from embedded assets, got err=%v", err)
	}
}

// TestFilesEmbedUpdatedUploadDefaults verifies the packed manifest assets keep
// the upload-size defaults aligned with the host source defaults.
func TestFilesEmbedUpdatedUploadDefaults(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("manifest/config/config.template.yaml"); errors.Is(err, os.ErrNotExist) {
		t.Skip("packed manifest assets have not been prepared")
	}

	configContent, err := fs.ReadFile(Files, "manifest/config/config.template.yaml")
	if err != nil {
		t.Fatalf("read packed config template: %v", err)
	}
	if !strings.Contains(string(configContent), "maxSize: 20") {
		t.Fatalf("expected packed config template to keep 20MB upload default, got %q", string(configContent))
	}
	if !strings.Contains(string(configContent), "enabled: true") {
		t.Fatalf("expected packed config template to include i18n enabled default, got %q", string(configContent))
	}
	config := readPackedConfigTemplate(t)
	if !config.I18n.Enabled {
		t.Fatal("expected packed config template to enable runtime i18n by default")
	}
	if len(config.I18n.Locales) == 0 {
		t.Fatal("expected packed config template to include runtime i18n locale metadata")
	}
	for _, locale := range config.I18n.Locales {
		if strings.TrimSpace(locale.Locale) == "" || strings.TrimSpace(locale.NativeName) == "" {
			t.Fatalf("expected packed i18n locale metadata to include locale and nativeName, got %+v", locale)
		}
	}
	assertPackedI18nSectionHasNoDirection(t, string(configContent))

	sqlContent, err := fs.ReadFile(Files, "manifest/sql/007-config-management.sql")
	if err != nil {
		t.Fatalf("read packed config-management sql: %v", err)
	}
	if !strings.Contains(string(sqlContent), "'sys.upload.maxSize', '20'") {
		t.Fatalf("expected packed config-management sql to keep 20MB upload default, got %q", string(sqlContent))
	}
}

// readPackedConfigTemplate parses the embedded config template used by packed
// manifest tests.
func readPackedConfigTemplate(t *testing.T) packedConfigTemplate {
	t.Helper()

	configContent, err := fs.ReadFile(Files, "manifest/config/config.template.yaml")
	if err != nil {
		t.Fatalf("read packed config template: %v", err)
	}

	var config packedConfigTemplate
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		t.Fatalf("parse packed config template: %v", err)
	}
	return config
}

// assertPackedI18nSectionHasNoDirection verifies locale direction remains a
// fixed LTR runtime convention instead of a packed configuration option.
func assertPackedI18nSectionHasNoDirection(t *testing.T, configContent string) {
	t.Helper()

	inI18nSection := false
	for _, line := range strings.Split(configContent, "\n") {
		if strings.HasPrefix(line, "i18n:") {
			inI18nSection = true
			continue
		}
		if inI18nSection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") {
			inI18nSection = false
		}
		if inI18nSection && strings.Contains(strings.TrimSpace(line), "direction") {
			t.Fatalf("expected packed i18n config section to omit direction, got line %q", line)
		}
	}
}
