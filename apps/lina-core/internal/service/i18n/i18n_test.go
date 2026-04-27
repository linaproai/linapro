// This file verifies locale normalization, runtime bundle aggregation, and
// context-aware translation behavior for the host i18n service.
package i18n

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gvalid"

	"lina-core/internal/model"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/pluginhost"
)

const testPluginID = "plugin-i18n-test"
const testCacheInvalidatePluginID = "plugin-i18n-cache-invalidate"

// stubConfigService supplies focused i18n config fixtures without requiring a
// full host config service implementation for locale tests.
type stubConfigService struct {
	hostconfig.Service
	cfg *hostconfig.I18nConfig
}

// GetI18n returns the fixture i18n config for locale tests.
func (s stubConfigService) GetI18n(_ context.Context) *hostconfig.I18nConfig {
	return s.cfg
}

// init registers one minimal source plugin fixture with embedded i18n assets.
func init() {
	plugin := pluginhost.NewSourcePlugin(testPluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"manifest/i18n/en-US.json": &fstest.MapFile{Data: []byte(`{
  "plugin": {
    "plugin-i18n-test": {
      "name": "Runtime Test Plugin"
    }
  }
}`)},
	})
	pluginhost.RegisterSourcePlugin(plugin)
}

// resetRuntimeBundleCache clears the in-memory runtime bundle cache between tests.
func resetRuntimeBundleCache() {
	invalidateRuntimeBundleCache()
	invalidateRuntimeLocaleCache()
}

