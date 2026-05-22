// Package toolutil provides small cross-platform helpers shared by linactl
// commands and internal components. The package intentionally contains only
// stateless primitives so command packages can keep orchestration separate from
// filesystem, environment, and CLI parsing details.
package toolutil

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

// ParseBool parses command-line boolean forms accepted by linactl.
func ParseBool(value string, _ bool) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
}

// IsConnectionFailure detects common database connection failure messages.
func IsConnectionFailure(text string) bool {
	patterns := []string{"dial tcp", "connection refused", "connect: connection", "failed to connect", "i/o timeout", "no such host"}
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// DownloadFile downloads a URL to a local file.
func DownloadFile(ctx context.Context, url string, dst string) error {
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

// ExecutableName returns the platform-specific executable filename.
func ExecutableName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		return name + ".exe"
	}
	return name
}

// ViteCommand returns the platform-specific Vite binary path.
func ViteCommand(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite.cmd")
	}
	return filepath.Join(root, "apps", "lina-vben", "node_modules", ".bin", "vite")
}

// RelativePath renders a path relative to the repository root when possible.
func RelativePath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// FirstNonEmpty returns the first non-empty value.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// NormalizeParamKey keeps make-style and CLI-style option keys equivalent.
func NormalizeParamKey(key string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(key)), "-", "_")
}

// FileExists reports whether a path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists reports whether a path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// EnvValue returns one environment value from a KEY=VALUE list.
func EnvValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

// SetEnvValue returns a copy of env with one key replaced or appended.
func SetEnvValue(env []string, key string, value string) []string {
	prefix := key + "="
	next := append([]string{}, env...)
	for index, item := range next {
		if strings.HasPrefix(item, prefix) {
			next[index] = prefix + value
			return next
		}
	}
	return append(next, prefix+value)
}

// RemoveEnvValue returns a copy of env without one key.
func RemoveEnvValue(env []string, key string) []string {
	prefix := key + "="
	next := make([]string, 0, len(env))
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			continue
		}
		next = append(next, item)
	}
	return next
}

// init silences the default flag package output for this custom parser.
func init() {
	flag.CommandLine.SetOutput(io.Discard)
}
