// Demo-record delete route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// DeleteDemoRecord deletes one plugin-owned demo record.
func (c *Controller) DeleteDemoRecord(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.DeleteDemoRecordPayload(readDemoRecordIDFromDeleteRoute(request))
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal deleted demo record payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

// readDemoRecordIDFromDeleteRoute reads the record identifier from the delete
// route path parameters.
func readDemoRecordIDFromDeleteRoute(request *pluginbridge.BridgeRequestEnvelopeV1) string {
	return readDynamicPathParam(request, "id")
}
