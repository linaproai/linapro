// This file verifies hosted OpenAPI binding, plugin asset path parsing, and
// route precedence for host API and frontend fallback handlers.

package httpstartup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"unsafe"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/dict"
	filesvc "lina-core/internal/service/file"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/menu"
	"lina-core/internal/service/middleware"
	"lina-core/internal/service/notify"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/internal/service/sysconfig"
	sysinfosvc "lina-core/internal/service/sysinfo"
	"lina-core/internal/service/user"
	"lina-core/internal/service/usermsg"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	pluginservicehostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	pluginservicemanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	pluginserviceconfig "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// fakeApiDocService is the apidoc stub used by hosted OpenAPI binding tests.
type fakeApiDocService struct {
	document *goai.OpenApiV3
}

// Build returns the preconfigured OpenAPI document for hosted-doc binding tests.
func (f *fakeApiDocService) Build(_ context.Context, _ *ghttp.Server) (*goai.OpenApiV3, error) {
	return f.document, nil
}

// ResolveRouteText returns fallback route text for hosted-doc binding tests.
func (f *fakeApiDocService) ResolveRouteText(_ context.Context, input apidoc.RouteTextInput) apidoc.RouteTextOutput {
	return apidoc.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
}

// ResolveRouteTexts returns fallback route text for hosted-doc binding tests.
func (f *fakeApiDocService) ResolveRouteTexts(_ context.Context, inputs []apidoc.RouteTextInput) []apidoc.RouteTextOutput {
	outputs := make([]apidoc.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, apidoc.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
	}
	return outputs
}

// FindRouteTitleOperationKeys returns no route-title matches for hosted-doc binding tests.
func (f *fakeApiDocService) FindRouteTitleOperationKeys(_ context.Context, _ string) []string {
	return []string{}
}

// TestBindHostedOpenAPIDocsDisablesBuiltInEndpointsAndBindsConfiguredPath
// verifies the host-owned OpenAPI route replaces the built-in GoFrame endpoints.
func TestBindHostedOpenAPIDocsDisablesBuiltInEndpointsAndBindsConfiguredPath(t *testing.T) {
	server := ghttp.GetServer("cmd-http-bind-openapi-" + t.Name())
	server.SetOpenApiPath("/builtin-api.json")
	server.SetSwaggerPath("/swagger")

	testI18nSvc := i18nsvc.New(bizctx.New(), config.New(), cachecoord.Default(cluster.New(config.New().GetCluster(context.Background()))))
	bindHostedOpenAPIDocs(
		context.Background(),
		server,
		&fakeApiDocService{document: &goai.OpenApiV3{}},
		"/api.json",
		testI18nSvc,
		bizctx.New(),
	)

	if server.GetOpenApiPath() != "" {
		t.Fatalf("expected built-in openapi path to be cleared, got %q", server.GetOpenApiPath())
	}

	configValue := reflect.ValueOf(server).Elem().FieldByName("config")
	swaggerPath := unsafeFieldString(configValue.FieldByName("SwaggerPath"))
	if swaggerPath != "" {
		t.Fatalf("expected built-in swagger path to be cleared, got %q", swaggerPath)
	}

	foundHostedRoute := false
	for _, route := range server.GetRoutes() {
		if route.Route == "/api.json" {
			foundHostedRoute = true
			break
		}
	}
	if !foundHostedRoute {
		t.Fatal("expected hosted OpenAPI route to be bound at /api.json")
	}
}

