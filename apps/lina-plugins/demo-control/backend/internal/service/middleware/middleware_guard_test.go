// This file verifies demo-control middleware allow and reject behavior.

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
)

// demoControlTestResponse stores one HTTP response snapshot used by the tests.
type demoControlTestResponse struct {
	status int
	body   string
}

// staticDemoControlEnablementReader returns one fixed enablement value for tests.
type staticDemoControlEnablementReader bool

// IsEnabled reports the fixed test enablement value.
func (r staticDemoControlEnablementReader) IsEnabled(_ context.Context, _ string) bool {
	return bool(r)
}

// TestGuardBypassesWriteRequestsWhenPluginDisabled verifies an unenabled
// plugin does not interfere with downstream write handlers.
func TestGuardBypassesWriteRequestsWhenPluginDisabled(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, false)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodPost, baseURL+"/api/v1/resource")
	if response.status != http.StatusOK {
		t.Fatalf("expected disabled plugin to keep POST allowed, got %d", response.status)
	}
	if response.body != "mutated" {
		t.Fatalf("expected downstream POST handler body, got %q", response.body)
	}
}

// TestGuardRejectsWriteRequestsWhenPluginEnabledWithoutAutoEnable verifies
// runtime interception no longer depends on plugin.autoEnable once the plugin
// has already been enabled.
func TestGuardRejectsWriteRequestsWhenPluginEnabledWithoutAutoEnable(t *testing.T) {
	setDemoControlTestConfig(t, `
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
plugin:
  dynamic:
    storagePath: "temp/output"
`)

	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodPut, baseURL+"/api/v1/resource")
	if response.status != http.StatusForbidden {
		t.Fatalf("expected PUT to be rejected after manual enable without autoEnable, got %d", response.status)
	}
	assertDemoControlRejectedResponse(t, response, http.MethodPut, "/api/v1/resource")
}

// TestGuardAllowsSafeReadMethods verifies an enabled plugin still allows read-only methods.
func TestGuardAllowsSafeReadMethods(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodGet, baseURL+"/api/v1/ping")
	if response.status != http.StatusOK {
		t.Fatalf("expected GET to stay allowed, got %d", response.status)
	}
	if response.body != "ok" {
		t.Fatalf("expected downstream GET handler body, got %q", response.body)
	}
}

// TestGuardRejectsWriteRequestsWhenPluginEnabled verifies an enabled plugin
// blocks write methods outside the session whitelist.
func TestGuardRejectsWriteRequestsWhenPluginEnabled(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodPut, baseURL+"/api/v1/resource")
	if response.status != http.StatusForbidden {
		t.Fatalf("expected PUT to be rejected, got %d", response.status)
	}
	assertDemoControlRejectedResponse(t, response, http.MethodPut, "/api/v1/resource")
}

// TestGuardCoversWholeSystemScope verifies the plugin guards non-API write
// routes as well when it is mounted on the whole system scope.
func TestGuardCoversWholeSystemScope(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodPost, baseURL+"/system/write")
	if response.status != http.StatusForbidden {
		t.Fatalf("expected whole-system POST to be rejected, got %d", response.status)
	}
	assertDemoControlRejectedResponse(t, response, http.MethodPost, "/system/write")
}

// TestGuardAllowsLoginAndLogoutWhitelist verifies the plugin preserves the
// minimal session whitelist needed for usable demos.
func TestGuardAllowsLoginAndLogoutWhitelist(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	loginResponse := doDemoControlRequest(t, http.MethodPost, baseURL+"/api/v1/auth/login")
	if loginResponse.status != http.StatusOK || loginResponse.body != "login-ok" {
		t.Fatalf("expected login whitelist to pass, got status=%d body=%q", loginResponse.status, loginResponse.body)
	}

	logoutResponse := doDemoControlRequest(t, http.MethodPost, baseURL+"/api/v1/auth/logout")
	if logoutResponse.status != http.StatusOK || logoutResponse.body != "logout-ok" {
		t.Fatalf("expected logout whitelist to pass, got status=%d body=%q", logoutResponse.status, logoutResponse.body)
	}
}

// TestGuardRejectsPluginManagementWriteRequests verifies demo mode blocks all
// plugin-governance write operations after the guard is enabled.
func TestGuardRejectsPluginManagementWriteRequests(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	rejectedRequests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/plugins/plugin-demo-source/install"},
		{method: http.MethodPut, path: "/api/v1/plugins/plugin-demo-source/enable"},
		{method: http.MethodPut, path: "/api/v1/plugins/plugin-demo-source/disable"},
		{method: http.MethodDelete, path: "/api/v1/plugins/plugin-demo-source"},
		{method: http.MethodPost, path: "/api/v1/plugins/demo-control/install"},
		{method: http.MethodPut, path: "/api/v1/plugins/demo-control/enable"},
		{method: http.MethodPut, path: "/api/v1/plugins/demo-control/disable"},
		{method: http.MethodDelete, path: "/api/v1/plugins/demo-control"},
		{method: http.MethodPost, path: "/api/v1/plugins/sync"},
		{method: http.MethodPost, path: "/api/v1/plugins/dynamic/package"},
	}

	for _, item := range rejectedRequests {
		response := doDemoControlRequest(t, item.method, baseURL+item.path)
		if response.status != http.StatusForbidden {
			t.Fatalf("expected %s %s to be rejected, got %d", item.method, item.path, response.status)
		}
		assertDemoControlRejectedResponse(t, response, item.method, item.path)
	}
}

