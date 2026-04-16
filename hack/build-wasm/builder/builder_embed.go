// This file loads declared embed resources and collects frontend and SQL
// assets from either embedded files or clear-text plugin directories.

package builder

import (
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadEmbeddedStaticResourceSet(pluginDir string) (*embeddedStaticResourceSet, error) {
	embedFilePath := filepath.Join(pluginDir, "plugin_embed.go")
	content, err := os.ReadFile(embedFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	patterns, err := parseGoEmbedPatterns(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynamic plugin embed patterns: %w", err)
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("dynamic plugin embed declaration missing //go:embed patterns: %s", embedFilePath)
	}

	files, err := collectEmbeddedPatternFiles(pluginDir, patterns)
	if err != nil {
		return nil, err
	}
	return &embeddedStaticResourceSet{files: files}, nil
}

func parseGoEmbedPatterns(content string) ([]string, error) {
	patterns := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//go:embed ") {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(trimmed, "//go:embed")))
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty //go:embed directive")
		}
		patterns = append(patterns, fields...)
	}
	return patterns, nil
}

func collectEmbeddedPatternFiles(pluginDir string, patterns []string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	for _, pattern := range patterns {
		normalizedPattern := strings.TrimSpace(pattern)
		if normalizedPattern == "" {
			continue
		}
		if strings.HasPrefix(normalizedPattern, "all:") {
			return nil, fmt.Errorf("dynamic plugin embed pattern does not support all: prefix: %s", normalizedPattern)
		}

		cleanPattern := filepath.Clean(filepath.FromSlash(normalizedPattern))
		if cleanPattern == "." || cleanPattern == ".." || filepath.IsAbs(cleanPattern) || strings.HasPrefix(cleanPattern, ".."+string(os.PathSeparator)) {
			return nil, fmt.Errorf("dynamic plugin embed pattern is invalid: %s", normalizedPattern)
		}

		matches, err := filepath.Glob(filepath.Join(pluginDir, cleanPattern))
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("dynamic plugin embed pattern matched nothing: %s", normalizedPattern)
		}
		for _, matchPath := range matches {
			if err = appendEmbeddedPathFiles(files, pluginDir, matchPath); err != nil {
				return nil, err
			}
		}
	}
	return files, nil
}

func appendEmbeddedPathFiles(files map[string][]byte, pluginDir string, targetPath string) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return appendEmbeddedFile(files, pluginDir, targetPath)
	}
	return filepath.WalkDir(targetPath, func(currentPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if currentPath != targetPath && shouldSkipEmbeddedDirectoryEntry(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		return appendEmbeddedFile(files, pluginDir, currentPath)
	})
}

func shouldSkipEmbeddedDirectoryEntry(name string) bool {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return false
	}

	switch trimmedName[0] {
	case '.', '_':
		return true
	default:
		return false
	}
}

func appendEmbeddedFile(files map[string][]byte, pluginDir string, filePath string) error {
	relativePath, err := filepath.Rel(pluginDir, filePath)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	files[filepath.ToSlash(filepath.Clean(relativePath))] = content
	return nil
}

func loadRuntimeBuildManifest(pluginDir string, embeddedResources *embeddedStaticResourceSet) (*pluginManifest, error) {
	manifest := &pluginManifest{}
	if embeddedResources != nil {
		content, ok := embeddedResources.ReadFile("plugin.yaml")
		if !ok {
			return nil, fmt.Errorf("dynamic plugin embedded resources missing plugin.yaml")
		}
		if err := yaml.Unmarshal(content, manifest); err != nil {
			return nil, fmt.Errorf("failed to load dynamic plugin manifest from embedded resources: %w", err)
		}
		return manifest, nil
	}

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if err := loadYAMLFile(manifestPath, manifest); err != nil {
		return nil, fmt.Errorf("failed to load dynamic plugin manifest: %w", err)
	}
	return manifest, nil
}

