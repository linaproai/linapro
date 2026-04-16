package packed

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

func TestFilesEmbedPreparedManifestAssets(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("manifest/config/config.template.yaml"); errors.Is(err, os.ErrNotExist) {
		t.Skip("packed manifest assets have not been prepared")
	}

	expectedPaths := []string{
		"manifest/sql/001-project-init.sql",
		"manifest/sql/mock-data/001-mock-depts.sql",
		"manifest/config/components.yaml",
		"manifest/config/config.template.yaml",
	}

	for _, path := range expectedPaths {
		if _, err := fs.ReadFile(Files, path); err != nil {
			t.Fatalf("read embedded manifest asset %q: %v", path, err)
		}
	}
}

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
