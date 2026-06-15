// This file owns catalog-local manifest read-model caches. Cached manifests are
// kept immutable by returning detached copies to callers, and dynamic artifacts
// are guarded by filesystem identity, size, and mtime so steady-state scans
// avoid reading full WASM files.

package catalog

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/resourcefs"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// sourceManifestCacheEntry stores one parsed source manifest with the manifest
// file identity used to detect development-time source-plugin changes.
type sourceManifestCacheEntry struct {
	pluginID  string
	sourceKey string
	size      int64
	modTime   time.Time
	manifest  *Manifest
}

// runtimeArtifactCacheEntry stores one parsed runtime artifact manifest and the
// filesystem identity used to decide whether it can be reused.
type runtimeArtifactCacheEntry struct {
	path     string
	fileInfo os.FileInfo
	size     int64
	modTime  time.Time
	manifest *Manifest
}

// InvalidateManifestCache removes cached manifest projections for one plugin
// ID. Empty pluginID clears all dynamic manifest cache entries.
func (s *serviceImpl) InvalidateManifestCache(pluginID string) {
	if s == nil {
		return
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if normalizedPluginID == "" {
		s.sourceManifestCache = make(map[string]*sourceManifestCacheEntry)
		s.runtimeArtifactCache = make(map[string]*runtimeArtifactCacheEntry)
		s.runtimePluginArtifactIndex = make(map[string]string)
		return
	}
	delete(s.sourceManifestCache, normalizedPluginID)
	if artifactPath, ok := s.runtimePluginArtifactIndex[normalizedPluginID]; ok {
		delete(s.runtimeArtifactCache, artifactPath)
		delete(s.runtimePluginArtifactIndex, normalizedPluginID)
		return
	}
	for artifactPath, entry := range s.runtimeArtifactCache {
		if entry == nil || entry.manifest == nil {
			continue
		}
		if strings.TrimSpace(entry.manifest.ID) == normalizedPluginID {
			delete(s.runtimeArtifactCache, artifactPath)
		}
	}
	delete(s.runtimePluginArtifactIndex, normalizedPluginID)
}

// getCachedSourceManifest returns a cached source manifest when the current
// embedded or filesystem manifest stat still matches the cached identity.
func (s *serviceImpl) getCachedSourceManifest(pluginID string, sourcePlugin any, manifestFS fs.FS) (*Manifest, bool) {
	if s == nil || manifestFS == nil || strings.TrimSpace(pluginID) == "" {
		return nil, false
	}
	info, err := fs.Stat(manifestFS, resourcefs.EmbeddedManifestPath)
	if err != nil {
		return nil, false
	}
	s.cacheMu.RLock()
	entry := s.sourceManifestCache[strings.TrimSpace(pluginID)]
	sourceKey := sourceManifestCacheSourceKey(sourcePlugin)
	if entry != nil && entry.sourceKey == sourceKey && entry.size == info.Size() && entry.modTime.Equal(info.ModTime()) {
		manifest := CloneManifest(entry.manifest)
		s.cacheMu.RUnlock()
		return manifest, true
	}
	s.cacheMu.RUnlock()
	return nil, false
}

// storeCachedSourceManifest records a detached source manifest with the current
// plugin.yaml stat identity.
func (s *serviceImpl) storeCachedSourceManifest(pluginID string, sourcePlugin any, manifestFS fs.FS, manifest *Manifest) {
	if s == nil || manifestFS == nil || manifest == nil || strings.TrimSpace(pluginID) == "" {
		return
	}
	info, err := fs.Stat(manifestFS, resourcefs.EmbeddedManifestPath)
	if err != nil {
		return
	}
	s.cacheMu.Lock()
	if s.sourceManifestCache == nil {
		s.sourceManifestCache = make(map[string]*sourceManifestCacheEntry)
	}
	s.sourceManifestCache[strings.TrimSpace(pluginID)] = &sourceManifestCacheEntry{
		pluginID:  strings.TrimSpace(pluginID),
		sourceKey: sourceManifestCacheSourceKey(sourcePlugin),
		size:      info.Size(),
		modTime:   info.ModTime(),
		manifest:  CloneManifest(manifest),
	}
	s.cacheMu.Unlock()
}

// sourceManifestCacheSourceKey distinguishes same-ID source-plugin registry
// replacements that keep the same plugin.yaml content but change callbacks or providers.
func sourceManifestCacheSourceKey(sourcePlugin any) string {
	if sourcePlugin == nil {
		return ""
	}
	value := reflect.ValueOf(sourcePlugin)
	if value.Kind() == reflect.Pointer && !value.IsNil() {
		return value.Type().String() + ":" + strconv.FormatUint(uint64(value.Pointer()), 10)
	}
	return value.Type().String()
}

// loadRuntimeManifestFromArtifactCache returns a cached dynamic manifest when
// the artifact path, size, and modification time are unchanged.
func (s *serviceImpl) loadRuntimeManifestFromArtifactCache(artifactPath string) (*Manifest, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	key, info, err := runtimeArtifactCacheKey(artifactPath)
	if err != nil {
		return nil, false, err
	}
	s.cacheMu.RLock()
	entry := s.runtimeArtifactCache[key]
	if entry != nil && entry.size == info.Size() && entry.modTime.Equal(info.ModTime()) && sameRuntimeArtifactFile(entry.fileInfo, info) {
		manifest := CloneManifest(entry.manifest)
		s.cacheMu.RUnlock()
		return manifest, true, nil
	}
	s.cacheMu.RUnlock()
	return nil, false, nil
}

// storeRuntimeManifestArtifactCache records a parsed dynamic manifest using its
// filesystem identity. The cached value is detached from the caller-owned copy.
func (s *serviceImpl) storeRuntimeManifestArtifactCache(artifactPath string, manifest *Manifest) error {
	if s == nil || manifest == nil {
		return nil
	}
	key, info, err := runtimeArtifactCacheKey(artifactPath)
	if err != nil {
		return err
	}
	cachedManifest := CloneManifest(manifest)
	s.cacheMu.Lock()
	s.runtimeArtifactCache[key] = &runtimeArtifactCacheEntry{
		path:     key,
		fileInfo: info,
		size:     info.Size(),
		modTime:  info.ModTime(),
		manifest: cachedManifest,
	}
	if cachedManifest != nil && strings.TrimSpace(cachedManifest.ID) != "" {
		s.runtimePluginArtifactIndex[cachedManifest.ID] = key
	}
	s.cacheMu.Unlock()
	return nil
}

// sameRuntimeArtifactFile reports whether two stat results still point to the
// same underlying artifact. Atomic file replacement can preserve path, size,
// and mtime on coarse filesystems, so inode/device identity keeps the stat guard
// from serving a stale desired manifest after a package overwrite.
func sameRuntimeArtifactFile(cached os.FileInfo, current os.FileInfo) bool {
	if cached == nil || current == nil {
		return false
	}
	return os.SameFile(cached, current)
}

// recordRuntimeArtifactParse increments a package-private counter used by
// cache-boundary tests to prove full artifact parsing stays bounded.
func (s *serviceImpl) recordRuntimeArtifactParse(artifactPath string) {
	if s == nil {
		return
	}
	key := filepath.Clean(strings.TrimSpace(artifactPath))
	if key == "" {
		return
	}
	s.cacheMu.Lock()
	if s.parseCounts == nil {
		s.parseCounts = make(map[string]int)
	}
	s.parseCounts[key]++
	s.cacheMu.Unlock()
}

// RuntimeArtifactParseCount returns how many times this service has fully
// parsed the given artifact path. It is intentionally unexported to keep cache
// instrumentation out of the production catalog contract.
func (s *serviceImpl) runtimeArtifactParseCount(artifactPath string) int {
	if s == nil {
		return 0
	}
	key := filepath.Clean(strings.TrimSpace(artifactPath))
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.parseCounts[key]
}

// runtimeArtifactCacheKey normalizes artifact path identity and returns its
// current filesystem metadata.
func runtimeArtifactCacheKey(artifactPath string) (string, os.FileInfo, error) {
	normalizedPath := strings.TrimSpace(artifactPath)
	if normalizedPath == "" {
		return "", nil, gerror.New("dynamic plugin artifact path cannot be empty")
	}
	absolutePath := normalizedPath
	if !filepath.IsAbs(absolutePath) {
		var err error
		absolutePath, err = filepath.Abs(absolutePath)
		if err != nil {
			return "", nil, err
		}
	}
	key := filepath.Clean(absolutePath)
	info, err := os.Stat(key)
	if err != nil {
		return "", nil, gerror.Wrapf(err, "stat dynamic plugin artifact failed: %s", key)
	}
	if info.IsDir() {
		return "", nil, gerror.Newf("dynamic plugin artifact path is a directory: %s", key)
	}
	return key, info, nil
}

// CloneManifest returns a detached copy of one plugin manifest so cached read
// models remain immutable across callers.
func CloneManifest(manifest *Manifest) *Manifest {
	if manifest == nil {
		return nil
	}
	out := *manifest
	out.SupportsMultiTenant = cloneBoolPtr(manifest.SupportsMultiTenant)
	out.I18N = cloneI18NConfig(manifest.I18N)
	out.Dependencies = plugintypes.CloneDependencySpec(manifest.Dependencies)
	out.Menus = cloneMenuSpecs(manifest.Menus)
	out.PublicAssets = ClonePublicAssetSpecs(manifest.PublicAssets)
	out.Hooks = CloneHookSpecs(manifest.Hooks)
	out.LifecycleHandlers = CloneLifecycleContracts(manifest.LifecycleHandlers)
	out.BackendResources = cloneResourceSpecMap(manifest.BackendResources)
	out.Routes = cloneRouteContracts(manifest.Routes)
	out.BridgeSpec = cloneBridgeSpec(manifest.BridgeSpec)
	out.HostCapabilities = cloneStringSet(manifest.HostCapabilities)
	out.HostServices = cloneHostServiceSpecs(manifest.HostServices)
	out.RuntimeArtifact = cloneArtifactSpec(manifest.RuntimeArtifact)
	return &out
}

// CloneManifests returns detached copies of all non-nil manifests.
func CloneManifests(items []*Manifest) []*Manifest {
	if len(items) == 0 {
		return []*Manifest{}
	}
	out := make([]*Manifest, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, CloneManifest(item))
	}
	return out
}

