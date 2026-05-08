// This file tests bridge spec validation and request-response codec round trips.

package codec

import (
	"bytes"
	"net/http"
	"testing"
)

// TestValidateBridgeSpecRejectsTextCodec verifies executable bridge specs only
// accept protobuf request and response codecs.
func TestValidateBridgeSpecRejectsTextCodec(t *testing.T) {
	spec := &BridgeSpec{
		ABIVersion:     ABIVersionV1,
		RuntimeKind:    RuntimeKindWasm,
		RouteExecution: true,
		RequestCodec:   "json",
		ResponseCodec:  "protobuf",
	}
	if err := ValidateBridgeSpec(spec); err == nil {
		t.Fatal("expected text request codec to be rejected")
	}
}

// TestValidateRouteContractsRejectsInvalidPublicPermission verifies public
// routes cannot retain permission requirements.
func TestValidateRouteContractsRejectsInvalidPublicPermission(t *testing.T) {
	routes := []*RouteContract{
		{
			Path:       "/review-summary",
			Method:     http.MethodGet,
			Access:     AccessPublic,
			Permission: "plugin-demo-dynamic:review:view",
		},
	}
	if err := ValidateRouteContracts("plugin-demo-dynamic", routes); err == nil {
		t.Fatal("expected public route with permission to be rejected")
	}
}

// TestEncodeDecodeRequestEnvelopeRoundTrip verifies the manual protobuf codec
// preserves nested route, request, and identity snapshots.
func TestEncodeDecodeRequestEnvelopeRoundTrip(t *testing.T) {
	input := &BridgeRequestEnvelopeV1{
		PluginID: "plugin-demo-dynamic",
		Route: &RouteMatchSnapshotV1{
			Method:       http.MethodGet,
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/review-summary",
			InternalPath: "/review-summary",
			RoutePath:    "/review-summary",
			Access:       AccessLogin,
			Permission:   "plugin-demo-dynamic:review:view",
			RequestType:  "ReviewSummaryReq",
			PathParams: map[string]string{
				"id": "42",
			},
			QueryValues: map[string][]string{
				"q": {"hello"},
			},
		},
		Request: &HTTPRequestSnapshotV1{
			Method:       http.MethodGet,
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/review-summary",
			InternalPath: "/review-summary",
			RawPath:      "/api/v1/extensions/plugin-demo-dynamic/review-summary",
			RawQuery:     "q=hello",
			Host:         "localhost:8080",
			Scheme:       "http",
			ClientIP:     "127.0.0.1",
			Headers: map[string][]string{
				"Accept": {"application/json"},
			},
			Cookies: map[string]string{
				"lang": "zh-CN",
			},
			Body: []byte(`{"hello":"world"}`),
		},
		Identity: &IdentitySnapshotV1{
			TokenID:      "token-1",
			UserID:       1,
			Username:     "admin",
			Status:       1,
			Permissions:  []string{"plugin-demo-dynamic:review:view"},
			RoleNames:    []string{"超级管理员"},
			DataScope:    1,
			IsSuperAdmin: true,
		},
		RequestID: "req-1",
	}

	content, err := EncodeRequestEnvelope(input)
	if err != nil {
		t.Fatalf("expected request encode to succeed, got error: %v", err)
	}
	output, err := DecodeRequestEnvelope(content)
	if err != nil {
		t.Fatalf("expected request decode to succeed, got error: %v", err)
	}
	if output.PluginID != input.PluginID || output.RequestID != input.RequestID {
		t.Fatalf("unexpected request identity fields: %#v", output)
	}
	if output.Route == nil || output.Route.Permission != input.Route.Permission {
		t.Fatalf("unexpected route snapshot: %#v", output.Route)
	}
	if output.Request == nil || output.Request.RawQuery != input.Request.RawQuery {
		t.Fatalf("unexpected request snapshot: %#v", output.Request)
	}
	if output.Identity == nil || !output.Identity.IsSuperAdmin {
		t.Fatalf("unexpected identity snapshot: %#v", output.Identity)
	}
	if output.Identity.DataScope != input.Identity.DataScope {
		t.Fatalf("unexpected identity data scope: %#v", output.Identity)
	}
	if !bytes.Equal(output.Request.Body, input.Request.Body) {
		t.Fatalf("unexpected request body: %q", string(output.Request.Body))
	}
}