func (s *embeddedStaticResourceSet) ReadFile(relativePath string) ([]byte, bool) {
	if s == nil {
		return nil, false
	}
	content, ok := s.files[normalizeEmbeddedResourcePath(relativePath)]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), content...), true
}

func (s *embeddedStaticResourceSet) ListFiles(prefix string, extension string) []string {
	if s == nil {
		return nil
	}
	normalizedPrefix := normalizeEmbeddedResourcePath(prefix)
	if normalizedPrefix != "" && !strings.HasSuffix(normalizedPrefix, "/") {
		normalizedPrefix += "/"
	}

	items := make([]string, 0)
	for filePath := range s.files {
		if normalizedPrefix != "" && !strings.HasPrefix(filePath, normalizedPrefix) {
			continue
		}
		if extension != "" && filepath.Ext(filePath) != extension {
			continue
		}
		items = append(items, filePath)
	}
	sort.Strings(items)
	return items
}

func normalizeEmbeddedResourcePath(value string) string {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
	if normalized == "." {
		return ""
	}
	return normalized
}

func collectFrontendAssets(pluginDir string, embeddedResources *embeddedStaticResourceSet) ([]*frontendAsset, error) {
	if embeddedResources != nil {
		paths := embeddedResources.ListFiles("frontend/pages", "")
		assets := make([]*frontendAsset, 0, len(paths))
		for _, filePath := range paths {
			content, ok := embeddedResources.ReadFile(filePath)
			if !ok {
				return nil, fmt.Errorf("embedded frontend asset not found: %s", filePath)
			}
			relativePath := strings.TrimPrefix(filePath, "frontend/pages/")
			contentType := mime.TypeByExtension(filepath.Ext(filePath))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			assets = append(assets, &frontendAsset{
				Path:          relativePath,
				ContentBase64: base64.StdEncoding.EncodeToString(content),
				ContentType:   contentType,
			})
		}
		return assets, nil
	}

	frontendDir := filepath.Join(pluginDir, "frontend", "pages")
	info, err := os.Stat(frontendDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*frontendAsset{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("runtime frontend pages path is not a directory: %s", frontendDir)
	}

	paths := make([]string, 0)
	if err = filepath.WalkDir(frontendDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(paths)
	assets := make([]*frontendAsset, 0, len(paths))
	for _, filePath := range paths {
		relativePath, err := filepath.Rel(frontendDir, filePath)
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assets = append(assets, &frontendAsset{
			Path:          filepath.ToSlash(relativePath),
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			ContentType:   contentType,
		})
	}
	return assets, nil
}

func collectSQLAssets(pluginDir string, embeddedResources *embeddedStaticResourceSet, uninstall bool) ([]*sqlAsset, error) {
	if embeddedResources != nil {
		searchPrefix := "manifest/sql"
		if uninstall {
			searchPrefix = "manifest/sql/uninstall"
		}

		paths := embeddedResources.ListFiles(searchPrefix, ".sql")
		assets := make([]*sqlAsset, 0, len(paths))
		for _, filePath := range paths {
			if !uninstall && strings.HasPrefix(filePath, "manifest/sql/uninstall/") {
				continue
			}
			content, ok := embeddedResources.ReadFile(filePath)
			if !ok {
				return nil, fmt.Errorf("embedded sql asset not found: %s", filePath)
			}
			assets = append(assets, &sqlAsset{
				Key:     filepath.Base(filePath),
				Content: strings.TrimSpace(string(content)),
			})
		}
		return assets, nil
	}

	searchDir := filepath.Join(pluginDir, "manifest", "sql")
	if uninstall {
		searchDir = filepath.Join(pluginDir, "manifest", "sql", "uninstall")
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*sqlAsset{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	assets := make([]*sqlAsset, 0, len(fileNames))
	for _, name := range fileNames {
		sqlPath := filepath.Join(searchDir, name)
		content, err := os.ReadFile(sqlPath)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &sqlAsset{
			Key:     name,
			Content: strings.TrimSpace(string(content)),
		})
	}
	return assets, nil
}
