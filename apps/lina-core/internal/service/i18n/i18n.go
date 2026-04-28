// Package i18n resolves request locales, aggregates file-backed runtime
// translation bundles, and translates dynamic host metadata for Lina core.
package i18n

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/packed"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/pluginhost"
)

const (
	// DefaultLocale is the host fallback locale used when the request does not
	// resolve to one supported language.
	DefaultLocale = "zh-CN"

	// EnglishLocale is the canonical English locale code exposed by the host.
	EnglishLocale = "en-US"

	hostI18nDir       = "manifest/i18n"
	pluginI18nDir     = "manifest/i18n"
	localeQueryKey    = "lang"
	acceptLanguageKey = "Accept-Language"
)

// LocaleDescriptor describes one locale exposed by the host runtime.
type LocaleDescriptor struct {
	Locale     string // Locale is the canonical locale code, for example zh-CN.
	Name       string // Name is the display name localized to the current request locale.
	NativeName string // NativeName is the locale's self-name rendered in its own language.
	Direction  string // Direction is currently fixed to ltr by host convention.
	IsDefault  bool   // IsDefault reports whether the locale is the host default.
}

// LocaleResolver defines request-locale resolution and request-context locale lookup.
type LocaleResolver interface {
	// ResolveRequestLocale resolves the effective locale for the current HTTP request.
	ResolveRequestLocale(r *ghttp.Request) string
	// ResolveLocale resolves one explicit locale override against the current request locale.
	ResolveLocale(ctx context.Context, locale string) string
	// GetLocale returns the locale stored in request business context.
	GetLocale(ctx context.Context) string
}

// Translator defines runtime message lookup and localized error rendering.
type Translator interface {
	// Translate returns one key from the current request locale only, falling
	// back to the caller-provided literal when the key is missing.
	//
	// It does not fall back to the runtime default locale. Example: with request
	// locale en-US, key "job.handler.host.cleanup.name" only present in zh-CN,
	// and fallback "Job Log Cleanup", this method returns "Job Log Cleanup".
	// Use this for normal UI text when showing another language would be worse
	// than showing a source/default literal.
	Translate(ctx context.Context, key string, fallback string) string
	// TranslateSourceText returns one key from the current request locale and
	// falls back to sourceText when the key is missing.
	//
	// This is a semantic wrapper for source-owned metadata whose fallback text
	// is maintained next to the source definition. Example: a built-in cron job
	// registers sourceText "Online Session Cleanup"; zh-CN can translate the
	// key to "在线会话清理", while en-US may omit the key and still display the
	// source English text. It must not return default-locale text from zh-CN
	// while the request locale is en-US.
	TranslateSourceText(ctx context.Context, key string, sourceText string) string
	// TranslateOrKey returns one key from the current request locale and falls
	// back to the key itself when the translation is missing.
	//
	// Example: with request locale en-US and missing key "menu.unknown.title",
	// this method returns "menu.unknown.title". Use this for diagnostics,
	// admin tooling, or development-time surfaces where an explicit placeholder
	// is better than hiding a missing translation.
	TranslateOrKey(ctx context.Context, key string) string
	// TranslateWithDefaultLocale returns one key from the current request locale,
	// then explicitly falls back to the runtime default locale, then to fallback.
	//
	// Example: with request locale en-US, default locale zh-CN, key present only
	// in zh-CN, and fallback "fallback", this method returns the zh-CN value.
	// Use this only for scenarios that intentionally tolerate mixed-language
	// fallback, such as maintenance diagnostics. Do not use it for ordinary UI
	// metadata where the selected language must not show another language.
	TranslateWithDefaultLocale(ctx context.Context, key string, fallback string) string
	// LocalizeError translates one request-scoped error into the effective locale.
	LocalizeError(ctx context.Context, err error) string
}

// DynamicPluginTranslator defines artifact-local translation lookup for
// dynamic-plugin release metadata that must render before the plugin is enabled.
type DynamicPluginTranslator interface {
	// TranslateDynamicPluginSourceText returns one key from the current request
	// locale by reading the latest dynamic-plugin release artifact directly,
	// falling back to sourceText when the plugin, artifact, locale, or key is
	// unavailable. It does not add inactive plugin resources to the global
	// runtime bundle cache.
	TranslateDynamicPluginSourceText(ctx context.Context, pluginID string, key string, sourceText string) string
}

