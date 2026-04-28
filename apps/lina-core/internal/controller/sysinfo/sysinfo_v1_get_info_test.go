// This file verifies localized system-info response projections.

package sysinfo

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"
)

// TestFormatRunDurationUsesRuntimeLocale verifies uptime strings use runtime i18n resources.
func TestFormatRunDurationUsesRuntimeLocale(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}

	testCases := []struct {
		name     string
		locale   string
		seconds  int64
		expected string
	}{
		{name: "default locale hours", locale: i18nsvc.DefaultLocale, seconds: 3661, expected: "1小时1分钟1秒"},
		{name: "traditional locale minutes", locale: "zh-TW", seconds: 125, expected: "2分鐘5秒"},
		{name: "english locale seconds", locale: i18nsvc.EnglishLocale, seconds: 42, expected: "42 seconds"},
		{name: "english locale hours", locale: i18nsvc.EnglishLocale, seconds: 7322, expected: "2 hours 2 minutes 2 seconds"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(
				context.Background(),
				gctx.StrKey("BizCtx"),
				&model.Context{Locale: testCase.locale},
			)
			if actual := controller.formatRunDuration(ctx, testCase.seconds); actual != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}
