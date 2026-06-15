// This file separates mutable discovery manifests from the currently active
// manifests so staged dynamic uploads do not immediately replace the release
// that the host is still serving.

package catalog

import (
	"errors"
	"os"
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
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, gerror.New("plugin ID cannot be empty")
	}
	if manifest, ok, err := s.getDesiredManifestFromRuntimeIndex(normalizedPluginID); err != nil || ok {
		return manifest, err
	}
	manifestByID, err := s.ScanManifestsByID()
	if err != nil {
		return nil, err
	}
	if manifest, ok := manifestByID[normalizedPluginID]; ok {
		return manifest, nil
	}
	return nil, gerror.New("plugin does not exist")
}

// getDesiredManifestFromRuntimeIndex resolves dynamic plugins through the
// cached pluginID-to-artifact index without scanning every artifact.
func (s *serviceImpl) getDesiredManifestFromRuntimeIndex(pluginID string) (*Manifest, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	s.cacheMu.RLock()
	artifactPath := s.runtimePluginArtifactIndex[pluginID]
	s.cacheMu.RUnlock()
	if strings.TrimSpace(artifactPath) == "" {
		return nil, false, nil
	}
	manifest, err := s.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.InvalidateManifestCache(pluginID)
			return nil, false, nil
		}
		return nil, true, err
	}
	if manifest == nil || strings.TrimSpace(manifest.ID) != pluginID {
		s.InvalidateManifestCache(pluginID)
		return nil, false, nil
	}
	return manifest, true, nil
}
