// This file provides manifest-derived metadata helpers used when store derives
// governance rows from catalog manifests.

package catalog

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BuildRegistryChecksum returns a review-friendly checksum derived from the manifest source.
// For dynamic plugins, the artifact checksum is returned directly. For source plugins the
// manifest YAML bytes are hashed using SHA-256.
func (s *serviceImpl) BuildRegistryChecksum(manifest *Manifest) string {
	if manifest == nil {
		return ""
	}
	if manifest.RuntimeArtifact != nil {
		return manifest.RuntimeArtifact.Checksum
	}
	content, err := s.ReadSourcePluginManifestContent(manifest)
	if err != nil || len(content) == 0 {
		return ""
	}
	return fmt.Sprintf("%x", sha256sum(content))
}

// BuildPackagePath returns the canonical package path for a manifest used in release rows.
func (s *serviceImpl) BuildPackagePath(manifest *Manifest) string {
	if manifest == nil {
		return ""
	}
	if HasSourcePluginEmbeddedFiles(manifest) {
		return "embedded/source-plugins/" + manifest.ID
	}
	if manifest.RuntimeArtifact != nil && strings.TrimSpace(manifest.RuntimeArtifact.Path) != "" {
		normalizedPath := filepath.ToSlash(filepath.Clean(manifest.RuntimeArtifact.Path))
		if marker := "/releases/"; strings.Contains(normalizedPath, marker) {
			return strings.TrimPrefix(normalizedPath[strings.LastIndex(normalizedPath, marker):], "/")
		}
		return filepath.ToSlash(filepath.Base(normalizedPath))
	}
	return filepath.ToSlash(manifest.RootDir)
}
