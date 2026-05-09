-- ------------------------------------------------------------
-- 013 dynamic plugin host service extension SQL file
-- 013 动态插件宿主服务扩展 SQL 文件
-- Dynamic plugin host service extension: KV cache and unified notification domain
-- 动态插件宿主服务扩展：KV缓存与统一通知域
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_kv_cache (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "owner_type"  VARCHAR(16) NOT NULL DEFAULT '',
    "owner_key"   VARCHAR(64) NOT NULL DEFAULT '',
    "namespace"   VARCHAR(64) NOT NULL DEFAULT '',
    "cache_key"   VARCHAR(128) NOT NULL DEFAULT '',
    "value_kind"  SMALLINT NOT NULL DEFAULT 1,
    "value_bytes" BYTEA NOT NULL,
    "value_int"   BIGINT NOT NULL DEFAULT 0,
    "expire_at"   TIMESTAMP NULL DEFAULT NULL,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE sys_kv_cache IS 'Host distributed KV cache table';
COMMENT ON COLUMN sys_kv_cache."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_kv_cache."owner_type" IS 'Owner type: plugin=dynamic plugin, module=host module';
COMMENT ON COLUMN sys_kv_cache."owner_key" IS 'Owner key: plugin ID or module name';
COMMENT ON COLUMN sys_kv_cache."namespace" IS 'Cache namespace mapped to the host-cache resource identifier';
COMMENT ON COLUMN sys_kv_cache."cache_key" IS 'Cache key';
COMMENT ON COLUMN sys_kv_cache."value_kind" IS 'Value type: 1=string, 2=integer';
COMMENT ON COLUMN sys_kv_cache."value_bytes" IS 'Cache byte value used by get/set';
COMMENT ON COLUMN sys_kv_cache."value_int" IS 'Cache integer value used by incr';
COMMENT ON COLUMN sys_kv_cache."expire_at" IS 'Expiration time, NULL means never expires';
COMMENT ON COLUMN sys_kv_cache."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_kv_cache."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_kv_cache_owner_namespace_key ON sys_kv_cache ("owner_type", "owner_key", "namespace", "cache_key");
CREATE INDEX IF NOT EXISTS idx_sys_kv_cache_expire_at ON sys_kv_cache ("expire_at");

CREATE TABLE IF NOT EXISTS sys_notify_channel (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "channel_key"  VARCHAR(64) NOT NULL DEFAULT '',
    "name"         VARCHAR(128) NOT NULL DEFAULT '',
    "channel_type" VARCHAR(32) NOT NULL DEFAULT '',
    "status"       SMALLINT NOT NULL DEFAULT 1,
    "config_json"  TEXT NOT NULL,
    "remark"       VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP NULL DEFAULT NULL
);

COMMENT ON TABLE sys_notify_channel IS 'Notification channel table';
COMMENT ON COLUMN sys_notify_channel."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_notify_channel."channel_key" IS 'Channel key';
COMMENT ON COLUMN sys_notify_channel."name" IS 'Channel name';
COMMENT ON COLUMN sys_notify_channel."channel_type" IS 'Channel type: inbox=in-app message, email=email, webhook=webhook';
COMMENT ON COLUMN sys_notify_channel."status" IS 'Status: 1=enabled, 0=disabled';
COMMENT ON COLUMN sys_notify_channel."config_json" IS 'Channel configuration JSON';
COMMENT ON COLUMN sys_notify_channel."remark" IS 'Remark';
COMMENT ON COLUMN sys_notify_channel."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_notify_channel."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_notify_channel."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_notify_channel_channel_key ON sys_notify_channel ("channel_key");

CREATE TABLE IF NOT EXISTS sys_notify_message (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "source_type"    VARCHAR(32) NOT NULL DEFAULT '',
    "source_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "category_code"  VARCHAR(32) NOT NULL DEFAULT '',
    "title"          VARCHAR(255) NOT NULL DEFAULT '',
    "content"        TEXT NOT NULL,
    "payload_json"   TEXT NOT NULL,
    "sender_user_id" BIGINT NOT NULL DEFAULT 0,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE sys_notify_message IS 'Notification message table';
COMMENT ON COLUMN sys_notify_message."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_notify_message."plugin_id" IS 'Source plugin ID, empty for host built-in flows';
COMMENT ON COLUMN sys_notify_message."source_type" IS 'Source type: notice=notice, plugin=plugin, system=system';
COMMENT ON COLUMN sys_notify_message."source_id" IS 'Source business ID';
COMMENT ON COLUMN sys_notify_message."category_code" IS 'Message category: notice=notification, announcement=announcement, other=other';
COMMENT ON COLUMN sys_notify_message."title" IS 'Message title';
COMMENT ON COLUMN sys_notify_message."content" IS 'Message body';
COMMENT ON COLUMN sys_notify_message."payload_json" IS 'Extended payload JSON';
COMMENT ON COLUMN sys_notify_message."sender_user_id" IS 'Sender user ID';
COMMENT ON COLUMN sys_notify_message."created_at" IS 'Creation time';

CREATE INDEX IF NOT EXISTS idx_sys_notify_message_source ON sys_notify_message ("source_type", "source_id");

CREATE TABLE IF NOT EXISTS sys_notify_delivery (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "message_id"      BIGINT NOT NULL DEFAULT 0,
    "channel_key"     VARCHAR(64) NOT NULL DEFAULT '',
    "channel_type"    VARCHAR(32) NOT NULL DEFAULT '',
    "recipient_type"  VARCHAR(32) NOT NULL DEFAULT '',
    "recipient_key"   VARCHAR(128) NOT NULL DEFAULT '',
    "user_id"         BIGINT NOT NULL DEFAULT 0,
    "delivery_status" SMALLINT NOT NULL DEFAULT 0,
    "is_read"         SMALLINT NOT NULL DEFAULT 0,
    "read_at"         TIMESTAMP NULL DEFAULT NULL,
    "error_message"   VARCHAR(1000) NOT NULL DEFAULT '',
    "sent_at"         TIMESTAMP NULL DEFAULT NULL,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP NULL DEFAULT NULL
);

COMMENT ON TABLE sys_notify_delivery IS 'Notification delivery record table';
COMMENT ON COLUMN sys_notify_delivery."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_notify_delivery."message_id" IS 'Notification message ID';
COMMENT ON COLUMN sys_notify_delivery."channel_key" IS 'Delivery channel key';
COMMENT ON COLUMN sys_notify_delivery."channel_type" IS 'Delivery channel type';
COMMENT ON COLUMN sys_notify_delivery."recipient_type" IS 'Recipient type: user=user, email=email, webhook=webhook';
COMMENT ON COLUMN sys_notify_delivery."recipient_key" IS 'Recipient key such as user ID, email address, or webhook identifier';
COMMENT ON COLUMN sys_notify_delivery."user_id" IS 'In-app message user ID, 0 for non-in-app delivery';
COMMENT ON COLUMN sys_notify_delivery."delivery_status" IS 'Delivery status: 0=pending, 1=succeeded, 2=failed';
COMMENT ON COLUMN sys_notify_delivery."is_read" IS 'Read flag: 0=unread, 1=read';
COMMENT ON COLUMN sys_notify_delivery."read_at" IS 'Read time';
COMMENT ON COLUMN sys_notify_delivery."error_message" IS 'Failure reason';
COMMENT ON COLUMN sys_notify_delivery."sent_at" IS 'Send completion time';
COMMENT ON COLUMN sys_notify_delivery."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_notify_delivery."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_notify_delivery."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_message_id ON sys_notify_delivery ("message_id");
CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_user_inbox ON sys_notify_delivery ("user_id", "channel_type", "delivery_status", "is_read");
CREATE INDEX IF NOT EXISTS idx_sys_notify_delivery_channel_status ON sys_notify_delivery ("channel_key", "delivery_status");

INSERT INTO sys_notify_channel (
    "channel_key",
    "name",
    "channel_type",
    "status",
    "config_json",
    "remark",
    "created_at",
    "updated_at"
) VALUES (
    'inbox',
    '站内信',
    'inbox',
    1,
    '{}',
    '系统内置站内信通道',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;
