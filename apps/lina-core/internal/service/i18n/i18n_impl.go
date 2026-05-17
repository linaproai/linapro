// This file implements request locale resolution, hot-path translation lookup,
// and runtime message bundle assembly for the i18n service.

package i18n

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/packed"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/pluginhost"
)

// ResolveRequestLocale resolves the effective locale for the current HTTP request.
func (s *serviceImpl) ResolveRequestLocale(r *ghttp.Request) string {
	if r == nil {
		return s.getDefaultRuntimeLocale(context.Background())
	}

	ctx := r.Context()
	if rawLocale := strings.TrimSpace(r.Get(localeQueryKey).String()); rawLocale != "" {
		if locale, ok := s.lookupSupportedLocale(ctx, rawLocale); ok {
			return locale
		}
		return s.getDefaultRuntimeLocale(ctx)
	}
	if locale := s.resolveAcceptLanguageLocale(ctx, r.Header.Get(acceptLanguageKey)); locale != "" {
		return locale
	}
	return s.getDefaultRuntimeLocale(ctx)
}

// ResolveLocale resolves one explicit locale override against the current request locale.
func (s *serviceImpl) ResolveLocale(ctx context.Context, locale string) string {
	if strings.TrimSpace(locale) == "" {
		return s.GetLocale(ctx)
	}
	if normalizedLocale, ok := s.lookupSupportedLocale(ctx, locale); ok {
		return normalizedLocale
	}
	return s.getDefaultRuntimeLocale(ctx)
}

// GetLocale returns the locale stored in request business context.
func (s *serviceImpl) GetLocale(ctx context.Context) string {
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		if locale, ok := s.lookupSupportedLocale(ctx, bizCtx.Locale); ok {
			return locale
		}
	}
	return s.getDefaultRuntimeLocale(ctx)
}

// Translate returns the current-locale value or the caller fallback. For
// example, en-US missing and zh-CN present still returns fallback, not zh-CN.
func (s *serviceImpl) Translate(ctx context.Context, key string, fallback string) string {
	return s.translateForLocale(ctx, s.GetLocale(ctx), key, fallback)
}

// TranslateSourceText returns the current-locale value or source text. For
// example, a code-owned cron handler can omit en-US JSON and fall back to its
// registered English display name.
func (s *serviceImpl) TranslateSourceText(ctx context.Context, key string, sourceText string) string {
	return s.translateForLocale(ctx, s.GetLocale(ctx), key, sourceText)
}

// TranslateOrKey returns the current-locale value or the key itself. For
// example, missing key menu.unknown.title renders as menu.unknown.title.
func (s *serviceImpl) TranslateOrKey(ctx context.Context, key string) string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return ""
	}
	return s.translateForLocale(ctx, s.GetLocale(ctx), trimmedKey, trimmedKey)
}

// TranslateWithDefaultLocale returns the current-locale value, default-locale
// value, or fallback literal. For example, en-US missing and zh-CN present
// returns zh-CN; use this only when mixed-language fallback is intentional.
func (s *serviceImpl) TranslateWithDefaultLocale(ctx context.Context, key string, fallback string) string {
	return s.translateWithDefaultLocaleForLocale(ctx, s.GetLocale(ctx), key, fallback)
}

// ListRuntimeLocales returns the runtime locales supported by the host.
func (s *serviceImpl) ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor {
	displayLocale := s.ResolveLocale(ctx, locale)
	records := s.loadEnabledRuntimeLocales(ctx)
	items := make([]LocaleDescriptor, 0, len(records))
	for _, supportedLocale := range records {
		nameFallback := strings.TrimSpace(supportedLocale.Name)
		if nameFallback == "" {
			nameFallback = supportedLocale.Locale
		}
		nativeNameFallback := strings.TrimSpace(supportedLocale.NativeName)
		if nativeNameFallback == "" {
			nativeNameFallback = supportedLocale.Locale
		}
		items = append(items, LocaleDescriptor{
			Locale:     supportedLocale.Locale,
			Name:       s.translateForLocale(ctx, displayLocale, buildLocaleNameKey(supportedLocale.Locale), nameFallback),
			NativeName: s.translateForLocale(ctx, supportedLocale.Locale, buildLocaleNativeNameKey(supportedLocale.Locale), nativeNameFallback),
			Direction:  supportedLocale.Direction,
			IsDefault:  supportedLocale.IsDefault,
		})
	}
	return items
}

