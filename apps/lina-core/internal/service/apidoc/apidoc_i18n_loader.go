// This file loads the dedicated API-documentation translation bundles. These
// resources are intentionally separate from runtime UI i18n bundles because
// OpenAPI documentation text is large and only needed when `/api.json` is built.

package apidoc

import (
	"context"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/internal/packed"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

const (
	// openAPIHostI18nDir is the host-owned apidoc translation resource path.
	openAPIHostI18nDir = "manifest/i18n"
	// openAPIPluginI18nDir is the source-plugin apidoc translation resource path.
	openAPIPluginI18nDir = "manifest/i18n"
	// openAPILocaleSubdir is the locale-scoped apidoc translation resource path.
	openAPILocaleSubdir = "apidoc"
	// sourcePluginManifestPath is the embedded source-plugin manifest path.
	sourcePluginManifestPath = "plugin.yaml"
)

// openAPIMessageCache stores merged apidoc translation bundles per locale.
var openAPIMessageCache = struct {
	sync.RWMutex
	bundles map[string]map[string]string
}{
	bundles: make(map[string]map[string]string),
}

func init() {
	pluginhost.RegisterSourcePluginRegistryListener(invalidateOpenAPIMessageCache)
}

// loadOpenAPIMessageCatalog returns the merged apidoc translation catalog for
// one locale, loading host and source-plugin embedded resources lazily.
func loadOpenAPIMessageCatalog(ctx context.Context, locale string) map[string]string {
	normalizedLocale := normalizeOpenAPILocale(locale)
	openAPIMessageCache.RLock()
	cached := openAPIMessageCache.bundles[normalizedLocale]
	openAPIMessageCache.RUnlock()
	if cached != nil {
		return cloneOpenAPIMessageCatalog(cached)
	}

	bundle := make(map[string]string)
	mergeOpenAPIMessageCatalog(bundle, loadOpenAPIEmbeddedBundle(ctx, packed.Files, openAPIHostI18nDir, normalizedLocale))
	mergeOpenAPIMessageCatalog(bundle, loadOpenAPISourcePluginBundles(ctx, normalizedLocale))

	openAPIMessageCache.Lock()
	openAPIMessageCache.bundles[normalizedLocale] = cloneOpenAPIMessageCatalog(bundle)
	openAPIMessageCache.Unlock()
	return bundle
}

// loadOpenAPIMessageCatalog returns the request catalog after merging dynamic
// plugin apidoc resources that are discovered at runtime.
func (s *serviceImpl) loadOpenAPIMessageCatalog(ctx context.Context, locale string) map[string]string {
	catalog := loadOpenAPIMessageCatalog(ctx, locale)
	normalizedLocale := normalizeOpenAPILocale(locale)
	mergeOpenAPIMessageCatalog(catalog, s.loadOpenAPIDynamicPluginBundles(ctx, normalizedLocale))
	return catalog
}

// invalidateOpenAPIMessageCache clears lazily merged apidoc translation bundles
// when source plugin registrations change.
func invalidateOpenAPIMessageCache() {
	openAPIMessageCache.Lock()
	openAPIMessageCache.bundles = make(map[string]map[string]string)
	openAPIMessageCache.Unlock()
}

// normalizeOpenAPILocale normalizes empty locale inputs to the English source
// locale used by API DTO metadata.
func normalizeOpenAPILocale(locale string) string {
	trimmedLocale := strings.TrimSpace(locale)
	if trimmedLocale == "" {
		return i18nsvc.EnglishLocale
	}
	return trimmedLocale
}

// loadOpenAPISourcePluginBundles loads apidoc translations shipped by embedded
// source plugins. Each plugin may only contribute keys under its own plugin
// namespace.
func loadOpenAPISourcePluginBundles(ctx context.Context, locale string) map[string]string {
	bundle := make(map[string]string)
	pluginBundles := openAPIResourceLoader(i18nresource.ResourceLoader{
		SourcePlugins: func() []i18nresource.SourcePlugin {
			return listOpenAPII18nSourcePlugins(ctx)
		},
		Subdir:      openAPIPluginI18nDir,
		PluginScope: i18nresource.PluginScopeRestrictedToPluginNamespace,
	}).LoadSourcePluginBundles(ctx, locale)
	if len(pluginBundles) == 0 {
		return bundle
	}

	pluginIDs := make([]string, 0, len(pluginBundles))
	for pluginID := range pluginBundles {
		pluginIDs = append(pluginIDs, pluginID)
	}
	sort.Strings(pluginIDs)
	for _, pluginID := range pluginIDs {
		mergeOpenAPIMessageCatalog(bundle, pluginBundles[pluginID])
	}
	return bundle
}

// openAPII18nSourcePlugin adapts an i18n-managed source manifest to the shared
// ResourceLoader interface while keeping the governance decision manifest-driven.
type openAPII18nSourcePlugin struct {
	id    string
	files fs.FS
}

// ID returns the source plugin identifier used for namespace filtering.
func (plugin openAPII18nSourcePlugin) ID() string {
	return plugin.id
}

// GetEmbeddedFiles returns the plugin-owned embedded resource filesystem.
func (plugin openAPII18nSourcePlugin) GetEmbeddedFiles() fs.FS {
	return plugin.files
}

// listOpenAPII18nSourcePlugins adapts i18n-managed source plugin manifests to
// the shared ResourceLoader interface.
func listOpenAPII18nSourcePlugins(ctx context.Context) []i18nresource.SourcePlugin {
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return []i18nresource.SourcePlugin{}
	}
	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})

	plugins := make([]i18nresource.SourcePlugin, 0, len(sourcePlugins))
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil {
			continue
		}
		embeddedFiles := sourcePlugin.GetEmbeddedFiles()
		if embeddedFiles == nil {
			logger.Warningf(ctx, "skip source plugin apidoc i18n resources because embedded files are missing plugin=%s", sourcePlugin.ID())
			continue
		}
		manifest, err := readOpenAPISourcePluginManifest(embeddedFiles)
		if err != nil {
			logger.Warningf(ctx, "skip source plugin apidoc i18n resources because manifest cannot be read plugin=%s err=%v", sourcePlugin.ID(), err)
			continue
		}
		if manifest == nil || !manifest.i18nEnabled() {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			pluginID = strings.TrimSpace(sourcePlugin.ID())
		}
		if pluginID == "" {
			continue
		}
		plugins = append(plugins, openAPII18nSourcePlugin{
			id:    pluginID,
			files: embeddedFiles,
		})
	}
	return plugins
}

