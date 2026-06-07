// This file covers exported dynamic-route dispatch helpers from outside the runtime package.

package runtime_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
)

// TestMatchDynamicRoutePathSupportsParams verifies parameter placeholders are
// extracted from public route paths.
func TestMatchDynamicRoutePathSupportsParams(t *testing.T) {
	params, ok := runtime.MatchDynamicRoutePath("/records/{id}/detail", "/records/42/detail")
	if !ok {
		t.Fatal("expected dynamic path match to succeed")
	}
	if params["id"] != "42" {
		t.Fatalf("expected path param id=42, got %#v", params)
	}
}

// TestBuildDynamicRouteMetadataMapsRouteGovernance verifies that matched route
// metadata is projected into a generic dynamic-route context.
func TestBuildDynamicRouteMetadataMapsRouteGovernance(t *testing.T) {
	metadata := runtime.BuildDynamicRouteMetadata(&runtime.DynamicRouteRuntimeState{
		Match: &runtime.DynamicRouteMatch{
			PluginID:   "linapro-demo-dynamic",
			PublicPath: "/x/linapro-demo-dynamic/api/v1/review",
			Route: &protocol.RouteContract{
				Method:  http.MethodGet,
				Tags:    []string{"plugin-review", "dynamic"},
				Summary: "Review summary",
				Meta: map[string]string{
					"x-route-purpose": "review",
				},
			},
		},
	})
	if metadata == nil {
		t.Fatal("expected dynamic route metadata to be built")
	}
	if metadata.PluginID != "linapro-demo-dynamic" {
		t.Fatalf("expected plugin id linapro-demo-dynamic, got %q", metadata.PluginID)
	}
	if metadata.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %q", metadata.Method)
	}
	if metadata.PublicPath != "/x/linapro-demo-dynamic/api/v1/review" {
		t.Fatalf("expected public path to be preserved, got %q", metadata.PublicPath)
	}
	if len(metadata.Tags) != 2 || metadata.Tags[0] != "plugin-review" || metadata.Tags[1] != "dynamic" {
		t.Fatalf("expected route tags to be preserved, got %#v", metadata.Tags)
	}
	if metadata.Summary != "Review summary" {
		t.Fatalf("expected summary to be preserved, got %q", metadata.Summary)
	}
	if metadata.Meta["x-route-purpose"] != "review" {
		t.Fatalf("expected route metadata x-route-purpose review, got %#v", metadata.Meta)
	}
}

// TestDispatchDynamicRouteReturnsNotFoundWhenTenantPluginDisabled verifies
// tenant-scoped dynamic routes are hidden unless the current tenant enabled the
// plugin, even when the platform registry row is installed and enabled.
func TestDispatchDynamicRouteReturnsNotFoundWhenTenantPluginDisabled(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = datascope.WithTenantForTest(context.Background(), 7001)
		pluginID = "plugin-dev-dynamic-route-tenant-disabled"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Tenant Disabled Route Plugin",
		"v1.0.0",
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
			AllocExport:    "allocate",
			ExecuteExport:  "execute",
		},
	)
	testutil.CleanupPluginGovernanceRowsHard(t, context.Background(), pluginID)
	if _, err := dao.SysPluginState.Ctx(context.Background()).
		Where(do.SysPluginState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("cleanup dynamic route plugin state failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dao.SysPluginState.Ctx(context.Background()).
			Where(do.SysPluginState{PluginId: pluginID}).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route plugin state failed: %v", err)
		}
		testutil.CleanupPluginGovernanceRowsHard(t, context.Background(), pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("load dynamic route manifest failed: %v", err)
	}
	manifest.ScopeNature = catalog.ScopeNatureTenantAware.String()
	manifest.DefaultInstallMode = catalog.InstallModeTenantScoped.String()
	if _, err = services.Catalog.SyncManifest(context.Background(), manifest); err != nil {
		t.Fatalf("sync dynamic route manifest failed: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(context.Background(), pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("set dynamic route plugin installed failed: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(context.Background(), pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("set dynamic route plugin enabled failed: %v", err)
	}
	if _, err = dao.SysPlugin.Ctx(context.Background()).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			ScopeNature: catalog.ScopeNatureTenantAware.String(),
			InstallMode: catalog.InstallModeTenantScoped.String(),
		}).
		Update(); err != nil {
		t.Fatalf("set dynamic route plugin tenant governance failed: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("dispatch disabled tenant plugin route failed: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected disabled tenant plugin route to return 404, got %#v", response)
	}

	if err = services.Integration.SetTenantPluginEnabledState(ctx, pluginID, datascope.CurrentTenantID(ctx), true); err != nil {
		t.Fatalf("enable plugin for tenant failed: %v", err)
	}
	response, err = services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err == nil && response != nil && response.StatusCode == http.StatusNotFound {
		t.Fatalf("expected enabled tenant plugin route to pass routing, got %#v", response)
	}
	if err != nil && strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected enabled tenant plugin route to pass routing, got error: %v", err)
	}
}

