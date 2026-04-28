-- Mock data: source plugin demo records.

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
    '该记录由 plugin-demo-source 的 mock-data 初始化，用于演示源码插件页面如何对插件自有数据表执行增删查改操作。',
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

INSERT IGNORE INTO `plugin_demo_source_record` (
    `title`,
    `content`,
    `attachment_name`,
    `attachment_path`,
    `created_at`,
    `updated_at`
)
SELECT
    '源码插件附件演示记录',
    '该记录用于演示源码插件记录列表中的附件字段展示，附件文件本身不会随 mock SQL 创建。',
    'source-plugin-demo.txt',
    'demo-record-files/source-plugin-demo.txt',
    '2026-04-17 10:30:00',
    '2026-04-17 10:30:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM `plugin_demo_source_record`
    WHERE `title` = '源码插件附件演示记录'
);
