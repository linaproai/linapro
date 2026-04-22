// This file contains unit tests for the route registrar helpers exposed by the
// pluginhost package to source plugins and host bootstrap code.

package pluginhost

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// testPluginPingReq defines one strict-route DTO used by route binding tests.
type testPluginPingReq struct {
	g.Meta `path:"/plugins/test-plugin/ping" method:"get"`
}

// testPluginPingRes is the response DTO for the strict-route ping handler.
type testPluginPingRes struct{}

// testPluginPingHandler is the strict-route handler used to verify route capture.
func testPluginPingHandler(ctx context.Context, req *testPluginPingReq) (*testPluginPingRes, error) {
	return &testPluginPingRes{}, nil
}

// TestNewRouteRegistrarExposeRootGroupAndPublishedMiddlewares verifies the
// registrar exposes the root group and the published middleware directory.
func TestNewRouteRegistrarExposeRootGroupAndPublishedMiddlewares(t *testing.T) {
	noop := func(r *ghttp.Request) {}
	middlewares := NewRouteMiddlewares(noop, noop, noop, noop, noop, noop, noop)
	server := g.Server("pluginhost-route-registrar-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, middlewares)
	typed, ok := registrar.(*routeRegistrar)
	if !ok {
		t.Fatalf("expected concrete route registrar type")
	}
	if typed.group == nil {
		t.Fatalf("expected root plugin route group to be initialized")
	}
	if registrar.Middlewares() == nil {
		t.Fatalf("expected published middleware directory to be available")
	}

	called := false
	registrar.Group("/api/v1", func(group RouteGroup) {
		called = true
		if group == nil {
			t.Fatalf("expected callback group to be initialized")
		}
		group.Middleware(
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
			middlewares.Auth(),
			middlewares.Permission(),
		)
	})
	if !called {
		t.Fatalf("expected group callback to be invoked during route registration")
	}
}

// TestRouteRegistrarCaptureSourceRouteBindings verifies the host captures
// plugin-owned route bindings while preserving nested group path composition.
func TestRouteRegistrarCaptureSourceRouteBindings(t *testing.T) {
	server := g.Server("pluginhost-route-registrar-binding-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, nil)
	registrar.Group("/api/v1", func(group RouteGroup) {
		group.Group("/plugins", func(group RouteGroup) {
			group.Bind(testPluginPingHandler)
			group.GET("/raw", func(r *ghttp.Request) {})
		})
	})

	bindings := registrar.RouteBindings()
	if len(bindings) != 2 {
		t.Fatalf("expected 2 route bindings, got %d", len(bindings))
	}
	if bindings[0].PluginID != "plugin-demo" {
		t.Fatalf("expected plugin id plugin-demo, got %s", bindings[0].PluginID)
	}
	if bindings[0].Method != "GET" {
		t.Fatalf("expected GET binding, got %s", bindings[0].Method)
	}
	if bindings[0].Path != "/api/v1/plugins/plugins/test-plugin/ping" {
		t.Fatalf("expected strict route path to include nested prefix, got %s", bindings[0].Path)
	}
	if !bindings[0].Documentable {
		t.Fatalf("expected strict DTO handler to be documentable")
	}
	if bindings[1].Path != "/api/v1/plugins/raw" {
		t.Fatalf("expected raw handler path /api/v1/plugins/raw, got %s", bindings[1].Path)
	}
	if bindings[1].Documentable {
		t.Fatalf("expected raw handler to be non-documentable")
	}
}

// TestNormalizeRoutePrefix verifies plugin route prefixes are normalized before
// the host binds them to a router group.
func TestNormalizeRoutePrefix(t *testing.T) {
	if got := normalizeRoutePrefix("api/v1/"); got != "/api/v1" {
		t.Fatalf("expected normalized prefix /api/v1, got %s", got)
	}
	if got := normalizeRoutePrefix(""); got != "/" {
		t.Fatalf("expected empty prefix to normalize to root, got %s", got)
	}
}
