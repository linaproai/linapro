// This file tests the integration-owned source-plugin HTTP registrar, including
// route capture, enablement guards, and global middleware behavior.

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/datascope"
	"lina-core/pkg/plugin/pluginhost"
)

// testSourceHTTPPingReq defines one strict-route DTO used by route binding tests.
type testSourceHTTPPingReq struct {
	g.Meta `path:"/plugins/test-plugin/ping" method:"get"`
}

// testSourceHTTPPingRes is the response DTO for the strict-route ping handler.
type testSourceHTTPPingRes struct{}

// testSourceHTTPIllegalXReq defines one invalid DTO route outside the owned /x namespace.
type testSourceHTTPIllegalXReq struct {
	g.Meta `path:"/x/other-plugin/health" method:"get"`
}

// testSourceHTTPIllegalXRes is the response DTO for an invalid /x route.
type testSourceHTTPIllegalXRes struct{}

// testSourceHTTPEchoReq defines one strict-route DTO used to verify handler-response capture.
type testSourceHTTPEchoReq struct {
	g.Meta `path:"/echo" method:"get"`
}

// testSourceHTTPEchoRes is the response DTO used to verify handler-response capture.
type testSourceHTTPEchoRes struct {
	Message string `json:"message"`
}

// testSourceHTTPPingHandler is the strict-route handler used to verify route capture.
func testSourceHTTPPingHandler(ctx context.Context, req *testSourceHTTPPingReq) (*testSourceHTTPPingRes, error) {
	return &testSourceHTTPPingRes{}, nil
}

// testSourceHTTPIllegalXHandler verifies /x namespace validation for DTO routes.
func testSourceHTTPIllegalXHandler(
	ctx context.Context,
	req *testSourceHTTPIllegalXReq,
) (*testSourceHTTPIllegalXRes, error) {
	return &testSourceHTTPIllegalXRes{}, nil
}

// testSourceHTTPEchoHandler returns one typed response so HandlerResponse can wrap it.
func testSourceHTTPEchoHandler(
	ctx context.Context,
	req *testSourceHTTPEchoReq,
) (*testSourceHTTPEchoRes, error) {
	return &testSourceHTTPEchoRes{Message: "ok"}, nil
}

// TestSourceHTTPRegistrarExposeRoutesAndGlobalMiddlewares verifies the
// integration registrar publishes route and global middleware helpers.
func TestSourceHTTPRegistrarExposeRoutesAndGlobalMiddlewares(t *testing.T) {
	server, rootGroup := newSourceHTTPTestServer(t, false)
	middlewares := newSourceHTTPNoopMiddlewares()
	registrar := newSourceHTTPRegistrar(
		server,
		rootGroup,
		"plugin-demo",
		middlewares,
		newSourceHTTPTestService(true),
	)

	if registrar == nil {
		t.Fatal("expected HTTP registrar to be initialized")
	}
	if registrar.Routes() == nil {
		t.Fatal("expected route registrar to be published")
	}
	if registrar.GlobalMiddlewares() == nil {
		t.Fatal("expected global middleware registrar to be published")
	}
	if registrar.Routes().Middlewares() == nil {
		t.Fatal("expected route middleware directory to be published")
	}
}

