// This file keeps runtime frontend assets in memory after they are parsed from
// the runtime wasm artifact stored in plugin.dynamic.storagePath. The wasm
// artifact remains the single source of truth while the in-memory bundle avoids
// extracting files to disk before the host can serve them.

package frontend

import (
	"bytes"
	"context"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
)

// bundle stores one dynamic plugin frontend asset set in memory.
type bundle struct {
	// PluginID is the owning plugin identifier.
	PluginID string
	// Version is the release version string this bundle corresponds to.
	Version string
	// Checksum is the artifact checksum at build time.
	Checksum string
	// ContentTypes maps asset path to its HTTP content type.
	ContentTypes map[string]string
	// FileSystem exposes the assets through the standard fs.FS interface.
	FileSystem *bundleFS
}

// bundleFS exposes runtime frontend assets through the standard fs.ReadFile
// contract so the host can treat WASM-embedded files like a read-only virtual
// filesystem without extracting them to disk.
type bundleFS struct {
	Files map[string]*bundleFile
}

// bundleFile stores one immutable runtime asset payload.
type bundleFile struct {
	Content []byte
}

type bundleOpenFile struct {
	name   string
	reader *bytes.Reader
}

type bundleFileInfo struct {
	name string
	size int64
}

var frontendBundleCache = struct {
	items map[string]*bundle
	mu    sync.RWMutex
}{
	items: map[string]*bundle{},
}

// NormalizeAssetPath normalizes a relative frontend asset path into the canonical
// form used as cache keys inside the bundle filesystem. Returns an empty string
// when the path is empty or would escape the bundle root.
func NormalizeAssetPath(relativePath string) string {
	normalized := strings.TrimSpace(relativePath)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return ""
	}
	normalized = filepath.ToSlash(filepath.Clean(normalized))
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return ""
	}
	return normalized
}

func buildBundleCacheKey(pluginID string, version string) string {
	return strings.TrimSpace(pluginID) + "@" + strings.TrimSpace(version)
}

func buildBundle(manifest *catalog.Manifest) (*bundle, error) {
	if manifest == nil {
		return nil, gerror.New("插件清单不能为空")
	}
	if manifest.RuntimeArtifact == nil {
		return nil, gerror.New("当前动态插件缺少有效产物")
	}
	if len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}

	var (
		contentTypes = make(map[string]string, len(manifest.RuntimeArtifact.FrontendAssets))
		files        = make(map[string]*bundleFile, len(manifest.RuntimeArtifact.FrontendAssets))
	)

	for _, asset := range manifest.RuntimeArtifact.FrontendAssets {
		if asset == nil {
			continue
		}
		if asset.Path == "" {
			return nil, gerror.New("当前动态插件前端资源路径不能为空")
		}
		if len(asset.Content) == 0 {
			return nil, gerror.Newf("当前动态插件前端资源内容为空: %s", asset.Path)
		}

		contentType := strings.TrimSpace(asset.ContentType)
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(asset.Path))
		}
		if contentType == "" {
			contentType = http.DetectContentType(asset.Content)
		}

		contentTypes[asset.Path] = contentType
		files[asset.Path] = &bundleFile{Content: asset.Content}
	}
	if len(files) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}

	return &bundle{
		PluginID:     manifest.ID,
		Version:      manifest.Version,
		Checksum:     manifest.RuntimeArtifact.Checksum,
		ContentTypes: contentTypes,
		FileSystem:   &bundleFS{Files: files},
	}, nil
}

func (b *bundle) matchesManifest(manifest *catalog.Manifest) bool {
	if b == nil || manifest == nil || manifest.RuntimeArtifact == nil {
		return false
	}
	if b.PluginID != manifest.ID || b.Version != manifest.Version {
		return false
	}
	checksum := strings.TrimSpace(manifest.RuntimeArtifact.Checksum)
	if checksum == "" {
		return true
	}
	return b.Checksum == checksum
}

// HasAsset reports whether the bundle contains an asset at the given relative path.
func (b *bundle) HasAsset(relativePath string) bool {
	if b == nil || b.FileSystem == nil {
		return false
	}
	normalizedPath, err := normalizeRequestedAssetPath(relativePath)
	if err != nil {
		return false
	}
	_, ok := b.FileSystem.Files[normalizedPath]
	return ok
}

// ReadAsset reads the asset at the given relative path and returns its content and content type.
func (b *bundle) ReadAsset(relativePath string) ([]byte, string, error) {
	if b == nil || b.FileSystem == nil {
		return nil, "", gerror.New("当前动态插件前端资源不可用")
	}
	normalizedPath, err := normalizeRequestedAssetPath(relativePath)
	if err != nil {
		return nil, "", err
	}
	content, err := fs.ReadFile(b.FileSystem, normalizedPath)
	if err != nil {
		return nil, "", gerror.New("当前动态插件前端资源不存在")
	}
	return content, b.ContentTypes[normalizedPath], nil
}

// ReadFile implements fs.ReadFileFS for the in-memory asset bundle.
func (fsys *bundleFS) ReadFile(name string) ([]byte, error) {
	if fsys == nil {
		return nil, fs.ErrNotExist
	}
	normalizedPath := NormalizeAssetPath(name)
	if normalizedPath == "" {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist}
	}
	file, ok := fsys.Files[normalizedPath]
	if !ok || file == nil || len(file.Content) == 0 {
		return nil, &fs.PathError{Op: "readfile", Path: normalizedPath, Err: fs.ErrNotExist}
	}
	return file.Content, nil
}

