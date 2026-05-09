// Package testutil provides shared helpers for plugin sub-component tests.
package testutil

import (
	"fmt"
	"os"

	_ "lina-core/pkg/dbdriver"

	configsvc "lina-core/internal/service/config"
)

// testDynamicStorageDir stores process-local dynamic plugin artifacts created by tests.
var testDynamicStorageDir string

// testSourcePluginRootDir stores process-local source plugin fixtures created by tests.
var testSourcePluginRootDir string

// init allocates isolated plugin test directories so tests do not mutate
// developer or repository fixture paths.
func init() {
	var err error
	testDynamicStorageDir, err = os.MkdirTemp("", "lina-plugin-dynamic-storage-*")
	if err != nil {
		panic(fmt.Sprintf("failed to create isolated dynamic storage dir: %v", err))
	}
	configsvc.SetPluginDynamicStoragePathOverride(testDynamicStorageDir)
	testSourcePluginRootDir, err = os.MkdirTemp("", "lina-plugin-source-root-*")
	if err != nil {
		panic(fmt.Sprintf("failed to create isolated source plugin root: %v", err))
	}
}

// TestDynamicStorageDir returns the process-local runtime storage directory for plugin tests.
func TestDynamicStorageDir() string {
	return testDynamicStorageDir
}
