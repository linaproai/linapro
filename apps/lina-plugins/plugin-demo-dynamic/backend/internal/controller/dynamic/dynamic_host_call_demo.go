// Host call demo route controller.

package dynamic

import (
	"context"
	"strings"

	"lina-core/pkg/pluginbridge"
	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// HostCallDemo demonstrates unified host service capabilities including runtime,
// governed storage, outbound HTTP, and structured data access.
func (c *Controller) HostCallDemo(
	ctx context.Context,
	req *v1.HostCallDemoReq,
) (res *v1.HostCallDemoRes, err error) {
	payload, err := c.dynamicSvc.BuildHostCallDemoPayload(
		buildHostCallDemoRouteInput(pluginbridge.RequestEnvelopeFromContext(ctx), req),
	)
	if err != nil {
		return nil, err
	}
	return &v1.HostCallDemoRes{
		VisitCount: payload.VisitCount,
		PluginID:   payload.PluginID,
		Runtime: &v1.HostCallDemoRuntimeRes{
			Now:  payload.Runtime.Now,
			UUID: payload.Runtime.UUID,
			Node: payload.Runtime.Node,
		},
		Storage: &v1.HostCallDemoStorageRes{
			PathPrefix:  payload.Storage.PathPrefix,
			ObjectPath:  payload.Storage.ObjectPath,
			Stored:      payload.Storage.Stored,
			ListedCount: payload.Storage.ListedCount,
			Deleted:     payload.Storage.Deleted,
		},
		Network: &v1.HostCallDemoNetworkRes{
			URL:         payload.Network.URL,
			Skipped:     payload.Network.Skipped,
			StatusCode:  payload.Network.StatusCode,
			ContentType: payload.Network.ContentType,
			BodyPreview: payload.Network.BodyPreview,
			Error:       payload.Network.Error,
		},
		Data: &v1.HostCallDemoDataRes{
			Table:      payload.Data.Table,
			RecordKey:  payload.Data.RecordKey,
			ListTotal:  payload.Data.ListTotal,
			CountTotal: payload.Data.CountTotal,
			Updated:    payload.Data.Updated,
			Deleted:    payload.Data.Deleted,
		},
		Message: payload.Message,
	}, nil
}

// buildHostCallDemoRouteInput extracts host-call demo metadata and flags from
// the bridge request envelope.
func buildHostCallDemoRouteInput(
	request *pluginbridge.BridgeRequestEnvelopeV1,
	req *v1.HostCallDemoReq,
) *dynamicservice.HostCallDemoInput {
	input := &dynamicservice.HostCallDemoInput{}
	if request == nil {
		if req != nil {
			input.SkipNetwork = req.SkipNetwork
		}
		return input
	}

	input.PluginID = strings.TrimSpace(request.PluginID)
	input.RequestID = strings.TrimSpace(request.RequestID)
	if request.Route != nil {
		input.RoutePath = strings.TrimSpace(request.Route.InternalPath)
	}
	if req != nil {
		input.SkipNetwork = req.SkipNetwork
	}
	if request.Identity != nil {
		input.Username = strings.TrimSpace(request.Identity.Username)
	}
	return input
}