// cloneBoolPtr returns a detached copy of one bool pointer.
func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

// cloneI18NConfig returns a detached copy of the host-shaped i18n config.
func cloneI18NConfig(config *hostconfig.I18nConfig) *hostconfig.I18nConfig {
	if config == nil {
		return nil
	}
	out := *config
	if len(config.Locales) > 0 {
		out.Locales = append([]hostconfig.I18nLocaleConfig(nil), config.Locales...)
	}
	return &out
}

// cloneMenuSpecs returns detached menu specs including query maps.
func cloneMenuSpecs(items []*MenuSpec) []*MenuSpec {
	if len(items) == 0 {
		return []*MenuSpec{}
	}
	out := make([]*MenuSpec, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.Visible = cloneIntPtr(item.Visible)
		next.Status = cloneIntPtr(item.Status)
		next.IsFrame = cloneIntPtr(item.IsFrame)
		next.IsCache = cloneIntPtr(item.IsCache)
		if len(item.Query) > 0 {
			next.Query = make(map[string]interface{}, len(item.Query))
			for key, value := range item.Query {
				next.Query[key] = value
			}
		}
		out = append(out, &next)
	}
	return out
}

// cloneIntPtr returns a detached copy of one int pointer.
func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

// cloneResourceSpecMap returns detached backend resource specs keyed by ID.
func cloneResourceSpecMap(items map[string]*ResourceSpec) map[string]*ResourceSpec {
	if len(items) == 0 {
		return map[string]*ResourceSpec{}
	}
	out := make(map[string]*ResourceSpec, len(items))
	for key, item := range items {
		out[key] = CloneResourceSpec(item)
	}
	return out
}