// TestNormalizeLocale verifies that raw locale aliases normalize to canonical locale codes.
func TestNormalizeLocale(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		raw      string
		expected string
	}{
		{name: "zh short tag", raw: "zh", expected: "zh"},
		{name: "zh underscore", raw: "zh_CN", expected: DefaultLocale},
		{name: "english us", raw: "en-US", expected: EnglishLocale},
		{name: "traditional chinese", raw: "zh_tw", expected: "zh-TW"},
		{name: "english gb", raw: "en-gb", expected: "en-GB"},
		{name: "french", raw: "fr-fr", expected: "fr-FR"},
		{name: "script tag", raw: "zh_hans_cn", expected: "zh-Hans-CN"},
		{name: "invalid", raw: "zh-中文", expected: ""},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if actual := normalizeLocale(testCase.raw); actual != testCase.expected {
				t.Fatalf("expected locale %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestNormalizeAcceptLanguage verifies that the first valid language tag is normalized.
func TestNormalizeAcceptLanguage(t *testing.T) {
	t.Parallel()

	header := "fr-FR, en-GB;q=0.8, zh-CN;q=0.6"
	if actual := normalizeAcceptLanguage(header); actual != "fr-FR" {
		t.Fatalf("expected accept-language locale %q, got %q", "fr-FR", actual)
	}
}

// TestResolveLocaleFallsBackToDefault verifies that explicit unsupported locales
// fall back to the configured runtime default language.
func TestResolveLocaleFallsBackToDefault(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	if actual := svc.ResolveLocale(context.Background(), "fr-FR"); actual != DefaultLocale {
		t.Fatalf("expected unsupported locale to fall back to %q, got %q", DefaultLocale, actual)
	}
}

// TestParseLocaleJSONSupportsNestedAndFlatKeys verifies that resource files can
// be maintained with flat keys while remaining backward compatible with nested JSON.
func TestParseLocaleJSONSupportsNestedAndFlatKeys(t *testing.T) {
	t.Parallel()

	flatCatalog := parseLocaleJSON([]byte(`{
  "menu.dashboard.title": "Workbench",
  "plugin.demo.name": "Demo"
}`))
	if actual := flatCatalog["menu.dashboard.title"]; actual != "Workbench" {
		t.Fatalf("expected flat key translation %q, got %q", "Workbench", actual)
	}

	nestedCatalog := parseLocaleJSON([]byte(`{
  "menu": {
    "dashboard": {
      "title": "Workbench"
    }
  }
}`))
	if actual := nestedCatalog["menu.dashboard.title"]; actual != "Workbench" {
		t.Fatalf("expected nested key translation %q, got %q", "Workbench", actual)
	}

	mixedCatalog := parseLocaleJSON([]byte(`{
  "menu": {
    "dashboard": {
      "title": "Nested Workbench"
    }
  },
  "menu.dashboard.title": "Flat Workbench"
}`))
	if actual := mixedCatalog["menu.dashboard.title"]; actual != "Flat Workbench" {
		t.Fatalf("expected flat key to override nested translation %q, got %q", "Flat Workbench", actual)
	}
}

// TestBuildRuntimeMessagesIncludesHostAndSourcePlugin verifies that the runtime
// message bundle merges host translations with registered source-plugin assets.
func TestBuildRuntimeMessagesIncludesHostAndSourcePlugin(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	messages := svc.BuildRuntimeMessages(context.Background(), EnglishLocale)

	if actual, ok := lookupMessageString(messages, "menu.dashboard.title"); !ok || actual != "Dashboard" {
		t.Fatalf("expected host menu translation %q, got %q (exists=%v)", "Dashboard", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "dict.cron_job_status.name"); !ok || actual != "Scheduled Job Status" {
		t.Fatalf("expected scheduled-job dict translation %q, got %q (exists=%v)", "Scheduled Job Status", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "dict.sys_menu_type.B.label"); !ok || actual != "Button" {
		t.Fatalf("expected built-in menu-type translation %q, got %q (exists=%v)", "Button", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "plugin.plugin-i18n-test.name"); !ok || actual != "Runtime Test Plugin" {
		t.Fatalf("expected plugin translation %q, got %q (exists=%v)", "Runtime Test Plugin", actual, ok)
	}
}

// TestListRuntimeLocalesUsesRequestedDisplayLocale verifies that the runtime
// locale list exposes localized display names and stable native names.
func TestListRuntimeLocalesUsesRequestedDisplayLocale(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	locales := svc.ListRuntimeLocales(context.Background(), EnglishLocale)
	if len(locales) != 3 {
		t.Fatalf("expected 3 runtime locales, got %d", len(locales))
	}

	localeMap := make(map[string]LocaleDescriptor, len(locales))
	for _, locale := range locales {
		localeMap[locale.Locale] = locale
	}

	zhLocale, ok := localeMap[DefaultLocale]
	if !ok {
		t.Fatalf("expected locale %q to be returned", DefaultLocale)
	}
	if zhLocale.Name != "Chinese (Simplified)" {
		t.Fatalf("expected localized locale name %q, got %q", "Chinese (Simplified)", zhLocale.Name)
	}
	if zhLocale.NativeName != "简体中文" {
		t.Fatalf("expected locale native name %q, got %q", "简体中文", zhLocale.NativeName)
	}
	if !zhLocale.IsDefault {
		t.Fatal("expected zh-CN locale to be marked as default")
	}

	enLocale, ok := localeMap[EnglishLocale]
	if !ok {
		t.Fatalf("expected locale %q to be returned", EnglishLocale)
	}
	if enLocale.Name != "English" {
		t.Fatalf("expected localized locale name %q, got %q", "English", enLocale.Name)
	}
	if enLocale.NativeName != "English" {
		t.Fatalf("expected locale native name %q, got %q", "English", enLocale.NativeName)
	}
	if enLocale.IsDefault {
		t.Fatal("expected en-US locale to not be marked as default")
	}

	twLocale, ok := localeMap["zh-TW"]
	if !ok {
		t.Fatalf("expected locale %q to be returned", "zh-TW")
	}
	if twLocale.Name != "Chinese (Traditional)" {
		t.Fatalf("expected localized locale name %q, got %q", "Chinese (Traditional)", twLocale.Name)
	}
	if twLocale.NativeName != "繁體中文" {
		t.Fatalf("expected locale native name %q, got %q", "繁體中文", twLocale.NativeName)
	}
	if twLocale.Direction != LocaleDirectionLTR.String() {
		t.Fatalf("expected locale direction %q, got %q", LocaleDirectionLTR.String(), twLocale.Direction)
	}
	if twLocale.IsDefault {
		t.Fatal("expected zh-TW locale to not be marked as default")
	}
}

// TestBuildConfiguredRuntimeLocalesUsesConfigLocalesAsWhitelist verifies that
// removing a locale from config i18n.locales disables it even when its JSON
// resource file still exists.
func TestBuildConfiguredRuntimeLocalesUsesConfigLocalesAsWhitelist(t *testing.T) {
	t.Parallel()

	config := &hostconfig.I18nConfig{
		Default: DefaultLocale,
		Enabled: true,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: EnglishLocale, NativeName: "English"},
			{Locale: DefaultLocale, NativeName: "简体中文"},
		},
	}
	locales := normalizeRuntimeLocales(buildConfiguredRuntimeLocales(
		[]string{DefaultLocale, EnglishLocale, "zh-TW"},
		config,
	), config.Default)

	if len(locales) != 2 {
		t.Fatalf("expected 2 enabled locales, got %d: %+v", len(locales), locales)
	}
	for _, locale := range locales {
		if locale.Locale == "zh-TW" {
			t.Fatalf("expected zh-TW to be disabled when missing from config locales: %+v", locales)
		}
	}
}

// TestBuildConfiguredRuntimeLocalesDisabledReturnsDefaultOnly verifies that
// i18n.enabled=false suppresses all non-default runtime locales.
func TestBuildConfiguredRuntimeLocalesDisabledReturnsDefaultOnly(t *testing.T) {
	t.Parallel()

	config := &hostconfig.I18nConfig{
		Default: DefaultLocale,
		Enabled: false,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: EnglishLocale, NativeName: "English"},
			{Locale: DefaultLocale, NativeName: "简体中文"},
			{Locale: "zh-TW", NativeName: "繁體中文"},
		},
	}
	locales := normalizeRuntimeLocales(buildConfiguredRuntimeLocales(
		[]string{DefaultLocale, EnglishLocale, "zh-TW"},
		config,
	), config.Default)

	if len(locales) != 1 {
		t.Fatalf("expected only one locale when i18n is disabled, got %d: %+v", len(locales), locales)
	}
	if locales[0].Locale != DefaultLocale || !locales[0].IsDefault {
		t.Fatalf("expected disabled i18n to keep only default locale, got %+v", locales[0])
	}
}

// TestFallbackRuntimeLocalesUsesConfiguredDefault verifies the last-resort
// runtime locale list is still driven by i18n.default.
func TestFallbackRuntimeLocalesUsesConfiguredDefault(t *testing.T) {
	t.Parallel()

	locales := fallbackRuntimeLocales(&hostconfig.I18nConfig{Default: EnglishLocale})

	if len(locales) != 1 {
		t.Fatalf("expected one fallback locale, got %d: %+v", len(locales), locales)
	}
	if locales[0].Locale != EnglishLocale || !locales[0].IsDefault {
		t.Fatalf("expected fallback locale to use configured default, got %+v", locales[0])
	}
}

// TestGetDefaultRuntimeLocaleUsesConfiguredDefault verifies default-locale
// resolution does not depend on the package-level test locale constants.
func TestGetDefaultRuntimeLocaleUsesConfiguredDefault(t *testing.T) {
	resetRuntimeBundleCache()

	cfg := &hostconfig.I18nConfig{
		Default: EnglishLocale,
		Enabled: false,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: DefaultLocale, NativeName: "简体中文"},
			{Locale: EnglishLocale, NativeName: "English"},
		},
	}
	svc := &serviceImpl{configSvc: stubConfigService{cfg: cfg}}

	if actual := svc.getDefaultRuntimeLocale(context.Background()); actual != EnglishLocale {
		t.Fatalf("expected configured default locale %q, got %q", EnglishLocale, actual)
	}
}

