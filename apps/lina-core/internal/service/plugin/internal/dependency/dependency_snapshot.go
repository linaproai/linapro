// This file prepares resolver snapshots from persisted plugin registry and
// release metadata without leaking host DAO or release entities to callers.

package dependency

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/statusflag"
)

// ApplyRegistrySnapshot prefers installed release snapshots for effective
// dependency metadata and marks unknown snapshots conservatively.
func ApplyRegistrySnapshot(
	ctx context.Context,
	storeSvc store.Service,
	snapshot *PluginSnapshot,
	registry *store.PluginRecord,
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
	snapshot.Installed = registry.Installed == statusflag.Installed.Int()
	if !snapshot.Installed {
		return
	}
	release, err := storeSvc.GetRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		snapshot.DependencySnapshotUnknown = true
		return
	}
	releaseSnapshot, err := storeSvc.ParseManifestSnapshot(release.ManifestSnapshot)
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
	snapshot.Dependencies = plugintypes.CloneDependencySpec(releaseSnapshot.Dependencies)
	snapshot.OwnerHostServices = ownerHostServiceSummariesFromSpecs(releaseSnapshot.RequestedHostServices)
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
		cloned.Dependencies = plugintypes.CloneDependencySpec(item.Dependencies)
		cloned.OwnerHostServices = cloneOwnerHostServiceSummaries(item.OwnerHostServices)
		out = append(out, &cloned)
	}
	return out
}
