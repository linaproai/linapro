// This file contains small cross-platform utility helpers used by linactl.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// parseBool parses command-line boolean forms accepted by linactl.
func parseBool(value string, _ bool) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
}

// isConnectionFailure detects common database connection failure messages.
func isConnectionFailure(text string) bool {
	patterns := []string{"dial tcp", "connection refused", "connect: connection", "failed to connect", "i/o timeout", "no such host"}
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// downloadFile downloads a URL to a local file.
func downloadFile(ctx context.Context, url string, dst string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close download response: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s returned %s", url, resp.Status)
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", dst, closeErr)
		}
	}()
	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

// executableName returns the platform-specific executable filename.
func executableName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		return name + ".exe"
	}
	return name
}

// viteCommand returns the platform-specific Vite binary path.
func viteCommand(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite.cmd")
	}
	return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite")
}

// relativePath renders a path relative to the repository root when possible.
func relativePath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// firstNonEmpty returns the first non-empty value.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// normalizeParamKey keeps make-style and CLI-style option keys equivalent.
func normalizeParamKey(key string) string {
	return strings.ReplaceAll(strings.TrimSpace(key), "-", "_")
}

// fileExists reports whether a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists reports whether a path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// init silences the default flag package output for this custom parser.
func init() {
	flag.CommandLine.SetOutput(io.Discard)
}
