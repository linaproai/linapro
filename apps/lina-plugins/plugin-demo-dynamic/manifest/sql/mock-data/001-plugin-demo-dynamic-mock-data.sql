-- Mock data: dynamic plugin demo records.
-- 模拟数据：动态插件演示记录。

INSERT INTO plugin_demo_dynamic_record (
    "id",
    "tenant_id",
    "title",
    "content",
    "attachment_name",
    "attachment_path",
    "created_at",
    "updated_at"
)
VALUES (
    'plugin-demo-dynamic-mock-record',
    0,
    'Dynamic Plugin SQL Demo Record',
    'This record is loaded from plugin-demo-dynamic mock-data and demonstrates CRUD operations against the data table created during plugin installation.',
    '',
    '',
    '2026-04-16 09:00:00',
    '2026-04-16 09:00:00'
)
ON CONFLICT DO NOTHING;

INSERT INTO plugin_demo_dynamic_record (
    "id",
    "tenant_id",
    "title",
    "content",
    "attachment_name",
    "attachment_path",
    "created_at",
    "updated_at"
)
VALUES (
    'plugin-demo-dynamic-attachment-mock',
    0,
    'Dynamic Plugin Attachment Demo',
    'This mock record demonstrates attachment metadata for the hosted dynamic plugin page. The file itself is not created by SQL.',
    'dynamic-plugin-demo.txt',
    'demo-record-files/dynamic-plugin-demo.txt',
    '2026-04-17 10:30:00',
    '2026-04-17 10:30:00'
)
ON CONFLICT DO NOTHING;
