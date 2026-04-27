// This file verifies that the layered runtime cache invalidates only the
// requested locale × sector slices, leaving unrelated entries hot.

package i18n

import (
	"testing"
)

// seedLocaleCache directly populates one locale entry's sector maps so tests
// can assert invalidate semantics without exercising the database loader.
func seedLocaleCache(locale string, populator func(lc *localeCache)) *localeCache {
	lc := runtimeBundleCache.getOrCreate(locale)
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.host = map[string]string{}
	lc.plugins = map[string]map[string]string{}
	lc.dynamic = map[string]map[string]string{}
	lc.db = map[string]string{}
	lc.dbSources = map[string]MessageSourceDescriptor{}
	if populator != nil {
		populator(lc)
	}
	merged, sources := mergeLocaleSectors(lc, locale)
	lc.merged = merged
	lc.sources = sources
	lc.version++
	return lc
}

// TestInvalidateRuntimeBundleCacheDatabaseSectorIsLocaleScoped verifies that
// importing translations for one locale only drops the database sector for
// that locale and leaves other locales' merged views intact.
func TestInvalidateRuntimeBundleCacheDatabaseSectorIsLocaleScoped(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	enLocale := EnglishLocale
	zhLocale := DefaultLocale

	enCache := seedLocaleCache(enLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
		lc.db = map[string]string{"menu.dashboard.title": "Database EN"}
		lc.dbSources = map[string]MessageSourceDescriptor{
			"menu.dashboard.title": {Type: string(messageOriginTypeDatabase), ScopeType: "host", ScopeKey: "core"},
		}
	})
	zhCache := seedLocaleCache(zhLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "工作台"}
		lc.db = map[string]string{"menu.dashboard.title": "数据库 ZH"}
		lc.dbSources = map[string]MessageSourceDescriptor{
			"menu.dashboard.title": {Type: string(messageOriginTypeDatabase), ScopeType: "host", ScopeKey: "core"},
		}
	})

	enVersionBefore := enCache.version
	zhVersionBefore := zhCache.version

	svc := New().(*serviceImpl)
	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{enLocale},
		Sectors: []Sector{SectorDatabase},
	})

	if enCache.snapshotMerged() != nil {
		t.Fatal("expected en-US merged catalog to be invalidated after database sector clear")
	}
	if enCache.db != nil {
		t.Fatal("expected en-US database sector to be cleared")
	}
	if enCache.host == nil {
		t.Fatal("expected en-US host sector to remain populated")
	}
	if enCache.version <= enVersionBefore {
		t.Fatalf("expected en-US version to increment after invalidation, before=%d after=%d", enVersionBefore, enCache.version)
	}

	if zhCache.snapshotMerged() == nil {
		t.Fatal("expected zh-CN merged catalog to remain hot after en-US scoped invalidation")
	}
	if zhCache.db == nil {
		t.Fatal("expected zh-CN database sector to remain populated after en-US scoped invalidation")
	}
	if zhCache.version != zhVersionBefore {
		t.Fatalf("expected zh-CN version to be unchanged, before=%d after=%d", zhVersionBefore, zhCache.version)
	}
}

// TestInvalidateRuntimeBundleCacheDynamicPluginIsPluginScoped verifies that a
// single dynamic plugin lifecycle change drops only that plugin's dynamic
// entry while keeping every other plugin and every other sector hot.
func TestInvalidateRuntimeBundleCacheDynamicPluginIsPluginScoped(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	const targetPluginID = "dynamic-plugin-target"
	const otherPluginID = "dynamic-plugin-other"

	enCache := seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
		lc.dynamic = map[string]map[string]string{
			targetPluginID: {"plugin." + targetPluginID + ".name": "Target Plugin"},
			otherPluginID:  {"plugin." + otherPluginID + ".name": "Other Plugin"},
		}
	})

	versionBefore := enCache.version

	svc := New().(*serviceImpl)
	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Sectors:         []Sector{SectorDynamicPlugin},
		DynamicPluginID: targetPluginID,
	})

	if _, ok := enCache.dynamic[targetPluginID]; ok {
		t.Fatalf("expected dynamic entry for %q to be removed", targetPluginID)
	}
	if _, ok := enCache.dynamic[otherPluginID]; !ok {
		t.Fatalf("expected dynamic entry for %q to remain populated", otherPluginID)
	}
	if enCache.host == nil {
		t.Fatal("expected host sector to remain populated")
	}
	if enCache.snapshotMerged() != nil {
		t.Fatal("expected merged catalog to be invalidated when any dynamic plugin entry changes")
	}
	if enCache.version <= versionBefore {
		t.Fatalf("expected per-locale version to increment, before=%d after=%d", versionBefore, enCache.version)
	}
}

