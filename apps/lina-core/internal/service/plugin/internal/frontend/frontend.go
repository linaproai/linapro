// Package frontend manages in-memory frontend asset bundles built from runtime WASM
// artifacts. The WASM artifact is the single source of truth; bundles are cached in
// memory and rebuilt on demand after a host restart, eliminating extraction to disk.

package frontend

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
)

// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
type RuntimeFrontendAssetOutput struct {
	// Content is the raw asset body.
	Content []byte
	// ContentType is the HTTP Content-Type header value returned to browsers.
	ContentType string
}

// Service defines the frontend service contract.
type Service interface {
	// EnsureBundleReader returns a BundleReader for the manifest, building and caching the bundle if needed.
	EnsureBundleReader(ctx context.Context, manifest *catalog.Manifest) (BundleReader, error)
	// ValidateRuntimeFrontendMenuBindings verifies that dynamic plugin menus only reference
	// hosted assets that exist in the plugin's in-memory bundle.
	ValidateRuntimeFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error
	// ValidateHostedMenuBindings is the exported form of validateHostedMenuBindings for cross-package access.
	ValidateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error
	// PrewarmRuntimeFrontendBundles rebuilds in-memory frontend bundles for all enabled
	// dynamic plugins during host startup. A single failed preload does not stop the host;
	// errors are collected and returned as one joined error.
	PrewarmRuntimeFrontendBundles(ctx context.Context) error
	// ResolveRuntimeFrontendAsset resolves one enabled dynamic plugin frontend asset for public serving.
	ResolveRuntimeFrontendAsset(
		ctx context.Context,
		pluginID string,
		version string,
		relativePath string,
	) (*RuntimeFrontendAssetOutput, error)
	// BuildRuntimeFrontendPublicBaseURL returns the stable public base URL for runtime plugin assets.
	BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string
	// InvalidateBundle removes all cached bundle entries for the given plugin ID.
	InvalidateBundle(ctx context.Context, pluginID string, reason string)
	// EnsureBundle guarantees an in-memory frontend bundle exists for the given manifest,
	// building and caching it if necessary. Returns the bundle for immediate use.
	// This is called by the runtime reconciler to pre-warm bundles after reconciliation.
	EnsureBundle(ctx context.Context, manifest *catalog.Manifest) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	catalogSvc catalog.Service
}

func New(catalogSvc catalog.Service) Service {
	return &serviceImpl{catalogSvc: catalogSvc}
}

// PrewarmRuntimeFrontendBundles rebuilds in-memory frontend bundles for all enabled
// dynamic plugins during host startup. A single failed preload does not stop the host;
// errors are collected and returned as one joined error.
func (s *serviceImpl) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "runtime frontend bundle prewarm started registries=%d", len(registries))
	failures := make([]string, 0)
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		if catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
			continue
		}
		if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
			s.InvalidateBundle(ctx, registry.PluginId, "plugin_not_enabled_during_prewarm")
			continue
		}

		manifest, manifestErr := s.loadActiveDynamicPluginManifest(ctx, registry)
		if manifestErr != nil {
			failures = append(
				failures,
				gerror.Wrapf(manifestErr, "预热动态插件前端资源失败: %s", registry.PluginId).Error(),
			)
			continue
		}
		if manifest.RuntimeArtifact == nil || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
			s.InvalidateBundle(ctx, manifest.ID, "no_embedded_frontend_assets")
			continue
		}

		if _, err = s.ensureBundle(ctx, manifest); err != nil {
			failures = append(
				failures,
				gerror.Wrapf(err, "预热动态插件前端资源失败: %s", manifest.ID).Error(),
			)
			logger.Debugf(ctx, "runtime frontend bundle prewarm failed plugin=%s err=%v", manifest.ID, err)
			continue
		}
		logger.Debugf(ctx, "runtime frontend bundle prewarm succeeded plugin=%s version=%s", manifest.ID, manifest.Version)
	}

	if len(failures) > 0 {
		return gerror.New(strings.Join(failures, "; "))
	}
	logger.Debugf(ctx, "runtime frontend bundle prewarm finished")
	return nil
}

// ResolveRuntimeFrontendAsset resolves one enabled dynamic plugin frontend asset for public serving.
func (s *serviceImpl) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		return nil, gerror.New("当前动态插件未启用")
	}

	if strings.TrimSpace(version) == "" {
		return nil, gerror.New("当前动态插件版本不存在或已切换")
	}
	release, err := s.catalogSvc.GetRelease(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.New("当前动态插件版本不存在或已切换")
	}
	if !isReleaseServable(release) {
		return nil, gerror.New("当前动态插件版本不存在或已切换")
	}

	manifest, err := s.catalogSvc.LoadReleaseManifest(ctx, release)
	if err != nil {
		return nil, err
	}
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}
	if manifest.RuntimeArtifact == nil || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}

	bundle, err := s.ensureBundle(ctx, manifest)
	if err != nil {
		return nil, err
	}

	content, contentType, err := bundle.ReadAsset(relativePath)
	if err != nil {
		return nil, err
	}
	logger.Debugf(
		ctx,
		"runtime frontend asset resolved plugin=%s version=%s path=%s contentType=%s",
		pluginID,
		version,
		strings.TrimSpace(relativePath),
		contentType,
	)
	return &RuntimeFrontendAssetOutput{
		Content:     content,
		ContentType: contentType,
	}, nil
}

// BuildRuntimeFrontendPublicBaseURL returns the stable public base URL for runtime plugin assets.
func (s *serviceImpl) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return "/plugin-assets/" + strings.TrimSpace(pluginID) + "/" + strings.TrimSpace(version) + "/"
}

// InvalidateBundle removes all cached bundle entries for the given plugin ID.
func (s *serviceImpl) InvalidateBundle(ctx context.Context, pluginID string, reason string) {
	invalidateBundle(ctx, pluginID, reason)
}

// EnsureBundle guarantees an in-memory frontend bundle exists for the given manifest,
// building and caching it if necessary. Returns the bundle for immediate use.
// This is called by the runtime reconciler to pre-warm bundles after reconciliation.
func (s *serviceImpl) EnsureBundle(ctx context.Context, manifest *catalog.Manifest) error {
	_, err := s.ensureBundle(ctx, manifest)
	return err
}

// HasFrontendAssets reports whether the manifest contains embedded frontend assets.
func HasFrontendAssets(manifest *catalog.Manifest) bool {
	return manifest != nil &&
		manifest.RuntimeArtifact != nil &&
		len(manifest.RuntimeArtifact.FrontendAssets) > 0
}

// loadActiveDynamicPluginManifest returns the currently active dynamic-plugin manifest
// reloaded from the stable release archive.
func (s *serviceImpl) loadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*catalog.Manifest, error) {
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
	return s.catalogSvc.LoadReleaseManifest(ctx, release)
}

// isReleaseServable reports whether a release row is in a state that allows frontend serving.
func isReleaseServable(release *entity.SysPluginRelease) bool {
	if release == nil {
		return false
	}
	switch strings.TrimSpace(release.Status) {
	case catalog.ReleaseStatusActive.String(), catalog.ReleaseStatusInstalled.String():
		return true
	default:
		return false
	}
}
