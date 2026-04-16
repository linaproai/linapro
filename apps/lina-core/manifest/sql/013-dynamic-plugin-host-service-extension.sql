-- ------------------------------------------------------------
-- 013-dynamic-plugin-host-service-extension.sql
-- 动态插件宿主服务扩展：KV缓存与统一通知域
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `sys_kv_cache` (
    `id`          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `owner_type`  VARCHAR(16)   NOT NULL DEFAULT '' COMMENT '所属类型：plugin=动态插件 module=宿主模块',
    `owner_key`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '所属标识：插件ID或模块名',
    `namespace`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '缓存命名空间，对应 host-cache 资源标识',
    `cache_key`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '缓存键',
    `value_kind`  TINYINT       NOT NULL DEFAULT 1  COMMENT '值类型：1=字符串 2=整数',
    `value_bytes` VARBINARY(4096) NOT NULL          COMMENT '缓存字节值，供 get/set 使用',
    `value_int`   BIGINT        NOT NULL DEFAULT 0  COMMENT '缓存整数值，供 incr 使用',
    `expire_at`   DATETIME      NULL DEFAULT NULL   COMMENT '过期时间，NULL表示永不过期',
    `created_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_owner_namespace_key` (`owner_type`, `owner_key`, `namespace`, `cache_key`),
    KEY `idx_expire_at` (`expire_at`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='宿主分布式KV缓存表';

CREATE TABLE IF NOT EXISTS `sys_notify_channel` (
    `id`           BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `channel_key`  VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '通道标识',
    `name`         VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '通道名称',
    `channel_type` VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '通道类型：inbox=站内信 email=邮件 webhook=Webhook',
    `status`       TINYINT       NOT NULL DEFAULT 1  COMMENT '状态：1=启用 0=停用',
    `config_json`  LONGTEXT      NOT NULL             COMMENT '通道配置JSON',
    `remark`       VARCHAR(500)  NOT NULL DEFAULT '' COMMENT '备注',
    `created_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME      NULL DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY `uk_channel_key` (`channel_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知通道表';

CREATE TABLE IF NOT EXISTS `sys_notify_message` (
    `id`             BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `plugin_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '来源插件ID，宿主内建流程为空',
    `source_type`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '来源类型：notice=公告 plugin=插件 system=系统',
    `source_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '来源业务ID',
    `category_code`  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '消息分类：notice=通知 announcement=公告 other=其他',
    `title`          VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '消息标题',
    `content`        LONGTEXT      NOT NULL             COMMENT '消息正文',
    `payload_json`   LONGTEXT      NOT NULL             COMMENT '扩展载荷JSON',
    `sender_user_id` BIGINT        NOT NULL DEFAULT 0  COMMENT '发送者用户ID',
    `created_at`     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY `idx_source` (`source_type`, `source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知消息主表';

CREATE TABLE IF NOT EXISTS `sys_notify_delivery` (
    `id`              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `message_id`      BIGINT        NOT NULL DEFAULT 0  COMMENT '通知消息ID',
    `channel_key`     VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '投递通道标识',
    `channel_type`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '投递通道类型',
    `recipient_type`  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '接收者类型：user=用户 email=邮箱 webhook=Webhook',
    `recipient_key`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '接收者标识，如用户ID邮箱地址或Webhook标识',
    `user_id`         BIGINT        NOT NULL DEFAULT 0  COMMENT '站内信用户ID，非站内信时为0',
    `delivery_status` TINYINT       NOT NULL DEFAULT 0  COMMENT '投递状态：0=待发送 1=成功 2=失败',
    `is_read`         TINYINT       NOT NULL DEFAULT 0  COMMENT '是否已读：0=未读 1=已读',
    `read_at`         DATETIME      NULL DEFAULT NULL COMMENT '已读时间',
    `error_message`   VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '失败原因',
    `sent_at`         DATETIME      NULL DEFAULT NULL COMMENT '发送完成时间',
    `created_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`      DATETIME      NULL DEFAULT NULL COMMENT '删除时间',
    KEY `idx_message_id` (`message_id`),
    KEY `idx_user_inbox` (`user_id`, `channel_type`, `delivery_status`, `is_read`),
    KEY `idx_channel_status` (`channel_key`, `delivery_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知投递记录表';

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
