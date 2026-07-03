// This file verifies hosted OpenAPI route binding and request-origin handling.

package httpstartup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
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

// unsafeFieldString reads an unexported string field value for test assertions.
func unsafeFieldString(value reflect.Value) string {
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().String()
}