// TestSourceRouteRegistrarExposeRootGroupAndPublishedMiddlewares verifies the
// registrar exposes the root group and the published middleware directory.
func TestSourceRouteRegistrarExposeRootGroupAndPublishedMiddlewares(t *testing.T) {
	_, rootGroup := newSourceHTTPTestServer(t, false)
	middlewares := newSourceHTTPNoopMiddlewares()
	registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", middlewares, newSourceHTTPTestService(true))
	if registrar.group == nil {
		t.Fatalf("expected root plugin route group to be initialized")
	}
	if registrar.Middlewares() == nil {
		t.Fatalf("expected published middleware directory to be available")
	}
	if registrar.APIPrefix() != "/x/plugin-demo" {
		t.Fatalf("expected plugin API prefix /x/plugin-demo, got %q", registrar.APIPrefix())
	}

	called := false
	registrar.Group(registrar.APIPrefix()+"/api/v1", func(group pluginhost.RouteGroup) {
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

// TestSourceRouteRegistrarCaptureSourceRouteBindings verifies the host captures
// plugin-owned route bindings while preserving nested group path composition.
func TestSourceRouteRegistrarCaptureSourceRouteBindings(t *testing.T) {
	_, rootGroup := newSourceHTTPTestServer(t, false)
	registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
	registrar.Group(registrar.APIPrefix()+"/api/v1", func(group pluginhost.RouteGroup) {
		group.Group("/plugins", func(group pluginhost.RouteGroup) {
			group.Bind(testSourceHTTPPingHandler)
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
	if bindings[0].Path != "/x/plugin-demo/api/v1/plugins/plugins/test-plugin/ping" {
		t.Fatalf("expected strict route path to include nested prefix, got %s", bindings[0].Path)
	}
	if !bindings[0].Documentable {
		t.Fatalf("expected strict DTO handler to be documentable")
	}
	if bindings[1].Path != "/x/plugin-demo/api/v1/plugins/raw" {
		t.Fatalf("expected raw handler path /x/plugin-demo/api/v1/plugins/raw, got %s", bindings[1].Path)
	}
	if bindings[1].Documentable {
		t.Fatalf("expected raw handler to be non-documentable")
	}
}

// TestSourceRouteRegistrarRejectsNonAPIRoutesUnderX verifies `/x` remains
// reserved for plugin API routes owned by the current plugin.
func TestSourceRouteRegistrarRejectsNonAPIRoutesUnderX(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "x root", prefix: "/x"},
		{name: "other plugin api", prefix: "/x/other-plugin/api/v1"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, rootGroup := newSourceHTTPTestServer(t, false)
			registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
			registrar.Group(testCase.prefix, func(group pluginhost.RouteGroup) {
				group.GET("/ping", func(request *ghttp.Request) {})
			})
			if registrar.Err() == nil {
				t.Fatalf("expected prefix %q to be rejected", testCase.prefix)
			}
		})
	}
}

// TestSourceRouteRegistrarAllowsPluginOwnedPathsUnderX verifies `/x/{pluginId}`
// is the only mandatory source-plugin API prefix.
func TestSourceRouteRegistrarAllowsPluginOwnedPathsUnderX(t *testing.T) {
	tests := []struct {
		name         string
		groupPrefix  string
		routePattern string
		expectedPath string
	}{
		{
			name:         "api v2",
			groupPrefix:  "/x/plugin-demo/api/v2",
			routePattern: "/items",
			expectedPath: "/x/plugin-demo/api/v2/items",
		},
		{
			name:         "interface path",
			groupPrefix:  "/x/plugin-demo/interface/m1",
			routePattern: "/items",
			expectedPath: "/x/plugin-demo/interface/m1/items",
		},
		{
			name:         "graphql",
			groupPrefix:  "/x/plugin-demo",
			routePattern: "/graphql",
			expectedPath: "/x/plugin-demo/graphql",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, rootGroup := newSourceHTTPTestServer(t, false)
			registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
			registrar.Group(testCase.groupPrefix, func(group pluginhost.RouteGroup) {
				group.GET(testCase.routePattern, func(request *ghttp.Request) {})
			})
			if err := registrar.Err(); err != nil {
				t.Fatalf("expected plugin-owned /x route to be accepted, got error: %v", err)
			}
			bindings := registrar.RouteBindings()
			if len(bindings) != 1 {
				t.Fatalf("expected 1 route binding, got %d", len(bindings))
			}
			if bindings[0].Path != testCase.expectedPath {
				t.Fatalf("expected route path %s, got %s", testCase.expectedPath, bindings[0].Path)
			}
		})
	}
}

// TestSourceRouteRegistrarRejectsNestedRoutesOutsideOwnedX verifies nested
// groups and DTO metadata cannot bypass the reserved /x plugin namespace guard.
func TestSourceRouteRegistrarRejectsNestedRoutesOutsideOwnedX(t *testing.T) {
	tests := []struct {
		name     string
		register func(registrar pluginhost.RouteRegistrar)
	}{
		{
			name: "nested group",
			register: func(registrar pluginhost.RouteRegistrar) {
				registrar.Group("/x/plugin-demo/api/v1", func(group pluginhost.RouteGroup) {
					group.Group("/../../x/other-plugin/assets", func(group pluginhost.RouteGroup) {
						group.GET("/logo", func(request *ghttp.Request) {})
					})
				})
			},
		},
		{
			name: "dto meta path",
			register: func(registrar pluginhost.RouteRegistrar) {
				registrar.Group("/", func(group pluginhost.RouteGroup) {
					group.Bind(testSourceHTTPIllegalXHandler)
				})
			},
		},
		{
			name: "method path",
			register: func(registrar pluginhost.RouteRegistrar) {
				registrar.Group("/", func(group pluginhost.RouteGroup) {
					group.GET("/x/other-plugin/health", func(request *ghttp.Request) {})
				})
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, rootGroup := newSourceHTTPTestServer(t, false)
			registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
			testCase.register(registrar)
			if registrar.Err() == nil {
				t.Fatalf("expected route registration to be rejected")
			}
		})
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

// TestSourceRouteRegistrarEnabledRouteAttachesPluginID verifies enabled plugin
// routes attach their source-plugin id to the active request.
func TestSourceRouteRegistrarEnabledRouteAttachesPluginID(t *testing.T) {
	server, rootGroup := newSourceHTTPTestServer(t, true)
	registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
	registrar.Group("/plugins/plugin-demo", func(group pluginhost.RouteGroup) {
		group.GET("/ping", func(request *ghttp.Request) {
			request.Response.Write(pluginhost.SourcePluginIDFromRequest(request))
		})
	})

	startSourceHTTPTestServer(t, server)
	statusCode, body := sourceHTTPTestGet(t, server, "/plugins/plugin-demo/ping")
	if statusCode != http.StatusOK {
		t.Fatalf("expected enabled route status 200, got %d body=%q", statusCode, body)
	}
	if body != "plugin-demo" {
		t.Fatalf("expected source plugin id plugin-demo, got %q", body)
	}
}

// TestSourceRouteRegistrarRejectsDisabledPlugin verifies disabled plugins do
// not execute route handlers and return 404.
func TestSourceRouteRegistrarRejectsDisabledPlugin(t *testing.T) {
	server, rootGroup := newSourceHTTPTestServer(t, true)
	registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(false))
	called := false
	registrar.Group("/plugins/plugin-demo", func(group pluginhost.RouteGroup) {
		group.GET("/ping", func(request *ghttp.Request) {
			called = true
			request.Response.Write("unexpected")
		})
	})

	startSourceHTTPTestServer(t, server)
	statusCode, _ := sourceHTTPTestGet(t, server, "/plugins/plugin-demo/ping")
	if statusCode != http.StatusNotFound {
		t.Fatalf("expected disabled route status 404, got %d", statusCode)
	}
	if called {
		t.Fatal("expected disabled plugin handler not to run")
	}
}

// TestSourceRouteRegistrarUsesRequestContextForEnablement verifies route guards
// evaluate the active request context instead of a background snapshot context.
func TestSourceRouteRegistrarUsesRequestContextForEnablement(t *testing.T) {
	server := g.Server(sourceHTTPTestServerName(t))
	server.SetDumpRouterMap(false)
	server.SetPort(0)

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(func(request *ghttp.Request) {
			request.SetCtx(datascope.WithTenantScope(request.Context(), 42))
			request.Middleware.Next()
		})
		rootGroup = group
	})

	registrar := newSourceRouteRegistrar(rootGroup, "plugin-demo", nil, newSourceHTTPTestService(true))
	called := false
	registrar.Group("/plugins/plugin-demo", func(group pluginhost.RouteGroup) {
		group.GET("/ping", func(request *ghttp.Request) {
			called = true
			request.Response.Write("unexpected")
		})
	})

	startSourceHTTPTestServer(t, server)
	statusCode, _ := sourceHTTPTestGet(t, server, "/plugins/plugin-demo/ping")
	if statusCode != http.StatusNotFound {
		t.Fatalf("expected tenant-scoped request without tenant registry access to return 404, got %d", statusCode)
	}
	if called {
		t.Fatal("expected request-context enablement guard to stop handler execution")
	}
}

// TestSourceGlobalMiddlewareRegistrarBypassesDisabledPlugin verifies disabled
// plugins do not execute their registered global middleware logic.
func TestSourceGlobalMiddlewareRegistrarBypassesDisabledPlugin(t *testing.T) {
	server, _ := newSourceHTTPTestServer(t, true)
	server.BindHandler("/api/v1/ping", func(request *ghttp.Request) {
		request.Response.Write("ok")
	})

	called := false
	registrar := &sourceGlobalMiddlewareRegistrar{
		server:   server,
		pluginID: "plugin-demo",
		service:  newSourceHTTPTestService(false),
	}
	if err := registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		called = true
		request.Middleware.Next()
	}); err != nil {
		t.Fatalf("expected middleware registration to succeed, got %v", err)
	}

	startSourceHTTPTestServer(t, server)
	statusCode, body := sourceHTTPTestGet(t, server, "/api/v1/ping")
	if statusCode != http.StatusOK {
		t.Fatalf("expected downstream status 200, got %d", statusCode)
	}
	if called {
		t.Fatal("expected disabled plugin middleware to be bypassed")
	}
	if body != "ok" {
		t.Fatalf("expected downstream handler response to stay intact, got %q", body)
	}
}

