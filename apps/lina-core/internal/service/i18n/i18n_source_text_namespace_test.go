// This file verifies source-text namespace registration for missing-message
// diagnostics.

package i18n

import (
	"context"
	"fmt"
	"testing"

	"lina-core/internal/model/do"
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
	key := fmt.Sprintf("test.source.unregistered.%s", t.Name())
	insertedID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       DefaultLocale,
		MessageKey:   key,
		MessageValue: "仅默认语言提供",
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nMessageByID(t, ctx, insertedID)
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
	key := fmt.Sprintf("%s%s", prefix, t.Name())
	insertedID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       DefaultLocale,
		MessageKey:   key,
		MessageValue: "仅默认语言提供",
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nMessageByID(t, ctx, insertedID)
	})

	items := New().CheckMissingMessages(ctx, EnglishLocale, prefix)
	if _, ok := findMissingMessage(items, key); ok {
		t.Fatalf("expected registered source-text key %q to be skipped", key)
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
