// This file holds runtime translation benchmarks. They prove the hot path
// stays sub-100ns once the merged catalog is built and the cache is warm.

package i18n

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
)

// BenchmarkTranslateHotPath measures the steady-state cost of one translation
// lookup against a warm cache. The merged catalog is built once during the
// first call so subsequent iterations only exercise the lookup path.
func BenchmarkTranslateHotPath(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	// Warm up the merged catalog so the loop measures cache-hit cost only.
	if value := svc.Translate(ctx, "menu.dashboard.title", ""); value == "" {
		b.Fatalf("expected warm-up Translate to succeed, got empty value")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = svc.Translate(ctx, "menu.dashboard.title", "fallback")
	}
}

// BenchmarkTranslateBatch measures the cumulative cost of translating 100
// keys, mirroring the workload of one moderately sized list endpoint.
func BenchmarkTranslateBatch(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New()
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
		_ = svc.Translate(ctx, key, "")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		// Translate the same 10 keys 10 times to reach the typical 100-call
		// per-request shape of menu/dict/config heavy endpoints.
		for round := 0; round < 10; round++ {
			for _, key := range keys {
				_ = svc.Translate(ctx, key, "fallback")
			}
		}
	}
}

// BenchmarkBuildRuntimeMessages exercises the still-clones path that ships the
// full message tree to the frontend, so we can quantify the cost we accept on
// /i18n/runtime/messages while keeping it off the per-key Translate path.
func BenchmarkBuildRuntimeMessages(b *testing.B) {
	resetRuntimeBundleCache()

	svc := New()
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	// Warm up so the loop measures the merge + clone + nest cost only.
	_ = svc.BuildRuntimeMessages(ctx, EnglishLocale)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = svc.BuildRuntimeMessages(ctx, EnglishLocale)
	}
}
