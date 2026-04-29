// Package pluginfs provides shared filesystem helpers for plugin manifests,
// embedded assets, and convention-based resource discovery.
package pluginfs

import (
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

const (
	// EmbeddedManifestPath is the canonical embedded plugin manifest path.
	EmbeddedManifestPath = "plugin.yaml"
)

// Shared filename patterns and allowed asset extensions used by plugin path validation.
var (
	sqlFileNamePattern = regexp.MustCompile(`^\d{3}-[a-z0-9-]+\.sql$`)
	vueFileExts        = map[string]struct{}{".vue": {}}
)

// BuildEmbeddedManifestPath builds the host-visible virtual manifest path for one embedded source plugin.
func BuildEmbeddedManifestPath(pluginID string, relativePath string) string {
	normalizedPluginID := strings.TrimSpace(pluginID)
	normalizedPath, err := NormalizeRelativePath(relativePath)
	if err != nil {
		normalizedPath = EmbeddedManifestPath
	}
	if normalizedPath == "" {
		normalizedPath = EmbeddedManifestPath
	}
	if normalizedPluginID == "" {
		return path.Join("embedded", "source-plugins", normalizedPath)
	}
	return path.Join("embedded", "source-plugins", normalizedPluginID, normalizedPath)
}

// NormalizeRelativePath normalizes one plugin-relative path and rejects empty or escaping values.
func NormalizeRelativePath(relativePath string) (string, error) {
	normalizedPath := path.Clean(strings.ReplaceAll(strings.TrimSpace(relativePath), "\\", "/"))
	if normalizedPath == "" || normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
		return "", gerror.Newf("plugin resource path is invalid: %s", relativePath)
	}
	return normalizedPath, nil
}

// ResolveResourcePath resolves one plugin-relative path to an absolute file inside the plugin root.
func ResolveResourcePath(rootDir string, relativePath string) (string, error) {
	normalizedPath, err := NormalizeRelativePath(relativePath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Clean(filepath.Join(rootDir, filepath.FromSlash(normalizedPath)))
	rootPath := filepath.Clean(rootDir)
	if fullPath != rootPath && !strings.HasPrefix(fullPath, rootPath+string(filepath.Separator)) {
		return "", gerror.Newf("plugin resource path escapes the plugin root: %s", relativePath)
	}
	if !gfile.Exists(fullPath) {
		return "", gerror.Newf("plugin resource file does not exist: %s", fullPath)
	}
	return fullPath, nil
}

// DiscoverSQLPaths discovers plugin SQL files by directory convention.
func DiscoverSQLPaths(rootDir string, uninstall bool) []string {
	var (
		searchDir = filepath.Join(rootDir, "manifest", "sql")
		relPrefix = "manifest/sql"
	)

	if uninstall {
		searchDir = filepath.Join(rootDir, "manifest", "sql", "uninstall")
		relPrefix = "manifest/sql/uninstall"
	}

	if !gfile.Exists(searchDir) || !gfile.IsDir(searchDir) {
		return []string{}
	}

	sqlFiles, err := gfile.ScanDirFile(searchDir, "*.sql", false)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(sqlFiles))
	for _, sqlFile := range sqlFiles {
		items = append(items, path.Join(relPrefix, filepath.Base(sqlFile)))
	}
	sort.Strings(items)
	return items
}

// DiscoverSQLPathsFromFS discovers plugin SQL files from one embedded filesystem.
func DiscoverSQLPathsFromFS(fileSystem fs.FS, uninstall bool) []string {
	searchDir := "manifest/sql"
	if uninstall {
		searchDir = "manifest/sql/uninstall"
	}

	entries, err := fs.ReadDir(fileSystem, searchDir)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		items = append(items, path.Join(searchDir, entry.Name()))
	}
	sort.Strings(items)
	return items
}

// DiscoverMockSQLPaths discovers plugin mock-data SQL files by directory convention.
// Mock-data files live under manifest/sql/mock-data/ and are deliberately excluded
// from the install/uninstall scans returned by DiscoverSQLPaths.
func DiscoverMockSQLPaths(rootDir string) []string {
	searchDir := filepath.Join(rootDir, "manifest", "sql", "mock-data")
	relPrefix := "manifest/sql/mock-data"

	if !gfile.Exists(searchDir) || !gfile.IsDir(searchDir) {
		return []string{}
	}

	sqlFiles, err := gfile.ScanDirFile(searchDir, "*.sql", false)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(sqlFiles))
	for _, sqlFile := range sqlFiles {
		items = append(items, path.Join(relPrefix, filepath.Base(sqlFile)))
	}
	sort.Strings(items)
	return items
}

