// This file builds and invalidates the process-local resource index for
// manifest-declared source-plugin consumer frontend mounts.

package plugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// sourceConsumerFrontendResourceIndex stores manifest-declared consumer
// frontend resources derived from source plugin manifests and asset listings.
type sourceConsumerFrontendResourceIndex struct {
	mounts []*sourceConsumerFrontendMountEntry
}

// sourceConsumerFrontendMountEntry stores one manifest-declared frontend mount
// and the normalized asset set available under frontend/consumer/.
type sourceConsumerFrontendMountEntry struct {
	pluginID    string
	version     string
	mountPath   string
	index       string
	spaFallback bool
	assets      map[string]struct{}
	indexAsset  *RuntimeFrontendAssetOutput
}

// loadSourceConsumerFrontendResourceIndex returns a process-local index of
// source-plugin consumer frontend resources. The index is rebuilt from embedded
// source manifests and their frontend/consumer asset listings after lifecycle
// invalidation or a cluster runtime cache revision refresh.
func (s *serviceImpl) loadSourceConsumerFrontendResourceIndex(ctx context.Context) (*sourceConsumerFrontendResourceIndex, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	s.sourceConsumerFrontendIndexMu.RLock()
	if s.sourceConsumerFrontendIndexReady {
		index := cloneSourceConsumerFrontendResourceIndex(s.sourceConsumerFrontendIndex)
		s.sourceConsumerFrontendIndexMu.RUnlock()
		return index, nil
	}
	s.sourceConsumerFrontendIndexMu.RUnlock()

	s.sourceConsumerFrontendIndexMu.Lock()
	defer s.sourceConsumerFrontendIndexMu.Unlock()
	if s.sourceConsumerFrontendIndexReady {
		return cloneSourceConsumerFrontendResourceIndex(s.sourceConsumerFrontendIndex), nil
	}

	manifests, err := s.catalogSvc.ScanEmbeddedSourceManifests()
	if err != nil {
		return nil, err
	}
	index := &sourceConsumerFrontendResourceIndex{mounts: make([]*sourceConsumerFrontendMountEntry, 0, len(manifests))}
	seenMounts := make(map[string]string, len(manifests))
	for _, manifest := range manifests {
		frontendSpec := activeSourceConsumerFrontendSpec(manifest)
		if frontendSpec == nil {
			continue
		}
		if existingPluginID, exists := seenMounts[frontendSpec.MountPath]; exists {
			return nil, gerror.Newf(
				"source consumer frontend mount path is duplicated: %s used by %s and %s",
				frontendSpec.MountPath,
				existingPluginID,
				manifest.ID,
			)
		}
		if conflictMountPath, conflictPluginID, exists := findSourceConsumerFrontendOverlappingMount(
			seenMounts,
			frontendSpec.MountPath,
		); exists {
			return nil, gerror.Newf(
				"source consumer frontend mount path overlaps: %s used by %s conflicts with %s used by %s",
				frontendSpec.MountPath,
				manifest.ID,
				conflictMountPath,
				conflictPluginID,
			)
		}
		seenMounts[frontendSpec.MountPath] = manifest.ID
		mountEntry := &sourceConsumerFrontendMountEntry{
			pluginID:    manifest.ID,
			version:     manifest.Version,
			mountPath:   frontendSpec.MountPath,
			index:       frontendSpec.Index,
			spaFallback: sourceConsumerSPAFallbackEnabled(frontendSpec),
			assets:      buildSourceConsumerFrontendAssetSet(s.catalogSvc.ListConsumerFrontendPaths(manifest)),
		}
		mountEntry.indexAsset = s.buildSourceConsumerFrontendMountIndexAsset(ctx, manifest, mountEntry)
		index.mounts = append(index.mounts, mountEntry)
	}
	s.sourceConsumerFrontendIndex = index
	s.sourceConsumerFrontendIndexReady = true
	return cloneSourceConsumerFrontendResourceIndex(s.sourceConsumerFrontendIndex), nil
}

// loadSourceConsumerFrontendMountEntries returns cloned mount entries from the
// resource index for internal callers that need mount-level projections.
func (s *serviceImpl) loadSourceConsumerFrontendMountEntries(ctx context.Context) ([]*sourceConsumerFrontendMountEntry, error) {
	index, err := s.loadSourceConsumerFrontendResourceIndex(ctx)
	if err != nil {
		return nil, err
	}
	return cloneSourceConsumerFrontendMountEntries(index.mounts), nil
}

// invalidateSourceConsumerFrontendMounts clears the process-local resource
// index so the next request rebuilds it from current plugin registry inputs.
func (s *serviceImpl) invalidateSourceConsumerFrontendMounts() {
	if s == nil {
		return
	}
	s.sourceConsumerFrontendIndexMu.Lock()
	s.sourceConsumerFrontendIndexReady = false
	s.sourceConsumerFrontendIndex = nil
	s.sourceConsumerFrontendIndexMu.Unlock()
}

