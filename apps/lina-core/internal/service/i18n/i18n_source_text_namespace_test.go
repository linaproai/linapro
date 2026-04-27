// This file verifies source-text namespace registration for missing-message
// diagnostics.

package i18n

import (
	"context"
	"fmt"
	"testing"
)

// TestCheckMissingMessagesDoesNotSkipUnregisteredSourceTextNamespace verifies
// unregistered namespaces still appear as missing translations.
func TestCheckMissingMessagesDoesNotSkipUnregisteredSourceTextNamespace(t *testing.T) {
	resetRuntimeBundleCache()
	resetSourceTextNamespacesForTest()
	t.Cleanup(func() {
		resetRuntimeBundleCache()
		resetSourceTextNamespacesForTest()
	})

	ctx := context.Background()
	pluginID := nextTestSourcePluginID()
	key := fmt.Sprintf("test.source.unregistered.%s", pluginID)
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"source":{"unregistered":{"%s":"仅默认语言提供"}}}}`, pluginID),
		"zh-TW":       fmt.Sprintf(`{"test":{"source":{"unregistered":{"%s":"繁體中文可用"}}}}`, pluginID),
	})

	items := New().CheckMissingMessages(ctx, EnglishLocale, "test.source.unregistered.")
	if _, ok := findMissingMessage(items, key); !ok {
		t.Fatalf("expected unregistered source-text key %q to remain missing", key)
	}
}

// TestCheckMissingMessagesSkipsRegisteredSourceTextNamespace verifies registered
// namespaces disappear from missing-translation results.
func TestCheckMissingMessagesSkipsRegisteredSourceTextNamespace(t *testing.T) {
	resetRuntimeBundleCache()
	resetSourceTextNamespacesForTest()
	t.Cleanup(func() {
		resetRuntimeBundleCache()
		resetSourceTextNamespacesForTest()
	})

	const prefix = "test.source.registered."
	RegisterSourceTextNamespace(prefix, "test source text")

	ctx := context.Background()
	pluginID := nextTestSourcePluginID()
	key := fmt.Sprintf("%s%s", prefix, pluginID)
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"source":{"registered":{"%s":"仅默认语言提供"}}}}`, pluginID),
		"zh-TW":       fmt.Sprintf(`{"test":{"source":{"registered":{"%s":"繁體中文可用"}}}}`, pluginID),
	})

	for _, locale := range []string{EnglishLocale, "zh-TW"} {
		items := New().CheckMissingMessages(ctx, locale, prefix)
		if _, ok := findMissingMessage(items, key); ok {
			t.Fatalf("expected registered source-text key %q to be skipped for %s", key, locale)
		}
	}
}

// TestRegisteredSourceTextNamespacesReturnsCopy verifies callers cannot mutate
// the registry through the query result.
func TestRegisteredSourceTextNamespacesReturnsCopy(t *testing.T) {
	resetSourceTextNamespacesForTest()
	t.Cleanup(resetSourceTextNamespacesForTest)

	RegisterSourceTextNamespace("test.copy.", "copy test")
	namespaces := RegisteredSourceTextNamespaces()
	namespaces["test.copy."] = "mutated"

	reason, ok := SourceTextNamespaceReason("test.copy.key")
	if !ok {
		t.Fatal("expected registered source-text namespace to resolve")
	}
	if reason != "copy test" {
		t.Fatalf("expected registry copy mutation to be ignored, got %q", reason)
	}
}
