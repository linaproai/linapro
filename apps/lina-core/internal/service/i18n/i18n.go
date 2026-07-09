// Package i18n resolves request locales, aggregates file-backed runtime
// translation bundles, and translates dynamic host metadata for Lina core.
package i18n

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cachecoord/revisionctrl"
	"lina-core/internal/service/config"
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

// RuntimeLocalesOutput describes the runtime locale-list state exposed by the
// host while keeping the HTTP DTO projection outside the service contract.
type RuntimeLocalesOutput struct {
	Enabled bool               // Enabled reports whether runtime language switching is enabled.
	Items   []LocaleDescriptor // Items contains the locale descriptors visible to callers.
}

// RuntimeBundleRevision describes one cached runtime message representation.
type RuntimeBundleRevision struct {
	// Version is the monotonic per-locale invalidation counter.
	Version uint64
	// Fingerprint is a deterministic digest of the merged flat message catalog.
	Fingerprint string
}

// Service defines the complete i18n service contract.
type Service interface {
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

	// Translate returns one key from the current request locale only, falling
	// back to the caller-provided literal when the key is missing.
	//
	// It does not fall back to the runtime default locale. Example: with request
	// locale en-US, key "job.handler.host.cleanup.name" only present in zh-CN,
	// and fallback "Job Log Cleanup", this method returns "Job Log Cleanup".
	// Callers should pass the source text, key, or another literal fallback
	// according to their own display semantics.
	Translate(ctx context.Context, key string, fallback string) string
	// LocalizeError translates one request-scoped error into the effective locale.
	LocalizeError(ctx context.Context, err error) string

	// TranslateDynamicPluginSourceText returns one key from the current request
	// locale by reading the latest dynamic-plugin release artifact directly,
	// falling back to sourceText when the plugin, artifact, locale, or key is
	// unavailable. It does not add inactive plugin resources to the global
	// runtime bundle cache.
	TranslateDynamicPluginSourceText(ctx context.Context, pluginID string, key string, sourceText string) string

	// BundleRevision synchronizes clustered runtime i18n cache revisions and
	// returns the per-locale translation bundle revision. The fingerprint is
	// populated after the locale's merged view has been built. Coordination
	// failures are returned so HTTP handlers can fail before emitting validators
	// based on stale state.
	BundleRevision(ctx context.Context, locale string) (RuntimeBundleRevision, error)
	// RuntimeLocales returns the runtime locale-switching state and locale
	// descriptors supported by the host, localizing display names for the
	// requested display locale. It returns default-only descriptors when
	// multi-language switching is disabled.
	RuntimeLocales(ctx context.Context, locale string) RuntimeLocalesOutput
	// BuildRuntimeMessages returns the current-locale runtime translation bundle.
	//
	// The returned bundle does not merge the runtime default locale into the
	// requested locale. Example: requesting en-US will include en-US host,
	// source-plugin, and dynamic-plugin resources; if a key only exists in zh-CN,
	// it is absent from this bundle so the frontend can show its own source text
	// or key placeholder instead of silently displaying Chinese.
	BuildRuntimeMessages(ctx context.Context, locale string) map[string]interface{}

	// InvalidateRuntimeBundleCache clears the cached runtime translation bundles
	// for the given scope. Callers should pass explicit locale/sector/plugin
	// scopes for ordinary business invalidation; an empty scope drops every
	// locale and every sector and is intended for maintenance paths.
	InvalidateRuntimeBundleCache(scope InvalidateScope)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc               bizctx.Service
	configSvc               config.Service
	runtimeCacheRevisionCtl *revisionctrl.Controller
}

// New creates an i18n service from explicit runtime-owned dependencies.
func New(bizCtxSvc bizctx.Service, configSvc config.Service, cacheCoordSvc cachecoord.Service) Service {
	service := &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		configSvc: configSvc,
	}
	service.runtimeCacheRevisionCtl = revisionctrl.NewControllerWithCoordinator(
		configSvc.IsClusterEnabled(context.Background()),
		cacheCoordSvc,
		runtimeI18nCacheObservedRevision,
		func(ctx context.Context, _ int64) error {
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
var runtimeI18nCacheObservedRevision = revisionctrl.NewObservedRevision()

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