// TestBindHostedOpenAPIDocsUsesRequestOrigin verifies the generated OpenAPI
// server URL follows the request entrypoint instead of static metadata.
func TestBindHostedOpenAPIDocsUsesRequestOrigin(t *testing.T) {
	testCases := []struct {
		name       string
		host       string
		proto      string
		wantOrigin string
	}{
		{
			name:       "backend direct mapped port",
			host:       "127.0.0.1:18088",
			wantOrigin: "http://127.0.0.1:18088",
		},
		{
			name:       "frontend proxy reaches backend port",
			host:       "localhost:9120",
			wantOrigin: "http://localhost:9120",
		},
		{
			name:       "https reverse proxy",
			host:       "api.example.com:8443",
			proto:      "https",
			wantOrigin: "https://api.example.com:8443",
		},
	}

	testI18nSvc := i18nsvc.New(bizctx.New(), config.New(), cachecoord.Default(cluster.New(config.New().GetCluster(context.Background()))))
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := ghttp.GetServer("cmd-http-openapi-origin-" + guid.S())
			server.SetPort(0)
			server.SetDumpRouterMap(false)
			bindHostedOpenAPIDocs(
				context.Background(),
				server,
				&fakeApiDocService{document: &goai.OpenApiV3{
					Servers: &goai.Servers{
						{
							URL:         "http://localhost:9120",
							Description: "CoreHostEndpoint",
						},
					},
				}},
				"/api.json",
				testI18nSvc,
				bizctx.New(),
			)
			if err := server.Start(); err != nil {
				t.Fatalf("start OpenAPI origin test server: %v", err)
			}
			t.Cleanup(func() {
				if err := server.Shutdown(); err != nil {
					t.Fatalf("shutdown OpenAPI origin test server: %v", err)
				}
			})

			request, err := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://127.0.0.1:%d/api.json", server.GetListenedPort()),
				nil,
			)
			if err != nil {
				t.Fatalf("create OpenAPI origin request: %v", err)
			}
			request.Host = testCase.host
			if testCase.proto != "" {
				request.Header.Set("X-Forwarded-Proto", testCase.proto)
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("request hosted OpenAPI document: %v", err)
			}
			defer func() {
				if closeErr := response.Body.Close(); closeErr != nil {
					t.Fatalf("close hosted OpenAPI response body: %v", closeErr)
				}
			}()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.StatusCode)
			}

			var payload struct {
				Servers []struct {
					URL         string `json:"url"`
					Description string `json:"description"`
				} `json:"servers"`
			}
			if err = json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode hosted OpenAPI response: %v", err)
			}
			if len(payload.Servers) != 1 {
				t.Fatalf("expected one OpenAPI server, got %#v", payload.Servers)
			}
			if payload.Servers[0].URL != testCase.wantOrigin {
				t.Fatalf("expected server url %q, got %q", testCase.wantOrigin, payload.Servers[0].URL)
			}
			if payload.Servers[0].Description != "CoreHostEndpoint" {
				t.Fatalf("expected server description to stay, got %q", payload.Servers[0].Description)
			}
		})
	}
}

// TestUploadedFileAccessRouteIsPublic verifies direct upload URLs remain
// browser-loadable without making the whole file controller public.
func TestUploadedFileAccessRouteIsPublic(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-upload-public-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)

	if err := server.Start(); err != nil {
		t.Fatalf("start route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown route test server: %v", err)
		}
	})

	uploadAccess := mustFindRoute(t, server, "GET", "/api/v1/uploads/*path")
	if strings.Contains(uploadAccess.Middleware, "Service.Auth") {
		t.Fatalf("expected upload URL access route to be public, middleware=%s", uploadAccess.Middleware)
	}
	if strings.Contains(uploadAccess.Middleware, "Service.Permission") {
		t.Fatalf("expected upload URL access route to skip permission middleware, middleware=%s", uploadAccess.Middleware)
	}

	fileUpload := mustFindRoute(t, server, "POST", "/api/v1/file/upload")
	if !strings.Contains(fileUpload.Middleware, "Service.Auth") {
		t.Fatalf("expected file upload route to remain authenticated, middleware=%s", fileUpload.Middleware)
	}
	if !strings.Contains(fileUpload.Middleware, "Service.Permission") {
		t.Fatalf("expected file upload route to keep permission middleware, middleware=%s", fileUpload.Middleware)
	}

	userBatchUpdate := mustFindRoute(t, server, "PUT", "/api/v1/user")
	if !strings.Contains(userBatchUpdate.Middleware, "Service.Auth") {
		t.Fatalf("expected user batch-update route to remain authenticated, middleware=%s", userBatchUpdate.Middleware)
	}
	if !strings.Contains(userBatchUpdate.Middleware, "Service.Permission") {
		t.Fatalf("expected user batch-update route to keep permission middleware, middleware=%s", userBatchUpdate.Middleware)
	}
}

// TestPluginManagementRuntimeRoutesAreBound verifies plugin runtime-management
// endpoints added by the upgrade workflow are reachable through the protected API tree.
func TestPluginManagementRuntimeRoutesAreBound(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-plugin-runtime-routes-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)

	if err := server.Start(); err != nil {
		t.Fatalf("start plugin route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown plugin route test server: %v", err)
		}
	})

	routes := []struct {
		method string
		route  string
	}{
		{method: "GET", route: "/api/v1/plugins/{id}"},
		{method: "GET", route: "/api/v1/plugins/{id}/dependencies"},
		{method: "GET", route: "/api/v1/plugins/{id}/upgrade/preview"},
		{method: "POST", route: "/api/v1/plugins/{id}/upgrade"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.route, func(t *testing.T) {
			item := mustFindRoute(t, server, route.method, route.route)
			if !strings.Contains(item.Middleware, "Service.Auth") {
				t.Fatalf("expected plugin runtime route to remain authenticated, middleware=%s", item.Middleware)
			}
			if !strings.Contains(item.Middleware, "Service.Permission") {
				t.Fatalf("expected plugin runtime route to keep permission middleware, middleware=%s", item.Middleware)
			}
		})
	}
}

