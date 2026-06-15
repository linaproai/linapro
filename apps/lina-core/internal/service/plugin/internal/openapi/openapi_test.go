// This file covers OpenAPI helper projection logic for dynamic plugin routes.

package openapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBuildRouteOpenAPIOperationUsesBridgeState verifies that projected
// responses follow the runtime bridge execution flag.
func TestBuildRouteOpenAPIOperationUsesBridgeState(t *testing.T) {
	operation := BuildRouteOpenAPIOperation("linapro-demo-dynamic", &protocol.RouteContract{
		Path:        "/api/v1/review-summary",
		Method:      http.MethodGet,
		Access:      protocol.AccessLogin,
		RequestType: "ReviewSummaryReq",
		Summary:     "Review Summary",
	}, &protocol.BridgeSpec{
		RouteExecution: true,
	})
	if operation == nil || operation.Responses["200"].Value == nil {
		t.Fatalf("expected executable bridge operation to expose 200 response, got %#v", operation)
	}
	if operation.Responses["500"].Value == nil {
		t.Fatalf("expected executable bridge operation to expose 500 response, got %#v", operation)
	}
	if operation.Responses["501"].Value != nil {
		t.Fatalf("expected executable bridge operation to hide 501 placeholder response, got %#v", operation)
	}
	if operation.Security == nil {
		t.Fatal("expected login route to project bearer security")
	}
	if operation.OperationID != "" {
		t.Fatalf("expected dynamic OpenAPI operationId to stay empty, got %s", operation.OperationID)
	}
	if len(operation.XExtensions) != 0 {
		t.Fatalf("expected dynamic OpenAPI operation to omit i18n extensions, got %#v", operation.XExtensions)
	}

	placeholder := BuildRouteOpenAPIOperation("linapro-demo-dynamic", &protocol.RouteContract{
		Path:        "/api/v1/placeholder",
		Method:      http.MethodGet,
		Access:      protocol.AccessPublic,
		RequestType: "PlaceholderReq",
	}, &protocol.BridgeSpec{
		RouteExecution: false,
	})
	if placeholder == nil || placeholder.Responses["501"].Value == nil {
		t.Fatalf("expected placeholder bridge operation to expose 501 response, got %#v", placeholder)
	}
	if placeholder.Responses["200"].Value != nil {
		t.Fatalf("expected placeholder bridge operation to omit 200 response, got %#v", placeholder)
	}
}

// TestBuildRoutePublicPathBuildsFixedPublicPath verifies that public route
// projection always uses the canonical dynamic plugin prefix.
func TestBuildRoutePublicPathBuildsFixedPublicPath(t *testing.T) {
	actual := BuildRoutePublicPath("plugin-openapi-projection", "/api/v1/review-summary")
	if actual != "/x/plugin-openapi-projection/api/v1/review-summary" {
		t.Fatalf("expected fixed public path projection, got %s", actual)
	}
}

// TestBuildRoutePublicPathPreservesPluginOwnedPathContent verifies `/api/v1`
// is only a plugin-local naming convention and not a forced public-path segment.
func TestBuildRoutePublicPathPreservesPluginOwnedPathContent(t *testing.T) {
	tests := []struct {
		name      string
		routePath string
		expected  string
	}{
		{
			name:      "api v2",
			routePath: "/api/v2/review-summary",
			expected:  "/x/plugin-openapi-projection/api/v2/review-summary",
		},
		{
			name:      "interface",
			routePath: "/interface/m1/review-summary",
			expected:  "/x/plugin-openapi-projection/interface/m1/review-summary",
		},
		{
			name:      "graphql",
			routePath: "/graphql",
			expected:  "/x/plugin-openapi-projection/graphql",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := BuildRoutePublicPath("plugin-openapi-projection", testCase.routePath)
			if actual != testCase.expected {
				t.Fatalf("expected public path %s, got %s", testCase.expected, actual)
			}
		})
	}
}