// DiscoverMockSQLPathsFromFS discovers plugin mock-data SQL files from one embedded filesystem.
func DiscoverMockSQLPathsFromFS(fileSystem fs.FS) []string {
	searchDir := "manifest/sql/mock-data"

	entries, err := fs.ReadDir(fileSystem, searchDir)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		items = append(items, path.Join(searchDir, entry.Name()))
	}
	sort.Strings(items)
	return items
}

// DiscoverVuePaths discovers plugin Vue resources under one relative directory.
func DiscoverVuePaths(rootDir string, relativeDir string) []string {
	searchDir := filepath.Join(rootDir, relativeDir)
	if !gfile.Exists(searchDir) || !gfile.IsDir(searchDir) {
		return []string{}
	}

	resourceFiles, err := gfile.ScanDirFile(searchDir, "*.vue", true)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(resourceFiles))
	for _, resourceFile := range resourceFiles {
		relativePath, relErr := filepath.Rel(rootDir, resourceFile)
		if relErr != nil {
			continue
		}
		items = append(items, path.Clean(strings.ReplaceAll(relativePath, "\\", "/")))
	}
	sort.Strings(items)
	return items
}

// DiscoverVuePathsFromFS discovers plugin Vue resources from one embedded filesystem.
func DiscoverVuePathsFromFS(fileSystem fs.FS, searchDir string) []string {
	items := make([]string, 0)
	if err := fs.WalkDir(fileSystem, searchDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil || d.IsDir() {
			return walkErr
		}
		if path.Ext(currentPath) != ".vue" {
			return nil
		}
		items = append(items, path.Clean(currentPath))
		return nil
	}); err != nil {
		return []string{}
	}
	sort.Strings(items)
	return items
}

// ValidateSQLPaths validates install or uninstall SQL asset paths under one plugin root.
func ValidateSQLPaths(rootDir string, relativePaths []string, uninstall bool) error {
	return validateSQLPaths(
		relativePaths,
		uninstall,
		func(normalizedPath string) bool {
			return gfile.Exists(filepath.Join(rootDir, filepath.FromSlash(normalizedPath)))
		},
	)
}

// ValidateSQLPathsFromFS validates install or uninstall SQL asset paths under one embedded filesystem.
func ValidateSQLPathsFromFS(fileSystem fs.FS, relativePaths []string, uninstall bool) error {
	return validateSQLPaths(
		relativePaths,
		uninstall,
		func(normalizedPath string) bool {
			_, err := fs.Stat(fileSystem, normalizedPath)
			return err == nil
		},
	)
}

// ValidateMockSQLPaths validates plugin mock-data SQL asset paths under one plugin root.
func ValidateMockSQLPaths(rootDir string, relativePaths []string) error {
	return validateMockSQLPaths(
		relativePaths,
		func(normalizedPath string) bool {
			return gfile.Exists(filepath.Join(rootDir, filepath.FromSlash(normalizedPath)))
		},
	)
}

// ValidateMockSQLPathsFromFS validates plugin mock-data SQL asset paths under one embedded filesystem.
func ValidateMockSQLPathsFromFS(fileSystem fs.FS, relativePaths []string) error {
	return validateMockSQLPaths(
		relativePaths,
		func(normalizedPath string) bool {
			_, err := fs.Stat(fileSystem, normalizedPath)
			return err == nil
		},
	)
}

// ValidateVuePaths validates plugin Vue asset paths under one plugin root.
func ValidateVuePaths(rootDir string, relativePaths []string, expectedPrefix string) error {
	return validateFilePaths(
		relativePaths,
		expectedPrefix,
		vueFileExts,
		func(normalizedPath string) bool {
			return gfile.Exists(filepath.Join(rootDir, filepath.FromSlash(normalizedPath)))
		},
	)
}

