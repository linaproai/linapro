-- ------------------------------------------------------------
-- 001 plugin-demo-dynamic records SQL file
-- 001 plugin-demo-dynamic 记录 SQL 文件
-- plugin-demo-dynamic demo record table
-- plugin-demo-dynamic 示例记录表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `plugin_demo_dynamic_record` (
    `id`              VARCHAR(64) PRIMARY KEY COMMENT  'Record ID',
    `title`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Record title',
    `content`         VARCHAR(1000) NOT NULL DEFAULT '' COMMENT  'Record content',
    `attachment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Original attachment file name',
    `attachment_path` VARCHAR(500) NOT NULL DEFAULT '' COMMENT  'Relative attachment storage path',
    `created_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time'
) COMMENT= 'Dynamic plugin demo record table';

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
    'plugin-demo-dynamic-mock-record',
    'Dynamic Plugin SQL Demo Record',
    'This record is seeded by the plugin-demo-dynamic install SQL and demonstrates CRUD operations against the data table created during plugin installation.',
    '',
    '',
    '2026-04-16 09:00:00',
    '2026-04-16 09:00:00'
);
