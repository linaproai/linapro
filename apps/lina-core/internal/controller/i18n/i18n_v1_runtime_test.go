// This file verifies the runtime i18n controller endpoints.

package i18n

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	v1 "lina-core/api/i18n/v1"
	"lina-core/internal/model"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	middlewaresvc "lina-core/internal/service/middleware"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	_ "lina-core/pkg/dbdriver"
)

// newTestI18nService creates a standalone i18n service for tests.
func newTestI18nService() i18nsvc.Service {
	configSvc := hostconfig.New()
	bizCtxSvc := bizctx.New()
	clusterSvc := cluster.New(configSvc.GetCluster(context.Background()))
	return i18nsvc.New(bizCtxSvc, configSvc, cachecoord.Default(clusterSvc))
}

// TestRuntimeMessagesUsesExplicitLangOverride verifies that the runtime
// messages endpoint honors the explicit lang query parameter.
func TestRuntimeMessagesUsesExplicitLangOverride(t *testing.T) {
	t.Parallel()

	i18nSvc := newTestI18nService()
	controller := &ControllerV1{
		localeResolver: i18nSvc,
		bundleProvider: i18nSvc,
		maintainer:     i18nSvc,
	}
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

	i18nSvc := newTestI18nService()
	controller := &ControllerV1{
		localeResolver: i18nSvc,
		bundleProvider: i18nSvc,
		maintainer:     i18nSvc,
	}
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
	if !res.Enabled {
		t.Fatal("expected runtime locale switch to be enabled by default")
	}
	expectedItems := i18nSvc.ListRuntimeLocales(ctx, i18nsvc.EnglishLocale)
	if len(res.Items) != len(expectedItems) {
		t.Fatalf("expected %d locale descriptors, got %d", len(expectedItems), len(res.Items))
	}
	for _, expected := range expectedItems {
		actual, ok := findRuntimeLocale(res.Items, expected.Locale)
		if !ok {
			t.Fatalf("expected locale %q in runtime locale list", expected.Locale)
		}
		if actual.Name != expected.Name ||
			actual.NativeName != expected.NativeName ||
			actual.Direction != expected.Direction ||
			actual.IsDefault != expected.IsDefault {
			t.Fatalf("unexpected locale descriptor for %s: got=%+v expected=%+v", expected.Locale, actual, expected)
		}
	}
}

// TestRuntimeLocalesReturnsDisabledDefaultOnly verifies the controller exposes
// disabled language-switch state and a default-only descriptor list.
func TestRuntimeLocalesReturnsDisabledDefaultOnly(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{
		localeResolver: disabledRuntimeLocaleService{},
		bundleProvider: disabledRuntimeLocaleService{},
	}

	res, err := controller.RuntimeLocales(context.Background(), &v1.RuntimeLocalesReq{Lang: i18nsvc.EnglishLocale})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Locale != i18nsvc.DefaultLocale {
		t.Fatalf("expected disabled runtime locale response locale %q, got %q", i18nsvc.DefaultLocale, res.Locale)
	}
	if res.Enabled {
		t.Fatal("expected disabled runtime locale response to report enabled=false")
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected only one default locale item, got %d", len(res.Items))
	}
	if res.Items[0].Locale != i18nsvc.DefaultLocale || !res.Items[0].IsDefault {
		t.Fatalf("expected default-only locale item, got %+v", res.Items[0])
	}
	if res.Items[0].Direction != i18nsvc.LocaleDirectionLTR.String() {
		t.Fatalf("expected fixed LTR direction, got %q", res.Items[0].Direction)
	}
}

// disabledRuntimeLocaleService is a narrow fake for controller disabled-i18n
// response-shape tests.
type disabledRuntimeLocaleService struct{}

const (
	// disabledRuntimeBundleFingerprint is a stable fake fingerprint for tests.
	disabledRuntimeBundleFingerprint = "00000000000000000000000000000000"
)

// ResolveRequestLocale always returns the configured default locale.
func (disabledRuntimeLocaleService) ResolveRequestLocale(_ *ghttp.Request) string {
	return i18nsvc.DefaultLocale
}

// ResolveLocale ignores explicit overrides and returns the configured default locale.
func (disabledRuntimeLocaleService) ResolveLocale(_ context.Context, _ string) string {
	return i18nsvc.DefaultLocale
}

// GetLocale always returns the configured default locale.
func (disabledRuntimeLocaleService) GetLocale(_ context.Context) string {
	return i18nsvc.DefaultLocale
}

