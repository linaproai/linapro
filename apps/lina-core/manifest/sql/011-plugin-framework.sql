-- ============================================================
-- Host plugin table
-- 宿主插件表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin (
    "id"             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "name"           VARCHAR(128) NOT NULL DEFAULT '',
    "version"        VARCHAR(32) NOT NULL DEFAULT '',
    "type"         VARCHAR(32) NOT NULL DEFAULT 'source',
    "installed"      SMALLINT NOT NULL DEFAULT 0,
    "status"         SMALLINT NOT NULL DEFAULT 0,
    "desired_state"  VARCHAR(32) NOT NULL DEFAULT 'uninstalled',
    "current_state"  VARCHAR(32) NOT NULL DEFAULT 'uninstalled',
    "generation"     BIGINT NOT NULL DEFAULT 1,
    "release_id"     INT NOT NULL DEFAULT 0,
    "manifest_path"  VARCHAR(255) NOT NULL DEFAULT '',
    "checksum"       VARCHAR(128) NOT NULL DEFAULT '',
    "installed_at"   TIMESTAMP,
    "enabled_at"     TIMESTAMP,
    "disabled_at"    TIMESTAMP,
    "remark"         VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"     TIMESTAMP,
    "updated_at"     TIMESTAMP,
    "deleted_at"     TIMESTAMP
);

COMMENT ON TABLE sys_plugin IS 'Plugin registry table';
COMMENT ON COLUMN sys_plugin."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin."name" IS 'Plugin name';
COMMENT ON COLUMN sys_plugin."version" IS 'Plugin version';
COMMENT ON COLUMN sys_plugin."type" IS 'Plugin top-level type: source/dynamic';
COMMENT ON COLUMN sys_plugin."installed" IS 'Installation status: 1=installed, 0=not installed';
COMMENT ON COLUMN sys_plugin."status" IS 'Enablement status: 1=enabled, 0=disabled';
COMMENT ON COLUMN sys_plugin."desired_state" IS 'Host desired state: uninstalled/installed/enabled';
COMMENT ON COLUMN sys_plugin."current_state" IS 'Host current state: uninstalled/installed/enabled/reconciling/failed';
COMMENT ON COLUMN sys_plugin."generation" IS 'Current host generation number';
COMMENT ON COLUMN sys_plugin."release_id" IS 'Current active host release ID';
COMMENT ON COLUMN sys_plugin."manifest_path" IS 'Plugin manifest file path';
COMMENT ON COLUMN sys_plugin."checksum" IS 'Plugin package checksum';
COMMENT ON COLUMN sys_plugin."installed_at" IS 'Installation time';
COMMENT ON COLUMN sys_plugin."enabled_at" IS 'Last enabled time';
COMMENT ON COLUMN sys_plugin."disabled_at" IS 'Last disabled time';
COMMENT ON COLUMN sys_plugin."remark" IS 'Remark';
COMMENT ON COLUMN sys_plugin."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_plugin."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_plugin_id ON sys_plugin ("plugin_id");

-- ============================================================
-- Plugin release record table
-- 插件发布记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_release (
    "id"                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"         VARCHAR(64) NOT NULL DEFAULT '',
    "release_version"   VARCHAR(32) NOT NULL DEFAULT '',
    "type"            VARCHAR(32) NOT NULL DEFAULT 'source',
    "runtime_kind"      VARCHAR(32) NOT NULL DEFAULT '',
    "schema_version"    VARCHAR(32) NOT NULL DEFAULT '',
    "min_host_version"  VARCHAR(32) NOT NULL DEFAULT '',
    "max_host_version"  VARCHAR(32) NOT NULL DEFAULT '',
    "status"            VARCHAR(32) NOT NULL DEFAULT '',
    "manifest_path"     VARCHAR(255) NOT NULL DEFAULT '',
    "package_path"      VARCHAR(255) NOT NULL DEFAULT '',
    "checksum"          VARCHAR(128) NOT NULL DEFAULT '',
    "manifest_snapshot" TEXT,
    "created_at"        TIMESTAMP,
    "updated_at"        TIMESTAMP,
    "deleted_at"        TIMESTAMP
);

COMMENT ON TABLE sys_plugin_release IS 'Plugin release record table';
COMMENT ON COLUMN sys_plugin_release."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_release."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_release."release_version" IS 'Plugin version';
COMMENT ON COLUMN sys_plugin_release."type" IS 'Plugin top-level type: source/dynamic';
COMMENT ON COLUMN sys_plugin_release."runtime_kind" IS 'Runtime artifact type (currently only wasm)';
COMMENT ON COLUMN sys_plugin_release."schema_version" IS 'plugin.yaml manifest schema version';
COMMENT ON COLUMN sys_plugin_release."min_host_version" IS 'Minimum compatible host version';
COMMENT ON COLUMN sys_plugin_release."max_host_version" IS 'Maximum compatible host version';
COMMENT ON COLUMN sys_plugin_release."status" IS 'Release status: prepared/installed/active/uninstalled/failed';
COMMENT ON COLUMN sys_plugin_release."manifest_path" IS 'Plugin manifest path';
COMMENT ON COLUMN sys_plugin_release."package_path" IS 'Plugin source directory or runtime artifact path';
COMMENT ON COLUMN sys_plugin_release."checksum" IS 'Plugin manifest or artifact checksum';
COMMENT ON COLUMN sys_plugin_release."manifest_snapshot" IS 'Plugin manifest and resource summary snapshot in YAML, without concrete SQL or frontend file paths';
COMMENT ON COLUMN sys_plugin_release."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_release."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_plugin_release."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_release_plugin_id_release_version ON sys_plugin_release ("plugin_id", "release_version");