// TestDynamicPluginRootRoutesPrecedeSPAFallback verifies root-level dynamic
// plugin paths are claimed before the catch-all frontend fallback can redirect
// API-style requests to the host SPA entry.
func TestDynamicPluginRootRoutesPrecedeSPAFallback(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-dynamic-plugin-root-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	bindHostAPIRoutes(ctx, server, runtime)
	if err := bindFrontendAssetRoutes(ctx, server, runtime.pluginSvc, "/admin"); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("start dynamic route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown dynamic route test server: %v", err)
		}
	})

	response, err := http.Get(fmt.Sprintf(
		"http://127.0.0.1:%d/x/plugin-dev-route-missing/api/v1/backend-summary",
		server.GetListenedPort(),
	))
	if err != nil {
		t.Fatalf("request root dynamic plugin route: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatalf("close dynamic route response body: %v", closeErr)
		}
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dynamic route response body: %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected dynamic route 404, got status=%d body=%q", response.StatusCode, string(body))
	}
	if strings.TrimSpace(string(body)) != "Dynamic route not found" {
		t.Fatalf("expected dynamic route not found body, got %q", string(body))
	}
	if response.Header.Get("Location") != "" {
		t.Fatalf("expected no SPA redirect for dynamic route, got location=%q", response.Header.Get("Location"))
	}
}

// TestFrontendAssetFallbackIsScopedToWorkspaceBasePath verifies the final SPA
// fallback only serves the built-in workspace under the configured base path.
func TestFrontendAssetFallbackIsScopedToWorkspaceBasePath(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-workspace-fallback-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	server.BindHandler("/*", func(r *ghttp.Request) {
		r.Response.WriteStatus(http.StatusNotFound)
		r.ExitAll()
	})
	if err := bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/admin", testFrontendFS()); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("start workspace fallback test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown workspace fallback test server: %v", err)
		}
	})

	adminResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/admin", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request admin workspace base: %v", err)
	}
	defer func() {
		if closeErr := adminResp.Body.Close(); closeErr != nil {
			t.Fatalf("close admin workspace response body: %v", closeErr)
		}
	}()
	if adminResp.StatusCode != http.StatusOK {
		adminBody, readErr := io.ReadAll(adminResp.Body)
		if readErr != nil {
			t.Fatalf("read admin workspace response body: %v", readErr)
		}
		for _, route := range server.GetRoutes() {
			t.Logf("registered route method=%s route=%s", route.Method, route.Route)
		}
		t.Fatalf("expected admin workspace status 200, got %d body=%q", adminResp.StatusCode, string(adminBody))
	}

	rootResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request root path: %v", err)
	}
	defer func() {
		if closeErr := rootResp.Body.Close(); closeErr != nil {
			t.Fatalf("close root response body: %v", closeErr)
		}
	}()
	if rootResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected root path to avoid SPA fallback with 404, got %d", rootResp.StatusCode)
	}

	for _, assetPath := range []string{"/admin/logo.webp", "/admin/stoplight/apidocs.html"} {
		t.Run(assetPath, func(t *testing.T) {
			response, requestErr := http.Get(fmt.Sprintf(
				"http://127.0.0.1:%d%s",
				server.GetListenedPort(),
				assetPath,
			))
			if requestErr != nil {
				t.Fatalf("request workspace asset path %s: %v", assetPath, requestErr)
			}
			defer func() {
				if closeErr := response.Body.Close(); closeErr != nil {
					t.Fatalf("close workspace asset response body: %v", closeErr)
				}
			}()
			body, readErr := io.ReadAll(response.Body)
			if readErr != nil {
				t.Fatalf("read workspace asset response body: %v", readErr)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected workspace asset %s status 200, got %d body=%q", assetPath, response.StatusCode, string(body))
			}
			if len(body) == 0 {
				t.Fatalf("expected workspace asset %s to return content", assetPath)
			}
		})
	}
}