// EnsureRuntimeBundleCacheFresh is a no-op for the disabled-locale fake.
func (disabledRuntimeLocaleService) EnsureRuntimeBundleCacheFresh(_ context.Context) error {
	return nil
}

// BundleRevision returns a stable fake bundle revision.
func (disabledRuntimeLocaleService) BundleRevision(_ string) i18nsvc.RuntimeBundleRevision {
	return i18nsvc.RuntimeBundleRevision{
		Version:     1,
		Fingerprint: disabledRuntimeBundleFingerprint,
	}
}

// BundleVersion returns a stable fake bundle version.
func (disabledRuntimeLocaleService) BundleVersion(_ string) uint64 {
	return 1
}

// ListRuntimeLocales returns only the default locale descriptor.
func (disabledRuntimeLocaleService) ListRuntimeLocales(_ context.Context, _ string) []i18nsvc.LocaleDescriptor {
	return []i18nsvc.LocaleDescriptor{
		{
			Locale:     i18nsvc.DefaultLocale,
			Name:       "简体中文",
			NativeName: "简体中文",
			Direction:  i18nsvc.LocaleDirectionLTR.String(),
			IsDefault:  true,
		},
	}
}

// IsMultiLanguageEnabled reports disabled runtime language switching.
func (disabledRuntimeLocaleService) IsMultiLanguageEnabled(_ context.Context) bool {
	return false
}

// BuildRuntimeMessages returns an empty fake runtime bundle.
func (disabledRuntimeLocaleService) BuildRuntimeMessages(_ context.Context, _ string) map[string]interface{} {
	return map[string]interface{}{}
}

// countingRuntimeLocaleService records runtime bundle builds for ETag tests.
type countingRuntimeLocaleService struct {
	disabledRuntimeLocaleService

	revision   i18nsvc.RuntimeBundleRevision
	buildCount atomic.Int64
}

// BundleRevision returns the configured fake runtime bundle revision.
func (s *countingRuntimeLocaleService) BundleRevision(_ string) i18nsvc.RuntimeBundleRevision {
	return s.revision
}

// BundleVersion returns the configured fake runtime bundle version.
func (s *countingRuntimeLocaleService) BundleVersion(_ string) uint64 {
	return s.revision.Version
}

