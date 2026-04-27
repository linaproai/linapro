package packed

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

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
		"manifest/i18n/zh-CN.json",
		"manifest/i18n/en-US.json",
	}

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

	sqlContent, err := fs.ReadFile(Files, "manifest/sql/007-config-management.sql")
	if err != nil {
		t.Fatalf("read packed config-management sql: %v", err)
	}
	if !strings.Contains(string(sqlContent), "'sys.upload.maxSize', '20'") {
		t.Fatalf("expected packed config-management sql to keep 20MB upload default, got %q", string(sqlContent))
	}
}
