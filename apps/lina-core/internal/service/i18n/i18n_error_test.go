// This file verifies structured runtime-message error localization.
package i18n

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

// TestLocalizeErrorUsesHostDataScopeErrorResources verifies every host
// data-permission error introduced for governed resources has shipped runtime
// translations in all built-in locales.
func TestLocalizeErrorUsesHostDataScopeErrorResources(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	svc := New()
	testCases := []struct {
		name     string
		key      string
		fallback string
		params   []bizerr.Param
		expected map[string]string
	}{
		{
			name:     "shared denied",
			key:      "error.datascope.denied",
			fallback: "Data is outside the current data permission scope",
			expected: map[string]string{
				DefaultLocale: "数据不在当前数据权限范围内",
				EnglishLocale: "Data is outside the current data permission scope",
				"zh-TW":       "數據不在當前數據權限範圍內",
			},
		},
		{
			name:     "shared unauthenticated",
			key:      "error.datascope.not.authenticated",
			fallback: "Not signed in",
			expected: map[string]string{
				DefaultLocale: "请先登录",
				EnglishLocale: "Not signed in",
				"zh-TW":       "請先登錄",
			},
		},
		{
			name:     "shared unsupported",
			key:      "error.datascope.unsupported",
			fallback: "Unsupported data permission scope: {scope}",
			params:   []bizerr.Param{bizerr.P("scope", 9)},
			expected: map[string]string{
				DefaultLocale: "不支持的数据权限范围：9",
				EnglishLocale: "Unsupported data permission scope: 9",
				"zh-TW":       "不支持的數據權限範圍：9",
			},
		},
		{
			name:     "user denied",
			key:      "error.user.data.scope.denied",
			fallback: "User data is outside the current data permission scope",
			expected: map[string]string{
				DefaultLocale: "用户数据超出当前数据权限范围",
				EnglishLocale: "User data is outside the current data permission scope",
				"zh-TW":       "用戶數據超出當前數據權限範圍",
			},
		},
		{
			name:     "file denied",
			key:      "error.file.data.scope.denied",
			fallback: "File data is outside the current data permission scope",
			expected: map[string]string{
				DefaultLocale: "文件数据不在当前数据权限范围内",
				EnglishLocale: "File data is outside the current data permission scope",
				"zh-TW":       "文件數據不在當前數據權限範圍內",
			},
		},
		{
			name:     "job denied",
			key:      "error.job.data.scope.denied",
			fallback: "Scheduled job data is outside the current data permission scope",
			expected: map[string]string{
				DefaultLocale: "定时任务数据不在当前数据权限范围内",
				EnglishLocale: "Scheduled job data is outside the current data permission scope",
				"zh-TW":       "定時任務數據不在當前數據權限範圍內",
			},
		},
		{
			name:     "role dept unavailable",
			key:      "error.role.data.scope.dept.unavailable",
			fallback: "Department data scope requires the organization management plugin to be enabled",
			expected: map[string]string{
				DefaultLocale: "本部门数据权限需要先启用组织管理插件",
				EnglishLocale: "Department data scope requires the organization management plugin to be enabled",
				"zh-TW":       "本部門數據權限需要先啟用組織管理插件",
			},
		},
	}

	for index, testCase := range testCases {
		testCase := testCase
		index := index
		t.Run(testCase.name, func(t *testing.T) {
			code := bizerr.MustDefineWithKey(
				fmt.Sprintf("TEST_HOST_DATASCOPE_ERROR_%d", index),
				testCase.key,
				testCase.fallback,
				gcode.CodeInvalidParameter,
			)
			for locale, expected := range testCase.expected {
				ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: locale})
				err := bizerr.NewCode(code, testCase.params...)
				if actual := svc.LocalizeError(ctx, err); actual != expected {
					t.Fatalf("expected %s localized error %q, got %q", locale, expected, actual)
				}
			}
		})
	}
}

