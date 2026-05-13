// This file verifies route text resolution used by audit logs reuses apidoc
// i18n catalogs instead of runtime UI language bundles.

package apidoc

import (
	"context"
	"reflect"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
)

// TestResolveRouteTextUsesApidocCatalog verifies route tags and summaries are
// resolved from the same structured apidoc keys used by OpenAPI localization.
func TestResolveRouteTextUsesApidocCatalog(t *testing.T) {
	operationKey := BuildRouteOperationKeyFromRequestType(reflect.TypeOf(testHostListReq{}))
	if operationKey != "core.internal.service.apidoc.testHostListReq" {
		t.Fatalf("expected static operation key to match apidoc component key, got %s", operationKey)
	}

	restoreCatalog := registerOpenAPITestCatalog("zh-CN", map[string]string{
		operationKey + ".meta.tags":    "用户管理",
		operationKey + ".meta.summary": "获取用户列表",
	})
	defer restoreCatalog()

	service := New(&testConfigProvider{}, bizctx.New(), i18nsvc.New(bizctx.New(), configsvc.New(), cachecoord.Default(nil)), &testPluginRouteProvider{}).(*serviceImpl)
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: "zh-CN"})
	output := service.ResolveRouteText(ctx, RouteTextInput{
		OperationKey:    operationKey,
		FallbackTitle:   "User Management",
		FallbackSummary: "Get user list",
	})
	if output.Title != "用户管理" {
		t.Fatalf("expected Chinese title 用户管理, got %s", output.Title)
	}
	if output.Summary != "获取用户列表" {
		t.Fatalf("expected Chinese summary 获取用户列表, got %s", output.Summary)
	}
}

// TestFindRouteTitleOperationKeys verifies localized title search returns
// operation key bases that can be matched against persisted audit records.
func TestFindRouteTitleOperationKeys(t *testing.T) {
	restoreCatalog := registerOpenAPITestCatalog("zh-CN", map[string]string{
		"core.api.user.v1.ListReq.meta.tags":                              "用户管理",
		"plugins.plugin_demo_dynamic.paths.get.backend_summary.meta.tags": "动态插件示例",
	})
	defer restoreCatalog()

	service := New(&testConfigProvider{}, bizctx.New(), i18nsvc.New(bizctx.New(), configsvc.New(), cachecoord.Default(nil)), &testPluginRouteProvider{}).(*serviceImpl)
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: "zh-CN"})
	keys := service.FindRouteTitleOperationKeys(ctx, "动态")
	expectedKey := "plugins.plugin_demo_dynamic.paths.get.backend_summary"
	for _, key := range keys {
		if key == expectedKey {
			return
		}
	}
	t.Fatalf("expected localized title matches to contain %s, got %#v", expectedKey, keys)
}

// TestBuildRouteOperationKeyFromPathNormalizesMethod verifies persisted route
// methods can be uppercase while apidoc path keys remain lowercase.
func TestBuildRouteOperationKeyFromPathNormalizesMethod(t *testing.T) {
	key := BuildRouteOperationKeyFromPath("/api/v1/extensions/plugin-demo-dynamic/backend-summary", "GET")
	if key != "plugins.plugin_demo_dynamic.paths.get.backend_summary" {
		t.Fatalf("expected lower-case dynamic path key, got %s", key)
	}
}
