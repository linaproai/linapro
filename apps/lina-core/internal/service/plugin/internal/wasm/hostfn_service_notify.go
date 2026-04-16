// This file implements the governed unified notify host service dispatcher.

package wasm

import (
	"context"
	"encoding/json"

	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/pluginbridge"
)

var notifyHostService = notifysvc.New()

func dispatchNotifyHostService(
	ctx context.Context,
	hcc *hostCallContext,
	channelKey string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, "host call context not available")
	}
	if channelKey == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusCapabilityDenied, "notify host service requires one authorized channel key")
	}

	switch method {
	case pluginbridge.HostServiceMethodNotifySend:
		request, err := pluginbridge.UnmarshalHostServiceNotifySendRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}

		var metadata map[string]any
		if len(request.PayloadJSON) > 0 {
			if err = json.Unmarshal(request.PayloadJSON, &metadata); err != nil {
				return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "notify payloadJson must be valid JSON")
			}
		}

		recipientUserIDs := request.RecipientUserIDs
		if len(recipientUserIDs) == 0 && hcc.identity != nil && hcc.identity.UserID > 0 {
			recipientUserIDs = []int64{int64(hcc.identity.UserID)}
		}

		output, callErr := notifyHostService.Send(ctx, notifysvc.SendInput{
			ChannelKey:       channelKey,
			PluginID:         hcc.pluginID,
			SourceType:       notifysvc.SourceType(request.SourceType),
			SourceID:         request.SourceID,
			CategoryCode:     notifysvc.CategoryCode(request.CategoryCode),
			Title:            request.Title,
			Content:          request.Content,
			Payload:          metadata,
			RecipientUserIDs: recipientUserIDs,
		})
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceNotifySendResponse(&pluginbridge.HostServiceNotifySendResponse{
			MessageID:     output.MessageID,
			DeliveryCount: int32(output.DeliveryCount),
		}))
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported notify host service method: "+method,
		)
	}
}