// TestRegisterSourcePluginInvalidatesRuntimeBundleCache verifies that source
// plugin registrations clear the cached runtime bundle so new translations are visible.
func TestRegisterSourcePluginInvalidatesRuntimeBundleCache(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	messages := svc.BuildRuntimeMessages(context.Background(), EnglishLocale)
	if _, ok := lookupMessageString(messages, "plugin."+testCacheInvalidatePluginID+".name"); ok {
		t.Fatalf("expected plugin %q translation to be absent before registration", testCacheInvalidatePluginID)
	}

	plugin := pluginhost.NewSourcePlugin(testCacheInvalidatePluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"manifest/i18n/en-US.json": &fstest.MapFile{Data: []byte(`{
  "plugin": {
    "plugin-i18n-cache-invalidate": {
      "name": "Cache Invalidation Plugin"
    }
  }
}`)},
	})
	pluginhost.RegisterSourcePlugin(plugin)

	messages = svc.BuildRuntimeMessages(context.Background(), EnglishLocale)
	if actual, ok := lookupMessageString(messages, "plugin."+testCacheInvalidatePluginID+".name"); !ok || actual != "Cache Invalidation Plugin" {
		t.Fatalf("expected cache-invalidated plugin translation %q, got %q (exists=%v)", "Cache Invalidation Plugin", actual, ok)
	}
}