// cloneRouteContracts returns detached dynamic route contracts.
func cloneRouteContracts(items []*protocol.RouteContract) []*protocol.RouteContract {
	if len(items) == 0 {
		return []*protocol.RouteContract{}
	}
	out := make([]*protocol.RouteContract, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.Tags = append([]string(nil), item.Tags...)
		if len(item.Meta) > 0 {
			next.Meta = make(map[string]string, len(item.Meta))
			for key, value := range item.Meta {
				next.Meta[key] = value
			}
		}
		out = append(out, &next)
	}
	return out
}

// cloneBridgeSpec returns a detached bridge ABI specification.
func cloneBridgeSpec(spec *protocol.BridgeSpec) *protocol.BridgeSpec {
	if spec == nil {
		return nil
	}
	out := *spec
	return &out
}

// cloneStringSet returns a detached string set.
func cloneStringSet(items map[string]struct{}) map[string]struct{} {
	if len(items) == 0 {
		return map[string]struct{}{}
	}
	out := make(map[string]struct{}, len(items))
	for key, value := range items {
		out[key] = value
	}
	return out
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

// cloneHostServiceResourceSpecs returns detached host service resource specs.
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
		if len(item.Attributes) > 0 {
			next.Attributes = make(map[string]string, len(item.Attributes))
			for key, value := range item.Attributes {
				next.Attributes[key] = value
			}
		}
		out = append(out, &next)
	}
	return out
}