// BuildRuntimeMessages returns the current-locale runtime translation bundle for
// one locale. For example, en-US output omits keys that only exist in zh-CN.
func (s *serviceImpl) BuildRuntimeMessages(ctx context.Context, locale string) map[string]interface{} {
	s.ensureRuntimeBundleCacheFreshBestEffort(ctx)
	// The returned tree leaves the cache so it MUST be a clone; nesting alone
	// does not isolate frontend mutations from concurrent cache reads.
	return nestFlatMessageMap(cloneFlatMessageMap(s.snapshotMergedCatalog(ctx, locale)))
}

// snapshotMergedCatalog returns a read-only reference to the merged catalog
// for one locale. Callers MUST treat the returned map as read-only; if they
// need to mutate or persist it they must call cloneFlatMessageMap first.
func (s *serviceImpl) snapshotMergedCatalog(ctx context.Context, locale string) map[string]string {
	normalizedLocale := s.ResolveLocale(ctx, locale)
	return s.ensureMergedCatalog(ctx, normalizedLocale)
}

// translateForLocale resolves one translation key against the requested locale
// using the layered cache without cloning the merged catalog. This is the
// hot path called by every menu, dict, config, and plugin localization site.
func (s *serviceImpl) translateForLocale(ctx context.Context, locale string, key string, fallback string) string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fallback
	}
	if value, ok := s.lookupBundleKey(ctx, locale, trimmedKey); ok {
		return value
	}
	return fallback
}

// translateWithDefaultLocaleForLocale resolves one translation key and
// explicitly allows cross-language default-locale fallback. Caller-provided
// `locale` is the previously-resolved request locale; the default locale
// fallback is only consulted when the key is absent in the request locale.
func (s *serviceImpl) translateWithDefaultLocaleForLocale(ctx context.Context, locale string, key string, fallback string) string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fallback
	}
	if value, ok := s.lookupBundleKey(ctx, locale, trimmedKey); ok {
		return value
	}
	defaultLocale := s.getDefaultRuntimeLocale(ctx)
	if locale != defaultLocale {
		if value, ok := s.lookupBundleKey(ctx, defaultLocale, trimmedKey); ok {
			return value
		}
	}
	return fallback
}

// lookupBundleKey reads one translation value from the merged catalog without
// cloning. Callers MUST pass an already-normalized locale (e.g. the result of
// ResolveLocale or GetLocale); skipping a redundant resolution keeps the read
// path at one map lookup plus one read-locked map index.
func (s *serviceImpl) lookupBundleKey(ctx context.Context, locale string, key string) (string, bool) {
	merged := s.ensureMergedCatalog(ctx, locale)
	value, ok := merged[key]
	return value, ok
}

// ensureMergedCatalog returns the merged flat catalog for one locale, building
// it on demand when invalidation has dropped the cached view.
func (s *serviceImpl) ensureMergedCatalog(ctx context.Context, locale string) map[string]string {
	lc := runtimeBundleCache.getOrCreate(locale)
	if merged := lc.snapshotMerged(); merged != nil {
		return merged
	}
	return s.rebuildMergedCatalog(ctx, lc, locale)
}