// Open implements fs.FS so standard-library helpers such as fs.ReadFile can read
// runtime assets from the in-memory bundle without a physical directory.
func (fsys *bundleFS) Open(name string) (fs.File, error) {
	content, err := fsys.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &bundleOpenFile{name: name, reader: bytes.NewReader(content)}, nil
}

func (f *bundleOpenFile) Stat() (fs.FileInfo, error) {
	if f == nil || f.reader == nil {
		return nil, fs.ErrInvalid
	}
	return bundleFileInfo{name: filepath.Base(f.name), size: f.reader.Size()}, nil
}

func (f *bundleOpenFile) Read(p []byte) (int, error) {
	if f == nil || f.reader == nil {
		return 0, fs.ErrInvalid
	}
	return f.reader.Read(p)
}

func (f *bundleOpenFile) Close() error { return nil }

func (fi bundleFileInfo) Name() string       { return fi.name }
func (fi bundleFileInfo) Size() int64        { return fi.size }
func (fi bundleFileInfo) Mode() fs.FileMode  { return 0o444 }
func (fi bundleFileInfo) ModTime() time.Time { return time.Time{} }
func (fi bundleFileInfo) IsDir() bool        { return false }
func (fi bundleFileInfo) Sys() interface{}   { return nil }

func normalizeRequestedAssetPath(relativePath string) (string, error) {
	trimmedPath := strings.TrimSpace(relativePath)
	if trimmedPath == "" || trimmedPath == "/" {
		return "index.html", nil
	}
	normalizedPath := NormalizeAssetPath(relativePath)
	if normalizedPath == "" {
		return "", gerror.Newf("运行时前端资源路径越界: %s", relativePath)
	}
	return normalizedPath, nil
}

// ensureBundle returns the cached bundle for the manifest, building and caching it if needed.
func (s *serviceImpl) ensureBundle(ctx context.Context, manifest *catalog.Manifest) (*bundle, error) {
	if manifest == nil {
		return nil, gerror.New("插件清单不能为空")
	}
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}
	if manifest.RuntimeArtifact == nil {
		return nil, gerror.New("当前动态插件缺少有效产物")
	}

	cacheKey := buildBundleCacheKey(manifest.ID, manifest.Version)
	frontendBundleCache.mu.RLock()
	cached := frontendBundleCache.items[cacheKey]
	frontendBundleCache.mu.RUnlock()
	if cached != nil && cached.matchesManifest(manifest) {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache hit plugin=%s version=%s checksum=%s",
			manifest.ID, manifest.Version, manifest.RuntimeArtifact.Checksum,
		)
		return cached, nil
	}
	if cached != nil {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache stale plugin=%s cachedVersion=%s requestedVersion=%s cachedChecksum=%s requestedChecksum=%s",
			manifest.ID, cached.Version, manifest.Version, cached.Checksum, manifest.RuntimeArtifact.Checksum,
		)
	} else {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache miss plugin=%s version=%s checksum=%s",
			manifest.ID, manifest.Version, manifest.RuntimeArtifact.Checksum,
		)
	}

	built, err := buildBundle(manifest)
	if err != nil {
		return nil, err
	}

	frontendBundleCache.mu.Lock()
	defer frontendBundleCache.mu.Unlock()

	current := frontendBundleCache.items[cacheKey]
	if current != nil && current.matchesManifest(manifest) {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache filled concurrently plugin=%s version=%s checksum=%s",
			manifest.ID, manifest.Version, manifest.RuntimeArtifact.Checksum,
		)
		return current, nil
	}
	frontendBundleCache.items[cacheKey] = built
	logger.Debugf(
		ctx,
		"runtime frontend bundle cached plugin=%s version=%s checksum=%s assets=%d",
		manifest.ID, manifest.Version, manifest.RuntimeArtifact.Checksum, len(built.FileSystem.Files),
	)
	return built, nil
}

// invalidateBundle removes all cached bundle entries for the given plugin ID.
func invalidateBundle(ctx context.Context, pluginID string, reason string) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}

	frontendBundleCache.mu.Lock()
	defer frontendBundleCache.mu.Unlock()

	deletedCount := 0
	for cacheKey := range frontendBundleCache.items {
		if !strings.HasPrefix(cacheKey, normalizedPluginID+"@") {
			continue
		}
		delete(frontendBundleCache.items, cacheKey)
		deletedCount++
	}
	if deletedCount > 0 {
		logger.Debugf(ctx, "runtime frontend bundle invalidated plugin=%s reason=%s deleted=%d", normalizedPluginID, strings.TrimSpace(reason), deletedCount)
		return
	}
	logger.Debugf(ctx, "runtime frontend bundle invalidate skipped plugin=%s reason=%s cache=empty", normalizedPluginID, strings.TrimSpace(reason))
}

// ResetBundleCache clears all in-memory frontend bundles. Intended for use in tests.
func ResetBundleCache() {
	frontendBundleCache.mu.Lock()
	defer frontendBundleCache.mu.Unlock()
	frontendBundleCache.items = map[string]*bundle{}
}

// BundleReader provides read access to a plugin's in-memory frontend asset bundle.
type BundleReader interface {
	// HasAsset reports whether the bundle contains an asset at the given relative path.
	HasAsset(relativePath string) bool
	// ReadAsset reads the asset at the given relative path and returns its content and content type.
	ReadAsset(relativePath string) ([]byte, string, error)
}

// EnsureBundleReader returns a BundleReader for the manifest, building and caching the bundle if needed.
func (s *serviceImpl) EnsureBundleReader(ctx context.Context, manifest *catalog.Manifest) (BundleReader, error) {
	return s.ensureBundle(ctx, manifest)
}