// TestFrontendAssetFallbackClaimsHostedPluginAssetNamespace verifies the final
// asset handler owns the hosted plugin asset namespace even when a broader root
// wildcard route has already been registered.
func TestFrontendAssetFallbackClaimsHostedPluginAssetNamespace(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-plugin-asset-fallback-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	server.BindHandler("/*", func(r *ghttp.Request) {
		r.Response.WriteStatus(http.StatusConflict)
		r.Response.Write("root wildcard")
		r.ExitAll()
	})
	if err := bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/admin", testFrontendFS()); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("start plugin asset fallback test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown plugin asset fallback test server: %v", err)
		}
	})

	response, err := http.Get(fmt.Sprintf(
		"http://127.0.0.1:%d/x-assets/plugin-missing/v0.1.0/app.js",
		server.GetListenedPort(),
	))
	if err != nil {
		t.Fatalf("request hosted plugin asset path: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatalf("close hosted plugin asset response body: %v", closeErr)
		}
	}()
	if response.StatusCode != http.StatusNotFound {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatalf("read hosted plugin asset response body: %v", readErr)
		}
		t.Fatalf("expected hosted plugin asset handler 404, got status=%d body=%q", response.StatusCode, string(body))
	}
}

// TestFrontendAssetFallbackSupportsRootWorkspaceBasePath verifies dedicated
// admin-domain deployments can mount the workspace at `/` while keeping host
// and plugin namespaces outside SPA fallback.
func TestFrontendAssetFallbackSupportsRootWorkspaceBasePath(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-root-workspace-fallback-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	if err := bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/", testFrontendFS()); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("start root workspace fallback test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown root workspace fallback test server: %v", err)
		}
	})

	rootResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request root workspace base: %v", err)
	}
	defer func() {
		if closeErr := rootResp.Body.Close(); closeErr != nil {
			t.Fatalf("close root workspace response body: %v", closeErr)
		}
	}()
	if rootResp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(rootResp.Body)
		if readErr != nil {
			t.Fatalf("read root workspace response body: %v", readErr)
		}
		t.Fatalf("expected root workspace status 200, got %d body=%q", rootResp.StatusCode, string(body))
	}

	nestedResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/system/user", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request nested root workspace route: %v", err)
	}
	defer func() {
		if closeErr := nestedResp.Body.Close(); closeErr != nil {
			t.Fatalf("close nested root workspace response body: %v", closeErr)
		}
	}()
	if nestedResp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(nestedResp.Body)
		if readErr != nil {
			t.Fatalf("read nested root workspace response body: %v", readErr)
		}
		t.Fatalf("expected nested root workspace status 200, got %d body=%q", nestedResp.StatusCode, string(body))
	}

	for _, reservedPath := range []string{
		"/api/v1/missing",
		"/x/plugin-missing/api/v1/missing",
		"/x-assets/plugin-missing/v0.1.0/app.js",
		"/api.json",
	} {
		t.Run(reservedPath, func(t *testing.T) {
			response, requestErr := http.Get(fmt.Sprintf(
				"http://127.0.0.1:%d%s",
				server.GetListenedPort(),
				reservedPath,
			))
			if requestErr != nil {
				t.Fatalf("request reserved path %s: %v", reservedPath, requestErr)
			}
			defer func() {
				if closeErr := response.Body.Close(); closeErr != nil {
					t.Fatalf("close reserved path response body: %v", closeErr)
				}
			}()
			if response.StatusCode != http.StatusNotFound {
				body, readErr := io.ReadAll(response.Body)
				if readErr != nil {
					t.Fatalf("read reserved path response body: %v", readErr)
				}
				t.Fatalf("expected reserved path %s to avoid SPA fallback with 404, got %d body=%q", reservedPath, response.StatusCode, string(body))
			}
		})
	}
}

// TestRootWorkspaceFallbackDoesNotOverrideHostedOpenAPI verifies the startup
// binding order keeps the explicit OpenAPI route reachable after root fallback.
func TestRootWorkspaceFallbackDoesNotOverrideHostedOpenAPI(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-root-workspace-openapi-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	if err := bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/", testFrontendFS()); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}
	bindHostedOpenAPIDocs(
		ctx,
		server,
		&fakeApiDocService{document: &goai.OpenApiV3{}},
		"/api.json",
		runtime.i18nSvc,
		runtime.bizCtxSvc,
	)

	if err := server.Start(); err != nil {
		t.Fatalf("start root workspace OpenAPI test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown root workspace OpenAPI test server: %v", err)
		}
	})

	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api.json", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request hosted OpenAPI under root workspace: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatalf("close hosted OpenAPI response body: %v", closeErr)
		}
	}()
	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatalf("read hosted OpenAPI response body: %v", readErr)
		}
		t.Fatalf("expected hosted OpenAPI status 200, got %d body=%q", response.StatusCode, string(body))
	}
}

