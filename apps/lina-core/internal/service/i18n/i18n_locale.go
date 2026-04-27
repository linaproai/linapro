// This file loads enabled runtime locales from manifest resources and resolves
// supported locale codes for requests.

package i18n

import (
	"context"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"lina-core/internal/packed"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

// runtimeLocaleCache stores enabled runtime locale descriptors discovered from
// manifest resources and the default config i18n metadata.
var runtimeLocaleCache = struct {
	sync.RWMutex
	loaded  bool
	locales []LocaleDescriptor
}{}

// invalidateRuntimeLocaleCache clears the cached locale descriptors. It is
// used by tests and development reload flows that change manifest metadata.
func invalidateRuntimeLocaleCache() {
	runtimeLocaleCache.Lock()
	defer runtimeLocaleCache.Unlock()
	runtimeLocaleCache.loaded = false
	runtimeLocaleCache.locales = nil
}

// loadEnabledRuntimeLocales returns the enabled runtime locale descriptors,
// discovering built-in host locales from manifest resources.
func (s *serviceImpl) loadEnabledRuntimeLocales(ctx context.Context) []LocaleDescriptor {
	runtimeLocaleCache.RLock()
	if runtimeLocaleCache.loaded {
		cachedLocales := cloneLocaleDescriptors(runtimeLocaleCache.locales)
		runtimeLocaleCache.RUnlock()
		return cachedLocales
	}
	runtimeLocaleCache.RUnlock()

	config := s.loadRuntimeI18nConfig(ctx)
	records := s.loadConfiguredRuntimeLocales(ctx, config)
	if len(records) == 0 {
		records = fallbackRuntimeLocales(config)
	}
	records = normalizeRuntimeLocales(records, config.Default)

	runtimeLocaleCache.Lock()
	runtimeLocaleCache.loaded = true
	runtimeLocaleCache.locales = cloneLocaleDescriptors(records)
	runtimeLocaleCache.Unlock()
	return cloneLocaleDescriptors(records)
}

// loadConfiguredRuntimeLocales returns file-backed runtime locales discovered
// from host i18n JSON files, with metadata from the default config i18n section.
func (s *serviceImpl) loadConfiguredRuntimeLocales(ctx context.Context, config *hostconfig.I18nConfig) []LocaleDescriptor {
	discoveredLocales := discoverHostRuntimeLocaleFiles(ctx)
	if len(discoveredLocales) == 0 {
		configuredLocales := buildRuntimeLocalesFromConfig(config)
		if len(configuredLocales) > 0 {
			return configuredLocales
		}
		return fallbackRuntimeLocales(config)
	}
	return buildConfiguredRuntimeLocales(discoveredLocales, config)
}

// discoverHostRuntimeLocaleFiles lists host manifest/i18n/*.json locale files.
func discoverHostRuntimeLocaleFiles(ctx context.Context) []string {
	entries, err := fs.ReadDir(packed.Files, hostI18nDir)
	if err != nil {
		logger.Warningf(ctx, "scan host i18n locale resources failed dir=%s err=%v", hostI18nDir, err)
		return []string{}
	}

	locales := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		locale := normalizeLocale(strings.TrimSuffix(name, ".json"))
		if locale == "" {
			continue
		}
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	return locales
}

// loadRuntimeI18nConfig loads runtime locale metadata from the shared config service.
func (s *serviceImpl) loadRuntimeI18nConfig(ctx context.Context) *hostconfig.I18nConfig {
	if s == nil || s.configSvc == nil {
		return hostconfig.New().GetI18n(ctx)
	}
	cfg := s.configSvc.GetI18n(ctx)
	if cfg == nil {
		return hostconfig.New().GetI18n(ctx)
	}
	return cfg
}

// buildRuntimeLocalesFromConfig builds locale descriptors from config metadata
// when runtime JSON resources cannot be discovered. This keeps the fallback
// list configuration-driven instead of embedding every supported language in Go.
func buildRuntimeLocalesFromConfig(config *hostconfig.I18nConfig) []LocaleDescriptor {
	if config == nil {
		return []LocaleDescriptor{}
	}

	orderedLocales := make([]LocaleDescriptor, 0, len(config.Locales))
	seenLocales := make(map[string]struct{}, len(config.Locales))
	defaultLocale := normalizeLocale(config.Default)

	for _, item := range config.Locales {
		locale := normalizeLocale(item.Locale)
		if locale == "" {
			continue
		}
		if _, ok := seenLocales[locale]; ok {
			continue
		}
		seenLocales[locale] = struct{}{}
		orderedLocales = append(orderedLocales, LocaleDescriptor{
			Locale:     locale,
			NativeName: strings.TrimSpace(item.NativeName),
			Direction:  LocaleDirectionLTR.String(),
			IsDefault:  defaultLocale != "" && locale == defaultLocale,
		})
	}

	if len(orderedLocales) == 0 && defaultLocale != "" {
		orderedLocales = append(orderedLocales, LocaleDescriptor{
			Locale:    defaultLocale,
			Direction: LocaleDirectionLTR.String(),
			IsDefault: true,
		})
	}
	if defaultLocale != "" {
		orderedLocales = ensureDefaultRuntimeLocaleDescriptor(orderedLocales, defaultLocale, hostconfig.I18nLocaleConfig{})
	}
	return filterRuntimeLocalesByI18nEnabled(orderedLocales, config.Enabled, defaultLocale)
}

// buildConfiguredRuntimeLocales applies config metadata and ordering to
// the discovered file-backed locale list.
func buildConfiguredRuntimeLocales(discoveredLocales []string, config *hostconfig.I18nConfig) []LocaleDescriptor {
	if config == nil {
		return []LocaleDescriptor{}
	}

	discoveredSet := make(map[string]struct{}, len(discoveredLocales))
	for _, locale := range discoveredLocales {
		discoveredSet[locale] = struct{}{}
	}

	metadataByLocale := make(map[string]hostconfig.I18nLocaleConfig, len(config.Locales))
	orderedLocales := make([]string, 0, len(config.Locales))
	seenOrderedLocales := make(map[string]struct{}, len(discoveredLocales))
	for _, item := range config.Locales {
		locale := normalizeLocale(item.Locale)
		if locale == "" {
			continue
		}
		item.Locale = locale
		metadataByLocale[locale] = item
		if _, ok := discoveredSet[locale]; !ok {
			continue
		}
		if _, ok := seenOrderedLocales[locale]; ok {
			continue
		}
		seenOrderedLocales[locale] = struct{}{}
		orderedLocales = append(orderedLocales, locale)
	}

	defaultLocale := normalizeLocale(config.Default)

	descriptors := make([]LocaleDescriptor, 0, len(orderedLocales))
	for _, locale := range orderedLocales {
		metadata := metadataByLocale[locale]
		descriptors = append(descriptors, LocaleDescriptor{
			Locale:     locale,
			NativeName: strings.TrimSpace(metadata.NativeName),
			Direction:  LocaleDirectionLTR.String(),
			IsDefault:  defaultLocale != "" && locale == defaultLocale,
		})
	}
	if defaultLocale != "" {
		descriptors = ensureDefaultRuntimeLocaleDescriptor(descriptors, defaultLocale, metadataByLocale[defaultLocale])
	}
	return filterRuntimeLocalesByI18nEnabled(descriptors, config.Enabled, defaultLocale)
}

// ensureDefaultRuntimeLocaleDescriptor keeps the configured default language
// available even when users remove it from the selectable locale list.
func ensureDefaultRuntimeLocaleDescriptor(descriptors []LocaleDescriptor, defaultLocale string, metadata hostconfig.I18nLocaleConfig) []LocaleDescriptor {
	foundDefault := false
	for index := range descriptors {
		if descriptors[index].Locale == defaultLocale {
			descriptors[index].IsDefault = true
			foundDefault = true
			continue
		}
		descriptors[index].IsDefault = false
	}
	if foundDefault {
		return descriptors
	}
	descriptors = append([]LocaleDescriptor{
		{
			Locale:     defaultLocale,
			NativeName: strings.TrimSpace(metadata.NativeName),
			Direction:  LocaleDirectionLTR.String(),
			IsDefault:  true,
		},
	}, descriptors...)
	for index := 1; index < len(descriptors); index++ {
		descriptors[index].IsDefault = false
	}
	return descriptors
}

// filterRuntimeLocalesByI18nEnabled returns only the default locale when
// multi-language switching is disabled in config.yaml.
func filterRuntimeLocalesByI18nEnabled(locales []LocaleDescriptor, enabled bool, defaultLocale string) []LocaleDescriptor {
	if enabled {
		return locales
	}
	normalizedDefaultLocale := normalizeLocale(defaultLocale)
	if normalizedDefaultLocale != "" {
		for _, locale := range locales {
			if locale.Locale == normalizedDefaultLocale {
				locale.IsDefault = true
				return []LocaleDescriptor{locale}
			}
		}
		return fallbackRuntimeLocales(&hostconfig.I18nConfig{Default: normalizedDefaultLocale})
	}
	for _, locale := range locales {
		if locale.IsDefault {
			return []LocaleDescriptor{locale}
		}
	}
	return []LocaleDescriptor{}
}

// IsMultiLanguageEnabled reports whether config.yaml enables runtime language switching.
func (s *serviceImpl) IsMultiLanguageEnabled(ctx context.Context) bool {
	return s.loadRuntimeI18nConfig(ctx).Enabled
}

// getDefaultRuntimeLocale returns the default runtime locale from the enabled
// locale descriptors, falling back to the configured i18n.default value.
func (s *serviceImpl) getDefaultRuntimeLocale(ctx context.Context) string {
	for _, locale := range s.loadEnabledRuntimeLocales(ctx) {
		if locale.IsDefault {
			return locale.Locale
		}
	}
	return normalizeLocale(s.loadRuntimeI18nConfig(ctx).Default)
}

// lookupSupportedLocale resolves one raw locale string against the enabled
// runtime locale descriptors. The hot path holds only a read lock and avoids
// cloning the descriptor slice; cache misses fall back to the public loader
// which reads locale metadata from manifest resources and config.yaml.
func (s *serviceImpl) lookupSupportedLocale(ctx context.Context, rawLocale string) (string, bool) {
	normalizedLocale := normalizeLocale(rawLocale)
	if normalizedLocale == "" {
		return "", false
	}
	if locale, hit := lookupCachedSupportedLocale(normalizedLocale); hit {
		return locale, true
	}
	for _, locale := range s.loadEnabledRuntimeLocales(ctx) {
		if strings.EqualFold(locale.Locale, normalizedLocale) {
			return locale.Locale, true
		}
	}
	return "", false
}

// lookupCachedSupportedLocale performs a read-only locale registry lookup
// without cloning. Returns (canonical locale, true) only when the cache is
// already loaded and the locale exists; otherwise the caller must fall back to
// the manifest-backed loader. Used by the Translate hot path where every
// avoided allocation matters.
func lookupCachedSupportedLocale(normalizedLocale string) (string, bool) {
	runtimeLocaleCache.RLock()
	defer runtimeLocaleCache.RUnlock()
	if !runtimeLocaleCache.loaded {
		return "", false
	}
	for _, locale := range runtimeLocaleCache.locales {
		if strings.EqualFold(locale.Locale, normalizedLocale) {
			return locale.Locale, true
		}
	}
	return "", false
}

// resolveAcceptLanguageLocale returns the first supported locale discovered in
// one Accept-Language header.
func (s *serviceImpl) resolveAcceptLanguageLocale(ctx context.Context, header string) string {
	for _, part := range strings.Split(header, ",") {
		languageTag := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale, ok := s.lookupSupportedLocale(ctx, languageTag); ok {
			return locale
		}
	}
	return ""
}

// fallbackRuntimeLocales returns a minimal configured default-locale list used
// only when embedded manifest resources and locale metadata cannot be read.
func fallbackRuntimeLocales(config *hostconfig.I18nConfig) []LocaleDescriptor {
	if config == nil {
		return []LocaleDescriptor{}
	}
	defaultLocale := normalizeLocale(config.Default)
	if defaultLocale == "" {
		return []LocaleDescriptor{}
	}
	return []LocaleDescriptor{
		{
			Locale:    defaultLocale,
			Direction: LocaleDirectionLTR.String(),
			IsDefault: true,
		},
	}
}

// normalizeRuntimeLocales ensures the runtime locale list always contains a
// single default locale and no duplicate locale codes.
func normalizeRuntimeLocales(locales []LocaleDescriptor, defaultLocale string) []LocaleDescriptor {
	normalizedDefaultLocale := normalizeLocale(defaultLocale)
	if len(locales) == 0 {
		return fallbackRuntimeLocales(&hostconfig.I18nConfig{Default: normalizedDefaultLocale})
	}

	items := make([]LocaleDescriptor, 0, len(locales))
	seenLocales := make(map[string]struct{}, len(locales))
	hasDefault := false
	for _, locale := range locales {
		normalizedLocale := normalizeLocale(locale.Locale)
		if normalizedLocale == "" {
			continue
		}
		if _, ok := seenLocales[normalizedLocale]; ok {
			continue
		}
		seenLocales[normalizedLocale] = struct{}{}
		locale.Locale = normalizedLocale
		locale.Direction = LocaleDirectionLTR.String()
		if normalizedDefaultLocale != "" {
			locale.IsDefault = normalizedLocale == normalizedDefaultLocale
		}
		if locale.IsDefault && !hasDefault {
			hasDefault = true
		} else {
			locale.IsDefault = false
		}
		items = append(items, locale)
	}

	if len(items) == 0 {
		return fallbackRuntimeLocales(&hostconfig.I18nConfig{Default: normalizedDefaultLocale})
	}
	if hasDefault {
		return items
	}
	if normalizedDefaultLocale != "" {
		return ensureDefaultRuntimeLocaleDescriptor(items, normalizedDefaultLocale, hostconfig.I18nLocaleConfig{})
	}
	items[0].IsDefault = true
	return items
}

// cloneLocaleDescriptors copies locale descriptors so callers may mutate them safely.
func cloneLocaleDescriptors(src []LocaleDescriptor) []LocaleDescriptor {
	if len(src) == 0 {
		return []LocaleDescriptor{}
	}
	dst := make([]LocaleDescriptor, len(src))
	copy(dst, src)
	return dst
}