// TestSourceGlobalMiddlewareRegistrarCapturesHandlerResponse verifies outer
// global middleware can observe the unified response after HandlerResponse completes.
func TestSourceGlobalMiddlewareRegistrarCapturesHandlerResponse(t *testing.T) {
	server, _ := newSourceHTTPTestServer(t, true)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(testSourceHTTPEchoHandler)
	})

	captured := ""
	registrar := &sourceGlobalMiddlewareRegistrar{
		server:   server,
		pluginID: "plugin-demo",
		service:  newSourceHTTPTestService(true),
	}
	if err := registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		request.Middleware.Next()
		captured = request.Response.BufferString()
	}); err != nil {
		t.Fatalf("expected middleware registration to succeed, got %v", err)
	}

	startSourceHTTPTestServer(t, server)
	_, _ = sourceHTTPTestGet(t, server, "/api/v1/echo")
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

// TestSourceGlobalMiddlewareRegistrarObservesDownstreamExitAll verifies outer
// global middleware still resumes after downstream handlers terminate the request early.
func TestSourceGlobalMiddlewareRegistrarObservesDownstreamExitAll(t *testing.T) {
	server, _ := newSourceHTTPTestServer(t, true)
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
	registrar := &sourceGlobalMiddlewareRegistrar{
		server:   server,
		pluginID: "plugin-demo",
		service:  newSourceHTTPTestService(true),
	}
	if err := registrar.Bind("/api/v1/*", func(request *ghttp.Request) {
		request.Middleware.Next()
		captured = request.Response.BufferString()
	}); err != nil {
		t.Fatalf("expected middleware registration to succeed, got %v", err)
	}

	startSourceHTTPTestServer(t, server)
	_, _ = sourceHTTPTestGet(t, server, "/api/v1/stop")
	if captured != "stopped" {
		t.Fatalf("expected global middleware to capture downstream early-exit response, got %q", captured)
	}
}