// TestRootWorkspaceFallbackConflictsWithExistingRootRoute verifies root-mounted
// workspace fallback cannot silently replace an already registered root route.
func TestRootWorkspaceFallbackConflictsWithExistingRootRoute(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_ROOT_WORKSPACE_CONFLICT_CHILD") != "" {
		server := ghttp.GetServer("cmd-http-root-workspace-conflict-" + guid.S())
		server.SetPort(0)
		server.SetDumpRouterMap(false)
		server.BindHandler("/", func(r *ghttp.Request) {
			r.Response.Write("root")
			r.ExitAll()
		})
		if err := bindFrontendAssetRoutesWithFS(server, nil, "/", testFrontendFS()); err != nil {
			t.Fatalf("bind frontend asset routes returned unexpected error: %v", err)
		}
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestRootWorkspaceFallbackConflictsWithExistingRootRoute$")
	command.Env = append(os.Environ(), "LINAPRO_TEST_ROOT_WORKSPACE_CONFLICT_CHILD=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("expected root workspace fallback conflict child process to fail, output=%s", string(output))
	}
	if !strings.Contains(string(output), "duplicated route registry") {
		t.Fatalf("expected duplicated route registry diagnostic, got output=%s", string(output))
	}
}

// TestTrimWorkspaceRequestPath verifies workspace base stripping used by the
// final frontend fallback is stable for base and nested admin routes.
func TestTrimWorkspaceRequestPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		basePath  string
		wantPath  string
		wantMatch bool
	}{
		{
			name:      "base path",
			path:      "admin",
			basePath:  "admin",
			wantPath:  "index.html",
			wantMatch: true,
		},
		{
			name:      "nested route",
			path:      "admin/system/user",
			basePath:  "/admin",
			wantPath:  "system/user",
			wantMatch: true,
		},
		{
			name:      "prefix sibling",
			path:      "administer",
			basePath:  "/admin",
			wantMatch: false,
		},
		{
			name:      "root path",
			path:      "",
			basePath:  "/admin",
			wantMatch: false,
		},
		{
			name:      "root base exact path",
			path:      "",
			basePath:  "/",
			wantPath:  "index.html",
			wantMatch: true,
		},
		{
			name:      "root base nested route",
			path:      "system/user",
			basePath:  "/",
			wantPath:  "system/user",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotMatch := trimWorkspaceRequestPath(tt.path, tt.basePath)
			if gotMatch != tt.wantMatch {
				t.Fatalf("expected match=%v, got %v", tt.wantMatch, gotMatch)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("expected path=%q, got %q", tt.wantPath, gotPath)
			}
		})
	}
}

