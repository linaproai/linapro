// This file writes dynamic route bridge responses back to GoFrame without the
// host API success wrapper changing plugin-owned payloads.

package runtime

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/logger"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
)

// writeDynamicRouteResponse writes the guest response back without going through
// GoFrame's default success wrapper, otherwise raw plugin payloads would be
// polluted by host-managed response formatting.
func (s *serviceImpl) writeDynamicRouteResponse(request *ghttp.Request, response *bridgecontract.BridgeResponseEnvelopeV1) {
	if request == nil || response == nil {
		return
	}
	metadata := GetDynamicRouteMetadata(request)
	if metadata != nil {
		metadata.ResponseBody = string(response.Body)
		metadata.ResponseContentType = strings.TrimSpace(response.ContentType)
	}
	for key, values := range response.Headers {
		for _, value := range values {
			request.Response.Header().Add(key, value)
		}
	}
	if strings.TrimSpace(response.ContentType) != "" {
		request.Response.Header().Set("Content-Type", response.ContentType)
	}
	statusCode := int(response.StatusCode)
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	// RawWriter preserves the exact status/body emitted by the bridge envelope.
	request.Response.RawWriter().WriteHeader(statusCode)
	if len(response.Body) > 0 {
		if _, err := request.Response.RawWriter().Write(response.Body); err != nil {
			logger.Warningf(request.Context(), "write dynamic route response body failed err=%v", err)
		}
	}
}