// openAPISourcePluginManifest is the minimal plugin.yaml projection needed by
// apidoc i18n loading. Full manifest validation remains owned by plugin catalog.
type openAPISourcePluginManifest struct {
	ID   string `yaml:"id"`
	I18N *struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"i18n"`
}

// i18nEnabled reports whether one source plugin participates in apidoc i18n governance.
func (manifest *openAPISourcePluginManifest) i18nEnabled() bool {
	return manifest != nil && manifest.I18N != nil && manifest.I18N.Enabled
}

// readOpenAPISourcePluginManifest reads the source-plugin manifest projection
// required for apidoc i18n filtering.
func readOpenAPISourcePluginManifest(filesystem fs.FS) (*openAPISourcePluginManifest, error) {
	if filesystem == nil {
		return nil, gerror.New("source plugin embedded files are nil")
	}
	content, err := fs.ReadFile(filesystem, sourcePluginManifestPath)
	if err != nil {
		return nil, err
	}
	manifest := &openAPISourcePluginManifest{}
	if err = yaml.Unmarshal(content, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

// loadOpenAPIEmbeddedBundle reads one locale bundle from an embedded filesystem.
func loadOpenAPIEmbeddedBundle(ctx context.Context, filesystem fs.FS, dir string, locale string) map[string]string {
	return openAPIResourceLoader(i18nresource.ResourceLoader{
		HostFS: filesystem,
		Subdir: dir,
	}).LoadHostBundle(ctx, locale)
}

// parseOpenAPIMessageCatalogJSON parses one apidoc bundle. Files may be
// maintained as nested JSON or flat dotted keys, while the service keeps a flat
// structured catalog internally.
func parseOpenAPIMessageCatalogJSON(content []byte) (map[string]string, error) {
	return i18nresource.ParseCatalog(content, i18nresource.ValueModeStringOnly)
}

// openAPIResourceLoader applies the common apidoc resource-loader defaults.
func openAPIResourceLoader(loader i18nresource.ResourceLoader) i18nresource.ResourceLoader {
	loader.LocaleSubdir = openAPILocaleSubdir
	loader.Recursive = true
	loader.ValueMode = i18nresource.ValueModeStringOnly
	loader.KeyFilter = func(key string) bool {
		return !isGeneratedEntityOpenAPIKey(key)
	}
	return loader
}

// mergeOpenAPIMessageCatalog merges source values into target values.
func mergeOpenAPIMessageCatalog(target map[string]string, source map[string]string) {
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		if isGeneratedEntityOpenAPIKey(trimmedKey) {
			continue
		}
		target[trimmedKey] = value
	}
}

// mergeOpenAPIPluginMessageCatalog merges only keys owned by the requested
// plugin namespace so plugin bundles cannot override host or sibling-plugin
// documentation strings.
func mergeOpenAPIPluginMessageCatalog(ctx context.Context, target map[string]string, pluginID string, source map[string]string) {
	prefix := "plugins." + sanitizeOpenAPIKeyPart(pluginID) + "."
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		if !strings.HasPrefix(trimmedKey, prefix) {
			logger.Warningf(ctx, "ignore apidoc i18n key outside plugin namespace plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		if isGeneratedEntityOpenAPIKey(trimmedKey) {
			logger.Warningf(ctx, "ignore generated entity apidoc i18n key plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		target[trimmedKey] = value
	}
}

// isGeneratedEntityOpenAPIKey reports whether a structured key points to
// metadata generated from internal/model/entity packages.
func isGeneratedEntityOpenAPIKey(key string) bool {
	return strings.Contains(key, ".internal.model.entity.")
}

// cloneOpenAPIMessageCatalog copies a catalog so callers may safely read it
// without sharing the cached map.
func cloneOpenAPIMessageCatalog(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
