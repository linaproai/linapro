// This file adapts notification host-service calls to the shared notification
// capability service.

package wasm

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"lina-core/pkg/plugin/capability/notifycap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchNotificationsHostService routes notification-domain host-service calls.
func dispatchNotificationsHostService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := notificationsServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("notifications")
	}
	switch method {
	case bridgehostservice.HostServiceMethodNotificationsBatchGetMessages:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceNotifications, method)
		result, err := service.BatchGet(ctx, capCtx, messageIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodNotificationsBatchGetBySource:
		var request notifycap.BatchGetBySourceInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceNotifications, method)
		result, err := service.BatchGetBySource(ctx, capCtx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodNotificationsEnsureVisible:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceNotifications, method)
		err := service.EnsureVisible(ctx, capCtx, messageIDs(request.IDs))
		return domainCapabilityResult(struct{}{}, err)
	case bridgehostservice.HostServiceMethodNotificationsSend:
		return handleNotificationsSend(ctx, hcc, service, resourceRef, method, payload)
	default:
		return domainMethodNotFound("notifications", method)
	}
}

// notificationsServiceForHostCall resolves the notification service for one host call.
func notificationsServiceForHostCall(hcc *hostCallContext) notifycap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Notifications()
}

// messageIDs converts transport string identifiers into typed notification IDs.
func messageIDs(ids []string) []notifycap.MessageID {
	out := make([]notifycap.MessageID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, notifycap.MessageID(value))
		}
	}
	return out
}

// handleNotificationsSend decodes one resource-scoped notification send request.
func handleNotificationsSend(
	ctx context.Context,
	hcc *hostCallContext,
	service notifycap.Service,
	channelKey string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if strings.TrimSpace(channelKey) == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusCapabilityDenied, "notifications.messages.send requires one authorized channel key")
	}
	request, err := bridgehostservice.UnmarshalHostServiceNotificationsSendRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	var metadata map[string]any
	if len(request.PayloadJSON) > 0 {
		if err = json.Unmarshal(request.PayloadJSON, &metadata); err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "notifications payloadJson must be valid JSON")
		}
	}
	recipients := recipientIDs(request.RecipientUserIDs)
	if len(recipients) == 0 && hcc != nil && hcc.identity != nil && hcc.identity.UserID > 0 {
		recipients = []string{strconv.FormatInt(int64(hcc.identity.UserID), 10)}
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceNotifications, method)
	output, callErr := service.Send(ctx, capCtx, notifycap.SendInput{
		ChannelKey: channelKey,
		Recipients: recipients,
		SourceType: notifycap.SourceType(request.SourceType),
		SourceID:   request.SourceID,
		Title:      request.Title,
		Content:    request.Content,
		Category:   notifycap.CategoryCode(request.CategoryCode),
		Payload:    metadata,
	})
	if callErr != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
	}
	response := &bridgehostservice.HostServiceNotificationsSendResponse{}
	if output != nil {
		response.DeliveryCount = int32(output.DeliveryCount)
		if id, parseErr := strconv.ParseInt(string(output.MessageID), 10, 64); parseErr == nil {
			response.MessageID = id
		}
	}
	return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceNotificationsSendResponse(response))
}

// recipientIDs converts transport numeric recipients into capability IDs.
func recipientIDs(ids []int64) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			out = append(out, strconv.FormatInt(id, 10))
		}
	}
	return out
}
