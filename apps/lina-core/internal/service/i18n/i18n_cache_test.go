// This file verifies that the layered runtime cache invalidates only the
// requested locale × sector slices, leaving unrelated entries hot.

package i18n

import (
	"context"
	"testing"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/config"
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
	merged := mergeLocaleSectors(lc)
	lc.merged = merged
	lc.fingerprint = runtimeBundleFingerprint(merged)
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

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{enLocale},
		Sectors: []Sector{SectorHost},
	})

	if enCache.snapshotMerged() != nil {
		t.Fatal("expected en-US merged catalog to be invalidated after host sector clear")
	}
	if enCache.fingerprint != "" {
		t.Fatal("expected en-US fingerprint to be cleared after host sector clear")
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

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
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
	if enCache.fingerprint != "" {
		t.Fatal("expected fingerprint to be cleared when any dynamic plugin entry changes")
	}
	if enCache.version <= versionBefore {
		t.Fatalf("expected per-locale version to increment, before=%d after=%d", versionBefore, enCache.version)
	}
}

// TestInvalidateRuntimeBundleCacheSourcePluginIsPluginScoped verifies that a
// source-plugin runtime upgrade drops only the upgraded plugin's source bundle.
func TestInvalidateRuntimeBundleCacheSourcePluginIsPluginScoped(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	const targetPluginID = "source-plugin-target"
	const otherPluginID = "source-plugin-other"

	enCache := seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
		lc.plugins = map[string]map[string]string{
			targetPluginID: {"plugin." + targetPluginID + ".name": "Target Source Plugin"},
			otherPluginID:  {"plugin." + otherPluginID + ".name": "Other Source Plugin"},
		}
	})

	versionBefore := enCache.version

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Sectors:        []Sector{SectorSourcePlugin},
		SourcePluginID: targetPluginID,
	})

	if _, ok := enCache.plugins[targetPluginID]; ok {
		t.Fatalf("expected source entry for %q to be removed", targetPluginID)
	}
	if _, ok := enCache.plugins[otherPluginID]; !ok {
		t.Fatalf("expected source entry for %q to remain populated", otherPluginID)
	}
	if _, ok := enCache.sourceDirty[targetPluginID]; !ok {
		t.Fatalf("expected source entry for %q to be marked dirty", targetPluginID)
	}
	if enCache.host == nil {
		t.Fatal("expected host sector to remain populated")
	}
	if enCache.snapshotMerged() != nil {
		t.Fatal("expected merged catalog to be invalidated when any source plugin entry changes")
	}
	if enCache.fingerprint != "" {
		t.Fatal("expected fingerprint to be cleared when any source plugin entry changes")
	}
	if enCache.version <= versionBefore {
		t.Fatalf("expected per-locale version to increment, before=%d after=%d", versionBefore, enCache.version)
	}
}

// TestBundleRevisionReportsCachedFingerprint verifies that the service exposes
// the cache-owned content digest without recomputing from a nested response map.
func TestBundleRevisionReportsCachedFingerprint(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	cache := seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{
			"app.sample.metrics.total":  "Total",
			"app.sample.overview.title": "Overview",
		}
	})

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	revision, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected bundle revision without error, got %v", err)
	}
	if revision.Version != cache.version {
		t.Fatalf("expected revision version %d, got %d", cache.version, revision.Version)
	}
	if revision.Fingerprint != cache.fingerprint {
		t.Fatalf("expected revision fingerprint %q, got %q", cache.fingerprint, revision.Fingerprint)
	}
	if len(revision.Fingerprint) != runtimeBundleFingerprintHexLength {
		t.Fatalf("expected %d-char fingerprint, got %q", runtimeBundleFingerprintHexLength, revision.Fingerprint)
	}
}

