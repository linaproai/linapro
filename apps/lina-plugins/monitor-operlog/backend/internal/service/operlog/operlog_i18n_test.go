// This file verifies operation-log runtime i18n projections.

package operlog

import (
	"context"
	"testing"

	"lina-plugin-monitor-operlog/backend/internal/model/operlogtype"
)

// fakeI18nService provides deterministic runtime translations for operation-log tests.
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

// TestExportHeadersUseRuntimeI18N verifies operation-log export headers resolve
// through runtime i18n keys.
func TestExportHeadersUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.monitor-operlog.fields.moduleName":     "模块名称",
		"plugin.monitor-operlog.fields.operSummary":    "操作摘要",
		"plugin.monitor-operlog.fields.operType":       "操作类型",
		"plugin.monitor-operlog.fields.operator":       "操作人员",
		"plugin.monitor-operlog.fields.requestMethod":  "请求方式",
		"plugin.monitor-operlog.fields.requestUrl":     "请求地址",
		"plugin.monitor-operlog.fields.ipAddress":      "IP地址",
		"plugin.monitor-operlog.fields.requestParams":  "请求参数",
		"plugin.monitor-operlog.fields.responseResult": "返回结果",
		"plugin.monitor-operlog.fields.operResult":     "操作结果",
		"plugin.monitor-operlog.fields.errorInfo":      "错误信息",
		"plugin.monitor-operlog.fields.durationMs":     "耗时(毫秒)",
		"plugin.monitor-operlog.fields.operTime":       "操作时间",
	}}}

	actual := service.exportHeaders(context.Background())
	expected := []string{
		"模块名称", "操作摘要", "操作类型", "操作人员", "请求方式", "请求地址", "IP地址",
		"请求参数", "返回结果", "操作结果", "错误信息", "耗时(毫秒)", "操作时间",
	}
	for index, item := range expected {
		if actual[index] != item {
			t.Fatalf("expected header %d to be %q, got %q", index, item, actual[index])
		}
	}
}

// TestExportTypeAndStatusTextUseRuntimeI18N verifies operation type and result
// export text resolves through dictionary runtime i18n keys.
func TestExportTypeAndStatusTextUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"dict.sys_oper_type.export.label": "导出",
		"dict.sys_oper_status.0.label":    "成功",
		"dict.sys_oper_status.1.label":    "失败",
	}}}

	if actual := service.exportOperTypeText(context.Background(), operlogtype.OperTypeExport.String(), nil); actual != "导出" {
		t.Fatalf("expected export type label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusSuccess, nil); actual != "成功" {
		t.Fatalf("expected success status label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusFail, nil); actual != "失败" {
		t.Fatalf("expected failure status label, got %q", actual)
	}
}
