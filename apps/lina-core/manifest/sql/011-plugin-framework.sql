-- ============================================================
-- 宿主插件表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin (
    id            INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    name          VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Plugin name',
    version       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT  'Plugin version',
    type          VARCHAR(32)  NOT NULL DEFAULT 'source' COMMENT  'Plugin top-level type: source/dynamic',
    installed     TINYINT      NOT NULL DEFAULT 0 COMMENT  'Installation status: 1=installed, 0=not installed',
    status        TINYINT      NOT NULL DEFAULT 0 COMMENT  'Enablement status: 1=enabled, 0=disabled',
    desired_state VARCHAR(32)  NOT NULL DEFAULT 'uninstalled' COMMENT  'Host desired state: uninstalled/installed/enabled',
    current_state VARCHAR(32)  NOT NULL DEFAULT 'uninstalled' COMMENT  'Host current state: uninstalled/installed/enabled/reconciling/failed',
    generation    BIGINT       NOT NULL DEFAULT 1 COMMENT  'Current host generation number',
    release_id    INT          NOT NULL DEFAULT 0 COMMENT  'Current active host release ID',
    manifest_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Plugin manifest file path',
    checksum      VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Plugin package checksum',
    installed_at  DATETIME                          COMMENT  'Installation time',
    enabled_at    DATETIME                          COMMENT  'Last enabled time',
    disabled_at   DATETIME                          COMMENT  'Last disabled time',
    remark        VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at    DATETIME                          COMMENT  'Creation time',
    updated_at    DATETIME                          COMMENT  'Update time',
    deleted_at    DATETIME                          COMMENT  'Deletion time',
    UNIQUE KEY uk_plugin_id (plugin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin registry table';

-- ============================================================
-- 插件发布记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_release (
    id                INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id         VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    release_version   VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Plugin version',
    type              VARCHAR(32)   NOT NULL DEFAULT 'source' COMMENT  'Plugin top-level type: source/dynamic',
    runtime_kind      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Runtime artifact type (currently only wasm)',
    schema_version    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'plugin.yaml manifest schema version',
    min_host_version  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Minimum compatible host version',
    max_host_version  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Maximum compatible host version',
    status            VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Release status: prepared/installed/active/uninstalled/failed',
    manifest_path     VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Plugin manifest path',
    package_path      VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Plugin source directory or runtime artifact path',
    checksum          VARCHAR(128)  NOT NULL DEFAULT '' COMMENT  'Plugin manifest or artifact checksum',
    manifest_snapshot LONGTEXT                           COMMENT  'Plugin manifest and resource summary snapshot in YAML, without concrete SQL or frontend file paths',
    created_at        DATETIME                           COMMENT  'Creation time',
    updated_at        DATETIME                           COMMENT  'Update time',
    deleted_at        DATETIME                           COMMENT  'Deletion time',
    UNIQUE KEY uk_plugin_release (plugin_id, release_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin release record table';

-- ============================================================
-- 插件迁移执行记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_migration (
    id              INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id       VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    release_id      INT           NOT NULL DEFAULT 0 COMMENT  'Owning plugin release ID',
    phase           VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Migration phase: install/uninstall/upgrade/rollback',
    migration_key   VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Migration execution key such as install-step-001, without concrete SQL path',
    checksum        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT  'Migration file checksum',
    execution_order INT           NOT NULL DEFAULT 0 COMMENT  'Execution order starting from 1',
    status          VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'Execution status: pending/succeeded/failed/skipped',
    executed_at     DATETIME                           COMMENT  'Execution time',
    error_message   VARCHAR(1024) NOT NULL DEFAULT '' COMMENT  'Failure reason or additional description',
    created_at      DATETIME                           COMMENT  'Creation time',
    updated_at      DATETIME                           COMMENT  'Update time',
    UNIQUE KEY uk_plugin_migration (plugin_id, release_id, phase, migration_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin migration execution record table';

-- ============================================================
-- 插件资源引用表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_resource_ref (
    id             INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    release_id     INT          NOT NULL DEFAULT 0 COMMENT  'Owning plugin release ID',
    resource_type  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Resource type: manifest/sql/frontend/menu/permission, etc.',
    resource_key   VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Resource unique key',
    resource_path  VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Resource location metadata, empty by default and without concrete frontend or SQL paths',
    owner_type     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Host object type: file/menu/route/slot, etc.',
    owner_key      VARCHAR(255) NOT NULL DEFAULT '' COMMENT  'Stable host object identifier',
    remark         VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at     DATETIME                          COMMENT  'Creation time',
    updated_at     DATETIME                          COMMENT  'Update time',
    deleted_at     DATETIME                          COMMENT  'Deletion time',
    UNIQUE KEY uk_plugin_resource_ref (plugin_id, release_id, resource_type, resource_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin resource reference table';

-- ============================================================
-- 插件节点状态表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_node_state (
    id                INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    release_id        INT          NOT NULL DEFAULT 0 COMMENT  'Owning plugin release ID',
    node_key          VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Node unique identifier',
    desired_state     VARCHAR(32)  NOT NULL DEFAULT '' COMMENT  'Node desired state',
    current_state     VARCHAR(32)  NOT NULL DEFAULT '' COMMENT  'Node current state',
    generation        BIGINT       NOT NULL DEFAULT 0 COMMENT  'Plugin generation number',
    last_heartbeat_at DATETIME                          COMMENT  'Last heartbeat time',
    error_message     VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Node error message',
    created_at        DATETIME                          COMMENT  'Creation time',
    updated_at        DATETIME                          COMMENT  'Update time',
    UNIQUE KEY uk_plugin_node_state (plugin_id, node_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin node state table';
