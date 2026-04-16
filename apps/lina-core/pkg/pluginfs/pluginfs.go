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
		return "", gerror.Newf("插件资源路径非法: %s", relativePath)
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
		return "", gerror.Newf("插件资源路径越界: %s", relativePath)
	}
	if !gfile.Exists(fullPath) {
		return "", gerror.Newf("插件资源文件不存在: %s", fullPath)
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
			return gerror.New("SQL 资源路径不能为空")
		}

		normalizedPath, err := NormalizeRelativePath(relativePath)
		if err != nil {
			return gerror.Newf("SQL 资源路径非法: %s", relativePath)
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("SQL 资源路径必须放在 %s: %s", expectedPrefix, relativePath)
		}
		if !uninstall && strings.HasPrefix(normalizedPath, "manifest/sql/uninstall/") {
			return gerror.Newf("安装 SQL 不允许放在 manifest/sql/uninstall/: %s", relativePath)
		}
		if path.Dir(normalizedPath) != expectedDir {
			return gerror.Newf("SQL 资源必须放在 %s 根目录: %s", expectedDir, relativePath)
		}
		if !IsValidSQLFileName(normalizedPath) {
			return gerror.Newf("SQL 文件名必须使用 {序号}-{当前迭代名称}.sql: %s", relativePath)
		}
		if !exists(normalizedPath) {
			return gerror.Newf("SQL 资源文件不存在: %s", relativePath)
		}
	}

	return nil
}

func validateFilePaths(
	relativePaths []string,
	expectedPrefix string,
	allowedExt map[string]struct{},
	exists func(normalizedPath string) bool,
) error {
	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("插件资源路径不能为空")
		}

		normalizedPath, err := NormalizeRelativePath(relativePath)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("插件资源路径必须放在 %s 下: %s", expectedPrefix, relativePath)
		}
		if len(allowedExt) > 0 {
			if _, ok := allowedExt[strings.ToLower(path.Ext(normalizedPath))]; !ok {
				return gerror.Newf("插件资源文件类型不支持: %s", relativePath)
			}
		}
		if !exists(normalizedPath) {
			return gerror.Newf("插件资源文件不存在: %s", relativePath)
		}
	}

	return nil
}
