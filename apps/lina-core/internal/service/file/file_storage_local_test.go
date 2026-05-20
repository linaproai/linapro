// This file verifies local file storage path generation and content writes.

package file

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"lina-core/internal/service/datascope"
)

// TestLocalStoragePutUsesTenantIDWithoutTenantPrefix verifies new uploads keep
// tenant partitioning while omitting the legacy literal t directory.
func TestLocalStoragePutUsesTenantIDWithoutTenantPrefix(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 42)
	storage := NewLocalStorage(t.TempDir())

	storagePath, err := storage.Put(ctx, "Avatar.PNG", strings.NewReader("avatar-content"))
	if err != nil {
		t.Fatalf("put local storage file: %v", err)
	}

	normalizedPath := filepath.ToSlash(storagePath)
	if strings.HasPrefix(normalizedPath, "t/42/") {
		t.Fatalf("expected storage path without literal tenant prefix, got %q", normalizedPath)
	}
	matched, err := regexp.MatchString(`^42/[0-9]{4}/[0-9]{2}/[0-9]{8}_[0-9]{6}_[A-Za-z0-9]{8}\.png$`, normalizedPath)
	if err != nil {
		t.Fatalf("compile storage path pattern: %v", err)
	}
	if !matched {
		t.Fatalf("unexpected storage path format: %q", normalizedPath)
	}

	content, err := os.ReadFile(filepath.Join(storage.basePath, filepath.FromSlash(normalizedPath)))
	if err != nil {
		t.Fatalf("read stored local file: %v", err)
	}
	if string(content) != "avatar-content" {
		t.Fatalf("expected stored content %q, got %q", "avatar-content", string(content))
	}
}
