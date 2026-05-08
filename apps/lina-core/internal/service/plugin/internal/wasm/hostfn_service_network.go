// This file implements the governed outbound HTTP host service backed by
// authorized URL-pattern matching and platform-level request protections.

package wasm

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// Default timeout and size limits for governed outbound network requests.
const (
	defaultNetworkTimeout      = 10 * time.Second
	defaultNetworkMaxBodyBytes = 2 << 20
)

// protectedNetworkRequestHeaders lists hop-by-hop or host-controlled headers blocked from guest code.
var protectedNetworkRequestHeaders = map[string]struct{}{
	"connection":        {},
	"content-length":    {},
	"host":              {},
	"proxy-connection":  {},
	"te":                {},
	"trailer":           {},
	"transfer-encoding": {},
	"upgrade":           {},
}

// dispatchNetworkHostService routes network host service methods to the
// governed outbound HTTP request handler.
func dispatchNetworkHostService(
	ctx context.Context,
	hcc *hostCallContext,
	targetURL string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if strings.TrimSpace(targetURL) == "" {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"network host service requires one authorized target URL",
		)
	}

	switch method {
	case bridgehostservice.HostServiceMethodNetworkRequest:
		return handleNetworkRequest(ctx, hcc, targetURL, payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported network host service method: "+method,
		)
	}
}

// handleNetworkRequest validates and executes one outbound HTTP request against
// the authorized upstream resource snapshot.
func handleNetworkRequest(
	ctx context.Context,
	hcc *hostCallContext,
	targetURL string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceNetworkRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if request == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "network request is required")
	}

	resolvedURL, err := normalizeNetworkTargetURL(targetURL)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if hcc != nil && !hcc.hasHostServiceAccess(bridgehostservice.HostServiceNetwork, bridgehostservice.HostServiceMethodNetworkRequest, resolvedURL, "") {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"network request target URL is not authorized: "+resolvedURL,
		)
	}

	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method == "" {
		method = http.MethodGet
	}
	if int64(len(request.Body)) > defaultNetworkMaxBodyBytes {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			"network request body exceeds platform size limit",
		)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, resolvedURL, bytes.NewReader(request.Body))
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = applyNetworkRequestHeaders(httpRequest, request.Headers); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	clientCtx, cancel := context.WithTimeout(ctx, defaultNetworkTimeout)
	defer cancel()
	httpRequest = httpRequest.WithContext(clientCtx)

	httpResponse, err := (&http.Client{}).Do(httpRequest)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	defer func() {
		if closeErr := httpResponse.Body.Close(); closeErr != nil {
			logger.Warningf(ctx, "close network response body failed err=%v", closeErr)
		}
	}()

	body, err := readNetworkResponseBody(httpResponse.Body, defaultNetworkMaxBodyBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	response := &bridgehostservice.HostServiceNetworkResponse{
		StatusCode:  int32(httpResponse.StatusCode),
		Headers:     flattenResponseHeaders(httpResponse.Header),
		Body:        body,
		ContentType: normalizeNetworkContentType(httpResponse.Header.Get("Content-Type")),
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceNetworkResponse(response),
	)
}

// normalizeNetworkTargetURL validates and canonicalizes the target URL.
func normalizeNetworkTargetURL(rawValue string) (string, error) {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return "", gerror.New("network request URL is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", gerror.Wrap(err, "network request URL is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", gerror.Newf("network request URL scheme is not supported: %s", parsed.Scheme)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", gerror.New("network request URL is missing host")
	}
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}

// applyNetworkRequestHeaders copies guest-requested headers while blocking protected ones.
func applyNetworkRequestHeaders(
	request *http.Request,
	headers map[string]string,
) error {
	if request == nil {
		return gerror.New("network request is nil")
	}

	for key, value := range headers {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" {
			continue
		}
		if _, ok := protectedNetworkRequestHeaders[normalizedKey]; ok {
			return gerror.Newf("network request header is not allowed: %s", key)
		}
		request.Header.Set(textproto.CanonicalMIMEHeaderKey(key), value)
	}
	return nil
}

