// This file adapts notification host-service calls to the shared notification
// capability service.

package wasm

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	usermsgv1 "lina-core/api/usermsg/v1"
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
		result, err := service.BatchGet(ctx, messageIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodNotificationsList:
		var request notifycap.ListInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.List(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodNotificationsBatchGetBySource:
		var request notifycap.BatchGetBySourceInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGetBySource(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodNotificationsEnsureVisible:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.EnsureVisible(ctx, messageIDs(request.IDs))
		return domainCapabilityResult(struct{}{}, err)
	case bridgehostservice.HostServiceMethodNotificationsSend:
		return handleNotificationsSend(ctx, hcc, service, resourceRef, method, payload)
	case bridgehostservice.HostServiceMethodNotificationsDelete:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Delete(ctx, messageIDs(request.IDs))
		return domainCapabilityResult(struct{}{}, err)
	case bridgehostservice.HostServiceMethodNotificationsDeleteBySource:
		var request notifycap.BatchGetBySourceInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.DeleteBySource(ctx, request.SourceType, request.SourceIDs)
		return domainCapabilityResult(struct{}{}, err)
	case bridgehostservice.HostServiceMethodNotificationsMarkRead:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.MarkRead(ctx, messageIDs(request.IDs))
		return domainCapabilityResult(struct{}{}, err)
	case bridgehostservice.HostServiceMethodNotificationsMarkUnread:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.MarkUnread(ctx, messageIDs(request.IDs))
		return domainCapabilityResult(struct{}{}, err)
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
	output, callErr := service.Send(ctx, notifycap.SendInput{
		ChannelKey: channelKey,
		Recipients: recipients,
		SourceType: usermsgv1.SourceType(request.SourceType),
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
