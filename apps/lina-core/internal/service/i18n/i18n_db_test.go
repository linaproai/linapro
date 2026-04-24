// This file verifies database-backed locale registry and message override behavior.

package i18n

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

const (
	testFrenchLocale = "fr-FR"
)

// TestBuildRuntimeMessagesIncludesDatabaseOverrides verifies that enabled
// database overrides replace file-based runtime messages.
func TestBuildRuntimeMessagesIncludesDatabaseOverrides(t *testing.T) {
	resetRuntimeBundleCache()

	ctx := context.Background()
	overrideKey := "menu.dashboard.title"
	overrideValue := "Database Workbench"
	insertedID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       EnglishLocale,
		MessageKey:   overrideKey,
		MessageValue: overrideValue,
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nMessageByID(t, ctx, insertedID)
		resetRuntimeBundleCache()
	})

	messages := New().BuildRuntimeMessages(ctx, EnglishLocale)
	if actual, ok := lookupMessageString(messages, overrideKey); !ok || actual != overrideValue {
		t.Fatalf("expected database override %q, got %q (exists=%v)", overrideValue, actual, ok)
	}
}

// TestImportMessagesCreatesAndUpdatesRows verifies that import can create and
// update database overrides and that exported messages reflect the result.
func TestImportMessagesCreatesAndUpdatesRows(t *testing.T) {
	resetRuntimeBundleCache()

	ctx := context.Background()
	svc := New()
	key := fmt.Sprintf("test.import.%s", t.Name())

	firstOutput, err := svc.ImportMessages(ctx, MessageImportInput{
		Locale:    EnglishLocale,
		Overwrite: true,
		Messages: map[string]string{
			key: "First Value",
		},
		Remark: t.Name(),
	})
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	if firstOutput.Created != 1 || firstOutput.Updated != 0 || firstOutput.Skipped != 0 {
		t.Fatalf("unexpected first import output: %+v", firstOutput)
	}

	secondOutput, err := svc.ImportMessages(ctx, MessageImportInput{
		Locale:    EnglishLocale,
		Overwrite: true,
		Messages: map[string]string{
			key: "Second Value",
		},
		Remark: t.Name(),
	})
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if secondOutput.Created != 0 || secondOutput.Updated != 1 || secondOutput.Skipped != 0 {
		t.Fatalf("unexpected second import output: %+v", secondOutput)
	}
	t.Cleanup(func() {
		deleteI18nMessageByUniqueKey(t, ctx, EnglishLocale, key, string(messageScopeTypeHost), hostMessageScopeKey)
	})

	exported := svc.ExportMessages(ctx, EnglishLocale, false)
	if actual, ok := exported.Messages[key]; !ok || actual != "Second Value" {
		t.Fatalf("expected exported imported value %q, got %q (exists=%v)", "Second Value", actual, ok)
	}
}

// TestListRuntimeLocalesIncludesDatabaseLocales verifies that enabled locale
// registry rows participate in runtime locale listing.
func TestListRuntimeLocalesIncludesDatabaseLocales(t *testing.T) {
	resetRuntimeBundleCache()

	ctx := context.Background()
	localeID := insertI18nLocaleForTest(t, ctx, do.SysI18NLocale{
		Locale:     testFrenchLocale,
		Name:       "法语",
		NativeName: "Français",
		Sort:       99,
		Status:     int(localeStatusEnabled),
		IsDefault:  int(localeDefaultNo),
		Remark:     t.Name(),
	})
	nameMessageID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       EnglishLocale,
		MessageKey:   buildLocaleNameKey(testFrenchLocale),
		MessageValue: "French",
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	nativeNameMessageID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       testFrenchLocale,
		MessageKey:   buildLocaleNativeNameKey(testFrenchLocale),
		MessageValue: "Français",
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nMessageByID(t, ctx, nativeNameMessageID)
		deleteI18nMessageByID(t, ctx, nameMessageID)
		deleteI18nLocaleByID(t, ctx, localeID)
		resetRuntimeBundleCache()
	})

	locales := New().ListRuntimeLocales(ctx, EnglishLocale)
	frenchLocale, ok := findLocaleDescriptor(locales, testFrenchLocale)
	if !ok {
		t.Fatalf("expected runtime locale %q to exist", testFrenchLocale)
	}
	if frenchLocale.Name != "French" {
		t.Fatalf("expected localized locale name %q, got %q", "French", frenchLocale.Name)
	}
	if frenchLocale.NativeName != "Français" {
		t.Fatalf("expected locale native name %q, got %q", "Français", frenchLocale.NativeName)
	}
	if frenchLocale.IsDefault {
		t.Fatal("expected test locale to not be marked as default")
	}
}

