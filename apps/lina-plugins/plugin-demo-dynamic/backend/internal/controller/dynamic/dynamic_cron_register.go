// Cron registration controller.

package dynamic

import "lina-core/pkg/pluginbridge"

// RegisterCrons publishes the dynamic sample plugin's built-in cron
// declarations for host-side discovery.
func (c *Controller) RegisterCrons(_ *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	if err := c.dynamicSvc.RegisterCrons(); err != nil {
		return nil, err
	}
	return pluginbridge.NewSuccessResponse(204, "text/plain; charset=utf-8", nil), nil
}