// ValidateVuePathsFromFS validates plugin Vue asset paths under one embedded filesystem.
func ValidateVuePathsFromFS(fileSystem fs.FS, relativePaths []string, expectedPrefix string) error {
	return validateFilePaths(
		relativePaths,
		expectedPrefix,
		vueFileExts,
		func(normalizedPath string) bool {
			_, err := fs.Stat(fileSystem, normalizedPath)
			return err == nil
		},
	)
}

// IsValidSQLFileName reports whether one SQL asset name matches the project naming convention.
func IsValidSQLFileName(name string) bool {
	return sqlFileNamePattern.MatchString(path.Base(strings.TrimSpace(name)))
}

// validateSQLPaths validates plugin SQL asset paths against directory, naming,
// and existence rules for install or uninstall manifests.
func validateSQLPaths(relativePaths []string, uninstall bool, exists func(normalizedPath string) bool) error {
	var (
		expectedDir    = "manifest/sql"
		expectedPrefix = "manifest/sql/"
	)

	if uninstall {
		expectedDir = "manifest/sql/uninstall"
		expectedPrefix = "manifest/sql/uninstall/"
	}

	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("SQL resource path cannot be empty")
		}

		normalizedPath, err := NormalizeRelativePath(relativePath)
		if err != nil {
			return gerror.Newf("SQL resource path is invalid: %s", relativePath)
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("SQL resource path must be under %s: %s", expectedPrefix, relativePath)
		}
		if !uninstall {
			if strings.HasPrefix(normalizedPath, "manifest/sql/uninstall/") {
				return gerror.Newf("install SQL cannot be placed under manifest/sql/uninstall/: %s", relativePath)
			}
			if strings.HasPrefix(normalizedPath, "manifest/sql/mock-data/") {
				return gerror.Newf("install SQL cannot be placed under manifest/sql/mock-data/: %s", relativePath)
			}
		}
		if path.Dir(normalizedPath) != expectedDir {
			return gerror.Newf("SQL resource must be placed directly under %s: %s", expectedDir, relativePath)
		}
		if !IsValidSQLFileName(normalizedPath) {
			return gerror.Newf("SQL filename must use {sequence}-{change-name}.sql: %s", relativePath)
		}
		if !exists(normalizedPath) {
			return gerror.Newf("SQL resource file does not exist: %s", relativePath)
		}
	}

	return nil
}

// validateMockSQLPaths validates plugin mock-data SQL asset paths against directory,
// naming, and existence rules.
func validateMockSQLPaths(relativePaths []string, exists func(normalizedPath string) bool) error {
	const (
		expectedDir    = "manifest/sql/mock-data"
		expectedPrefix = "manifest/sql/mock-data/"
	)

	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("Mock SQL resource path cannot be empty")
		}

		normalizedPath, err := NormalizeRelativePath(relativePath)
		if err != nil {
			return gerror.Newf("Mock SQL resource path is invalid: %s", relativePath)
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("Mock SQL resource path must be under %s: %s", expectedPrefix, relativePath)
		}
		if path.Dir(normalizedPath) != expectedDir {
			return gerror.Newf("Mock SQL resource must be placed directly under %s: %s", expectedDir, relativePath)
		}
		if !IsValidSQLFileName(normalizedPath) {
			return gerror.Newf("Mock SQL filename must use {sequence}-{change-name}.sql: %s", relativePath)
		}
		if !exists(normalizedPath) {
			return gerror.Newf("Mock SQL resource file does not exist: %s", relativePath)
		}
	}

	return nil
}

// validateFilePaths validates relative asset paths against the expected
// directory prefix, extension allowlist, and existence contract.
func validateFilePaths(
	relativePaths []string,
	expectedPrefix string,
	allowedExt map[string]struct{},
	exists func(normalizedPath string) bool,
) error {
	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("plugin resource path cannot be empty")
		}

		normalizedPath, err := NormalizeRelativePath(relativePath)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("plugin resource path must be under %s: %s", expectedPrefix, relativePath)
		}
		if len(allowedExt) > 0 {
			if _, ok := allowedExt[strings.ToLower(path.Ext(normalizedPath))]; !ok {
				return gerror.Newf("plugin resource file type is unsupported: %s", relativePath)
			}
		}
		if !exists(normalizedPath) {
			return gerror.Newf("plugin resource file does not exist: %s", relativePath)
		}
	}

	return nil
}
