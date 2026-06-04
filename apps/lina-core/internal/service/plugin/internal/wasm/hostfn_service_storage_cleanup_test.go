// This file tests governed storage cleanup wiring for dynamic-plugin host services.

package wasm

import (
	"context"
	"strings"
	"testing"
)

// TestPurgeAuthorizedStoragePathsRequiresConfiguredService verifies lifecycle
// cleanup fails explicitly when storage configuration was not wired.
func TestPurgeAuthorizedStoragePathsRequiresConfiguredService(t *testing.T) {
	previousConfigSvc := currentStorageConfigReader()
	setStorageConfigServiceForTest(nil)
	t.Cleanup(func() {
		setStorageConfigServiceForTest(previousConfigSvc)
	})

	err := PurgeAuthorizedStoragePaths(context.Background(), "test-plugin-storage-cleanup", nil)
	if err == nil {
		t.Fatal("expected missing storage config service to fail cleanup")
	}
	if !strings.Contains(err.Error(), "storage host service is not configured") {
		t.Fatalf("expected storage configuration error, got %v", err)
	}
}
