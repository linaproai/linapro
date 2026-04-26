// This file loads the dedicated API-documentation translation bundles. These
// resources are intentionally separate from runtime UI i18n bundles because
// OpenAPI documentation text is large and only needed when `/api.json` is built.

package apidoc

import (
	"context"
	"encoding/json"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"lina-core/internal/packed"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

const (
	// openAPIHostI18nDir is the host-owned apidoc translation resource path.
	openAPIHostI18nDir = "manifest/i18n/apidoc"
	// openAPIPluginI18nDir is the source-plugin apidoc translation resource path.
	openAPIPluginI18nDir = "manifest/i18n/apidoc"
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
	mergeOpenAPIMessageCatalog(catalog, loadOpenAPIWorkspacePluginBundles(ctx, normalizedLocale))
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
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return bundle
	}

	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil || sourcePlugin.GetEmbeddedFiles() == nil {
			continue
		}
		mergeOpenAPIPluginMessageCatalog(
			ctx,
			bundle,
			sourcePlugin.ID(),
			loadOpenAPIEmbeddedBundle(ctx, sourcePlugin.GetEmbeddedFiles(), openAPIPluginI18nDir, locale),
		)
	}
	return bundle
}

// loadOpenAPIEmbeddedBundle reads one locale bundle from an embedded filesystem.
func loadOpenAPIEmbeddedBundle(ctx context.Context, filesystem fs.FS, dir string, locale string) map[string]string {
	if filesystem == nil {
		return map[string]string{}
	}
	content, err := fs.ReadFile(filesystem, path.Join(dir, locale+".json"))
	if err != nil || len(content) == 0 {
		return map[string]string{}
	}

	var bundle map[string]string
	if err = json.Unmarshal(content, &bundle); err != nil {
		logger.Warningf(ctx, "parse apidoc i18n bundle failed locale=%s err=%v", locale, err)
		return map[string]string{}
	}
	return bundle
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