// TestDispatchDynamicRouteReturnsUpgradeRequiredWhenPendingUpgrade verifies
// dynamic business routes are blocked while a newer artifact awaits runtime upgrade.
func TestDispatchDynamicRouteReturnsUpgradeRequiredWhenPendingUpgrade(t *testing.T) {
	var (
		services   = testutil.NewServices()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-route-pending-upgrade"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Pending Upgrade Plugin",
		oldVersion,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic route manifest to load, got error: %v", err)
	}
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic route manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic route plugin install state to be set, got error: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic route plugin enable state to be set, got error: %v", err)
	}

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Pending Upgrade Plugin",
		newVersion,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)
	newManifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic route manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic route manifest sync to succeed, got error: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("expected pending-upgrade dynamic route to return bridge failure response, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusConflict {
		t.Fatalf("expected pending-upgrade dynamic route to return 409, got %#v", response)
	}
	if response.Failure == nil || response.Failure.Code != runtime.CodePluginRuntimeUpgradeRequired.RuntimeCode() {
		t.Fatalf("expected stable upgrade-required failure code, got %#v", response)
	}
}

// TestDispatchDynamicRouteBlocksFailedActiveRelease verifies rollback failure
// state uses the conservative exposure policy and does not serve a failed
// dynamic release even if the archived route manifest is still loadable.
func TestDispatchDynamicRouteBlocksFailedActiveRelease(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-route-failed-active-release"
		version  = "v1.0.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Failed Active Release Route Plugin",
		version,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath); err != nil {
		t.Fatalf("expected dynamic route manifest to load, got error: %v", err)
	}
	if err := services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected dynamic route plugin install to succeed, got error: %v", err)
	}
	if err := services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic route plugin enable state to be set, got error: %v", err)
	}
	registry, err := services.Catalog.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.ReleaseId <= 0 {
		t.Fatalf("expected installed plugin to point at active release, got %#v", registry)
	}
	if err = services.Catalog.UpdateReleaseState(ctx, registry.ReleaseId, catalog.ReleaseStatusFailed, ""); err != nil {
		t.Fatalf("expected release failure marker update to succeed, got error: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("expected failed-release dynamic route to return bridge failure response, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusConflict {
		t.Fatalf("expected failed-release dynamic route to return 409, got %#v", response)
	}
	if response.Failure == nil || response.Failure.Code != runtime.CodePluginRuntimeUpgradeRequired.RuntimeCode() {
		t.Fatalf("expected stable upgrade-required failure code, got %#v", response)
	}
}

// TestDispatchDynamicRouteAllowsPluginOwnedPathShapes verifies the runtime only
// forces the `/x/{pluginId}` prefix and preserves the following plugin-owned
// path content for contract matching and bridge metadata.
func TestDispatchDynamicRouteAllowsPluginOwnedPathShapes(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-route-owned-paths"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Owned Paths Plugin",
		"v1.0.0",
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v2/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryV2Req",
			},
			{
				Path:        "/interface/m1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "InterfaceSummaryReq",
			},
			{
				Path:        "/graphql",
				Method:      http.MethodPost,
				Access:      protocol.AccessPublic,
				RequestType: "GraphQLReq",
			},
			{
				Path:        "/",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "RootReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: false,
		},
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("load dynamic route manifest failed: %v", err)
	}
	manifest.ScopeNature = catalog.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = catalog.InstallModeGlobal.String()
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("sync dynamic route manifest failed: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("set dynamic route plugin installed failed: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("set dynamic route plugin enabled failed: %v", err)
	}
	services.Integration.SetPluginEnabledState(pluginID, true)

	tests := []struct {
		name       string
		method     string
		publicPath string
	}{
		{
			name:       "api v2",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID + "/api/v2/summary",
		},
		{
			name:       "interface",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID + "/interface/m1/summary",
		},
		{
			name:       "graphql",
			method:     http.MethodPost,
			publicPath: "/x/" + pluginID + "/graphql",
		},
		{
			name:       "root",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			request := &ghttp.Request{}
			request.Request = httptest.NewRequest(testCase.method, testCase.publicPath, nil)
			response, err := services.Runtime.DispatchDynamicRoute(
				ctx,
				&runtime.DynamicRouteDispatchInput{Request: request},
			)
			if err != nil {
				t.Fatalf("expected dynamic route dispatch to return bridge response, got error: %v", err)
			}
			if response == nil || response.StatusCode != http.StatusNotImplemented {
				t.Fatalf("expected matched placeholder route to return 501, got %#v", response)
			}
		})
	}
}

