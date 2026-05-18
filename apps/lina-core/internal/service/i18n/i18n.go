// Package i18n resolves request locales, aggregates file-backed runtime
// translation bundles, and translates dynamic host metadata for Lina core.
package i18n

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/config"
	"lina-core/internal/service/pluginruntimecache"
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
	// ResolveRequestLocale resolves the effective locale for the current HTTP
	// request from query parameters, Accept-Language, and runtime defaults. It
	// never mutates request state and returns a supported locale.
	ResolveRequestLocale(r *ghttp.Request) string
	// ResolveLocale resolves one explicit locale override against runtime
	// support metadata, falling back to the current request locale or default.
	ResolveLocale(ctx context.Context, locale string) string
	// GetLocale returns the locale stored in request business context, falling
	// back to the configured default when the context is absent or unsupported.
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
	// registers sourceText "Online Session Cleanup"; another locale can translate
	// the key, while en-US may omit the key and still display the
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

// RuntimeBundleRevision describes one cached runtime message representation.
type RuntimeBundleRevision struct {
	// Version is the monotonic per-locale invalidation counter.
	Version uint64
	// Fingerprint is a deterministic digest of the merged flat message catalog.
	Fingerprint string
}

// BundleProvider defines runtime locale descriptors, runtime bundles, and bundle versioning.
type BundleProvider interface {
	// EnsureRuntimeBundleCacheFresh synchronizes clustered plugin-runtime cache
	// revisions before callers make HTTP cache decisions. Coordination failures
	// are returned so HTTP handlers can decide whether to fail or degrade.
	EnsureRuntimeBundleCacheFresh(ctx context.Context) error
	// BundleRevision returns the per-locale runtime translation bundle revision.
	// The fingerprint is populated after the locale's merged view has been
	// built, so callers can render content-sensitive HTTP validators without
	// walking the nested response payload.
	BundleRevision(locale string) RuntimeBundleRevision
	// BundleVersion returns the per-locale runtime translation bundle version.
	// It increases monotonically whenever any sector that contributes to that
	// locale's merged view is invalidated, so HTTP ETag handlers can produce
	// stable identifiers. New HTTP validators should prefer BundleRevision.
	BundleVersion(locale string) uint64
	// ListRuntimeLocales returns the runtime locales supported by the host,
	// localizing display names for the requested display locale.
	ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor
	// IsMultiLanguageEnabled reports whether the host allows runtime language
	// switching according to runtime config. Disabled mode should make callers
	// render only the default locale.
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
	// for the given scope. Callers should pass explicit locale/sector/plugin
	// scopes for ordinary business invalidation; an empty scope drops every
	// locale and every sector and is intended for maintenance paths.
	InvalidateRuntimeBundleCache(scope InvalidateScope)
	// ExportMessages exports flat runtime messages for one locale without
	// mutating runtime caches.
	ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput
	// CheckMissingMessages reports translation keys missing from one locale
	// relative to the configured default/source catalogs.
	CheckMissingMessages(ctx context.Context, locale string, keyPrefix string) []MissingMessageItem
	// DiagnoseMessages reports the effective source of runtime messages for one
	// locale, including host/source-plugin/dynamic-plugin origins.
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
	bizCtxSvc               bizctx.Service
	configSvc               config.Service
	runtimeCacheRevisionCtl *pluginruntimecache.Controller
}

// New creates an i18n service from explicit runtime-owned dependencies.
func New(bizCtxSvc bizctx.Service, configSvc config.Service, cacheCoordSvc cachecoord.Service) Service {
	service := &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		configSvc: configSvc,
	}
	service.runtimeCacheRevisionCtl = pluginruntimecache.NewControllerWithCoordinator(
		configSvc.IsClusterEnabled(context.Background()),
		cacheCoordSvc,
		runtimeI18nCacheObservedRevision,
		func(ctx context.Context) error {
			service.InvalidateRuntimeBundleCache(InvalidateScope{
				Sectors: []Sector{
					SectorSourcePlugin,
					SectorDynamicPlugin,
				},
			})
			return nil
		},
	)
	return service
}

// runtimeI18nCacheObservedRevision records the shared revision consumed by the
// runtime i18n cache domain inside this process.
var runtimeI18nCacheObservedRevision = pluginruntimecache.NewObservedRevision()

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
