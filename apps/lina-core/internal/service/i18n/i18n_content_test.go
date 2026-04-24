// This file verifies sys_i18n_content fallback rules and cache invalidation behavior.

package i18n

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

// TestGetContentReturnsRequestedLocaleVariant verifies that sys_i18n_content
// prefers the requested locale when that variant exists.
func TestGetContentReturnsRequestedLocaleVariant(t *testing.T) {
	resetContentCache()

	ctx := context.Background()
	businessType := "notice"
	businessID := fmt.Sprintf("content-%s", t.Name())
	field := "title"
	zhID := insertI18nContentForTest(t, ctx, do.SysI18NContent{
		BusinessType: businessType,
		BusinessId:   businessID,
		Field:        field,
		Locale:       DefaultLocale,
		ContentType:  string(ContentTypePlain),
		Content:      "中文标题",
		Status:       int(contentStatusEnabled),
		Remark:       t.Name(),
	})
	enID := insertI18nContentForTest(t, ctx, do.SysI18NContent{
		BusinessType: businessType,
		BusinessId:   businessID,
		Field:        field,
		Locale:       EnglishLocale,
		ContentType:  string(ContentTypePlain),
		Content:      "English Title",
		Status:       int(contentStatusEnabled),
		Remark:       t.Name(),
	})
	resetContentCache()
	t.Cleanup(func() {
		deleteI18nContentByID(t, ctx, enID)
		deleteI18nContentByID(t, ctx, zhID)
		resetContentCache()
	})

	output, err := New().GetContent(ctx, ContentLookupInput{
		BusinessType: businessType,
		BusinessID:   businessID,
		Field:        field,
		Locale:       EnglishLocale,
	})
	if err != nil {
		t.Fatalf("get content: %v", err)
	}
	if !output.Found {
		t.Fatal("expected content variant to be found")
	}
	if output.Content != "English Title" {
		t.Fatalf("expected english content %q, got %q", "English Title", output.Content)
	}
	if output.EffectiveLocale != EnglishLocale {
		t.Fatalf("expected effective locale %q, got %q", EnglishLocale, output.EffectiveLocale)
	}
	if output.ResolvedFallback {
		t.Fatal("expected requested locale to win without fallback")
	}
}

// TestGetContentFallsBackToDefaultLocale verifies that missing locale variants
// fall back to the runtime default locale.
func TestGetContentFallsBackToDefaultLocale(t *testing.T) {
	resetContentCache()

	ctx := context.Background()
	businessType := "notice"
	businessID := fmt.Sprintf("content-%s", t.Name())
	field := "summary"
	zhID := insertI18nContentForTest(t, ctx, do.SysI18NContent{
		BusinessType: businessType,
		BusinessId:   businessID,
		Field:        field,
		Locale:       DefaultLocale,
		ContentType:  string(ContentTypeMarkdown),
		Content:      "默认语言摘要",
		Status:       int(contentStatusEnabled),
		Remark:       t.Name(),
	})
	resetContentCache()
	t.Cleanup(func() {
		deleteI18nContentByID(t, ctx, zhID)
		resetContentCache()
	})

	output, err := New().GetContent(ctx, ContentLookupInput{
		BusinessType: businessType,
		BusinessID:   businessID,
		Field:        field,
		Locale:       EnglishLocale,
	})
	if err != nil {
		t.Fatalf("get content: %v", err)
	}
	if output.Content != "默认语言摘要" {
		t.Fatalf("expected default-locale fallback content %q, got %q", "默认语言摘要", output.Content)
	}
	if output.ContentType != ContentTypeMarkdown {
		t.Fatalf("expected content type %q, got %q", ContentTypeMarkdown, output.ContentType)
	}
	if output.EffectiveLocale != DefaultLocale {
		t.Fatalf("expected fallback locale %q, got %q", DefaultLocale, output.EffectiveLocale)
	}
	if !output.ResolvedFallback {
		t.Fatal("expected output to report default-locale fallback")
	}
}

