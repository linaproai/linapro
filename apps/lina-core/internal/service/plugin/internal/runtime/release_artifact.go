// This file keeps versioned dynamic-plugin artifacts in a release archive so
// the host can stage a new upload without losing access to the currently active
// release that is still serving in-flight requests and old plugin pages.

package runtime

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/logger"
)

// buildReleaseArtifactRelativePath returns the versioned archive location used by
// sys_plugin_release.package_path for dynamic-plugin artifacts.
func buildReleaseArtifactRelativePath(pluginID string, version string, checksum string) string {
	return filepath.ToSlash(
		filepath.Join(
			"releases",
			strings.TrimSpace(pluginID),
			strings.TrimSpace(version),
			strings.TrimSpace(checksum),
			buildArtifactFileName(pluginID),
		),
	)
}

// archiveReleaseArtifact copies the currently discovered runtime artifact into a
// checksum-versioned archive path and returns that stable relative path.
func (s *serviceImpl) archiveReleaseArtifact(ctx context.Context, manifest *catalog.Manifest) (string, error) {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return "", gerror.New("dynamic plugin archive requires a valid artifact")
	}
	checksum := strings.TrimSpace(manifest.RuntimeArtifact.Checksum)
	if checksum == "" {
		return "", gerror.New("dynamic plugin archive requires an artifact checksum")
	}

	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return "", err
	}

	relativePath := buildReleaseArtifactRelativePath(manifest.ID, manifest.Version, checksum)
	targetPath := filepath.Join(storageDir, filepath.FromSlash(relativePath))

	sourcePath := strings.TrimSpace(manifest.RuntimeArtifact.Path)
	if sourcePath == "" {
		return "", gerror.New("dynamic plugin archive is missing artifact path")
	}

	content := gfile.GetBytes(sourcePath)
	if len(content) == 0 {
		return "", gerror.Newf("read dynamic plugin archive artifact failed: %s", sourcePath)
	}
	if gfile.Exists(targetPath) {
		existingContent := gfile.GetBytes(targetPath)
		// Reuse the immutable checksum path only when the bytes are identical.
		if bytes.Equal(existingContent, content) {
			return relativePath, nil
		}
	}
	if err = gfile.Mkdir(filepath.Dir(targetPath)); err != nil {
		return "", gerror.Wrap(err, "create dynamic plugin release archive directory failed")
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return "", gerror.Wrap(err, "write dynamic plugin release archive file failed")
	}
	return relativePath, nil
}

// cleanupStaleReleaseArtifacts removes archived dynamic artifacts that are no
// longer referenced by any release row for the plugin. Active and historical
// release rows keep their current package_path entries, so rollback/drain
// artifacts remain intact while same-version refresh leftovers are cleaned.
func (s *serviceImpl) cleanupStaleReleaseArtifacts(ctx context.Context, pluginID string) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return
	}

	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		logger.Warningf(ctx, "resolve runtime storage dir for artifact cleanup failed plugin=%s err=%v", pluginID, err)
		return
	}

	keepPaths, err := s.referencedReleaseArtifactPaths(ctx, storageDir, pluginID)
	if err != nil {
		logger.Warningf(ctx, "load referenced release artifact paths failed plugin=%s err=%v", pluginID, err)
		return
	}

	releaseRoot := filepath.Join(storageDir, "releases", pluginID)
	if !gfile.Exists(releaseRoot) {
		return
	}
	files, err := gfile.ScanDirFile(releaseRoot, "*.wasm", true)
	if err != nil {
		logger.Warningf(ctx, "scan release artifact archive failed plugin=%s err=%v", pluginID, err)
		return
	}
	for _, file := range files {
		absolutePath := filepath.Clean(file)
		if _, ok := keepPaths[absolutePath]; ok {
			continue
		}
		wasm.InvalidateCache(ctx, absolutePath)
		if removeErr := gfile.Remove(absolutePath); removeErr != nil {
			logger.Warningf(ctx, "remove stale release artifact failed plugin=%s path=%s err=%v", pluginID, absolutePath, removeErr)
		}
	}
}

// referencedReleaseArtifactPaths returns absolute archive paths still referenced
// by release rows so cleanup never removes the current active artifact.
func (s *serviceImpl) referencedReleaseArtifactPaths(
	ctx context.Context,
	storageDir string,
	pluginID string,
) (map[string]struct{}, error) {
	var releases []*entity.SysPluginRelease
	if err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{PluginId: pluginID}).
		Scan(&releases); err != nil {
		return nil, err
	}

	keepPaths := make(map[string]struct{}, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}
		packagePath := strings.TrimSpace(release.PackagePath)
		if packagePath == "" {
			continue
		}
		if filepath.IsAbs(packagePath) {
			keepPaths[filepath.Clean(packagePath)] = struct{}{}
			continue
		}
		keepPaths[filepath.Clean(filepath.Join(storageDir, filepath.FromSlash(packagePath)))] = struct{}{}
	}
	return keepPaths, nil
}

// resolveReleasePackagePath resolves one persisted release package path into an
// absolute host path. Relative paths are anchored at the runtime storage directory.
func (s *serviceImpl) resolveReleasePackagePath(ctx context.Context, release *entity.SysPluginRelease) (string, error) {
	if release == nil {
		return "", gerror.New("plugin release cannot be nil")
	}

	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return "", gerror.Newf("plugin release is missing package_path: %s@%s", release.PluginId, release.ReleaseVersion)
	}
	if filepath.IsAbs(packagePath) {
		return filepath.Clean(packagePath), nil
	}

	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(storageDir, filepath.FromSlash(packagePath))), nil
}

// loadManifestFromRelease reloads one dynamic manifest from its persisted release archive.
func (s *serviceImpl) loadManifestFromRelease(ctx context.Context, release *entity.SysPluginRelease) (*catalog.Manifest, error) {
	if release == nil {
		return nil, gerror.New("plugin release cannot be nil")
	}
	return s.catalogSvc.LoadReleaseManifest(ctx, release)
}

// LoadActiveDynamicPluginManifest implements catalog.DynamicManifestLoader.
// It returns the currently active dynamic-plugin manifest reloaded from the stable
// release archive so live traffic sees the stable version during staged upgrades.
func (s *serviceImpl) LoadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error) {
	if registry == nil {
		return nil, gerror.New("plugin registry record cannot be nil")
	}
	if catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil, gerror.New("current plugin is not dynamic")
	}

	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("dynamic plugin is missing active release: %s", registry.PluginId)
	}
	return s.loadManifestFromRelease(ctx, release)
}

// loadActiveManifest is the private helper used internally by the reconciler
// to reload the active manifest without going through the catalog interface.
func (s *serviceImpl) loadActiveManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error) {
	return s.LoadActiveDynamicPluginManifest(ctx, registry)
}
