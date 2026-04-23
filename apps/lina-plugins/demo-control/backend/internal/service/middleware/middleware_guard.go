// This file implements request classification and rejection logic for the
// demo-control middleware.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// demo-control middleware constants define safe methods, whitelist routes, and
// the normalized JSON error envelope returned for blocked write operations.
const (
	demoControlErrorCode = 1
	demoControlMessage   = "演示模式已开启，禁止执行写操作"
)

// demoControlErrorResponse defines the JSON payload returned for blocked demo writes.
type demoControlErrorResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// Guard enforces the demo-mode read-only policy on API requests.
func (s *serviceImpl) Guard(request *ghttp.Request) {
	if request == nil {
		return
	}
	if isDemoControlAllowedRequest(request) {
		request.Middleware.Next()
		return
	}
	writeDemoControlError(request)
}

// isDemoControlAllowedRequest reports whether the incoming request should bypass
// demo-mode write protection.
func isDemoControlAllowedRequest(request *ghttp.Request) bool {
	if request == nil {
		return true
	}

	method := normalizeDemoControlMethod(request.Method)
	if isDemoControlSafeMethod(method) {
		return true
	}
	return isDemoControlSessionWhitelist(method, request.URL.Path)
}

// normalizeDemoControlMethod trims and uppercases one request method.
func normalizeDemoControlMethod(method string) string {
	return strings.ToUpper(strings.TrimSpace(method))
}

// isDemoControlSafeMethod reports whether the method is read-only under the
// demo-mode guard contract.
func isDemoControlSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// isDemoControlSessionWhitelist reports whether the request belongs to the
// minimal session-management whitelist required by demo environments.
func isDemoControlSessionWhitelist(method string, path string) bool {
	if method != http.MethodPost {
		return false
	}

	switch normalizeDemoControlPath(path) {
	case "/api/v1/auth/login", "/api/v1/auth/logout":
		return true
	default:
		return false
	}
}

// normalizeDemoControlPath canonicalizes one request path for whitelist matching.
func normalizeDemoControlPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if len(trimmed) > 1 {
		trimmed = strings.TrimRight(trimmed, "/")
		if trimmed == "" {
			return "/"
		}
	}
	return trimmed
}

// writeDemoControlError writes one JSON error response for blocked write requests.
func writeDemoControlError(request *ghttp.Request) {
	err := gerror.New(demoControlMessage)
	request.SetError(err)
	request.Response.WriteStatus(http.StatusForbidden)
	request.Response.WriteJson(demoControlErrorResponse{
		Code:    demoControlErrorCode,
		Data:    nil,
		Message: demoControlMessage,
	})
}