// TestLocalizeErrorUsesRealPluginErrorResources verifies representative source
// plugin business errors render through the shipped plugin runtime i18n files.
func TestLocalizeErrorUsesRealPluginErrorResources(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	repoRoot := findRepoRootForI18NTest(t)
	pluginDirs := []string{
		"content-notice",
		"org-center",
		"monitor-loginlog",
		"monitor-operlog",
		"plugin-demo-source",
		"plugin-demo-dynamic",
	}
	for _, pluginDir := range pluginDirs {
		registerSourcePluginDirectoryI18N(t, repoRoot, pluginDir)
	}

	svc := New()
	testCases := []struct {
		name     string
		key      string
		fallback string
		params   []bizerr.Param
		expected map[string]string
	}{
		{
			name:     "content notice not found",
			key:      "error.content.notice.not.found",
			fallback: "Notice does not exist",
			expected: map[string]string{
				DefaultLocale: "通知公告不存在",
				EnglishLocale: "Notice does not exist",
				"zh-TW":       "通知公告不存在",
			},
		},
		{
			name:     "org department not found",
			key:      "error.org.dept.not.found",
			fallback: "Department does not exist",
			expected: map[string]string{
				DefaultLocale: "部门不存在",
				EnglishLocale: "Department does not exist",
				"zh-TW":       "部門不存在",
			},
		},
		{
			name:     "org post assigned",
			key:      "error.org.post.assigned.delete.denied",
			fallback: "Post {id} has assigned users and cannot be deleted",
			params:   []bizerr.Param{bizerr.P("id", 17)},
			expected: map[string]string{
				DefaultLocale: "岗位ID 17 已分配给用户，不能删除",
				EnglishLocale: "Post 17 has assigned users and cannot be deleted",
				"zh-TW":       "崗位ID 17 已分配給用戶，不能刪除",
			},
		},
		{
			name:     "login log not found",
			key:      "error.monitor.loginlog.not.found",
			fallback: "Login log does not exist",
			expected: map[string]string{
				DefaultLocale: "登录日志不存在",
				EnglishLocale: "Login log does not exist",
				"zh-TW":       "登錄日誌不存在",
			},
		},
		{
			name:     "operation log not found",
			key:      "error.monitor.operlog.not.found",
			fallback: "Operation log does not exist",
			expected: map[string]string{
				DefaultLocale: "操作日志不存在",
				EnglishLocale: "Operation log does not exist",
				"zh-TW":       "操作日誌不存在",
			},
		},
		{
			name:     "source demo attachment size",
			key:      "error.plugin.demo.source.attachment.size.too.large",
			fallback: "Attachment size must not exceed {maxSizeMB}MB",
			params:   []bizerr.Param{bizerr.P("maxSizeMB", 5)},
			expected: map[string]string{
				DefaultLocale: "附件大小不能超过5MB",
				EnglishLocale: "Attachment size must not exceed 5MB",
				"zh-TW":       "附件大小不能超過5MB",
			},
		},
		{
			name:     "dynamic demo title length",
			key:      "error.plugin.demo.dynamic.record.title.too.long",
			fallback: "Record title must not exceed {maxChars} characters",
			params:   []bizerr.Param{bizerr.P("maxChars", 128)},
			expected: map[string]string{
				DefaultLocale: "记录标题长度不能超过128个字符",
				EnglishLocale: "Record title must not exceed 128 characters",
				"zh-TW":       "記錄標題長度不能超過128個字符",
			},
		},
	}

	for index, testCase := range testCases {
		testCase := testCase
		index := index
		t.Run(testCase.name, func(t *testing.T) {
			code := bizerr.MustDefineWithKey(
				fmt.Sprintf("TEST_PLUGIN_ERROR_%d", index),
				testCase.key,
				testCase.fallback,
				gcode.CodeInvalidParameter,
			)
			for locale, expected := range testCase.expected {
				ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: locale})
				err := bizerr.NewCode(code, testCase.params...)
				if actual := svc.LocalizeError(ctx, err); actual != expected {
					t.Fatalf("expected %s localized error %q, got %q", locale, expected, actual)
				}
			}
		})
	}
}

// findRepoRootForI18NTest resolves the repository root for tests that need
// actual source-plugin manifest resources.
func findRepoRootForI18NTest(t *testing.T) string {
	t.Helper()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}
	current := workingDir
	for {
		if _, statErr := os.Stat(filepath.Join(current, "apps", "lina-core", "go.mod")); statErr == nil {
			if _, pluginErr := os.Stat(filepath.Join(current, "apps", "lina-plugins")); pluginErr == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatalf("repository root not found from %s", workingDir)
		}
		current = parent
	}
}

// registerSourcePluginDirectoryI18N registers one source plugin backed by the
// plugin's real manifest directory from the repository checkout.
func registerSourcePluginDirectoryI18N(t *testing.T, repoRoot string, pluginDir string) {
	t.Helper()

	pluginPath := filepath.Join(repoRoot, "apps", "lina-plugins", pluginDir)
	if _, err := os.Stat(filepath.Join(pluginPath, "manifest", "i18n")); err != nil {
		t.Fatalf("plugin i18n directory missing for %s: %v", pluginDir, err)
	}
	pluginID := nextTestSourcePluginID() + "-" + pluginDir
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(os.DirFS(pluginPath))
	pluginhost.RegisterSourcePlugin(plugin)
	resetRuntimeBundleCache()
}