// BuildRuntimeMessages records that a full response body was built.
func (s *countingRuntimeLocaleService) BuildRuntimeMessages(_ context.Context, _ string) map[string]interface{} {
	s.buildCount.Add(1)
	return map[string]interface{}{
		"app": map[string]interface{}{
			"sample": map[string]interface{}{
				"title": "Sample",
			},
		},
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

// TestBuildRuntimeMessagesETagUsesBundleRevision verifies that the strong ETag
// formats the cache-owned version and content fingerprint.
func TestBuildRuntimeMessagesETagUsesBundleRevision(t *testing.T) {
	t.Parallel()

	revision := i18nsvc.RuntimeBundleRevision{
		Version:     42,
		Fingerprint: "0123456789abcdef0123456789abcdef",
	}
	got, ok := buildRuntimeMessagesETag(i18nsvc.EnglishLocale, revision)
	if !ok {
		t.Fatal("expected non-empty fingerprint to produce an ETag")
	}
	if got != `"en-US-42-0123456789abcdef0123456789abcdef"` {
		t.Fatalf("expected ETag to include locale, version and fingerprint, got %q", got)
	}
	if empty, ok := buildRuntimeMessagesETag(i18nsvc.EnglishLocale, i18nsvc.RuntimeBundleRevision{}); ok || empty != "" {
		t.Fatalf("expected empty fingerprint to skip ETag, got ok=%v value=%q", ok, empty)
	}
}

// TestMatchesIfNoneMatchAcceptsExactWildcardAndMultiValues verifies the
// If-None-Match matcher honors RFC 7232 semantics: exact match, the `*` wildcard,
// and comma-separated candidate lists.
func TestMatchesIfNoneMatchAcceptsExactWildcardAndMultiValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		headerValue string
		etag        string
		shouldMatch bool
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
	firstRequest, err := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	if err != nil {
		t.Fatalf("create first request: %v", err)
	}
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
	secondRequest, err := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	if err != nil {
		t.Fatalf("create second request: %v", err)
	}
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

	// Invalidate the host sector so the bundle version advances; the same
	// If-None-Match should now miss and a fresh 200 must arrive.
	newTestI18nService().InvalidateRuntimeBundleCache(i18nsvc.InvalidateScope{
		Locales: []string{i18nsvc.EnglishLocale},
		Sectors: []i18nsvc.Sector{i18nsvc.SectorHost},
	})

	thirdRequest, err := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.EnglishLocale, nil)
	if err != nil {
		t.Fatalf("create third request: %v", err)
	}
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

// TestRuntimeMessagesWarmETagSkipsBundleBuild proves the 304 warm-cache path
// does not build the nested runtime message response body.
func TestRuntimeMessagesWarmETagSkipsBundleBuild(t *testing.T) {
	bundleSvc := &countingRuntimeLocaleService{
		revision: i18nsvc.RuntimeBundleRevision{
			Version:     7,
			Fingerprint: "abcdef0123456789abcdef0123456789",
		},
	}
	controller := &ControllerV1{
		localeResolver: bundleSvc,
		bundleProvider: bundleSvc,
	}
	address := startRuntimeMessagesControllerTestServer(t, controller)
	etag, ok := buildRuntimeMessagesETag(i18nsvc.DefaultLocale, bundleSvc.revision)
	if !ok {
		t.Fatal("expected fake revision to produce an ETag")
	}

	request, err := http.NewRequest(http.MethodGet, address+"/i18n/runtime/messages?lang="+i18nsvc.DefaultLocale, nil)
	if err != nil {
		t.Fatalf("create warm-cache request: %v", err)
	}
	request.Header.Set(runtimeMessagesIfNoneMatchHeader, etag)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("warm-cache request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotModified {
		t.Fatalf("expected warm-cache request status 304, got %d", response.StatusCode)
	}
	if response.Header.Get(runtimeMessagesETagHeader) != etag {
		t.Fatalf("expected 304 to echo ETag %q, got %q", etag, response.Header.Get(runtimeMessagesETagHeader))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read warm-cache body: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("expected empty body on warm-cache 304, got %d bytes", len(body))
	}
	if got := bundleSvc.buildCount.Load(); got != 0 {
		t.Fatalf("expected warm-cache 304 to skip BuildRuntimeMessages, got %d calls", got)
	}
}

// startRuntimeMessagesTestServer wires the runtime i18n controller with the
// host response middleware on a randomly chosen port and returns the base URL.
func startRuntimeMessagesTestServer(t *testing.T) string {
	t.Helper()
	return startRuntimeMessagesControllerTestServer(t, NewV1(newTestI18nService()))
}

// startRuntimeMessagesControllerTestServer wires a supplied runtime i18n
// controller with the host response middleware for endpoint-level tests.
func startRuntimeMessagesControllerTestServer(t *testing.T, controller any) string {
	t.Helper()

	serverName := "i18n-runtime-test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	server := ghttp.GetServer(serverName)
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	middlewareSvc := newRuntimeMessagesTestMiddleware()
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(middlewareSvc.Response)
		group.Bind(controller)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("start runtime messages test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown runtime messages test server: %v", err)
		}
	})

	listenedPort := server.GetListenedPort()
	if listenedPort <= 0 {
		t.Fatal("expected randomly allocated port to be positive")
	}
	return "http://127.0.0.1:" + strconv.Itoa(listenedPort)
}

// newRuntimeMessagesTestMiddleware constructs response middleware with
// explicit dependencies so controller tests do not rely on disabled defaults.
func newRuntimeMessagesTestMiddleware() middlewaresvc.Service {
	configSvc := hostconfig.New()
	bizCtxSvc := bizctx.New()
	i18nSvc := newTestI18nService()
	cacheCoordSvc := cachecoord.Default(nil)
	pluginSvc := pluginsvc.New(nil, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, session.NewDBStore())
	orgCapSvc := orgcap.New(pluginSvc)
	tenantSvc := tenantcapsvc.New(pluginSvc, nil)
	roleSvc := role.New(pluginSvc, bizCtxSvc, configSvc, i18nSvc, nil, orgCapSvc, tenantSvc)
	roleSvc.SetDataScopeService(datascope.New(bizCtxSvc, roleSvc, orgCapSvc))
	authSvc := auth.New(configSvc, pluginSvc, orgCapSvc, roleSvc, tenantSvc, session.NewDBStore(), kvcache.New())
	return middlewaresvc.New(authSvc, bizCtxSvc, configSvc, i18nSvc, pluginSvc, roleSvc, tenantSvc)
}
