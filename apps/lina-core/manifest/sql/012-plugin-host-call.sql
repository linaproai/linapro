-- ------------------------------------------------------------
-- 012 plugin host-call SQL file
-- 012 插件宿主调用 SQL 文件
-- Plugin host call: key-value state storage table
-- 插件宿主调用：键值状态存储表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_plugin_state (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"   VARCHAR(64) NOT NULL DEFAULT '',
    "state_key"   VARCHAR(255) NOT NULL DEFAULT '',
    "state_value" TEXT,
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP
);

COMMENT ON TABLE sys_plugin_state IS 'Plugin key-value state storage table';
COMMENT ON COLUMN sys_plugin_state."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_state."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_state."state_key" IS 'State key';
COMMENT ON COLUMN sys_plugin_state."state_value" IS 'State value with JSON support';
COMMENT ON COLUMN sys_plugin_state."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_state."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_state_plugin_id_state_key ON sys_plugin_state ("plugin_id", "state_key");
