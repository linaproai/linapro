// This file separates mutable discovery manifests from the currently active
// manifests so staged dynamic uploads do not immediately replace the release
// that the host is still serving.

package catalog

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// GetDesiredManifest returns the latest discovered manifest for the given plugin ID.
// For dynamic plugins this is the mutable staging artifact stored at the configured
// runtime storage path. Changes here do not take effect until the reconciler archives
// the artifact as an active release.
func (s *serviceImpl) GetDesiredManifest(pluginID string) (*Manifest, error) {
	if pluginID == "" {
		return nil, gerror.New("插件ID不能为空")
	}
	manifests, err := s.ScanManifests()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			return manifest, nil
		}
	}
	return nil, gerror.New("插件不存在")
}

// GetActiveManifest returns the manifest currently in use by the host for serving.
// For dynamic plugins this reloads from the archived active release so live traffic
// sees the stable version while staging changes accumulate. Source plugins always
// return the discovered manifest directly.
func (s *serviceImpl) GetActiveManifest(ctx context.Context, pluginID string) (*Manifest, error) {
	manifest, err := s.GetDesiredManifest(pluginID)
	if err != nil {
		return nil, err
	}
	if manifest == nil || NormalizeType(manifest.Type) != TypeDynamic {
		return manifest, nil
	}

	registry, err := s.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil || registry.Installed != InstalledYes || registry.ReleaseId <= 0 {
		return manifest, nil
	}
	if s.dynamicManifestLoader == nil {
		return manifest, nil
	}
	return s.dynamicManifestLoader.LoadActiveDynamicPluginManifest(ctx, registry)
}
