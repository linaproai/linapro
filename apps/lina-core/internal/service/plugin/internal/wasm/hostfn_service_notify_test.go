// This file tests notify host service dispatch, default recipients, and authorization.

package wasm

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginbridge"
)

const createPluginNotifyTablesSQL = `
CREATE TABLE IF NOT EXISTS sys_notify_channel (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    channel_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通道标识',
    name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通道名称',
    channel_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通道类型：inbox=站内信 email=邮件 webhook=Webhook',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=停用',
    config_json LONGTEXT NOT NULL COMMENT '通道配置JSON',
    remark VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY uk_channel_key (channel_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知通道表';

CREATE TABLE IF NOT EXISTS sys_notify_message (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源插件ID，宿主内建流程为空',
    source_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型：notice=公告 plugin=插件 system=系统',
    source_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
    category_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '消息分类：notice=通知 announcement=公告 other=其他',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息标题',
    content LONGTEXT NOT NULL COMMENT '消息正文',
    payload_json LONGTEXT NOT NULL COMMENT '扩展载荷JSON',
    sender_user_id BIGINT NOT NULL DEFAULT 0 COMMENT '发送者用户ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY idx_source (source_type, source_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知消息主表';

CREATE TABLE IF NOT EXISTS sys_notify_delivery (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    message_id BIGINT NOT NULL DEFAULT 0 COMMENT '通知消息ID',
    channel_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '投递通道标识',
    channel_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '投递通道类型',
    recipient_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '接收者类型：user=用户 email=邮箱 webhook=Webhook',
    recipient_key VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接收者标识，如用户ID邮箱地址或Webhook标识',
    user_id BIGINT NOT NULL DEFAULT 0 COMMENT '站内信用户ID，非站内信时为0',
    delivery_status TINYINT NOT NULL DEFAULT 0 COMMENT '投递状态：0=待发送 1=成功 2=失败',
    is_read TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0=未读 1=已读',
    read_at DATETIME NULL DEFAULT NULL COMMENT '已读时间',
    error_message VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '失败原因',
    sent_at DATETIME NULL DEFAULT NULL COMMENT '发送完成时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    KEY idx_message_id (message_id),
    KEY idx_user_inbox (user_id, channel_type, delivery_status, is_read),
    KEY idx_channel_status (channel_key, delivery_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知投递记录表';
`

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

func ensurePluginNotifyTables(t *testing.T, ctx context.Context) {
	t.Helper()
	if _, err := g.DB().Exec(ctx, createPluginNotifyTablesSQL); err != nil {
		t.Fatalf("expected notify tables to be created, got error: %v", err)
	}
	if _, err := g.DB().Exec(ctx, `
INSERT INTO sys_notify_channel (
    channel_key, name, channel_type, status, config_json, remark, created_at, updated_at, deleted_at
) VALUES (
    'inbox', '站内信', 'inbox', 1, '{}', '系统内置站内信通道', NOW(), NOW(), NULL
)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    channel_type = VALUES(channel_type),
    status = VALUES(status),
    config_json = VALUES(config_json),
    remark = VALUES(remark),
    deleted_at = NULL
`); err != nil {
		t.Fatalf("expected inbox channel seed to upsert, got error: %v", err)
	}
}

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