// cloneArtifactSpec returns a detached runtime artifact projection.
func cloneArtifactSpec(item *ArtifactSpec) *ArtifactSpec {
	if item == nil {
		return nil
	}
	out := *item
	out.Manifest = cloneArtifactManifest(item.Manifest)
	out.FrontendAssets = cloneArtifactFrontendAssets(item.FrontendAssets)
	out.InstallSQLAssets = cloneArtifactSQLAssets(item.InstallSQLAssets)
	out.UninstallSQLAssets = cloneArtifactSQLAssets(item.UninstallSQLAssets)
	out.MockSQLAssets = cloneArtifactSQLAssets(item.MockSQLAssets)
	out.ManifestResources = cloneArtifactManifestResources(item.ManifestResources)
	out.HookSpecs = CloneHookSpecs(item.HookSpecs)
	out.LifecycleContracts = CloneLifecycleContracts(item.LifecycleContracts)
	if len(item.ResourceSpecs) > 0 {
		out.ResourceSpecs = make([]*ResourceSpec, 0, len(item.ResourceSpecs))
		for _, resource := range item.ResourceSpecs {
			out.ResourceSpecs = append(out.ResourceSpecs, CloneResourceSpec(resource))
		}
	}
	out.RouteContracts = cloneRouteContracts(item.RouteContracts)
	out.BridgeSpec = cloneBridgeSpec(item.BridgeSpec)
	out.Capabilities = append([]string(nil), item.Capabilities...)
	out.HostServices = cloneHostServiceSpecs(item.HostServices)
	return &out
}

// cloneArtifactManifest returns a detached embedded manifest identity snapshot.
func cloneArtifactManifest(item *ArtifactManifest) *ArtifactManifest {
	if item == nil {
		return nil
	}
	out := *item
	out.SupportsMultiTenant = cloneBoolPtr(item.SupportsMultiTenant)
	out.Dependencies = plugintypes.CloneDependencySpec(item.Dependencies)
	out.Menus = cloneMenuSpecs(item.Menus)
	out.PublicAssets = ClonePublicAssetSpecs(item.PublicAssets)
	return &out
}

// cloneArtifactFrontendAssets returns detached frontend asset payloads.
func cloneArtifactFrontendAssets(items []*ArtifactFrontendAsset) []*ArtifactFrontendAsset {
	if len(items) == 0 {
		return []*ArtifactFrontendAsset{}
	}
	out := make([]*ArtifactFrontendAsset, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.Content = append([]byte(nil), item.Content...)
		out = append(out, &next)
	}
	return out
}

// cloneArtifactSQLAssets returns detached SQL asset payloads.
func cloneArtifactSQLAssets(items []*ArtifactSQLAsset) []*ArtifactSQLAsset {
	if len(items) == 0 {
		return []*ArtifactSQLAsset{}
	}
	out := make([]*ArtifactSQLAsset, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		out = append(out, &next)
	}
	return out
}

// cloneArtifactManifestResources returns detached manifest resource payloads.
func cloneArtifactManifestResources(items []*ArtifactManifestResource) []*ArtifactManifestResource {
	if len(items) == 0 {
		return []*ArtifactManifestResource{}
	}
	out := make([]*ArtifactManifestResource, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		next.Content = append([]byte(nil), item.Content...)
		out = append(out, &next)
	}
	return out
}
