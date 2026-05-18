// This file owns the layered runtime translation cache: per-locale, per-sector
// raw bundles, a derived merged read-view, source descriptors for diagnostics,
// and per-locale revision metadata that drives ETag invalidation.

package i18n

import (
	"sync"
	"sync/atomic"

	"lina-core/pkg/pluginhost"
)

// Sector identifies one logical contributor to the merged runtime translation
// bundle. Higher sectors override lower sectors when the merged view is built;
// the canonical priority is host < source-plugin < dynamic-plugin.
type Sector string

const (
	// SectorHost is the host-embedded `manifest/i18n/<locale>/*.json` resource.
	SectorHost Sector = "host"
	// SectorSourcePlugin is the source-plugin embedded `manifest/i18n/<locale>/*.json` resource.
	SectorSourcePlugin Sector = "source-plugin"
	// SectorDynamicPlugin is the active dynamic-plugin release `i18n_assets` custom section.
	SectorDynamicPlugin Sector = "dynamic-plugin"
)

// InvalidateScope describes which slice of the runtime cache should be cleared.
// An empty Locales slice means "every locale"; an empty Sectors slice means
// "every sector". SourcePluginID and DynamicPluginID further narrow plugin
// sectors to one owning plugin, leaving sibling plugin entries intact.
type InvalidateScope struct {
	Locales         []string
	Sectors         []Sector
	SourcePluginID  string
	DynamicPluginID string
}

// localeCache stores per-sector raw flat catalogs for one locale plus the
// derived merged view. The merged view is regenerated on demand after any
// sector mutation and serves the high-traffic `Translate*` lookup path.
type localeCache struct {
	mu sync.RWMutex

	host    map[string]string
	plugins map[string]map[string]string // sourcePluginID  -> messages
	dynamic map[string]map[string]string // dynamicPluginID -> messages

	// sourceDirty and dynamicDirty mark plugin entries that must be reloaded on
	// the next merge. A plugin can be added or upgraded after the sector was
	// already loaded without it, so deleting the old entry alone is not enough.
	sourceDirty  map[string]struct{}
	dynamicDirty map[string]struct{}

	// sources records the effective origin of every key seen during merge so
	// admin diagnostics can report where a translation actually came from.
	sources map[string]MessageSourceDescriptor

	// merged is the derived flat view; nil after invalidation, lazily rebuilt
	// on the next read. Once built it is never mutated; invalidation replaces
	// the whole map reference.
	merged map[string]string

	// fingerprint is the deterministic content digest for the merged flat view.
	// It is empty while the merged view is invalid and set when merged is built.
	fingerprint string

	// version increments on every mutation that could change the merged view.
	// Used with fingerprint to render HTTP ETags and decide 304 protocol.
	version uint64
}

// runtimeCache is the package-level cache shared by all i18n service instances.
type runtimeCache struct {
	mu      sync.RWMutex
	locales map[string]*localeCache
}

// runtimeBundleCache is the singleton runtime bundle cache.
var runtimeBundleCache = &runtimeCache{
	locales: make(map[string]*localeCache),
}

// totalInvalidationsObserved counts every successful invalidation call across
// the process lifetime. Tests can sample it to assert that scoped invalidates
// run, even when the targeted scope is already empty.
var totalInvalidationsObserved atomic.Uint64

func init() {
	// Source-plugin registry mutations affect the source-plugin sector across
	// every locale; emit a scoped invalidation rather than wiping the cache.
	pluginhost.RegisterSourcePluginRegistryListener(func() {
		runtimeBundleCache.invalidate(InvalidateScope{Sectors: []Sector{SectorSourcePlugin}})
	})
}

// InvalidateRuntimeBundleCache clears the cached runtime translation bundles
// for the given scope. An empty scope clears every locale and every sector.
func (s *serviceImpl) InvalidateRuntimeBundleCache(scope InvalidateScope) {
	runtimeBundleCache.invalidate(scope)
	// Locale descriptors are cached separately; any cache change should also
	// re-read manifest metadata on next access so resource additions become
	// discoverable without a process restart in tests and dev reload flows.
	resetRuntimeLocaleCache()
}

// BundleRevision returns the per-locale runtime translation bundle revision.
// The fingerprint is populated after the merged view has been built.
func (s *serviceImpl) BundleRevision(locale string) RuntimeBundleRevision {
	normalizedLocale := normalizeLocale(locale)
	if normalizedLocale == "" {
		return RuntimeBundleRevision{}
	}
	lc := runtimeBundleCache.lookup(normalizedLocale)
	if lc == nil {
		return RuntimeBundleRevision{}
	}
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return RuntimeBundleRevision{
		Version:     lc.version,
		Fingerprint: lc.fingerprint,
	}
}

