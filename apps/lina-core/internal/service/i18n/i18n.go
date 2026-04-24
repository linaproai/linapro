// Package i18n resolves request locales, aggregates runtime translation bundles,
// and translates dynamic host metadata for the Lina core service.
package i18n

import (
	"context"
	"encoding/json"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/packed"
	"lina-core/internal/service/bizctx"
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
	IsDefault  bool   // IsDefault reports whether the locale is the host default.
}

// Service defines the i18n service contract.
type Service interface {
	// ResolveRequestLocale resolves the effective locale for the current HTTP request.
	ResolveRequestLocale(r *ghttp.Request) string
	// ResolveLocale resolves one explicit locale override against the current request locale.
	ResolveLocale(ctx context.Context, locale string) string
	// GetLocale returns the locale stored in request business context.
	GetLocale(ctx context.Context) string
	// Translate returns the localized value for one translation key.
	Translate(ctx context.Context, key string, fallback string) string
	// LocalizeError translates one request-scoped error into the effective locale.
	LocalizeError(ctx context.Context, err error) string
	// InvalidateRuntimeBundleCache clears the cached runtime translation bundles.
	InvalidateRuntimeBundleCache()
	// ListRuntimeLocales returns the runtime locales supported by the host.
	ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor
	// BuildRuntimeMessages returns the effective runtime translation bundle for one locale.
	BuildRuntimeMessages(ctx context.Context, locale string) map[string]interface{}
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
}

// runtimeBundleCache stores flat per-locale message bundles discovered from host
// and source-plugin resources.
var runtimeBundleCache = struct {
	sync.RWMutex
	bundles map[string]map[string]string
}{
	bundles: make(map[string]map[string]string),
}

// supportedRuntimeLocales declares the runtime locales currently shipped by the host.
var supportedRuntimeLocales = []string{
	DefaultLocale,
	EnglishLocale,
}

// New creates and returns a new i18n service instance.
func New() Service {
	return &serviceImpl{
		bizCtxSvc: bizctx.New(),
	}
}

// ResolveRequestLocale resolves the effective locale for the current HTTP request.
func (s *serviceImpl) ResolveRequestLocale(r *ghttp.Request) string {
	if r == nil {
		return DefaultLocale
	}

	if locale := normalizeLocale(r.Get(localeQueryKey).String()); locale != "" {
		return locale
	}
	if locale := normalizeAcceptLanguage(r.Header.Get(acceptLanguageKey)); locale != "" {
		return locale
	}
	return DefaultLocale
}

// ResolveLocale resolves one explicit locale override against the current request locale.
func (s *serviceImpl) ResolveLocale(ctx context.Context, locale string) string {
	if normalizedLocale := normalizeLocale(locale); normalizedLocale != "" {
		return normalizedLocale
	}
	return s.GetLocale(ctx)
}

// GetLocale returns the locale stored in request business context.
func (s *serviceImpl) GetLocale(ctx context.Context) string {
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		if locale := normalizeLocale(bizCtx.Locale); locale != "" {
			return locale
		}
	}
	return DefaultLocale
}

// Translate returns the localized value for one translation key.
func (s *serviceImpl) Translate(ctx context.Context, key string, fallback string) string {
	return s.translateForLocale(ctx, s.GetLocale(ctx), key, fallback)
}

// ListRuntimeLocales returns the runtime locales supported by the host.
func (s *serviceImpl) ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor {
	displayLocale := s.ResolveLocale(ctx, locale)
	items := make([]LocaleDescriptor, 0, len(supportedRuntimeLocales))
	for _, supportedLocale := range supportedRuntimeLocales {
		items = append(items, LocaleDescriptor{
			Locale:     supportedLocale,
			Name:       s.translateForLocale(ctx, displayLocale, buildLocaleNameKey(supportedLocale), supportedLocale),
			NativeName: s.translateForLocale(ctx, supportedLocale, buildLocaleNativeNameKey(supportedLocale), supportedLocale),
			IsDefault:  supportedLocale == DefaultLocale,
		})
	}
	return items
}

// BuildRuntimeMessages returns the effective runtime translation bundle for one locale.
func (s *serviceImpl) BuildRuntimeMessages(ctx context.Context, locale string) map[string]interface{} {
	return nestFlatMessageMap(s.buildRuntimeMessageCatalog(ctx, locale))
}

// buildRuntimeMessageCatalog returns the effective flat translation catalog for one locale.
func (s *serviceImpl) buildRuntimeMessageCatalog(ctx context.Context, locale string) map[string]string {
	normalizedLocale := normalizeLocale(locale)
	if normalizedLocale == "" {
		normalizedLocale = DefaultLocale
	}

	defaultBundle := s.loadRawLocaleBundle(ctx, DefaultLocale)
	if normalizedLocale == DefaultLocale {
		return cloneFlatMessageMap(defaultBundle)
	}

	effective := cloneFlatMessageMap(defaultBundle)
	mergeFlatMessageMaps(effective, s.loadRawLocaleBundle(ctx, normalizedLocale))
	return effective
}