// TestFrontendAssetFallbackProxiesWorkspaceBasePathInDevelopment verifies the
// optional Vite proxy is scoped to the workspace path and leaves root routes free.
func TestFrontendAssetFallbackProxiesWorkspaceBasePathInDevelopment(t *testing.T) {
	ctx := context.Background()
	devServer := http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/admin/", "/admin/system/user", "/admin/logo.webp", "/admin/stoplight/apidocs.html", "/admin/stoplight/styles.min.css":
			default:
				t.Errorf("expected proxied workspace path, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("vite-admin-dev:" + r.URL.Path)); err != nil {
				t.Errorf("write dev proxy response: %v", err)
			}
		}),
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen frontend dev server: %v", err)
	}
	go func() {
		if serveErr := devServer.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			t.Errorf("serve frontend dev server: %v", serveErr)
		}
	}()
	t.Cleanup(func() {
		if shutdownErr := devServer.Shutdown(context.Background()); shutdownErr != nil {
			t.Fatalf("shutdown frontend dev server: %v", shutdownErr)
		}
	})
	t.Setenv(frontendDevServerURLEnv, "http://"+listener.Addr().String())

	server := ghttp.GetServer("cmd-http-workspace-dev-proxy-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	if err = bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/admin", testFrontendFS()); err != nil {
		t.Fatalf("bind frontend asset routes: %v", err)
	}

	if err = server.Start(); err != nil {
		t.Fatalf("start workspace dev proxy test server: %v", err)
	}
	t.Cleanup(func() {
		if shutdownErr := server.Shutdown(); shutdownErr != nil {
			t.Fatalf("shutdown workspace dev proxy test server: %v", shutdownErr)
		}
	})

	baseResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/admin", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request proxied admin workspace base: %v", err)
	}
	defer func() {
		if closeErr := baseResp.Body.Close(); closeErr != nil {
			t.Fatalf("close proxied admin base response body: %v", closeErr)
		}
	}()
	baseBody, err := io.ReadAll(baseResp.Body)
	if err != nil {
		t.Fatalf("read proxied admin base response body: %v", err)
	}
	if string(baseBody) != "vite-admin-dev:/admin/" {
		t.Fatalf("expected normalized dev proxy base body, got %q", string(baseBody))
	}

	adminResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/admin/system/user", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request proxied admin workspace path: %v", err)
	}
	defer func() {
		if closeErr := adminResp.Body.Close(); closeErr != nil {
			t.Fatalf("close proxied admin response body: %v", closeErr)
		}
	}()
	body, err := io.ReadAll(adminResp.Body)
	if err != nil {
		t.Fatalf("read proxied admin response body: %v", err)
	}
	if string(body) != "vite-admin-dev:/admin/system/user" {
		t.Fatalf("expected dev proxy body, got %q", string(body))
	}

	for _, assetPath := range []string{"/admin/logo.webp", "/admin/stoplight/apidocs.html", "/admin/stoplight/styles.min.css"} {
		t.Run(assetPath, func(t *testing.T) {
			assetResp, requestErr := http.Get(fmt.Sprintf(
				"http://127.0.0.1:%d%s",
				server.GetListenedPort(),
				assetPath,
			))
			if requestErr != nil {
				t.Fatalf("request workspace asset with dev proxy enabled: %v", requestErr)
			}
			defer func() {
				if closeErr := assetResp.Body.Close(); closeErr != nil {
					t.Fatalf("close workspace asset response body: %v", closeErr)
				}
			}()
			assetBody, readErr := io.ReadAll(assetResp.Body)
			if readErr != nil {
				t.Fatalf("read workspace asset response body: %v", readErr)
			}
			expectedBody := "vite-admin-dev:" + assetPath
			if string(assetBody) != expectedBody {
				t.Fatalf("expected dev proxied workspace asset body %q, got %q", expectedBody, string(assetBody))
			}
		})
	}

	rootResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request root path with dev proxy enabled: %v", err)
	}
	defer func() {
		if closeErr := rootResp.Body.Close(); closeErr != nil {
			t.Fatalf("close root response body: %v", closeErr)
		}
	}()
	if rootResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected root path to avoid dev proxy fallback with 404, got %d", rootResp.StatusCode)
	}
}

// TestFrontendAssetFallbackProxiesRootWorkspaceBasePathInDevelopment verifies
// the optional Vite proxy also works when the workspace is mounted at `/`.
func TestFrontendAssetFallbackProxiesRootWorkspaceBasePathInDevelopment(t *testing.T) {
	ctx := context.Background()
	devServer := http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" && r.URL.Path != "/system/user" {
				t.Errorf("expected proxied root workspace path, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("vite-root-dev:" + r.URL.Path)); err != nil {
				t.Errorf("write root dev proxy response: %v", err)
			}
		}),
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen root frontend dev server: %v", err)
	}
	go func() {
		if serveErr := devServer.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			t.Errorf("serve root frontend dev server: %v", serveErr)
		}
	}()
	t.Cleanup(func() {
		if shutdownErr := devServer.Shutdown(context.Background()); shutdownErr != nil {
			t.Fatalf("shutdown root frontend dev server: %v", shutdownErr)
		}
	})
	t.Setenv(frontendDevServerURLEnv, "http://"+listener.Addr().String())

	server := ghttp.GetServer("cmd-http-root-workspace-dev-proxy-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	runtime := newRouteBindingTestRuntime(ctx)
	if err = bindFrontendAssetRoutesWithFS(server, runtime.pluginSvc, "/", testFrontendFS()); err != nil {
		t.Fatalf("bind root frontend asset routes: %v", err)
	}

	if err = server.Start(); err != nil {
		t.Fatalf("start root workspace dev proxy test server: %v", err)
	}
	t.Cleanup(func() {
		if shutdownErr := server.Shutdown(); shutdownErr != nil {
			t.Fatalf("shutdown root workspace dev proxy test server: %v", shutdownErr)
		}
	})

	baseResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request proxied root workspace base: %v", err)
	}
	defer func() {
		if closeErr := baseResp.Body.Close(); closeErr != nil {
			t.Fatalf("close proxied root base response body: %v", closeErr)
		}
	}()
	baseBody, err := io.ReadAll(baseResp.Body)
	if err != nil {
		t.Fatalf("read proxied root base response body: %v", err)
	}
	if string(baseBody) != "vite-root-dev:/" {
		t.Fatalf("expected normalized root dev proxy base body, got %q", string(baseBody))
	}

	nestedResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/system/user", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request proxied root workspace path: %v", err)
	}
	defer func() {
		if closeErr := nestedResp.Body.Close(); closeErr != nil {
			t.Fatalf("close proxied root workspace response body: %v", closeErr)
		}
	}()
	nestedBody, err := io.ReadAll(nestedResp.Body)
	if err != nil {
		t.Fatalf("read proxied root workspace response body: %v", err)
	}
	if string(nestedBody) != "vite-root-dev:/system/user" {
		t.Fatalf("expected root dev proxy body, got %q", string(nestedBody))
	}

	reservedResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/missing", server.GetListenedPort()))
	if err != nil {
		t.Fatalf("request reserved path with root dev proxy enabled: %v", err)
	}
	defer func() {
		if closeErr := reservedResp.Body.Close(); closeErr != nil {
			t.Fatalf("close reserved root dev proxy response body: %v", closeErr)
		}
	}()
	if reservedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected reserved path to avoid root dev proxy fallback with 404, got %d", reservedResp.StatusCode)
	}
}