// TestInvalidateContentCacheByBusinessTypeKeepsOtherTypes verifies that
// dropping a business-type's content variants leaves unrelated business types
// in cache and avoids a database round-trip on the next read.
func TestInvalidateContentCacheByBusinessTypeKeepsOtherTypes(t *testing.T) {
	runtimeContentCache.Lock()
	runtimeContentCache.variants = map[string]map[string]ContentVariant{
		"notice\x00101\x00title": {
			EnglishLocale: {Locale: EnglishLocale, Content: "Notice EN"},
			DefaultLocale: {Locale: DefaultLocale, Content: "通知"},
		},
		"notice\x00102\x00title": {
			DefaultLocale: {Locale: DefaultLocale, Content: "通知2"},
		},
		"announcement\x00200\x00title": {
			EnglishLocale: {Locale: EnglishLocale, Content: "Announcement EN"},
		},
	}
	runtimeContentCache.Unlock()
	t.Cleanup(func() {
		invalidateContentCache(ContentInvalidateScope{})
	})

	svc := New().(*serviceImpl)
	svc.InvalidateContentCache(ContentInvalidateScope{BusinessType: "notice"})

	runtimeContentCache.RLock()
	defer runtimeContentCache.RUnlock()
	if _, ok := runtimeContentCache.variants["notice\x00101\x00title"]; ok {
		t.Fatal("expected notice anchor 101 to be dropped")
	}
	if _, ok := runtimeContentCache.variants["notice\x00102\x00title"]; ok {
		t.Fatal("expected notice anchor 102 to be dropped")
	}
	if _, ok := runtimeContentCache.variants["announcement\x00200\x00title"]; !ok {
		t.Fatal("expected announcement anchor 200 to remain cached")
	}
}

// TestInvalidateContentCacheByLocaleOnlyDropsLocaleSlice verifies that a
// locale-only scope drops the locale entry from every anchor without losing
// the other locales already present.
func TestInvalidateContentCacheByLocaleOnlyDropsLocaleSlice(t *testing.T) {
	runtimeContentCache.Lock()
	runtimeContentCache.variants = map[string]map[string]ContentVariant{
		"notice\x00101\x00title": {
			EnglishLocale: {Locale: EnglishLocale, Content: "Notice EN"},
			DefaultLocale: {Locale: DefaultLocale, Content: "通知"},
		},
		"announcement\x00200\x00title": {
			EnglishLocale: {Locale: EnglishLocale, Content: "Announcement EN"},
		},
	}
	runtimeContentCache.Unlock()
	t.Cleanup(func() {
		invalidateContentCache(ContentInvalidateScope{})
	})

	svc := New().(*serviceImpl)
	svc.InvalidateContentCache(ContentInvalidateScope{Locale: EnglishLocale})

	runtimeContentCache.RLock()
	defer runtimeContentCache.RUnlock()

	noticeVariants, ok := runtimeContentCache.variants["notice\x00101\x00title"]
	if !ok {
		t.Fatal("expected notice anchor to remain after locale-only invalidation")
	}
	if _, hit := noticeVariants[EnglishLocale]; hit {
		t.Fatal("expected notice en-US variant to be removed")
	}
	if _, hit := noticeVariants[DefaultLocale]; !hit {
		t.Fatal("expected notice zh-CN variant to remain")
	}

	if _, ok := runtimeContentCache.variants["announcement\x00200\x00title"]; ok {
		t.Fatal("expected announcement anchor with only en-US to be removed entirely")
	}
}

// TestBundleVersionIncrementsOnInvalidate verifies that BundleVersion reports
// monotonically increasing values whenever any sector contributing to a
// locale is invalidated, supporting the future ETag protocol.
func TestBundleVersionIncrementsOnInvalidate(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	cache := seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
	})

	svc := New().(*serviceImpl)
	versionBefore := svc.BundleVersion(EnglishLocale)
	if versionBefore != cache.version {
		t.Fatalf("expected BundleVersion to report cache version, got service=%d cache=%d", versionBefore, cache.version)
	}

	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{EnglishLocale},
		Sectors: []Sector{SectorHost},
	})

	versionAfter := svc.BundleVersion(EnglishLocale)
	if versionAfter <= versionBefore {
		t.Fatalf("expected BundleVersion to advance after invalidation, before=%d after=%d", versionBefore, versionAfter)
	}
}