// TestProjectDynamicRoutesToOpenAPICacheKeys verifies dynamic route projection
// is cached by runtime revision, locale, and runtime bundle version.
func TestProjectDynamicRoutesToOpenAPICacheKeys(t *testing.T) {
	var (
		ctx        = context.Background()
		catalogSvc = &fakeProjectionCatalog{manifest: dynamicOpenAPITestManifest("OpenAPI Summary")}
		storeSvc   = &fakeProjectionStore{manifest: dynamicOpenAPITestManifest("OpenAPI Summary")}
		revision   = &fakeRevisionReader{revision: 1}
		locale     = &fakeLocaleBundleReader{locale: "zh-CN", version: 1}
		service    = New(catalogSvc, storeSvc, revision, locale)
	)

	firstPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, firstPaths); err != nil {
		t.Fatalf("project first dynamic routes: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 1 {
		t.Fatalf("expected first projection to scan once, got %d", got)
	}
	firstPaths["/x/plugin-openapi-cache/api/v1/review"].Get.Summary = "mutated by caller"

	secondPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, secondPaths); err != nil {
		t.Fatalf("project second dynamic routes: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 1 {
		t.Fatalf("expected cache hit not to rescan, got %d", got)
	}
	if summary := secondPaths["/x/plugin-openapi-cache/api/v1/review"].Get.Summary; summary != "OpenAPI Summary" {
		t.Fatalf("expected cached projection clone to keep original summary, got %q", summary)
	}

	revision.revision = 2
	thirdPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, thirdPaths); err != nil {
		t.Fatalf("project after revision change: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 2 {
		t.Fatalf("expected runtime revision change to rebuild projection, got scans=%d", got)
	}

	locale.locale = "en-US"
	fourthPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, fourthPaths); err != nil {
		t.Fatalf("project after locale change: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 3 {
		t.Fatalf("expected locale change to rebuild projection, got scans=%d", got)
	}

	locale.version = 2
	fifthPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, fifthPaths); err != nil {
		t.Fatalf("project after bundle version change: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 4 {
		t.Fatalf("expected bundle version change to rebuild projection, got scans=%d", got)
	}

	service.InvalidateProjectionCache(ctx, "test_runtime_changed")
	sixthPaths := goai.Paths{}
	if err := service.ProjectDynamicRoutesToOpenAPI(ctx, sixthPaths); err != nil {
		t.Fatalf("project after explicit invalidation: %v", err)
	}
	if got := catalogSvc.scanCalls; got != 5 {
		t.Fatalf("expected explicit invalidation to rebuild projection, got scans=%d", got)
	}
}

func dynamicOpenAPITestManifest(summary string) *catalog.Manifest {
	return &catalog.Manifest{
		ID:      "plugin-openapi-cache",
		Name:    "Plugin OpenAPI Cache",
		Version: "v0.1.0",
		Type:    plugintypes.TypeDynamic.String(),
		Routes: []*protocol.RouteContract{
			{
				Path:    "/api/v1/review",
				Method:  http.MethodGet,
				Access:  protocol.AccessLogin,
				Summary: summary,
			},
		},
		BridgeSpec: &protocol.BridgeSpec{RouteExecution: true},
	}
}

type fakeProjectionCatalog struct {
	manifest  *catalog.Manifest
	scanCalls int
}

func (f *fakeProjectionCatalog) ScanManifests() ([]*catalog.Manifest, error) {
	f.scanCalls++
	return []*catalog.Manifest{catalog.CloneManifest(f.manifest)}, nil
}

func (f *fakeProjectionCatalog) GetDesiredManifest(pluginID string) (*catalog.Manifest, error) {
	if pluginID != f.manifest.ID {
		return nil, nil
	}
	return catalog.CloneManifest(f.manifest), nil
}

type fakeProjectionStore struct {
	manifest *catalog.Manifest
}

func (f *fakeProjectionStore) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (f *fakeProjectionStore) ListAllRegistries(context.Context) ([]*store.PluginRecord, error) {
	return []*store.PluginRecord{
		{
			PluginId:  f.manifest.ID,
			Type:      plugintypes.TypeDynamic.String(),
			Installed: plugintypes.InstalledYes,
			Status:    plugintypes.StatusEnabled,
			ReleaseId: 1,
		},
	}, nil
}

func (f *fakeProjectionStore) GetRegistry(_ context.Context, pluginID string) (*store.PluginRecord, error) {
	if pluginID != f.manifest.ID {
		return nil, nil
	}
	return &store.PluginRecord{
		PluginId:  f.manifest.ID,
		Type:      plugintypes.TypeDynamic.String(),
		Installed: plugintypes.InstalledYes,
		Status:    plugintypes.StatusEnabled,
		ReleaseId: 1,
	}, nil
}

func (f *fakeProjectionStore) GetRegistryRelease(context.Context, *store.PluginRecord) (*store.ReleaseRecord, error) {
	return &store.ReleaseRecord{Id: 1, PluginId: f.manifest.ID, ReleaseVersion: f.manifest.Version}, nil
}

func (f *fakeProjectionStore) LoadReleaseManifest(context.Context, *store.ReleaseRecord) (*catalog.Manifest, error) {
	return catalog.CloneManifest(f.manifest), nil
}

type fakeRevisionReader struct {
	revision int64
}

func (f *fakeRevisionReader) CurrentRevision(context.Context) (int64, error) {
	return f.revision, nil
}

type fakeLocaleBundleReader struct {
	locale  string
	version uint64
}

func (f *fakeLocaleBundleReader) GetLocale(context.Context) string {
	return f.locale
}

func (f *fakeLocaleBundleReader) BundleVersion(string) uint64 {
	return f.version
}
