// This file verifies hosted OpenAPI binding and plugin asset path parsing.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
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
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/menu"
	"lina-core/internal/service/middleware"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/internal/service/sysconfig"
	sysinfosvc "lina-core/internal/service/sysinfo"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/internal/service/user"
	"lina-core/internal/service/usermsg"
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
	server.SetOpenApiPath("/legacy-api.json")
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
			host:       "localhost:8080",
			wantOrigin: "http://localhost:8080",
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
							URL:         "http://localhost:8080",
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

// newRouteBindingTestRuntime creates the shared service graph required by
// route-binding tests without starting cluster, plugin, or cron lifecycles.
func newRouteBindingTestRuntime(ctx context.Context) *httpRuntime {
	configSvc := config.New()
	clusterSvc := cluster.New(configSvc.GetCluster(ctx))
	bizCtxSvc := bizctx.New()
	sessionStore := session.NewDBStore()
	cacheCoordSvc := cachecoord.Default(clusterSvc)
	i18nService := i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
	pluginSvc, err := pluginsvc.New(clusterSvc, configSvc, bizCtxSvc, cacheCoordSvc, i18nService, sessionStore, nil)
	if err != nil {
		panic(err)
	}
	orgCapSvc := orgcap.New(pluginSvc)
	tenantSvc := tenantcapsvc.New(pluginSvc, bizCtxSvc)
	pluginSvc.SetTenantCapability(tenantSvc)
	roleSvc := role.New(pluginSvc, bizCtxSvc, configSvc, i18nService, nil, orgCapSvc, tenantSvc)
	kvCacheSvc := kvcache.New()
	dictSvc := dict.New(i18nService)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
	roleSvc.SetDataScopeService(scopeSvc)
	menuSvc := menu.New(pluginSvc, i18nService, roleSvc)
	notifySvc := notify.New(tenantSvc)
	authSvc := auth.New(configSvc, pluginSvc, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
	fileSvc := filesvc.New(configSvc, filesvc.NewLocalStorage(configSvc.GetUploadPath(ctx)), bizCtxSvc, dictSvc, scopeSvc)
	sysConfigSvc := sysconfig.New(configSvc, i18nService)
	sysInfoSvc, err := sysinfosvc.New(configSvc, clusterSvc, nil, cacheCoordSvc)
	if err != nil {
		panic(err)
	}
	userSvc := user.New(authSvc, bizCtxSvc, i18nService, orgCapSvc, roleSvc, scopeSvc, tenantSvc)
	userMsgSvc := usermsg.New(bizCtxSvc, notifySvc, i18nService)
	jobRegistry := jobhandlersvc.New()
	return &httpRuntime{
		configSvc:       configSvc,
		clusterSvc:      clusterSvc,
		pluginSvc:       pluginSvc,
		authSvc:         authSvc,
		authTokenIssuer: authSvc.(auth.TenantTokenIssuer),
		bizCtxSvc:       bizCtxSvc,
		i18nSvc:         i18nService,
		orgCapSvc:       orgCapSvc,
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
		middlewareSvc:   middleware.New(authSvc, bizCtxSvc, configSvc, i18nService, pluginSvc, roleSvc, tenantSvc),
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
			path:          "plugin-assets/plugin-demo-dynamic/v0.1.0/standalone.html",
			wantPluginID:  "plugin-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "standalone.html",
			wantOK:        true,
		},
		{
			name:          "embedded mount entry",
			path:          "/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js",
			wantPluginID:  "plugin-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "mount.js",
			wantOK:        true,
		},
		{
			name:          "version root path",
			path:          "/plugin-assets/plugin-demo-dynamic/v0.1.0/",
			wantPluginID:  "plugin-demo-dynamic",
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
			path:   "/plugin-assets/plugin-demo-dynamic",
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
