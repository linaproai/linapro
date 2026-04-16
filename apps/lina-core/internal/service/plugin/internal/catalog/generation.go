// This file provides small helpers for the host-level generation model used by
// dynamic plugin installs, upgrades, rollbacks, and node-state convergence.

package catalog

import (
	"fmt"
	"strings"

	"lina-core/internal/model/entity"
)

// CompareSemanticVersions compares two validated semantic-version strings.
// It returns -1 when left < right, 0 when equal, and 1 when left > right.
func CompareSemanticVersions(left string, right string) (int, error) {
	leftVersion, err := parseSemanticVersion(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := parseSemanticVersion(right)
	if err != nil {
		return 0, err
	}

	switch {
	case leftVersion.Major < rightVersion.Major:
		return -1, nil
	case leftVersion.Major > rightVersion.Major:
		return 1, nil
	case leftVersion.Minor < rightVersion.Minor:
		return -1, nil
	case leftVersion.Minor > rightVersion.Minor:
		return 1, nil
	case leftVersion.Patch < rightVersion.Patch:
		return -1, nil
	case leftVersion.Patch > rightVersion.Patch:
		return 1, nil
	default:
		return 0, nil
	}
}

// NextGeneration returns the next stable generation number for one plugin registry row.
func NextGeneration(registry *entity.SysPlugin) int64 {
	if registry != nil && registry.Generation > 0 {
		return registry.Generation + 1
	}
	return 1
}

// BuildStableHostState rebuilds the stable host-state enum from current install and
// enablement flags, ignoring any transient reconciling/failed markers.
func BuildStableHostState(registry *entity.SysPlugin) string {
	if registry == nil {
		return HostStateUninstalled.String()
	}
	return DeriveHostState(registry.Installed, registry.Status)
}

// ShouldTrackStagedDynamicRelease reports whether discovery found a newer dynamic
// artifact that should remain staged instead of immediately replacing the active registry version.
func ShouldTrackStagedDynamicRelease(registry *entity.SysPlugin, manifest *Manifest) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if NormalizeType(registry.Type) != TypeDynamic ||
		NormalizeType(manifest.Type) != TypeDynamic {
		return false
	}
	if registry.Installed != InstalledYes {
		return false
	}
	if strings.TrimSpace(registry.Version) == "" || strings.TrimSpace(manifest.Version) == "" {
		return false
	}
	return strings.TrimSpace(registry.Version) != strings.TrimSpace(manifest.Version)
}

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