// newSourceHTTPTestService creates a minimal integration service with loaded
// platform enablement state for the test plugin.
func newSourceHTTPTestService(enabled bool) *serviceImpl {
	service := &serviceImpl{sharedState: newSharedState()}
	service.SetPluginEnabledState("plugin-demo", enabled)
	return service
}

// newSourceHTTPNoopMiddlewares creates a published route middleware directory
// with no-op handlers for registrar wiring tests.
func newSourceHTTPNoopMiddlewares() pluginhost.RouteMiddlewares {
	noop := func(r *ghttp.Request) {}
	return pluginhost.NewRouteMiddlewares(noop, noop, noop, noop, noop, noop, noop, noop)
}

// newSourceHTTPTestServer creates a named GoFrame test server and root group.
func newSourceHTTPTestServer(t *testing.T, listen bool) (*ghttp.Server, *ghttp.RouterGroup) {
	t.Helper()
	server := g.Server(sourceHTTPTestServerName(t))
	if listen {
		server.SetDumpRouterMap(false)
		server.SetPort(0)
	}

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})
	return server, rootGroup
}

// sourceHTTPTestServerName returns a unique GoFrame server name for each test.
func sourceHTTPTestServerName(t *testing.T) string {
	t.Helper()
	replacer := strings.NewReplacer("/", "-", " ", "-", "_", "-")
	return "integration-source-http-" + replacer.Replace(t.Name())
}

// startSourceHTTPTestServer starts a GoFrame test server and registers cleanup.
func startSourceHTTPTestServer(t *testing.T, server *ghttp.Server) {
	t.Helper()
	server.Start()
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Logf("shutdown source HTTP test server failed: %v", err)
		}
	})
	time.Sleep(100 * time.Millisecond)
}

// sourceHTTPTestGet issues one HTTP GET request against a started GoFrame server.
func sourceHTTPTestGet(t *testing.T, server *ghttp.Server, path string) (int, string) {
	t.Helper()
	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", server.GetListenedPort(), path))
	if err != nil {
		t.Fatalf("expected HTTP GET %s to succeed, got %v", path, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Logf("close response body for %s failed: %v", path, closeErr)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("expected response body for %s to be readable, got %v", path, err)
	}
	return response.StatusCode, string(body)
}
