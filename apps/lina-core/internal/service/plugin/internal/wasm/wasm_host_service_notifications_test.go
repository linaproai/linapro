// This file tests notification-domain host service dispatch, method-level
// resource authorization, and default recipient handling.

package wasm

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingNotificationsService records notification sends while returning
// deterministic output for host-service dispatch tests.
type trackingNotificationsService struct {
	sendCalls        int
	ensureCalls      int
	lastCapCtx       capmodel.CapabilityContext
	lastInput        notifycap.SendInput
	lastBatchIDs     []notifycap.MessageID
	lastSourceInput  notifycap.BatchGetBySourceInput
	lastEnsureIDs    []notifycap.MessageID
}

// BatchGet records requested message IDs and returns typed projections.
func (s *trackingNotificationsService) BatchGet(_ context.Context, capCtx capmodel.CapabilityContext, ids []notifycap.MessageID) (*capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID], error) {
	s.lastCapCtx = capCtx
	s.lastBatchIDs = append([]notifycap.MessageID(nil), ids...)
	result := &capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID]{
		Items:      map[notifycap.MessageID]*notifycap.MessageProjection{},
		MissingIDs: []notifycap.MessageID{},
	}
	for _, id := range ids {
		result.Items[id] = &notifycap.MessageProjection{
			ID:           id,
			TenantID:     41,
			PluginID:     capCtx.PluginID,
			SourceType:   notifycap.SourceTypePlugin,
			SourceID:     "job-1",
			CategoryCode: notifycap.CategoryCodeOther,
			Title:        "sync done",
			CreatedAt:    1700000000000,
		}
	}
	return result, nil
}

// BatchGetBySource records source lookup input and returns grouped projections.
func (s *trackingNotificationsService) BatchGetBySource(_ context.Context, capCtx capmodel.CapabilityContext, input notifycap.BatchGetBySourceInput) (*notifycap.BatchGetBySourceResult, error) {
	s.lastCapCtx = capCtx
	s.lastSourceInput = input
	return &notifycap.BatchGetBySourceResult{
		Items: map[string][]*notifycap.MessageProjection{
			"job-1": {
				{ID: "9001", TenantID: 41, PluginID: capCtx.PluginID, SourceType: input.SourceType, SourceID: "job-1", CategoryCode: notifycap.CategoryCodeOther, Title: "sync done"},
			},
		},
		MissingIDs: []string{"job-missing"},
	}, nil
}

// EnsureVisible records requested message IDs.
func (s *trackingNotificationsService) EnsureVisible(_ context.Context, capCtx capmodel.CapabilityContext, ids []notifycap.MessageID) error {
	s.ensureCalls++
	s.lastCapCtx = capCtx
	s.lastEnsureIDs = append([]notifycap.MessageID(nil), ids...)
	return nil
}

// Send records one governed notification send request.
func (s *trackingNotificationsService) Send(_ context.Context, capCtx capmodel.CapabilityContext, input notifycap.SendInput) (*notifycap.SendResult, error) {
	s.sendCalls++
	s.lastCapCtx = capCtx
	s.lastInput = input
	return &notifycap.SendResult{
		MessageID:     notifycap.MessageID("9001"),
		DeliveryCount: len(input.Recipients),
	}, nil
}