// TestGuardAllowsPluginManagementReadRequests verifies plugin management reads
// still pass through because they use safe HTTP methods.
func TestGuardAllowsPluginManagementReadRequests(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	allowedRequests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/plugins/plugin-demo-source"},
		{method: http.MethodHead, path: "/api/v1/plugins/plugin-demo-source"},
		{method: http.MethodOptions, path: "/api/v1/plugins/plugin-demo-source"},
	}

	for _, item := range allowedRequests {
		response := doDemoControlRequest(t, item.method, baseURL+item.path)
		if response.status != http.StatusOK {
			t.Fatalf("expected %s %s to stay allowed, got %d", item.method, item.path, response.status)
		}
	}
}

// startDemoControlTestServer boots one ephemeral HTTP server with the
// demo-control middleware mounted on `/*`.
func startDemoControlTestServer(t *testing.T, enabled bool) (string, func()) {
	t.Helper()

	server := g.Server(fmt.Sprintf("demo-control-middleware-test-%d", time.Now().UnixNano()))
	server.SetDumpRouterMap(false)
	server.SetPort(0)

	guardSvc := New(nil, staticDemoControlEnablementReader(enabled))
	server.BindMiddleware("/*", guardSvc.Guard)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.ALL("/ping", func(request *ghttp.Request) {
			request.Response.Write("ok")
		})
		group.ALL("/resource", func(request *ghttp.Request) {
			request.Response.Write("mutated")
		})
		group.ALL("/auth/login", func(request *ghttp.Request) {
			request.Response.Write("login-ok")
		})
		group.ALL("/auth/logout", func(request *ghttp.Request) {
			request.Response.Write("logout-ok")
		})
		group.ALL("/plugins/sync", func(request *ghttp.Request) {
			request.Response.Write("plugin-sync-ok")
		})
		group.ALL("/plugins/dynamic/package", func(request *ghttp.Request) {
			request.Response.Write("plugin-dynamic-upload-ok")
		})
		group.ALL("/plugins/plugin-demo-source/install", func(request *ghttp.Request) {
			request.Response.Write("plugin-install-ok")
		})
		group.ALL("/plugins/plugin-demo-source/enable", func(request *ghttp.Request) {
			request.Response.Write("plugin-enable-ok")
		})
		group.ALL("/plugins/plugin-demo-source/disable", func(request *ghttp.Request) {
			request.Response.Write("plugin-disable-ok")
		})
		group.ALL("/plugins/plugin-demo-source", func(request *ghttp.Request) {
			request.Response.Write("plugin-uninstall-ok")
		})
		group.ALL("/plugins/demo-control/install", func(request *ghttp.Request) {
			request.Response.Write("demo-control-install-ok")
		})
		group.ALL("/plugins/demo-control/enable", func(request *ghttp.Request) {
			request.Response.Write("demo-control-enable-ok")
		})
		group.ALL("/plugins/demo-control/disable", func(request *ghttp.Request) {
			request.Response.Write("demo-control-disable-ok")
		})
		group.ALL("/plugins/demo-control", func(request *ghttp.Request) {
			request.Response.Write("demo-control-uninstall-ok")
		})
	})
	server.BindHandler("/system/write", func(request *ghttp.Request) {
		request.Response.Write("system-write-ok")
	})

	server.Start()
	time.Sleep(100 * time.Millisecond)

	return fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()), func() {
		server.Shutdown()
	}
}

// doDemoControlRequest sends one HTTP request and captures the response snapshot for assertions.
func doDemoControlRequest(t *testing.T, method string, targetURL string) demoControlTestResponse {
	t.Helper()

	request, err := http.NewRequest(method, targetURL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return demoControlTestResponse{status: response.StatusCode, body: string(body)}
}

// assertDemoControlRejectedResponse verifies one blocked demo-control response
// stays as valid JSON so the frontend can surface the explicit demo-mode
// message instead of falling back to a generic 403 string.
func assertDemoControlRejectedResponse(
	t *testing.T,
	response demoControlTestResponse,
	method string,
	path string,
) {
	t.Helper()

	var payload demoControlErrorResponse
	if err := json.Unmarshal([]byte(response.body), &payload); err != nil {
		t.Fatalf("expected valid JSON rejection body for %s %s, got %q (err=%v)", method, path, response.body, err)
	}
	if payload.Code != CodeDemoControlWriteDenied.TypeCode().Code() {
		t.Fatalf("expected rejection code %d for %s %s, got %d", CodeDemoControlWriteDenied.TypeCode().Code(), method, path, payload.Code)
	}
	if payload.Message == "" {
		t.Fatalf("expected non-empty rejection message for %s %s", method, path)
	}
	if payload.ErrorCode != CodeDemoControlWriteDenied.RuntimeCode() {
		t.Fatalf("expected rejection error code %q for %s %s, got %q", CodeDemoControlWriteDenied.RuntimeCode(), method, path, payload.ErrorCode)
	}
	if payload.MessageKey != CodeDemoControlWriteDenied.MessageKey() {
		t.Fatalf("expected rejection message key %q for %s %s, got %q", CodeDemoControlWriteDenied.MessageKey(), method, path, payload.MessageKey)
	}
}

// setDemoControlTestConfig replaces the process config adapter for one test
// case so the middleware can prove runtime interception no longer depends on
// plugin.autoEnable.
func setDemoControlTestConfig(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)
	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}
