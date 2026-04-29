// This file verifies organization post runtime i18n projections.

package post

import (
	"context"
	"testing"
)

// fakeI18nService provides deterministic runtime translations for post tests.
type fakeI18nService struct {
	messages map[string]string
}

// GetLocale returns the fixed test locale.
func (s fakeI18nService) GetLocale(_ context.Context) string {
	return "zh-CN"
}

// Translate resolves known test keys and otherwise returns the fallback text.
func (s fakeI18nService) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := s.messages[key]; ok {
		return value
	}
	return fallback
}

// FindMessageKeys is unused by these tests and returns no matches.
func (s fakeI18nService) FindMessageKeys(_ context.Context, _ string, _ string) []string {
	return []string{}
}

// TestExportHeadersUseRuntimeI18N verifies post export headers resolve through
// runtime i18n keys.
func TestExportHeadersUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		postExportHeaderCodeKey:      "岗位编码",
		postExportHeaderNameKey:      "岗位名称",
		postExportHeaderSortKey:      "排序",
		postExportHeaderStatusKey:    "状态",
		postExportHeaderRemarkKey:    "备注",
		postExportHeaderCreatedAtKey: "创建时间",
	}}}

	actual := service.exportHeaders(context.Background())
	expected := []string{"岗位编码", "岗位名称", "排序", "状态", "备注", "创建时间"}
	for index, item := range expected {
		if actual[index] != item {
			t.Fatalf("expected header %d to be %q, got %q", index, item, actual[index])
		}
	}
}

// TestExportStatusTextUseRuntimeI18N verifies post status export text resolves
// through runtime i18n keys.
func TestExportStatusTextUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		postExportStatusEnabledKey:  "正常",
		postExportStatusDisabledKey: "停用",
	}}}

	if actual := service.exportStatusText(context.Background(), 1); actual != "正常" {
		t.Fatalf("expected enabled label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), 0); actual != "停用" {
		t.Fatalf("expected disabled label, got %q", actual)
	}
}