// TestHandleHostServiceInvokeNotificationsSendDefaultsToCurrentUser verifies
// sends default to the caller when no explicit recipients are provided.
func TestHandleHostServiceInvokeNotificationsSendDefaultsToCurrentUser(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications", "inbox", 42)
	response := invokeNotificationsHostService(
		t,
		hcc,
		"inbox",
		protocol.MarshalHostServiceNotificationsSendRequest(&protocol.HostServiceNotificationsSendRequest{
			Title:        "sync done",
			Content:      "order sync finished",
			SourceType:   "plugin",
			SourceID:     "job-1",
			CategoryCode: "other",
			PayloadJSON:  []byte(`{"scope":"orders"}`),
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("send: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}

	payload, err := protocol.UnmarshalHostServiceNotificationsSendResponse(response.Payload)
	if err != nil {
		t.Fatalf("send payload decode failed: %v", err)
	}
	if payload.MessageID != 9001 || payload.DeliveryCount != 1 {
		t.Fatalf("send payload: got %#v", payload)
	}
	if notifications.sendCalls != 1 {
		t.Fatalf("expected one notification send, got %d", notifications.sendCalls)
	}
	if notifications.lastCapCtx.PluginID != hcc.pluginID || notifications.lastCapCtx.Actor.UserID != 42 {
		t.Fatalf("expected plugin-scoped actor context, got %#v", notifications.lastCapCtx)
	}
	if notifications.lastInput.ChannelKey != "inbox" || notifications.lastInput.SourceID != "job-1" {
		t.Fatalf("expected channel and source to be forwarded, got %#v", notifications.lastInput)
	}
	if len(notifications.lastInput.Recipients) != 1 || notifications.lastInput.Recipients[0] != "42" {
		t.Fatalf("expected current user fallback recipient, got %#v", notifications.lastInput.Recipients)
	}
	if notifications.lastInput.Payload["scope"] != "orders" {
		t.Fatalf("expected payload metadata to be forwarded, got %#v", notifications.lastInput.Payload)
	}
}

// TestHandleHostServiceInvokeNotificationsRejectsInvalidPayloadJSON verifies
// malformed payloadJson content is rejected before the domain service is called.
func TestHandleHostServiceInvokeNotificationsRejectsInvalidPayloadJSON(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	response := invokeNotificationsHostService(
		t,
		newNotificationsHostCallContext("test-plugin-notifications-invalid", "inbox", 1),
		"inbox",
		protocol.MarshalHostServiceNotificationsSendRequest(&protocol.HostServiceNotificationsSendRequest{
			Title:       "sync done",
			Content:     "order sync finished",
			PayloadJSON: []byte("{"),
		}),
	)
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for malformed payloadJson, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if notifications.sendCalls != 0 {
		t.Fatalf("expected no notification send after invalid payload, got %d", notifications.sendCalls)
	}
}

// TestHandleHostServiceInvokeNotificationsRejectsUnauthorizedChannel verifies
// plugins cannot send through notification channels outside their grants.
func TestHandleHostServiceInvokeNotificationsRejectsUnauthorizedChannel(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	response := invokeNotificationsHostService(
		t,
		newNotificationsHostCallContext("test-plugin-notifications-denied", "inbox", 1),
		"ops-webhook",
		protocol.MarshalHostServiceNotificationsSendRequest(&protocol.HostServiceNotificationsSendRequest{
			Title:   "sync done",
			Content: "order sync finished",
		}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized channel, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if notifications.sendCalls != 0 {
		t.Fatalf("expected no notification send after authorization denial, got %d", notifications.sendCalls)
	}
}

// TestHandleHostServiceInvokeNotificationsBatchGetReturnsTypedProjection verifies
// dynamic reads return the stable MessageProjection DTO.
func TestHandleHostServiceInvokeNotificationsBatchGetReturnsTypedProjection(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-batch", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsBatchGetMessages)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsBatchGetMessages,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"9001"}}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("batch_get: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID]
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if out.Items["9001"] == nil || out.Items["9001"].Title != "sync done" {
		t.Fatalf("unexpected typed message projection: %#v", out.Items)
	}
	if len(notifications.lastBatchIDs) != 1 || notifications.lastBatchIDs[0] != "9001" {
		t.Fatalf("expected batch IDs to reach service, got %#v", notifications.lastBatchIDs)
	}
}

// TestHandleHostServiceInvokeNotificationsBatchGetBySource verifies source
// lookups use the notification service once with the full source ID set.
func TestHandleHostServiceInvokeNotificationsBatchGetBySource(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-source", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsBatchGetBySource)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsBatchGetBySource,
		marshalCapabilityJSONRequest(t, notifycap.BatchGetBySourceInput{
			SourceType: notifycap.SourceTypePlugin,
			SourceIDs:  []string{"job-1", "job-missing"},
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("by_source.batch_get: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out notifycap.BatchGetBySourceResult
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if len(out.Items["job-1"]) != 1 || out.Items["job-1"][0].ID != "9001" || len(out.MissingIDs) != 1 {
		t.Fatalf("unexpected source batch response: %#v", out)
	}
	if notifications.lastSourceInput.SourceType != notifycap.SourceTypePlugin || len(notifications.lastSourceInput.SourceIDs) != 2 {
		t.Fatalf("expected source input to reach service, got %#v", notifications.lastSourceInput)
	}
}

// TestHandleHostServiceInvokeNotificationsEnsureVisible verifies visibility
// checks route through the notification service.
func TestHandleHostServiceInvokeNotificationsEnsureVisible(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-visible", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsEnsureVisible)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsEnsureVisible,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"9001"}}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("visible.ensure: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if notifications.ensureCalls != 1 || len(notifications.lastEnsureIDs) != 1 || notifications.lastEnsureIDs[0] != "9001" {
		t.Fatalf("expected visibility IDs to reach service, got calls=%d ids=%#v", notifications.ensureCalls, notifications.lastEnsureIDs)
	}
}

// TestHandleHostServiceInvokeRejectsStandaloneNotifyService verifies the old
// notify service is no longer a public runtime service.
func TestHandleHostServiceInvokeRejectsStandaloneNotifyService(t *testing.T) {
	response := handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, newNotificationsHostCallContext("test-plugin-old-notify", "inbox", 1)),
		protocol.MarshalHostServiceRequestEnvelope(&protocol.HostServiceRequestEnvelope{
			Service:     "notify",
			Method:      "send",
			ResourceRef: "inbox",
			Payload: protocol.MarshalHostServiceNotificationsSendRequest(&protocol.HostServiceNotificationsSendRequest{
				Title:   "sync done",
				Content: "order sync finished",
			}),
		}),
	)
	if response.Status != protocol.HostCallStatusNotFound {
		t.Fatalf("expected old notify service to be not found, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// newNotificationsHostCallContext constructs a notifications-capable host call
// context for one authorized channel and caller identity snapshot.
func newNotificationsHostCallContext(pluginID string, channelKey string, userID int32) *hostCallContext {
	return &hostCallContext{
		pluginID: pluginID,
		capabilities: map[string]struct{}{
			protocol.CapabilityNotifications: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceNotifications,
			Methods: []string{
				protocol.HostServiceMethodNotificationsSend,
			},
			Resources: []*protocol.HostServiceResourceSpec{
				{Ref: channelKey},
			},
		}},
		identity: &protocol.IdentitySnapshotV1{UserID: userID},
	}
}

// invokeNotificationsDomainHostService routes one ordinary notifications domain method.
func invokeNotificationsDomainHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(&protocol.HostServiceRequestEnvelope{
			Service: protocol.HostServiceNotifications,
			Method:  method,
			Payload: payload,
		}),
	)
}

// invokeNotificationsHostService routes one notification host-service request
// through the shared dispatcher and returns its raw response envelope.
func invokeNotificationsHostService(
	t *testing.T,
	hcc *hostCallContext,
	channelKey string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(&protocol.HostServiceRequestEnvelope{
			Service:     protocol.HostServiceNotifications,
			Method:      protocol.HostServiceMethodNotificationsSend,
			ResourceRef: channelKey,
			Payload:     payload,
		}),
	)
}
