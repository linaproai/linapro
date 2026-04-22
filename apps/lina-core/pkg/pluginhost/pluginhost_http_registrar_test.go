// This file contains unit tests for the published HTTP registrar and guarded
// global middleware registration helpers exposed to source plugins.

package pluginhost

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// testHTTPRegistrarEchoReq defines one strict-route DTO used to verify handler-response capture.
type testHTTPRegistrarEchoReq struct {
	g.Meta `path:"/echo" method:"get"`
}

// testHTTPRegistrarEchoRes is the response DTO used to verify handler-response capture.
type testHTTPRegistrarEchoRes struct {
	Message string `json:"message"`
}

// testHTTPRegistrarEchoHandler returns one typed response so HandlerResponse can wrap it.
func testHTTPRegistrarEchoHandler(
	ctx context.Context,
	req *testHTTPRegistrarEchoReq,
) (*testHTTPRegistrarEchoRes, error) {
	return &testHTTPRegistrarEchoRes{Message: "ok"}, nil
}

// TestNewHTTPRegistrarExposeRoutesAndGlobalMiddlewares verifies the published
// HTTP registrar exposes both route and global middleware registration helpers.
func TestNewHTTPRegistrarExposeRoutesAndGlobalMiddlewares(t *testing.T) {
	server := g.Server("pluginhost-http-registrar-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	middlewares := NewRouteMiddlewares(
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
		func(r *ghttp.Request) {},
	)
	registrar := NewHTTPRegistrar(server, rootGroup, "plugin-demo", nil, middlewares)
	if registrar == nil {
		t.Fatal("expected HTTP registrar to be initialized")
	}
	if registrar.Routes() == nil {
		t.Fatal("expected route registrar to be published")
	}
	if registrar.GlobalMiddlewares() == nil {
		t.Fatal("expected global middleware registrar to be published")
	}
}

// TestGlobalMiddlewareRegistrarBypassesDisabledPlugin verifies disabled plugins
// do not execute their registered global middleware logic.
func TestGlobalMiddlewareRegistrarBypassesDisabledPlugin(t *testing.T) {
	server := g.Server("pluginhost-http-middleware-disabled-test")
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.BindHandler("/api/v1/ping", func(request *ghttp.Request) {
		request.Response.Write("ok")
	})

	called := false
	registrar := NewGlobalMiddlewareRegistrar(server, "plugin-demo", func(pluginID string) bool {
		return false
	})
	registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		called = true
		request.Middleware.Next()
	})

	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	client := g.Client()
	client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()))
	body := client.GetContent(context.Background(), "/api/v1/ping")

	if called {
		t.Fatal("expected disabled plugin middleware to be bypassed")
	}
	if body != "ok" {
		t.Fatalf("expected downstream handler response to stay intact, got %q", body)
	}
}

// TestGlobalMiddlewareRegistrarCapturesHandlerResponse verifies outer global
// middleware can observe the unified response after HandlerResponse completes.
func TestGlobalMiddlewareRegistrarCapturesHandlerResponse(t *testing.T) {
	server := g.Server("pluginhost-http-middleware-response-test")
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(testHTTPRegistrarEchoHandler)
	})

	captured := ""
	registrar := NewGlobalMiddlewareRegistrar(server, "plugin-demo", func(pluginID string) bool {
		return true
	})
	registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		request.Middleware.Next()
		captured = request.Response.BufferString()
	})

	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	client := g.Client()
	client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()))
	_ = client.GetContent(context.Background(), "/api/v1/echo")

	if captured == "" {
		t.Fatal("expected global middleware to capture one response body")
	}
	if !strings.Contains(captured, `"code":0`) {
		t.Fatalf("expected unified handler response wrapper, got %q", captured)
	}
	if !strings.Contains(captured, `"message":"ok"`) {
		t.Fatalf("expected typed handler payload to be visible, got %q", captured)
	}
}

// TestGlobalMiddlewareRegistrarObservesDownstreamExitAll verifies outer global
// middleware still resumes after downstream handlers terminate the request early.
func TestGlobalMiddlewareRegistrarObservesDownstreamExitAll(t *testing.T) {
	server := g.Server("pluginhost-http-middleware-exitall-test")
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(func(request *ghttp.Request) {
			request.Response.Write("stopped")
			request.ExitAll()
		})
		group.ALL("/stop", func(request *ghttp.Request) {
			request.Response.Write("unexpected")
		})
	})

	captured := ""
	registrar := NewGlobalMiddlewareRegistrar(server, "plugin-demo", func(pluginID string) bool {
		return true
	})
	registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		request.Middleware.Next()
		captured = request.Response.BufferString()
	})

	server.Start()
	defer server.Shutdown()
	time.Sleep(100 * time.Millisecond)

	client := g.Client()
	client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()))
	_ = client.GetContent(context.Background(), "/api/v1/stop")

	if captured != "stopped" {
		t.Fatalf("expected global middleware to capture downstream early-exit response, got %q", captured)
	}
}
