// This file owns store-local read-model caches for immutable release manifests
// and persisted YAML manifest snapshots. Cached values are returned as detached
// copies so authorization refresh and projection builders cannot mutate shared
// cache state.

package store

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// releaseManifestCacheEntry stores one parsed release manifest together with the
// artifact file identity used to reject stale entries after archive removal or
// same-path replacement.
type releaseManifestCacheEntry struct {
	fileInfo os.FileInfo
	size     int64
	modTime  time.Time
	manifest *catalog.Manifest
}

// releaseManifestCacheKey builds the immutable identity key for a release
// manifest parse result.
func releaseManifestCacheKey(release *ReleaseRecord, absolutePath string) string {
	if release == nil {
		return ""
	}
	return strings.Join([]string{
		fmt.Sprintf("%d", release.Id),
		strings.TrimSpace(release.PluginId),
		strings.TrimSpace(release.ReleaseVersion),
		strings.TrimSpace(release.Checksum),
		filepath.Clean(strings.TrimSpace(absolutePath)),
	}, "|")
}

// getCachedReleaseManifest returns a detached cached release manifest when the
// active archive file still has the same filesystem identity.
func (s *serviceImpl) getCachedReleaseManifest(key string, absolutePath string) *catalog.Manifest {
	if s == nil || strings.TrimSpace(key) == "" {
		return nil
	}
	info, err := os.Stat(filepath.Clean(strings.TrimSpace(absolutePath)))
	if err != nil || info.IsDir() {
		s.removeCachedReleaseManifest(key)
		return nil
	}
	s.cacheMu.RLock()
	entry := s.releaseManifestCache[key]
	s.cacheMu.RUnlock()
	if entry == nil ||
		entry.size != info.Size() ||
		!entry.modTime.Equal(info.ModTime()) ||
		!sameReleaseManifestFile(entry.fileInfo, info) {
		s.removeCachedReleaseManifest(key)
		return nil
	}
	return catalog.CloneManifest(entry.manifest)
}

// storeCachedReleaseManifest records a detached release manifest.
func (s *serviceImpl) storeCachedReleaseManifest(key string, absolutePath string, manifest *catalog.Manifest) {
	if s == nil || strings.TrimSpace(key) == "" || strings.TrimSpace(absolutePath) == "" || manifest == nil {
		return
	}
	info, err := os.Stat(filepath.Clean(strings.TrimSpace(absolutePath)))
	if err != nil || info.IsDir() {
		return
	}
	s.cacheMu.Lock()
	if s.releaseManifestCache == nil {
		s.releaseManifestCache = make(map[string]*releaseManifestCacheEntry)
	}
	s.releaseManifestCache[key] = &releaseManifestCacheEntry{
		fileInfo: info,
		size:     info.Size(),
		modTime:  info.ModTime(),
		manifest: catalog.CloneManifest(manifest),
	}
	s.cacheMu.Unlock()
}

// removeCachedReleaseManifest deletes one release cache entry after a failed file
// guard check. It is safe to call for missing keys.
func (s *serviceImpl) removeCachedReleaseManifest(key string) {
	if s == nil || strings.TrimSpace(key) == "" {
		return
	}
	s.cacheMu.Lock()
	delete(s.releaseManifestCache, key)
	s.cacheMu.Unlock()
}

// sameReleaseManifestFile reports whether two stat results refer to the same
// artifact archive.
func sameReleaseManifestFile(cached os.FileInfo, current os.FileInfo) bool {
	if cached == nil || current == nil {
		return false
	}
	return os.SameFile(cached, current)
}

