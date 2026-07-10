// This file verifies guest-side notification host-service clients call the
// published dynamic notification methods.

package domainhostcall

import (
	"encoding/json"
	"testing"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// notificationsHostCallRecorder records guest notification host-service calls.
type notificationsHostCallRecorder struct {
	service string
	method  string
	ids     idsRequest
	list    notifycap.ListInput
	source  notifycap.BatchGetBySourceInput
}

// invoke decodes notification JSON requests and writes deterministic list output.
func (r *notificationsHostCallRecorder) invoke(service string, method string, request []byte, out any) error {
	r.service = service
	r.method = method
	if len(request) > 0 {
		envelope, err := protocol.UnmarshalHostServiceJSONRequest(request)
		if err != nil {
			return err
		}
		switch method {
		case protocol.HostServiceMethodNotificationsList:
			if err = json.Unmarshal(envelope.Value, &r.list); err != nil {
				return err
			}
		case protocol.HostServiceMethodNotificationsDeleteBySource:
			if err = json.Unmarshal(envelope.Value, &r.source); err != nil {
				return err
			}
		default:
			if err = json.Unmarshal(envelope.Value, &r.ids); err != nil {
				return err
			}
		}
	}
	if page, ok := out.(*capmodel.PageResult[*notifycap.MessageInfo]); ok {
		*page = capmodel.PageResult[*notifycap.MessageInfo]{
			Items: []*notifycap.MessageInfo{{ID: "9001", SourceID: "job-1", Title: "sync done"}},
			Total: 1,
		}
	}
	return nil
}

// TestNotificationsListCallsPublishedHostService verifies List uses messages.list.
func TestNotificationsListCallsPublishedHostService(t *testing.T) {
	recorder := &notificationsHostCallRecorder{}
	service := Notifications(recorder.invoke, nil)

	page, err := service.List(t.Context(), notifycap.ListInput{
		SourceType: usermsgv1.SourceTypePlugin,
		SourceID:   "job-1",
		Page:       capmodel.PageRequest{PageNum: 1, PageSize: 10},
	})
	if err != nil {
		t.Fatalf("list through host service: %v", err)
	}
	if page == nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %#v", page)
	}
	if recorder.service != protocol.HostServiceNotifications || recorder.method != protocol.HostServiceMethodNotificationsList {
		t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
	}
	if recorder.list.SourceID != "job-1" || recorder.list.Page.PageSize != 10 {
		t.Fatalf("unexpected list request: %#v", recorder.list)
	}
}

// TestNotificationsDeleteCallsPublishedHostService verifies Delete uses messages.delete.
func TestNotificationsDeleteCallsPublishedHostService(t *testing.T) {
	recorder := &notificationsHostCallRecorder{}
	service := Notifications(recorder.invoke, nil)

	if err := service.Delete(t.Context(), []notifycap.MessageID{"9001"}); err != nil {
		t.Fatalf("delete through host service: %v", err)
	}
	if recorder.service != protocol.HostServiceNotifications || recorder.method != protocol.HostServiceMethodNotificationsDelete {
		t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
	}
	if len(recorder.ids.IDs) != 1 || recorder.ids.IDs[0] != "9001" {
		t.Fatalf("unexpected delete request: %#v", recorder.ids)
	}
}

// TestNotificationsDeleteBySourceCallsPublishedHostService verifies source
// deletes use messages.by_source.delete.
func TestNotificationsDeleteBySourceCallsPublishedHostService(t *testing.T) {
	recorder := &notificationsHostCallRecorder{}
	service := Notifications(recorder.invoke, nil)

	if err := service.DeleteBySource(t.Context(), usermsgv1.SourceTypePlugin, []string{"job-1", "job-2"}); err != nil {
		t.Fatalf("delete by source through host service: %v", err)
	}
	if recorder.service != protocol.HostServiceNotifications || recorder.method != protocol.HostServiceMethodNotificationsDeleteBySource {
		t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
	}
	if recorder.source.SourceType != usermsgv1.SourceTypePlugin || len(recorder.source.SourceIDs) != 2 {
		t.Fatalf("unexpected source delete request: %#v", recorder.source)
	}
}

// TestNotificationsReadStateCallsPublishedHostService verifies read-state calls
// use messages.mark_read and messages.mark_unread.
func TestNotificationsReadStateCallsPublishedHostService(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(notifycap.Service) error
		method string
		id     string
	}{
		{
			name: "mark_read",
			call: func(service notifycap.Service) error {
				return service.MarkRead(t.Context(), []notifycap.MessageID{"9001"})
			},
			method: protocol.HostServiceMethodNotificationsMarkRead,
			id:     "9001",
		},
		{
			name: "mark_unread",
			call: func(service notifycap.Service) error {
				return service.MarkUnread(t.Context(), []notifycap.MessageID{"9002"})
			},
			method: protocol.HostServiceMethodNotificationsMarkUnread,
			id:     "9002",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &notificationsHostCallRecorder{}
			service := Notifications(recorder.invoke, nil)

			if err := tc.call(service); err != nil {
				t.Fatalf("%s through host service: %v", tc.name, err)
			}
			if recorder.service != protocol.HostServiceNotifications || recorder.method != tc.method {
				t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
			}
			if len(recorder.ids.IDs) != 1 || recorder.ids.IDs[0] != tc.id {
				t.Fatalf("unexpected read-state request: %#v", recorder.ids)
			}
		})
	}
}
