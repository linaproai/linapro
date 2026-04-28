// This file verifies structured runtime-message error localization.
package i18n

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginhost"
)

// TestLocalizeErrorSupportsStructuredRuntimeMessages verifies structured
// runtime-message errors render from the request locale with named parameters.
func TestLocalizeErrorSupportsStructuredRuntimeMessages(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	key := "test.structured." + pluginID
	code := bizerr.MustDefineWithKey(
		"TEST_STRUCTURED_ERROR",
		key,
		"User {username} does not exist",
		gcode.CodeNotFound,
	)
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"structured":{"%s":"用户 {username} 不存在"}}}`, pluginID),
		EnglishLocale: fmt.Sprintf(`{"test":{"structured":{"%s":"User {username} does not exist"}}}`, pluginID),
		"zh-TW":       fmt.Sprintf(`{"test":{"structured":{"%s":"使用者 {username} 不存在"}}}`, pluginID),
	})

	svc := New()
	testCases := []struct {
		locale   string
		expected string
	}{
		{locale: DefaultLocale, expected: "用户 alice 不存在"},
		{locale: EnglishLocale, expected: "User alice does not exist"},
		{locale: "zh-TW", expected: "使用者 alice 不存在"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.locale, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: testCase.locale})
			err := bizerr.NewCode(code, bizerr.P("username", "alice"))
			if actual := svc.LocalizeError(ctx, err); actual != testCase.expected {
				t.Fatalf("expected localized structured error %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestLocalizeErrorUsesStructuredFallback verifies missing runtime-message keys
// render the English fallback with named parameters instead of leaking the key.
func TestLocalizeErrorUsesStructuredFallback(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})
	code := bizerr.MustDefineWithKey(
		"TEST_STRUCTURED_MISSING_KEY",
		"test.structured.missingKey",
		"User {username} does not exist",
		gcode.CodeNotFound,
	)
	err := bizerr.NewCode(code, bizerr.P("username", "alice"))
	if actual := svc.LocalizeError(ctx, err); actual != "User alice does not exist" {
		t.Fatalf("expected structured fallback %q, got %q", "User alice does not exist", actual)
	}
}

// TestLocalizeErrorUsesRuntimeBundleCache verifies structured-error rendering
// reads through the normal runtime bundle cache and does not require a bespoke
// per-error catalog build path.
func TestLocalizeErrorUsesRuntimeBundleCache(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	key := "test.structured.cache." + pluginID
	code := bizerr.MustDefineWithKey(
		"TEST_STRUCTURED_CACHE",
		key,
		"Fallback {value}",
		gcode.CodeInvalidParameter,
	)
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"structured":{"cache":{"%s":"Cached {value}"}}}}`,
			pluginID,
		))},
	})
	pluginhost.RegisterSourcePlugin(plugin)
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})
	err := bizerr.NewCode(code, bizerr.P("value", "message"))
	if actual := svc.LocalizeError(ctx, err); actual != "Cached message" {
		t.Fatalf("expected cached runtime bundle translation %q, got %q", "Cached message", actual)
	}
}
