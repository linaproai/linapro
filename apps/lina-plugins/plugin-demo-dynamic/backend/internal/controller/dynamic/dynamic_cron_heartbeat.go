// Declared cron heartbeat controller.

package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// CronHeartbeat executes the declared cron heartbeat task for the dynamic
// sample plugin.
func (c *Controller) CronHeartbeat(_ *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.BuildCronHeartbeatPayload()
	if err != nil {
		return nil, err
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal cron heartbeat payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}
