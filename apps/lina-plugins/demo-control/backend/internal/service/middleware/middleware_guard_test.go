// This file verifies demo-control middleware allow and reject behavior.

package middleware

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// demoControlTestResponse stores one HTTP response snapshot used by the tests.
type demoControlTestResponse struct {
	status int
	body   string
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

// TestGuardRejectsWriteRequestsWhenDemoControlEnabled verifies an enabled
// plugin blocks write methods outside the session whitelist.
func TestGuardRejectsWriteRequestsWhenPluginEnabled(t *testing.T) {
	baseURL, shutdown := startDemoControlTestServer(t, true)
	defer shutdown()

	response := doDemoControlRequest(t, http.MethodPut, baseURL+"/api/v1/resource")
	if response.status != http.StatusForbidden {
		t.Fatalf("expected PUT to be rejected, got %d", response.status)
	}
	if !strings.Contains(response.body, demoControlMessage) {
		t.Fatalf("expected rejection body to mention demo-control message, got %q", response.body)
	}
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

// startDemoControlTestServer boots one ephemeral HTTP server with the
// demo-control middleware mounted on `/api/v1/*`.
func startDemoControlTestServer(t *testing.T, enabled bool) (string, func()) {
	t.Helper()

	server := g.Server(fmt.Sprintf("demo-control-middleware-test-%d", time.Now().UnixNano()))
	server.SetDumpRouterMap(false)
	server.SetPort(0)

	if enabled {
		guardSvc := New()
		server.BindMiddleware("/api/v1/*", guardSvc.Guard)
	}
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
