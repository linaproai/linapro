// This file covers exported dynamic-route dispatch helpers from outside the runtime package.

package runtime_test

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
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
			PluginID:   "plugin-demo-dynamic",
			PublicPath: "/api/v1/extensions/plugin-demo-dynamic/review",
			Route: &pluginbridge.RouteContract{
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
	if metadata.PluginID != "plugin-demo-dynamic" {
		t.Fatalf("expected plugin id plugin-demo-dynamic, got %q", metadata.PluginID)
	}
	if metadata.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %q", metadata.Method)
	}
	if metadata.PublicPath != "/api/v1/extensions/plugin-demo-dynamic/review" {
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

// TestExecuteDynamicWasmBridgeReturnsGuestResponse verifies that a bundled
// runtime plugin route executes and returns the guest response unchanged.
func TestExecuteDynamicWasmBridgeReturnsGuestResponse(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: "plugin-demo-dynamic",
		Route: &pluginbridge.RouteMatchSnapshotV1{
			InternalPath: "/backend-summary",
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/backend-summary",
			Access:       pluginbridge.AccessLogin,
			Permission:   "plugin-demo-dynamic:backend:view",
		},
		Identity: &pluginbridge.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &pluginbridge.HTTPRequestSnapshotV1{
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
	if got := response.Headers["X-Lina-Plugin-Bridge"]; len(got) != 1 || got[0] != "plugin-demo-dynamic" {
		t.Fatalf("expected guest bridge header to be preserved, got %#v", response.Headers)
	}
	if got := response.Headers["X-Lina-Plugin-Middleware"]; len(got) != 1 || got[0] != "backend-summary" {
		t.Fatalf("expected guest-local middleware header to be preserved, got %#v", response.Headers)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected guest response body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "plugin-demo-dynamic" {
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

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID:  "plugin-demo-dynamic",
		RequestID: "req-host-call-demo",
		Route: &pluginbridge.RouteMatchSnapshotV1{
			InternalPath: "/host-call-demo",
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/host-call-demo",
			Access:       pluginbridge.AccessLogin,
			Permission:   "plugin-demo-dynamic:backend:view",
			RequestType:  "HostCallDemoReq",
			QueryValues: map[string][]string{
				"skipNetwork": {"1"},
			},
		},
		Identity: &pluginbridge.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &pluginbridge.HTTPRequestSnapshotV1{
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
	if payload["pluginId"] != "plugin-demo-dynamic" {
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
	if dataPayload["table"] != "sys_plugin_node_state" {
		t.Fatalf("expected data table sys_plugin_node_state, got %#v", dataPayload)
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

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName("plugin-demo-dynamic"))
	return services.Catalog.LoadManifestFromArtifactPath(artifactPath)
}

// findCronContract locates one declared cron contract by stable plugin-local name.
func findCronContract(contracts []*pluginbridge.CronContract, name string) *pluginbridge.CronContract {
	for _, item := range contracts {
		if item != nil && item.Name == name {
			return item
		}
	}
	return nil
}
