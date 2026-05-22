// This file verifies resource-backed runtime message diagnostics without
// relying on database-backed i18n override tables.

package i18n

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/config"
	"lina-core/pkg/pluginhost"
)

var testSourcePluginSequence atomic.Uint64

// nextTestSourcePluginID returns a unique source-plugin fixture ID.
func nextTestSourcePluginID() string {
	return fmt.Sprintf("plugin-i18n-test-%d", testSourcePluginSequence.Add(1))
}

// registerTestSourcePluginI18N registers a source-plugin fixture with the
// provided locale JSON resources and invalidates the runtime cache.
func registerTestSourcePluginI18N(t *testing.T, pluginID string, localeFiles map[string]string) {
	t.Helper()

	fileSystem := fstest.MapFS{}
	fileSystem["plugin.yaml"] = &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(pluginID, true))}
	for locale, content := range localeFiles {
		normalizedLocale := normalizeLocale(locale)
		if normalizedLocale == "" {
			t.Fatalf("invalid locale in test fixture: %q", locale)
		}
		fileSystem["manifest/i18n/"+normalizedLocale+"/plugin.json"] = &fstest.MapFile{Data: []byte(content)}
	}

	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(fileSystem)
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}
	resetRuntimeBundleCache()
}

// sourcePluginI18NManifestFixture builds the minimal embedded plugin.yaml used
// by runtime i18n tests that register in-memory source plugins.
func sourcePluginI18NManifestFixture(pluginID string, enabled bool) string {
	enabledValue := "false"
	if enabled {
		enabledValue = "true"
	}
	return fmt.Sprintf(`id: %s
name: Runtime I18N Test Plugin
version: v0.1.0
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
i18n:
  enabled: %s
  default: zh-CN
  locales:
    - locale: zh-CN
      nativeName: 简体中文
    - locale: en-US
      nativeName: English
`, pluginID, enabledValue)
}

// TestRuntimeTranslationsDoNotImplicitlyUseDefaultLocale verifies that current
// locale translation methods never return default-locale text unless the caller
// explicitly asks for default-locale fallback.
func TestRuntimeTranslationsDoNotImplicitlyUseDefaultLocale(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"strict":{"%s":{"title":"仅默认语言提供"}}}}`, pluginID),
	})
	key := "test.strict." + pluginID + ".title"

	ctx := context.Background()
	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil))
	messages := svc.BuildRuntimeMessages(ctx, EnglishLocale)
	if value, ok := lookupMessageString(messages, key); ok {
		t.Fatalf("expected en-US runtime messages to omit default-locale-only key, got %q", value)
	}

	enCtx := context.WithValue(ctx, gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})
	if actual := svc.Translate(enCtx, key, "Source Fallback"); actual != "Source Fallback" {
		t.Fatalf("expected Translate to use caller fallback, got %q", actual)
	}
	if actual := svc.TranslateSourceText(enCtx, key, "Source Text"); actual != "Source Text" {
		t.Fatalf("expected TranslateSourceText to use source text, got %q", actual)
	}
	if actual := svc.TranslateOrKey(enCtx, key); actual != key {
		t.Fatalf("expected TranslateOrKey to return key placeholder %q, got %q", key, actual)
	}
	if actual := svc.TranslateWithDefaultLocale(enCtx, key, "fallback"); actual != "仅默认语言提供" {
		t.Fatalf("expected explicit default-locale fallback %q, got %q", "仅默认语言提供", actual)
	}
}

// TestBuildRuntimeMessagesHonorsSourcePluginI18NPolicy verifies runtime UI
// translations follow the same plugin.yaml opt-in rule as apidoc resources.
func TestBuildRuntimeMessagesHonorsSourcePluginI18NPolicy(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	managedPluginID := nextTestSourcePluginID()
	managedPlugin := pluginhost.NewSourcePlugin(managedPluginID)
	managedPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(managedPluginID, true))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Runtime Managed"}}}`,
			managedPluginID,
		))},
	})
	cleanupManaged, err := pluginhost.RegisterSourcePluginForTest(managedPlugin)
	if err != nil {
		t.Fatalf("register managed source plugin failed: %v", err)
	}
	t.Cleanup(cleanupManaged)

	optOutPluginID := nextTestSourcePluginID()
	optOutPlugin := pluginhost.NewSourcePlugin(optOutPluginID)
	optOutPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(optOutPluginID, false))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Should Not Load"}}}`,
			optOutPluginID,
		))},
	})
	cleanupOptOut, err := pluginhost.RegisterSourcePluginForTest(optOutPlugin)
	if err != nil {
		t.Fatalf("register opt-out source plugin failed: %v", err)
	}
	t.Cleanup(cleanupOptOut)

	missingPolicyPluginID := nextTestSourcePluginID()
	missingPolicyPlugin := pluginhost.NewSourcePlugin(missingPolicyPluginID)
	missingPolicyPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(fmt.Sprintf(`id: %s
name: Runtime I18N Missing Policy Test Plugin
version: v0.1.0
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
`, missingPolicyPluginID))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Should Also Not Load"}}}`,
			missingPolicyPluginID,
		))},
	})
	cleanupMissingPolicy, err := pluginhost.RegisterSourcePluginForTest(missingPolicyPlugin)
	if err != nil {
		t.Fatalf("register missing-policy source plugin failed: %v", err)
	}
	t.Cleanup(cleanupMissingPolicy)
	resetRuntimeBundleCache()

	messages := New(bizctx.New(), config.New(), cachecoord.Default(nil)).BuildRuntimeMessages(context.Background(), EnglishLocale)
	managedKey := "test.policy." + managedPluginID
	if actual, ok := lookupMessageString(messages, managedKey); !ok || actual != "Runtime Managed" {
		t.Fatalf("expected managed plugin translation %q, got %q (exists=%v)", "Runtime Managed", actual, ok)
	}
	if value, ok := lookupMessageString(messages, "test.policy."+optOutPluginID); ok {
		t.Fatalf("expected i18n.enabled=false plugin translation to be skipped, got %q", value)
	}
	if value, ok := lookupMessageString(messages, "test.policy."+missingPolicyPluginID); ok {
		t.Fatalf("expected plugin without i18n policy translation to be skipped, got %q", value)
	}
}