// TestGetContentFallsBackToDefaultContent verifies that callers can provide a
// stable business-field fallback when no multilingual variant exists.
func TestGetContentFallsBackToDefaultContent(t *testing.T) {
	resetContentCache()

	output, err := New().GetContent(context.Background(), ContentLookupInput{
		BusinessType:   "notice",
		BusinessID:     fmt.Sprintf("content-%s", t.Name()),
		DefaultContent: "Original Title",
		Field:          "title",
		Locale:         EnglishLocale,
	})
	if err != nil {
		t.Fatalf("get content: %v", err)
	}
	if output.Found {
		t.Fatal("expected lookup to miss sys_i18n_content")
	}
	if !output.Defaulted {
		t.Fatal("expected caller fallback content to be used")
	}
	if output.Content != "Original Title" {
		t.Fatalf("expected caller fallback content %q, got %q", "Original Title", output.Content)
	}
}

// TestContentCacheRequiresInvalidationAfterWrite verifies that cached anchor
// lookups stay stable until callers explicitly invalidate the content cache.
func TestContentCacheRequiresInvalidationAfterWrite(t *testing.T) {
	resetContentCache()

	ctx := context.Background()
	svc := New()
	businessType := "notice"
	businessID := fmt.Sprintf("content-%s", t.Name())
	field := "body"

	firstOutput, err := svc.GetContent(ctx, ContentLookupInput{
		BusinessType:   businessType,
		BusinessID:     businessID,
		DefaultContent: "Original Body",
		Field:          field,
		Locale:         EnglishLocale,
	})
	if err != nil {
		t.Fatalf("first get content: %v", err)
	}
	if !firstOutput.Defaulted {
		t.Fatal("expected first lookup to use caller fallback content")
	}

	insertedID := insertI18nContentForTest(t, ctx, do.SysI18NContent{
		BusinessType: businessType,
		BusinessId:   businessID,
		Field:        field,
		Locale:       EnglishLocale,
		ContentType:  string(ContentTypeHTML),
		Content:      "<p>Localized Body</p>",
		Status:       int(contentStatusEnabled),
		Remark:       t.Name(),
	})
	t.Cleanup(func() {
		deleteI18nContentByID(t, ctx, insertedID)
		resetContentCache()
	})

	secondOutput, err := svc.GetContent(ctx, ContentLookupInput{
		BusinessType:   businessType,
		BusinessID:     businessID,
		DefaultContent: "Original Body",
		Field:          field,
		Locale:         EnglishLocale,
	})
	if err != nil {
		t.Fatalf("second get content: %v", err)
	}
	if secondOutput.Content != "Original Body" {
		t.Fatalf("expected cached fallback content %q, got %q", "Original Body", secondOutput.Content)
	}

	svc.InvalidateContentCache()
	thirdOutput, err := svc.GetContent(ctx, ContentLookupInput{
		BusinessType: businessType,
		BusinessID:   businessID,
		Field:        field,
		Locale:       EnglishLocale,
	})
	if err != nil {
		t.Fatalf("third get content: %v", err)
	}
	if thirdOutput.Content != "<p>Localized Body</p>" {
		t.Fatalf("expected invalidated cache to return stored content %q, got %q", "<p>Localized Body</p>", thirdOutput.Content)
	}
	if thirdOutput.ContentType != ContentTypeHTML {
		t.Fatalf("expected invalidated content type %q, got %q", ContentTypeHTML, thirdOutput.ContentType)
	}
}

// resetContentCache clears cached sys_i18n_content variants between tests.
func resetContentCache() {
	invalidateContentCache()
}

// insertI18nContentForTest inserts one sys_i18n_content row without touching caches.
func insertI18nContentForTest(t *testing.T, ctx context.Context, data do.SysI18NContent) uint64 {
	t.Helper()

	insertedID, err := dao.SysI18NContent.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert i18n content: %v", err)
	}
	return uint64(insertedID)
}

// deleteI18nContentByID removes one sys_i18n_content row by id.
func deleteI18nContentByID(t *testing.T, ctx context.Context, id uint64) {
	t.Helper()

	if _, err := dao.SysI18NContent.Ctx(ctx).Unscoped().Where(do.SysI18NContent{Id: id}).Delete(); err != nil {
		t.Fatalf("delete i18n content %d: %v", id, err)
	}
}
