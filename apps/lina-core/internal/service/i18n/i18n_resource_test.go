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
	for locale, content := range localeFiles {
		normalizedLocale := normalizeLocale(locale)
		if normalizedLocale == "" {
			t.Fatalf("invalid locale in test fixture: %q", locale)
		}
		fileSystem["manifest/i18n/"+normalizedLocale+"/plugin.json"] = &fstest.MapFile{Data: []byte(content)}
	}

	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(fileSystem)
	pluginhost.RegisterSourcePlugin(plugin)
	resetRuntimeBundleCache()
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
	svc := New()
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

	items := New().CheckMissingMessages(context.Background(), EnglishLocale, "test.missing.")
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

	items := New().DiagnoseMessages(context.Background(), EnglishLocale, "test.diagnose.")
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
