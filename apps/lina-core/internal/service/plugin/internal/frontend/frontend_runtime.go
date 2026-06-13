// This file implements runtime frontend bundle prewarming, declared public
// asset serving, and invalidation operations for enabled plugins.

package frontend

import (
	"context"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/resourcefs"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// PrewarmRuntimeFrontendBundles rebuilds in-memory frontend bundles for all enabled
// dynamic plugins during host startup. A single failed preload does not stop the host;
// errors are collected and returned as one joined error.
func (s *serviceImpl) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "runtime frontend bundle prewarm started registries=%d", len(registries))
	failures := make([]string, 0)
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		if plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic {
			continue
		}
		if registry.Installed != plugintypes.InstalledYes || registry.Status != plugintypes.StatusEnabled {
			s.InvalidateBundle(ctx, registry.PluginId, "plugin_not_enabled_during_prewarm")
			continue
		}

		manifest, manifestErr := s.loadActiveDynamicPluginManifest(ctx, registry)
		if manifestErr != nil {
			failures = append(
				failures,
				gerror.Wrapf(manifestErr, "prewarm dynamic plugin frontend assets failed: %s", registry.PluginId).Error(),
			)
			continue
		}
		if manifest.RuntimeArtifact == nil || len(manifest.PublicAssets) == 0 || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
			s.InvalidateBundle(ctx, manifest.ID, "no_embedded_frontend_assets")
			continue
		}

		if _, err = s.ensureBundle(ctx, manifest); err != nil {
			failures = append(
				failures,
				gerror.Wrapf(err, "prewarm dynamic plugin frontend assets failed: %s", manifest.ID).Error(),
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

// ResolveRuntimeFrontendAsset resolves one declared public asset for public serving.
func (s *serviceImpl) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	if strings.TrimSpace(version) == "" {
		return nil, gerror.New("current plugin version does not exist or has switched")
	}
	manifest, err := s.resolvePublicAssetManifest(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if manifest == nil {
		return nil, gerror.New("current plugin manifest does not exist")
	}
	if strings.TrimSpace(manifest.Version) != strings.TrimSpace(version) {
		return nil, gerror.New("current plugin version does not exist or has switched")
	}
	if len(manifest.PublicAssets) == 0 {
		return nil, gerror.New("current plugin does not declare public assets")
	}
	if plugintypes.NormalizeType(manifest.Type) == plugintypes.TypeSource {
		return s.resolveSourcePublicAsset(ctx, manifest, relativePath)
	}
	if manifest.RuntimeArtifact == nil || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
		return nil, gerror.New("current dynamic plugin does not declare frontend assets")
	}

	resolvedAssetPath, err := resolvePublicAssetDeclaration(manifest.PublicAssets, relativePath)
	if err != nil {
		return nil, err
	}
	bundle, err := s.ensureBundle(ctx, manifest)
	if err != nil {
		return nil, err
	}

	content, contentType, err := bundle.ReadAsset(resolvedAssetPath)
	if err != nil {
		return nil, err
	}
	logger.Debugf(
		ctx,
		"plugin public asset resolved plugin=%s version=%s path=%s contentType=%s",
		pluginID,
		version,
		resolvedAssetPath,
		contentType,
	)
	return &RuntimeFrontendAssetOutput{
		Content:     content,
		ContentType: contentType,
	}, nil
}

// BuildRuntimeFrontendPublicBaseURL returns the stable public base URL for plugin public assets.
func (s *serviceImpl) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return pluginhost.HostedAssetURLPrefix + strings.TrimSpace(pluginID) + "/" + strings.TrimSpace(version) + "/"
}

// InvalidateBundle removes all cached bundle entries for the given plugin ID.
func (s *serviceImpl) InvalidateBundle(ctx context.Context, pluginID string, reason string) {
	invalidateBundle(ctx, pluginID, reason)
}

// InvalidateAllBundles removes every cached runtime frontend bundle.
func (s *serviceImpl) InvalidateAllBundles(ctx context.Context, reason string) {
	invalidateAllBundles(ctx, reason)
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
func (s *serviceImpl) loadActiveDynamicPluginManifest(ctx context.Context, registry *store.PluginRecord) (*catalog.Manifest, error) {
	if registry == nil {
		return nil, gerror.New("plugin registry record cannot be nil")
	}
	if plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic {
		return nil, gerror.New("current plugin is not dynamic")
	}

	release, err := s.storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("dynamic plugin is missing active release: %s", registry.PluginId)
	}
	return s.storeSvc.LoadReleaseManifest(ctx, release)
}

