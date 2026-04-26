// This file covers host-managed OpenAPI document construction for host routes,
// source-plugin routes, and dynamic-plugin route projection.

package apidoc

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/pluginhost"
)

// testConfigProvider provides fixed OpenAPI metadata for builder tests.
type testConfigProvider struct{}

// testPluginRouteProvider provides controllable source and dynamic plugin route
// projection inputs for builder tests.
type testPluginRouteProvider struct {
	enabledByID  map[string]bool
	sourceRoutes []pluginhost.SourceRouteBinding
}

// testHostListReq defines one host-owned DTO route used in apidoc builder tests.
type testHostListReq struct {
	g.Meta  `path:"/host/items" method:"get" tags:"User Management" summary:"Get user list" dc:"Query the paginated user list, support filtering by user name, nickname, status, mobile phone number, gender, department, creation time and other conditions, support custom sorting"`
	PageNum int `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
}

// testHostListRes is the response DTO for the host route test handler.
type testHostListRes struct {
	Message string `json:"message" dc:"Message title" eg:"System maintenance notification"`
}

// testSourceEnabledReq defines one enabled source-plugin DTO route used in tests.
type testSourceEnabledReq struct {
	g.Meta `path:"/plugins/enabled/ping" method:"get" tags:"Source Plugin Demo" summary:"Query source plugin example public ping" dc:"Return source plugin example public ping information."`
}

// testSourceEnabledRes is the response DTO for the enabled source-plugin handler.
type testSourceEnabledRes struct{}

// testSourceDisabledReq defines one disabled source-plugin DTO route used in tests.
type testSourceDisabledReq struct {
	g.Meta `path:"/plugins/disabled/ping" method:"get" tags:"Source Plugin Demo" summary:"Query source plugin example public ping" dc:"Return source plugin example public ping information."`
}

// testSourceDisabledRes is the response DTO for the disabled source-plugin handler.
type testSourceDisabledRes struct{}

// GetOpenApi returns fixed host document metadata for builder tests.
func (p *testConfigProvider) GetOpenApi(ctx context.Context) *configsvc.OpenApiConfig {
	return &configsvc.OpenApiConfig{
		Title:             "Hosted API",
		Description:       "Host managed OpenAPI document",
		Version:           "v-test",
		ServerUrl:         "https://api.example.com",
		ServerDescription: "Test API Server",
	}
}

// ListSourceRouteBindings returns the test-controlled source route snapshot.
func (p *testPluginRouteProvider) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	return pluginhost.CloneSourceRouteBindings(p.sourceRoutes)
}

// IsEnabled returns the configured enablement state for the requested plugin.
func (p *testPluginRouteProvider) IsEnabled(ctx context.Context, pluginID string) bool {
	return p.enabledByID[pluginID]
}

// ProjectDynamicRoutesToOpenAPI injects one synthetic dynamic-plugin route into
// the document under test.
func (p *testPluginRouteProvider) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	paths["/api/v1/extensions/plugin-demo-dynamic/backend-summary"] = goai.Path{
		Get: &goai.Operation{
			Tags:        []string{"Dynamic Plugin Demo"},
			Summary:     "Query the dynamic plugin backend execution summary",
			Description: "Return the current bridge execution summary for plugin-demo-dynamic when dispatched through the host prefix /api/v1/extensions/{pluginId}/..., including plugin ID, route information, and current user context.",
		},
	}
	return nil
}

// testHostListHandler is the strict-route host handler used by the builder test.
func testHostListHandler(ctx context.Context, req *testHostListReq) (*testHostListRes, error) {
	return &testHostListRes{}, nil
}

// testSourceEnabledHandler is the strict-route source-plugin handler for the enabled case.
func testSourceEnabledHandler(ctx context.Context, req *testSourceEnabledReq) (*testSourceEnabledRes, error) {
	return &testSourceEnabledRes{}, nil
}

// testSourceDisabledHandler is the strict-route source-plugin handler for the disabled case.
func testSourceDisabledHandler(ctx context.Context, req *testSourceDisabledReq) (*testSourceDisabledRes, error) {
	return &testSourceDisabledRes{}, nil
}

// TestBuildProjectsHostAndEnabledPluginRoutes verifies the host-managed OpenAPI
// document keeps host routes, filters disabled source routes, and includes
// dynamic-plugin projections.
func TestBuildProjectsHostAndEnabledPluginRoutes(t *testing.T) {
	server := g.Server("apidoc-builder-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Bind(testHostListHandler)
		group.Bind(testSourceEnabledHandler)
		group.Bind(testSourceDisabledHandler)
	})
	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	pluginProvider := &testPluginRouteProvider{
		enabledByID: map[string]bool{
			"plugin-source-enabled":  true,
			"plugin-source-disabled": false,
		},
		sourceRoutes: []pluginhost.SourceRouteBinding{
			{
				PluginID:     "plugin-source-enabled",
				Method:       "GET",
				Path:         "/api/v1/plugins/enabled/ping",
				Handler:      testSourceEnabledHandler,
				Documentable: true,
			},
			{
				PluginID:     "plugin-source-disabled",
				Method:       "GET",
				Path:         "/api/v1/plugins/disabled/ping",
				Handler:      testSourceDisabledHandler,
				Documentable: true,
			},
		},
	}

	service := New(&testConfigProvider{}, pluginProvider)
	document, err := service.Build(context.Background(), server)
	if err != nil {
		t.Fatalf("expected hosted apidoc build to succeed, got %v", err)
	}
	if document.Info.Title != "Hosted API" {
		t.Fatalf("expected hosted title Hosted API, got %s", document.Info.Title)
	}
	if document.Info.Version != "v-test" {
		t.Fatalf("expected hosted version v-test, got %s", document.Info.Version)
	}
	if document.Security == nil {
		t.Fatalf("expected hosted document to publish bearer security")
	}
	if _, ok := document.Paths["/api/v1/host/items"]; !ok {
		t.Fatalf("expected host static route to stay in hosted document")
	}
	if _, ok := document.Paths["/api/v1/plugins/enabled/ping"]; !ok {
		t.Fatalf("expected enabled source-plugin route to be projected")
	}
	if _, ok := document.Paths["/api/v1/plugins/disabled/ping"]; ok {
		t.Fatalf("expected disabled source-plugin route to be removed from hosted document")
	}
	if _, ok := document.Paths["/api/v1/extensions/plugin-demo-dynamic/backend-summary"]; !ok {
		t.Fatalf("expected dynamic-plugin route projection to stay available")
	}
}

// TestBuildLocalizesOpenAPIForRequestLocale verifies the hosted OpenAPI output
// localizes operation groups, descriptions, parameter descriptions, response
// schemas, examples, and dynamic-plugin route metadata for request locales.
func TestBuildLocalizesOpenAPIForRequestLocale(t *testing.T) {
	server := g.Server("apidoc-i18n-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Bind(testHostListHandler)
		group.Bind(testSourceEnabledHandler)
	})
	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	pluginProvider := &testPluginRouteProvider{
		enabledByID: map[string]bool{
			"plugin-source-enabled": true,
		},
		sourceRoutes: []pluginhost.SourceRouteBinding{
			{
				PluginID:     "plugin-source-enabled",
				Method:       "GET",
				Path:         "/api/v1/plugins/enabled/ping",
				Handler:      testSourceEnabledHandler,
				Documentable: true,
			},
		},
	}

	ctx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.EnglishLocale},
	)
	service := New(&testConfigProvider{}, pluginProvider)
	document, err := service.Build(ctx, server)
	if err != nil {
		t.Fatalf("expected hosted apidoc build to succeed, got %v", err)
	}

	hostOperation := document.Paths["/api/v1/host/items"].Get
	if hostOperation == nil {
		t.Fatalf("expected host operation to be present")
	}
	if got := hostOperation.Tags[0]; got != "User Management" {
		t.Fatalf("expected localized host tag User Management, got %s", got)
	}
	if got := hostOperation.Summary; got != "Get user list" {
		t.Fatalf("expected localized host summary Get user list, got %s", got)
	}
	if len(hostOperation.Parameters) == 0 || hostOperation.Parameters[0].Value == nil {
		t.Fatalf("expected host operation to expose query parameters")
	}
	if got := hostOperation.Parameters[0].Value.Description; got != "Page number" {
		t.Fatalf("expected localized page parameter description Page number, got %s", got)
	}

	sourceOperation := document.Paths["/api/v1/plugins/enabled/ping"].Get
	if sourceOperation == nil || sourceOperation.Tags[0] != "Source Plugin Demo" {
		t.Fatalf("expected source-plugin operation tag to be localized")
	}
	dynamicOperation := document.Paths["/api/v1/extensions/plugin-demo-dynamic/backend-summary"].Get
	if dynamicOperation == nil || dynamicOperation.Summary != "Query the dynamic plugin backend execution summary" {
		t.Fatalf("expected dynamic-plugin operation summary to be localized")
	}

	restoreCatalog := registerOpenAPITestCatalog("zh-CN", map[string]string{
		"core.internal.service.apidoc.testHostListReq.fields.pageNum.dc":     "页码",
		"core.internal.service.apidoc.testHostListReq.meta.summary":          "获取用户列表",
		"core.internal.service.apidoc.testHostListReq.meta.tags":             "用户管理",
		"core.internal.service.apidoc.testSourceEnabledReq.meta.tags":        "源码插件示例",
		"plugins.plugin_demo_dynamic.paths.get.backend_summary.meta.summary": "查询动态插件后端执行摘要",
	})
	defer restoreCatalog()

	zhCtx := context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: "zh-CN"},
	)
	zhDocument, err := service.Build(zhCtx, server)
	if err != nil {
		t.Fatalf("expected hosted Chinese apidoc build to succeed, got %v", err)
	}
	zhHostOperation := zhDocument.Paths["/api/v1/host/items"].Get
	if zhHostOperation == nil {
		t.Fatalf("expected localized Chinese host operation to be present")
	}
	if got := zhHostOperation.Tags[0]; got != "用户管理" {
		t.Fatalf("expected Chinese host tag 用户管理, got %s", got)
	}
	if got := zhHostOperation.Summary; got != "获取用户列表" {
		t.Fatalf("expected Chinese host summary 获取用户列表, got %s", got)
	}
	if len(zhHostOperation.Parameters) == 0 || zhHostOperation.Parameters[0].Value == nil {
		t.Fatalf("expected Chinese host operation to expose query parameters")
	}
	if got := zhHostOperation.Parameters[0].Value.Description; got != "页码" {
		t.Fatalf("expected Chinese page parameter description 页码, got %s", got)
	}
	zhSourceOperation := zhDocument.Paths["/api/v1/plugins/enabled/ping"].Get
	if zhSourceOperation == nil || zhSourceOperation.Tags[0] != "源码插件示例" {
		t.Fatalf("expected source-plugin operation tag to be localized to Chinese")
	}
	zhDynamicOperation := zhDocument.Paths["/api/v1/extensions/plugin-demo-dynamic/backend-summary"].Get
	if zhDynamicOperation == nil || zhDynamicOperation.Summary != "查询动态插件后端执行摘要" {
		t.Fatalf("expected dynamic-plugin operation summary to be localized to Chinese")
	}
}

// registerOpenAPITestCatalog injects test-only route translations into the
// apidoc catalog cache without shipping fixture keys in production resources.
func registerOpenAPITestCatalog(locale string, entries map[string]string) func() {
	catalog := loadOpenAPIMessageCatalog(context.Background(), locale)
	mergeOpenAPIMessageCatalog(catalog, entries)

	normalizedLocale := normalizeOpenAPILocale(locale)
	openAPIMessageCache.Lock()
	openAPIMessageCache.bundles[normalizedLocale] = cloneOpenAPIMessageCatalog(catalog)
	openAPIMessageCache.Unlock()

	return invalidateOpenAPIMessageCache
}

// TestLocalizeSchemaTranslatesAlreadySeenDirectMetadata verifies recursive
// cycle guards do not skip direct schema display text for shared schema nodes.
func TestLocalizeSchemaTranslatesAlreadySeenDirectMetadata(t *testing.T) {
	service := New(&testConfigProvider{}, &testPluginRouteProvider{}).(*serviceImpl)
	localizer := &openAPILocalizer{
		catalog: map[string]string{
			"test.schema.title":   "Parameter ID",
			"test.schema.dc":      "Parameter ID",
			"test.schema.default": "Parameter ID",
		},
	}
	schema := &goai.Schema{
		Title:       "参数ID",
		Description: "参数ID",
		Default:     "参数ID",
		Example:     "参数ID",
	}
	seenSchemas := map[*goai.Schema]struct{}{
		schema: {},
	}

	service.localizeSchema(localizer, schema, "test.schema", seenSchemas)

	if schema.Title != "Parameter ID" {
		t.Fatalf("expected already-seen schema title to be localized, got %s", schema.Title)
	}
	if schema.Description != "Parameter ID" {
		t.Fatalf("expected already-seen schema description to be localized, got %s", schema.Description)
	}
	if schema.Default != "Parameter ID" {
		t.Fatalf("expected already-seen schema default to be localized, got %v", schema.Default)
	}
	if schema.Example != "参数ID" {
		t.Fatalf("expected already-seen schema example to stay unchanged, got %v", schema.Example)
	}
}

// TestEnglishLocalizerPreservesGeneratedSchemaMetadata verifies en-US output
// keeps generated entity and framework metadata exactly as its source provides
// it when the empty en-US apidoc bundle has no translation entry.
func TestEnglishLocalizerPreservesGeneratedSchemaMetadata(t *testing.T) {
	service := New(&testConfigProvider{}, &testPluginRouteProvider{}).(*serviceImpl)
	localizer := &openAPILocalizer{
		locale:  i18nsvc.EnglishLocale,
		catalog: map[string]string{},
	}
	schema := &goai.Schema{
		Title:       "参数ID",
		Description: "参数ID",
		Default:     "参数ID",
		Example:     "参数ID",
		Properties:  &goai.Schemas{},
	}
	schema.Properties.Set("createdAt", goai.SchemaRef{
		Value: &goai.Schema{Description: "创建时间"},
	})
	schema.Properties.Set("status", goai.SchemaRef{
		Value: &goai.Schema{Description: "状态（0停用 1正常）"},
	})

	service.localizeSchema(localizer, schema, "core.api.config.v1.GetReq.responses.200.content.application_json.fields.data", map[*goai.Schema]struct{}{})

	if schema.Title != "参数ID" {
		t.Fatalf("expected generated schema title to stay as source 参数ID, got %s", schema.Title)
	}
	if schema.Description != "参数ID" {
		t.Fatalf("expected generated schema description to stay as source 参数ID, got %s", schema.Description)
	}
	if schema.Default != "参数ID" {
		t.Fatalf("expected generated schema default to stay as source 参数ID, got %v", schema.Default)
	}
	if schema.Example != "参数ID" {
		t.Fatalf("expected generated schema example to stay unchanged, got %v", schema.Example)
	}

	createdAtRef := schema.Properties.Get("createdAt")
	if createdAtRef == nil {
		t.Fatalf("expected createdAt property to be present")
	}
	createdAt := createdAtRef.Value
	if createdAt == nil || createdAt.Description != "创建时间" {
		t.Fatalf("expected createdAt description to stay as source 创建时间, got %#v", createdAt)
	}
	statusRef := schema.Properties.Get("status")
	if statusRef == nil {
		t.Fatalf("expected status property to be present")
	}
	status := statusRef.Value
	if status == nil || status.Description != "状态（0停用 1正常）" {
		t.Fatalf("expected status description to stay as source 状态（0停用 1正常）, got %#v", status)
	}
	if got := localizer.translate("core.api.auth.v1.LoginReq.responses.200.content.application_json.fields.code.dc", "错误码"); got != "错误码" {
		t.Fatalf("expected common response code description to stay as source 错误码, got %s", got)
	}
}
