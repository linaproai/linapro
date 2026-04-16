// This file covers OpenAPI helper projection logic for dynamic plugin routes.

package openapi

import (
	"net/http"
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestBuildRouteOpenAPIOperationUsesBridgeState(t *testing.T) {
	operation := BuildRouteOpenAPIOperation("plugin-demo-dynamic", &pluginbridge.RouteContract{
		Path:    "/review-summary",
		Method:  http.MethodGet,
		Access:  pluginbridge.AccessLogin,
		Summary: "Review Summary",
	}, &pluginbridge.BridgeSpec{
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

	placeholder := BuildRouteOpenAPIOperation("plugin-demo-dynamic", &pluginbridge.RouteContract{
		Path:   "/placeholder",
		Method: http.MethodGet,
		Access: pluginbridge.AccessPublic,
	}, &pluginbridge.BridgeSpec{
		RouteExecution: false,
	})
	if placeholder == nil || placeholder.Responses["501"].Value == nil {
		t.Fatalf("expected placeholder bridge operation to expose 501 response, got %#v", placeholder)
	}
	if placeholder.Responses["200"].Value != nil {
		t.Fatalf("expected placeholder bridge operation to omit 200 response, got %#v", placeholder)
	}
}

func TestBuildRoutePublicPathBuildsFixedPublicPath(t *testing.T) {
	actual := BuildRoutePublicPath("plugin-openapi-projection", "/review-summary")
	if actual != "/api/v1/extensions/plugin-openapi-projection/review-summary" {
		t.Fatalf("expected fixed public path projection, got %s", actual)
	}
}
