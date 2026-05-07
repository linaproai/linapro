// This file verifies hosted OpenAPI binding and plugin asset path parsing.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
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

	bindHostedOpenAPIDocs(
		context.Background(),
		server,
		&fakeApiDocService{document: &goai.OpenApiV3{}},
		"/api.json",
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

// TestUploadedFileAccessRouteIsPublic verifies direct upload URLs remain
// browser-loadable without making the whole file controller public.
func TestUploadedFileAccessRouteIsPublic(t *testing.T) {
	ctx := context.Background()
	server := ghttp.GetServer("cmd-http-upload-public-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)

	configSvc := config.New()
	jobRegistry := jobhandlersvc.New()
	runtime := &httpRuntime{
		configSvc:     configSvc,
		clusterSvc:    cluster.New(configSvc.GetCluster(ctx)),
		pluginSvc:     pluginsvc.New(nil),
		jobRegistry:   jobRegistry,
		jobMgmtSvc:    jobmgmtsvc.New(configSvc, jobRegistry, nil),
		middlewareSvc: middleware.New(),
	}
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
