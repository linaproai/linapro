// Demo-record detail route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// DemoRecord returns one plugin-owned demo record detail.
func (c *Controller) DemoRecord(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.GetDemoRecordPayload(readDemoRecordIDFromDetailRoute(request))
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal demo record payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

func readDemoRecordIDFromDetailRoute(request *pluginbridge.BridgeRequestEnvelopeV1) string {
	return readDynamicPathParam(request, "id")
}
