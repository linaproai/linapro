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
	"github.com/gogf/gf/v2/util/guid"

	configsvc "lina-core/internal/service/config"
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
	g.Meta `path:"/host/items" method:"get" tags:"宿主接口" summary:"查询宿主列表"`
}

// testHostListRes is the response DTO for the host route test handler.
type testHostListRes struct{}

// testSourceEnabledReq defines one enabled source-plugin DTO route used in tests.
type testSourceEnabledReq struct {
	g.Meta `path:"/plugins/enabled/ping" method:"get" tags:"源码插件启用" summary:"源码插件启用路由"`
}

// testSourceEnabledRes is the response DTO for the enabled source-plugin handler.
type testSourceEnabledRes struct{}

// testSourceDisabledReq defines one disabled source-plugin DTO route used in tests.
type testSourceDisabledReq struct {
	g.Meta `path:"/plugins/disabled/ping" method:"get" tags:"源码插件禁用" summary:"源码插件禁用路由"`
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
	paths["/api/v1/extensions/plugin-dynamic/review-summary"] = goai.Path{
		Get: &goai.Operation{
			Summary: "动态插件接口",
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
	if _, ok := document.Paths["/api/v1/extensions/plugin-dynamic/review-summary"]; !ok {
		t.Fatalf("expected dynamic-plugin route projection to stay available")
	}
}
