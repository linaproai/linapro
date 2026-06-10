// This file defines plugin-owned manifest/ files as read-only raw resources for
// source and dynamic plugins. Config, SQL, and i18n files remain governed by
// their dedicated lifecycle pipelines when they need to take effect.
package manifestcap

import (
	"bytes"
	"context"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"
)

// EmbeddedFilesResolver returns embedded manifest resources for one source
// plugin. It is injected by the host runtime so this public capability package
// does not depend on pluginhost.
type EmbeddedFilesResolver func(pluginID string) fs.FS

// serviceAdapter reads raw resources under one plugin manifest root.
type serviceAdapter struct {
	pluginID          string
	developmentRoot   string
	embeddedResolver  EmbeddedFilesResolver
	embeddedFiles     fs.FS
	artifactResources map[string][]byte
}

// NewFactory creates a manifest service factory.
func NewFactory(developmentRoot string, embeddedResolvers ...EmbeddedFilesResolver) ServiceFactory {
	var resolver EmbeddedFilesResolver
	if len(embeddedResolvers) > 0 {
		resolver = embeddedResolvers[0]
	}
	return &serviceAdapter{
		developmentRoot:  strings.TrimSpace(developmentRoot),
		embeddedResolver: resolver,
	}
}

// ForPlugin returns a manifest reader scoped to pluginID.
func (s *serviceAdapter) ForPlugin(pluginID string) Service {
	clone := s.clone()
	clone.pluginID = strings.TrimSpace(pluginID)
	if clone.embeddedResolver != nil {
		clone.embeddedFiles = clone.embeddedResolver(clone.pluginID)
	}
	return clone
}

// WithArtifactResources returns a factory clone carrying release-bound manifest
// resources for pluginID. Resource paths are relative to manifest/.
func (s *serviceAdapter) WithArtifactResources(pluginID string, resources map[string][]byte) ServiceFactory {
	clone := s.clone()
	if strings.TrimSpace(pluginID) == "" || len(resources) == 0 {
		return clone
	}
	if clone.artifactResources == nil {
		clone.artifactResources = make(map[string][]byte)
	}
	for path, content := range resources {
		clone.artifactResources[strings.TrimSpace(pluginID)+"\x00"+path] = append([]byte(nil), content...)
	}
	return clone
}

// clone returns a detached adapter copy.
func (s *serviceAdapter) clone() *serviceAdapter {
	if s == nil {
		return &serviceAdapter{}
	}
	clone := &serviceAdapter{
		pluginID:         s.pluginID,
		developmentRoot:  s.developmentRoot,
		embeddedResolver: s.embeddedResolver,
		embeddedFiles:    s.embeddedFiles,
	}
	if len(s.artifactResources) > 0 {
		clone.artifactResources = make(map[string][]byte, len(s.artifactResources))
		for key, content := range s.artifactResources {
			clone.artifactResources[key] = append([]byte(nil), content...)
		}
	}
	return clone
}

