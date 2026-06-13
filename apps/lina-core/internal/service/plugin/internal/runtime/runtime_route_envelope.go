// This file builds dynamic route bridge request envelopes and snapshots the
// HTTP request data that guest route handlers are allowed to observe.

package runtime

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
)

// buildDynamicRouteRequestEnvelopeWithIdentity snapshots the matched request
// into the bridge payload forwarded to guest code.
func (s *serviceImpl) buildDynamicRouteRequestEnvelopeWithIdentity(
	match *dynamicRouteMatch,
	request *ghttp.Request,
	identity *bridgecontract.IdentitySnapshotV1,
) (*bridgecontract.BridgeRequestEnvelopeV1, error) {
	body := request.GetBody()
	queryValues := request.URL.Query()
	return &bridgecontract.BridgeRequestEnvelopeV1{
		PluginID: match.PluginID,
		Route: &bridgecontract.RouteMatchSnapshotV1{
			Method:       strings.ToUpper(strings.TrimSpace(request.Method)),
			PublicPath:   match.PublicPath,
			InternalPath: match.InternalPath,
			RoutePath:    match.Route.Path,
			Access:       match.Route.Access,
			Permission:   match.Route.Permission,
			RequestType:  match.Route.RequestType,
			PathParams:   cloneStringMap(match.PathParams),
			QueryValues:  cloneURLValues(queryValues),
		},
		Request: &bridgecontract.HTTPRequestSnapshotV1{
			Method:       strings.ToUpper(strings.TrimSpace(request.Method)),
			PublicPath:   match.PublicPath,
			InternalPath: match.InternalPath,
			RawPath:      request.URL.Path,
			RawQuery:     request.URL.RawQuery,
			Host:         request.Host,
			Scheme:       request.URL.Scheme,
			RemoteAddr:   request.Request.RemoteAddr,
			ClientIP:     request.GetClientIp(),
			ContentType:  request.Header.Get("Content-Type"),
			Headers:      sanitizeDynamicRouteHeaders(request.Header),
			Cookies:      collectRequestCookies(request),
			Body:         append([]byte(nil), body...),
		},
		Identity:  identity,
		RequestID: buildDynamicRequestID(match, request),
	}, nil
}

// sanitizeDynamicRouteHeaders clones request headers while stripping bearer tokens.
func sanitizeDynamicRouteHeaders(headers http.Header) map[string][]string {
	result := make(map[string][]string)
	if len(headers) == 0 {
		return result
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if strings.EqualFold(key, "Authorization") {
			continue
		}
		values := headers.Values(key)
		if len(values) == 0 {
			continue
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}

// collectRequestCookies snapshots request cookies into a simple name-value map.
func collectRequestCookies(request *ghttp.Request) map[string]string {
	result := make(map[string]string)
	if request == nil || request.Request == nil {
		return result
	}
	for _, cookie := range request.Request.Cookies() {
		if cookie == nil {
			continue
		}
		result[cookie.Name] = cookie.Value
	}
	return result
}

// cloneURLValues deep-copies URL query values for bridge payload serialization.
func cloneURLValues(values url.Values) map[string][]string {
	result := make(map[string][]string, len(values))
	for key, items := range values {
		result[key] = append([]string(nil), items...)
	}
	return result
}

// cloneStringMap deep-copies string maps used in request and route snapshots.
func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

// buildDynamicRequestID derives a stable host-side request ID for bridge logging.
func buildDynamicRequestID(match *dynamicRouteMatch, request *ghttp.Request) string {
	if request == nil {
		return match.PluginID + ":" + base64.StdEncoding.EncodeToString([]byte(match.InternalPath))
	}
	return match.PluginID + ":" + request.Method + ":" + match.InternalPath
}
