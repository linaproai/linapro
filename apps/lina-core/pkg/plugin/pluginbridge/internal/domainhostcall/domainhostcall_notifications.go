// This file implements guest-side notification capability hostcall clients.

package domainhostcall

import (
	"context"
	"encoding/json"
	"strconv"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// notificationsService adapts notification projection reads to host services.
type notificationsService struct{ baseService }

// Notifications creates the notification-domain guest client.
func Notifications(invoker Invoker, hostInvoker HostServiceInvoker) notifycap.Service {
	return notificationsService{baseService: newBaseServiceWithHostService(invoker, hostInvoker)}
}

// BatchGet returns visible message projections and opaque missing IDs.
func (s notificationsService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []notifycap.MessageID) (*capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID], error) {
	out := &capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID]{Items: map[notifycap.MessageID]*notifycap.MessageProjection{}}
	err := s.callJSONRequest(protocol.HostServiceNotifications, protocol.HostServiceMethodNotificationsBatchGetMessages, idsRequest{IDs: messageIDsToStrings(ids)}, out)
	return out, err
}

// BatchGetBySource returns visible message projections grouped by source ID.
func (s notificationsService) BatchGetBySource(_ context.Context, _ capmodel.CapabilityContext, input notifycap.BatchGetBySourceInput) (*notifycap.BatchGetBySourceResult, error) {
	out := &notifycap.BatchGetBySourceResult{Items: map[string][]*notifycap.MessageProjection{}}
	err := s.callJSONRequest(protocol.HostServiceNotifications, protocol.HostServiceMethodNotificationsBatchGetBySource, input, out)
	return out, err
}

// EnsureVisible rejects when any requested message is absent or outside caller scope.
func (s notificationsService) EnsureVisible(_ context.Context, _ capmodel.CapabilityContext, ids []notifycap.MessageID) error {
	return s.callJSONRequest(protocol.HostServiceNotifications, protocol.HostServiceMethodNotificationsEnsureVisible, idsRequest{IDs: messageIDsToStrings(ids)}, nil)
}

// Send sends one governed notification message.
func (s notificationsService) Send(_ context.Context, _ capmodel.CapabilityContext, input notifycap.SendInput) (*notifycap.SendResult, error) {
	var payloadJSON []byte
	if len(input.Payload) > 0 {
		encoded, err := json.Marshal(input.Payload)
		if err != nil {
			return nil, err
		}
		payloadJSON = encoded
	}
	payload, err := s.callHostService(
		protocol.HostServiceNotifications,
		protocol.HostServiceMethodNotificationsSend,
		input.ChannelKey,
		"",
		protocol.MarshalHostServiceNotificationsSendRequest(&protocol.HostServiceNotificationsSendRequest{
			Title:            input.Title,
			Content:          input.Content,
			SourceType:       string(input.SourceType),
			SourceID:         input.SourceID,
			CategoryCode:     string(input.Category),
			RecipientUserIDs: parseRecipientUserIDs(input.Recipients),
			PayloadJSON:      payloadJSON,
		}),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceNotificationsSendResponse(payload)
	if err != nil {
		return nil, err
	}
	return &notifycap.SendResult{
		MessageID:     notifycap.MessageID(strconv.FormatInt(response.MessageID, 10)),
		DeliveryCount: int(response.DeliveryCount),
	}, nil
}

// messageIDsToStrings converts notification message IDs to transport strings.
func messageIDsToStrings(ids []notifycap.MessageID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// parseRecipientUserIDs converts recipient domain IDs to transport numeric IDs.
func parseRecipientUserIDs(ids []string) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err == nil && parsed > 0 {
			out = append(out, parsed)
		}
	}
	return out
}

var _ notifycap.Service = (*notificationsService)(nil)