// findSourceConsumerFrontendOverlappingMount reports whether mountPath would
// nest under or contain any previously indexed consumer frontend mount.
func findSourceConsumerFrontendOverlappingMount(
	seenMounts map[string]string,
	mountPath string,
) (conflictMountPath string, conflictPluginID string, exists bool) {
	normalizedMountPath := strings.TrimRight(strings.TrimSpace(mountPath), "/")
	for existingMountPath, existingPluginID := range seenMounts {
		normalizedExistingMountPath := strings.TrimRight(strings.TrimSpace(existingMountPath), "/")
		if normalizedMountPath == normalizedExistingMountPath {
			return existingMountPath, existingPluginID, true
		}
		if strings.HasPrefix(normalizedMountPath, normalizedExistingMountPath+"/") ||
			strings.HasPrefix(normalizedExistingMountPath, normalizedMountPath+"/") {
			return existingMountPath, existingPluginID, true
		}
	}
	return "", "", false
}

// match returns the most specific mount entry matching requestPath and the
// asset-relative path under that mount.
func (index *sourceConsumerFrontendResourceIndex) match(requestPath string) (*sourceConsumerFrontendMountEntry, string) {
	if index == nil {
		return nil, ""
	}
	var (
		matchedMount        *sourceConsumerFrontendMountEntry
		matchedRelativePath string
	)
	for _, mount := range index.mounts {
		relativePath, ok := matchSourceConsumerFrontendMountPath(requestPath, mount.mountPath)
		if !ok {
			continue
		}
		if matchedMount == nil || len(strings.TrimRight(mount.mountPath, "/")) > len(strings.TrimRight(matchedMount.mountPath, "/")) {
			matchedMount = mount
			matchedRelativePath = relativePath
		}
	}
	return matchedMount, matchedRelativePath
}

// assetDeclared reports whether the mount index includes the requested asset-relative path.
func (mount *sourceConsumerFrontendMountEntry) assetDeclared(relativePath string) bool {
	if mount == nil {
		return false
	}
	assetPath, err := normalizeSourceConsumerFrontendAssetPath(relativePath)
	if err != nil {
		return false
	}
	return sourceConsumerFrontendAssetSet(mount.assets).has(assetPath)
}

// cloneSourceConsumerFrontendResourceIndex copies the resource index so callers cannot mutate cached state.
func cloneSourceConsumerFrontendResourceIndex(index *sourceConsumerFrontendResourceIndex) *sourceConsumerFrontendResourceIndex {
	if index == nil {
		return &sourceConsumerFrontendResourceIndex{}
	}
	return &sourceConsumerFrontendResourceIndex{mounts: cloneSourceConsumerFrontendMountEntries(index.mounts)}
}

// cloneSourceConsumerFrontendMountEntries copies mount entries so callers cannot mutate the cached index.
func cloneSourceConsumerFrontendMountEntries(mounts []*sourceConsumerFrontendMountEntry) []*sourceConsumerFrontendMountEntry {
	clonedMounts := make([]*sourceConsumerFrontendMountEntry, 0, len(mounts))
	for _, mount := range mounts {
		if mount == nil {
			continue
		}
		clonedMount := *mount
		clonedMount.assets = cloneSourceConsumerFrontendAssetSet(mount.assets)
		clonedMount.indexAsset = cloneFrontendAssetOutput(mount.indexAsset)
		clonedMounts = append(clonedMounts, &clonedMount)
	}
	return clonedMounts
}

// cloneSourceConsumerFrontendAssetSet copies asset membership for immutable index reads.
func cloneSourceConsumerFrontendAssetSet(assets map[string]struct{}) map[string]struct{} {
	if len(assets) == 0 {
		return nil
	}
	clonedAssets := make(map[string]struct{}, len(assets))
	for assetPath := range assets {
		clonedAssets[assetPath] = struct{}{}
	}
	return clonedAssets
}

// activeSourceConsumerFrontendSpec returns the normalized enabled frontend spec for a source plugin.
func activeSourceConsumerFrontendSpec(manifest *catalog.Manifest) *catalog.ConsumerFrontendSpec {
	if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeSource {
		return nil
	}
	if err := catalog.NormalizeConsumerSpec(manifest); err != nil {
		return nil
	}
	if manifest.Consumer == nil || manifest.Consumer.Frontend == nil {
		return nil
	}
	frontendSpec := manifest.Consumer.Frontend
	if strings.TrimSpace(frontendSpec.MountPath) == "" {
		return nil
	}
	return frontendSpec
}