-- ============================================================
-- Plugin migration execution record table
-- 插件迁移执行记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_migration (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"       VARCHAR(64) NOT NULL DEFAULT '',
    "release_id"      INT NOT NULL DEFAULT 0,
    "phase"           VARCHAR(32) NOT NULL DEFAULT '',
    "migration_key"   VARCHAR(255) NOT NULL DEFAULT '',
    "checksum"        VARCHAR(128) NOT NULL DEFAULT '',
    "execution_order" INT NOT NULL DEFAULT 0,
    "status"          VARCHAR(32) NOT NULL DEFAULT '',
    "executed_at"     TIMESTAMP,
    "error_message"   VARCHAR(1024) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP,
    "updated_at"      TIMESTAMP
);

COMMENT ON TABLE sys_plugin_migration IS 'Plugin migration execution record table';
COMMENT ON COLUMN sys_plugin_migration."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_migration."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_migration."release_id" IS 'Owning plugin release ID';
COMMENT ON COLUMN sys_plugin_migration."phase" IS 'Migration phase: install/uninstall/upgrade/rollback/mock';
COMMENT ON COLUMN sys_plugin_migration."migration_key" IS 'Migration execution key such as install-step-001, without concrete SQL path';
COMMENT ON COLUMN sys_plugin_migration."checksum" IS 'Migration file checksum';
COMMENT ON COLUMN sys_plugin_migration."execution_order" IS 'Execution order starting from 1';
COMMENT ON COLUMN sys_plugin_migration."status" IS 'Execution status: pending/succeeded/failed/skipped';
COMMENT ON COLUMN sys_plugin_migration."executed_at" IS 'Execution time';
COMMENT ON COLUMN sys_plugin_migration."error_message" IS 'Failure reason or additional description';
COMMENT ON COLUMN sys_plugin_migration."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_migration."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_migration_plugin_release_phase_key ON sys_plugin_migration ("plugin_id", "release_id", "phase", "migration_key");

-- ============================================================
-- Plugin resource reference table
-- 插件资源引用表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_resource_ref (
    "id"            INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"     VARCHAR(64) NOT NULL DEFAULT '',
    "release_id"    INT NOT NULL DEFAULT 0,
    "resource_type" VARCHAR(64) NOT NULL DEFAULT '',
    "resource_key"  VARCHAR(255) NOT NULL DEFAULT '',
    "resource_path" VARCHAR(255) NOT NULL DEFAULT '',
    "owner_type"    VARCHAR(64) NOT NULL DEFAULT '',
    "owner_key"     VARCHAR(255) NOT NULL DEFAULT '',
    "remark"        VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"    TIMESTAMP,
    "updated_at"    TIMESTAMP,
    "deleted_at"    TIMESTAMP
);

COMMENT ON TABLE sys_plugin_resource_ref IS 'Plugin resource reference table';
COMMENT ON COLUMN sys_plugin_resource_ref."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_resource_ref."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_resource_ref."release_id" IS 'Owning plugin release ID';
COMMENT ON COLUMN sys_plugin_resource_ref."resource_type" IS 'Resource type: manifest/sql/frontend/menu/permission, etc.';
COMMENT ON COLUMN sys_plugin_resource_ref."resource_key" IS 'Resource unique key';
COMMENT ON COLUMN sys_plugin_resource_ref."resource_path" IS 'Resource location metadata, empty by default and without concrete frontend or SQL paths';
COMMENT ON COLUMN sys_plugin_resource_ref."owner_type" IS 'Host object type: file/menu/route/slot, etc.';
COMMENT ON COLUMN sys_plugin_resource_ref."owner_key" IS 'Stable host object identifier';
COMMENT ON COLUMN sys_plugin_resource_ref."remark" IS 'Remark';
COMMENT ON COLUMN sys_plugin_resource_ref."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_resource_ref."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_plugin_resource_ref."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_resource_ref_plugin_release_type_key ON sys_plugin_resource_ref ("plugin_id", "release_id", "resource_type", "resource_key");

-- ============================================================
-- Plugin node state table
-- 插件节点状态表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_node_state (
    "id"                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_id"         VARCHAR(64) NOT NULL DEFAULT '',
    "release_id"        INT NOT NULL DEFAULT 0,
    "node_key"          VARCHAR(128) NOT NULL DEFAULT '',
    "desired_state"     VARCHAR(32) NOT NULL DEFAULT '',
    "current_state"     VARCHAR(32) NOT NULL DEFAULT '',
    "generation"        BIGINT NOT NULL DEFAULT 0,
    "last_heartbeat_at" TIMESTAMP,
    "error_message"     VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"        TIMESTAMP,
    "updated_at"        TIMESTAMP
);

COMMENT ON TABLE sys_plugin_node_state IS 'Plugin node state table';
COMMENT ON COLUMN sys_plugin_node_state."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_plugin_node_state."plugin_id" IS 'Plugin unique identifier (kebab-case)';
COMMENT ON COLUMN sys_plugin_node_state."release_id" IS 'Owning plugin release ID';
COMMENT ON COLUMN sys_plugin_node_state."node_key" IS 'Node unique identifier';
COMMENT ON COLUMN sys_plugin_node_state."desired_state" IS 'Node desired state';
COMMENT ON COLUMN sys_plugin_node_state."current_state" IS 'Node current state';
COMMENT ON COLUMN sys_plugin_node_state."generation" IS 'Plugin generation number';
COMMENT ON COLUMN sys_plugin_node_state."last_heartbeat_at" IS 'Last heartbeat time';
COMMENT ON COLUMN sys_plugin_node_state."error_message" IS 'Node error message';
COMMENT ON COLUMN sys_plugin_node_state."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_plugin_node_state."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_plugin_node_state_plugin_id_node_key ON sys_plugin_node_state ("plugin_id", "node_key");
