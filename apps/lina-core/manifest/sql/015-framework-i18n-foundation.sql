-- 015: Framework i18n foundation
-- 包含：语言注册表、消息翻译覆写表、业务内容多语言表

-- ============================================================
-- 语言注册表
-- ============================================================
CREATE TABLE IF NOT EXISTS `sys_i18n_locale` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '语言ID',
    `locale`      VARCHAR(32)     NOT NULL DEFAULT ''     COMMENT '语言编码，如 zh-CN、en-US',
    `name`        VARCHAR(128)    NOT NULL DEFAULT ''     COMMENT '语言名称默认展示值',
    `native_name` VARCHAR(128)    NOT NULL DEFAULT ''     COMMENT '语言原生名称',
    `sort`        INT             NOT NULL DEFAULT 0      COMMENT '显示排序',
    `status`      TINYINT         NOT NULL DEFAULT 1      COMMENT '状态（0=停用 1=启用）',
    `is_default`  TINYINT         NOT NULL DEFAULT 0      COMMENT '是否默认语言（0=否 1=是）',
    `remark`      VARCHAR(500)    NOT NULL DEFAULT ''     COMMENT '备注',
    `created_at`  DATETIME        DEFAULT NULL            COMMENT '创建时间',
    `updated_at`  DATETIME        DEFAULT NULL            COMMENT '更新时间',
    `deleted_at`  DATETIME        DEFAULT NULL            COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_locale` (`locale`),
    KEY `idx_status_sort` (`status`, `sort`, `locale`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='国际化语言注册表';

-- ============================================================
-- 消息翻译覆写表
-- ============================================================
CREATE TABLE IF NOT EXISTS `sys_i18n_message` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
    `locale`        VARCHAR(32)     NOT NULL DEFAULT ''     COMMENT '语言编码',
    `message_key`   VARCHAR(255)    NOT NULL DEFAULT ''     COMMENT '翻译键',
    `message_value` LONGTEXT        NOT NULL                COMMENT '翻译值',
    `scope_type`    VARCHAR(32)     NOT NULL DEFAULT 'host' COMMENT '作用域类型（host/project/plugin/business）',
    `scope_key`     VARCHAR(128)    NOT NULL DEFAULT ''     COMMENT '作用域标识，如 core、plugin_id、project_code',
    `source_type`   VARCHAR(32)     NOT NULL DEFAULT 'manual' COMMENT '来源类型（manual/import/sync）',
    `status`        TINYINT         NOT NULL DEFAULT 1      COMMENT '状态（0=停用 1=启用）',
    `remark`        VARCHAR(500)    NOT NULL DEFAULT ''     COMMENT '备注',
    `created_at`    DATETIME        DEFAULT NULL            COMMENT '创建时间',
    `updated_at`    DATETIME        DEFAULT NULL            COMMENT '更新时间',
    `deleted_at`    DATETIME        DEFAULT NULL            COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_locale_message_scope` (`locale`, `message_key`, `scope_type`, `scope_key`),
    KEY `idx_locale_status_key` (`locale`, `status`, `message_key`),
    KEY `idx_scope_status` (`scope_type`, `scope_key`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='国际化消息翻译表';

-- ============================================================
-- 业务内容多语言表
-- ============================================================
CREATE TABLE IF NOT EXISTS `sys_i18n_content` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '内容ID',
    `business_type` VARCHAR(64)     NOT NULL DEFAULT ''     COMMENT '业务类型',
    `business_id`   VARCHAR(128)    NOT NULL DEFAULT ''     COMMENT '业务主键或稳定业务标识',
    `field`         VARCHAR(64)     NOT NULL DEFAULT ''     COMMENT '业务字段名',
    `locale`        VARCHAR(32)     NOT NULL DEFAULT ''     COMMENT '语言编码',
    `content_type`  VARCHAR(32)     NOT NULL DEFAULT 'plain' COMMENT '内容类型（plain/markdown/html/json）',
    `content`       LONGTEXT        NOT NULL                COMMENT '多语言内容值',
    `status`        TINYINT         NOT NULL DEFAULT 1      COMMENT '状态（0=停用 1=启用）',
    `remark`        VARCHAR(500)    NOT NULL DEFAULT ''     COMMENT '备注',
    `created_at`    DATETIME        DEFAULT NULL            COMMENT '创建时间',
    `updated_at`    DATETIME        DEFAULT NULL            COMMENT '更新时间',
    `deleted_at`    DATETIME        DEFAULT NULL            COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_business_field_locale` (`business_type`, `business_id`, `field`, `locale`),
    KEY `idx_business_lookup` (`business_type`, `business_id`, `field`, `status`),
    KEY `idx_locale_status` (`locale`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务内容多语言表';

-- ============================================================
-- 初始语言数据
-- ============================================================
INSERT IGNORE INTO `sys_i18n_locale` (`locale`, `name`, `native_name`, `sort`, `status`, `is_default`, `remark`, `created_at`, `updated_at`) VALUES
('zh-CN', '简体中文', '简体中文', 1, 1, 1, '宿主默认语言', NOW(), NOW()),
('en-US', '英语', 'English', 2, 1, 0, '宿主内置英语语言包', NOW(), NOW());
