// This file loads enabled runtime locales from the database and resolves
// supported locale codes for request processing.

package i18n

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
)

// runtimeLocaleCache stores enabled runtime locale descriptors discovered from
// the locale registry table.
var runtimeLocaleCache = struct {
	sync.RWMutex
	loaded  bool
	locales []LocaleDescriptor
}{}

// loadEnabledRuntimeLocales returns the enabled runtime locale descriptors,
// preferring the database registry and falling back to built-in host locales.
func (s *serviceImpl) loadEnabledRuntimeLocales(ctx context.Context) []LocaleDescriptor {
	runtimeLocaleCache.RLock()
	if runtimeLocaleCache.loaded {
		cachedLocales := cloneLocaleDescriptors(runtimeLocaleCache.locales)
		runtimeLocaleCache.RUnlock()
		return cachedLocales
	}
	runtimeLocaleCache.RUnlock()

	records, err := s.queryEnabledRuntimeLocales(ctx)
	if err != nil {
		logger.Warningf(ctx, "load enabled runtime locales fallback to built-ins: %v", err)
		records = builtinRuntimeLocales()
	}
	if len(records) == 0 {
		records = builtinRuntimeLocales()
	}
	records = normalizeRuntimeLocales(records)

	runtimeLocaleCache.Lock()
	runtimeLocaleCache.loaded = true
	runtimeLocaleCache.locales = cloneLocaleDescriptors(records)
	runtimeLocaleCache.Unlock()
	return cloneLocaleDescriptors(records)
}

// queryEnabledRuntimeLocales loads enabled locale rows from sys_i18n_locale.
func (s *serviceImpl) queryEnabledRuntimeLocales(ctx context.Context) (items []LocaleDescriptor, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("query enabled runtime locales panic: %v", recovered)
		}
	}()

	var rows []*entity.SysI18NLocale
	err = dao.SysI18NLocale.Ctx(ctx).
		Where(do.SysI18NLocale{Status: int(localeStatusEnabled)}).
		OrderAsc(dao.SysI18NLocale.Columns().Sort).
		OrderAsc(dao.SysI18NLocale.Columns().Locale).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	items = make([]LocaleDescriptor, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		normalizedLocale := normalizeLocale(row.Locale)
		if normalizedLocale == "" {
			continue
		}
		items = append(items, LocaleDescriptor{
			Locale:     normalizedLocale,
			Name:       strings.TrimSpace(row.Name),
			NativeName: strings.TrimSpace(row.NativeName),
			IsDefault:  row.IsDefault == int(localeDefaultYes),
		})
	}
	return items, nil
}

// getDefaultRuntimeLocale returns the default runtime locale from the enabled
// locale registry, falling back to the built-in host default when needed.
func (s *serviceImpl) getDefaultRuntimeLocale(ctx context.Context) string {
	for _, locale := range s.loadEnabledRuntimeLocales(ctx) {
		if locale.IsDefault {
			return locale.Locale
		}
	}
	return DefaultLocale
}

// lookupSupportedLocale resolves one raw locale string against the enabled
// runtime locale registry. The hot path holds only a read lock and avoids
// cloning the descriptor slice; cache misses fall back to the public loader
// which performs the database query.
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
// the database-backed loader. Used by the Translate hot path where every
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

// builtinRuntimeLocales returns the host-shipped runtime locales used as a
// fallback when the locale registry is unavailable.
func builtinRuntimeLocales() []LocaleDescriptor {
	return []LocaleDescriptor{
		{
			Locale:     DefaultLocale,
			Name:       "简体中文",
			NativeName: "简体中文",
			IsDefault:  true,
		},
		{
			Locale:     EnglishLocale,
			Name:       "英语",
			NativeName: "English",
			IsDefault:  false,
		},
	}
}

// normalizeRuntimeLocales ensures the runtime locale list always contains a
// single default locale and no duplicate locale codes.
func normalizeRuntimeLocales(locales []LocaleDescriptor) []LocaleDescriptor {
	if len(locales) == 0 {
		return builtinRuntimeLocales()
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
		if locale.IsDefault && !hasDefault {
			hasDefault = true
		} else {
			locale.IsDefault = false
		}
		items = append(items, locale)
	}

	if len(items) == 0 {
		return builtinRuntimeLocales()
	}
	if hasDefault {
		return items
	}
	for index := range items {
		if items[index].Locale == DefaultLocale {
			items[index].IsDefault = true
			return items
		}
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
