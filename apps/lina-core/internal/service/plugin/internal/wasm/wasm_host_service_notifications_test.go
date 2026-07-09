// This file tests notification-domain host service dispatch, method-level
// resource authorization, and default recipient handling.

package wasm

import (
	"context"
	"testing"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingNotificationsService records notification sends while returning
// deterministic output for host-service dispatch tests.
type trackingNotificationsService struct {
	sendCalls            int
	ensureCalls          int
	deleteCalls          int
	deleteBySourceCalls  int
	markReadCalls        int
	markUnreadCalls      int
	lastCurrent          bizctxcap.CurrentContext
	lastInput            notifycap.SendInput
	lastBatchIDs         []notifycap.MessageID
	lastListInput        notifycap.ListInput
	lastSourceInput      notifycap.BatchGetBySourceInput
	lastEnsureIDs        []notifycap.MessageID
	lastDeleteIDs        []notifycap.MessageID
	lastDeleteSourceType usermsgv1.SourceType
	lastDeleteSourceIDs  []string
	lastMarkReadIDs      []notifycap.MessageID
	lastMarkUnreadIDs    []notifycap.MessageID
}

// BatchGet records requested message IDs and returns typed projections.
func (s *trackingNotificationsService) BatchGet(ctx context.Context, ids []notifycap.MessageID) (*capmodel.BatchResult[*notifycap.MessageInfo, notifycap.MessageID], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastBatchIDs = append([]notifycap.MessageID(nil), ids...)
	result := &capmodel.BatchResult[*notifycap.MessageInfo, notifycap.MessageID]{
		Items:      map[notifycap.MessageID]*notifycap.MessageInfo{},
		MissingIDs: []notifycap.MessageID{},
	}
	for _, id := range ids {
		result.Items[id] = &notifycap.MessageInfo{
			ID:           id,
			TenantID:     41,
			PluginID:     "test-plugin-notifications",
			SourceType:   usermsgv1.SourceTypePlugin,
			SourceID:     "job-1",
			CategoryCode: notifycap.CategoryCodeOther,
			Title:        "sync done",
			CreatedAt:    1700000000000,
		}
	}
	return result, nil
}

// Get returns one typed message projection.
func (s *trackingNotificationsService) Get(ctx context.Context, id notifycap.MessageID) (*notifycap.MessageInfo, error) {
	result, err := s.BatchGet(ctx, []notifycap.MessageID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

// List records list input and returns one typed projection.
func (s *trackingNotificationsService) List(ctx context.Context, input notifycap.ListInput) (*capmodel.PageResult[*notifycap.MessageInfo], error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastListInput = input
	return &capmodel.PageResult[*notifycap.MessageInfo]{
		Items: []*notifycap.MessageInfo{
			{ID: "9001", TenantID: 41, PluginID: "test-plugin-notifications", SourceType: input.SourceType, SourceID: input.SourceID, CategoryCode: notifycap.CategoryCodeOther, Title: "sync done"},
		},
		Total: 1,
	}, nil
}

// BatchGetBySource records source lookup input and returns grouped projections.
func (s *trackingNotificationsService) BatchGetBySource(ctx context.Context, input notifycap.BatchGetBySourceInput) (*notifycap.BatchGetBySourceResult, error) {
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastSourceInput = input
	return &notifycap.BatchGetBySourceResult{
		Items: map[string][]*notifycap.MessageInfo{
			"job-1": {
				{ID: "9001", TenantID: 41, PluginID: "test-plugin-notifications", SourceType: input.SourceType, SourceID: "job-1", CategoryCode: notifycap.CategoryCodeOther, Title: "sync done"},
			},
		},
		MissingIDs: []string{"job-missing"},
	}, nil
}

// EnsureVisible records requested message IDs.
func (s *trackingNotificationsService) EnsureVisible(ctx context.Context, ids []notifycap.MessageID) error {
	s.ensureCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastEnsureIDs = append([]notifycap.MessageID(nil), ids...)
	return nil
}

// Send records one governed notification send request.
func (s *trackingNotificationsService) Send(ctx context.Context, input notifycap.SendInput) (*notifycap.SendResult, error) {
	s.sendCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastInput = input
	return &notifycap.SendResult{
		MessageID:     notifycap.MessageID("9001"),
		DeliveryCount: len(input.Recipients),
	}, nil
}

// Delete records requested message IDs.
func (s *trackingNotificationsService) Delete(ctx context.Context, ids []notifycap.MessageID) error {
	s.deleteCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteIDs = append([]notifycap.MessageID(nil), ids...)
	return nil
}

// DeleteBySource records requested business source IDs.
func (s *trackingNotificationsService) DeleteBySource(ctx context.Context, sourceType usermsgv1.SourceType, sourceIDs []string) error {
	s.deleteBySourceCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastDeleteSourceType = sourceType
	s.lastDeleteSourceIDs = append([]string(nil), sourceIDs...)
	return nil
}

// MarkRead records requested read-state changes.
func (s *trackingNotificationsService) MarkRead(ctx context.Context, ids []notifycap.MessageID) error {
	s.markReadCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastMarkReadIDs = append([]notifycap.MessageID(nil), ids...)
	return nil
}

// MarkUnread records requested read-state changes.
func (s *trackingNotificationsService) MarkUnread(ctx context.Context, ids []notifycap.MessageID) error {
	s.markUnreadCalls++
	s.lastCurrent = bizctxcap.CurrentFromContext(ctx)
	s.lastMarkUnreadIDs = append([]notifycap.MessageID(nil), ids...)
	return nil
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
	if notifications.lastCurrent.UserID != 42 {
		t.Fatalf("expected current user context, got %#v", notifications.lastCurrent)
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
// dynamic reads return the stable MessageInfo DTO.
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
	var out capmodel.BatchResult[*notifycap.MessageInfo, notifycap.MessageID]
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if out.Items["9001"] == nil || out.Items["9001"].Title != "sync done" {
		t.Fatalf("unexpected typed message projection: %#v", out.Items)
	}
	if len(notifications.lastBatchIDs) != 1 || notifications.lastBatchIDs[0] != "9001" {
		t.Fatalf("expected batch IDs to reach service, got %#v", notifications.lastBatchIDs)
	}
}

// TestHandleHostServiceInvokeNotificationsList verifies dynamic list calls are
// routed to the notification capability service with bounded page input.
func TestHandleHostServiceInvokeNotificationsList(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-list", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsList)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsList,
		marshalCapabilityJSONRequest(t, notifycap.ListInput{
			SourceType: usermsgv1.SourceTypePlugin,
			SourceID:   "job-1",
			Page:       capmodel.PageRequest{PageNum: 2, PageSize: 10},
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("messages.list: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	var out capmodel.PageResult[*notifycap.MessageInfo]
	decodeCapabilityJSONResponse(t, response.Payload, &out)
	if out.Total != 1 || len(out.Items) != 1 || out.Items[0].SourceID != "job-1" {
		t.Fatalf("unexpected list response: %#v", out)
	}
	if notifications.lastListInput.SourceID != "job-1" || notifications.lastListInput.Page.PageSize != 10 {
		t.Fatalf("expected list input to reach service, got %#v", notifications.lastListInput)
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
			SourceType: usermsgv1.SourceTypePlugin,
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
	if notifications.lastSourceInput.SourceType != usermsgv1.SourceTypePlugin || len(notifications.lastSourceInput.SourceIDs) != 2 {
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

// TestHandleHostServiceInvokeNotificationsDelete verifies dynamic delete calls
// route through the notification capability service.
func TestHandleHostServiceInvokeNotificationsDelete(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-delete", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsDelete)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsDelete,
		marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{"9001"}}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("messages.delete: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if notifications.deleteCalls != 1 || len(notifications.lastDeleteIDs) != 1 || notifications.lastDeleteIDs[0] != "9001" {
		t.Fatalf("expected delete IDs to reach service, got calls=%d ids=%#v", notifications.deleteCalls, notifications.lastDeleteIDs)
	}
}

// TestHandleHostServiceInvokeNotificationsDeleteBySource verifies source-based
// dynamic deletes forward source type and source IDs.
func TestHandleHostServiceInvokeNotificationsDeleteBySource(t *testing.T) {
	notifications := &trackingNotificationsService{}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
		notifications: notifications,
	})

	hcc := newNotificationsHostCallContext("test-plugin-notifications-delete-source", "inbox", 42)
	hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, protocol.HostServiceMethodNotificationsDeleteBySource)
	response := invokeNotificationsDomainHostService(
		t,
		hcc,
		protocol.HostServiceMethodNotificationsDeleteBySource,
		marshalCapabilityJSONRequest(t, notifycap.BatchGetBySourceInput{
			SourceType: usermsgv1.SourceTypePlugin,
			SourceIDs:  []string{"job-1", "job-2"},
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("messages.by_source.delete: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if notifications.deleteBySourceCalls != 1 || notifications.lastDeleteSourceType != usermsgv1.SourceTypePlugin || len(notifications.lastDeleteSourceIDs) != 2 {
		t.Fatalf("expected source delete input to reach service, got calls=%d type=%s ids=%#v", notifications.deleteBySourceCalls, notifications.lastDeleteSourceType, notifications.lastDeleteSourceIDs)
	}
}

// TestHandleHostServiceInvokeNotificationsReadState verifies dynamic read-state
// methods route through the notification capability service.
func TestHandleHostServiceInvokeNotificationsReadState(t *testing.T) {
	for _, tc := range []struct {
		name       string
		method     string
		assertCall func(*trackingNotificationsService)
	}{
		{
			name:   "mark_read",
			method: protocol.HostServiceMethodNotificationsMarkRead,
			assertCall: func(notifications *trackingNotificationsService) {
				if notifications.markReadCalls != 1 || len(notifications.lastMarkReadIDs) != 1 || notifications.lastMarkReadIDs[0] != "9001" {
					t.Fatalf("expected mark read IDs to reach service, got calls=%d ids=%#v", notifications.markReadCalls, notifications.lastMarkReadIDs)
				}
			},
		},
		{
			name:   "mark_unread",
			method: protocol.HostServiceMethodNotificationsMarkUnread,
			assertCall: func(notifications *trackingNotificationsService) {
				if notifications.markUnreadCalls != 1 || len(notifications.lastMarkUnreadIDs) != 1 || notifications.lastMarkUnreadIDs[0] != "9002" {
					t.Fatalf("expected mark unread IDs to reach service, got calls=%d ids=%#v", notifications.markUnreadCalls, notifications.lastMarkUnreadIDs)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			notifications := &trackingNotificationsService{}
			configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{
				notifications: notifications,
			})
			hcc := newNotificationsHostCallContext("test-plugin-notifications-"+tc.name, "inbox", 42)
			hcc.hostServices[0].Methods = append(hcc.hostServices[0].Methods, tc.method)
			id := "9001"
			if tc.method == protocol.HostServiceMethodNotificationsMarkUnread {
				id = "9002"
			}
			response := invokeNotificationsDomainHostService(
				t,
				hcc,
				tc.method,
				marshalCapabilityJSONRequest(t, idsRequest{IDs: []string{id}}),
			)
			if response.Status != protocol.HostCallStatusSuccess {
				t.Fatalf("%s: expected success, got status=%d payload=%s", tc.method, response.Status, string(response.Payload))
			}
			tc.assertCall(notifications)
		})
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