// resolvePublicAssetManifest loads the manifest that owns a requested
// /x-assets plugin version. Dynamic plugins may serve previously installed
// releases while the plugin remains enabled; source plugins use the active
// discovered manifest because they are compiled into the host.
func (s *serviceImpl) resolvePublicAssetManifest(ctx context.Context, pluginID string, version string) (*catalog.Manifest, error) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry != nil && plugintypes.NormalizeType(registry.Type) == plugintypes.TypeDynamic {
		release, releaseErr := s.storeSvc.GetRelease(ctx, pluginID, version)
		if releaseErr != nil {
			return nil, releaseErr
		}
		if release == nil || !isReleaseServable(release) {
			return nil, gerror.New("current dynamic plugin version does not exist or has switched")
		}
		return s.storeSvc.LoadReleaseManifest(ctx, release)
	}

	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// isReleaseServable reports whether a release row is in a state that allows
// versioned public asset serving.
func isReleaseServable(release *store.ReleaseRecord) bool {
	if release == nil {
		return false
	}
	switch strings.TrimSpace(release.Status) {
	case plugintypes.ReleaseStatusActive.String(), plugintypes.ReleaseStatusInstalled.String():
		return true
	default:
		return false
	}
}

// resolveSourcePublicAsset reads one declared source-plugin asset from the
// plugin embedded filesystem or filesystem root.
func (s *serviceImpl) resolveSourcePublicAsset(
	ctx context.Context,
	manifest *catalog.Manifest,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	resolvedAssetPath, err := resolvePublicAssetDeclaration(manifest.PublicAssets, relativePath)
	if err != nil {
		return nil, err
	}
	var content []byte
	if embeddedFiles := catalog.GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		if err = resourcefs.ValidateNoSymlinkPathFromFS(embeddedFiles, resolvedAssetPath); err == nil {
			content, err = fs.ReadFile(embeddedFiles, resolvedAssetPath)
		}
	} else {
		var fullPath string
		fullPath, err = resourcefs.ResolveResourcePath(manifest.RootDir, resolvedAssetPath)
		if err == nil {
			content, err = os.ReadFile(fullPath)
		}
	}
	if err != nil {
		return nil, gerror.Wrapf(err, "source plugin public asset does not exist: %s", resolvedAssetPath)
	}
	contentType := mime.TypeByExtension(filepath.Ext(resolvedAssetPath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	logger.Debugf(ctx, "source plugin public asset resolved plugin=%s version=%s path=%s contentType=%s", manifest.ID, manifest.Version, resolvedAssetPath, contentType)
	return &RuntimeFrontendAssetOutput{
		Content:     content,
		ContentType: contentType,
	}, nil
}

// resolvePublicAssetDeclaration maps one /x-assets relative request path to a
// declared plugin asset path and rejects undeclared resources.
func resolvePublicAssetDeclaration(declarations []*catalog.PublicAssetSpec, requestPath string) (string, error) {
	normalizedRequestPath, ok := normalizePublicAssetRequestPath(requestPath)
	if !ok {
		return "", gerror.Newf("plugin public asset path is invalid: %s", requestPath)
	}
	for _, declaration := range declarations {
		if declaration == nil {
			continue
		}
		mount, mountOK := normalizePublicAssetRequestPath(declaration.Mount)
		source, sourceOK := normalizePublicAssetRequestPath(declaration.Source)
		if !mountOK || !sourceOK {
			continue
		}
		mount = strings.Trim(mount, "/")
		source = strings.Trim(source, "/")
		if source == "" {
			continue
		}
		if mount != "" && normalizedRequestPath != mount && !strings.HasPrefix(normalizedRequestPath, mount+"/") {
			continue
		}
		assetSuffix := normalizedRequestPath
		if mount != "" {
			assetSuffix = strings.TrimPrefix(normalizedRequestPath, mount)
			assetSuffix = strings.TrimPrefix(assetSuffix, "/")
		}
		if assetSuffix == "" {
			assetSuffix = publicAssetIndexFile(declaration)
		}
		resolvedPath := path.Join(source, assetSuffix)
		if resolvedPath != source && !strings.HasPrefix(resolvedPath, source+"/") {
			return "", gerror.Newf("plugin public asset is outside declared source: %s", requestPath)
		}
		return resolvedPath, nil
	}
	return "", gerror.Newf("plugin public asset is not declared: %s", requestPath)
}

// publicAssetIndexFile returns the declaration-specific directory index file.
func publicAssetIndexFile(declaration *catalog.PublicAssetSpec) string {
	if declaration == nil || strings.TrimSpace(declaration.Index) == "" {
		return catalog.DefaultPublicAssetIndex()
	}
	return strings.TrimSpace(declaration.Index)
}

// normalizePublicAssetRequestPath normalizes a browser-facing relative asset
// path while preventing traversal outside the public asset namespace.
func normalizePublicAssetRequestPath(value string) (string, bool) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = path.Clean(normalized)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		if strings.TrimSpace(value) == "" || strings.TrimSpace(value) == "/" || strings.TrimSpace(value) == "." {
			return "", true
		}
		return "", false
	}
	return strings.Trim(normalized, "/"), true
}
