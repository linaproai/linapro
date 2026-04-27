-- ------------------------------------------------------------
-- 001-plugin-demo-dynamic-records.sql
-- plugin-demo-dynamic 示例记录表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `plugin_demo_dynamic_record` (
    `id`              VARCHAR(64) PRIMARY KEY COMMENT '记录ID',
    `title`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT '记录标题',
    `content`         VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '记录内容',
    `attachment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '附件原始文件名',
    `attachment_path` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '附件相对存储路径',
    `created_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) COMMENT='动态插件示例记录表';

INSERT IGNORE INTO `plugin_demo_dynamic_record` (
    `id`,
    `title`,
    `content`,
    `attachment_name`,
    `attachment_path`,
    `created_at`,
    `updated_at`
)
VALUES (
    'plugin-demo-dynamic-seed-record',
    'Dynamic Plugin SQL Demo Record',
    'This record is seeded by the plugin-demo-dynamic install SQL and demonstrates CRUD operations against the data table created during plugin installation.',
    '',
    '',
    '2026-04-16T00:00:00+08:00',
    '2026-04-16T00:00:00+08:00'
);
