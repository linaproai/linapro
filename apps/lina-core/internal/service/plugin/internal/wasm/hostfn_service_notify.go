// This file implements the governed unified notify host service dispatcher.

package wasm

import (
	"context"
	"encoding/json"

	notifysvc "lina-core/internal/service/notify"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// notifyHostService is the shared governed notification backend used by wasm host calls.
// It must be configured via ConfigureNotifyHostService before use.
var notifyHostService notifysvc.Service

// ConfigureNotifyHostService replaces the governed notification backend used
// by wasm host calls. The service must be non-nil.
func ConfigureNotifyHostService(service notifysvc.Service) {
	if service == nil {
		panic("wasm notify host service requires a non-nil notify service")
	}
	notifyHostService = service
}

// dispatchNotifyHostService routes notify host service methods to the governed notification backend.
func dispatchNotifyHostService(
	ctx context.Context,
	hcc *hostCallContext,
	channelKey string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host call context not available")
	}
	if channelKey == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusCapabilityDenied, "notify host service requires one authorized channel key")
	}

	switch method {
	case bridgehostservice.HostServiceMethodNotifySend:
		request, err := bridgehostservice.UnmarshalHostServiceNotifySendRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}

		var metadata map[string]any
		if len(request.PayloadJSON) > 0 {
			if err = json.Unmarshal(request.PayloadJSON, &metadata); err != nil {
				return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "notify payloadJson must be valid JSON")
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
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceNotifySendResponse(&bridgehostservice.HostServiceNotifySendResponse{
			MessageID:     output.MessageID,
			DeliveryCount: int32(output.DeliveryCount),
		}))
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported notify host service method: "+method,
		)
	}
}
