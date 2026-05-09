-- ------------------------------------------------------------
-- 001 plugin-demo-source records SQL file
-- 001 plugin-demo-source 记录 SQL 文件
-- plugin-demo-source demo record table
-- plugin-demo-source 示例记录表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_demo_source_record (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "title"           VARCHAR(128) NOT NULL DEFAULT '',
    "content"         VARCHAR(1000) NOT NULL DEFAULT '',
    "attachment_name" VARCHAR(255) NOT NULL DEFAULT '',
    "attachment_path" VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_demo_source_record IS 'Source plugin demo record table';
COMMENT ON COLUMN plugin_demo_source_record."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_demo_source_record."title" IS 'Record title';
COMMENT ON COLUMN plugin_demo_source_record."content" IS 'Record content';
COMMENT ON COLUMN plugin_demo_source_record."attachment_name" IS 'Original attachment file name';
COMMENT ON COLUMN plugin_demo_source_record."attachment_path" IS 'Relative attachment storage path';
COMMENT ON COLUMN plugin_demo_source_record."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_demo_source_record."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_demo_source_record_title ON plugin_demo_source_record ("title");

INSERT INTO plugin_demo_source_record (
    "title",
    "content",
    "attachment_name",
    "attachment_path",
    "created_at",
    "updated_at"
)
VALUES (
    '源码插件 SQL 示例记录',
    '该记录由 plugin-demo-source 安装 SQL 初始化，用于演示源码插件页面如何对插件自有数据表执行增删查改操作。',
    '',
    '',
    '2026-04-16 09:00:00',
    '2026-04-16 09:00:00'
)
ON CONFLICT DO NOTHING;
