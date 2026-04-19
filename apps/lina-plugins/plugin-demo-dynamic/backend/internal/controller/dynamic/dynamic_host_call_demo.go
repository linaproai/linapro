// Host call demo route controller.

package dynamic

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// HostCallDemo demonstrates unified host service capabilities including runtime,
// governed storage, outbound HTTP, and structured data access.
func (c *Controller) HostCallDemo(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.BuildHostCallDemoPayload(buildHostCallDemoRouteInput(request))
	if err != nil {
		return nil, err
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal host call demo payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

// buildHostCallDemoRouteInput extracts host-call demo metadata and flags from
// the bridge request envelope.
func buildHostCallDemoRouteInput(request *pluginbridge.BridgeRequestEnvelopeV1) *dynamicservice.HostCallDemoInput {
	input := &dynamicservice.HostCallDemoInput{}
	if request == nil {
		return input
	}

	input.PluginID = strings.TrimSpace(request.PluginID)
	input.RequestID = strings.TrimSpace(request.RequestID)
	if request.Route != nil {
		input.RoutePath = strings.TrimSpace(request.Route.InternalPath)
		input.SkipNetwork = hasDynamicQueryFlag(request, "skipNetwork")
	}
	if request.Identity != nil {
		input.Username = strings.TrimSpace(request.Identity.Username)
	}
	return input
}
