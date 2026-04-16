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

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// buildReleaseArtifactRelativePath returns the versioned archive location used by
// sys_plugin_release.package_path for dynamic-plugin artifacts.
func buildReleaseArtifactRelativePath(pluginID string, version string) string {
	return filepath.ToSlash(
		filepath.Join(
			"releases",
			strings.TrimSpace(pluginID),
			strings.TrimSpace(version),
			buildArtifactFileName(pluginID),
		),
	)
}

// archiveReleaseArtifact copies the currently discovered runtime artifact into a
// versioned archive path and returns that stable relative path. Same-version
// refreshes overwrite the archive when bytes differ so the active release always
// points at the exact content currently reconciled.
func (s *serviceImpl) archiveReleaseArtifact(ctx context.Context, manifest *catalog.Manifest) (string, error) {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return "", gerror.New("动态插件归档要求存在有效产物")
	}

	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return "", err
	}

	relativePath := buildReleaseArtifactRelativePath(manifest.ID, manifest.Version)
	targetPath := filepath.Join(storageDir, filepath.FromSlash(relativePath))

	sourcePath := strings.TrimSpace(manifest.RuntimeArtifact.Path)
	if sourcePath == "" {
		return "", gerror.New("动态插件归档缺少产物路径")
	}

	content := gfile.GetBytes(sourcePath)
	if len(content) == 0 {
		return "", gerror.Newf("动态插件归档读取产物失败: %s", sourcePath)
	}
	if gfile.Exists(targetPath) {
		existingContent := gfile.GetBytes(targetPath)
		// Reuse the archived file only when the bytes are identical. A rebuilt
		// artifact with the same version must replace the old archive content.
		if bytes.Equal(existingContent, content) {
			return relativePath, nil
		}
	}
	if err = gfile.Mkdir(filepath.Dir(targetPath)); err != nil {
		return "", gerror.Wrap(err, "创建动态插件 release 归档目录失败")
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return "", gerror.Wrap(err, "写入动态插件 release 归档文件失败")
	}
	return relativePath, nil
}

// resolveReleasePackagePath resolves one persisted release package path into an
// absolute host path. Relative paths are anchored at the runtime storage directory.
func (s *serviceImpl) resolveReleasePackagePath(ctx context.Context, release *entity.SysPluginRelease) (string, error) {
	if release == nil {
		return "", gerror.New("插件 release 不能为空")
	}

	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return "", gerror.Newf("插件 release 缺少 package_path: %s@%s", release.PluginId, release.ReleaseVersion)
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
		return nil, gerror.New("插件 release 不能为空")
	}
	return s.catalogSvc.LoadReleaseManifest(ctx, release)
}

// LoadActiveDynamicPluginManifest implements catalog.DynamicManifestLoader.
// It returns the currently active dynamic-plugin manifest reloaded from the stable
// release archive so live traffic sees the stable version during staged upgrades.
func (s *serviceImpl) LoadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error) {
	if registry == nil {
		return nil, gerror.New("插件注册记录不能为空")
	}
	if catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}

	release, err := s.catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("动态插件缺少当前生效 release: %s", registry.PluginId)
	}
	return s.loadManifestFromRelease(ctx, release)
}

// loadActiveManifest is the private helper used internally by the reconciler
// to reload the active manifest without going through the catalog interface.
func (s *serviceImpl) loadActiveManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error) {
	return s.LoadActiveDynamicPluginManifest(ctx, registry)
}
