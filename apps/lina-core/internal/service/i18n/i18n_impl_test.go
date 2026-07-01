// This file verifies resource-backed runtime messages and benchmarks runtime
// translation paths without relying on database-backed i18n override tables.

package i18n

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/config"
	"lina-core/pkg/plugin/pluginhost"
)

var testSourcePluginSequence atomic.Uint64

var (
	benchmarkRuntimeMessagesSink map[string]interface{}
	benchmarkTranslateSink       string
)

// nextTestSourcePluginID returns a unique source-plugin fixture ID.
func nextTestSourcePluginID() string {
	return fmt.Sprintf("plugin-i18n-test-%d", testSourcePluginSequence.Add(1))
}

// registerTestSourcePluginI18N registers a source-plugin fixture with the
// provided locale JSON resources and invalidates the runtime cache.
func registerTestSourcePluginI18N(t *testing.T, pluginID string, localeFiles map[string]string) {
	t.Helper()

	fileSystem := fstest.MapFS{}
	fileSystem["plugin.yaml"] = &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(pluginID, true))}
	for locale, content := range localeFiles {
		normalizedLocale := normalizeLocale(locale)
		if normalizedLocale == "" {
			t.Fatalf("invalid locale in test fixture: %q", locale)
		}
		fileSystem["manifest/i18n/"+normalizedLocale+"/plugin.json"] = &fstest.MapFile{Data: []byte(content)}
	}

	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(fileSystem)
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}
	resetRuntimeBundleCache()
}

// sourcePluginI18NManifestFixture builds the minimal embedded plugin.yaml used
// by runtime i18n tests that register in-memory source plugins.
func sourcePluginI18NManifestFixture(pluginID string, enabled bool) string {
	enabledValue := "false"
	if enabled {
		enabledValue = "true"
	}
	return fmt.Sprintf(`id: %s
name: Runtime I18N Test Plugin
version: v0.1.0
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
i18n:
  enabled: %s
  default: zh-CN
  locales:
    - locale: zh-CN
      nativeName: 简体中文
    - locale: en-US
      nativeName: English
`, pluginID, enabledValue)
}

// TestRuntimeTranslationsDoNotImplicitlyUseDefaultLocale verifies that current
// locale translation never returns default-locale text unless the caller uses a
// default-locale context.
func TestRuntimeTranslationsDoNotImplicitlyUseDefaultLocale(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	pluginID := nextTestSourcePluginID()
	registerTestSourcePluginI18N(t, pluginID, map[string]string{
		DefaultLocale: fmt.Sprintf(`{"test":{"strict":{"%s":{"title":"仅默认语言提供"}}}}`, pluginID),
	})
	key := "test.strict." + pluginID + ".title"

	var (
		ctx      = context.Background()
		svc      = New(bizctx.New(), config.New(), cachecoord.Default(nil))
		messages = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	)
	if value, ok := lookupMessageString(messages, key); ok {
		t.Fatalf("expected en-US runtime messages to omit default-locale-only key, got %q", value)
	}

	enCtx := context.WithValue(ctx, gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})
	if actual := svc.Translate(enCtx, key, "Source Fallback"); actual != "Source Fallback" {
		t.Fatalf("expected Translate to use caller fallback, got %q", actual)
	}
	if actual := svc.Translate(context.Background(), key, "fallback"); actual != "仅默认语言提供" {
		t.Fatalf("expected default-locale context to return %q, got %q", "仅默认语言提供", actual)
	}
}

