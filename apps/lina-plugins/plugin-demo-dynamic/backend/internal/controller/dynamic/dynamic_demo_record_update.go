// Demo-record update route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// UpdateDemoRecord updates one plugin-owned demo record.
func (c *Controller) UpdateDemoRecord(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	recordID, input, err := buildDemoRecordUpdateRouteInput(request)
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}

	payload, err := c.dynamicSvc.UpdateDemoRecordPayload(recordID, input)
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal updated demo record payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

func buildDemoRecordUpdateRouteInput(
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (string, *dynamicservice.DemoRecordMutationInput, error) {
	input, err := decodeDemoRecordMutationBody(request)
	if err != nil {
		return "", nil, err
	}
	return readDynamicPathParam(request, "id"), input, nil
}
