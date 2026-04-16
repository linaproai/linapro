// This file tests network host service URL authorization, wildcard matching,
// protected-header filtering, timeout handling, and bounded response bodies.

package wasm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lina-core/pkg/pluginbridge"
)

func TestHandleHostServiceInvokeNetworkRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method: got %s want %s", request.Method, http.MethodPost)
		}
		if request.URL.Path != "/api/v1/ping" {
			t.Fatalf("path: got %s want /api/v1/ping", request.URL.Path)
		}
		if request.URL.Query().Get("tenant") != "demo" {
			t.Fatalf("query: got %s want demo", request.URL.Query().Get("tenant"))
		}
		if request.Header.Get("X-Request-Id") != "req-1" {
			t.Fatalf("x-request-id: got %s", request.Header.Get("X-Request-Id"))
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.Header().Set("X-Upstream", "crm")
		if _, err := writer.Write([]byte(`{"ok":true}`)); err != nil {
			t.Fatalf("write success response failed: %v", err)
		}
	}))
	defer server.Close()

	hcc := newNetworkHostCallContext(server.URL + "/api/v1")

	response := invokeNetworkHostService(
		t,
		context.Background(),
		hcc,
		server.URL+"/api/v1/ping?tenant=demo",
		&pluginbridge.HostServiceNetworkRequest{
			Method: http.MethodPost,
			Headers: map[string]string{
				"x-request-id": "req-1",
			},
			Body: []byte(`{"name":"ticket"}`),
		},
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}

	payload, err := pluginbridge.UnmarshalHostServiceNetworkResponse(response.Payload)
	if err != nil {
		t.Fatalf("payload decode failed: %v", err)
	}
	if payload.StatusCode != http.StatusOK {
		t.Fatalf("statusCode: got %d want %d", payload.StatusCode, http.StatusOK)
	}
	if payload.ContentType != "application/json" {
		t.Fatalf("contentType: got %s want application/json", payload.ContentType)
	}
	if payload.Headers["X-Upstream"] != "crm" {
		t.Fatalf("headers: got %#v", payload.Headers)
	}
	if string(payload.Body) != `{"ok":true}` {
		t.Fatalf("body: got %q", payload.Body)
	}
}

func TestHandleHostServiceInvokeNetworkRejectsUnauthorizedURL(t *testing.T) {
	hcc := newNetworkHostCallContext("https://api.example.com/v1")

	response := invokeNetworkHostService(
		t,
		context.Background(),
		hcc,
		"https://evil.example.com/v1/ping",
		&pluginbridge.HostServiceNetworkRequest{Method: http.MethodGet},
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestHandleHostServiceInvokeNetworkRejectsProtectedHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("ok")); err != nil {
			t.Fatalf("write protected-header response failed: %v", err)
		}
	}))
	defer server.Close()

	hcc := newNetworkHostCallContext(server.URL)

	response := invokeNetworkHostService(
		t,
		context.Background(),
		hcc,
		server.URL+"/ping",
		&pluginbridge.HostServiceNetworkRequest{
			Method: http.MethodGet,
			Headers: map[string]string{
				"Host": "evil.example.com",
			},
		},
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestHandleHostServiceInvokeNetworkRejectsOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write(make([]byte, defaultNetworkMaxBodyBytes+1)); err != nil {
			t.Fatalf("write oversized response failed: %v", err)
		}
	}))
	defer server.Close()

	hcc := newNetworkHostCallContext(server.URL)

	response := invokeNetworkHostService(
		t,
		context.Background(),
		hcc,
		server.URL+"/ping",
		&pluginbridge.HostServiceNetworkRequest{Method: http.MethodGet},
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestHandleHostServiceInvokeNetworkTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(50 * time.Millisecond)
		if _, err := writer.Write([]byte("slow")); err != nil {
			t.Fatalf("write slow response failed: %v", err)
		}
	}))
	defer server.Close()

	hcc := newNetworkHostCallContext(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	response := invokeNetworkHostService(
		t,
		ctx,
		hcc,
		server.URL+"/ping",
		&pluginbridge.HostServiceNetworkRequest{Method: http.MethodGet},
	)
	if response.Status != pluginbridge.HostCallStatusInternalError {
		t.Fatalf("expected internal error, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestMatchAuthorizedNetworkResourceSupportsWildcardHost(t *testing.T) {
	specs := []*pluginbridge.HostServiceSpec{
		{
			Service: pluginbridge.HostServiceNetwork,
			Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
			Resources: []*pluginbridge.HostServiceResourceSpec{
				{Ref: "https://*.example.com/api"},
			},
		},
	}

	resource := matchAuthorizedNetworkResource(specs, "https://foo.bar.example.com/api/orders?id=1")
	if resource == nil || resource.Ref != "https://*.example.com/api" {
		t.Fatalf("expected wildcard network resource match, got %#v", resource)
	}
}

func newNetworkHostCallContext(pattern string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-network",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityHTTPRequest: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{
			{
				Service: pluginbridge.HostServiceNetwork,
				Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
				Resources: []*pluginbridge.HostServiceResourceSpec{
					{Ref: pattern},
				},
			},
		},
	}
}

func invokeNetworkHostService(
	t *testing.T,
	ctx context.Context,
	hcc *hostCallContext,
	targetURL string,
	request *pluginbridge.HostServiceNetworkRequest,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	envelope := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceNetwork,
		Method:      pluginbridge.HostServiceMethodNetworkRequest,
		ResourceRef: targetURL,
		Payload:     pluginbridge.MarshalHostServiceNetworkRequest(request),
	}
	return handleHostServiceInvoke(
		ctx,
		hcc,
		pluginbridge.MarshalHostServiceRequestEnvelope(envelope),
	)
}
