-- ------------------------------------------------------------
-- 001 plugin-demo-dynamic records SQL file
-- 001 plugin-demo-dynamic 记录 SQL 文件
-- Purpose: Stores tenant-scoped demo records used by the dynamic plugin sample to demonstrate host-service CRUD and attachments.
-- 用途：存储动态插件示例使用的租户级演示记录，用于展示宿主服务 CRUD 与附件能力。
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_demo_dynamic_record (
    "id"              VARCHAR(64) PRIMARY KEY,
    "tenant_id"       INT NOT NULL DEFAULT 0,
    "title"           VARCHAR(128) NOT NULL DEFAULT '',
    "content"         VARCHAR(1000) NOT NULL DEFAULT '',
    "attachment_name" VARCHAR(255) NOT NULL DEFAULT '',
    "attachment_path" VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_demo_dynamic_record IS 'Dynamic plugin demo record table';
COMMENT ON COLUMN plugin_demo_dynamic_record."id" IS 'Record ID';
COMMENT ON COLUMN plugin_demo_dynamic_record."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_demo_dynamic_record."title" IS 'Record title';
COMMENT ON COLUMN plugin_demo_dynamic_record."content" IS 'Record content';
COMMENT ON COLUMN plugin_demo_dynamic_record."attachment_name" IS 'Original attachment file name';
COMMENT ON COLUMN plugin_demo_dynamic_record."attachment_path" IS 'Relative attachment storage path';
COMMENT ON COLUMN plugin_demo_dynamic_record."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_demo_dynamic_record."updated_at" IS 'Update time';

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
    'This record is seeded by the plugin-demo-dynamic install SQL and demonstrates CRUD operations against the data table created during plugin installation.',
    '',
    '',
    '2026-04-16 09:00:00',
    '2026-04-16 09:00:00'
)
ON CONFLICT DO NOTHING;