// testFrontendFS returns the minimal host frontend bundle needed by fallback
// route tests. The real embedded bundle is generated and not tracked in Git.
func testFrontendFS() fs.FS {
	return fstest.MapFS{
		"index.html": {
			Data: []byte("<!doctype html><title>LinaPro Test Workspace</title>"),
			Mode: 0o644,
		},
		"logo.webp": {
			Data: []byte("test-logo"),
			Mode: 0o644,
		},
		"stoplight/apidocs.html": {
			Data: []byte("<!doctype html><title>API Documentation</title>"),
			Mode: 0o644,
		},
		"assets/app.js": {
			Data: []byte("console.log('linapro-test-workspace');"),
			Mode: 0o644,
		},
	}
}

// newRouteBindingTestRuntime creates the shared service graph required by
// route-binding tests without starting cluster, plugin, or cron lifecycles.
func newRouteBindingTestRuntime(ctx context.Context) *httpRuntime {
	configSvc := config.New()
	clusterSvc := cluster.New(configSvc.GetCluster(ctx))
	bizCtxSvc := bizctx.New()
	sessionStore := session.NewDBStore()
	cacheCoordSvc := cachecoord.Default(clusterSvc)
	i18nService := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	lockerSvc := locker.New()
	pluginRuntime := pluginsvc.NewRuntimeDelegate()
	orgCapSvc := orgspi.New(orgspi.NewManager(), pluginRuntime)
	orgProjection := orgCapSvc
	tenantSvc := tenantspi.New(tenantspi.NewManager(), pluginRuntime, bizCtxSvc)
	roleSvc := role.New(pluginRuntime, bizCtxSvc, configSvc, i18nService, orgCapSvc, tenantSvc)
	kvCacheSvc := kvcache.New()
	dictSvc := dict.New(i18nService)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	roleSvc.SetDataScopeService(scopeSvc)
	menuSvc := menu.New(pluginRuntime, i18nService, roleSvc, tenantSvc)
	notifySvc := notify.New(tenantSvc)
	authSvc := auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
	fileSvc := filesvc.New(configSvc, filesvc.NewLocalStorage(configSvc.GetUploadPath(ctx)), bizCtxSvc, dictSvc, scopeSvc)
	sysConfigSvc := sysconfig.New(configSvc, i18nService)
	sysInfoSvc, err := sysinfosvc.New(configSvc, clusterSvc, nil, cacheCoordSvc)
	if err != nil {
		panic(err)
	}
	userSvc := user.New(authSvc, bizCtxSvc, i18nService, orgCapSvc, orgCapSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc, tenantSvc, tenantSvc)
	userMsgSvc := usermsg.New(bizCtxSvc, notifySvc, i18nService)
	jobRegistry := jobhandlersvc.New()
	hostConfigReader, ok := configSvc.(pluginservicehostconfig.RawConfigReader)
	if !ok {
		panic("test config service does not support raw host config reads")
	}
	hostConfigSvc := pluginservicehostconfig.New(hostConfigReader)
	hostLockSvc, err := hostlock.New(lockerSvc)
	if err != nil {
		panic(err)
	}
	capabilities, err := pluginsvc.NewHostServices(
		apidoc.New(configSvc, bizCtxSvc, i18nService, pluginRuntime),
		authSvc,
		authSvc.(auth.TenantTokenIssuer),
		bizCtxSvc,
		hostConfigSvc,
		scopeSvc,
		i18nService,
		pluginRuntime,
		pluginRuntime,
		sessionStore,
		aitext.New(nil, pluginRuntime),
		orgCapSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
		hostLockSvc,
		pluginsvc.NewStorageProviderRuntime(pluginRuntime),
		pluginsvc.NewLocalStorageProvider(configSvc.GetPluginDynamicStoragePath(ctx)),
	)
	if err != nil {
		panic(err)
	}
	pluginSvc, err := pluginsvc.New(
		clusterSvc,
		configSvc,
		bizCtxSvc,
		cacheCoordSvc,
		i18nService,
		sessionStore,
		roleSvc,
		lockerSvc,
		nil,
		capabilities,
		orgCapSvc,
		tenantSvc,
		tenantSvc,
		tenantSvc,
		pluginserviceconfig.NewConfigFactory("", ""),
		hostConfigSvc,
		pluginservicemanifest.NewFactory(""),
	)
	if err != nil {
		panic(err)
	}
	if err = pluginRuntime.BindService(pluginSvc); err != nil {
		panic(err)
	}
	return &httpRuntime{
		configSvc:       configSvc,
		clusterSvc:      clusterSvc,
		pluginSvc:       pluginSvc,
		authSvc:         authSvc,
		authTokenIssuer: authSvc.(auth.TenantTokenIssuer),
		bizCtxSvc:       bizCtxSvc,
		i18nSvc:         i18nService,
		orgCapSvc:       orgCapSvc,
		orgProjection:   orgProjection,
		roleSvc:         roleSvc,
		sessionStore:    sessionStore,
		tenantSvc:       tenantSvc,
		kvCacheSvc:      kvCacheSvc,
		dictSvc:         dictSvc,
		fileSvc:         fileSvc,
		menuSvc:         menuSvc,
		notifySvc:       notifySvc,
		sysConfigSvc:    sysConfigSvc,
		sysInfoSvc:      sysInfoSvc,
		userSvc:         userSvc,
		userMsgSvc:      userMsgSvc,
		jobRegistry:     jobRegistry,
		jobMgmtSvc:      jobmgmtsvc.New(bizCtxSvc, configSvc, i18nService, jobRegistry, nil, scopeSvc),
		middlewareSvc:   middleware.New(authSvc, bizCtxSvc, configSvc, i18nService, roleSvc, tenantSvc),
	}
}

