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
	"lina-core/pkg/pluginbridge"
)

const (
	defaultNetworkTimeout      = 10 * time.Second
	defaultNetworkMaxBodyBytes = 2 << 20
)

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

func dispatchNetworkHostService(
	ctx context.Context,
	hcc *hostCallContext,
	targetURL string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if strings.TrimSpace(targetURL) == "" {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			"network host service requires one authorized target URL",
		)
	}

	switch method {
	case pluginbridge.HostServiceMethodNetworkRequest:
		return handleNetworkRequest(ctx, hcc, targetURL, payload)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported network host service method: "+method,
		)
	}
}

func handleNetworkRequest(
	ctx context.Context,
	hcc *hostCallContext,
	targetURL string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceNetworkRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if request == nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "network request is required")
	}

	resolvedURL, err := normalizeNetworkTargetURL(targetURL)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if hcc != nil && !hcc.hasHostServiceAccess(pluginbridge.HostServiceNetwork, pluginbridge.HostServiceMethodNetworkRequest, resolvedURL, "") {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			"network request target URL is not authorized: "+resolvedURL,
		)
	}

	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method == "" {
		method = http.MethodGet
	}
	if int64(len(request.Body)) > defaultNetworkMaxBodyBytes {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusInvalidRequest,
			"network request body exceeds platform size limit",
		)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, resolvedURL, bytes.NewReader(request.Body))
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = applyNetworkRequestHeaders(httpRequest, request.Headers); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	clientCtx, cancel := context.WithTimeout(ctx, defaultNetworkTimeout)
	defer cancel()
	httpRequest = httpRequest.WithContext(clientCtx)

	httpResponse, err := (&http.Client{}).Do(httpRequest)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	defer func() {
		if closeErr := httpResponse.Body.Close(); closeErr != nil {
			logger.Warningf(ctx, "close network response body failed err=%v", closeErr)
		}
	}()

	body, err := readNetworkResponseBody(httpResponse.Body, defaultNetworkMaxBodyBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	response := &pluginbridge.HostServiceNetworkResponse{
		StatusCode:  int32(httpResponse.StatusCode),
		Headers:     flattenResponseHeaders(httpResponse.Header),
		Body:        body,
		ContentType: normalizeNetworkContentType(httpResponse.Header.Get("Content-Type")),
	}
	return pluginbridge.NewHostCallSuccessResponse(
		pluginbridge.MarshalHostServiceNetworkResponse(response),
	)
}

func normalizeNetworkTargetURL(rawValue string) (string, error) {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return "", gerror.New("network request URL is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", gerror.Wrap(err, "network request URL 非法")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", gerror.Newf("network request URL scheme 不支持: %s", parsed.Scheme)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", gerror.New("network request URL 缺少 host")
	}
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}

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
			return gerror.Newf("network request header 不允许设置: %s", key)
		}
		request.Header.Set(textproto.CanonicalMIMEHeaderKey(key), value)
	}
	return nil
}

func matchAuthorizedNetworkResource(
	specs []*pluginbridge.HostServiceSpec,
	targetURL string,
) *pluginbridge.HostServiceResourceSpec {
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
		if spec == nil || spec.Service != pluginbridge.HostServiceNetwork {
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

func matchNetworkPathPrefix(patternPath string, targetPath string) bool {
	normalizedPattern := normalizeAuthorizedNetworkPath(patternPath)
	normalizedTarget := normalizeAuthorizedNetworkPath(targetPath)
	if normalizedPattern == "/" {
		return true
	}
	return normalizedTarget == normalizedPattern || strings.HasPrefix(normalizedTarget, normalizedPattern+"/")
}

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