// TestExecuteDynamicWasmBridgeReturnsGuestResponse verifies that a bundled
// runtime plugin route executes and returns the guest response unchanged.
func TestExecuteDynamicWasmBridgeReturnsGuestResponse(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID: "linapro-demo-dynamic",
		Route: &protocol.RouteMatchSnapshotV1{
			InternalPath: "/api/v1/backend-summary",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/backend-summary",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:backend:view",
			RequestType:  "BackendSummaryReq",
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected dynamic wasm execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected guest bridge response 200, got %#v", response)
	}
	if string(response.Body) == "" {
		t.Fatal("expected guest bridge response body to be non-empty")
	}
	if got := response.Headers["X-Lina-Plugin-Bridge"]; len(got) != 1 || got[0] != "linapro-demo-dynamic" {
		t.Fatalf("expected guest bridge header to be preserved, got %#v", response.Headers)
	}
	if got := response.Headers["X-Lina-Plugin-Middleware"]; len(got) != 1 || got[0] != "backend-summary" {
		t.Fatalf("expected guest-local middleware header to be preserved, got %#v", response.Headers)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected guest response body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "linapro-demo-dynamic" {
		t.Fatalf("expected guest payload pluginId to be preserved, got %#v", payload)
	}
	if payload["authenticated"] != true {
		t.Fatalf("expected guest payload authenticated=true, got %#v", payload)
	}
}

// TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices verifies
// that structured host-service declarations are available inside guest code.
func TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)
	ensureDynamicDemoRecordTable(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID:  "linapro-demo-dynamic",
		RequestID: "req-host-call-demo",
		Route: &protocol.RouteMatchSnapshotV1{
			InternalPath: "/api/v1/host-call-demo",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/host-call-demo",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:backend:view",
			RequestType:  "HostCallDemoReq",
			QueryValues: map[string][]string{
				"skipNetwork": {"1"},
			},
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected host call demo execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected host call demo response 200, got %#v", response)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected host call demo body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "linapro-demo-dynamic" {
		t.Fatalf("expected pluginId to be preserved, got %#v", payload)
	}
	if payload["visitCount"] == nil {
		t.Fatalf("expected visitCount to be returned, got %#v", payload)
	}

	runtimePayload, ok := payload["runtime"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected runtime payload object, got %#v full payload=%#v body=%s", payload["runtime"], payload, string(response.Body))
	}
	if runtimePayload["uuid"] == "" || runtimePayload["node"] == "" {
		t.Fatalf("expected runtime payload to include uuid and node, got %#v", runtimePayload)
	}

	storagePayload, ok := payload["storage"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected storage payload object, got %#v", payload["storage"])
	}
	if storagePayload["pathPrefix"] != "host-call-demo/" {
		t.Fatalf("expected storage pathPrefix host-call-demo/, got %#v", storagePayload)
	}
	if storagePayload["stored"] != true || storagePayload["deleted"] != true {
		t.Fatalf("expected storage payload to confirm store/delete lifecycle, got %#v", storagePayload)
	}

	dataPayload, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data payload object, got %#v", payload["data"])
	}
	if dataPayload["table"] != "plugin_linapro_demo_dynamic_record" {
		t.Fatalf("expected data table plugin_linapro_demo_dynamic_record, got %#v", dataPayload)
	}
	if dataPayload["updated"] != true || dataPayload["deleted"] != true {
		t.Fatalf("expected data payload to confirm update/delete lifecycle, got %#v", dataPayload)
	}

	networkPayload, ok := payload["network"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected network payload object, got %#v", payload["network"])
	}
	if networkPayload["url"] != "https://example.com" {
		t.Fatalf("expected network url https://example.com, got %#v", networkPayload)
	}
	if networkPayload["skipped"] != true {
		t.Fatalf("expected network payload skipped=true during offline-safe test run, got %#v", networkPayload)
	}

	orgPayload, ok := payload["org"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected org payload object, got %#v", payload["org"])
	}
	if _, ok = orgPayload["available"].(bool); !ok {
		t.Fatalf("expected org payload to include availability, got %#v", orgPayload)
	}
	if orgPayload["assignmentCount"] == nil || orgPayload["currentUserDeptCount"] == nil {
		t.Fatalf("expected org payload to include current-user organization projections, got %#v", orgPayload)
	}

	tenantPayload, ok := payload["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tenant payload object, got %#v", payload["tenant"])
	}
	if tenantPayload["visible"] != true {
		t.Fatalf("expected tenant visibility check to pass, got %#v", tenantPayload)
	}
	if tenantPayload["currentTenantId"] == nil || tenantPayload["userTenantCount"] == nil {
		t.Fatalf("expected tenant payload to include current tenant and user tenant count, got %#v", tenantPayload)
	}
}