// translateForLocale resolves one translation key against the requested locale.
func (s *serviceImpl) translateForLocale(ctx context.Context, locale string, key string, fallback string) string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fallback
	}

	if value, ok := s.buildRuntimeMessageCatalog(ctx, locale)[trimmedKey]; ok {
		return value
	}
	return fallback
}

// loadRawLocaleBundle loads the non-fallback flat runtime messages for one locale.
func (s *serviceImpl) loadRawLocaleBundle(ctx context.Context, locale string) map[string]string {
	normalizedLocale := normalizeLocale(locale)
	if normalizedLocale == "" {
		normalizedLocale = DefaultLocale
	}

	runtimeBundleCache.RLock()
	cached := runtimeBundleCache.bundles[normalizedLocale]
	runtimeBundleCache.RUnlock()
	if cached != nil {
		return cloneFlatMessageMap(cached)
	}

	bundle := make(map[string]string)
	mergeFlatMessageMaps(bundle, loadEmbeddedHostLocaleBundle(normalizedLocale))
	mergeFlatMessageMaps(bundle, loadSourcePluginLocaleBundle(normalizedLocale))

	runtimeBundleCache.Lock()
	runtimeBundleCache.bundles[normalizedLocale] = cloneFlatMessageMap(bundle)
	runtimeBundleCache.Unlock()
	return bundle
}

// loadSourcePluginLocaleBundle loads source-plugin translation resources from
// registered embedded plugin filesystems.
func loadSourcePluginLocaleBundle(locale string) map[string]string {
	bundle := make(map[string]string)
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return bundle
	}

	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})

	relativePath := path.Join(pluginI18nDir, locale+".json")
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil || sourcePlugin.GetEmbeddedFiles() == nil {
			continue
		}
		content, err := fs.ReadFile(sourcePlugin.GetEmbeddedFiles(), relativePath)
		if err != nil || len(content) == 0 {
			continue
		}
		mergeFlatMessageMaps(bundle, parseLocaleJSON(content))
	}
	return bundle
}

// loadEmbeddedHostLocaleBundle loads host runtime messages from embedded manifest assets.
func loadEmbeddedHostLocaleBundle(locale string) map[string]string {
	content, err := fs.ReadFile(packed.Files, path.Join(hostI18nDir, locale+".json"))
	if err != nil {
		return map[string]string{}
	}
	return parseLocaleJSON(content)
}

// parseLocaleJSON unmarshals one locale JSON file into a flat message catalog.
func parseLocaleJSON(content []byte) map[string]string {
	result := make(map[string]interface{})
	if len(content) == 0 {
		return map[string]string{}
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return map[string]string{}
	}

	flatMessages := make(map[string]string)
	flattenMessageValue("", result, flatMessages)
	return flatMessages
}

// flattenMessageValue flattens one nested message value into dotted keys.
func flattenMessageValue(prefix string, value interface{}, output map[string]string) {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		for key, item := range typedValue {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				continue
			}
			nextPrefix := trimmedKey
			if prefix != "" {
				nextPrefix = prefix + "." + trimmedKey
			}
			flattenMessageValue(nextPrefix, item, output)
		}
	default:
		if prefix == "" {
			return
		}
		output[prefix] = gconv.String(typedValue)
	}
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

// mergeFlatMessageMaps merges source flat messages into destination messages.
func mergeFlatMessageMaps(dst map[string]string, src map[string]string) {
	if len(src) == 0 {
		return
	}
	for key, value := range src {
		dst[key] = value
	}
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

// normalizeAcceptLanguage converts an Accept-Language header into one supported locale.
func normalizeAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		languageTag := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale := normalizeLocale(languageTag); locale != "" {
			return locale
		}
	}
	return ""
}

// normalizeLocale canonicalizes one raw locale value to the host-supported locale set.
func normalizeLocale(value string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(value, "_", "-"))
	if normalized == "" {
		return ""
	}
	switch strings.ToLower(normalized) {
	case "zh", "zh-cn", "zh-hans", "zh-hans-cn":
		return DefaultLocale
	case "en", "en-us", "en-gb":
		return EnglishLocale
	default:
		return ""
	}
}

// buildLocaleNameKey builds the runtime translation key used for one locale display name.
func buildLocaleNameKey(locale string) string {
	return "locale." + locale + ".name"
}

// buildLocaleNativeNameKey builds the runtime translation key used for one locale native display name.
func buildLocaleNativeNameKey(locale string) string {
	return "locale." + locale + ".nativeName"
}