// TestParsePluginAssetRequestPath verifies hosted runtime asset URLs are parsed
// into plugin ID, version, and relative asset path segments.
func TestParsePluginAssetRequestPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantPluginID  string
		wantVersion   string
		wantAssetPath string
		wantOK        bool
	}{
		{
			name:          "hosted asset file",
			path:          "x-assets/linapro-demo-dynamic/v0.1.0/standalone.html",
			wantPluginID:  "linapro-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "standalone.html",
			wantOK:        true,
		},
		{
			name:          "embedded mount entry",
			path:          "/x-assets/linapro-demo-dynamic/v0.1.0/mount.js",
			wantPluginID:  "linapro-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "mount.js",
			wantOK:        true,
		},
		{
			name:          "version root path",
			path:          "/x-assets/linapro-demo-dynamic/v0.1.0/",
			wantPluginID:  "linapro-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "",
			wantOK:        true,
		},
		{
			name:   "non plugin path",
			path:   "/assets/index.js",
			wantOK: false,
		},
		{
			name:   "missing version",
			path:   "/x-assets/linapro-demo-dynamic",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPluginID, gotVersion, gotAssetPath, gotOK := parsePluginAssetRequestPath(tt.path)
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, gotOK)
			}
			if gotPluginID != tt.wantPluginID {
				t.Fatalf("expected pluginID=%q, got %q", tt.wantPluginID, gotPluginID)
			}
			if gotVersion != tt.wantVersion {
				t.Fatalf("expected version=%q, got %q", tt.wantVersion, gotVersion)
			}
			if gotAssetPath != tt.wantAssetPath {
				t.Fatalf("expected assetPath=%q, got %q", tt.wantAssetPath, gotAssetPath)
			}
		})
	}
}

// mustFindRoute returns one route item by method and path.
func mustFindRoute(t *testing.T, server *ghttp.Server, method string, route string) ghttp.RouterItem {
	t.Helper()

	for _, item := range server.GetRoutes() {
		if item.Method == method && item.Route == route {
			return item
		}
	}
	t.Fatalf("expected route %s %s to be registered", method, route)
	return ghttp.RouterItem{}
}

// unsafeFieldString reads an unexported string field value for test assertions.
func unsafeFieldString(value reflect.Value) string {
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().String()
}