// matchAuthorizedNetworkResource finds the authorized upstream resource that matches the target URL.
func matchAuthorizedNetworkResource(
	specs []*bridgehostservice.HostServiceSpec,
	targetURL string,
) *bridgehostservice.HostServiceResourceSpec {
	normalizedTarget, err := url.Parse(strings.TrimSpace(targetURL))
	if err != nil || normalizedTarget == nil {
		return nil
	}
	// Network authorization is matched structurally so one approved pattern can
	// cover host wildcards and path prefixes while query and fragment remain
	// irrelevant to governance decisions.
	var (
		targetHost = strings.ToLower(strings.TrimSpace(normalizedTarget.Hostname()))
		targetPort = strings.TrimSpace(normalizedTarget.Port())
		targetPath = normalizeAuthorizedNetworkPath(normalizedTarget.Path)
	)

	for _, spec := range specs {
		if spec == nil || spec.Service != bridgehostservice.HostServiceNetwork {
			continue
		}
		for _, resource := range spec.Resources {
			if resource == nil {
				continue
			}
			pattern, err := url.Parse(strings.TrimSpace(resource.Ref))
			if err != nil || pattern == nil {
				continue
			}
			if !strings.EqualFold(pattern.Scheme, normalizedTarget.Scheme) {
				continue
			}
			if !matchNetworkHostPattern(strings.ToLower(strings.TrimSpace(pattern.Hostname())), targetHost) {
				continue
			}
			patternPort := strings.TrimSpace(pattern.Port())
			if patternPort != "" && patternPort != targetPort {
				continue
			}
			if !matchNetworkPathPrefix(normalizeAuthorizedNetworkPath(pattern.Path), targetPath) {
				continue
			}
			return resource
		}
	}
	return nil
}

// matchNetworkHostPattern matches exact hosts and leading-wildcard host patterns.
func matchNetworkHostPattern(pattern string, target string) bool {
	if pattern == "" || target == "" {
		return false
	}
	if pattern == target {
		return true
	}
	matched, err := path.Match(pattern, target)
	return err == nil && matched
}

// normalizeAuthorizedNetworkPath canonicalizes authorization paths for prefix matching.
func normalizeAuthorizedNetworkPath(rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	normalized := path.Clean("/" + strings.TrimPrefix(strings.ReplaceAll(trimmed, "\\", "/"), "/"))
	if normalized == "." {
		return "/"
	}
	return normalized
}

// matchNetworkPathPrefix reports whether the target path is within the authorized path prefix.
func matchNetworkPathPrefix(patternPath string, targetPath string) bool {
	normalizedPattern := normalizeAuthorizedNetworkPath(patternPath)
	normalizedTarget := normalizeAuthorizedNetworkPath(targetPath)
	if normalizedPattern == "/" {
		return true
	}
	return normalizedTarget == normalizedPattern || strings.HasPrefix(normalizedTarget, normalizedPattern+"/")
}

// readNetworkResponseBody reads at most maxBodyBytes+1 bytes to enforce response size limits.
func readNetworkResponseBody(reader io.Reader, maxBodyBytes int64) ([]byte, error) {
	if reader == nil {
		return nil, nil
	}
	if maxBodyBytes <= 0 {
		return io.ReadAll(reader)
	}

	limited := io.LimitReader(reader, maxBodyBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBodyBytes {
		return nil, gerror.New("network response body exceeds platform size limit")
	}
	return body, nil
}

// flattenResponseHeaders joins multi-value response headers into a string map.
func flattenResponseHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		result[textproto.CanonicalMIMEHeaderKey(key)] = values[0]
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// normalizeNetworkContentType trims parameters and lowercases the content type.
func normalizeNetworkContentType(contentType string) string {
	trimmed := strings.TrimSpace(contentType)
	if trimmed == "" {
		return ""
	}
	if index := strings.Index(trimmed, ";"); index >= 0 {
		trimmed = trimmed[:index]
	}
	return strings.ToLower(strings.TrimSpace(trimmed))
}
