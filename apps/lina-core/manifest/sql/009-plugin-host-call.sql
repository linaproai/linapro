-- ------------------------------------------------------------
-- 009: Plugin Host Call
-- 009：插件宿主调用
-- Purpose: Stores plugin key-value runtime state and tenant-scoped enablement flags exposed through host-call services.
-- 用途：存储通过宿主调用服务暴露的插件键值运行态与租户级启用标记。
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_plugin_state (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"   VARCHAR(64) NOT NULL DEFAULT '',
    "tenant_id"   INT NOT NULL DEFAULT 0,
    "state_key"   VARCHAR(255) NOT NULL DEFAULT '',
    "state_value" TEXT,
    "enabled"     BOOL NOT NULL DEFAULT FALSE,
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP
);

COMMENT ON TABLE sys_plugin_state IS 'Plugin key-value state storage table';
COMMENT ON COLUMN sys_plugin_state."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_state."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_state."tenant_id" IS 'Plugin state tenant ID, 0 means platform/global state';
COMMENT ON COLUMN sys_plugin_state."state_key" IS 'State key';
COMMENT ON COLUMN sys_plugin_state."state_value" IS 'State value with JSON support';
COMMENT ON COLUMN sys_plugin_state."enabled" IS 'Whether the plugin is enabled for the tenant';
COMMENT ON COLUMN sys_plugin_state."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_state."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_state_plugin_tenant_key ON sys_plugin_state ("plugin_id", "tenant_id", "state_key");
CREATE INDEX IF NOT EXISTS idx_sys_plugin_state_tenant_enabled ON sys_plugin_state ("tenant_id", "plugin_id", "enabled");
