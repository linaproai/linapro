// Package playwright checks local Playwright browser prerequisites for linactl
// E2E commands. It keeps OS-specific cache detection outside the command
// entrypoint while leaving browser installation to the setup command.
package playwright

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureBrowsers checks that Playwright's Chromium browser is installed.
// If the browser cache directory is missing, it prints a clear error with the fix command.
func EnsureBrowsers(_ context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	var cacheDir string
	switch runtime.GOOS {
	case "linux":
		cacheDir = filepath.Join(home, ".cache", "ms-playwright")
	case "darwin":
		cacheDir = filepath.Join(home, "Library", "Caches", "ms-playwright")
	default:
		// Windows and others: Playwright uses a self-contained bundle; skip detection.
		return nil
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Playwright browsers not installed. Run: make env.setup")
		}
		return fmt.Errorf("check Playwright browser cache: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "chromium") {
			return nil
		}
	}

	return fmt.Errorf("Playwright Chromium browser not found in %s. Run: make env.setup", cacheDir)
}