// rebuildMergedCatalog reloads any missing sectors and recomputes the merged
// view under the locale entry's write lock. Callers that race here will block
// briefly, but each subsequent read enjoys a full O(1) hit.
func (s *serviceImpl) rebuildMergedCatalog(ctx context.Context, lc *localeCache, locale string) map[string]string {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.merged != nil {
		return lc.merged
	}

	if lc.host == nil {
		lc.host = loadEmbeddedHostLocaleBundle(ctx, locale)
	}
	if lc.plugins == nil {
		lc.plugins = loadSourcePluginLocaleBundles(ctx, locale)
		lc.sourceDirty = nil
	} else if len(lc.sourceDirty) > 0 {
		for pluginID := range lc.sourceDirty {
			bundle := loadSourcePluginLocaleBundle(ctx, locale, pluginID)
			if len(bundle) == 0 {
				delete(lc.plugins, pluginID)
				continue
			}
			lc.plugins[pluginID] = bundle
		}
		lc.sourceDirty = nil
	}
	if lc.dynamic == nil {
		lc.dynamic = s.loadDynamicPluginLocaleBundles(ctx, locale)
		lc.dynamicDirty = nil
	} else if len(lc.dynamicDirty) > 0 {
		for pluginID := range lc.dynamicDirty {
			bundle := s.loadDynamicPluginLocaleBundle(ctx, locale, pluginID)
			if len(bundle) == 0 {
				delete(lc.dynamic, pluginID)
				continue
			}
			lc.dynamic[pluginID] = bundle
		}
		lc.dynamicDirty = nil
	}

	merged, sources := mergeLocaleSectors(lc, locale)
	lc.merged = merged
	lc.sources = sources
	lc.fingerprint = runtimeBundleFingerprint(merged)
	lc.version++
	return merged
}

// mergeLocaleSectors composes the merged catalog and source descriptor map for
// one locale entry. Higher-priority sectors overwrite lower ones; per-key
// origin is recorded for diagnostics.
func mergeLocaleSectors(lc *localeCache, locale string) (map[string]string, map[string]MessageSourceDescriptor) {
	merged := make(map[string]string, len(lc.host))
	sources := make(map[string]MessageSourceDescriptor, len(lc.host))

	for key, value := range lc.host {
		merged[key] = value
		sources[key] = MessageSourceDescriptor{
			Type:      string(messageOriginTypeHostFile),
			ScopeType: string(messageScopeTypeHost),
			ScopeKey:  hostMessageScopeKey,
		}
	}

	pluginIDs := make([]string, 0, len(lc.plugins))
	for pluginID := range lc.plugins {
		pluginIDs = append(pluginIDs, pluginID)
	}
	sort.Strings(pluginIDs)
	for _, pluginID := range pluginIDs {
		for key, value := range lc.plugins[pluginID] {
			merged[key] = value
			sources[key] = MessageSourceDescriptor{
				Type:      string(messageOriginTypePluginFile),
				ScopeType: string(messageScopeTypePlugin),
				ScopeKey:  pluginID,
			}
		}
	}

	dynamicIDs := make([]string, 0, len(lc.dynamic))
	for pluginID := range lc.dynamic {
		dynamicIDs = append(dynamicIDs, pluginID)
	}
	sort.Strings(dynamicIDs)
	for _, pluginID := range dynamicIDs {
		for key, value := range lc.dynamic[pluginID] {
			merged[key] = value
			sources[key] = MessageSourceDescriptor{
				Type:      string(messageOriginTypePluginFile),
				ScopeType: string(messageScopeTypePlugin),
				ScopeKey:  pluginID,
			}
		}
	}

	return merged, sources
}

// loadSourcePluginLocaleBundles loads source-plugin translation resources from
// registered embedded plugin filesystems, returning a per-plugin map so the
// cache can attribute each key to its owning plugin.
func loadSourcePluginLocaleBundles(ctx context.Context, locale string) map[string]map[string]string {
	return i18nresource.ResourceLoader{
		SourcePlugins: listRuntimeI18nSourcePlugins,
		Subdir:        pluginI18nDir,
		PluginScope:   i18nresource.PluginScopeOpen,
		ValueMode:     i18nresource.ValueModeStringifyScalars,
	}.LoadSourcePluginBundles(ctx, locale)
}