// TestTranslateUsesContextLocaleAndFallback verifies that Translate resolves the
// locale from business context and falls back to the provided literal when needed.
func TestTranslateUsesContextLocaleAndFallback(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	if actual := svc.Translate(ctx, "framework.description", "fallback"); actual == "fallback" {
		t.Fatal("expected translated framework description, got fallback")
	}
	if actual := svc.Translate(ctx, "missing.translation.key", "fallback"); actual != "fallback" {
		t.Fatalf("expected fallback value %q, got %q", "fallback", actual)
	}
	if actual := svc.TranslateSourceText(ctx, "job.handler.host.session-cleanup.name", "Online Session Cleanup"); actual != "Online Session Cleanup" {
		t.Fatalf("expected source text fallback %q, got %q", "Online Session Cleanup", actual)
	}
}

// TestCheckMissingMessagesSkipsSourceTextBackedKeys verifies that missing
// diagnostics do not require JSON copies for source-owned keys.
func TestCheckMissingMessagesSkipsSourceTextBackedKeys(t *testing.T) {
	resetRuntimeBundleCache()
	resetSourceTextNamespacesForTest()
	RegisterSourceTextNamespace("job.handler.", "test job handler source text")
	RegisterSourceTextNamespace("job.group.default.", "test default group source text")
	t.Cleanup(func() {
		resetRuntimeBundleCache()
		resetSourceTextNamespacesForTest()
	})

	for _, locale := range []string{EnglishLocale, "zh-TW"} {
		items := New().CheckMissingMessages(context.Background(), locale, "job.")
		namespaces := RegisteredSourceTextNamespaces()
		for _, item := range items {
			for prefix := range namespaces {
				if strings.HasPrefix(item.Key, prefix) {
					t.Fatalf("expected source-text-backed key %q to be skipped for %s", item.Key, locale)
				}
			}
		}
	}
}

// TestTraditionalChineseRuntimeCatalogHasNoMissingMessages verifies that the
// shipped Traditional Chinese runtime bundle covers the default-language
// baseline except source-owned keys.
func TestTraditionalChineseRuntimeCatalogHasNoMissingMessages(t *testing.T) {
	resetRuntimeBundleCache()
	resetSourceTextNamespacesForTest()
	RegisterSourceTextNamespace("job.handler.", "test job handler source text")
	RegisterSourceTextNamespace("job.group.default.", "test default group source text")
	t.Cleanup(func() {
		resetRuntimeBundleCache()
		resetSourceTextNamespacesForTest()
	})

	items := New().CheckMissingMessages(context.Background(), "zh-TW", "")
	items = filterExternalDynamicPluginMissingMessagesForTest(items)
	if len(items) == 0 {
		return
	}

	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
		if len(keys) >= 20 {
			break
		}
	}
	t.Fatalf("expected zh-TW missing translation total=0, got %d; first keys: %s", len(items), strings.Join(keys, ", "))
}

// filterExternalDynamicPluginMissingMessagesForTest removes gaps contributed by
// previously installed dynamic-plugin release artifacts in the developer
// database. This test verifies shipped host/source resources; dynamic-plugin
// artifact freshness is covered by the focused dynamic-plugin tests and E2E.
func filterExternalDynamicPluginMissingMessagesForTest(items []MissingMessageItem) []MissingMessageItem {
	filteredItems := make([]MissingMessageItem, 0, len(items))
	for _, item := range items {
		if item.Source.ScopeKey == "plugin-demo-dynamic" {
			continue
		}
		filteredItems = append(filteredItems, item)
	}
	return filteredItems
}

// TestLocalizeErrorSupportsFormattedBusinessKeys verifies that backend error
// keys can be formatted after translation using gerror text arguments.
func TestLocalizeErrorSupportsFormattedBusinessKeys(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	actual := svc.LocalizeError(ctx, gerror.Newf("error.upload.fileTooLarge", 20))
	if actual != "File size must not exceed 20MB" {
		t.Fatalf("expected localized formatted error %q, got %q", "File size must not exceed 20MB", actual)
	}
}

// TestLocalizeErrorSupportsValidationKeys verifies that flat validation keys
// are translated after validation when they were stored as message IDs.
func TestLocalizeErrorSupportsValidationKeys(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	err := gvalid.New().
		Data("").
		Rules("required").
		Messages("validation.auth.login.username.required").
		Run(ctx)
	if err == nil {
		t.Fatal("expected validation error")
	}

	actual := svc.LocalizeError(ctx, err)
	if actual != "Please enter a username" {
		t.Fatalf("expected localized validation error %q, got %q", "Please enter a username", actual)
	}
}