// Get returns one raw resource under the current plugin manifest root.
func (s *serviceAdapter) Get(_ context.Context, resourcePath string) ([]byte, error) {
	normalizedPath, err := normalizeManifestResourcePath(resourcePath)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(s.pluginID) == "" {
		return nil, gerror.New("manifest service requires plugin scope")
	}
	if content := s.artifactResourceContent(normalizedPath); len(content) > 0 {
		return content, nil
	}
	if s.embeddedFiles != nil {
		content, err := fs.ReadFile(s.embeddedFiles, path.Join("manifest", normalizedPath))
		if err == nil {
			return content, nil
		}
		if !isFSNotExist(err) {
			return nil, gerror.Wrapf(err, "read embedded manifest resource failed plugin=%s path=%s", s.pluginID, normalizedPath)
		}
	}
	if root := resolveManifestDevelopmentRoot(s.developmentRoot); root != "" {
		filePath := filepath.Join(root, "apps", "lina-plugins", s.pluginID, "manifest", filepath.FromSlash(normalizedPath))
		content, err := readContainedFile(filePath, filepath.Join(root, "apps", "lina-plugins", s.pluginID, "manifest"))
		if err == nil {
			return content, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, nil
}

// Exists reports whether one allowed manifest resource exists.
func (s *serviceAdapter) Exists(ctx context.Context, resourcePath string) (bool, error) {
	content, err := s.Get(ctx, resourcePath)
	if err != nil {
		return false, err
	}
	return len(content) > 0, nil
}

// Scan unmarshals the selected YAML resource, or the nested key inside it, into target.
func (s *serviceAdapter) Scan(ctx context.Context, resourcePath string, key string, target any) error {
	if target == nil {
		return gerror.New("manifest scan target cannot be nil")
	}
	content, err := s.Get(ctx, resourcePath)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return nil
	}
	if strings.TrimSpace(key) == "" {
		if err = yaml.Unmarshal(content, target); err != nil {
			return gerror.Wrapf(err, "scan manifest resource failed path=%s", resourcePath)
		}
		return nil
	}
	jsonDoc, err := gjson.LoadYaml(content)
	if err != nil {
		return gerror.Wrapf(err, "parse manifest resource failed path=%s", resourcePath)
	}
	if err = jsonDoc.Get(strings.TrimSpace(key)).Scan(target); err != nil {
		return gerror.Wrapf(err, "scan manifest resource failed path=%s key=%s", resourcePath, key)
	}
	return nil
}

// artifactResourceContent returns one release-bound manifest resource.
func (s *serviceAdapter) artifactResourceContent(resourcePath string) []byte {
	if s == nil || len(s.artifactResources) == 0 {
		return nil
	}
	content := s.artifactResources[strings.TrimSpace(s.pluginID)+"\x00"+resourcePath]
	if len(content) == 0 {
		return nil
	}
	return append([]byte(nil), content...)
}

// normalizeManifestResourcePath validates one manifest-relative resource path.
func normalizeManifestResourcePath(resourcePath string) (string, error) {
	raw := strings.ReplaceAll(strings.TrimSpace(resourcePath), "\\", "/")
	if raw == "" || raw == "." {
		return "", gerror.New("manifest resource path cannot be empty or root")
	}
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err == nil && parsed.Scheme != "" {
			return "", gerror.Newf("manifest resource path cannot be URL: %s", resourcePath)
		}
	}
	if strings.HasPrefix(raw, "/") {
		return "", gerror.Newf("manifest resource path cannot be absolute: %s", resourcePath)
	}
	if len(raw) >= 2 && ((raw[0] >= 'A' && raw[0] <= 'Z') || (raw[0] >= 'a' && raw[0] <= 'z')) && raw[1] == ':' {
		return "", gerror.Newf("manifest resource path cannot contain drive prefix: %s", resourcePath)
	}
	normalized := path.Clean(raw)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.Newf("manifest resource path escapes manifest root: %s", resourcePath)
	}
	if strings.HasPrefix(normalized, "manifest/") || normalized == "manifest" {
		return "", gerror.Newf("manifest resource path must be relative to manifest root: %s", resourcePath)
	}
	return normalized, nil
}

// readContainedFile reads filePath only when it remains under rootDir.
func readContainedFile(filePath string, rootDir string) ([]byte, error) {
	cleanRoot, err := filepath.Abs(filepath.Clean(rootDir))
	if err != nil {
		return nil, err
	}
	cleanPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
		return nil, gerror.New("manifest resource path escapes manifest root")
	}
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// resolveManifestDevelopmentRoot resolves the repository root for development reads.
func resolveManifestDevelopmentRoot(override string) string {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return filepath.Clean(trimmed)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	current, err := filepath.Abs(workingDir)
	if err != nil {
		return ""
	}
	for depth := 0; depth < 12; depth++ {
		if fileExists(filepath.Join(current, "go.work")) && fileExists(filepath.Join(current, "apps", "lina-core")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

// isFSNotExist reports whether err is a missing embedded file error.
func isFSNotExist(err error) bool {
	return err != nil && (os.IsNotExist(err) || strings.Contains(err.Error(), "file does not exist"))
}

// fileExists reports whether one filesystem path exists.
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
