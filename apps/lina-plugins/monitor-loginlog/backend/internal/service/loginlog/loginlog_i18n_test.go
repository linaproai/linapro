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
