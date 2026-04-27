// This file verifies the runtime i18n controller endpoints.

package i18n

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "lina-core/api/i18n/v1"
	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"
	middlewaresvc "lina-core/internal/service/middleware"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

// TestRuntimeMessagesUsesExplicitLangOverride verifies that the runtime
// messages endpoint honors the explicit lang query parameter.
func TestRuntimeMessagesUsesExplicitLangOverride(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}
	ctx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.DefaultLocale},
	)

	res, err := controller.RuntimeMessages(ctx, &v1.RuntimeMessagesReq{Lang: i18nsvc.EnglishLocale})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Locale != i18nsvc.EnglishLocale {
		t.Fatalf("expected runtime locale %q, got %q", i18nsvc.EnglishLocale, res.Locale)
	}

	actual, ok := lookupRuntimeMessage(res.Messages, "menu.dashboard.title")
	if !ok {
		t.Fatal("expected menu.dashboard.title to exist in runtime messages")
	}
	if actual != "Dashboard" {
		t.Fatalf("expected English runtime message %q, got %q", "Dashboard", actual)
	}
}

// TestRuntimeLocalesReturnsLocalizedDescriptors verifies that the runtime
// locale endpoint returns localized display names with stable native names.
func TestRuntimeLocalesReturnsLocalizedDescriptors(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}
	ctx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.DefaultLocale},
	)

	res, err := controller.RuntimeLocales(ctx, &v1.RuntimeLocalesReq{Lang: i18nsvc.EnglishLocale})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Locale != i18nsvc.EnglishLocale {
		t.Fatalf("expected runtime locale %q, got %q", i18nsvc.EnglishLocale, res.Locale)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 locale descriptors, got %d", len(res.Items))
	}

	zhLocale, ok := findRuntimeLocale(res.Items, i18nsvc.DefaultLocale)
	if !ok {
		t.Fatalf("expected locale %q in runtime locale list", i18nsvc.DefaultLocale)
	}
	if zhLocale.Name != "Chinese (Simplified)" || zhLocale.NativeName != "简体中文" || !zhLocale.IsDefault {
		t.Fatalf("unexpected zh-CN locale descriptor: %+v", zhLocale)
	}
}

// lookupRuntimeMessage reads one dotted runtime message path from the nested response payload.
func lookupRuntimeMessage(messages map[string]interface{}, key string) (string, bool) {
	current := interface{}(messages)
	for _, segment := range strings.Split(strings.TrimSpace(key), ".") {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		current, ok = currentMap[segment]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}

// findRuntimeLocale locates one locale descriptor by locale code.
func findRuntimeLocale(items []v1.RuntimeLocaleItem, locale string) (v1.RuntimeLocaleItem, bool) {
	for _, item := range items {
		if item.Locale == locale {
			return item, true
		}
	}
	return v1.RuntimeLocaleItem{}, false
}

// TestBuildRuntimeMessagesETagFormatsLocaleAndVersion verifies that the strong
// ETag format used by the runtime messages endpoint is `"<locale>-<version>"`.
func TestBuildRuntimeMessagesETagFormatsLocaleAndVersion(t *testing.T) {
	t.Parallel()

	got := buildRuntimeMessagesETag(i18nsvc.EnglishLocale, 42)
	if got != `"en-US-42"` {
		t.Fatalf("expected ETag %q, got %q", `"en-US-42"`, got)
	}

	got = buildRuntimeMessagesETag(i18nsvc.DefaultLocale, 0)
	if got != `"zh-CN-0"` {
		t.Fatalf("expected ETag %q, got %q", `"zh-CN-0"`, got)
	}
}

// TestMatchesIfNoneMatchAcceptsExactWildcardAndMultiValues verifies the
// If-None-Match matcher honors RFC 7232 semantics: exact match, the `*` wildcard,
// and comma-separated candidate lists.
func TestMatchesIfNoneMatchAcceptsExactWildcardAndMultiValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		headerValue  string
		etag         string
		shouldMatch  bool
	}{
		{name: "empty header", headerValue: "", etag: `"en-US-1"`, shouldMatch: false},
		{name: "exact match", headerValue: `"en-US-1"`, etag: `"en-US-1"`, shouldMatch: true},
		{name: "version mismatch", headerValue: `"en-US-1"`, etag: `"en-US-2"`, shouldMatch: false},
		{name: "wildcard", headerValue: "*", etag: `"en-US-1"`, shouldMatch: true},
		{name: "multi-value with match", headerValue: `"old", "en-US-1"`, etag: `"en-US-1"`, shouldMatch: true},
		{name: "multi-value without match", headerValue: `"old", "older"`, etag: `"en-US-1"`, shouldMatch: false},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if matches := matchesIfNoneMatch(testCase.headerValue, testCase.etag); matches != testCase.shouldMatch {
				t.Fatalf("expected matches=%v, got %v", testCase.shouldMatch, matches)
			}
		})
	}
}