// TestCheckMissingMessagesReturnsLocaleGaps verifies that missing translation
// diagnostics compare the target locale against the default locale baseline.
func TestCheckMissingMessagesReturnsLocaleGaps(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"missing":{"%s":"仅默认语言提供"}}}`, pluginID),
	})
	key := "test.missing." + pluginID

	items := New(bizctx.New(), config.New(), cachecoord.Default(nil)).CheckMissingMessages(context.Background(), EnglishLocale, "test.missing.")
	missingItem, ok := findMissingMessage(items, key)
	if !ok {
		t.Fatalf("expected missing translation key %q", key)
	}
	if missingItem.DefaultValue != "仅默认语言提供" {
		t.Fatalf("expected default fallback value %q, got %q", "仅默认语言提供", missingItem.DefaultValue)
	}
	if missingItem.Source.Type != string(messageOriginTypePluginFile) {
		t.Fatalf("expected plugin source type %q, got %q", string(messageOriginTypePluginFile), missingItem.Source.Type)
	}
	if missingItem.Source.ScopeKey != pluginID {
		t.Fatalf("expected plugin source key %q, got %q", pluginID, missingItem.Source.ScopeKey)
	}
}

// TestDiagnoseMessagesReportsPluginSource verifies that source diagnostics
// report resource-backed plugin messages from the resource-only source model.
func TestDiagnoseMessagesReportsPluginSource(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		EnglishLocale: fmt.Sprintf(`{"test":{"diagnose":{"%s":"Plugin Diagnose Value"}}}`, pluginID),
	})
	key := "test.diagnose." + pluginID

	items := New(bizctx.New(), config.New(), cachecoord.Default(nil)).DiagnoseMessages(context.Background(), EnglishLocale, "test.diagnose.")
	diagnosticItem, ok := findDiagnosticMessage(items, key)
	if !ok {
		t.Fatalf("expected diagnostic translation key %q", key)
	}
	if diagnosticItem.Value != "Plugin Diagnose Value" {
		t.Fatalf("expected diagnostic value %q, got %q", "Plugin Diagnose Value", diagnosticItem.Value)
	}
	if diagnosticItem.FromFallback {
		t.Fatal("expected diagnostic item to not use fallback")
	}
	if diagnosticItem.Source.Type != string(messageOriginTypePluginFile) {
		t.Fatalf("expected plugin source type %q, got %q", string(messageOriginTypePluginFile), diagnosticItem.Source.Type)
	}
	if diagnosticItem.Source.ScopeKey != pluginID {
		t.Fatalf("expected plugin source key %q, got %q", pluginID, diagnosticItem.Source.ScopeKey)
	}
}

// findMissingMessage locates one missing-translation item by key.
func findMissingMessage(items []MissingMessageItem, key string) (MissingMessageItem, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return MissingMessageItem{}, false
}

// findDiagnosticMessage locates one source-diagnostic item by key.
func findDiagnosticMessage(items []MessageDiagnosticItem, key string) (MessageDiagnosticItem, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return MessageDiagnosticItem{}, false
}
