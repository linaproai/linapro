// Demo-record list route controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// DemoRecordList returns one paged list of plugin-owned demo records.
func (c *Controller) DemoRecordList(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.ListDemoRecordsPayload(buildDemoRecordListRouteInput(request))
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal demo record list payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}

func buildDemoRecordListRouteInput(request *pluginbridge.BridgeRequestEnvelopeV1) *dynamicservice.DemoRecordListInput {
	return &dynamicservice.DemoRecordListInput{
		PageNum:  readDynamicQueryInt(request, "pageNum"),
		PageSize: readDynamicQueryInt(request, "pageSize"),
		Keyword:  readDynamicQueryValue(request, "keyword"),
	}
}