// invalidateReleaseManifestCacheForPlugin removes release manifest cache
// entries for a plugin after release metadata or authorization changes.
func (s *serviceImpl) invalidateReleaseManifestCacheForPlugin(pluginID string) {
	if s == nil {
		return
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if normalizedPluginID == "" {
		s.releaseManifestCache = make(map[string]*releaseManifestCacheEntry)
		return
	}
	for key, entry := range s.releaseManifestCache {
		if entry != nil && entry.manifest != nil && strings.TrimSpace(entry.manifest.ID) == normalizedPluginID {
			delete(s.releaseManifestCache, key)
			continue
		}
		parts := strings.Split(key, "|")
		if len(parts) > 1 && parts[1] == normalizedPluginID {
			delete(s.releaseManifestCache, key)
		}
	}
}

// manifestSnapshotCacheKey hashes snapshot YAML content after trim
// normalization. Snapshot content is the authority for this cache.
func manifestSnapshotCacheKey(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(trimmed)))
}

// getCachedManifestSnapshot returns a detached parsed YAML snapshot.
func (s *serviceImpl) getCachedManifestSnapshot(key string) *ManifestSnapshot {
	if s == nil || strings.TrimSpace(key) == "" {
		return nil
	}
	s.cacheMu.RLock()
	snapshot := s.manifestSnapshotCache[key]
	s.cacheMu.RUnlock()
	return cloneManifestSnapshot(snapshot)
}

// storeCachedManifestSnapshot records a detached parsed YAML snapshot.
func (s *serviceImpl) storeCachedManifestSnapshot(key string, snapshot *ManifestSnapshot) {
	if s == nil || strings.TrimSpace(key) == "" || snapshot == nil {
		return
	}
	s.cacheMu.Lock()
	if s.manifestSnapshotCache == nil {
		s.manifestSnapshotCache = make(map[string]*ManifestSnapshot)
	}
	s.manifestSnapshotCache[key] = cloneManifestSnapshot(snapshot)
	s.cacheMu.Unlock()
}

// cloneManifestSnapshot returns a detached manifest snapshot projection.
func cloneManifestSnapshot(snapshot *ManifestSnapshot) *ManifestSnapshot {
	if snapshot == nil {
		return nil
	}
	out := *snapshot
	out.Dependencies = plugintypes.CloneDependencySpec(snapshot.Dependencies)
	out.Routes = cloneRouteContracts(snapshot.Routes)
	out.PublicAssets = catalog.ClonePublicAssetSpecs(snapshot.PublicAssets)
	out.RequestedHostServices = cloneHostServiceSpecs(snapshot.RequestedHostServices)
	out.AuthorizedHostServices = cloneHostServiceSpecs(snapshot.AuthorizedHostServices)
	if snapshot.UninstallPurgeStorageData != nil {
		value := *snapshot.UninstallPurgeStorageData
		out.UninstallPurgeStorageData = &value
	}
	return &out
}

// cloneHostServiceSpecs returns detached host service declarations.
func cloneHostServiceSpecs(items []*protocol.HostServiceSpec) []*protocol.HostServiceSpec {
	if len(items) == 0 {
		return []*protocol.HostServiceSpec{}
	}
	out := make([]*protocol.HostServiceSpec, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.Methods = append([]string(nil), item.Methods...)
		next.Paths = append([]string(nil), item.Paths...)
		next.Tables = append([]string(nil), item.Tables...)
		next.Keys = append([]string(nil), item.Keys...)
		next.Resources = cloneHostServiceResourceSpecs(item.Resources)
		out = append(out, &next)
	}
	return out
}

// cloneHostServiceResourceSpecs returns detached governed resource specs.
func cloneHostServiceResourceSpecs(items []*protocol.HostServiceResourceSpec) []*protocol.HostServiceResourceSpec {
	if len(items) == 0 {
		return []*protocol.HostServiceResourceSpec{}
	}
	out := make([]*protocol.HostServiceResourceSpec, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.AllowMethods = append([]string(nil), item.AllowMethods...)
		next.HeaderAllowList = append([]string(nil), item.HeaderAllowList...)
		next.Attributes = cloneStringMap(item.Attributes)
		out = append(out, &next)
	}
	return out
}