// TestRuntimeBundleFingerprintIsStableAndContentSensitive verifies that flat
// catalog fingerprints ignore map iteration order and change when content does.
func TestRuntimeBundleFingerprintIsStableAndContentSensitive(t *testing.T) {
	t.Parallel()

	first := runtimeBundleFingerprint(map[string]string{
		"app.sample.metrics.total":  "Total",
		"app.sample.overview.title": "Overview",
	})
	second := runtimeBundleFingerprint(map[string]string{
		"app.sample.overview.title": "Overview",
		"app.sample.metrics.total":  "Total",
	})
	if second != first {
		t.Fatalf("expected same content to produce stable fingerprint %q, got %q", first, second)
	}

	changed := runtimeBundleFingerprint(map[string]string{
		"app.sample.metrics.total":  "Total",
		"app.sample.overview.title": "Dashboard",
	})
	if changed == first {
		t.Fatalf("expected content change to alter fingerprint %q", changed)
	}
}

// TestBundleRevisionIncrementsOnInvalidate verifies that BundleRevision reports
// monotonically increasing values whenever any sector contributing to a
// locale is invalidated, supporting the future ETag protocol.
func TestBundleRevisionIncrementsOnInvalidate(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	cache := seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
	})

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	revisionBefore, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected bundle revision before invalidation without error, got %v", err)
	}
	versionBefore := revisionBefore.Version
	if versionBefore != cache.version {
		t.Fatalf("expected BundleRevision to report cache version, got service=%d cache=%d", versionBefore, cache.version)
	}

	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{EnglishLocale},
		Sectors: []Sector{SectorHost},
	})

	revisionAfter, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected bundle revision after invalidation without error, got %v", err)
	}
	versionAfter := revisionAfter.Version
	if versionAfter <= versionBefore {
		t.Fatalf("expected BundleRevision to advance after invalidation, before=%d after=%d", versionBefore, versionAfter)
	}
}

// TestBundleRevisionIncrementsOnLocaleWideInvalidate verifies all-sector
// invalidation for a locale preserves monotonic ETag versions.
func TestBundleRevisionIncrementsOnLocaleWideInvalidate(t *testing.T) {
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

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	revisionBefore, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected bundle revision before locale-wide invalidation without error, got %v", err)
	}
	versionBefore := revisionBefore.Version

	svc.InvalidateRuntimeBundleCache(InvalidateScope{
		Locales: []string{EnglishLocale},
	})

	revisionAfter, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected bundle revision after locale-wide invalidation without error, got %v", err)
	}
	versionAfter := revisionAfter.Version
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

// TestBundleRevisionIncrementsOnFullInvalidate verifies a full cache invalidation
// clears every cached locale while keeping each cached locale version monotonic.
func TestBundleRevisionIncrementsOnFullInvalidate(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	seedLocaleCache(EnglishLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "Dashboard"}
	})
	seedLocaleCache(DefaultLocale, func(lc *localeCache) {
		lc.host = map[string]string{"menu.dashboard.title": "工作台"}
	})

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil)).(*serviceImpl)
	enRevisionBefore, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected en-US bundle revision before full invalidation without error, got %v", err)
	}
	zhRevisionBefore, err := svc.BundleRevision(context.Background(), DefaultLocale)
	if err != nil {
		t.Fatalf("expected zh-CN bundle revision before full invalidation without error, got %v", err)
	}
	enVersionBefore := enRevisionBefore.Version
	zhVersionBefore := zhRevisionBefore.Version

	svc.InvalidateRuntimeBundleCache(InvalidateScope{})

	enRevisionAfter, err := svc.BundleRevision(context.Background(), EnglishLocale)
	if err != nil {
		t.Fatalf("expected en-US bundle revision after full invalidation without error, got %v", err)
	}
	zhRevisionAfter, err := svc.BundleRevision(context.Background(), DefaultLocale)
	if err != nil {
		t.Fatalf("expected zh-CN bundle revision after full invalidation without error, got %v", err)
	}
	enVersionAfter := enRevisionAfter.Version
	zhVersionAfter := zhRevisionAfter.Version
	if enVersionAfter <= enVersionBefore {
		t.Fatalf("expected en-US version to advance after full invalidation, before=%d after=%d", enVersionBefore, enVersionAfter)
	}
	if zhVersionAfter <= zhVersionBefore {
		t.Fatalf("expected zh-CN version to advance after full invalidation, before=%d after=%d", zhVersionBefore, zhVersionAfter)
	}
}