// loadSourcePluginLocaleBundle loads one source-plugin translation bundle from
// the registered embedded plugin filesystem.
func loadSourcePluginLocaleBundle(ctx context.Context, locale string, pluginID string) map[string]string {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return map[string]string{}
	}
	sourcePlugin, ok := pluginhost.GetSourcePlugin(normalizedPluginID)
	if !ok || sourcePlugin == nil {
		return map[string]string{}
	}
	bundles := i18nresource.ResourceLoader{
		SourcePlugins: func() []i18nresource.SourcePlugin {
			return []i18nresource.SourcePlugin{sourcePlugin}
		},
		Subdir:      pluginI18nDir,
		PluginScope: i18nresource.PluginScopeOpen,
		ValueMode:   i18nresource.ValueModeStringifyScalars,
	}.LoadSourcePluginBundles(ctx, locale)
	return bundles[normalizedPluginID]
}

// listRuntimeI18nSourcePlugins adapts pluginhost definitions to the shared
// ResourceLoader interface without coupling the loader package to pluginhost.
func listRuntimeI18nSourcePlugins() []i18nresource.SourcePlugin {
	sourcePlugins := pluginhost.ListSourcePlugins()
	plugins := make([]i18nresource.SourcePlugin, 0, len(sourcePlugins))
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil {
			continue
		}
		plugins = append(plugins, sourcePlugin)
	}
	return plugins
}

// loadEmbeddedHostLocaleBundle loads host runtime messages from embedded manifest assets.
func loadEmbeddedHostLocaleBundle(ctx context.Context, locale string) map[string]string {
	return i18nresource.ResourceLoader{
		HostFS:      packed.Files,
		Subdir:      hostI18nDir,
		PluginScope: i18nresource.PluginScopeOpen,
		ValueMode:   i18nresource.ValueModeStringifyScalars,
	}.LoadHostBundle(ctx, locale)
}

// parseLocaleJSON unmarshals one locale JSON file into a flat message catalog.
// Flat keys override equivalent nested paths, which keeps mixed-format locale
// files deterministic during gradual authoring-format migrations.
func parseLocaleJSON(content []byte) map[string]string {
	bundle, err := i18nresource.ParseCatalog(content, i18nresource.ValueModeStringifyScalars)
	if err != nil {
		return map[string]string{}
	}
	return bundle
}

// lookupMessageString retrieves one string message by dotted key path.
func lookupMessageString(messages map[string]interface{}, key string) (string, bool) {
	if len(messages) == 0 {
		return "", false
	}

	current := interface{}(messages)
	for _, segment := range strings.Split(strings.TrimSpace(key), ".") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return "", false
		}
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		next, ok := currentMap[segment]
		if !ok {
			return "", false
		}
		current = next
	}
	value, ok := current.(string)
	return value, ok
}

// cloneFlatMessageMap clones one flat message map so callers can safely mutate it.
func cloneFlatMessageMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// nestFlatMessageMap converts one flat catalog into the nested object tree expected by the frontend runtime i18n loader.
func nestFlatMessageMap(src map[string]string) map[string]interface{} {
	if len(src) == 0 {
		return map[string]interface{}{}
	}

	keys := make([]string, 0, len(src))
	for key := range src {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	output := make(map[string]interface{})
	for _, key := range keys {
		setNestedMessageValue(output, key, src[key])
	}
	return output
}

// setNestedMessageValue writes one dotted key into the nested runtime message object.
func setNestedMessageValue(output map[string]interface{}, key string, value string) {
	segments := strings.Split(strings.TrimSpace(key), ".")
	if len(segments) == 0 {
		return
	}

	current := output
	for index, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return
		}
		if index == len(segments)-1 {
			current[segment] = value
			return
		}

		next, ok := current[segment].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			current[segment] = next
		}
		current = next
	}
}
