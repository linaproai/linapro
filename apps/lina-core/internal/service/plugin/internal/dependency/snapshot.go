// This file prepares resolver snapshots from persisted plugin registry and
// release metadata without leaking host DAO or release entities to callers.

package dependency

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// ReleaseSnapshotReader is the catalog slice required to hydrate effective
// dependency metadata from an installed plugin release.
type ReleaseSnapshotReader interface {
	// GetRegistryRelease returns the active release row for a plugin registry.
	GetRegistryRelease(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPluginRelease, error)
	// ParseManifestSnapshot parses a persisted release manifest snapshot.
	ParseManifestSnapshot(snapshot string) (*catalog.ManifestSnapshot, error)
}

// ApplyRegistrySnapshot prefers installed release snapshots for effective
// dependency metadata and marks unknown snapshots conservatively.
func ApplyRegistrySnapshot(
	ctx context.Context,
	catalogSvc ReleaseSnapshotReader,
	snapshot *PluginSnapshot,
	registry *entity.SysPlugin,
) {
	if snapshot == nil || registry == nil {
		return
	}
	if strings.TrimSpace(registry.Name) != "" {
		snapshot.Name = strings.TrimSpace(registry.Name)
	}
	if strings.TrimSpace(registry.Version) != "" {
		snapshot.Version = strings.TrimSpace(registry.Version)
	}
	snapshot.Installed = registry.Installed == catalog.InstalledYes
	if !snapshot.Installed {
		return
	}
	release, err := catalogSvc.GetRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		snapshot.DependencySnapshotUnknown = true
		return
	}
	releaseSnapshot, err := catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil || releaseSnapshot == nil {
		snapshot.DependencySnapshotUnknown = true
		return
	}
	if strings.TrimSpace(releaseSnapshot.Name) != "" {
		snapshot.Name = strings.TrimSpace(releaseSnapshot.Name)
	}
	if strings.TrimSpace(releaseSnapshot.Version) != "" {
		snapshot.Version = strings.TrimSpace(releaseSnapshot.Version)
	}
	snapshot.Dependencies = catalog.CloneDependencySpec(releaseSnapshot.Dependencies)
}

// ClonePluginSnapshots returns a detached copy so callers cannot mutate the
// cached dependency snapshot slice for later checks in the same request.
func ClonePluginSnapshots(items []*PluginSnapshot) []*PluginSnapshot {
	out := make([]*PluginSnapshot, 0, len(items))
	for _, item := range items {
		if item == nil {
			out = append(out, nil)
			continue
		}
		cloned := *item
		cloned.Dependencies = catalog.CloneDependencySpec(item.Dependencies)
		out = append(out, &cloned)
	}
	return out
}
