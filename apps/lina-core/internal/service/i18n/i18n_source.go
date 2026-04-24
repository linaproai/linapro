// This file resolves source-aware runtime i18n bundles for diagnostics and maintenance APIs.

package i18n

import (
	"context"
	"io/fs"
	"path"
	"sort"

	"lina-core/internal/packed"
	"lina-core/pkg/pluginhost"
)

// loadRawLocaleBundleWithSources loads the non-fallback bundle for one locale
// and retains per-key source descriptors.
func (s *serviceImpl) loadRawLocaleBundleWithSources(ctx context.Context, locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	bundle := make(map[string]string)
	sources := make(map[string]MessageSourceDescriptor)

	hostBundle, hostSources := loadEmbeddedHostLocaleBundleWithSources(resolvedLocale)
	mergeFlatMessageMaps(bundle, hostBundle)
	mergeMessageSources(sources, hostSources)

	pluginBundle, pluginSources := loadSourcePluginLocaleBundleWithSources(resolvedLocale)
	mergeFlatMessageMaps(bundle, pluginBundle)
	mergeMessageSources(sources, pluginSources)

	dynamicPluginBundle, dynamicPluginSources := s.loadDynamicPluginLocaleBundleWithSources(ctx, resolvedLocale)
	mergeFlatMessageMaps(bundle, dynamicPluginBundle)
	mergeMessageSources(sources, dynamicPluginSources)

	databaseBundle, databaseSources := s.loadDatabaseLocaleBundleWithSources(ctx, resolvedLocale)
	mergeFlatMessageMaps(bundle, databaseBundle)
	mergeMessageSources(sources, databaseSources)
	return bundle, sources
}

// loadEmbeddedHostLocaleBundleWithSources loads host runtime messages and their source descriptors.
func loadEmbeddedHostLocaleBundleWithSources(locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	content, err := fs.ReadFile(packed.Files, path.Join(hostI18nDir, locale+".json"))
	if err != nil {
		return map[string]string{}, map[string]MessageSourceDescriptor{}
	}
	bundle := parseLocaleJSON(content)
	sources := make(map[string]MessageSourceDescriptor, len(bundle))
	for key := range bundle {
		sources[key] = MessageSourceDescriptor{
			Type:      string(messageOriginTypeHostFile),
			ScopeType: string(messageScopeTypeHost),
			ScopeKey:  hostMessageScopeKey,
		}
	}
	return bundle, sources
}

// loadSourcePluginLocaleBundleWithSources loads source-plugin runtime messages and their source descriptors.
func loadSourcePluginLocaleBundleWithSources(locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	bundle := make(map[string]string)
	sources := make(map[string]MessageSourceDescriptor)
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return bundle, sources
	}

	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})

	relativePath := path.Join(pluginI18nDir, locale+".json")
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil || sourcePlugin.GetEmbeddedFiles() == nil {
			continue
		}
		content, err := fs.ReadFile(sourcePlugin.GetEmbeddedFiles(), relativePath)
		if err != nil || len(content) == 0 {
			continue
		}
		pluginBundle := parseLocaleJSON(content)
		for key, value := range pluginBundle {
			bundle[key] = value
			sources[key] = MessageSourceDescriptor{
				Type:      string(messageOriginTypePluginFile),
				ScopeType: string(messageScopeTypePlugin),
				ScopeKey:  sourcePlugin.ID(),
			}
		}
	}
	return bundle, sources
}

// mergeMessageSources merges source descriptors using the same overwrite semantics as bundle merging.
func mergeMessageSources(dst map[string]MessageSourceDescriptor, src map[string]MessageSourceDescriptor) {
	if len(src) == 0 {
		return
	}
	for key, value := range src {
		dst[key] = value
	}
}
