// This file tests notify host service dispatch, default recipients, and authorization.

package wasm

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/internal/service/tenantcap"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

// trackingNotifyService records notify sends while returning deterministic
// output for shared-instance wiring tests.
type trackingNotifyService struct {
	sendCalls int
	lastInput notifysvc.SendInput
}

// InboxUnreadCount is unused by host-service notify tests.
func (s *trackingNotifyService) InboxUnreadCount(context.Context, int64) (int, error) {
	return 0, nil
}

// InboxList is unused by host-service notify tests.
func (s *trackingNotifyService) InboxList(context.Context, notifysvc.InboxListInput) (*notifysvc.InboxListOutput, error) {
	return &notifysvc.InboxListOutput{}, nil
}

// InboxMarkRead is unused by host-service notify tests.
func (s *trackingNotifyService) InboxMarkRead(context.Context, int64, int64) error { return nil }

// InboxMarkAllRead is unused by host-service notify tests.
func (s *trackingNotifyService) InboxMarkAllRead(context.Context, int64) error { return nil }

// InboxDelete is unused by host-service notify tests.
func (s *trackingNotifyService) InboxDelete(context.Context, int64, int64) error { return nil }

// InboxClear is unused by host-service notify tests.
func (s *trackingNotifyService) InboxClear(context.Context, int64) error { return nil }

// DeleteBySource is unused by host-service notify tests.
func (s *trackingNotifyService) DeleteBySource(context.Context, notifysvc.SourceType, []string) error {
	return nil
}

// Send records one host-service notify send request.
func (s *trackingNotifyService) Send(_ context.Context, in notifysvc.SendInput) (*notifysvc.SendOutput, error) {
	s.sendCalls++
	s.lastInput = in
	return &notifysvc.SendOutput{MessageID: 9001, DeliveryCount: len(in.RecipientUserIDs)}, nil
}

// SendNoticePublication is unused by host-service notify tests.
func (s *trackingNotifyService) SendNoticePublication(context.Context, notifysvc.NoticePublishInput) (*notifysvc.SendOutput, error) {
	return &notifysvc.SendOutput{}, nil
}

// createPluginNotifyTablesSQL provisions the notify tables required by the host
// service integration tests when they are absent in the test database.
const createPluginNotifyTablesSQL = `
CREATE TABLE IF NOT EXISTS sys_notify_channel (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    channel_key  VARCHAR(64) NOT NULL DEFAULT '',
    name         VARCHAR(128) NOT NULL DEFAULT '',
    channel_type VARCHAR(32) NOT NULL DEFAULT '',
    status       SMALLINT NOT NULL DEFAULT 1,
    config_json  TEXT NOT NULL,
    remark       VARCHAR(500) NOT NULL DEFAULT '',
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP NULL DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_notify_channel_channel_key ON sys_notify_channel (channel_key);

CREATE TABLE IF NOT EXISTS sys_notify_message (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plugin_id      VARCHAR(64) NOT NULL DEFAULT '',
    source_type    VARCHAR(32) NOT NULL DEFAULT '',
    source_id      VARCHAR(64) NOT NULL DEFAULT '',
    category_code  VARCHAR(32) NOT NULL DEFAULT '',
    title          VARCHAR(255) NOT NULL DEFAULT '',
    content        TEXT NOT NULL,
    payload_json   TEXT NOT NULL,
    sender_user_id BIGINT NOT NULL DEFAULT 0,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sys_notify_message_source ON sys_notify_message (source_type, source_id);

CREATE TABLE IF NOT EXISTS sys_notify_delivery (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message_id      BIGINT NOT NULL DEFAULT 0,
    channel_key     VARCHAR(64) NOT NULL DEFAULT '',
    channel_type    VARCHAR(32) NOT NULL DEFAULT '',
    recipient_type  VARCHAR(32) NOT NULL DEFAULT '',
    recipient_key   VARCHAR(128) NOT NULL DEFAULT '',
    user_id         BIGINT NOT NULL DEFAULT 0,
    delivery_status SMALLINT NOT NULL DEFAULT 0,
    is_read         SMALLINT NOT NULL DEFAULT 0,
    read_at         TIMESTAMP NULL DEFAULT NULL,
    error_message   VARCHAR(1000) NOT NULL DEFAULT '',
    sent_at         TIMESTAMP NULL DEFAULT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP NULL DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_message_id ON sys_notify_delivery (message_id);
CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_user_inbox ON sys_notify_delivery (user_id, channel_type, delivery_status, is_read);
CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_channel_status ON sys_notify_delivery (channel_key, delivery_status);
`