// TestRuntimeMessagesEmitsETagAndShortCircuits304 verifies the runtime messages
// endpoint over a real HTTP cycle: first request returns the bundle plus an
// ETag, a follow-up If-None-Match request returns 304 with no body, and after a
// scoped invalidation the version increments and a fresh 200 is served again.
func TestRuntimeMessagesEmitsETagAndShortCircuits304(t *testing.T) {
	address := startRuntimeMessagesTestServer(t)

	// First request: server emits ETag and a 200 response.
	firstRequest, _ := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	firstResponse, err := http.DefaultClient.Do(firstRequest)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	defer firstResponse.Body.Close()
	if firstResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", firstResponse.StatusCode)
	}
	etag := firstResponse.Header.Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header on first response")
	}
	if cacheControl := firstResponse.Header.Get("Cache-Control"); cacheControl != "private, must-revalidate" {
		t.Fatalf("expected Cache-Control %q, got %q", "private, must-revalidate", cacheControl)
	}
	body, err := io.ReadAll(firstResponse.Body)
	if err != nil {
		t.Fatalf("read first body: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected first response body to contain the bundle JSON")
	}

	// Second request: matching If-None-Match returns 304 with no body.
	secondRequest, _ := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	secondRequest.Header.Set("If-None-Match", etag)
	secondResponse, err := http.DefaultClient.Do(secondRequest)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	defer secondResponse.Body.Close()
	if secondResponse.StatusCode != http.StatusNotModified {
		t.Fatalf("expected second request status 304, got %d", secondResponse.StatusCode)
	}
	if secondResponse.Header.Get("ETag") != etag {
		t.Fatalf("expected 304 to echo the same ETag %q, got %q", etag, secondResponse.Header.Get("ETag"))
	}
	secondBody, err := io.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatalf("read second body: %v", err)
	}
	if len(secondBody) != 0 {
		t.Fatalf("expected empty body on 304, got %d bytes: %s", len(secondBody), string(secondBody))
	}

	// Invalidate the database sector so the bundle version advances; the same
	// If-None-Match should now miss and a fresh 200 must arrive.
	i18nsvc.New().InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
		Locales: []string{i18nsvc.EnglishLocale},
		Sectors: []i18nsvc.Sector{i18nsvc.SectorDatabase},
	})

	thirdRequest, _ := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	thirdRequest.Header.Set("If-None-Match", etag)
	thirdResponse, err := http.DefaultClient.Do(thirdRequest)
	if err != nil {
		t.Fatalf("third request: %v", err)
	}
	defer thirdResponse.Body.Close()
	if thirdResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected post-invalidation request to return 200, got %d", thirdResponse.StatusCode)
	}
	freshETag := thirdResponse.Header.Get("ETag")
	if freshETag == "" || freshETag == etag {
		t.Fatalf("expected fresh ETag distinct from %q, got %q", etag, freshETag)
	}
}

// startRuntimeMessagesTestServer wires the runtime i18n controller with the
// host response middleware on a randomly chosen port and returns the base URL.
func startRuntimeMessagesTestServer(t *testing.T) string {
	t.Helper()

	serverName := "i18n-runtime-test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	server := ghttp.GetServer(serverName)
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	middlewareSvc := middlewaresvc.New()
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(middlewareSvc.Response)
		group.Bind(NewV1())
	})

	if err := server.Start(); err != nil {
		t.Fatalf("start runtime messages test server: %v", err)
	}
	t.Cleanup(func() {
		_ = server.Shutdown()
	})

	listenedPort := server.GetListenedPort()
	if listenedPort <= 0 {
		t.Fatal("expected randomly allocated port to be positive")
	}
	return "http://127.0.0.1:" + strconv.Itoa(listenedPort)
}