// BundleVersion returns the per-locale runtime translation bundle version. The
// value increases monotonically across the process lifetime whenever any
// sector that contributes to the locale's merged view is invalidated.
func (s *serviceImpl) BundleVersion(locale string) uint64 {
	return s.BundleRevision(locale).Version
}

// lookup returns the cached locale entry without creating it.
func (rc *runtimeCache) lookup(locale string) *localeCache {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.locales[locale]
}

// resetRuntimeLocaleCache invalidates the locale descriptor cache so
// subsequent reads pick up newly added language resources.
func resetRuntimeLocaleCache() {
	runtimeLocaleCache.Lock()
	runtimeLocaleCache.loaded = false
	runtimeLocaleCache.locales = nil
	runtimeLocaleCache.Unlock()
}

// invalidateRuntimeBundleCache is kept as a private convenience for unit tests
// and historical no-arg listeners that target every locale and every sector.
func invalidateRuntimeBundleCache() {
	runtimeBundleCache.invalidate(InvalidateScope{})
	resetRuntimeLocaleCache()
}

// invalidate applies one scoped invalidation. A locale entry whose sectors are
// all touched rebuilds on next read while preserving its monotonic version; a
// per-sector clear keeps unaffected sectors hot so unrelated locales avoid
// cascading rebuilds.
func (rc *runtimeCache) invalidate(scope InvalidateScope) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	totalInvalidationsObserved.Add(1)
	targetLocales := rc.resolveLocaleTargets(scope.Locales)

	// An empty sector list means every sector. Keep locale entries so ETag
	// versions remain monotonic across full reloads.
	if len(scope.Sectors) == 0 {
		for _, locale := range targetLocales {
			if lc, ok := rc.locales[locale]; ok {
				lc.invalidateSectors(InvalidateScope{
					Sectors: []Sector{SectorHost, SectorSourcePlugin, SectorDynamicPlugin},
				})
			}
		}
		return
	}

	for _, locale := range targetLocales {
		lc, ok := rc.locales[locale]
		if !ok {
			continue
		}
		lc.invalidateSectors(scope)
	}
}

// resolveLocaleTargets returns the set of locale codes the invalidation should
// touch. Empty input means "every cached locale".
func (rc *runtimeCache) resolveLocaleTargets(locales []string) []string {
	if len(locales) == 0 {
		targets := make([]string, 0, len(rc.locales))
		for locale := range rc.locales {
			targets = append(targets, locale)
		}
		return targets
	}
	targets := make([]string, 0, len(locales))
	for _, locale := range locales {
		normalized := normalizeLocale(locale)
		if normalized == "" {
			continue
		}
		targets = append(targets, normalized)
	}
	return targets
}

// invalidateSectors clears one or more sectors on this locale entry. The merged
// view is dropped because any sector change can shift effective values.
func (lc *localeCache) invalidateSectors(scope InvalidateScope) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	for _, sector := range scope.Sectors {
		switch sector {
		case SectorHost:
			lc.host = nil
		case SectorSourcePlugin:
			if scope.SourcePluginID == "" {
				lc.plugins = nil
				lc.sourceDirty = nil
			} else if lc.plugins != nil {
				delete(lc.plugins, scope.SourcePluginID)
				if lc.sourceDirty == nil {
					lc.sourceDirty = make(map[string]struct{}, 1)
				}
				lc.sourceDirty[scope.SourcePluginID] = struct{}{}
			}
		case SectorDynamicPlugin:
			if scope.DynamicPluginID == "" {
				lc.dynamic = nil
				lc.dynamicDirty = nil
			} else if lc.dynamic != nil {
				delete(lc.dynamic, scope.DynamicPluginID)
				if lc.dynamicDirty == nil {
					lc.dynamicDirty = make(map[string]struct{}, 1)
				}
				lc.dynamicDirty[scope.DynamicPluginID] = struct{}{}
			}
		}
	}
	lc.merged = nil
	lc.sources = nil
	lc.fingerprint = ""
	lc.version++
}

// getOrCreate returns the locale cache entry, creating it on first access.
func (rc *runtimeCache) getOrCreate(locale string) *localeCache {
	rc.mu.RLock()
	if lc, ok := rc.locales[locale]; ok {
		rc.mu.RUnlock()
		return lc
	}
	rc.mu.RUnlock()

	rc.mu.Lock()
	defer rc.mu.Unlock()
	if lc, ok := rc.locales[locale]; ok {
		return lc
	}
	lc := &localeCache{}
	rc.locales[locale] = lc
	return lc
}

// snapshotMerged returns the current merged view if it is already built. The
// caller must treat the returned map as read-only.
func (lc *localeCache) snapshotMerged() map[string]string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.merged
}
