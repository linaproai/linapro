// This file separates mutable discovery manifests from the currently active
// manifests so staged dynamic uploads do not immediately replace the release
// that the host is still serving.

package catalog

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ScanManifestsByID returns the discovered manifests keyed by plugin ID so list
// and projection callers can reuse one scan result across many registry rows.
func (s *serviceImpl) ScanManifestsByID() (map[string]*Manifest, error) {
	manifests, err := s.ScanManifests()
	if err != nil {
		return nil, err
	}

	manifestByID := make(map[string]*Manifest, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		manifestByID[pluginID] = manifest
	}
	return manifestByID, nil
}

// GetDesiredManifest returns the latest discovered manifest for the given plugin ID.
// For dynamic plugins this is the mutable staging artifact stored at the configured
// runtime storage path. Changes here do not take effect until the reconciler archives
// the artifact as an active release.
func (s *serviceImpl) GetDesiredManifest(pluginID string) (*Manifest, error) {
	if pluginID == "" {
		return nil, gerror.New("plugin ID cannot be empty")
	}
	manifestByID, err := s.ScanManifestsByID()
	if err != nil {
		return nil, err
	}
	if manifest, ok := manifestByID[pluginID]; ok {
		return manifest, nil
	}
	return nil, gerror.New("plugin does not exist")
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
