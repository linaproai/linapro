-- ------------------------------------------------------------
-- 001 plugin-demo-source records SQL file
-- 001 plugin-demo-source 记录 SQL 文件
-- plugin-demo-source demo record table
-- plugin-demo-source 示例记录表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `plugin_demo_source_record` (
    `id`              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    `title`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Record title',
    `content`         VARCHAR(1000) NOT NULL DEFAULT '' COMMENT  'Record content',
    `attachment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Original attachment file name',
    `attachment_path` VARCHAR(500) NOT NULL DEFAULT '' COMMENT  'Relative attachment storage path',
    `created_at`      DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at`      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Source plugin demo record table';

INSERT IGNORE INTO `plugin_demo_source_record` (
    `title`,
    `content`,
    `attachment_name`,
    `attachment_path`,
    `created_at`,
    `updated_at`
)
SELECT
    '源码插件 SQL 示例记录',
    '该记录由 plugin-demo-source 安装 SQL 初始化，用于演示源码插件页面如何对插件自有数据表执行增删查改操作。',
    '',
    '',
    '2026-04-16 09:00:00',
    '2026-04-16 09:00:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM `plugin_demo_source_record`
    WHERE `title` = '源码插件 SQL 示例记录'
);
