// Demo-record create route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// CreateDemoRecord creates one plugin-owned demo record.
func (c *Controller) CreateDemoRecord(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	input, err := buildDemoRecordCreateRouteInput(request)
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}

	payload, err := c.dynamicSvc.CreateDemoRecordPayload(input)
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal created demo record payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

func buildDemoRecordCreateRouteInput(request *pluginbridge.BridgeRequestEnvelopeV1) (*dynamicservice.DemoRecordMutationInput, error) {
	return decodeDemoRecordMutationBody(request)
}