// TestHandleHostServiceInvokeNotifySendDefaultsToCurrentUser verifies notify
// sends default to the caller when no explicit recipients are provided.
func TestHandleHostServiceInvokeNotifySendDefaultsToCurrentUser(t *testing.T) {
	configureDefaultNotifyHostService(t)

	ctx := context.Background()
	ensurePluginNotifyTables(t, ctx)

	pluginID := "test-plugin-notify"
	cleanupPluginNotifyMessages(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginNotifyMessages(t, ctx, pluginID)
	})

	hcc := newNotifyHostCallContext(pluginID, "inbox", 1)
	response := invokeNotifyHostService(
		t,
		hcc,
		"inbox",
		pluginbridge.MarshalHostServiceNotifySendRequest(&pluginbridge.HostServiceNotifySendRequest{
			Title:        "同步完成",
			Content:      "订单同步已完成",
			SourceType:   "plugin",
			SourceID:     "job-1",
			CategoryCode: "other",
			PayloadJSON:  []byte(`{"scope":"orders"}`),
		}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("send: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}

	payload, err := pluginbridge.UnmarshalHostServiceNotifySendResponse(response.Payload)
	if err != nil {
		t.Fatalf("send payload decode failed: %v", err)
	}
	if payload.MessageID <= 0 || payload.DeliveryCount != 1 {
		t.Fatalf("send payload: got %#v", payload)
	}

	var message *entity.SysNotifyMessage
	if err = dao.SysNotifyMessage.Ctx(ctx).Where("id", payload.MessageID).Scan(&message); err != nil {
		t.Fatalf("query notify message failed: %v", err)
	}
	if message == nil || message.PluginId != pluginID || message.SourceId != "job-1" {
		t.Fatalf("notify message: got %#v", message)
	}

	var delivery *entity.SysNotifyDelivery
	if err = dao.SysNotifyDelivery.Ctx(ctx).Where("message_id", payload.MessageID).Scan(&delivery); err != nil {
		t.Fatalf("query notify delivery failed: %v", err)
	}
	if delivery == nil || delivery.UserId != 1 || delivery.ChannelKey != "inbox" {
		t.Fatalf("notify delivery: got %#v", delivery)
	}
}

// TestHandleHostServiceInvokeNotifyRejectsInvalidPayloadJSON verifies malformed
// payloadJson content is rejected before any persistence occurs.
func TestHandleHostServiceInvokeNotifyRejectsInvalidPayloadJSON(t *testing.T) {
	ctx := context.Background()
	ensurePluginNotifyTables(t, ctx)
	configureDefaultNotifyHostService(t)

	hcc := newNotifyHostCallContext("test-plugin-notify-invalid", "inbox", 1)
	response := invokeNotifyHostService(
		t,
		hcc,
		"inbox",
		pluginbridge.MarshalHostServiceNotifySendRequest(&pluginbridge.HostServiceNotifySendRequest{
			Title:       "同步完成",
			Content:     "订单同步已完成",
			PayloadJSON: []byte("{"),
		}),
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for malformed notify payloadJson, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeNotifyRejectsUnauthorizedChannel verifies plugins
// cannot send through channels outside their granted resources.
func TestHandleHostServiceInvokeNotifyRejectsUnauthorizedChannel(t *testing.T) {
	hcc := newNotifyHostCallContext("test-plugin-notify-denied", "inbox", 1)
	response := invokeNotifyHostService(
		t,
		hcc,
		"ops-webhook",
		pluginbridge.MarshalHostServiceNotifySendRequest(&pluginbridge.HostServiceNotifySendRequest{
			Title:   "同步完成",
			Content: "订单同步已完成",
		}),
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized notify channel, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeNotifyUsesConfiguredSharedService verifies notify
// host service dispatch reuses the explicitly configured shared instance.
func TestHandleHostServiceInvokeNotifyUsesConfiguredSharedService(t *testing.T) {
	notifySvc := &trackingNotifyService{}
	if err := ConfigureNotifyHostService(notifySvc); err != nil {
		t.Fatalf("configure notify host service failed: %v", err)
	}
	t.Cleanup(func() {
		notifyHostService = nil
	})

	hcc := newNotifyHostCallContext("test-plugin-notify-shared", "inbox", 42)
	response := invokeNotifyHostService(
		t,
		hcc,
		"inbox",
		pluginbridge.MarshalHostServiceNotifySendRequest(&pluginbridge.HostServiceNotifySendRequest{
			Title:        "done",
			Content:      "finished",
			SourceType:   "plugin",
			SourceID:     "job-2",
			CategoryCode: "other",
		}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("send through shared notify: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := pluginbridge.UnmarshalHostServiceNotifySendResponse(response.Payload)
	if err != nil {
		t.Fatalf("decode shared notify response: %v", err)
	}
	if payload.MessageID != 9001 || payload.DeliveryCount != 1 {
		t.Fatalf("expected shared notify output, got %#v", payload)
	}
	if notifySvc.sendCalls != 1 {
		t.Fatalf("expected shared notify service to receive one send, got %d", notifySvc.sendCalls)
	}
	if notifySvc.lastInput.PluginID != hcc.pluginID || len(notifySvc.lastInput.RecipientUserIDs) != 1 || notifySvc.lastInput.RecipientUserIDs[0] != 42 {
		t.Fatalf("expected plugin and default recipient to be forwarded, got %#v", notifySvc.lastInput)
	}
}

// configureDefaultNotifyHostService wires tests to the same notify-backed host
// service shape used by startup and restores the previous package state.
func configureDefaultNotifyHostService(t *testing.T) {
	t.Helper()
	previousNotifySvc := notifyHostService
	if err := ConfigureNotifyHostService(notifysvc.New(tenantcap.New(nil, nil))); err != nil {
		t.Fatalf("configure notify host service failed: %v", err)
	}
	t.Cleanup(func() {
		notifyHostService = previousNotifySvc
	})
}

// ensurePluginNotifyTables creates the notify schema and seeds the inbox
// channel used by the notify host-service tests.
func ensurePluginNotifyTables(t *testing.T, ctx context.Context) {
	t.Helper()
	for _, statement := range dialect.SplitSQLStatements(createPluginNotifyTablesSQL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("expected notify tables to be created, got error: %v\nSQL:\n%s", err, statement)
		}
	}
	if _, err := g.DB().Exec(ctx, `
INSERT INTO sys_notify_channel (
    channel_key, name, channel_type, status, config_json, remark, created_at, updated_at, deleted_at
) VALUES (
    'inbox', '站内信', 'inbox', 1, '{}', '系统内置站内信通道', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL
)
ON CONFLICT DO NOTHING
`); err != nil {
		t.Fatalf("expected inbox channel seed to insert idempotently, got error: %v", err)
	}
}

// cleanupPluginNotifyMessages removes notify messages and dependent deliveries
// created for one plugin so tests stay isolated across reruns.
func cleanupPluginNotifyMessages(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := g.DB().Exec(ctx, `
DELETE FROM sys_notify_delivery
WHERE message_id IN (SELECT id FROM sys_notify_message WHERE plugin_id = ?)
`, pluginID); err != nil {
		t.Fatalf("failed to delete notify deliveries for %s: %v", pluginID, err)
	}
	if _, err := g.DB().Exec(ctx, `DELETE FROM sys_notify_message WHERE plugin_id = ?`, pluginID); err != nil {
		t.Fatalf("failed to delete notify messages for %s: %v", pluginID, err)
	}
}

// newNotifyHostCallContext constructs a notify-capable host call context for
// one authorized channel and caller identity snapshot.
func newNotifyHostCallContext(pluginID string, channelKey string, userID int32) *hostCallContext {
	return &hostCallContext{
		pluginID: pluginID,
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityNotify: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceNotify,
			Methods: []string{pluginbridge.HostServiceMethodNotifySend},
			Resources: []*pluginbridge.HostServiceResourceSpec{
				{Ref: channelKey},
			},
		}},
		identity: &pluginbridge.IdentitySnapshotV1{UserID: userID},
	}
}

// invokeNotifyHostService routes a notify host-service request through the
// shared dispatcher and returns its raw response envelope.
func invokeNotifyHostService(
	t *testing.T,
	hcc *hostCallContext,
	channelKey string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceNotify,
		Method:      pluginbridge.HostServiceMethodNotifySend,
		ResourceRef: channelKey,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		hcc,
		pluginbridge.MarshalHostServiceRequestEnvelope(request),
	)
}