// TestExecuteDynamicWasmBridgeCreatesDemoRecord verifies the bundled dynamic
// sample can execute the create demo-record route with data and storage host
// services, matching the E2E CRUD path.
func TestExecuteDynamicWasmBridgeCreatesDemoRecord(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)
	ensureDynamicDemoRecordTable(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID:  "linapro-demo-dynamic",
		RequestID: "req-demo-record-create",
		Route: &protocol.RouteMatchSnapshotV1{
			Method:       http.MethodPost,
			InternalPath: "/api/v1/demo-records",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/demo-records",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:record:create",
			RequestType:  "CreateDemoRecordReq",
			RoutePath:    "/api/v1/demo-records",
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method:      http.MethodPost,
			ContentType: "application/json",
			Body: []byte(`{
				"title":"Dynamic route create test",
				"content":"Created through the bundled dynamic WASM bridge",
				"attachmentName":"linapro-demo-dynamic-note.txt",
				"attachmentContentBase64":"bGluYXByby1kZW1vLWR5bmFtaWMgYXR0YWNobWVudCBmaXh0dXJl",
				"attachmentContentType":"text/plain"
			}`),
		},
	})
	if err != nil {
		t.Fatalf("expected demo-record create route to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected demo-record create response 200, got %#v body=%s", response, responseBodyForTest(response))
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected demo-record create body to be valid json, got error: %v body=%s", err, string(response.Body))
	}
	if payload["title"] != "Dynamic route create test" || payload["hasAttachment"] != true {
		t.Fatalf("expected created demo-record payload with attachment, got %#v", payload)
	}
}

// TestExecuteDeclaredCronJobUsesWasmBridge verifies that dynamic-plugin cron
// discovery and execution both reuse the shared Wasm bridge.
func TestExecuteDeclaredCronJobUsesWasmBridge(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	contracts, err := services.Runtime.DiscoverCronContracts(context.Background(), manifest)
	if err != nil {
		t.Fatalf("expected bundled runtime cron discovery to succeed, got error: %v", err)
	}
	contract := findCronContract(contracts, "heartbeat")
	if contract == nil {
		t.Fatalf("expected bundled runtime manifest to discover heartbeat cron contract, got %#v", contracts)
	}

	ctx := context.Background()
	if _, err = dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: manifest.ID,
			StateKey: "cron_heartbeat_count",
		}).
		Delete(); err != nil {
		t.Fatalf("expected cron state cleanup to succeed, got error: %v", err)
	}

	if err = services.Runtime.ExecuteDeclaredCronJob(ctx, manifest, contract); err != nil {
		t.Fatalf("expected declared cron execution to succeed, got error: %v", err)
	}

	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: manifest.ID,
			StateKey: "cron_heartbeat_count",
		}).
		Value("state_value")
	if err != nil {
		t.Fatalf("expected cron state query to succeed, got error: %v", err)
	}
	if value.IsEmpty() || value.String() != "1" {
		t.Fatalf("expected cron heartbeat state value to be 1, got %#v", value)
	}
}

// loadBundledDynamicSampleManifest loads the bundled demo runtime artifact from test storage.
func loadBundledDynamicSampleManifest(t *testing.T, services *testutil.Services) (*catalog.Manifest, error) {
	t.Helper()

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName("linapro-demo-dynamic"))
	return services.Catalog.LoadManifestFromArtifactPath(artifactPath)
}

// ensureDynamicDemoRecordTable provisions the dynamic sample table needed by
// bundled route tests that bypass the full plugin install lifecycle.
func ensureDynamicDemoRecordTable(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := g.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS plugin_linapro_demo_dynamic_record (
    "id"              VARCHAR(64) PRIMARY KEY,
    "tenant_id"       INT NOT NULL DEFAULT 0,
    "title"           VARCHAR(128) NOT NULL DEFAULT '',
    "content"         VARCHAR(1000) NOT NULL DEFAULT '',
    "attachment_name" VARCHAR(255) NOT NULL DEFAULT '',
    "attachment_path" VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		t.Fatalf("failed to create dynamic demo record table: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := g.DB().Exec(ctx, `DELETE FROM plugin_linapro_demo_dynamic_record WHERE "title" = ?`, "Dynamic route create test"); cleanupErr != nil {
			t.Fatalf("failed to cleanup dynamic demo record table: %v", cleanupErr)
		}
	})
}

// responseBodyForTest returns response body bytes without forcing every failure
// assertion to nil-check the response first.
func responseBodyForTest(response *protocol.BridgeResponseEnvelopeV1) []byte {
	if response == nil {
		return nil
	}
	return response.Body
}

// findCronContract locates one declared cron contract by stable plugin-local name.
func findCronContract(contracts []*protocol.CronContract, name string) *protocol.CronContract {
	for _, item := range contracts {
		if item != nil && item.Name == name {
			return item
		}
	}
	return nil
}
