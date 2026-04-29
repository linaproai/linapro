// This file verifies login-log message localization behavior.

package loginlog

import (
	"context"
	"testing"

	"lina-core/pkg/pluginhost"
)

// fakeI18nService provides deterministic runtime translations for unit tests.
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

// TestTranslateLoginLogMessageResolvesStableReason verifies that login-log
// display messages are translated from stable auth lifecycle reason codes.
func TestTranslateLoginLogMessageResolvesStableReason(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		loginLogMessagePrefix + ".loginSuccessful": "登录成功",
	}}}

	actual := service.translateLoginLogMessage(context.Background(), pluginhost.AuthHookReasonLoginSuccessful)
	if actual != "登录成功" {
		t.Fatalf("expected stable reason to resolve, got %q", actual)
	}
}

// TestTranslateLoginLogMessagePreservesRawMessages verifies that custom raw
// audit messages are not interpreted through legacy text-to-key mappings.
func TestTranslateLoginLogMessagePreservesRawMessages(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.monitor-loginlog.logMessage.loginSuccessful": "登录成功",
	}}}

	actual := service.translateLoginLogMessage(context.Background(), "Login successful")
	if actual != "Login successful" {
		t.Fatalf("expected raw message to remain unchanged, got %q", actual)
	}
}

// TestExportHeadersUseRuntimeI18N verifies login-log export headers resolve
// through runtime i18n keys.
func TestExportHeadersUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.monitor-loginlog.fields.userName":  "用户账号",
		"plugin.monitor-loginlog.fields.status":    "登录状态",
		"plugin.monitor-loginlog.fields.ipAddress": "IP地址",
		"plugin.monitor-loginlog.fields.browser":   "浏览器",
		"plugin.monitor-loginlog.fields.os":        "操作系统",
		"plugin.monitor-loginlog.fields.message":   "消息",
		"plugin.monitor-loginlog.fields.loginTime": "登录时间",
	}}}

	actual := service.exportHeaders(context.Background())
	expected := []string{"用户账号", "登录状态", "IP地址", "浏览器", "操作系统", "消息", "登录时间"}
	for index, item := range expected {
		if actual[index] != item {
			t.Fatalf("expected header %d to be %q, got %q", index, item, actual[index])
		}
	}
}

// TestExportStatusTextUseRuntimeI18N verifies fallback login status labels are
// still resolved through dictionary runtime i18n keys.
func TestExportStatusTextUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"dict.sys_login_status.0.label": "成功",
		"dict.sys_login_status.1.label": "失败",
	}}}

	if actual := service.exportStatusText(context.Background(), LoginStatusSuccess, nil); actual != "成功" {
		t.Fatalf("expected success label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), LoginStatusFail, nil); actual != "失败" {
		t.Fatalf("expected failed label, got %q", actual)
	}
}