// BundleProvider defines runtime locale descriptors, runtime bundles, and bundle versioning.
type BundleProvider interface {
	// BundleVersion returns the per-locale runtime translation bundle version.
	// It increases monotonically whenever any sector that contributes to that
	// locale's merged view is invalidated, so HTTP ETag handlers can produce
	// stable identifiers without recomputing the catalog.
	BundleVersion(locale string) uint64
	// ListRuntimeLocales returns the runtime locales supported by the host.
	ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor
	// IsMultiLanguageEnabled reports whether the host allows runtime language switching.
	IsMultiLanguageEnabled(ctx context.Context) bool
	// BuildRuntimeMessages returns the current-locale runtime translation bundle.
	//
	// The returned bundle does not merge the runtime default locale into the
	// requested locale. Example: requesting en-US will include en-US host,
	// source-plugin, and dynamic-plugin resources; if a key only exists in zh-CN,
	// it is absent from this bundle so the frontend can show its own source text
	// or key placeholder instead of silently displaying Chinese.
	BuildRuntimeMessages(ctx context.Context, locale string) map[string]interface{}
}

// Maintainer defines administrative i18n message maintenance and cache invalidation operations.
type Maintainer interface {
	// InvalidateRuntimeBundleCache clears the cached runtime translation bundles
	// for the given scope. An empty scope drops every locale and every sector.
	InvalidateRuntimeBundleCache(scope InvalidateScope)
	// ExportMessages exports flat runtime messages for one locale.
	ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput
	// CheckMissingMessages reports translation keys missing from one locale.
	CheckMissingMessages(ctx context.Context, locale string, keyPrefix string) []MissingMessageItem
	// DiagnoseMessages reports the effective source of runtime messages for one locale.
	DiagnoseMessages(ctx context.Context, locale string, keyPrefix string) []MessageDiagnosticItem
}

// Service defines the complete i18n service contract.
type Service interface {
	LocaleResolver
	Translator
	DynamicPluginTranslator
	BundleProvider
	Maintainer
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	configSvc config.Service
}

// New creates and returns a new i18n service instance.
func New() Service {
	return &serviceImpl{
		bizCtxSvc: bizctx.New(),
		configSvc: config.New(),
	}
}

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
		LayoutMode:    i18nresource.LayoutModeLocaleDirectory,
		ValueMode:     i18nresource.ValueModeStringifyScalars,
	}.LoadSourcePluginBundles(ctx, locale)
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
		LayoutMode:  i18nresource.LayoutModeLocaleDirectory,
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

// normalizeAcceptLanguage converts an Accept-Language header into the first valid locale tag.
func normalizeAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		languageTag := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale := normalizeLocale(languageTag); locale != "" {
			return locale
		}
	}
	return ""
}

// normalizeLocale canonicalizes one raw locale value into a stable locale code.
func normalizeLocale(value string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(value, "_", "-"))
	if normalized == "" {
		return ""
	}

	segments := strings.Split(normalized, "-")
	if len(segments) == 0 {
		return ""
	}
	for index, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" || !isAlphaNumericLocaleSegment(segment) {
			return ""
		}
		switch {
		case index == 0:
			segments[index] = strings.ToLower(segment)
		case len(segment) == 4:
			segments[index] = strings.ToUpper(segment[:1]) + strings.ToLower(segment[1:])
		case len(segment) == 2 || len(segment) == 3:
			segments[index] = strings.ToUpper(segment)
		default:
			segments[index] = strings.ToLower(segment)
		}
	}
	return strings.Join(segments, "-")
}

// buildLocaleNameKey builds the runtime translation key used for one locale display name.
func buildLocaleNameKey(locale string) string {
	return "locale." + locale + ".name"
}

// buildLocaleNativeNameKey builds the runtime translation key used for one locale native display name.
func buildLocaleNativeNameKey(locale string) string {
	return "locale." + locale + ".nativeName"
}

// isAlphaNumericLocaleSegment reports whether one locale segment contains only ASCII letters or digits.
func isAlphaNumericLocaleSegment(segment string) bool {
	for _, char := range segment {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		default:
			return false
		}
	}
	return true
}
