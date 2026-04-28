// This file implements request classification and rejection logic for the
// demo-control middleware.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
)

// demo-control middleware constants define safe methods, whitelist routes, and
// the normalized JSON error envelope returned for blocked write operations.
const (
	demoControlPluginID = "demo-control"
)

// demoControlErrorResponse defines the JSON payload returned for blocked demo writes.
type demoControlErrorResponse struct {
	Code          int            `json:"code"`
	Data          any            `json:"data"`
	Message       string         `json:"message"`
	ErrorCode     string         `json:"errorCode,omitempty"`
	MessageKey    string         `json:"messageKey,omitempty"`
	MessageParams map[string]any `json:"messageParams,omitempty"`
}

// Guard enforces the demo-mode read-only policy on whole-system requests.
func (s *serviceImpl) Guard(request *ghttp.Request) {
	if request == nil {
		return
	}
	if isDemoControlAllowedRequest(request) {
		request.Middleware.Next()
		return
	}
	s.writeDemoControlError(request)
}

// isDemoControlAllowedRequest reports whether the incoming request should bypass
// demo-mode write protection.
func isDemoControlAllowedRequest(request *ghttp.Request) bool {
	if request == nil {
		return true
	}

	method := normalizeDemoControlMethod(request.Method)
	path := normalizeDemoControlPath(request.URL.Path)
	if isDemoControlSafeMethod(method) {
		return true
	}
	if isDemoControlSessionWhitelist(method, path) {
		return true
	}
	return isDemoControlPluginManagementWhitelist(method, path)
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

// isDemoControlPluginManagementWhitelist reports whether the request belongs to
// the minimal plugin-governance whitelist preserved for demo environments.
func isDemoControlPluginManagementWhitelist(method string, path string) bool {
	segments := splitDemoControlPathSegments(path)
	if len(segments) < 4 {
		return false
	}
	if segments[0] != "api" || segments[1] != "v1" || segments[2] != "plugins" {
		return false
	}

	pluginID := strings.TrimSpace(segments[3])
	if pluginID == "" || pluginID == demoControlPluginID {
		return false
	}

	switch method {
	case http.MethodPost:
		return len(segments) == 5 && segments[4] == "install"
	case http.MethodPut:
		return len(segments) == 5 && (segments[4] == "enable" || segments[4] == "disable")
	case http.MethodDelete:
		return len(segments) == 4
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

// splitDemoControlPathSegments converts one canonical request path into ordered
// non-empty path segments for whitelist classification.
func splitDemoControlPathSegments(path string) []string {
	normalizedPath := normalizeDemoControlPath(path)
	if normalizedPath == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(normalizedPath, "/"), "/")
}

// writeDemoControlError writes one JSON error response for blocked write requests.
func (s *serviceImpl) writeDemoControlError(request *ghttp.Request) {
	err := bizerr.NewCode(CodeDemoControlWriteDenied)
	message := err.Error()
	if s != nil && s.i18nSvc != nil {
		message = s.i18nSvc.Translate(
			request.Context(),
			CodeDemoControlWriteDenied.MessageKey(),
			CodeDemoControlWriteDenied.Fallback(),
		)
	}

	request.SetError(err)
	request.Response.Status = http.StatusForbidden
	response := demoControlErrorResponse{
		Code:    CodeDemoControlWriteDenied.TypeCode().Code(),
		Data:    nil,
		Message: message,
	}
	applyDemoControlErrorMetadata(&response, err)
	request.Response.WriteJson(response)
	request.ExitAll()
}

// applyDemoControlErrorMetadata copies structured runtime-message metadata into
// the demo-control rejection response.
func applyDemoControlErrorMetadata(response *demoControlErrorResponse, err error) {
	if response == nil || err == nil {
		return
	}
	messageErr, ok := bizerr.As(err)
	if !ok {
		return
	}
	response.Code = messageErr.TypeCode().Code()
	response.ErrorCode = messageErr.RuntimeCode()
	response.MessageKey = messageErr.MessageKey()
	response.MessageParams = messageErr.Params()
}
