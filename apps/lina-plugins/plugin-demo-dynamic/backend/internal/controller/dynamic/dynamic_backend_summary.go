// Backend summary route controller.

package dynamic

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// BackendSummary returns plugin bridge execution summary including plugin
// identity, route metadata, and current user context.
func (c *Controller) BackendSummary(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload := c.dynamicSvc.BuildBackendSummaryPayload(buildBackendSummaryRouteInput(request))
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal backend summary payload failed")
	}
	response := pluginbridge.NewJSONResponse(200, content)
	response.Headers = map[string][]string{
		"X-Lina-Plugin-Bridge":     {"plugin-demo-dynamic"},
		"X-Lina-Plugin-Middleware": {"backend-summary"},
	}
	return response, nil
}

func buildBackendSummaryRouteInput(request *pluginbridge.BridgeRequestEnvelopeV1) *dynamicservice.BackendSummaryInput {
	input := &dynamicservice.BackendSummaryInput{}
	if request == nil {
		return input
	}

	input.PluginID = strings.TrimSpace(request.PluginID)
	if request.Route != nil {
		input.PublicPath = strings.TrimSpace(request.Route.PublicPath)
		input.Access = strings.TrimSpace(request.Route.Access)
		input.Permission = strings.TrimSpace(request.Route.Permission)
	}
	if request.Identity != nil {
		input.Authenticated = request.Identity.UserID > 0
		input.HasIdentity = true
		input.Username = strings.TrimSpace(request.Identity.Username)
		input.IsSuperAdmin = request.Identity.IsSuperAdmin
	}
	return input
}
