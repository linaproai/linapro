// This file verifies that the layered runtime cache invalidates only the
// requested locale × sector slices, leaving unrelated entries hot.

package i18n

import (
	"testing"
)

// seedLocaleCache directly populates one locale entry's sector maps so tests
// can assert invalidate semantics without exercising resource loaders.
func seedLocaleCache(locale string, populator func(lc *localeCache)) *localeCache {
	lc := runtimeBundleCache.getOrCreate(locale)
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.host = map[string]string{}
	lc.plugins = map[string]map[string]string{}
	lc.dynamic = map[string]map[string]string{}
	if populator != nil {
		populator(lc)
	}
	merged, sources := mergeLocaleSectors(lc, locale)
	lc.merged = merged
	lc.sources = sources
	lc.version++
	return lc
}

// TestInvalidateRuntimeBundleCacheHostSectorIsLocaleScoped verifies that
// clearing one locale's host sector leaves other locales' merged views intact.
func TestInvalidateRuntimeBundleCacheHostSectorIsLocaleScoped(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	enLocale := EnglishLocale
	zhLocale := DefaultLocale

	enCache := seedLocaleCache(enLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
	})
	zhCache := seedLocaleCache(zhLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "工作台"}
	})

	enVersionBefore := enCache.version
	zhVersionBefore := zhCache.version

	svc := New().(*serviceImpl)
	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{enLocale},
		Sectors: []Sector{SectorHost},
	})

	if enCache.snapshotMerged() != nil {
		t.Fatal("expected en-US merged catalog to be invalidated after host sector clear")
	}
	if enCache.host != nil {
		t.Fatal("expected en-US host sector to be cleared")
	}
	if enCache.version <= enVersionBefore {
		t.Fatalf("expected en-US version to increment after invalidation, before=%d after=%d", enVersionBefore, enCache.version)
	}

	if zhCache.snapshotMerged() == nil {
		t.Fatal("expected zh-CN merged catalog to remain hot after en-US scoped invalidation")
	}
	if zhCache.host == nil {
		t.Fatal("expected zh-CN host sector to remain populated after en-US scoped invalidation")
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
	if _, ok := enCache.dynamicDirty[targetPluginID]; !ok {
		t.Fatalf("expected dynamic entry for %q to be marked dirty", targetPluginID)
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

// TestBundleVersionIncrementsOnLocaleWideInvalidate verifies all-sector
// invalidation for a locale preserves monotonic ETag versions.
func TestBundleVersionIncrementsOnLocaleWideInvalidate(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
		lc.plugins = map[string]map[string]string{
			"plugin-a": {"plugin.plugin-a.name": "Plugin A"},
		}
		lc.dynamic = map[string]map[string]string{
			"plugin-b": {"plugin.plugin-b.name": "Plugin B"},
		}
	})

	svc := New().(*serviceImpl)
	versionBefore := svc.BundleVersion(EnglishLocale)

	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{EnglishLocale},
	})

	versionAfter := svc.BundleVersion(EnglishLocale)
	if versionAfter <= versionBefore {
		t.Fatalf("expected locale-wide invalidation to advance version, before=%d after=%d", versionBefore, versionAfter)
	}

	lc := runtimeBundleCache.getOrCreate(EnglishLocale)
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	if lc.host != nil || lc.plugins != nil || lc.dynamic != nil || lc.merged != nil {
		t.Fatalf("expected every sector and merged view to be cleared, got host=%v plugins=%v dynamic=%v merged=%v", lc.host, lc.plugins, lc.dynamic, lc.merged)
	}
}

// TestBundleVersionIncrementsOnFullInvalidate verifies a full cache invalidation
// clears every cached locale while keeping each cached locale version monotonic.
func TestBundleVersionIncrementsOnFullInvalidate(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
	})
	seedLocaleCache(DefaultLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "工作台"}
	})

	svc := New().(*serviceImpl)
	enVersionBefore := svc.BundleVersion(EnglishLocale)
	zhVersionBefore := svc.BundleVersion(DefaultLocale)

	svc.InvalidateRuntimeBundleCache(InvalidateScope{})

	enVersionAfter := svc.BundleVersion(EnglishLocale)
	zhVersionAfter := svc.BundleVersion(DefaultLocale)
	if enVersionAfter <= enVersionBefore {
		t.Fatalf("expected en-US version to advance after full invalidation, before=%d after=%d", enVersionBefore, enVersionAfter)
	}
	if zhVersionAfter <= zhVersionBefore {
		t.Fatalf("expected zh-CN version to advance after full invalidation, before=%d after=%d", zhVersionBefore, zhVersionAfter)
	}
}
