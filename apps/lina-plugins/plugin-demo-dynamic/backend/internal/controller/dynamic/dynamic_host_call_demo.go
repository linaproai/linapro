// Host call demo route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// HostCallDemo demonstrates unified host service capabilities including runtime,
// governed storage, outbound HTTP, and structured data access.
func (c *Controller) HostCallDemo(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.BuildHostCallDemoPayload(request)
	if err != nil {
		return nil, err
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal host call demo payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}
