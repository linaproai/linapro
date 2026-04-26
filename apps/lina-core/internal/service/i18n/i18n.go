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
	"lina-core/internal/service/config"
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
	// InvalidateContentCache clears cached sys_i18n_content lookup results.
	InvalidateContentCache()
	// ExportMessages exports flat runtime messages for one locale.
	ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput
	// CheckMissingMessages reports translation keys missing from one locale.
	CheckMissingMessages(ctx context.Context, locale string, keyPrefix string) []MissingMessageItem
	// DiagnoseMessages reports the effective source of runtime messages for one locale.
	DiagnoseMessages(ctx context.Context, locale string, keyPrefix string) []MessageDiagnosticItem
	// ImportMessages writes flat translation messages into sys_i18n_message.
	ImportMessages(ctx context.Context, input MessageImportInput) (MessageImportOutput, error)
	// GetContent resolves one business-content translation from sys_i18n_content and
	// falls back to the runtime default locale or caller-provided default content.
	GetContent(ctx context.Context, input ContentLookupInput) (ContentLookupOutput, error)
	// ListContentVariants lists all enabled locale variants for one business-content anchor.
	ListContentVariants(ctx context.Context, businessType string, businessID string, field string) ([]ContentVariant, error)
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
	configSvc config.Service
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

// Translate returns the localized value for one translation key.
func (s *serviceImpl) Translate(ctx context.Context, key string, fallback string) string {
	return s.translateForLocale(ctx, s.GetLocale(ctx), key, fallback)
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
			IsDefault:  supportedLocale.IsDefault,
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
	normalizedLocale := s.ResolveLocale(ctx, locale)
	defaultLocale := s.getDefaultRuntimeLocale(ctx)

	defaultBundle := s.loadRawLocaleBundle(ctx, defaultLocale)
	if normalizedLocale == defaultLocale {
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
	normalizedLocale := s.ResolveLocale(ctx, locale)

	runtimeBundleCache.RLock()
	cached := runtimeBundleCache.bundles[normalizedLocale]
	runtimeBundleCache.RUnlock()
	if cached != nil {
		return cloneFlatMessageMap(cached)
	}

	bundle := make(map[string]string)
	mergeFlatMessageMaps(bundle, loadEmbeddedHostLocaleBundle(normalizedLocale))
	mergeFlatMessageMaps(bundle, loadSourcePluginLocaleBundle(normalizedLocale))
	mergeFlatMessageMaps(bundle, s.loadDynamicPluginLocaleBundle(ctx, normalizedLocale))
	mergeFlatMessageMaps(bundle, s.loadDatabaseLocaleBundle(ctx, normalizedLocale))

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
// Flat keys override equivalent nested paths, which keeps mixed-format locale
// files deterministic during gradual authoring-format migrations.
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
		keys := make([]string, 0, len(typedValue))
		for key := range typedValue {
			if strings.TrimSpace(key) == "" {
				continue
			}
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			left := strings.TrimSpace(keys[i])
			right := strings.TrimSpace(keys[j])
			if left == right {
				return keys[i] < keys[j]
			}
			return left < right
		})
		for _, key := range keys {
			trimmedKey := strings.TrimSpace(key)
			nextPrefix := trimmedKey
			if prefix != "" {
				nextPrefix = prefix + "." + trimmedKey
			}
			flattenMessageValue(nextPrefix, typedValue[key], output)
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
