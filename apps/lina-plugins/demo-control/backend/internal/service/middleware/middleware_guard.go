// This file implements request classification and rejection logic for the
// demo-control middleware.

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
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
	if !s.isDemoControlEnabled(request.Context()) {
		request.Middleware.Next()
		return
	}
	s.writeDemoControlError(request)
}

// isDemoControlEnabled reports whether the current plugin state activates demo protection.
func (s *serviceImpl) isDemoControlEnabled(ctx context.Context) bool {
	if s == nil || s.enablementReader == nil {
		return false
	}
	return s.enablementReader.IsEnabled(ctx, demoControlPluginID)
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
	return false
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
