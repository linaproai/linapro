// This file contains unit tests for the route registrar helpers exposed by the
// pluginhost package to source plugins and host bootstrap code.

package pluginhost

import (
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

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
	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		called = true
		if group == nil {
			t.Fatalf("expected callback group to be initialized")
		}
		group.Middleware(middlewares.Ctx(), middlewares.Auth(), middlewares.Permission())
	})
	if !called {
		t.Fatalf("expected group callback to be invoked during route registration")
	}
}

func TestNormalizeRoutePrefix(t *testing.T) {
	if got := normalizeRoutePrefix("api/v1/"); got != "/api/v1" {
		t.Fatalf("expected normalized prefix /api/v1, got %s", got)
	}
	if got := normalizeRoutePrefix(""); got != "/" {
		t.Fatalf("expected empty prefix to normalize to root, got %s", got)
	}
}
