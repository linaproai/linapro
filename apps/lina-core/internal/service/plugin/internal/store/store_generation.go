// This file provides small helpers for the host-level generation model used by
// dynamic plugin installs, upgrades, rollbacks, and node-state convergence.

package store

import (
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// NextGeneration returns the next stable generation number for one plugin registry row.
func NextGeneration(registry *PluginRecord) int64 {
	if registry != nil && registry.Generation > 0 {
		return registry.Generation + 1
	}
	return 1
}

// BuildStableHostState rebuilds the stable host-state enum from current install and
// enablement flags, ignoring any transient reconciling/failed markers.
func BuildStableHostState(registry *PluginRecord) string {
	if registry == nil {
		return plugintypes.HostStateUninstalled.String()
	}
	return plugintypes.DeriveHostState(registry.Installed, registry.Status)
}

// ShouldTrackStagedDynamicRelease reports whether discovery found a newer dynamic
// artifact that should remain staged instead of immediately replacing the active registry version.
func ShouldTrackStagedDynamicRelease(registry *PluginRecord, manifest *catalog.Manifest) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if plugintypes.NormalizeType(registry.Type) != plugintypes.TypeDynamic ||
		plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic {
		return false
	}
	if registry.Installed != plugintypes.InstalledYes {
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
func (s *serviceImpl) BuildRegistryChecksum(manifest *catalog.Manifest) string {
	if manifest == nil || s.catalogSvc == nil {
		return ""
	}
	return s.catalogSvc.BuildRegistryChecksum(manifest)
}
