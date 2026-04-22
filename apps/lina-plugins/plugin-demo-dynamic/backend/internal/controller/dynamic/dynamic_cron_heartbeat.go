// Declared cron heartbeat controller.

package dynamic

import "lina-core/pkg/pluginbridge"

// CronHeartbeat executes the declared cron heartbeat task for the dynamic
// sample plugin.
func (c *Controller) CronHeartbeat(_ *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.BuildCronHeartbeatPayload()
	if err != nil {
		return nil, err
	}
	return pluginbridge.WriteJSON(200, payload)
}
