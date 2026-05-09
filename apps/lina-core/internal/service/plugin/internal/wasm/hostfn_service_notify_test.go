// This file tests notify host service dispatch, default recipients, and authorization.

package wasm

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

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