// TestBuildRuntimeMessagesHonorsSourcePluginI18NPolicy verifies runtime UI
// translations follow the same plugin.yaml opt-in rule as apidoc resources.
func TestBuildRuntimeMessagesHonorsSourcePluginI18NPolicy(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	managedPluginID := nextTestSourcePluginID()
	managedPlugin := pluginhost.NewDeclarations(managedPluginID)
	managedPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(managedPluginID, true))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Runtime Managed"}}}`,
			managedPluginID,
		))},
	})
	cleanupManaged, err := pluginhost.RegisterSourcePluginForTest(managedPlugin)
	if err != nil {
		t.Fatalf("register managed source plugin failed: %v", err)
	}
	t.Cleanup(cleanupManaged)

	optOutPluginID := nextTestSourcePluginID()
	optOutPlugin := pluginhost.NewDeclarations(optOutPluginID)
	optOutPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(optOutPluginID, false))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Should Not Load"}}}`,
			optOutPluginID,
		))},
	})
	cleanupOptOut, err := pluginhost.RegisterSourcePluginForTest(optOutPlugin)
	if err != nil {
		t.Fatalf("register opt-out source plugin failed: %v", err)
	}
	t.Cleanup(cleanupOptOut)

	missingPolicyPluginID := nextTestSourcePluginID()
	missingPolicyPlugin := pluginhost.NewDeclarations(missingPolicyPluginID)
	missingPolicyPlugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(fmt.Sprintf(`id: %s
name: Runtime I18N Missing Policy Test Plugin
version: v0.1.0
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
`, missingPolicyPluginID))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(
			`{"test":{"policy":{"%s":"Should Also Not Load"}}}`,
			missingPolicyPluginID,
		))},
	})
	cleanupMissingPolicy, err := pluginhost.RegisterSourcePluginForTest(missingPolicyPlugin)
	if err != nil {
		t.Fatalf("register missing-policy source plugin failed: %v", err)
	}
	t.Cleanup(cleanupMissingPolicy)
	resetRuntimeBundleCache()

	messages := New(bizctx.New(), config.New(), cachecoord.Default(nil)).BuildRuntimeMessages(context.Background(), EnglishLocale)
	managedKey := "test.policy." + managedPluginID
	if actual, ok := lookupMessageString(messages, managedKey); !ok || actual != "Runtime Managed" {
		t.Fatalf("expected managed plugin translation %q, got %q (exists=%v)", "Runtime Managed", actual, ok)
	}
	if value, ok := lookupMessageString(messages, "test.policy."+optOutPluginID); ok {
		t.Fatalf("expected i18n.enabled=false plugin translation to be skipped, got %q", value)
	}
	if value, ok := lookupMessageString(messages, "test.policy."+missingPolicyPluginID); ok {
		t.Fatalf("expected plugin without i18n policy translation to be skipped, got %q", value)
	}
}

// BenchmarkTranslateHotPath measures the steady-state cost of one translation
// lookup against a warm cache. The merged catalog is built once during the
// first call so subsequent iterations only exercise the lookup path.
func BenchmarkTranslateHotPath(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	// Warm up the merged catalog so the loop measures cache-hit cost only.
	if value := svc.Translate(ctx, "menu.dashboard.title", ""); value == "" {
		b.Fatalf("expected warm-up Translate to succeed, got empty value")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		benchmarkTranslateSink = svc.Translate(ctx, "menu.dashboard.title", "fallback")
	}
}

// BenchmarkTranslateBatch measures the cumulative cost of translating 100
// keys, mirroring the workload of one moderately sized list endpoint.
func BenchmarkTranslateBatch(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	keys := []string{
		"menu.dashboard.title",
		"menu.system.title",
		"dict.cron_job_status.name",
		"dict.sys_menu_type.B.label",
		"plugin.plugin-i18n-test.name",
		"locale.zh-CN.name",
		"locale.en-US.name",
		"dict.cron_job_status.0.label",
		"dict.cron_job_status.1.label",
		"dict.cron_job_status.2.label",
	}
	// Warm up.
	for _, key := range keys {
		benchmarkTranslateSink = svc.Translate(ctx, key, "")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		// Translate the same 10 keys 10 times to reach the typical 100-call
		// per-request shape of menu/dict/config heavy endpoints.
		for round := 0; round < 10; round++ {
			for _, key := range keys {
				benchmarkTranslateSink = svc.Translate(ctx, key, "fallback")
			}
		}
	}
}

// BenchmarkTranslateCacheMissRebuild measures the cost of rebuilding one
// locale's merged catalog after the cache entry was invalidated. This keeps the
// expensive cold path visible without mixing it into the hot-path benchmark.
func BenchmarkTranslateCacheMissRebuild(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	// Warm once so package-level plugin registrations and configuration readers
	// perform their one-time setup before the cold-cache loop starts.
	benchmarkTranslateSink = svc.Translate(ctx, "menu.dashboard.title", "fallback")

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		runtimeBundleCache.invalidate(InvalidateScope{Locales: []string{EnglishLocale}})
		benchmarkTranslateSink = svc.Translate(ctx, "menu.dashboard.title", "fallback")
	}
}

// BenchmarkBuildRuntimeMessages exercises the still-clones path that ships the
// full message tree to the frontend, so we can quantify the cost we accept on
// /i18n/runtime/messages while keeping it off the per-key Translate path.
func BenchmarkBuildRuntimeMessages(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), config.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	// Warm up so the loop measures the merge + clone + nest cost only.
	benchmarkRuntimeMessagesSink = svc.BuildRuntimeMessages(ctx, EnglishLocale)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		benchmarkRuntimeMessagesSink = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	}
}