// TestCheckMissingMessagesReturnsLocaleGaps verifies that missing translation
// diagnostics compare the target locale against the default locale baseline.
func TestCheckMissingMessagesReturnsLocaleGaps(t *testing.T) {
	resetRuntimeBundleCache()

	ctx := context.Background()
	key := fmt.Sprintf("test.missing.%s", t.Name())
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
		resetRuntimeBundleCache()
	})

	items := New().CheckMissingMessages(ctx, EnglishLocale, "test.missing.")
	missingItem, ok := findMissingMessage(items, key)
	if !ok {
		t.Fatalf("expected missing translation key %q", key)
	}
	if missingItem.DefaultValue != "仅默认语言提供" {
		t.Fatalf("expected default fallback value %q, got %q", "仅默认语言提供", missingItem.DefaultValue)
	}
	if missingItem.Source.Type != string(messageOriginTypeDatabase) {
		t.Fatalf("expected database source type %q, got %q", string(messageOriginTypeDatabase), missingItem.Source.Type)
	}
}

// TestDiagnoseMessagesReportsDatabaseSource verifies that source diagnostics can
// report database-backed overrides without falling back.
func TestDiagnoseMessagesReportsDatabaseSource(t *testing.T) {
	resetRuntimeBundleCache()

	ctx := context.Background()
	key := fmt.Sprintf("test.diagnose.%s", t.Name())
	insertedID := insertI18nMessageForTest(t, ctx, do.SysI18NMessage{
		Locale:       EnglishLocale,
		MessageKey:   key,
		MessageValue: "Database Diagnose Value",
		ScopeType:    "host",
		ScopeKey:     "core",
		SourceType:   "manual",
		Status:       int(messageStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nMessageByID(t, ctx, insertedID)
		resetRuntimeBundleCache()
	})

	items := New().DiagnoseMessages(ctx, EnglishLocale, "test.diagnose.")
	diagnosticItem, ok := findDiagnosticMessage(items, key)
	if !ok {
		t.Fatalf("expected diagnostic translation key %q", key)
	}
	if diagnosticItem.Value != "Database Diagnose Value" {
		t.Fatalf("expected diagnostic value %q, got %q", "Database Diagnose Value", diagnosticItem.Value)
	}
	if diagnosticItem.FromFallback {
		t.Fatal("expected diagnostic item to not use fallback")
	}
	if diagnosticItem.Source.Type != string(messageOriginTypeDatabase) {
		t.Fatalf("expected diagnostic source type %q, got %q", string(messageOriginTypeDatabase), diagnosticItem.Source.Type)
	}
}

// insertI18nLocaleForTest inserts one locale registry row and invalidates runtime caches.
func insertI18nLocaleForTest(t *testing.T, ctx context.Context, data do.SysI18NLocale) uint64 {
	t.Helper()

	insertedID, err := dao.SysI18NLocale.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert i18n locale: %v", err)
	}
	resetRuntimeBundleCache()
	return uint64(insertedID)
}

// insertI18nMessageForTest inserts one message override row and invalidates runtime caches.
func insertI18nMessageForTest(t *testing.T, ctx context.Context, data do.SysI18NMessage) uint64 {
	t.Helper()

	insertedID, err := dao.SysI18NMessage.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert i18n message: %v", err)
	}
	resetRuntimeBundleCache()
	return uint64(insertedID)
}

// deleteI18nLocaleByID removes one locale registry row and invalidates runtime caches.
func deleteI18nLocaleByID(t *testing.T, ctx context.Context, id uint64) {
	t.Helper()

	if _, err := dao.SysI18NLocale.Ctx(ctx).Unscoped().Where(do.SysI18NLocale{Id: id}).Delete(); err != nil {
		t.Fatalf("delete i18n locale %d: %v", id, err)
	}
	resetRuntimeBundleCache()
}

// deleteI18nMessageByID removes one message override row and invalidates runtime caches.
func deleteI18nMessageByID(t *testing.T, ctx context.Context, id uint64) {
	t.Helper()

	if _, err := dao.SysI18NMessage.Ctx(ctx).Unscoped().Where(do.SysI18NMessage{Id: id}).Delete(); err != nil {
		t.Fatalf("delete i18n message %d: %v", id, err)
	}
	resetRuntimeBundleCache()
}

// deleteI18nMessageByUniqueKey removes one message override row by its unique scope key.
func deleteI18nMessageByUniqueKey(t *testing.T, ctx context.Context, locale string, key string, scopeType string, scopeKey string) {
	t.Helper()

	if _, err := dao.SysI18NMessage.Ctx(ctx).
		Unscoped().
		Where(do.SysI18NMessage{
			Locale:     locale,
			MessageKey: key,
			ScopeType:  scopeType,
			ScopeKey:   scopeKey,
		}).
		Delete(); err != nil {
		t.Fatalf("delete i18n message %s/%s/%s/%s: %v", locale, key, scopeType, scopeKey, err)
	}
	resetRuntimeBundleCache()
}

// findLocaleDescriptor locates one runtime locale descriptor by locale code.
func findLocaleDescriptor(items []LocaleDescriptor, locale string) (LocaleDescriptor, bool) {
	for _, item := range items {
		if item.Locale == locale {
			return item, true
		}
	}
	return LocaleDescriptor{}, false
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
