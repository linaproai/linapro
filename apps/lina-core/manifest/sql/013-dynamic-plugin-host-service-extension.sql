-- ------------------------------------------------------------
-- 013 dynamic plugin host service extension SQL file
-- 013 动态插件宿主服务扩展 SQL 文件
-- Dynamic plugin host service extension: KV cache and unified notification domain
-- 动态插件宿主服务扩展：KV缓存与统一通知域
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `sys_kv_cache` (
    `id`          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    `owner_type`  VARCHAR(16)   NOT NULL DEFAULT '' COMMENT  'Owner type: plugin=dynamic plugin, module=host module',
    `owner_key`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Owner key: plugin ID or module name',
    `namespace`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Cache namespace mapped to the host-cache resource identifier',
    `cache_key`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT  'Cache key',
    `value_kind`  TINYINT       NOT NULL DEFAULT 1  COMMENT  'Value type: 1=string, 2=integer',
    `value_bytes` VARBINARY(4096) NOT NULL          COMMENT  'Cache byte value used by get/set',
    `value_int`   BIGINT        NOT NULL DEFAULT 0  COMMENT  'Cache integer value used by incr',
    `expire_at`   DATETIME      NULL DEFAULT NULL   COMMENT  'Expiration time, NULL means never expires',
    `created_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time',
    UNIQUE KEY `uk_owner_namespace_key` (`owner_type`, `owner_key`, `namespace`, `cache_key`),
    KEY `idx_expire_at` (`expire_at`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Host distributed KV cache table';

CREATE TABLE IF NOT EXISTS `sys_notify_channel` (
    `id`           BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    `channel_key`  VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Channel key',
    `name`         VARCHAR(128)  NOT NULL DEFAULT '' COMMENT  'Channel name',
    `channel_type` VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Channel type: inbox=in-app message, email=email, webhook=webhook',
    `status`       TINYINT       NOT NULL DEFAULT 1  COMMENT  'Status: 1=enabled, 0=disabled',
    `config_json`  LONGTEXT      NOT NULL             COMMENT  'Channel configuration JSON',
    `remark`       VARCHAR(500)  NOT NULL DEFAULT '' COMMENT  'Remark',
    `created_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time',
    `deleted_at`   DATETIME      NULL DEFAULT NULL COMMENT  'Deletion time',
    UNIQUE KEY `uk_channel_key` (`channel_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Notification channel table';

CREATE TABLE IF NOT EXISTS `sys_notify_message` (
    `id`             BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    `plugin_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Source plugin ID, empty for host built-in flows',
    `source_type`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Source type: notice=notice, plugin=plugin, system=system',
    `source_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Source business ID',
    `category_code`  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Message category: notice=notification, announcement=announcement, other=other',
    `title`          VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Message title',
    `content`        LONGTEXT      NOT NULL             COMMENT  'Message body',
    `payload_json`   LONGTEXT      NOT NULL             COMMENT  'Extended payload JSON',
    `sender_user_id` BIGINT        NOT NULL DEFAULT 0  COMMENT  'Sender user ID',
    `created_at`     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    KEY `idx_source` (`source_type`, `source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Notification message table';

CREATE TABLE IF NOT EXISTS `sys_notify_delivery` (
    `id`              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    `message_id`      BIGINT        NOT NULL DEFAULT 0  COMMENT  'Notification message ID',
    `channel_key`     VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Delivery channel key',
    `channel_type`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Delivery channel type',
    `recipient_type`  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Recipient type: user=user, email=email, webhook=webhook',
    `recipient_key`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT  'Recipient key such as user ID, email address, or webhook identifier',
    `user_id`         BIGINT        NOT NULL DEFAULT 0  COMMENT  'In-app message user ID, 0 for non-in-app delivery',
    `delivery_status` TINYINT       NOT NULL DEFAULT 0  COMMENT  'Delivery status: 0=pending, 1=succeeded, 2=failed',
    `is_read`         TINYINT       NOT NULL DEFAULT 0  COMMENT  'Read flag: 0=unread, 1=read',
    `read_at`         DATETIME      NULL DEFAULT NULL COMMENT  'Read time',
    `error_message`   VARCHAR(1000) NOT NULL DEFAULT '' COMMENT  'Failure reason',
    `sent_at`         DATETIME      NULL DEFAULT NULL COMMENT  'Send completion time',
    `created_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time',
    `deleted_at`      DATETIME      NULL DEFAULT NULL COMMENT  'Deletion time',
    KEY `idx_message_id` (`message_id`),
    KEY `idx_user_inbox` (`user_id`, `channel_type`, `delivery_status`, `is_read`),
    KEY `idx_channel_status` (`channel_key`, `delivery_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Notification delivery record table';

INSERT IGNORE INTO `sys_notify_channel` (
    `channel_key`,
    `name`,
    `channel_type`,
    `status`,
    `config_json`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    'inbox',
    '站内信',
    'inbox',
    1,
    '{}',
    '系统内置站内信通道',
    NOW(),
    NOW()
);
